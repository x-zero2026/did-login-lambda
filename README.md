# DID Login Lambda - 后端服务

基于 DID（去中心化身份）的登录和项目管理平台后端服务。

## 📦 必需文件

### Lambda 部署必需
- `template.yaml` - AWS SAM 模板
- `samconfig.toml` - SAM 部署配置
- `Makefile` - 构建脚本
- `env.json` - Lambda 环境变量配置
- `go/` - Go 源代码目录

### 本地开发必需
- `go/.env` - 本地环境变量
- `go/local-server/` - 本地开发服务器
- `go/cmd/` - Lambda 函数源码
- `go/pkg/` - 共享包

## 🚀 快速开始

### 前置要求

- Go 1.21+
- AWS CLI
- AWS SAM CLI
- PostgreSQL 数据库（Supabase）

### 1. 配置环境变量

#### Lambda 环境变量 (env.json)
```bash
# 从示例文件创建配置
cp env.json.example env.json

# 编辑 env.json，填入实际配置
vim env.json
```

配置内容：
```json
{
  "Parameters": {
    "SupabaseURL": "https://xxx.supabase.co",
    "SupabaseAPIKey": "your-api-key",
    "DBPassword": "your-password",
    "JWTSecret": "your-secret"
  }
}
```

**注意**: `env.json` 包含敏感信息，已在 `.gitignore` 中，不会被提交到 Git

#### 本地开发环境变量 (go/.env)
```bash
cd go
cp .env.example .env
# 编辑 .env 文件
```

### 2. 本地开发

#### 方式 1: 使用本地服务器（推荐）
```bash
cd go
go run local-server/main.go
# 访问 http://localhost:8080
```

#### 方式 2: 使用 SAM Local
```bash
# 构建
sam build

# 本地运行
sam local start-api --env-vars env.json

# 访问 http://127.0.0.1:3000
```

### 3. 部署到 AWS Lambda

#### 首次部署
```bash
# 1. 构建
sam build

# 2. 部署（交互式配置）
sam deploy --guided

# 按提示输入：
# Stack Name: xz-login
# AWS Region: us-east-1
# Confirm changes: Y
# Allow SAM CLI IAM role creation: Y
# Save arguments to configuration file: Y
```

#### 后续部署
```bash
# 构建
sam build

# 部署（使用保存的配置）
sam deploy
```

#### 使用 Makefile
```bash
# 构建并部署
make deploy

# 只构建
make build

# 清理
make clean
```

## 📝 部署步骤详解

### 步骤 1: 准备数据库

1. 登录 Supabase 控制台
2. 进入 SQL Editor
3. 运行数据库初始化脚本（见 `docs/backend/database/database-schema.sql`）

### 步骤 2: 配置 AWS 凭证

```bash
# 配置 AWS CLI
aws configure

# 输入：
# AWS Access Key ID
# AWS Secret Access Key
# Default region: us-east-1
# Default output format: json
```

### 步骤 3: 配置环境变量

编辑 `env.json`，填入实际的数据库和 JWT 配置：

```json
{
  "RegisterFunction": {
    "SUPABASE_URL": "https://rbpsksuuvtzmathnmyxn.supabase.co",
    "DB_PASSWORD": "your-actual-password",
    "JWT_SECRET": "your-secret-key"
  },
  "LoginFunction": {
    "SUPABASE_URL": "https://rbpsksuuvtzmathnmyxn.supabase.co",
    "DB_PASSWORD": "your-actual-password",
    "JWT_SECRET": "your-secret-key"
  }
  // ... 其他函数配置相同
}
```

### 步骤 4: 构建项目

```bash
# 使用 SAM 构建
sam build

# 或使用 Makefile
make build
```

构建过程：
- 编译所有 Go Lambda 函数
- 生成 ARM64 架构的二进制文件
- 输出到 `.aws-sam/build/` 目录

### 步骤 5: 部署到 AWS

#### 首次部署（交互式）
```bash
sam deploy --guided
```

配置说明：
- **Stack Name**: `xz-login`（或自定义名称）
- **AWS Region**: `us-east-1`
- **Confirm changes**: `Y`（部署前确认）
- **Allow SAM CLI IAM role creation**: `Y`（允许创建 IAM 角色）
- **Disable rollback**: `N`（失败时回滚）
- **Save arguments**: `Y`（保存配置到 samconfig.toml）

#### 后续部署（使用保存的配置）
```bash
sam deploy
```

### 步骤 6: 获取 API Gateway URL

部署成功后，输出会显示 API Gateway URL：

```
CloudFormation outputs from deployed stack
---------------------------------------------------------
Outputs
---------------------------------------------------------
Key                 ApiGatewayUrl
Description         API Gateway endpoint URL
Value               https://xxx.execute-api.us-east-1.amazonaws.com/prod
---------------------------------------------------------
```

复制这个 URL，配置到前端的 `VITE_API_BASE_URL` 环境变量。

### 步骤 7: 测试部署

```bash
# 测试注册接口
curl -X POST https://xxx.execute-api.us-east-1.amazonaws.com/prod/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "username": "testuser",
    "password": "Password123"
  }'
```

## 🏗️ 项目结构

```
did-login-lambda/
├── template.yaml              # SAM 模板（定义所有 Lambda 函数）
├── samconfig.toml            # SAM 部署配置
├── Makefile                  # 构建和部署脚本
├── env.json                  # Lambda 环境变量
├── README.md                 # 本文件
│
└── go/                       # Go 源代码
    ├── cmd/                  # Lambda 函数入口
    │   ├── register/         # 注册函数
    │   ├── login/            # 登录函数
    │   ├── recover/          # 账户恢复
    │   ├── reset-password/   # 重置密码
    │   ├── get-profile/      # 获取用户信息
    │   ├── list-projects/    # 列出项目
    │   ├── update-project/   # 更新项目
    │   ├── list-apps/        # 列出应用
    │   ├── create-app/       # 创建应用
    │   ├── update-app/       # 更新应用
    │   ├── delete-app/       # 删除应用
    │   ├── set-app-global/   # 设置全局应用
    │   └── toggle-app/       # 开关应用
    │
    ├── pkg/                  # 共享包
    │   ├── auth/            # 认证相关（JWT, DID, 密码）
    │   ├── db/              # 数据库连接
    │   ├── models/          # 数据模型
    │   └── response/        # 响应格式
    │
    ├── local-server/        # 本地开发服务器
    │   └── main.go
    │
    ├── go.mod               # Go 模块定义
    ├── go.sum               # 依赖锁定
    ├── .env.example         # 环境变量示例
    └── README.md            # Go 项目说明
```

## 🔧 常用命令

### 构建
```bash
# SAM 构建
sam build

# 构建特定函数
sam build RegisterFunction
```

### 本地测试
```bash
# 启动本地 API
sam local start-api --env-vars env.json

# 调用特定函数
sam local invoke RegisterFunction --event events/register.json
```

### 部署
```bash
# 完整部署流程
sam build && sam deploy

# 使用 Makefile
make deploy
```

### 查看日志
```bash
# 查看特定函数日志
sam logs -n RegisterFunction --stack-name xz-login --tail

# 查看所有函数日志
sam logs --stack-name xz-login --tail
```

### 删除部署
```bash
# 删除 CloudFormation Stack
aws cloudformation delete-stack --stack-name xz-login
```

## 🔍 故障排查

### 问题 1: 构建失败

**错误**: `GOOS=linux GOARCH=arm64 go build: command not found`

**解决**:
```bash
# 确保安装了 Go
go version

# 确保在正确目录
cd did-login-lambda
sam build
```

### 问题 2: 部署失败 - 权限错误

**错误**: `User is not authorized to perform: cloudformation:CreateStack`

**解决**:
```bash
# 检查 AWS 凭证
aws sts get-caller-identity

# 确保有足够权限（需要 CloudFormation, Lambda, API Gateway, IAM 权限）
```

### 问题 3: Lambda 函数超时

**错误**: `Task timed out after 3.00 seconds`

**解决**: 在 `template.yaml` 中增加超时时间
```yaml
Timeout: 30  # 从 3 秒增加到 30 秒
```

### 问题 4: 数据库连接失败

**错误**: `Database connection failed`

**检查**:
1. `env.json` 中的数据库配置是否正确
2. Supabase 数据库是否可访问
3. 使用 Pooler 端口 6543（不是 5432）

### 问题 5: 环境变量未生效

**解决**:
```bash
# 确保 env.json 格式正确
cat env.json | jq .

# 重新部署
sam build && sam deploy
```

## 📚 相关文档

- [API 参考文档](../docs/did-login/API-REFERENCE.md) - 完整 API 文档
- [数据库文档](../docs/backend/database/) - 数据库结构和清理脚本
- [前端部署](../docs/frontend/DEPLOY-TO-AMPLIFY.md) - 前端部署指南
- [系统设计](../docs/general/DESIGN.md) - 系统架构设计

## 🔐 环境变量和安全

### 敏感文件（不要提交到 Git）

以下文件包含敏感信息，已在 `.gitignore` 中：

- `env.json` - Lambda 环境变量（数据库密码、JWT 密钥）
- `go/.env` - 本地开发环境变量
- `samconfig.toml` - SAM 部署配置（可能包含 AWS 账号信息）

### 示例文件（应该提交）

- `env.json.example` - 环境变量模板
- `go/.env.example` - 本地环境变量模板

### 新开发者设置

```bash
# 1. 克隆仓库
git clone <repository-url>
cd did-login-lambda

# 2. 配置环境变量
cp env.json.example env.json
vim env.json  # 填入实际配置

# 3. 配置本地开发环境
cd go
cp .env.example .env
vim .env  # 填入实际配置
```

详细说明请查看 [GITIGNORE-GUIDE.md](./GITIGNORE-GUIDE.md)

---

## 🛠️ 技术栈

- **语言**: Go 1.21+
- **部署**: AWS Lambda (ARM64)
- **API**: AWS API Gateway (REST API)
- **数据库**: PostgreSQL (Supabase)
- **认证**: JWT (HS256, 7天有效期)
- **加密**: bcrypt (密码), ed25519 (DID)
- **助记词**: BIP39 (12个单词)

## 📊 Lambda 函数列表

| 函数名 | 路径 | 方法 | 说明 |
|--------|------|------|------|
| RegisterFunction | /api/auth/register | POST | 用户注册 |
| LoginFunction | /api/auth/login | POST | 用户登录 |
| RecoverFunction | /api/auth/recover | POST | 账户恢复 |
| ResetPasswordFunction | /api/auth/reset-password | POST | 重置密码 |
| GetProfileFunction | /api/user/profile | GET | 获取用户信息 |
| ListProjectsFunction | /api/projects | GET | 列出项目 |
| UpdateProjectFunction | /api/projects/{id} | PUT | 更新项目 |
| ListAppsFunction | /api/projects/{id}/apps | GET | 列出应用 |
| CreateAppFunction | /api/apps | POST | 创建应用 |
| UpdateAppFunction | /api/apps/{id} | PUT | 更新应用 |
| DeleteAppFunction | /api/apps/{id} | DELETE | 删除应用 |
| SetAppGlobalFunction | /api/apps/{id}/set-global | PUT | 设置全局应用 |
| ToggleAppFunction | /api/projects/{id}/apps/{appId}/toggle | PUT | 开关应用 |

## 🔐 环境变量说明

### SUPABASE_URL
Supabase 项目 URL，格式：`https://xxx.supabase.co`

### DB_PASSWORD
数据库密码，用于连接 PostgreSQL

### JWT_SECRET
JWT 签名密钥，建议使用强随机字符串

### 连接字符串格式
```
postgresql://postgres.{project_ref}:{password}@aws-1-ap-south-1.pooler.supabase.com:6543/postgres
```

注意：使用 Pooler 端口 6543，不是直连端口 5432

## 🎯 下一步

1. ✅ 部署后端到 AWS Lambda
2. ✅ 获取 API Gateway URL
3. ✅ 配置前端环境变量
4. ✅ 部署前端到 AWS Amplify
5. ✅ 测试完整流程

## 🆘 需要帮助？

- 查看 [故障排查文档](../docs/general/TROUBLESHOOTING.md)
- 查看 [API 文档](../docs/did-login/API-REFERENCE.md)
- 检查 CloudWatch 日志
