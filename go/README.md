# X-Zero Platform - Backend (Go + AWS Lambda)

基于DID的登录和项目管理平台后端服务。

## 技术栈

- Go 1.21+
- AWS Lambda + API Gateway
- PostgreSQL (Supabase)
- JWT认证
- Ed25519 + BIP39

## 项目结构

```
go/
├── cmd/                    # Lambda函数入口
│   ├── register/          # 用户注册
│   ├── login/             # 用户登录
│   ├── recover/           # 助记词恢复
│   ├── reset-password/    # 重置密码
│   ├── get-profile/       # 获取用户信息
│   ├── list-projects/     # 获取项目列表
│   ├── update-project/    # 更新项目
│   ├── list-apps/         # 获取应用列表
│   ├── create-app/        # 创建应用
│   ├── update-app/        # 更新应用
│   ├── delete-app/        # 删除应用
│   ├── set-app-global/    # 设置应用为全局
│   └── toggle-app/        # 开关应用
├── pkg/                   # 共享包
│   ├── auth/             # 认证相关（JWT、密码、DID）
│   ├── db/               # 数据库连接
│   ├── models/           # 数据模型
│   └── response/         # API响应
├── go.mod
├── go.sum
└── .env.example
```

## 环境配置

1. 复制环境变量模板：
```bash
cp .env.example .env
```

2. 编辑 `.env` 文件，填入实际值：
```
SUPABASE_URL=https://xxx.supabase.co
SUPABASE_API_KEY=your_api_key
JWT_SECRET=your_strong_random_secret
JWT_EXPIRY=168h
```

## 数据库初始化

在Supabase SQL编辑器中运行 `../database-schema.sql` 文件。

## 本地开发

### 安装依赖

```bash
go mod download
```

### 使用AWS SAM本地调试

1. 安装AWS SAM CLI：
```bash
# macOS
brew install aws-sam-cli

# 其他平台参考：https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/install-sam-cli.html
```

2. 构建项目：
```bash
cd ..  # 回到did-login-lambda目录
sam build
```

3. 本地运行API：
```bash
sam local start-api --env-vars env.json
```

创建 `env.json` 文件：
```json
{
  "Parameters": {
    "SupabaseURL": "https://xxx.supabase.co",
    "SupabaseAPIKey": "your_api_key",
    "JWTSecret": "your_jwt_secret"
  }
}
```

## 部署到AWS

1. 构建：
```bash
sam build
```

2. 首次部署（引导式）：
```bash
sam deploy --guided
```

按提示输入：
- Stack Name: xzero-did-login
- AWS Region: us-east-1 (或您的区域)
- Parameter SupabaseURL: 您的Supabase URL
- Parameter SupabaseAPIKey: 您的Supabase API Key
- Parameter JWTSecret: 您的JWT密钥
- Confirm changes before deploy: Y
- Allow SAM CLI IAM role creation: Y
- Save arguments to configuration file: Y

3. 后续部署：
```bash
sam deploy
```

## API端点

部署后，您将获得API Gateway URL，例如：
```
https://xxx.execute-api.us-east-1.amazonaws.com/prod/
```

### 认证相关
- POST `/api/auth/register` - 注册
- POST `/api/auth/login` - 登录
- POST `/api/auth/recover` - 助记词恢复
- POST `/api/auth/reset-password` - 重置密码

### 用户相关
- GET `/api/user/profile` - 获取用户信息（需要认证）

### 项目相关
- GET `/api/projects` - 获取项目列表（需要认证）
- PUT `/api/projects/:id` - 更新项目（需要认证）

### 应用相关
- GET `/api/projects/:project_id/apps` - 获取项目应用列表（需要认证）
- POST `/api/apps` - 创建应用（需要认证）
- PUT `/api/apps/:id` - 更新应用（需要认证）
- DELETE `/api/apps/:id` - 删除应用（需要认证）
- PUT `/api/apps/:id/set-global` - 设置应用为全局（需要认证）
- PUT `/api/projects/:project_id/apps/:app_id/toggle` - 开关应用（需要认证）

## 注意事项

1. **数据库连接**：需要在 `pkg/db/postgres.go` 中配置正确的数据库密码
2. **JWT密钥**：生产环境务必使用强随机密钥
3. **CORS**：已配置允许所有来源，生产环境建议限制具体域名
4. **错误处理**：所有API都返回统一的JSON格式

## 故障排查

### Lambda函数超时
- 检查数据库连接是否正常
- 增加Lambda函数的超时时间（在template.yaml中）

### 数据库连接失败
- 确认Supabase URL和API Key正确
- 检查数据库密码是否正确配置

### JWT验证失败
- 确认JWT_SECRET在所有Lambda函数中一致
- 检查token是否过期
