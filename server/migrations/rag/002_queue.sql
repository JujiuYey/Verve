CREATE TABLE IF NOT EXISTS rag_index_batches (
  id VARCHAR(32) PRIMARY KEY,
  root_folder_id VARCHAR(32) NOT NULL REFERENCES wiki_folders(id) ON DELETE CASCADE,
  status VARCHAR(32) NOT NULL,
  total_count INTEGER NOT NULL DEFAULT 0,
  error_message TEXT,
  started_at TIMESTAMPTZ,
  finished_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT chk_rag_index_batches_status CHECK (status IN ('pending', 'running', 'completed', 'failed', 'canceled'))
);

ALTER TABLE rag_index_jobs ADD COLUMN IF NOT EXISTS batch_id VARCHAR(32) REFERENCES rag_index_batches(id) ON DELETE CASCADE;
ALTER TABLE rag_index_jobs ADD COLUMN IF NOT EXISTS asynq_task_id VARCHAR(128);
ALTER TABLE rag_index_jobs ADD COLUMN IF NOT EXISTS attempt_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE rag_index_jobs ADD COLUMN IF NOT EXISTS max_attempts INTEGER NOT NULL DEFAULT 3;

CREATE INDEX IF NOT EXISTS idx_rag_batches_root_created ON rag_index_batches(root_folder_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_rag_jobs_batch_status ON rag_index_jobs(batch_id, status);
CREATE INDEX IF NOT EXISTS idx_rag_jobs_asynq_task ON rag_index_jobs(asynq_task_id);
