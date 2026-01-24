# 数据库文档

## 📁 文件说明

### database-schema.sql
**用途**: 数据库表结构初始化脚本  
**使用**: 在 Supabase SQL Editor 中运行，创建所有表、索引和约束

### 清理脚本

#### quick-cleanup.sql
**用途**: 一键清理所有测试数据（推荐）  
**安全性**: ⚠️ 会删除所有数据，保留表结构

#### view-test-data.sql
**用途**: 查看当前数据库中的所有数据  
**安全性**: ✅ 只读，不会修改数据

#### cleanup-test-data.sql
**用途**: 详细的清理脚本，带注释  
**安全性**: ⚠️ 会删除所有数据

#### cleanup-specific-user.sql
**用途**: 清理特定用户及其相关数据  
**安全性**: ⚠️ 会删除指定用户的所有数据

## 📚 详细文档

完整的使用指南请查看 [DATABASE-CLEANUP-GUIDE.md](./DATABASE-CLEANUP-GUIDE.md)

## 🚀 快速使用

### 初始化数据库
```sql
-- 在 Supabase SQL Editor 中运行
\i database-schema.sql
```

### 查看数据
```sql
\i view-test-data.sql
```

### 清理所有数据
```sql
\i quick-cleanup.sql
```

## 🔗 相关文档

- [DATABASE-CLEANUP-GUIDE.md](./DATABASE-CLEANUP-GUIDE.md) - 详细清理指南
- [API 文档](../../did-login/API-REFERENCE.md) - API 接口文档
