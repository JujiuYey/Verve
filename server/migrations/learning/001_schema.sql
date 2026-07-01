-- ============================================
-- Learning 模块数据库架构
-- 依赖:
--   system/001_schema.sql(sys_users)
--   wiki/001_schema.sql(wiki_folders, wiki_documents)
-- ============================================

-- 1. 学习小节状态(来自 Wiki 文档)
CREATE TABLE learning_objectives (
    id VARCHAR(32) PRIMARY KEY,
    user_id VARCHAR(32) NOT NULL REFERENCES sys_users(id) ON DELETE CASCADE,
    stage_title VARCHAR(255),
    title VARCHAR(255) NOT NULL,
    detail TEXT,
    source_document_id VARCHAR(32) NOT NULL REFERENCES wiki_documents(id) ON DELETE CASCADE,
    source_folder_id VARCHAR(32) NOT NULL REFERENCES wiki_folders(id) ON DELETE CASCADE,
    source_folder_path TEXT,
    order_index INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    mastery_level VARCHAR(20) NOT NULL DEFAULT 'none',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_learning_objectives_status CHECK (status IN ('pending', 'active', 'completed', 'review')),
    CONSTRAINT chk_learning_objectives_mastery CHECK (mastery_level IN ('none','seen','heard','explained','written','verified'))
);
CREATE INDEX idx_learning_objectives_folder_order ON learning_objectives(source_folder_id, order_index);
CREATE INDEX idx_learning_objectives_user_id ON learning_objectives(user_id);
CREATE INDEX idx_learning_objectives_status ON learning_objectives(status);
CREATE INDEX idx_learning_objectives_source_document_id ON learning_objectives(source_document_id);
CREATE INDEX idx_learning_objectives_source_folder_id ON learning_objectives(source_folder_id);

-- 2. 学习会话(一节课)
CREATE TABLE learning_sessions (
    id VARCHAR(32) PRIMARY KEY,
    user_id VARCHAR(32) NOT NULL REFERENCES sys_users(id) ON DELETE CASCADE,
    objective_id VARCHAR(32) NOT NULL REFERENCES learning_objectives(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    summary TEXT,
    message_count INTEGER DEFAULT 0,
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_learning_sessions_status CHECK (status IN ('active', 'completed', 'abandoned'))
);
CREATE INDEX idx_learning_sessions_user_id ON learning_sessions(user_id);
CREATE INDEX idx_learning_sessions_objective_id ON learning_sessions(objective_id);
CREATE INDEX idx_learning_sessions_created_at ON learning_sessions(created_at DESC);

-- 3. 陪练消息
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
    CONSTRAINT chk_learning_messages_agent CHECK (agent_type IS NULL OR agent_type IN ('tutor','examiner','guide'))
);
CREATE INDEX idx_learning_messages_session_id ON learning_messages(session_id, created_at);
CREATE INDEX idx_learning_messages_agent_type ON learning_messages(agent_type) WHERE agent_type IS NOT NULL;
CREATE INDEX idx_learning_messages_total_tokens ON learning_messages(total_tokens) WHERE total_tokens IS NOT NULL;
CREATE INDEX idx_learning_messages_tool_result ON learning_messages USING GIN (tool_result);

-- 4. 练习与验证
CREATE TABLE learning_exercises (
    id VARCHAR(32) PRIMARY KEY,
    session_id VARCHAR(32) NOT NULL REFERENCES learning_sessions(id) ON DELETE CASCADE,
    objective_id VARCHAR(32) NOT NULL REFERENCES learning_objectives(id) ON DELETE CASCADE,
    user_id VARCHAR(32) NOT NULL REFERENCES sys_users(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL,
    prompt TEXT NOT NULL,
    user_answer TEXT,
    verdict VARCHAR(20),
    mastery_after VARCHAR(20),
    feedback TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_learning_exercises_type CHECK (type IN ('explain','choice','cloze','paste_output','code_snippet')),
    CONSTRAINT chk_learning_exercises_verdict CHECK (verdict IS NULL OR verdict IN ('pass','partial','fail')),
    CONSTRAINT chk_learning_exercises_mastery CHECK (mastery_after IS NULL OR mastery_after IN ('none','seen','heard','explained','written','verified'))
);
CREATE INDEX idx_learning_exercises_session_id ON learning_exercises(session_id);
CREATE INDEX idx_learning_exercises_objective_id ON learning_exercises(objective_id);
CREATE INDEX idx_learning_exercises_user_id ON learning_exercises(user_id);

-- 5. 学习画像(一个 Wiki 文件夹一份)
CREATE TABLE learning_profiles (
    id VARCHAR(32) PRIMARY KEY,
    user_id VARCHAR(32) NOT NULL REFERENCES sys_users(id) ON DELETE CASCADE,
    folder_id VARCHAR(32) NOT NULL REFERENCES wiki_folders(id) ON DELETE CASCADE,
    current_level VARCHAR(50),
    completed_topics JSONB,
    weak_points JSONB,
    verification_habits TEXT,
    next_goal TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_learning_profiles_folder UNIQUE (user_id, folder_id)
);
CREATE INDEX idx_learning_profiles_user_id ON learning_profiles(user_id);
CREATE INDEX idx_learning_profiles_folder_id ON learning_profiles(folder_id);

-- 6. 学习日志(按 Wiki 文件夹记录)
CREATE TABLE learning_journals (
    id VARCHAR(32) PRIMARY KEY,
    user_id VARCHAR(32) NOT NULL REFERENCES sys_users(id) ON DELETE CASCADE,
    folder_id VARCHAR(32) NOT NULL REFERENCES wiki_folders(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    learned TEXT,
    evidence TEXT,
    weak_points TEXT,
    next_step TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_learning_journal_day UNIQUE (user_id, folder_id, date)
);
CREATE INDEX idx_learning_journals_user_date ON learning_journals(user_id, date DESC);
CREATE INDEX idx_learning_journals_folder_id ON learning_journals(folder_id);

-- ============================================
-- 表注释
-- ============================================
COMMENT ON TABLE learning_objectives IS '学习小节状态(来自 Wiki 文档)';
COMMENT ON TABLE learning_sessions IS '学习会话(一节课)';
COMMENT ON TABLE learning_messages IS '陪练对话消息';
COMMENT ON TABLE learning_exercises IS '练习与验证记录';
COMMENT ON TABLE learning_profiles IS '学习画像(一个 Wiki 文件夹一份)';
COMMENT ON TABLE learning_journals IS '学习日志(按 Wiki 文件夹记录)';
