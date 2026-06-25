-- ============================================
-- System 模块数据库架构
-- ============================================

-- 1. 部门表 (sys_departments)
CREATE TABLE sys_departments (
    id VARCHAR(32) PRIMARY KEY,
    -- 部门名称
    name VARCHAR(100) NOT NULL,
    -- 部门描述
    description TEXT,
    -- 父部门 ID（支持部门层级结构）
    parent_id VARCHAR(32) REFERENCES sys_departments(id) ON DELETE SET NULL,
    -- 时间戳
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_sys_departments_parent_id ON sys_departments(parent_id);
CREATE INDEX idx_sys_departments_name ON sys_departments(name);

-- 2. 角色表 (sys_roles)
CREATE TABLE sys_roles (
    id VARCHAR(32) PRIMARY KEY,
    -- 角色名称
    name VARCHAR(50) NOT NULL UNIQUE,
    -- 角色描述
    description TEXT,
    -- 时间戳
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_sys_roles_name ON sys_roles(name);

-- 3. 用户表 (sys_users)
CREATE TABLE sys_users (
    id VARCHAR(32) PRIMARY KEY,
    -- 用户名
    username VARCHAR(50) NOT NULL UNIQUE,
    -- 邮箱
    email VARCHAR(100) NOT NULL UNIQUE,
    -- 密码哈希
    password VARCHAR(255) NOT NULL,
    -- 用户姓名
    full_name VARCHAR(100),
    -- 用户头像 URL
    avatar VARCHAR(500),
    -- 用户状态: active, inactive, suspended
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    -- 时间戳
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_sys_users_status CHECK (status IN ('active', 'inactive', 'suspended'))
);

CREATE INDEX idx_sys_users_username ON sys_users(username);
CREATE INDEX idx_sys_users_email ON sys_users(email);
CREATE INDEX idx_sys_users_status ON sys_users(status);

-- 4. 用户部门关联表 (sys_user_departments)
CREATE TABLE sys_user_departments (
    id VARCHAR(32) PRIMARY KEY,
    user_id VARCHAR(32) NOT NULL REFERENCES sys_users(id) ON DELETE CASCADE,
    department_id VARCHAR(32) NOT NULL REFERENCES sys_departments(id) ON DELETE CASCADE,
    is_primary BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_user_department UNIQUE (user_id, department_id)
);

CREATE INDEX idx_sys_user_departments_user_id ON sys_user_departments(user_id);
CREATE INDEX idx_sys_user_departments_department_id ON sys_user_departments(department_id);
CREATE INDEX idx_sys_user_departments_is_primary ON sys_user_departments(is_primary) WHERE is_primary = true;

-- 5. 用户角色关联表 (sys_user_roles)
CREATE TABLE sys_user_roles (
    id VARCHAR(32) PRIMARY KEY,
    user_id VARCHAR(32) NOT NULL REFERENCES sys_users(id) ON DELETE CASCADE,
    role_id VARCHAR(32) NOT NULL REFERENCES sys_roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_user_role UNIQUE (user_id, role_id)
);

CREATE INDEX idx_sys_user_roles_user_id ON sys_user_roles(user_id);
CREATE INDEX idx_sys_user_roles_role_id ON sys_user_roles(role_id);

-- ============================================
-- 表注释
-- ============================================
COMMENT ON TABLE sys_departments IS '部门表';
COMMENT ON TABLE sys_roles IS '角色表';
COMMENT ON TABLE sys_users IS '用户表';
COMMENT ON TABLE sys_user_departments IS '用户部门关联表';
COMMENT ON TABLE sys_user_roles IS '用户角色关联表';

-- ============================================
-- 初始化数据
-- ============================================

-- 初始化角色
INSERT INTO sys_roles (id, name, description, created_at, updated_at) VALUES
    ('r000000000000000000000000000001', 'admin', '系统管理员，拥有所有权限', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('r000000000000000000000000000002', 'user', '普通用户，可以使用基本功能', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('r000000000000000000000000000003', 'manager', '部门管理员，可以管理本部门资源', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- 初始化部门
INSERT INTO sys_departments (id, name, description, parent_id, created_at, updated_at) VALUES
    ('d000000000000000000000000000001', '技术部', '负责技术研发和系统维护', NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('d000000000000000000000000000002', '产品部', '负责产品设计和需求管理', NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('d000000000000000000000000000003', '运营部', '负责运营和市场推广', NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- 初始化用户（密码为 '123456' 的 bcrypt hash）
INSERT INTO sys_users (id, username, email, password, full_name, status, created_at, updated_at) VALUES
    ('u000000000000000000000000000001', 'admin', 'admin@example.com',
    '$2a$10$VDDBXqMXSNoLSi5Jg6Ltm.1zIMhKTB1opt31CQM6F5m1AC7269a6K', '系统管理员', 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('u000000000000000000000000000002', 'user1', 'user1@example.com', '$2a$10$VDDBXqMXSNoLSi5Jg6Ltm.1zIMhKTB1opt31CQM6F5m1AC7269a6K', '张三', 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('u000000000000000000000000000003', 'user2', 'user2@example.com', '$2a$10$VDDBXqMXSNoLSi5Jg6Ltm.1zIMhKTB1opt31CQM6F5m1AC7269a6K', '李四', 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- 分配用户角色
INSERT INTO sys_user_roles (id, user_id, role_id, created_at) VALUES
    ('ur00000000000000000000000000001', 'u000000000000000000000000000001', 'r000000000000000000000000000001', CURRENT_TIMESTAMP),
    ('ur00000000000000000000000000002', 'u000000000000000000000000000002', 'r000000000000000000000000000002', CURRENT_TIMESTAMP),
    ('ur00000000000000000000000000003', 'u000000000000000000000000000003', 'r000000000000000000000000000002', CURRENT_TIMESTAMP);

-- 分配用户部门
INSERT INTO sys_user_departments (id, user_id, department_id, is_primary, created_at) VALUES
    ('ud0000000000000000000000000001', 'u000000000000000000000000000001', 'd000000000000000000000000000001', true, CURRENT_TIMESTAMP),
    ('ud0000000000000000000000000002', 'u000000000000000000000000000002', 'd000000000000000000000000000001', true, CURRENT_TIMESTAMP),
    ('ud0000000000000000000000000003', 'u000000000000000000000000000003', 'd000000000000000000000000000002', true, CURRENT_TIMESTAMP);
