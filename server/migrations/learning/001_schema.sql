-- ============================================
-- Learning 模块数据库架构
-- 依赖:
--   system/001_schema.sql(sys_users)
--   wiki/001_schema.sql(wiki_folders, wiki_documents)
-- ============================================

-- 1. 文档学习会话
CREATE TABLE learning_sessions (
    id VARCHAR(32) PRIMARY KEY,
    user_id VARCHAR(32) NOT NULL REFERENCES sys_users(id) ON DELETE CASCADE,
    document_id VARCHAR(32) NOT NULL REFERENCES wiki_documents(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    summary TEXT,
    message_count INTEGER NOT NULL DEFAULT 0,
    started_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_learning_sessions_status CHECK (status IN ('active', 'completed', 'abandoned'))
);
CREATE INDEX idx_learning_sessions_user_id ON learning_sessions(user_id);
CREATE INDEX idx_learning_sessions_document_id ON learning_sessions(document_id);
CREATE INDEX idx_learning_sessions_created_at ON learning_sessions(created_at DESC);

-- 2. 陪练消息
CREATE TABLE learning_messages (
    id VARCHAR(32) PRIMARY KEY,
    session_id VARCHAR(32) NOT NULL REFERENCES learning_sessions(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL,
    agent_type VARCHAR(20),
    content TEXT NOT NULL,
    tool_used VARCHAR(100),
    tool_result JSONB,
    prompt_tokens BIGINT,
    completion_tokens BIGINT,
    total_tokens BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_learning_messages_role CHECK (role IN ('user', 'assistant', 'system')),
    CONSTRAINT chk_learning_messages_agent CHECK (agent_type IS NULL OR agent_type IN ('tutor', 'examiner', 'guide'))
);
CREATE INDEX idx_learning_messages_session_id ON learning_messages(session_id, created_at);
CREATE INDEX idx_learning_messages_agent_type ON learning_messages(agent_type) WHERE agent_type IS NOT NULL;
CREATE INDEX idx_learning_messages_total_tokens ON learning_messages(total_tokens) WHERE total_tokens IS NOT NULL;
CREATE INDEX idx_learning_messages_tool_result ON learning_messages USING GIN (tool_result);

-- 3. 费曼解释评审
CREATE TABLE learning_explanation_reviews (
    id VARCHAR(32) PRIMARY KEY,
    session_id VARCHAR(32) NOT NULL REFERENCES learning_sessions(id) ON DELETE CASCADE,
    document_id VARCHAR(32) NOT NULL REFERENCES wiki_documents(id) ON DELETE CASCADE,
    user_id VARCHAR(32) NOT NULL REFERENCES sys_users(id) ON DELETE CASCADE,
    explanation TEXT NOT NULL,
    heard_summary TEXT NOT NULL,
    clear_points JSONB NOT NULL DEFAULT '[]'::jsonb,
    confusing_points JSONB NOT NULL DEFAULT '[]'::jsonb,
    misconceptions JSONB NOT NULL DEFAULT '[]'::jsonb,
    follow_up_question TEXT NOT NULL DEFAULT '',
    explanation_summary TEXT NOT NULL,
    ready_to_wrap_up BOOLEAN NOT NULL DEFAULT FALSE,
    context_sufficient BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_learning_explanation_reviews_session_id
    ON learning_explanation_reviews(session_id, created_at, id);
CREATE INDEX idx_learning_explanation_reviews_document_id
    ON learning_explanation_reviews(document_id);
CREATE INDEX idx_learning_explanation_reviews_user_id
    ON learning_explanation_reviews(user_id);

COMMENT ON TABLE learning_sessions IS '以整篇 Wiki 文档为范围的学习会话';
COMMENT ON TABLE learning_messages IS '陪练对话消息';
COMMENT ON TABLE learning_explanation_reviews IS '费曼解释的多轮结构化评审';
