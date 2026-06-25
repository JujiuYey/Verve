-- ============================================
-- AI 模块数据库架构
-- 依赖: system/001_schema.sql, wiki/001_schema.sql
-- ============================================

-- 1. 模型配置表 (model_config)
CREATE TABLE IF NOT EXISTS ai_model_config (
  id VARCHAR(32) PRIMARY KEY,
  -- 模型厂商/服务商，例如 OpenAI、Anthropic、硅基流动等
  vendor VARCHAR(32) NOT NULL, 
  -- 配置名称
  name VARCHAR(32) NOT NULL UNIQUE,
  -- API Key
  api_key VARCHAR(256) NOT NULL,
  -- 服务基地址 
  base_url VARCHAR(256) NOT NULL, 
  -- 模型类型：chat-对话模型, embedding-向量模型
  model_type VARCHAR(16) NOT NULL DEFAULT 'chat',
  -- 模型名称，例如 gpt-4o-mini、text-embedding-3-small
  model VARCHAR(128) NOT NULL, 
  -- 随机性/创造性（chat 模型使用）
  temperature DECIMAL(3,2) NOT NULL DEFAULT 0.7, 
  -- 核采样（chat 模型使用）
  top_p DECIMAL(3,2) NOT NULL DEFAULT 0.9, 
  -- 单次响应最大 token数（chat 模型使用）
  max_tokens INTEGER,  
  -- 采样时考虑的最高概率 token 数（chat 模型使用）
  top_k INTEGER,  
  -- 是否启用
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  -- 是否为默认配置
  is_default BOOLEAN NOT NULL DEFAULT FALSE,
  -- 创建时间
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, 
  -- 更新时间
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP 
);

-- 按启用状态查询时使用
CREATE INDEX IF NOT EXISTS idx_ai_model_config_active
ON ai_model_config(is_active);

-- 保证同一时间只有一个一种类型的默认模型配置
CREATE UNIQUE INDEX idx_ai_model_config_type_default                                      
ON ai_model_config(model_type, is_default) WHERE is_default = TRUE;

-- ============================================
-- 表注释
-- ============================================
COMMENT ON TABLE ai_model_config IS '模型配置表';

COMMENT ON COLUMN ai_model_config.name IS '名称';
COMMENT ON COLUMN ai_model_config.vendor IS '模型厂商/服务';
COMMENT ON COLUMN ai_model_config.api_key IS 'API Key，敏感信息，需加密存储';
COMMENT ON COLUMN ai_model_config.base_url IS '服务基地址，例如 https://api.openai.com/v1';
COMMENT ON COLUMN ai_model_config.model_type IS '模型类型，chat-对话模型, embedding-向量模型';
COMMENT ON COLUMN ai_model_config.model IS '模型名称，例如 gpt-4o-mini、text-embedding-3-small';
COMMENT ON COLUMN ai_model_config.temperature IS '随机性/创造性，值越高输出越随机，通常在 0.0 到 1.0 之间（chat 模型使用）';
COMMENT ON COLUMN ai_model_config.top_p IS '核采样，值越高输出越随机，通常在 0.0 到 1.0 之间（chat 模型使用）';
COMMENT ON COLUMN ai_model_config.max_tokens IS '单次响应最大 token数（chat 模型使用）';
COMMENT ON COLUMN ai_model_config.top_k IS '采样时考虑的最高概率 token 数（chat 模型使用）';
COMMENT ON COLUMN ai_model_config.is_active IS '是否启用，1 表示启用，0 表示禁用';
COMMENT ON COLUMN ai_model_config.is_default IS '是否为默认配置，1 表示是，0 表示否';
