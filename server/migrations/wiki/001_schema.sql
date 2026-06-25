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
CREATE INDEX idx_folders_user_id ON wiki_folders(user_id);
CREATE INDEX idx_folders_created_by ON wiki_folders(created_by);
CREATE INDEX idx_folders_updated_by ON wiki_folders(updated_by);

-- 2. 文件夹权限表 (wiki_folder_permissions)
CREATE TABLE wiki_folder_permissions (
    id VARCHAR(32) PRIMARY KEY,
    -- 文件夹 ID
    folder_id VARCHAR(32) NOT NULL REFERENCES wiki_folders(id) ON DELETE CASCADE,
    -- 被授权的用户 ID（与 department_id 二选一）
    user_id VARCHAR(32) REFERENCES sys_users(id) ON DELETE CASCADE,
    -- 被授权的部门 ID（与 user_id 二选一）
    department_id VARCHAR(32) REFERENCES sys_departments(id) ON DELETE CASCADE,
    -- 权限类型: manage, edit, read
    permission_type VARCHAR(20) NOT NULL DEFAULT 'read',
    -- 时间戳
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_folder_permission_target CHECK (user_id IS NOT NULL OR department_id IS NOT NULL),
    CONSTRAINT chk_folder_permission_type CHECK (permission_type IN ('manage', 'edit', 'read')),
    CONSTRAINT uq_folder_permission_user UNIQUE (folder_id, user_id),
    CONSTRAINT uq_folder_permission_department UNIQUE (folder_id, department_id)
);

CREATE INDEX idx_folder_permissions_folder_id ON wiki_folder_permissions(folder_id);
CREATE INDEX idx_folder_permissions_user_id ON wiki_folder_permissions(user_id);
CREATE INDEX idx_folder_permissions_department_id ON wiki_folder_permissions(department_id);
CREATE INDEX idx_folder_permissions_permission_type ON wiki_folder_permissions(permission_type);

-- 3. 文档表 (wiki_documents)
CREATE TABLE wiki_documents (
    id VARCHAR(32) PRIMARY KEY,
    -- 文件名
    filename VARCHAR(255) NOT NULL,
    -- 文件大小
    file_size BIGINT NOT NULL,
    -- 内容类型
    content_type VARCHAR(100) DEFAULT 'text/markdown',
    -- MinIO 中的文件路径
    file_path VARCHAR(500),
    -- 所属文件夹
    folder_id VARCHAR(32) NOT NULL REFERENCES wiki_folders(id) ON DELETE CASCADE,
    -- 上传者
    uploaded_by VARCHAR(32) REFERENCES sys_users(id) ON DELETE SET NULL,
    -- 处理状态: pending, processing, completed, failed
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    -- 文档被分割成的块数
    chunk_count INTEGER DEFAULT 0,
    -- 错误信息（处理失败时记录）
    error_message TEXT,
    -- 时间戳
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_documents_status CHECK (status IN ('pending', 'processing', 'completed', 'failed'))
);

CREATE INDEX idx_documents_status ON wiki_documents(status);
CREATE INDEX idx_documents_created_at ON wiki_documents(created_at DESC);
CREATE INDEX idx_documents_folder_id ON wiki_documents(folder_id);
CREATE INDEX idx_documents_uploaded_by ON wiki_documents(uploaded_by);

-- ============================================
-- 表注释
-- ============================================
COMMENT ON TABLE wiki_folders IS '文件夹表，支持树形结构';
COMMENT ON TABLE wiki_folder_permissions IS '文件夹权限表，存储用户或部门对文件夹的访问权限';
COMMENT ON TABLE wiki_documents IS '文档元数据表';

COMMENT ON COLUMN wiki_folders.user_id IS '文件夹所属用户';
COMMENT ON COLUMN wiki_folders.parent_id IS '父文件夹ID，NULL表示根目录';
COMMENT ON COLUMN wiki_folders.updated_by IS '最后更新者';
COMMENT ON COLUMN wiki_folder_permissions.folder_id IS '文件夹ID';
COMMENT ON COLUMN wiki_folder_permissions.user_id IS '被授权的用户ID';
COMMENT ON COLUMN wiki_folder_permissions.department_id IS '被授权的部门ID';
COMMENT ON COLUMN wiki_folder_permissions.permission_type IS '权限类型: manage(可管理), edit(可编辑), read(可阅读)';
