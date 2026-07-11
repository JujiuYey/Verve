-- ============================================
-- 三 Agent 学习轮次数据基础
-- 依赖:
--   learning/001_schema.sql
--   learning/003_memory.sql
-- ============================================

CREATE TABLE learning_turns (
    id VARCHAR(32) PRIMARY KEY,
    session_id VARCHAR(32) NOT NULL REFERENCES learning_sessions(id) ON DELETE CASCADE,
    request_id VARCHAR(64) NOT NULL,
    agent_type VARCHAR(32) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'processing',
    error_code VARCHAR(64),
    error_message TEXT,
    started_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_learning_turns_session_request UNIQUE (session_id, request_id),
    CONSTRAINT chk_learning_turns_agent CHECK (agent_type IN ('listener', 'teacher', 'curator')),
    CONSTRAINT chk_learning_turns_status CHECK (status IN ('processing', 'completed', 'failed'))
);
CREATE INDEX idx_learning_turns_session_created ON learning_turns(session_id, created_at, id);
CREATE INDEX idx_learning_turns_status ON learning_turns(status);

ALTER TABLE learning_messages ADD COLUMN turn_id VARCHAR(32);
ALTER TABLE learning_explanation_reviews ADD COLUMN turn_id VARCHAR(32);

-- 历史 review 和 message 无法可靠配对，分别回填为可审计的 legacy turn。
INSERT INTO learning_turns (
    id, session_id, request_id, agent_type, status,
    started_at, completed_at, created_at, updated_at
)
SELECT
    md5('legacy-review:' || id), session_id, 'legacy-review:' || id, 'listener', 'completed',
    created_at, created_at, created_at, created_at
FROM learning_explanation_reviews;

UPDATE learning_explanation_reviews
SET turn_id = md5('legacy-review:' || id);

INSERT INTO learning_turns (
    id, session_id, request_id, agent_type, status,
    started_at, completed_at, created_at, updated_at
)
SELECT
    md5('legacy-message:' || id), session_id, 'legacy-message:' || id,
    CASE WHEN agent_type = 'guide' THEN 'teacher' ELSE 'listener' END,
    'completed', created_at, created_at, created_at, updated_at
FROM learning_messages;

UPDATE learning_messages
SET turn_id = md5('legacy-message:' || id);

-- review 原有解释和结构化回复迁入消息表，保证历史对话仍可读取。
INSERT INTO learning_messages (
    id, session_id, turn_id, role, agent_type, content, created_at, updated_at
)
SELECT
    md5('legacy-review-user:' || id), session_id, md5('legacy-review:' || id),
    'user', 'tutor', explanation, created_at, created_at
FROM learning_explanation_reviews;

INSERT INTO learning_messages (
    id, session_id, turn_id, role, agent_type, content, created_at, updated_at
)
SELECT
    md5('legacy-review-assistant:' || id), session_id, md5('legacy-review:' || id),
    'assistant', 'tutor',
    jsonb_build_object(
        'heard_summary', heard_summary,
        'clear_points', clear_points,
        'confusing_points', confusing_points,
        'misconceptions', misconceptions,
        'follow_up_question', follow_up_question,
        'explanation_summary', explanation_summary,
        'ready_to_wrap_up', ready_to_wrap_up,
        'context_sufficient', context_sufficient
    )::text,
    created_at, created_at
FROM learning_explanation_reviews;

ALTER TABLE learning_messages
    ALTER COLUMN turn_id SET NOT NULL,
    ADD CONSTRAINT fk_learning_messages_turn
        FOREIGN KEY (turn_id) REFERENCES learning_turns(id) ON DELETE CASCADE;
ALTER TABLE learning_explanation_reviews
    ALTER COLUMN turn_id SET NOT NULL,
    ADD CONSTRAINT fk_learning_explanation_reviews_turn
        FOREIGN KEY (turn_id) REFERENCES learning_turns(id) ON DELETE CASCADE,
    ADD CONSTRAINT uq_learning_explanation_reviews_turn UNIQUE (turn_id);

DROP INDEX IF EXISTS idx_learning_messages_agent_type;
DROP INDEX IF EXISTS idx_learning_explanation_reviews_session_id;
DROP INDEX IF EXISTS idx_learning_explanation_reviews_document_id;
DROP INDEX IF EXISTS idx_learning_explanation_reviews_user_id;

CREATE INDEX idx_learning_messages_turn_id ON learning_messages(turn_id, created_at, id);
CREATE INDEX idx_learning_explanation_reviews_created_at ON learning_explanation_reviews(created_at, id);

ALTER TABLE learning_sessions DROP COLUMN message_count;
ALTER TABLE learning_messages DROP CONSTRAINT chk_learning_messages_agent;
ALTER TABLE learning_messages DROP COLUMN agent_type;
ALTER TABLE learning_explanation_reviews
    DROP COLUMN session_id,
    DROP COLUMN document_id,
    DROP COLUMN user_id,
    DROP COLUMN explanation;

CREATE TABLE learning_teaching_interventions (
    id VARCHAR(32) PRIMARY KEY,
    turn_id VARCHAR(32) NOT NULL REFERENCES learning_turns(id) ON DELETE CASCADE,
    question_summary TEXT NOT NULL,
    knowledge_gaps JSONB NOT NULL DEFAULT '[]'::jsonb,
    explanation_summary TEXT NOT NULL,
    key_points JSONB NOT NULL DEFAULT '[]'::jsonb,
    examples JSONB NOT NULL DEFAULT '[]'::jsonb,
    evidence JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_learning_teaching_interventions_turn UNIQUE (turn_id)
);

CREATE UNIQUE INDEX uq_learning_memory_events_source
    ON learning_memory_events(source_type, source_id, event_type)
    WHERE source_id IS NOT NULL;

COMMENT ON TABLE learning_turns IS '单次用户输入的 Agent 处理轮次';
COMMENT ON COLUMN learning_turns.id IS '轮次ID';
COMMENT ON COLUMN learning_turns.session_id IS '学习会话ID';
COMMENT ON COLUMN learning_turns.request_id IS '请求幂等标识';
COMMENT ON COLUMN learning_turns.agent_type IS '处理Agent类型';
COMMENT ON COLUMN learning_turns.status IS '处理状态';
COMMENT ON COLUMN learning_turns.error_code IS '失败错误码';
COMMENT ON COLUMN learning_turns.error_message IS '失败原因';
COMMENT ON COLUMN learning_turns.started_at IS '开始时间';
COMMENT ON COLUMN learning_turns.completed_at IS '完成时间';
COMMENT ON COLUMN learning_turns.created_at IS '创建时间';
COMMENT ON COLUMN learning_turns.updated_at IS '更新时间';

COMMENT ON COLUMN learning_messages.turn_id IS '处理轮次ID';
COMMENT ON COLUMN learning_explanation_reviews.turn_id IS '倾听轮次ID';

COMMENT ON TABLE learning_teaching_interventions IS 'LearningTeacher 教学干预';
COMMENT ON COLUMN learning_teaching_interventions.id IS '干预ID';
COMMENT ON COLUMN learning_teaching_interventions.turn_id IS '教学轮次ID';
COMMENT ON COLUMN learning_teaching_interventions.question_summary IS '卡点摘要';
COMMENT ON COLUMN learning_teaching_interventions.knowledge_gaps IS '前置知识缺口';
COMMENT ON COLUMN learning_teaching_interventions.explanation_summary IS '教学内容摘要';
COMMENT ON COLUMN learning_teaching_interventions.key_points IS '讲解关键点';
COMMENT ON COLUMN learning_teaching_interventions.examples IS '教学示例';
COMMENT ON COLUMN learning_teaching_interventions.evidence IS '文档与检索依据';
COMMENT ON COLUMN learning_teaching_interventions.created_at IS '创建时间';
