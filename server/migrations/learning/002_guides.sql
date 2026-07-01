-- 导学 Agent 结果缓存:避免同一学习小节、同一份 Markdown 反复消耗 token
CREATE TABLE learning_guides (
    id VARCHAR(32) PRIMARY KEY,
    objective_id VARCHAR(32) NOT NULL REFERENCES learning_objectives(id) ON DELETE CASCADE,
    user_id VARCHAR(32) NOT NULL REFERENCES sys_users(id) ON DELETE CASCADE,
    content_hash VARCHAR(64) NOT NULL,
    result JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_learning_guides_objective_hash UNIQUE (objective_id, content_hash)
);

CREATE INDEX idx_learning_guides_user_id ON learning_guides(user_id);
CREATE INDEX idx_learning_guides_objective_id ON learning_guides(objective_id);
