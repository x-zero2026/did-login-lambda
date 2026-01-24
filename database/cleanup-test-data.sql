-- ⚠️ 清理测试数据脚本
-- 警告：此脚本会删除所有数据！请谨慎使用！
-- 
-- 使用方法：
-- 1. 登录 Supabase Dashboard
-- 2. 进入 SQL Editor
-- 3. 复制并执行此脚本
-- 
-- 或者使用 psql 命令行：
-- psql "postgresql://postgres.rbpsksuuvtzmathnmyxn:iPass4xz2026%21@aws-1-ap-south-1.pooler.supabase.com:6543/postgres" -f cleanup-test-data.sql

-- ============================================
-- 方案 1: 清理所有数据（保留表结构）
-- ============================================

-- 禁用外键约束检查（PostgreSQL 不支持，但可以按顺序删除）
-- 按照依赖关系的逆序删除数据

BEGIN;

-- 1. 删除项目应用设置（依赖 projects 和 apps）
DELETE FROM project_app_settings;

-- 2. 删除应用-项目关联（依赖 apps 和 projects）
DELETE FROM app_projects;

-- 3. 删除用户-项目关联（依赖 users 和 projects）
DELETE FROM user_projects;

-- 4. 删除应用（依赖 users）
DELETE FROM apps;

-- 5. 删除项目（依赖 users）
DELETE FROM projects;

-- 6. 删除用户（最后删除，因为其他表都依赖它）
DELETE FROM users;

COMMIT;

-- 验证清理结果
SELECT 'users' as table_name, COUNT(*) as count FROM users
UNION ALL
SELECT 'projects', COUNT(*) FROM projects
UNION ALL
SELECT 'user_projects', COUNT(*) FROM user_projects
UNION ALL
SELECT 'apps', COUNT(*) FROM apps
UNION ALL
SELECT 'app_projects', COUNT(*) FROM app_projects
UNION ALL
SELECT 'project_app_settings', COUNT(*) FROM project_app_settings;

-- 应该显示所有表的 count 都是 0
