package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/hibiken/asynq"

	rag_db "verve/app/rag/models/db"
	rag_repo "verve/app/rag/repository"
	rag_service "verve/app/rag/service"
	wiki_repo "verve/app/wiki/repository"
)

const (
	TypeIndexDocument = "rag:index_document"
	QueueRAG          = "rag"
	MaxIndexAttempts  = 3
)

type IndexDocumentPayload struct {
	BatchID    string `json:"batch_id"`
	JobID      string `json:"job_id"`
	DocumentID string `json:"document_id"`
}

type Enqueuer struct {
	client  *asynq.Client
	batches *rag_repo.IndexBatchRepository
	jobs    *rag_repo.IndexJobRepository
	folders wiki_repo.FolderRepository
	docs    *wiki_repo.DocumentRepository
	indexer *rag_service.Indexer
}

func NewEnqueuer(
	client *asynq.Client,
	batches *rag_repo.IndexBatchRepository,
	jobs *rag_repo.IndexJobRepository,
	folders wiki_repo.FolderRepository,
	docs *wiki_repo.DocumentRepository,
	indexer *rag_service.Indexer,
) *Enqueuer {
	return &Enqueuer{client: client, batches: batches, jobs: jobs, folders: folders, docs: docs, indexer: indexer}
}

func (e *Enqueuer) EnqueueFolder(ctx context.Context, rootFolderID string) (*rag_db.IndexBatch, int, error) {
	if err := e.indexer.CheckReady(ctx); err != nil {
		return nil, 0, err
	}
	folderIDs, err := e.folders.GetAllSubFolderIDs(ctx, rootFolderID)
	if err != nil {
		return nil, 0, err
	}
	docs, err := e.docs.GetDocumentsByFolderIDs(ctx, folderIDs)
	if err != nil {
		return nil, 0, err
	}
	batch, err := e.batches.Create(ctx, rootFolderID, len(docs))
	if err != nil {
		return nil, 0, err
	}
	if len(docs) == 0 {
		if err := e.batches.MarkCompleted(ctx, batch.ID); err != nil {
			return nil, 0, err
		}
		return batch, 0, nil
	}
	if err := e.batches.MarkRunning(ctx, batch.ID); err != nil {
		return nil, 0, err
	}
	for _, doc := range docs {
		job, err := e.jobs.CreateQueued(ctx, batch.ID, rootFolderID, doc.ID, MaxIndexAttempts)
		if err != nil {
			return nil, 0, err
		}
		payload, err := json.Marshal(IndexDocumentPayload{BatchID: batch.ID, JobID: job.ID, DocumentID: doc.ID})
		if err != nil {
			return nil, 0, err
		}
		info, err := e.client.Enqueue(
			asynq.NewTask(TypeIndexDocument, payload),
			asynq.Queue(QueueRAG),
			asynq.MaxRetry(MaxIndexAttempts-1),
			asynq.Timeout(10*time.Minute),
		)
		if err != nil {
			return nil, 0, fmt.Errorf("enqueue index job failed: %w", err)
		}
		if err := e.jobs.SetTaskID(ctx, job.ID, info.ID); err != nil {
			return nil, 0, err
		}
	}
	return batch, len(docs), nil
}

type Processor struct {
	indexer *rag_service.Indexer
	batches *rag_repo.IndexBatchRepository
}

func NewProcessor(indexer *rag_service.Indexer, batches *rag_repo.IndexBatchRepository) *Processor {
	return &Processor{indexer: indexer, batches: batches}
}

func (p *Processor) HandleIndexDocument(ctx context.Context, task *asynq.Task) error {
	var payload IndexDocumentPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("invalid index document payload: %v: %w", err, asynq.SkipRetry)
	}
	if err := p.batches.MarkRunning(ctx, payload.BatchID); err != nil {
		return err
	}
	err := p.indexer.ProcessJob(ctx, payload.JobID)
	if refreshErr := p.batches.RefreshStatus(ctx, payload.BatchID); refreshErr != nil && err == nil {
		return refreshErr
	}
	if err != nil {
		if errors.Is(err, rag_service.ErrJobAttemptsExhausted) {
			return fmt.Errorf("%v: %w", err, asynq.SkipRetry)
		}
		return err
	}
	return nil
}
