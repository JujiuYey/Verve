DROP INDEX IF EXISTS idx_rag_jobs_asynq_task;
DROP INDEX IF EXISTS idx_rag_jobs_batch_status;
DROP INDEX IF EXISTS idx_rag_batches_root_created;

ALTER TABLE rag_index_jobs DROP COLUMN IF EXISTS asynq_task_id;
ALTER TABLE rag_index_jobs DROP COLUMN IF EXISTS batch_id;

DROP TABLE IF EXISTS rag_index_batches;
