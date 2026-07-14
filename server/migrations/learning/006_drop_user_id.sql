-- ============================================
-- 清理所有 user_id / created_by / updated_by / requested_by / changed_by 列,
-- 以及依赖的外键和索引。
-- 幂等(IF EXISTS),可重复跑。
--
-- 顺序约束:必须先于 999_drop_user_auth.sql 执行,否则 sys_users 上的
-- 外键会让 DROP 失败。
-- ============================================

-- 1. wiki_folders: user_id, created_by, updated_by
ALTER TABLE wiki_folders
    DROP CONSTRAINT IF EXISTS wiki_folders_user_id_fkey,
    DROP CONSTRAINT IF EXISTS wiki_folders_created_by_fkey,
    DROP CONSTRAINT IF EXISTS wiki_folders_updated_by_fkey;

DROP INDEX IF EXISTS idx_folders_user_id;
DROP INDEX IF EXISTS idx_folders_created_by;
DROP INDEX IF EXISTS idx_folders_updated_by;

ALTER TABLE wiki_folders
    DROP COLUMN IF EXISTS user_id,
    DROP COLUMN IF EXISTS created_by,
    DROP COLUMN IF EXISTS updated_by;

-- 2. wiki_document_revisions: changed_by
ALTER TABLE wiki_document_revisions
    DROP CONSTRAINT IF EXISTS wiki_document_revisions_changed_by_fkey;

ALTER TABLE wiki_document_revisions
    DROP COLUMN IF EXISTS changed_by;

-- 3. wiki_document_change_requests: requested_by
ALTER TABLE wiki_document_change_requests
    DROP CONSTRAINT IF EXISTS wiki_document_change_requests_requested_by_fkey;

DROP INDEX IF EXISTS idx_wiki_change_requests_requested_status;

ALTER TABLE wiki_document_change_requests
    DROP COLUMN IF EXISTS requested_by;

-- 4. learning_sessions: user_id
ALTER TABLE learning_sessions
    DROP CONSTRAINT IF EXISTS learning_sessions_user_id_fkey;

DROP INDEX IF EXISTS idx_learning_sessions_user_id;

ALTER TABLE learning_sessions
    DROP COLUMN IF EXISTS user_id;

-- 5. learning_memory_events: user_id
ALTER TABLE learning_memory_events
    DROP CONSTRAINT IF EXISTS learning_memory_events_user_id_fkey;

DROP INDEX IF EXISTS idx_learning_memory_events_user_time;

ALTER TABLE learning_memory_events
    DROP COLUMN IF EXISTS user_id;

-- 6. learning_memory_items: user_id
ALTER TABLE learning_memory_items
    DROP CONSTRAINT IF EXISTS learning_memory_items_user_id_fkey;

DROP INDEX IF EXISTS idx_learning_memory_items_user_kind;

ALTER TABLE learning_memory_items
    DROP COLUMN IF EXISTS user_id;

-- 7. learning_memory_summaries: user_id + 重建 folder 唯一索引
DROP INDEX IF EXISTS uq_learning_memory_summaries_user_folder;
DROP INDEX IF EXISTS uq_learning_memory_summaries_user_global;

ALTER TABLE learning_memory_summaries
    DROP CONSTRAINT IF EXISTS learning_memory_summaries_user_id_fkey;

DROP INDEX IF EXISTS idx_learning_memory_summaries_user_id;

ALTER TABLE learning_memory_summaries
    DROP COLUMN IF EXISTS user_id;

-- 替代原 (user_id, folder_id) WHERE folder_id IS NOT NULL:
-- 单租户下每 folder 一份 memory summary。
CREATE UNIQUE INDEX IF NOT EXISTS uq_learning_memory_summaries_folder
    ON learning_memory_summaries(folder_id)
    WHERE folder_id IS NOT NULL;

-- 替代原 (user_id) WHERE folder_id IS NULL:
-- 用常量表达式保证 NULL folder_id 的"全局汇总"最多只有一份。
CREATE UNIQUE INDEX IF NOT EXISTS uq_learning_memory_summaries_global
    ON learning_memory_summaries((1))
    WHERE folder_id IS NULL;
