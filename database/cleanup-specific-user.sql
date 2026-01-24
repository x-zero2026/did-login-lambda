-- 清理特定用户的数据
-- 
-- 使用方法：
-- 1. 替换下面的 'YOUR_USERNAME' 或 'YOUR_EMAIL' 为实际值
-- 2. 在 Supabase SQL Editor 中执行
-- 
-- 示例：
-- WHERE username = 'testfix'
-- WHERE email = 'test@example.com'

-- ============================================
-- 方案 2: 清理特定用户及其相关数据
-- ============================================

BEGIN;

-- 查找要删除的用户 DID
DO $
DECLARE
    target_did VARCHAR(66);
BEGIN
    -- 方式 1: 通过用户名查找
    SELECT did INTO target_did FROM users WHERE username = 'testfix';
    
    -- 方式 2: 通过邮箱查找（注释掉上面一行，使用这行）
    -- SELECT did INTO target_did FROM users WHERE email = 'test@example.com';
    
    IF target_did IS NOT NULL THEN
        RAISE NOTICE '找到用户 DID: %', target_did;
        
        -- 1. 删除该用户的项目应用设置
        DELETE FROM project_app_settings 
        WHERE project_id IN (
            SELECT project_id FROM projects WHERE creator_did = target_did
        );
        RAISE NOTICE '已删除项目应用设置';
        
        -- 2. 删除该用户的应用-项目关联
        DELETE FROM app_projects 
        WHERE project_id IN (
            SELECT project_id FROM projects WHERE creator_did = target_did
        );
        RAISE NOTICE '已删除应用-项目关联';
        
        -- 3. 删除该用户创建的应用
        DELETE FROM apps WHERE created_by_did = target_did;
        RAISE NOTICE '已删除用户创建的应用';
        
        -- 4. 删除该用户的项目关联
        DELETE FROM user_projects WHERE user_did = target_did;
        RAISE NOTICE '已删除用户-项目关联';
        
        -- 5. 删除该用户创建的项目
        DELETE FROM projects WHERE creator_did = target_did;
        RAISE NOTICE '已删除用户创建的项目';
        
        -- 6. 删除用户
        DELETE FROM users WHERE did = target_did;
        RAISE NOTICE '已删除用户';
        
        RAISE NOTICE '用户 % 及其所有相关数据已清理完成', target_did;
    ELSE
        RAISE NOTICE '未找到指定用户';
    END IF;
END $;

COMMIT;
