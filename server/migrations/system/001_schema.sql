-- ============================================
-- System 模块数据库架构
-- ============================================

-- 1. 用户表 (sys_users)
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

-- ============================================
-- 表注释
-- ============================================
COMMENT ON TABLE sys_users IS '用户表';

-- ============================================
-- 初始化数据
-- ============================================

-- 初始化用户（密码为 '123456' 的 bcrypt hash）
INSERT INTO sys_users (id, username, email, password, full_name, status, created_at, updated_at) VALUES
    ('u000000000000000000000000000001', 'admin', 'admin@example.com',
    '$2a$10$VDDBXqMXSNoLSi5Jg6Ltm.1zIMhKTB1opt31CQM6F5m1AC7269a6K', '系统管理员', 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('u000000000000000000000000000002', 'user1', 'user1@example.com', '$2a$10$VDDBXqMXSNoLSi5Jg6Ltm.1zIMhKTB1opt31CQM6F5m1AC7269a6K', '张三', 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('u000000000000000000000000000003', 'user2', 'user2@example.com', '$2a$10$VDDBXqMXSNoLSi5Jg6Ltm.1zIMhKTB1opt31CQM6F5m1AC7269a6K', '李四', 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);