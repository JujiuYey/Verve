-- Wiki 文档不可变版本与变更申请。

ALTER TABLE wiki_documents
    ADD COLUMN current_version BIGINT NOT NULL DEFAULT 1,
    ADD COLUMN content_hash VARCHAR(64);

CREATE TABLE wiki_document_change_requests (
    id VARCHAR(32) PRIMARY KEY,
    document_id VARCHAR(32) NOT NULL REFERENCES wiki_documents(id) ON DELETE CASCADE,
    requested_by VARCHAR(32) NOT NULL REFERENCES sys_users(id) ON DELETE CASCADE,
    source_type VARCHAR(32) NOT NULL,
    source_id VARCHAR(32) NOT NULL,
    request_id VARCHAR(64) NOT NULL,
    replaces_change_request_id VARCHAR(32) REFERENCES wiki_document_change_requests(id) ON DELETE SET NULL,
    base_version BIGINT NOT NULL,
    instruction TEXT NOT NULL,
    change_summary TEXT NOT NULL,
    proposed_content TEXT NOT NULL,
    proposed_diff TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'proposed',
    error_message TEXT,
    applied_version BIGINT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    applied_at TIMESTAMP,
    CONSTRAINT uq_wiki_change_requests_document_request UNIQUE (document_id, request_id),
    CONSTRAINT chk_wiki_change_requests_status CHECK (status IN ('proposed', 'applied', 'failed', 'cancelled', 'conflict'))
);

CREATE TABLE wiki_document_revisions (
    id VARCHAR(32) PRIMARY KEY,
    document_id VARCHAR(32) NOT NULL REFERENCES wiki_documents(id) ON DELETE CASCADE,
    version BIGINT NOT NULL,
    object_path TEXT NOT NULL,
    content_hash VARCHAR(64) NOT NULL,
    file_size BIGINT NOT NULL,
    change_request_id VARCHAR(32) REFERENCES wiki_document_change_requests(id) ON DELETE SET NULL,
    changed_by VARCHAR(32) NOT NULL REFERENCES sys_users(id) ON DELETE RESTRICT,
    change_summary TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_wiki_document_revisions_document_version UNIQUE (document_id, version)
);

CREATE INDEX idx_wiki_change_requests_document_created
    ON wiki_document_change_requests(document_id, created_at DESC);
CREATE INDEX idx_wiki_change_requests_requested_status
    ON wiki_document_change_requests(requested_by, status, created_at DESC);
CREATE INDEX idx_wiki_document_revisions_document_version
    ON wiki_document_revisions(document_id, version DESC);

COMMENT ON TABLE wiki_document_change_requests IS 'Wiki文档变更申请表';
COMMENT ON COLUMN wiki_documents.current_version IS '当前文档版本号';
COMMENT ON COLUMN wiki_documents.content_hash IS '当前内容哈希';
COMMENT ON COLUMN wiki_document_change_requests.id IS '主键ID';
COMMENT ON COLUMN wiki_document_change_requests.document_id IS '文档ID';
COMMENT ON COLUMN wiki_document_change_requests.requested_by IS '申请用户ID';
COMMENT ON COLUMN wiki_document_change_requests.source_type IS '来源类型';
COMMENT ON COLUMN wiki_document_change_requests.source_id IS '来源记录ID';
COMMENT ON COLUMN wiki_document_change_requests.request_id IS '幂等请求ID';
COMMENT ON COLUMN wiki_document_change_requests.replaces_change_request_id IS '替代的变更申请ID';
COMMENT ON COLUMN wiki_document_change_requests.base_version IS '生成申请时的文档版本号';
COMMENT ON COLUMN wiki_document_change_requests.instruction IS '用户修改要求';
COMMENT ON COLUMN wiki_document_change_requests.change_summary IS '修改摘要';
COMMENT ON COLUMN wiki_document_change_requests.proposed_content IS '建议完整内容';
COMMENT ON COLUMN wiki_document_change_requests.proposed_diff IS '建议内容差异';
COMMENT ON COLUMN wiki_document_change_requests.status IS '申请状态';
COMMENT ON COLUMN wiki_document_change_requests.error_message IS '失败原因';
COMMENT ON COLUMN wiki_document_change_requests.applied_version IS '已应用版本号';
COMMENT ON COLUMN wiki_document_change_requests.created_at IS '创建时间';
COMMENT ON COLUMN wiki_document_change_requests.updated_at IS '更新时间';
COMMENT ON COLUMN wiki_document_change_requests.applied_at IS '应用时间';
COMMENT ON TABLE wiki_document_revisions IS 'Wiki文档不可变修订表';
COMMENT ON COLUMN wiki_document_revisions.id IS '主键ID';
COMMENT ON COLUMN wiki_document_revisions.document_id IS '文档ID';
COMMENT ON COLUMN wiki_document_revisions.version IS '文档版本号';
COMMENT ON COLUMN wiki_document_revisions.object_path IS '不可变对象路径';
COMMENT ON COLUMN wiki_document_revisions.content_hash IS '内容哈希';
COMMENT ON COLUMN wiki_document_revisions.file_size IS '文件大小(字节)';
COMMENT ON COLUMN wiki_document_revisions.change_request_id IS '变更申请ID';
COMMENT ON COLUMN wiki_document_revisions.changed_by IS '修改用户ID';
COMMENT ON COLUMN wiki_document_revisions.change_summary IS '修改摘要';
COMMENT ON COLUMN wiki_document_revisions.created_at IS '创建时间';
