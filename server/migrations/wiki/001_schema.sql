-- ============================================
-- Wiki 模块数据库架构
-- 依赖: system/001_schema.sql
-- ============================================

-- 1. 文件夹表 (wiki_folders)
CREATE TABLE wiki_folders (
    id VARCHAR(32) PRIMARY KEY,
    -- 文件夹名称
    name VARCHAR(100) NOT NULL,
    -- 文件夹描述
    description TEXT,
    -- 父文件夹（NULL 表示根目录）
    parent_id VARCHAR(32) REFERENCES wiki_folders(id) ON DELETE CASCADE,
    -- 所属用户
    user_id VARCHAR(32) REFERENCES sys_users(id) ON DELETE SET NULL,
    -- 排序权重（同级内按升序展示）
    sort_order INTEGER NOT NULL DEFAULT 0,
    -- 创建者
    created_by VARCHAR(32) REFERENCES sys_users(id) ON DELETE SET NULL,
    -- 最后更新者
    updated_by VARCHAR(32) REFERENCES sys_users(id) ON DELETE SET NULL,
    -- 时间戳
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_folders_name ON wiki_folders(name);
CREATE INDEX idx_folders_parent_id ON wiki_folders(parent_id);
CREATE INDEX idx_folders_parent_sort ON wiki_folders(parent_id, sort_order);
CREATE INDEX idx_folders_user_id ON wiki_folders(user_id);
CREATE INDEX idx_folders_created_by ON wiki_folders(created_by);
CREATE INDEX idx_folders_updated_by ON wiki_folders(updated_by);

-- 2. 文档表 (wiki_documents)
CREATE TABLE wiki_documents (
    id VARCHAR(32) PRIMARY KEY,
    -- 文件名
    filename VARCHAR(255) NOT NULL,
    -- 文件大小
    file_size BIGINT NOT NULL,
    -- 内容类型
    content_type VARCHAR(100) DEFAULT 'text/markdown',
    -- MinIO 中的文件路径
    file_path VARCHAR(500) NOT NULL,
    -- 所属文件夹
    folder_id VARCHAR(32) NOT NULL REFERENCES wiki_folders(id) ON DELETE CASCADE,
    -- 排序权重（同一文件夹内按升序展示）
    sort_order INTEGER NOT NULL DEFAULT 0,
    -- 时间戳
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_documents_created_at ON wiki_documents(created_at DESC);
CREATE INDEX idx_documents_folder_id ON wiki_documents(folder_id);
CREATE INDEX idx_documents_folder_sort ON wiki_documents(folder_id, sort_order);

-- ============================================
-- 表注释
-- ============================================
COMMENT ON TABLE wiki_folders IS '文件夹表，支持树形结构';
COMMENT ON TABLE wiki_documents IS '文档元数据表';

COMMENT ON COLUMN wiki_folders.user_id IS '文件夹所属用户';
COMMENT ON COLUMN wiki_folders.parent_id IS '父文件夹ID，NULL表示根目录';
COMMENT ON COLUMN wiki_folders.sort_order IS '同级文件夹排序权重，升序展示';
COMMENT ON COLUMN wiki_folders.updated_by IS '最后更新者';
COMMENT ON COLUMN wiki_documents.sort_order IS '同一文件夹内文档排序权重，升序展示';
COMMENT ON COLUMN wiki_folder_permissions.folder_id IS '文件夹ID';
COMMENT ON COLUMN wiki_folder_permissions.user_id IS '被授权的用户ID';
COMMENT ON COLUMN wiki_folder_permissions.department_id IS '被授权的部门ID';
COMMENT ON COLUMN wiki_folder_permissions.permission_type IS '权限类型: manage(可管理), edit(可编辑), read(可阅读)';
