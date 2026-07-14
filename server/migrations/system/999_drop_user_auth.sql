-- ============================================
-- 清理 sys_users 表与 admin / user1 / user2 种子数据。
-- 兜底清理 wiki_folder_permissions 幻表(全仓库无 CREATE TABLE,
-- 仅历史 SQL 注释里被引用)。
--
-- 顺序约束:必须在 006_drop_user_id.sql 之后执行。
-- 幂等(IF EXISTS),可重复跑。
-- ============================================

DO $$
BEGIN
    IF to_regclass('public.sys_users') IS NOT NULL THEN
        EXECUTE 'TRUNCATE TABLE sys_users';
        EXECUTE 'DROP TABLE sys_users CASCADE';
    END IF;
END$$;

-- 防御性:任何残留同名表(包括历史部署可能手工创建过的)都清掉。
DROP TABLE IF EXISTS wiki_folder_permissions CASCADE;
