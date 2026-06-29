-- ============================================
-- 系统模型平台与模型配置
-- ============================================

-- 1. AI 模型平台表
CREATE TABLE IF NOT EXISTS sys_model_platforms (
  id VARCHAR(32) PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  provider_type VARCHAR(50) NOT NULL DEFAULT 'openai_compatible',
  default_base_url VARCHAR(500) NOT NULL DEFAULT '',
  base_url VARCHAR(500) NOT NULL DEFAULT '',
  api_key_ciphertext TEXT,
  api_key_hint VARCHAR(100),
  extra_headers JSONB,
  model_list_path VARCHAR(100) NOT NULL DEFAULT '/models',
  auth_scheme VARCHAR(30) NOT NULL DEFAULT 'bearer',
  docs_url VARCHAR(500),
  api_key_url VARCHAR(500),
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  sort_order INTEGER NOT NULL DEFAULT 0,
  last_model_sync_at TIMESTAMP,
  metadata JSONB,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT uk_sys_model_platforms_name UNIQUE (name),
  CONSTRAINT chk_sys_model_platforms_provider_type CHECK (provider_type IN ('openai_compatible', 'custom')),
  CONSTRAINT chk_sys_model_platforms_auth_scheme CHECK (auth_scheme IN ('bearer', 'x_api_key', 'both'))
);

CREATE INDEX IF NOT EXISTS idx_sys_model_platforms_enabled_sort
ON sys_model_platforms(enabled, sort_order);

-- 2. AI 模型表
CREATE TABLE IF NOT EXISTS sys_models (
  id VARCHAR(32) PRIMARY KEY,
  platform_id VARCHAR(32) NOT NULL REFERENCES sys_model_platforms(id) ON DELETE CASCADE,
  model_name VARCHAR(200) NOT NULL,
  display_name VARCHAR(200) NOT NULL,
  model_type VARCHAR(30) NOT NULL DEFAULT 'chat',
  capabilities JSONB,
  source VARCHAR(20) NOT NULL DEFAULT 'remote',
  status VARCHAR(20) NOT NULL DEFAULT 'active',
  is_default BOOLEAN NOT NULL DEFAULT FALSE,
  temperature DECIMAL(3,2) NOT NULL DEFAULT 0.7,
  top_p DECIMAL(3,2) NOT NULL DEFAULT 0.9,
  max_tokens INTEGER,
  top_k INTEGER,
  last_synced_at TIMESTAMP,
  metadata JSONB,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT uk_sys_models_platform_model_name UNIQUE (platform_id, model_name),
  CONSTRAINT chk_sys_models_model_type CHECK (model_type IN ('chat', 'embedding', 'rerank')),
  CONSTRAINT chk_sys_models_source CHECK (source IN ('remote', 'manual')),
  CONSTRAINT chk_sys_models_status CHECK (status IN ('active', 'inactive'))
);

CREATE INDEX IF NOT EXISTS idx_sys_models_platform_type_status
ON sys_models(platform_id, model_type, status);

CREATE UNIQUE INDEX IF NOT EXISTS idx_sys_models_type_default
ON sys_models(model_type)
WHERE is_default = TRUE;

-- ============================================
-- 表注释
-- ============================================

COMMENT ON TABLE sys_model_platforms IS '系统模型平台表';
COMMENT ON TABLE sys_models IS '系统模型表';

COMMENT ON COLUMN sys_model_platforms.name IS '平台展示名称';
COMMENT ON COLUMN sys_model_platforms.provider_type IS '平台类型：openai_compatible/custom';
COMMENT ON COLUMN sys_model_platforms.default_base_url IS '平台默认 API 地址';
COMMENT ON COLUMN sys_model_platforms.base_url IS '实际使用的 API 地址';
COMMENT ON COLUMN sys_model_platforms.api_key_ciphertext IS '加密后的 API Key';
COMMENT ON COLUMN sys_model_platforms.api_key_hint IS '脱敏展示用 Key 提示';
COMMENT ON COLUMN sys_model_platforms.extra_headers IS '附加请求头';
COMMENT ON COLUMN sys_model_platforms.model_list_path IS '拉取模型列表路径';
COMMENT ON COLUMN sys_model_platforms.auth_scheme IS '认证方式：bearer/x_api_key/both';
COMMENT ON COLUMN sys_model_platforms.docs_url IS '平台文档地址';
COMMENT ON COLUMN sys_model_platforms.api_key_url IS 'API Key 管理页面地址';
COMMENT ON COLUMN sys_model_platforms.enabled IS '是否启用平台';
COMMENT ON COLUMN sys_model_platforms.sort_order IS '展示排序';
COMMENT ON COLUMN sys_model_platforms.last_model_sync_at IS '最近同步模型列表时间';
COMMENT ON COLUMN sys_model_platforms.metadata IS '平台扩展信息';

COMMENT ON COLUMN sys_models.platform_id IS '所属模型平台 ID';
COMMENT ON COLUMN sys_models.model_name IS 'API 调用模型名称';
COMMENT ON COLUMN sys_models.display_name IS '前端展示名称';
COMMENT ON COLUMN sys_models.model_type IS '模型类型：chat/embedding/rerank';
COMMENT ON COLUMN sys_models.capabilities IS '模型能力标签';
COMMENT ON COLUMN sys_models.source IS '模型来源：remote/manual';
COMMENT ON COLUMN sys_models.status IS '模型状态：active/inactive';
COMMENT ON COLUMN sys_models.is_default IS '是否为该模型类型的默认模型';
COMMENT ON COLUMN sys_models.temperature IS '随机性/创造性，chat 模型使用';
COMMENT ON COLUMN sys_models.top_p IS '核采样，chat 模型使用';
COMMENT ON COLUMN sys_models.max_tokens IS '单次响应最大 token 数，chat 模型使用';
COMMENT ON COLUMN sys_models.top_k IS '采样时考虑的最高概率 token 数，chat 模型使用';
COMMENT ON COLUMN sys_models.last_synced_at IS '最近同步时间';
COMMENT ON COLUMN sys_models.metadata IS '模型扩展信息';
