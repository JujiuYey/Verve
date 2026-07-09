-- ============================================
-- 精简系统模型表
-- ============================================

DROP INDEX IF EXISTS idx_sys_models_type_default;
DROP INDEX IF EXISTS idx_sys_models_platform_type_status;

ALTER TABLE sys_models
  DROP CONSTRAINT IF EXISTS chk_sys_models_model_type,
  DROP CONSTRAINT IF EXISTS chk_sys_models_source;

ALTER TABLE sys_models
  DROP COLUMN IF EXISTS capabilities,
  DROP COLUMN IF EXISTS is_default,
  DROP COLUMN IF EXISTS temperature,
  DROP COLUMN IF EXISTS top_p,
  DROP COLUMN IF EXISTS max_tokens,
  DROP COLUMN IF EXISTS top_k,
  DROP COLUMN IF EXISTS metadata,
  DROP COLUMN IF EXISTS source,
  DROP COLUMN IF EXISTS model_type;

CREATE INDEX IF NOT EXISTS idx_sys_models_platform_status
ON sys_models(platform_id, status);
