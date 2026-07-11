-- ============================================
-- Learning Memory 存储
-- 依赖:
--   system/001_schema.sql(sys_users)
--   wiki/001_schema.sql(wiki_folders, wiki_documents)
--   learning/001_schema.sql(learning_sessions)
-- ============================================

-- 1. 学习记忆事件(原始观察)
CREATE TABLE learning_memory_events (
    id VARCHAR(32) PRIMARY KEY,
    user_id VARCHAR(32) NOT NULL REFERENCES sys_users(id) ON DELETE CASCADE,
    folder_id VARCHAR(32) REFERENCES wiki_folders(id) ON DELETE CASCADE,
    document_id VARCHAR(32) REFERENCES wiki_documents(id) ON DELETE SET NULL,
    session_id VARCHAR(32) REFERENCES learning_sessions(id) ON DELETE SET NULL,
    source_type VARCHAR(40) NOT NULL,
    source_id VARCHAR(32),
    event_type VARCHAR(40) NOT NULL,
    content TEXT NOT NULL,
    evidence JSONB NOT NULL DEFAULT '{}'::jsonb,
    occurred_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_learning_memory_events_user_time ON learning_memory_events(user_id, occurred_at DESC);
CREATE INDEX idx_learning_memory_events_folder_time ON learning_memory_events(folder_id, occurred_at DESC);
CREATE INDEX idx_learning_memory_events_document_id ON learning_memory_events(document_id);
CREATE INDEX idx_learning_memory_events_session_id ON learning_memory_events(session_id);

-- 2. 学习记忆条目(可复用事实)
CREATE TABLE learning_memory_items (
    id VARCHAR(32) PRIMARY KEY,
    user_id VARCHAR(32) NOT NULL REFERENCES sys_users(id) ON DELETE CASCADE,
    folder_id VARCHAR(32) REFERENCES wiki_folders(id) ON DELETE CASCADE,
    document_id VARCHAR(32) REFERENCES wiki_documents(id) ON DELETE SET NULL,
    kind VARCHAR(40) NOT NULL,
    statement TEXT NOT NULL,
    evidence_event_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
    confidence VARCHAR(20) NOT NULL DEFAULT 'observed',
    last_seen_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_learning_memory_items_user_kind ON learning_memory_items(user_id, kind);
CREATE INDEX idx_learning_memory_items_folder_id ON learning_memory_items(folder_id);
CREATE INDEX idx_learning_memory_items_document_id ON learning_memory_items(document_id);

-- 3. 学习记忆汇总(按 Wiki 文件夹)
CREATE TABLE learning_memory_summaries (
    id VARCHAR(32) PRIMARY KEY,
    user_id VARCHAR(32) NOT NULL REFERENCES sys_users(id) ON DELETE CASCADE,
    folder_id VARCHAR(32) REFERENCES wiki_folders(id) ON DELETE CASCADE,
    summary TEXT NOT NULL,
    highlights JSONB NOT NULL DEFAULT '[]'::jsonb,
    generated_from_event_id VARCHAR(32),
    generated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_learning_memory_summaries_user_id ON learning_memory_summaries(user_id);
CREATE INDEX idx_learning_memory_summaries_folder_id ON learning_memory_summaries(folder_id);
CREATE UNIQUE INDEX uq_learning_memory_summaries_user_folder
    ON learning_memory_summaries(user_id, folder_id)
    WHERE folder_id IS NOT NULL;
CREATE UNIQUE INDEX uq_learning_memory_summaries_user_global
    ON learning_memory_summaries(user_id)
    WHERE folder_id IS NULL;

COMMENT ON TABLE learning_memory_events IS '学习记忆事件(原始观察)';
COMMENT ON TABLE learning_memory_items IS '学习记忆条目(可复用事实)';
COMMENT ON TABLE learning_memory_summaries IS '学习记忆汇总(按 Wiki 文件夹)';
