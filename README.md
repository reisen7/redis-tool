# Redis Web Manager

一个现代化的 Redis 数据库 Web 管理工具，支持多用户、多连接管理。

## 功能特性

- 🔐 **用户认证** - 支持用户注册、登录、双重认证(2FA/TOTP)
- 👥 **多用户隔离** - 每个用户的连接和数据完全独立
- 🔗 **连接管理** - 支持多个 Redis 连接，分组管理
- 📊 **数据浏览** - 支持 String、List、Hash、Set、ZSet 等数据类型
- 💻 **命令终端** - 内置 Redis 命令行终端
- 📡 **Pub/Sub** - 支持发布/订阅消息监控
- 🎨 **主题切换** - 支持亮色/暗色主题
- 🔒 **安全存储** - 连接密码 AES-256 加密存储

## 技术栈

**后端:**
- Go 1.21+
- Gin Web Framework
- GORM + SQLite
- go-redis

**前端:**
- Vue 3 + TypeScript
- Vite
- Element Plus
- Pinia
- CodeMirror

## 快速开始

### 环境要求

- Go 1.21+
- Node.js 18+
- pnpm

### 后端启动

```bash
cd backend

# 复制配置文件
cp config.toml.example config.toml

# 编辑配置（可选）
# vim config.toml

# 运行
go run main.go
```

后端默认运行在 `http://localhost:8080`

### 前端启动

```bash
cd frontend

# 安装依赖
pnpm install

# 开发模式
pnpm dev
```

前端默认运行在 `http://localhost:5173`

### 默认账户

首次启动会自动创建管理员账户：
- 用户名: `admin`
- 密码: `admin123`

**请在生产环境中及时修改默认密码！**

## 配置说明

后端配置文件 `backend/config.toml`:

```toml
[server]
port = 8080

[database]
path = "./data/redis_manager.db"

[security]
jwt_secret = "your-secret-key"
default_admin_user = "admin"
default_admin_pass = "admin123"
```

## 项目结构

```
├── backend/                # Go 后端
│   ├── config/            # 配置管理
│   ├── handler/           # HTTP 处理器
│   ├── middleware/        # 中间件
│   ├── model/             # 数据模型
│   ├── router/            # 路由配置
│   ├── service/           # 业务逻辑
│   └── main.go
├── frontend/              # Vue 前端
│   ├── src/
│   │   ├── api/          # API 接口
│   │   ├── components/   # 组件
│   │   ├── stores/       # Pinia 状态
│   │   ├── types/        # TypeScript 类型
│   │   └── views/        # 页面
│   └── package.json
└── README.md
```

## 开发

### 后端测试

```bash
cd backend
go test ./...
```

### 前端测试

```bash
cd frontend
pnpm test
```

### 构建

```bash
# 后端
cd backend
go build -o redis-web-manager

# 前端
cd frontend
pnpm build
```

## License

MIT
