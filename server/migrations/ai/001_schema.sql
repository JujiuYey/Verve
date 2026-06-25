-- ============================================
-- AI 模块数据库架构
-- 依赖: system/001_schema.sql, wiki/001_schema.sql
-- ============================================

-- 1. Agent 会话表 (ai_agent_sessions)
CREATE TABLE ai_agent_sessions (
    id VARCHAR(32) PRIMARY KEY,
    -- 会话标题（可以从第一条消息自动生成）
    title VARCHAR(255),
    -- 所属用户
    user_id VARCHAR(32) REFERENCES sys_users(id) ON DELETE CASCADE,
    -- 消息计数（冗余字段，提高查询效率）
    message_count INTEGER DEFAULT 0,
    -- 时间戳
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_agent_sessions_created_at ON ai_agent_sessions(created_at DESC);
CREATE INDEX idx_agent_sessions_updated_at ON ai_agent_sessions(updated_at DESC);
CREATE INDEX idx_agent_sessions_user_id ON ai_agent_sessions(user_id);

-- 2. Agent 消息表 (ai_agent_messages)
CREATE TABLE ai_agent_messages (
    id VARCHAR(32) PRIMARY KEY,
    -- 关联的会话 ID
    session_id VARCHAR(32) NOT NULL REFERENCES ai_agent_sessions(id) ON DELETE CASCADE,
    -- 角色: user 或 assistant
    role VARCHAR(20) NOT NULL,
    -- 消息内容
    content TEXT NOT NULL,
    -- 使用的工具名称（仅 assistant 消息有值）
    tool_used VARCHAR(100),
    -- 工具执行结果（JSON 格式存储）
    tool_result JSONB,
    -- Token 用量
    prompt_tokens BIGINT,
    completion_tokens BIGINT,
    total_tokens BIGINT,
    -- 创建时间
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_agent_messages_role CHECK (role IN ('user', 'assistant'))
);

CREATE INDEX idx_agent_messages_session_id ON ai_agent_messages(session_id, created_at);
CREATE INDEX idx_agent_messages_created_at ON ai_agent_messages(created_at DESC);
CREATE INDEX idx_agent_messages_tool_used ON ai_agent_messages(tool_used) WHERE tool_used IS NOT NULL;
CREATE INDEX idx_agent_messages_total_tokens ON ai_agent_messages(total_tokens) WHERE total_tokens IS NOT NULL;
CREATE INDEX idx_agent_messages_tool_result ON ai_agent_messages USING GIN (tool_result);

-- 3. RAG 会话表 (ai_rag_sessions)
CREATE TABLE ai_rag_sessions (
    id VARCHAR(32) PRIMARY KEY,
    -- 会话标题（可以从第一条消息自动生成）
    title VARCHAR(255),
    -- 所属用户
    user_id VARCHAR(32) REFERENCES sys_users(id) ON DELETE CASCADE,
    -- 关联的文件夹（必填）
    folder_id VARCHAR(32) NOT NULL REFERENCES wiki_folders(id) ON DELETE CASCADE,
    -- 关联的文档（可选，如果指定则该会话只针对该文档）
    document_id VARCHAR(32) REFERENCES wiki_documents(id) ON DELETE CASCADE,
    -- 消息计数（冗余字段，提高查询效率）
    message_count INTEGER DEFAULT 0,
    -- 时间戳
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_rag_sessions_created_at ON ai_rag_sessions(created_at DESC);
CREATE INDEX idx_rag_sessions_updated_at ON ai_rag_sessions(updated_at DESC);
CREATE INDEX idx_rag_sessions_user_id ON ai_rag_sessions(user_id);
CREATE INDEX idx_rag_sessions_folder_id ON ai_rag_sessions(folder_id);
CREATE INDEX idx_rag_sessions_document_id ON ai_rag_sessions(document_id);

-- 4. RAG 消息表 (ai_rag_messages)
CREATE TABLE ai_rag_messages (
    id VARCHAR(32) PRIMARY KEY,
    -- 关联的会话 ID
    session_id VARCHAR(32) NOT NULL REFERENCES ai_rag_sessions(id) ON DELETE CASCADE,
    -- 角色: user 或 assistant
    role VARCHAR(20) NOT NULL,
    -- 消息内容
    content TEXT NOT NULL,
    -- RAG 检索到的文档片段（JSON 格式存储）
    retrieved_chunks JSONB,
    -- 相关性分数（可选，存储平均相关性）
    relevance_score FLOAT,
    -- Token 用量
    prompt_tokens BIGINT,
    completion_tokens BIGINT,
    total_tokens BIGINT,
    -- 创建时间
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_rag_messages_role CHECK (role IN ('user', 'assistant'))
);

CREATE INDEX idx_rag_messages_session_id ON ai_rag_messages(session_id, created_at);
CREATE INDEX idx_rag_messages_created_at ON ai_rag_messages(created_at DESC);
CREATE INDEX idx_rag_messages_relevance_score ON ai_rag_messages(relevance_score) WHERE relevance_score IS NOT NULL;
CREATE INDEX idx_rag_messages_total_tokens ON ai_rag_messages(total_tokens) WHERE total_tokens IS NOT NULL;
CREATE INDEX idx_rag_messages_retrieved_chunks ON ai_rag_messages USING GIN (retrieved_chunks);

-- ============================================
-- 触发器
-- ============================================

-- 自动更新 Agent 会话的消息计数
CREATE OR REPLACE FUNCTION update_agent_session_message_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE ai_agent_sessions
        SET message_count = message_count + 1,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = NEW.session_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE ai_agent_sessions
        SET message_count = message_count - 1,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = OLD.session_id;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_agent_session_message_count_trigger
    AFTER INSERT OR DELETE ON ai_agent_messages
    FOR EACH ROW
    EXECUTE FUNCTION update_agent_session_message_count();

-- 自动更新 RAG 会话的消息计数
CREATE OR REPLACE FUNCTION update_rag_session_message_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE ai_rag_sessions
        SET message_count = message_count + 1,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = NEW.session_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE ai_rag_sessions
        SET message_count = message_count - 1,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = OLD.session_id;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_rag_session_message_count_trigger
    AFTER INSERT OR DELETE ON ai_rag_messages
    FOR EACH ROW
    EXECUTE FUNCTION update_rag_session_message_count();

-- ============================================
-- 表注释
-- ============================================
COMMENT ON TABLE ai_agent_sessions IS 'Agent 对话会话表';
COMMENT ON TABLE ai_agent_messages IS 'Agent 对话消息表';
COMMENT ON TABLE ai_rag_sessions IS 'RAG 对话会话表';
COMMENT ON TABLE ai_rag_messages IS 'RAG 对话消息表';

COMMENT ON COLUMN ai_rag_messages.prompt_tokens IS '输入 token 数量（用户消息和上下文）';
COMMENT ON COLUMN ai_rag_messages.completion_tokens IS '输出 token 数量（AI 回复）';
COMMENT ON COLUMN ai_rag_messages.total_tokens IS '总 token 数量（prompt_tokens + completion_tokens）';
COMMENT ON COLUMN ai_agent_messages.prompt_tokens IS '输入 token 数量（用户消息和上下文）';
COMMENT ON COLUMN ai_agent_messages.completion_tokens IS '输出 token 数量（AI 回复）';
COMMENT ON COLUMN ai_agent_messages.total_tokens IS '总 token 数量（prompt_tokens + completion_tokens）';
