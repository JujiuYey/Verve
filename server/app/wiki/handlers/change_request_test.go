package handlers

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	rag_db "verve/app/rag/models/db"
	wiki_db "verve/app/wiki/models/db"
)

type fakeChangeRequestVersionService struct {
	job *rag_db.IndexJob
}

func (f *fakeChangeRequestVersionService) ApplyChangeRequest(context.Context, string) (*wiki_db.DocumentRevision, *rag_db.IndexJob, error) {
	return &wiki_db.DocumentRevision{ID: "revision-1"}, f.job, nil
}

func (f *fakeChangeRequestVersionService) CancelChangeRequest(context.Context, string) error {
	return nil
}

type fakeChangeRequestFinder struct{}

func (f *fakeChangeRequestFinder) FindChangeRequest(context.Context, string) (*wiki_db.DocumentChangeRequest, error) {
	return &wiki_db.DocumentChangeRequest{ID: "request-1"}, nil
}

type recordingJobProcessor struct {
	processed chan string
}

func (r *recordingJobProcessor) ProcessJob(_ context.Context, jobID string) error {
	r.processed <- jobID
	return nil
}

func TestChangeRequestApplyStartsCreatedIndexJob(t *testing.T) {
	processor := &recordingJobProcessor{processed: make(chan string, 1)}
	handler := NewChangeRequestHandlerWithDependencies(
		&fakeChangeRequestVersionService{job: &rag_db.IndexJob{ID: "job-1", Status: "pending"}},
		&fakeChangeRequestFinder{},
		processor,
	)
	app := fiber.New()
	app.Post("/change-requests/:id/apply", func(c *fiber.Ctx) error {
		return handler.Apply(c)
	})

	resp, err := app.Test(httptest.NewRequest("POST", "/change-requests/request-1/apply", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}

	select {
	case jobID := <-processor.processed:
		if jobID != "job-1" {
			t.Fatalf("processed job = %q, want job-1", jobID)
		}
	case <-time.After(time.Second):
		t.Fatal("created index job was not processed")
	}
}
