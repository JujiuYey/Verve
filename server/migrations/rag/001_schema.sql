CREATE TABLE IF NOT EXISTS rag_wiki_chunks (
  id VARCHAR(32) PRIMARY KEY,
  root_folder_id VARCHAR(32) NOT NULL REFERENCES wiki_folders(id) ON DELETE CASCADE,
  folder_id VARCHAR(32) NOT NULL REFERENCES wiki_folders(id) ON DELETE CASCADE,
  document_id VARCHAR(32) NOT NULL REFERENCES wiki_documents(id) ON DELETE CASCADE,
  document_title VARCHAR(255) NOT NULL,
  folder_path TEXT NOT NULL,
  heading_path TEXT NOT NULL,
  chunk_index INTEGER NOT NULL,
  block_type VARCHAR(32) NOT NULL,
  content TEXT NOT NULL,
  content_hash VARCHAR(64) NOT NULL,
  token_count INTEGER NOT NULL DEFAULT 0,
  vector_point_id VARCHAR(64) NOT NULL,
  embedding_model VARCHAR(128) NOT NULL,
  indexed_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT uk_rag_chunks_doc_hash UNIQUE (document_id, content_hash),
  CONSTRAINT uk_rag_chunks_doc_index UNIQUE (document_id, chunk_index)
);

CREATE TABLE IF NOT EXISTS rag_index_jobs (
  id VARCHAR(32) PRIMARY KEY,
  document_id VARCHAR(32) NOT NULL REFERENCES wiki_documents(id) ON DELETE CASCADE,
  root_folder_id VARCHAR(32) REFERENCES wiki_folders(id) ON DELETE SET NULL,
  status VARCHAR(32) NOT NULL,
  error_message TEXT,
  chunk_count INTEGER NOT NULL DEFAULT 0,
  started_at TIMESTAMPTZ,
  finished_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT chk_rag_index_jobs_status CHECK (status IN ('pending', 'running', 'completed', 'failed'))
);

CREATE INDEX IF NOT EXISTS idx_rag_chunks_root_folder ON rag_wiki_chunks(root_folder_id);
CREATE INDEX IF NOT EXISTS idx_rag_chunks_document ON rag_wiki_chunks(document_id);
CREATE INDEX IF NOT EXISTS idx_rag_chunks_folder ON rag_wiki_chunks(folder_id);
CREATE INDEX IF NOT EXISTS idx_rag_chunks_heading ON rag_wiki_chunks(heading_path);
CREATE INDEX IF NOT EXISTS idx_rag_jobs_document_status ON rag_index_jobs(document_id, status);
CREATE INDEX IF NOT EXISTS idx_rag_jobs_status_created ON rag_index_jobs(status, created_at DESC);
