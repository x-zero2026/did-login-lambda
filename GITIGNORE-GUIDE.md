# .gitignore 说明文档

## 📁 忽略文件说明

### 🔐 敏感信息（必须忽略）

#### env.json
**包含**: Lambda 环境变量（数据库密码、JWT 密钥）  
**原因**: 包含敏感信息，不能提交到 Git  
**替代**: 使用 `env.json.example` 作为模板

#### .env
**包含**: 本地开发环境变量  
**原因**: 包含敏感信息  
**替代**: 使用 `.env.example` 作为模板

#### samconfig.toml
**包含**: SAM 部署配置（可能包含 AWS 账号信息）  
**原因**: 包含部署特定配置  
**替代**: 每个开发者自己运行 `sam deploy --guided` 生成

---

### 🏗️ 构建产物（应该忽略）

#### .aws-sam/
**包含**: SAM 构建输出  
**原因**: 自动生成，体积大  
**重建**: 运行 `sam build`

#### bootstrap
**包含**: Go Lambda 二进制文件  
**原因**: 自动生成，体积大  
**重建**: 运行 `sam build` 或 `make build`

#### *.zip
**包含**: Lambda 部署包  
**原因**: 自动生成  
**重建**: 运行 `sam build`

---

### 💻 IDE 和 OS 文件（应该忽略）

#### .vscode/, .idea/
**包含**: IDE 配置  
**原因**: 个人偏好，不应强制给其他开发者

#### .DS_Store
**包含**: macOS 文件夹元数据  
**原因**: 系统文件，无用

#### Thumbs.db
**包含**: Windows 缩略图缓存  
**原因**: 系统文件，无用

---

### 📊 测试和日志（应该忽略）

#### *.log
**包含**: 日志文件  
**原因**: 运行时生成，体积可能很大

#### *.coverprofile, coverage.txt
**包含**: 测试覆盖率报告  
**原因**: 自动生成  
**重建**: 运行 `go test -cover`

---

## 📝 应该提交的文件

### ✅ 必须提交

- `template.yaml` - SAM 模板
- `Makefile` - 构建脚本
- `README.md` - 文档
- `go/` - 源代码
  - `cmd/` - Lambda 函数
  - `pkg/` - 共享包
  - `go.mod` - Go 模块定义
  - `go.sum` - 依赖锁定

### ✅ 示例文件（应该提交）

- `env.json.example` - 环境变量模板
- `go/.env.example` - 本地环境变量模板

---

## 🔧 配置示例文件

### 创建 env.json.example

```bash
# 从实际配置创建示例（移除敏感信息）
cp env.json env.json.example

# 编辑 env.json.example，替换敏感信息为占位符
vim env.json.example
```

**env.json.example 内容**:
```json
{
  "RegisterFunction": {
    "SUPABASE_URL": "https://your-project.supabase.co",
    "DB_PASSWORD": "your-database-password",
    "JWT_SECRET": "your-jwt-secret-key"
  },
  "LoginFunction": {
    "SUPABASE_URL": "https://your-project.supabase.co",
    "DB_PASSWORD": "your-database-password",
    "JWT_SECRET": "your-jwt-secret-key"
  }
}
```

---

## 🚀 新开发者设置流程

### 1. 克隆仓库
```bash
git clone <repository-url>
cd did-login-lambda
```

### 2. 配置环境变量
```bash
# 复制示例文件
cp env.json.example env.json

# 编辑并填入实际配置
vim env.json

# 配置本地开发环境
cd go
cp .env.example .env
vim .env
```

### 3. 构建和部署
```bash
# 构建
sam build

# 部署（首次）
sam deploy --guided
# 这会生成 samconfig.toml（已在 .gitignore 中）
```

---

## ⚠️ 安全检查

### 提交前检查

```bash
# 检查是否有敏感文件
git status

# 确保这些文件不在列表中：
# - env.json
# - .env
# - samconfig.toml
# - 任何包含密码或密钥的文件
```

### 如果不小心提交了敏感信息

```bash
# 从 Git 历史中移除文件
git filter-branch --force --index-filter \
  "git rm --cached --ignore-unmatch env.json" \
  --prune-empty --tag-name-filter cat -- --all

# 强制推送（危险！确保团队知晓）
git push origin --force --all

# 更重要的是：立即更换泄露的密码和密钥！
```

---

## 📋 .gitignore 检查清单

### 提交代码前

- [ ] `env.json` 不在 Git 中
- [ ] `.env` 不在 Git 中
- [ ] `samconfig.toml` 不在 Git 中
- [ ] `.aws-sam/` 不在 Git 中
- [ ] 构建产物（bootstrap, *.zip）不在 Git 中
- [ ] IDE 配置文件不在 Git 中
- [ ] 日志文件不在 Git 中

### 提交代码时

- [ ] `env.json.example` 已提交
- [ ] `go/.env.example` 已提交
- [ ] 所有源代码已提交
- [ ] `template.yaml` 已提交
- [ ] `Makefile` 已提交
- [ ] `README.md` 已提交

---

## 🔍 验证 .gitignore

### 检查被忽略的文件

```bash
# 查看所有被忽略的文件
git status --ignored

# 检查特定文件是否被忽略
git check-ignore -v env.json
git check-ignore -v .aws-sam/
```

### 测试 .gitignore

```bash
# 创建测试文件
touch env.json
touch .aws-sam/test.txt

# 检查 git status（这些文件不应该出现）
git status

# 清理测试文件
rm env.json
rm -rf .aws-sam/
```

---

## 📚 相关文档

- [Git 官方文档 - .gitignore](https://git-scm.com/docs/gitignore)
- [GitHub .gitignore 模板](https://github.com/github/gitignore)
- [AWS SAM 最佳实践](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-best-practices.html)

---

## 🆘 常见问题

### Q: 为什么 samconfig.toml 要忽略？

**A**: `samconfig.toml` 包含部署特定的配置，如 AWS 区域、Stack 名称等。不同开发者可能有不同的配置，所以应该让每个人自己生成。

### Q: 如果团队需要共享 SAM 配置怎么办？

**A**: 可以创建 `samconfig.toml.example`，但要确保移除任何敏感信息。

### Q: .aws-sam/ 目录很大，可以删除吗？

**A**: 可以！运行 `sam build` 会重新生成。或者使用 `make clean` 清理。

### Q: 我想提交 .vscode/ 配置怎么办？

**A**: 如果是团队共享的配置（如推荐的扩展），可以创建 `.vscode/extensions.json` 并从 .gitignore 中排除它：
```gitignore
.vscode/*
!.vscode/extensions.json
```

---

## ✅ 最佳实践

1. **永远不要提交敏感信息**
   - 使用环境变量
   - 使用 AWS Secrets Manager
   - 使用示例文件（.example）

2. **保持 .gitignore 简洁**
   - 只忽略必要的文件
   - 使用注释说明原因

3. **定期审查**
   - 检查是否有新的文件类型需要忽略
   - 移除不再需要的规则

4. **团队协作**
   - 确保所有开发者理解 .gitignore
   - 在 README 中说明环境配置流程

---

完成！🎉
