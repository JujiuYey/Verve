-- RAG 索引基础表。
-- 当前方案只支持“单篇文档上传/保存后异步解析”，不再维护文件夹级 batch、Redis/asynq 队列或批量解析表。

-- 文档切块表：保存 Markdown 文档切分后的文本块，以及它们在向量库中的 point 映射。
-- Qdrant 负责相似度搜索；本表负责把搜索结果还原成文档、目录、标题路径和原文片段。
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

COMMENT ON TABLE rag_wiki_chunks IS 'RAG 文档切块表。';
COMMENT ON COLUMN rag_wiki_chunks.id IS '主键。';
COMMENT ON COLUMN rag_wiki_chunks.root_folder_id IS '知识库根文件夹 ID。';
COMMENT ON COLUMN rag_wiki_chunks.folder_id IS '所属文件夹 ID。';
COMMENT ON COLUMN rag_wiki_chunks.document_id IS '文档 ID。';
COMMENT ON COLUMN rag_wiki_chunks.document_title IS '文档标题。';
COMMENT ON COLUMN rag_wiki_chunks.folder_path IS '文件夹路径。';
COMMENT ON COLUMN rag_wiki_chunks.heading_path IS '标题路径。';
COMMENT ON COLUMN rag_wiki_chunks.chunk_index IS '切块序号。';
COMMENT ON COLUMN rag_wiki_chunks.block_type IS '内容块类型。';
COMMENT ON COLUMN rag_wiki_chunks.content IS '切块内容。';
COMMENT ON COLUMN rag_wiki_chunks.content_hash IS '内容哈希。';
COMMENT ON COLUMN rag_wiki_chunks.token_count IS '估算 token 数。';
COMMENT ON COLUMN rag_wiki_chunks.vector_point_id IS '向量库 point ID。';
COMMENT ON COLUMN rag_wiki_chunks.embedding_model IS 'Embedding 模型。';
COMMENT ON COLUMN rag_wiki_chunks.indexed_at IS '索引时间。';
COMMENT ON COLUMN rag_wiki_chunks.created_at IS '创建时间。';
COMMENT ON COLUMN rag_wiki_chunks.updated_at IS '更新时间。';

-- 单篇文档索引状态表：记录每次上传/保存文档后触发的索引结果。
CREATE TABLE IF NOT EXISTS rag_index_jobs (
  id VARCHAR(32) PRIMARY KEY,
  document_id VARCHAR(32) NOT NULL REFERENCES wiki_documents(id) ON DELETE CASCADE,
  root_folder_id VARCHAR(32) REFERENCES wiki_folders(id) ON DELETE SET NULL,
  status VARCHAR(32) NOT NULL,
  error_message TEXT,
  chunk_count INTEGER NOT NULL DEFAULT 0,
  attempt_count INTEGER NOT NULL DEFAULT 0,
  max_attempts INTEGER NOT NULL DEFAULT 3,
  started_at TIMESTAMPTZ,
  finished_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT chk_rag_index_jobs_status CHECK (status IN ('pending', 'running', 'completed', 'failed'))
);

COMMENT ON TABLE rag_index_jobs IS '单篇文档索引状态表。';
COMMENT ON COLUMN rag_index_jobs.id IS '主键。';
COMMENT ON COLUMN rag_index_jobs.document_id IS '文档 ID。';
COMMENT ON COLUMN rag_index_jobs.root_folder_id IS '知识库根文件夹 ID。';
COMMENT ON COLUMN rag_index_jobs.status IS '索引状态。';
COMMENT ON COLUMN rag_index_jobs.error_message IS '错误信息。';
COMMENT ON COLUMN rag_index_jobs.chunk_count IS '切块数量。';
COMMENT ON COLUMN rag_index_jobs.attempt_count IS '尝试次数。';
COMMENT ON COLUMN rag_index_jobs.max_attempts IS '最大尝试次数。';
COMMENT ON COLUMN rag_index_jobs.started_at IS '开始时间。';
COMMENT ON COLUMN rag_index_jobs.finished_at IS '结束时间。';
COMMENT ON COLUMN rag_index_jobs.created_at IS '创建时间。';
COMMENT ON COLUMN rag_index_jobs.updated_at IS '更新时间。';

-- 检索和清理常用索引。
CREATE INDEX IF NOT EXISTS idx_rag_chunks_root_folder ON rag_wiki_chunks(root_folder_id);
CREATE INDEX IF NOT EXISTS idx_rag_chunks_document ON rag_wiki_chunks(document_id);
CREATE INDEX IF NOT EXISTS idx_rag_chunks_folder ON rag_wiki_chunks(folder_id);
CREATE INDEX IF NOT EXISTS idx_rag_chunks_heading ON rag_wiki_chunks(heading_path);
CREATE INDEX IF NOT EXISTS idx_rag_jobs_document_status ON rag_index_jobs(document_id, status);
CREATE INDEX IF NOT EXISTS idx_rag_jobs_status_created ON rag_index_jobs(status, created_at DESC);
