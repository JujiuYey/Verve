-- Agent/场景模型使用配置：解决“什么场景用哪个模型”的问题。
CREATE TABLE IF NOT EXISTS sys_agent_model_configs (
  id VARCHAR(32) PRIMARY KEY,
  agent_key VARCHAR(64) NOT NULL,
  scene_key VARCHAR(64) NOT NULL,
  model_id VARCHAR(32) NOT NULL REFERENCES sys_models(id) ON DELETE CASCADE,
  params JSONB NOT NULL DEFAULT '{}'::jsonb,
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT uk_sys_agent_model_configs_agent_scene UNIQUE (agent_key, scene_key)
);

CREATE INDEX IF NOT EXISTS idx_sys_agent_model_configs_model
ON sys_agent_model_configs(model_id);

COMMENT ON TABLE sys_agent_model_configs IS 'Agent 场景模型使用配置';
COMMENT ON COLUMN sys_agent_model_configs.agent_key IS 'Agent 标识，如 coach/wiki_rag';
COMMENT ON COLUMN sys_agent_model_configs.scene_key IS '场景标识，如 chat/embedding/rerank';
COMMENT ON COLUMN sys_agent_model_configs.model_id IS '该场景绑定的模型 ID';
COMMENT ON COLUMN sys_agent_model_configs.params IS '该场景调用参数';
