-- 🚀 快速清理脚本
-- 一键清理所有测试数据
-- 
-- ⚠️ 警告：此操作不可逆！
-- 
-- 使用方法：
-- 1. 在 Supabase SQL Editor 中打开此文件
-- 2. 点击 "Run" 执行
-- 3. 查看执行结果

-- ============================================
-- 快速清理所有数据
-- ============================================

DO $
DECLARE
    v_users_count INT;
    v_projects_count INT;
    v_apps_count INT;
BEGIN
    -- 记录清理前的数据量
    SELECT COUNT(*) INTO v_users_count FROM users;
    SELECT COUNT(*) INTO v_projects_count FROM projects;
    SELECT COUNT(*) INTO v_apps_count FROM apps;
    
    RAISE NOTICE '========================================';
    RAISE NOTICE '开始清理测试数据';
    RAISE NOTICE '========================================';
    RAISE NOTICE '清理前统计:';
    RAISE NOTICE '  用户数: %', v_users_count;
    RAISE NOTICE '  项目数: %', v_projects_count;
    RAISE NOTICE '  应用数: %', v_apps_count;
    RAISE NOTICE '========================================';
    
    -- 按顺序删除数据
    DELETE FROM project_app_settings;
    RAISE NOTICE '✓ 已清理 project_app_settings';
    
    DELETE FROM app_projects;
    RAISE NOTICE '✓ 已清理 app_projects';
    
    DELETE FROM user_projects;
    RAISE NOTICE '✓ 已清理 user_projects';
    
    DELETE FROM apps;
    RAISE NOTICE '✓ 已清理 apps';
    
    DELETE FROM projects;
    RAISE NOTICE '✓ 已清理 projects';
    
    DELETE FROM users;
    RAISE NOTICE '✓ 已清理 users';
    
    RAISE NOTICE '========================================';
    RAISE NOTICE '清理完成！';
    RAISE NOTICE '已删除:';
    RAISE NOTICE '  用户: % 个', v_users_count;
    RAISE NOTICE '  项目: % 个', v_projects_count;
    RAISE NOTICE '  应用: % 个', v_apps_count;
    RAISE NOTICE '========================================';
END $;

-- 验证清理结果
SELECT 
    'users' as table_name, 
    COUNT(*) as remaining_count
FROM users
UNION ALL
SELECT 'projects', COUNT(*) FROM projects
UNION ALL
SELECT 'user_projects', COUNT(*) FROM user_projects
UNION ALL
SELECT 'apps', COUNT(*) FROM apps
UNION ALL
SELECT 'app_projects', COUNT(*) FROM app_projects
UNION ALL
SELECT 'project_app_settings', COUNT(*) FROM project_app_settings
ORDER BY table_name;

-- 应该显示所有表的 remaining_count 都是 0
