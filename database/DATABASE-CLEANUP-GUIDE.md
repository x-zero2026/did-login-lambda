# 🗑️ 数据库清理指南

## 📋 清理脚本说明

### 1. view-test-data.sql
**用途**: 查看当前数据库中的所有测试数据  
**安全性**: ✅ 只读，不会修改数据

### 2. cleanup-test-data.sql
**用途**: 清理所有测试数据（保留表结构）  
**安全性**: ⚠️ 危险！会删除所有数据

### 3. cleanup-specific-user.sql
**用途**: 清理特定用户及其相关数据  
**安全性**: ⚠️ 会删除指定用户的所有数据

---

## 🚀 使用方法

### 方法 1: Supabase Dashboard（推荐）

#### 步骤 1: 登录 Supabase
1. 访问：https://supabase.com/dashboard
2. 选择项目：`rbpsksuuvtzmathnmyxn`

#### 步骤 2: 打开 SQL Editor
1. 点击左侧菜单 **"SQL Editor"**
2. 点击 **"New query"**

#### 步骤 3: 查看数据（可选）
```sql
-- 复制 view-test-data.sql 的内容
-- 粘贴到 SQL Editor
-- 点击 "Run" 或按 Ctrl+Enter
```

查看结果，确认要删除的数据。

#### 步骤 4: 执行清理
```sql
-- 复制 cleanup-test-data.sql 的内容
-- 粘贴到 SQL Editor
-- 点击 "Run" 或按 Ctrl+Enter
```

#### 步骤 5: 验证清理结果
再次运行 `view-test-data.sql`，确认所有表的 count 都是 0。

---

### 方法 2: psql 命令行

#### 前提条件
- 已安装 PostgreSQL 客户端
- 有数据库连接权限

#### 连接数据库
```bash
# 使用 Supabase Pooler 连接
psql "postgresql://postgres.rbpsksuuvtzmathnmyxn:iPass4xz2026%21@aws-1-ap-south-1.pooler.supabase.com:6543/postgres"
```

#### 查看数据
```bash
# 在 psql 中执行
\i docs/backend/database/view-test-data.sql
```

#### 清理数据
```bash
# 在 psql 中执行
\i docs/backend/database/cleanup-test-data.sql
```

---

## 📊 清理场景

### 场景 1: 清理所有测试数据

**适用情况**:
- 开发测试完成，准备上线
- 数据混乱，需要重新开始
- 定期清理测试环境

**步骤**:
1. 运行 `view-test-data.sql` 查看数据
2. 确认要删除所有数据
3. 运行 `cleanup-test-data.sql`
4. 验证清理结果

**SQL**:
```sql
-- 使用 cleanup-test-data.sql 的内容
```

---

### 场景 2: 清理特定用户

**适用情况**:
- 只删除某个测试用户
- 保留其他用户数据
- 清理错误注册的用户

**步骤**:
1. 找到要删除的用户名或邮箱
2. 编辑 `cleanup-specific-user.sql`
3. 修改用户名或邮箱
4. 运行脚本

**示例**:
```sql
-- 删除用户名为 testfix 的用户
SELECT did INTO target_did FROM users WHERE username = 'testfix';

-- 或删除邮箱为 test@example.com 的用户
SELECT did INTO target_did FROM users WHERE email = 'test@example.com';
```

---

### 场景 3: 清理特定时间段的数据

**适用情况**:
- 只删除某个时间段的测试数据
- 保留早期数据

**SQL**:
```sql
BEGIN;

-- 删除今天创建的用户及其相关数据
DELETE FROM project_app_settings 
WHERE project_id IN (
    SELECT project_id FROM projects 
    WHERE creator_did IN (
        SELECT did FROM users WHERE created_at >= CURRENT_DATE
    )
);

DELETE FROM app_projects 
WHERE project_id IN (
    SELECT project_id FROM projects 
    WHERE creator_did IN (
        SELECT did FROM users WHERE created_at >= CURRENT_DATE
    )
);

DELETE FROM apps 
WHERE created_by_did IN (
    SELECT did FROM users WHERE created_at >= CURRENT_DATE
);

DELETE FROM user_projects 
WHERE user_did IN (
    SELECT did FROM users WHERE created_at >= CURRENT_DATE
);

DELETE FROM projects 
WHERE creator_did IN (
    SELECT did FROM users WHERE created_at >= CURRENT_DATE
);

DELETE FROM users WHERE created_at >= CURRENT_DATE;

COMMIT;
```

---

## ⚠️ 安全提示

### 清理前检查

1. **备份数据**（如果需要）
   ```sql
   -- 导出用户数据
   COPY users TO '/tmp/users_backup.csv' CSV HEADER;
   ```

2. **确认环境**
   - ✅ 确认是测试环境
   - ❌ 不要在生产环境执行

3. **查看数据**
   - 先运行 `view-test-data.sql`
   - 确认要删除的数据

### 清理后验证

1. **检查表计数**
   ```sql
   SELECT COUNT(*) FROM users;
   SELECT COUNT(*) FROM projects;
   SELECT COUNT(*) FROM apps;
   ```

2. **测试注册**
   - 尝试注册新用户
   - 确认功能正常

---

## 🔄 常见问题

### Q1: 删除失败，提示外键约束错误

**原因**: 删除顺序不对，子表数据依赖父表

**解决**: 按照脚本中的顺序删除（从子表到父表）

---

### Q2: 想保留某些用户，只删除其他数据

**方案 1**: 使用 `cleanup-specific-user.sql`，多次执行删除不同用户

**方案 2**: 自定义 SQL
```sql
-- 删除除了特定用户外的所有用户
DELETE FROM users 
WHERE username NOT IN ('admin', 'keep-user');
```

---

### Q3: 清理后无法注册新用户

**检查**:
1. 表结构是否完整
2. 触发器是否存在
3. 索引是否存在

**修复**: 重新运行 `database-schema.sql`

---

## 📝 快速命令

### 查看所有用户
```sql
SELECT username, email, created_at FROM users ORDER BY created_at DESC;
```

### 查看数据统计
```sql
SELECT 
    (SELECT COUNT(*) FROM users) as users,
    (SELECT COUNT(*) FROM projects) as projects,
    (SELECT COUNT(*) FROM apps) as apps;
```

### 清理所有数据（快速版）
```sql
BEGIN;
DELETE FROM project_app_settings;
DELETE FROM app_projects;
DELETE FROM user_projects;
DELETE FROM apps;
DELETE FROM projects;
DELETE FROM users;
COMMIT;
```

### 清理特定用户（快速版）
```sql
-- 替换 'testfix' 为实际用户名
DELETE FROM users WHERE username = 'testfix';
-- 注意：由于外键约束 ON DELETE CASCADE，相关数据会自动删除
```

---

## 🎯 推荐流程

### 开发测试阶段
1. 每天结束前运行 `view-test-data.sql` 查看数据
2. 每周运行一次 `cleanup-test-data.sql` 清理

### 准备上线前
1. 运行 `view-test-data.sql` 确认测试数据
2. 运行 `cleanup-test-data.sql` 清理所有测试数据
3. 测试注册、登录等核心功能
4. 创建第一个正式用户

### 生产环境
- ❌ 不要运行清理脚本
- ✅ 只删除特定用户（如果必要）
- ✅ 做好数据备份

---

## 📚 相关文档

- [database-schema.sql](./database-schema.sql) - 数据库结构
- [APP-API-GUIDE.md](../APP-API-GUIDE.md) - API 文档

---

## 🆘 需要帮助？

如果遇到问题：
1. 检查 Supabase 日志
2. 查看错误信息
3. 参考本文档的常见问题
4. 联系技术支持
