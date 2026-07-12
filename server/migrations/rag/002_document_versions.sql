-- RAG 索引任务和切块与 Wiki 文档版本绑定。

ALTER TABLE rag_wiki_chunks
    ADD COLUMN document_version BIGINT NOT NULL DEFAULT 1;

ALTER TABLE rag_index_jobs
    ADD COLUMN document_version BIGINT NOT NULL DEFAULT 1,
    ADD COLUMN object_path TEXT;

UPDATE rag_index_jobs rij
SET object_path = d.file_path
FROM wiki_documents d
WHERE d.id = rij.document_id;

ALTER TABLE rag_index_jobs
    ALTER COLUMN object_path SET NOT NULL;

ALTER TABLE rag_wiki_chunks
    DROP CONSTRAINT uk_rag_chunks_doc_hash,
    DROP CONSTRAINT uk_rag_chunks_doc_index,
    ADD CONSTRAINT uk_rag_chunks_doc_version_hash UNIQUE (document_id, document_version, content_hash),
    ADD CONSTRAINT uk_rag_chunks_doc_version_index UNIQUE (document_id, document_version, chunk_index);

ALTER TABLE rag_index_jobs
    DROP CONSTRAINT chk_rag_index_jobs_status,
    ADD CONSTRAINT chk_rag_index_jobs_status CHECK (status IN ('pending', 'running', 'completed', 'failed', 'superseded'));

WITH ranked_jobs AS (
    SELECT id, row_number() OVER (PARTITION BY document_id ORDER BY created_at DESC, id DESC) AS row_num
    FROM rag_index_jobs
)
UPDATE rag_index_jobs rij
SET status = 'superseded', updated_at = CURRENT_TIMESTAMP
FROM ranked_jobs ranked
WHERE ranked.id = rij.id AND ranked.row_num > 1;

CREATE UNIQUE INDEX uq_rag_index_jobs_current_document_version
    ON rag_index_jobs(document_id, document_version)
    WHERE status <> 'superseded';
CREATE INDEX idx_rag_chunks_document_version
    ON rag_wiki_chunks(document_id, document_version);

COMMENT ON COLUMN rag_wiki_chunks.document_version IS '文档版本号。';
COMMENT ON COLUMN rag_index_jobs.document_version IS '文档版本号。';
COMMENT ON COLUMN rag_index_jobs.object_path IS '待索引对象路径。';
