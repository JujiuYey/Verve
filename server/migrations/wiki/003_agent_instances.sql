-- Wiki 学习 Agent 实例：一个用户在一个知识库根目录下对应一个长期学习 agent。
-- RAG 负责知识检索，本表负责把“当前学习的是哪套 wiki”持久化。

CREATE TABLE IF NOT EXISTS wiki_agent_instances (
  id VARCHAR(32) PRIMARY KEY,
  user_id VARCHAR(32) NOT NULL REFERENCES sys_users(id) ON DELETE CASCADE,
  root_folder_id VARCHAR(32) NOT NULL REFERENCES wiki_folders(id) ON DELETE CASCADE,
  agent_key VARCHAR(64) NOT NULL DEFAULT 'wiki_learning',
  name VARCHAR(120) NOT NULL,
  description TEXT,
  status VARCHAR(20) NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT uk_wiki_agent_instances_user_root UNIQUE (user_id, root_folder_id),
  CONSTRAINT chk_wiki_agent_instances_status CHECK (status IN ('active', 'archived'))
);

CREATE INDEX IF NOT EXISTS idx_wiki_agent_instances_root
ON wiki_agent_instances(root_folder_id);

CREATE INDEX IF NOT EXISTS idx_wiki_agent_instances_user_status
ON wiki_agent_instances(user_id, status);

COMMENT ON TABLE wiki_agent_instances IS 'Wiki 学习 Agent 实例表';
COMMENT ON COLUMN wiki_agent_instances.user_id IS '实例所属用户';
COMMENT ON COLUMN wiki_agent_instances.root_folder_id IS '绑定的知识库根目录';
COMMENT ON COLUMN wiki_agent_instances.agent_key IS 'Agent 类型标识';
COMMENT ON COLUMN wiki_agent_instances.name IS '实例展示名称';
COMMENT ON COLUMN wiki_agent_instances.description IS '实例说明';
COMMENT ON COLUMN wiki_agent_instances.status IS '实例状态';
