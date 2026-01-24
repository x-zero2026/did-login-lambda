-- 查看当前测试数据
-- 在清理前运行此脚本，查看数据库中有哪些数据

-- ============================================
-- 查看所有表的数据统计
-- ============================================

SELECT 
    'users' as table_name, 
    COUNT(*) as total_count,
    MIN(created_at) as oldest_record,
    MAX(created_at) as newest_record
FROM users
UNION ALL
SELECT 
    'projects', 
    COUNT(*),
    MIN(created_at),
    MAX(created_at)
FROM projects
UNION ALL
SELECT 
    'user_projects', 
    COUNT(*),
    MIN(joined_at),
    MAX(joined_at)
FROM user_projects
UNION ALL
SELECT 
    'apps', 
    COUNT(*),
    MIN(created_at),
    MAX(created_at)
FROM apps
UNION ALL
SELECT 
    'app_projects', 
    COUNT(*),
    MIN(added_at),
    MAX(added_at)
FROM app_projects
UNION ALL
SELECT 
    'project_app_settings', 
    COUNT(*),
    MIN(updated_at),
    MAX(updated_at)
FROM project_app_settings;

-- ============================================
-- 查看所有用户
-- ============================================

SELECT 
    username,
    email,
    LEFT(did, 20) || '...' as did_preview,
    created_at
FROM users
ORDER BY created_at DESC;

-- ============================================
-- 查看所有项目
-- ============================================

SELECT 
    p.project_name,
    u.username as creator,
    p.created_at,
    (SELECT COUNT(*) FROM user_projects WHERE project_id = p.project_id) as member_count,
    (SELECT COUNT(*) FROM app_projects WHERE project_id = p.project_id) as app_count
FROM projects p
JOIN users u ON p.creator_did = u.did
ORDER BY p.created_at DESC;

-- ============================================
-- 查看所有应用
-- ============================================

SELECT 
    a.app_name,
    a.emoji,
    a.is_global,
    u.username as creator,
    a.created_at,
    (SELECT COUNT(*) FROM app_projects WHERE app_id = a.app_id) as project_count
FROM apps a
JOIN users u ON a.created_by_did = u.did
ORDER BY a.created_at DESC;

-- ============================================
-- 查看用户-项目关系
-- ============================================

SELECT 
    u.username,
    p.project_name,
    up.role,
    up.joined_at
FROM user_projects up
JOIN users u ON up.user_did = u.did
JOIN projects p ON up.project_id = p.project_id
ORDER BY up.joined_at DESC;
