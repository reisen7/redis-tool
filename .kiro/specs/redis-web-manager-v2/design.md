# Design Document

## Overview

Redis Web Manager V2 在现有架构基础上进行重大升级，引入用户认证系统、多 Tab 工作区、Pub/Sub 功能和管理员控制面板。系统采用前后端分离架构，后端从无状态转变为有状态服务，使用 SQLite 持久化用户数据和连接配置。

核心设计变更：
- **有状态后端**：后端使用 SQLite 存储用户、连接配置和系统设置
- **JWT 认证**：所有 API 请求需要携带 JWT token
- **WebSocket 支持**：Pub/Sub 功能使用 WebSocket 实现实时消息推送
- **多 Tab 架构**：前端支持每个连接打开多个独立工作区
- **RBAC 权限**：区分管理员和普通用户权限

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           Browser (Frontend)                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌─────────────────────────────────────────────────────┐  │
│  │   Auth       │  │                  Main Application                    │  │
│  │   Pages      │  │  ┌─────────────┐  ┌─────────────────────────────┐   │  │
│  │  - Login     │  │  │ Connection  │  │      Tab Workspace          │   │  │
│  │  - Register  │  │  │   Sidebar   │  │  ┌─────────────────────┐    │   │  │
│  │  - 2FA Setup │  │  │             │  │  │   Keys List (Top)   │    │   │  │
│  └──────────────┘  │  │             │  │  ├─────────────────────┤    │   │  │
│                    │  │             │  │  │  Terminal (Bottom)  │    │   │  │
│  ┌──────────────┐  │  │             │  │  └─────────────────────┘    │   │  │
│  │   Admin      │  │  │             │  │  [Tab1] [Tab2] [Tab3] [+]   │   │  │
│  │   Panel      │  │  └─────────────┘  └─────────────────────────────┘   │  │
│  │  - Users     │  │                                                      │  │
│  │  - Settings  │  │  ┌─────────────────────────────────────────────┐    │  │
│  └──────────────┘  │  │              Pub/Sub Panel                   │    │  │
│                    │  │  - Subscribe/Unsubscribe                     │    │  │
│                    │  │  - Message History                           │    │  │
│                    │  │  - Publish Message                           │    │  │
│                    │  └─────────────────────────────────────────────┘    │  │
│                    └─────────────────────────────────────────────────────┘  │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                    Pinia Stores (State Management)                   │   │
│  │  - authStore (JWT, user info)                                        │   │
│  │  - connectionStore (connections, active connection)                  │   │
│  │  - tabStore (tabs, active tab)                                       │   │
│  │  - pubsubStore (subscriptions, messages)                             │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘
                    │ HTTP REST API + WebSocket
                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                          Go Backend (Gin)                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │   Router    │  │ Middleware  │  │  Handlers   │  │     Services        │ │
│  │   (Gin)     │──│  - CORS     │──│  - Auth     │──│  - UserService      │ │
│  │             │  │  - JWT Auth │  │  - User     │  │  - ConnectionService│ │
│  │             │  │             │  │  - Admin    │  │  - RedisService     │ │
│  │             │  │             │  │  - Redis    │  │  - PubSubService    │ │
│  │             │  │             │  │  - PubSub   │  │  - SettingsService  │ │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────────────┘ │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                    SQLite Database                                   │    │
│  │  - users (id, username, password_hash, role, totp_secret, ...)      │    │
│  │  - connections (id, user_id, name, host, port, password_enc, ...)   │    │
│  │  - settings (key, value)                                             │    │
│  │  - recovery_codes (id, user_id, code_hash, used)                    │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                    WebSocket Hub (Pub/Sub)                           │    │
│  │  - Client connections management                                     │    │
│  │  - Redis subscription forwarding                                     │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                          Redis Server(s)                                     │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Components and Interfaces

### Frontend Components (New/Modified)

```
src/
├── views/
│   ├── LoginView.vue          # 登录页面
│   ├── RegisterView.vue       # 注册页面
│   ├── TwoFactorView.vue      # 2FA 验证页面
│   ├── MainView.vue           # 主应用视图
│   └── AdminView.vue          # 管理员面板
├── components/
│   ├── auth/
│   │   ├── LoginForm.vue      # 登录表单
│   │   ├── RegisterForm.vue   # 注册表单
│   │   └── TotpSetup.vue      # 2FA 设置组件
│   ├── workspace/
│   │   ├── TabBar.vue         # Tab 栏
│   │   ├── TabContent.vue     # Tab 内容容器
│   │   ├── WorkspacePanel.vue # 工作区面板（Keys + Terminal）
│   │   └── ResizableDivider.vue # 可调整分隔条
│   ├── pubsub/
│   │   ├── PubSubPanel.vue    # Pub/Sub 主面板
│   │   ├── SubscriptionList.vue # 订阅列表
│   │   ├── MessageHistory.vue # 消息历史
│   │   └── PublishForm.vue    # 发布消息表单
│   ├── admin/
│   │   ├── UserList.vue       # 用户列表
│   │   ├── UserForm.vue       # 用户表单
│   │   └── SettingsPanel.vue  # 系统设置面板
│   └── ... (existing components)
├── stores/
│   ├── auth.ts                # 认证状态
│   ├── tabs.ts                # Tab 状态管理
│   ├── pubsub.ts              # Pub/Sub 状态
│   └── ... (existing stores)
├── api/
│   ├── auth.ts                # 认证 API
│   ├── admin.ts               # 管理员 API
│   ├── pubsub.ts              # Pub/Sub API
│   └── ... (existing APIs)
├── composables/
│   ├── useAuth.ts             # 认证逻辑
│   ├── useWebSocket.ts        # WebSocket 管理
│   └── ... (existing composables)
└── router/
    └── index.ts               # 路由配置（含守卫）
```

### Backend Structure (New/Modified)

```
backend/
├── main.go
├── config/
│   └── config.go              # 配置管理
├── router/
│   └── router.go              # 路由定义
├── middleware/
│   ├── cors.go                # CORS 中间件
│   └── auth.go                # JWT 认证中间件
├── handler/
│   ├── auth.go                # 认证处理（登录、注册、2FA）
│   ├── user.go                # 用户管理
│   ├── admin.go               # 管理员操作
│   ├── connection.go          # 连接管理（修改：关联用户）
│   ├── command.go             # 命令执行
│   ├── keys.go                # Key 管理
│   ├── info.go                # 服务器信息
│   └── pubsub.go              # Pub/Sub WebSocket
├── service/
│   ├── user.go                # 用户服务
│   ├── auth.go                # 认证服务（JWT、TOTP）
│   ├── connection.go          # 连接服务
│   ├── redis.go               # Redis 操作
│   ├── pubsub.go              # Pub/Sub 服务
│   ├── settings.go            # 系统设置服务
│   └── database.go            # 数据库服务
├── model/
│   ├── user.go                # 用户模型
│   ├── connection.go          # 连接模型
│   ├── settings.go            # 设置模型
│   └── types.go               # 通用类型
└── websocket/
    └── hub.go                 # WebSocket Hub
```

### API Interfaces (New)

#### Authentication APIs

| Endpoint | Method | Auth | Description | Request Body | Response |
|----------|--------|------|-------------|--------------|----------|
| `/api/auth/login` | POST | No | 用户登录 | `{username, password}` | `{token, user, requires2FA}` |
| `/api/auth/register` | POST | No | 用户注册 | `{username, password}` | `{success, message}` |
| `/api/auth/verify-2fa` | POST | No | 验证 2FA | `{tempToken, code}` | `{token, user}` |
| `/api/auth/logout` | POST | Yes | 登出 | - | `{success}` |
| `/api/auth/me` | GET | Yes | 获取当前用户 | - | `{user}` |

#### 2FA APIs

| Endpoint | Method | Auth | Description | Request Body | Response |
|----------|--------|------|-------------|--------------|----------|
| `/api/2fa/setup` | POST | Yes | 启用 2FA | - | `{secret, qrCode, recoveryCodes}` |
| `/api/2fa/verify` | POST | Yes | 验证并激活 | `{code}` | `{success}` |
| `/api/2fa/disable` | POST | Yes | 禁用 2FA | `{password, code}` | `{success}` |

#### Admin APIs

| Endpoint | Method | Auth | Description | Request Body | Response |
|----------|--------|------|-------------|--------------|----------|
| `/api/admin/users` | GET | Admin | 获取用户列表 | - | `{users}` |
| `/api/admin/users` | POST | Admin | 创建用户 | `{username, password, role}` | `{user}` |
| `/api/admin/users/:id` | PUT | Admin | 更新用户 | `{enabled, role}` | `{user}` |
| `/api/admin/users/:id` | DELETE | Admin | 删除用户 | - | `{success}` |
| `/api/admin/users/:id/reset-password` | POST | Admin | 重置密码 | - | `{tempPassword}` |
| `/api/admin/settings` | GET | Admin | 获取设置 | - | `{settings}` |
| `/api/admin/settings` | PUT | Admin | 更新设置 | `{key, value}` | `{success}` |

#### Connection APIs (Modified)

| Endpoint | Method | Auth | Description | Request Body | Response |
|----------|--------|------|-------------|--------------|----------|
| `/api/connections` | GET | Yes | 获取用户连接列表 | - | `{connections}` |
| `/api/connections` | POST | Yes | 创建连接 | `ConnectionConfig` | `{connection}` |
| `/api/connections/:id` | PUT | Yes | 更新连接 | `ConnectionConfig` | `{connection}` |
| `/api/connections/:id` | DELETE | Yes | 删除连接 | - | `{success}` |
| `/api/connections/:id/test` | POST | Yes | 测试连接 | - | `{success, message}` |

#### Pub/Sub APIs

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/api/pubsub/ws` | WebSocket | Yes | Pub/Sub WebSocket 连接 |
| `/api/pubsub/publish` | POST | Yes | 发布消息 |

## Data Models

### User

```go
type User struct {
    ID            uint      `json:"id" gorm:"primaryKey"`
    Username      string    `json:"username" gorm:"uniqueIndex;size:50"`
    PasswordHash  string    `json:"-" gorm:"size:100"`
    Role          string    `json:"role" gorm:"size:20;default:'user'"` // "admin" or "user"
    Enabled       bool      `json:"enabled" gorm:"default:true"`
    TotpSecret    string    `json:"-" gorm:"size:100"`
    TotpEnabled   bool      `json:"totpEnabled" gorm:"default:false"`
    MustChangePwd bool      `json:"mustChangePassword" gorm:"default:false"`
    CreatedAt     time.Time `json:"createdAt"`
    UpdatedAt     time.Time `json:"updatedAt"`
}
```

### Connection (Modified)

```go
type Connection struct {
    ID           uint      `json:"id" gorm:"primaryKey"`
    UserID       uint      `json:"userId" gorm:"index"`
    Name         string    `json:"name" gorm:"size:100"`
    Host         string    `json:"host" gorm:"size:255"`
    Port         int       `json:"port" gorm:"default:6379"`
    PasswordEnc  string    `json:"-" gorm:"size:500"` // AES-256 encrypted
    DB           int       `json:"db" gorm:"default:0"`
    Timeout      int       `json:"timeout" gorm:"default:5"`
    CreatedAt    time.Time `json:"createdAt"`
    UpdatedAt    time.Time `json:"updatedAt"`
}
```

### RecoveryCode

```go
type RecoveryCode struct {
    ID       uint   `json:"id" gorm:"primaryKey"`
    UserID   uint   `json:"userId" gorm:"index"`
    CodeHash string `json:"-" gorm:"size:100"`
    Used     bool   `json:"used" gorm:"default:false"`
}
```

### SystemSetting

```go
type SystemSetting struct {
    Key   string `json:"key" gorm:"primaryKey;size:50"`
    Value string `json:"value" gorm:"size:500"`
}
```

### Frontend Types

```typescript
// Auth Types
interface User {
  id: number;
  username: string;
  role: 'admin' | 'user';
  enabled: boolean;
  totpEnabled: boolean;
  mustChangePassword: boolean;
  createdAt: string;
}

interface LoginResponse {
  token?: string;
  user?: User;
  requires2FA?: boolean;
  tempToken?: string;
}

// Tab Types
interface Tab {
  id: string;
  connectionId: number;
  title: string;
  keysState: KeysState;
  terminalState: TerminalState;
  dividerPosition: number; // percentage
}

interface KeysState {
  pattern: string;
  cursor: string;
  selectedKeys: string[];
  scrollPosition: number;
}

interface TerminalState {
  history: CommandHistoryItem[];
  inputValue: string;
}

// Pub/Sub Types
interface PubSubMessage {
  id: string;
  channel: string;
  pattern?: string;
  message: string;
  timestamp: Date;
}

interface Subscription {
  id: string;
  pattern: string;
  isPattern: boolean;
  active: boolean;
}

// WebSocket Message Types
interface WSMessage {
  type: 'subscribe' | 'unsubscribe' | 'message' | 'error' | 'status';
  payload: any;
}
```



## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

Based on the acceptance criteria analysis, the following correctness properties must be verified through property-based testing:

### Property 1: Tab State Persistence

*For any* Tab with any state (scroll position, selected keys, divider position, terminal history), switching to another Tab and switching back should restore the exact same state.

**Validates: Requirements 1.5, 1.7**

### Property 2: Tab Creation Increases Count

*For any* workspace with N tabs, clicking "New Tab" should result in exactly N+1 tabs, with the new tab having an independent state.

**Validates: Requirements 1.2**

### Property 3: Tab Removal Decreases Count

*For any* workspace with N tabs (N > 0), closing a tab should result in exactly N-1 tabs, and the closed tab's state should no longer be accessible.

**Validates: Requirements 1.3**

### Property 4: Subscription Management Consistency

*For any* channel pattern, subscribing should add it to the subscription list, and unsubscribing should remove it from the list. The subscription list should accurately reflect all active subscriptions.

**Validates: Requirements 2.2, 2.5**

### Property 5: Message Display Completeness

*For any* Pub/Sub message received, the displayed message should contain the channel name, the message content, and a timestamp.

**Validates: Requirements 2.3**

### Property 6: Clear Messages Preserves Subscriptions

*For any* state with active subscriptions and message history, clearing the message history should remove all messages while the subscription list remains unchanged.

**Validates: Requirements 2.8**

### Property 7: JWT Included in Authenticated Requests

*For any* API request made while the user is authenticated, the request should include the JWT token in the Authorization header with "Bearer" prefix.

**Validates: Requirements 3.6**

### Property 8: Protected Endpoints Require Valid JWT

*For any* protected API endpoint, a request without a valid JWT token should return HTTP 401 Unauthorized status.

**Validates: Requirements 3.7**

### Property 9: Registration Disabled Returns 403

*For any* registration attempt when registration is disabled in system settings, the backend should return HTTP 403 Forbidden status.

**Validates: Requirements 4.2**

### Property 10: Password Validation Rules

*For any* password string, the validation function should return true if and only if the password contains at least 8 characters, at least one uppercase letter, at least one lowercase letter, and at least one number.

**Validates: Requirements 4.5, 4.6**

### Property 11: Invalid TOTP Denied

*For any* TOTP code that does not match the current or adjacent time windows for the user's secret, the verification should fail and return an error.

**Validates: Requirements 5.4**

### Property 12: User Data Isolation

*For any* user querying their connections, the result should contain only connections where the user_id matches the requesting user's ID.

**Validates: Requirements 6.2**

### Property 13: Cross-User Access Denied

*For any* user attempting to access, modify, or delete a connection owned by a different user, the backend should return HTTP 403 Forbidden status.

**Validates: Requirements 6.3**

### Property 14: User Enabled State Affects Login

*For any* user account, login should succeed if and only if the account is enabled. Disabled accounts should always fail login with an appropriate error.

**Validates: Requirements 7.3, 7.4**

### Property 15: Admin-Only Endpoints Enforce Role

*For any* admin-only API endpoint, a request from a non-admin user should return HTTP 403 Forbidden status.

**Validates: Requirements 7.7**

### Property 16: Settings Persistence and Immediate Effect

*For any* system setting change, the new value should be immediately retrievable and should affect system behavior without requiring restart.

**Validates: Requirements 8.1, 8.3**

### Property 17: Connection Password Encryption Round-Trip

*For any* connection password, encrypting it with AES-256 and then decrypting should produce the original password.

**Validates: Requirements 10.4**

## Error Handling

### Frontend Error Handling

| Error Type | Handling Strategy | User Feedback |
|------------|-------------------|---------------|
| 401 Unauthorized | Clear auth state, redirect to login | "会话已过期，请重新登录" |
| 403 Forbidden | Display error, stay on current page | "您没有权限执行此操作" |
| Network Error | Retry once, then show error | "网络连接失败，请检查网络" |
| WebSocket Disconnect | Auto-reconnect with backoff | "连接已断开，正在重连..." |
| Validation Error | Highlight invalid fields | 显示具体字段错误信息 |
| 2FA Required | Redirect to 2FA verification | "请输入双重认证码" |

### Backend Error Handling

| Error Type | HTTP Status | Response Format |
|------------|-------------|-----------------|
| Invalid Credentials | 401 | `{success: false, error: "用户名或密码错误"}` |
| Invalid Token | 401 | `{success: false, error: "认证已过期"}` |
| Forbidden | 403 | `{success: false, error: "没有权限"}` |
| Not Found | 404 | `{success: false, error: "资源不存在"}` |
| Validation Error | 400 | `{success: false, error: "参数错误", details: {...}}` |
| Registration Disabled | 403 | `{success: false, error: "注册功能已关闭"}` |
| Duplicate Username | 409 | `{success: false, error: "用户名已存在"}` |
| Internal Error | 500 | `{success: false, error: "服务器内部错误"}` |

### WebSocket Error Handling

| Event | Handling Strategy |
|-------|-------------------|
| Connection Failed | Retry with exponential backoff (1s, 2s, 4s, max 30s) |
| Auth Failed | Close connection, redirect to login |
| Subscription Error | Notify user, remove from subscription list |
| Message Parse Error | Log error, skip message |

## Testing Strategy

### Unit Testing

使用 Vitest 进行前端单元测试，使用 Go testing 包进行后端单元测试。

**Frontend Unit Tests:**
- `stores/auth.ts` - 认证状态管理
- `stores/tabs.ts` - Tab 状态管理
- `stores/pubsub.ts` - Pub/Sub 状态管理
- `composables/useAuth.ts` - 认证逻辑
- `composables/useWebSocket.ts` - WebSocket 管理
- 密码验证函数
- Tab 状态序列化/反序列化

**Backend Unit Tests:**
- `service/auth.go` - JWT 生成和验证、TOTP 验证
- `service/user.go` - 用户 CRUD、密码哈希
- `service/connection.go` - 连接 CRUD、权限验证
- `service/settings.go` - 设置管理
- `middleware/auth.go` - JWT 中间件
- AES-256 加密/解密

### Property-Based Testing

**Frontend (使用 fast-check):**
- Tab 状态持久化 round-trip (Property 1)
- Tab 创建/删除计数 (Property 2, 3)
- 订阅管理一致性 (Property 4)
- 消息显示完整性 (Property 5)
- 清除消息保留订阅 (Property 6)
- 密码验证规则 (Property 10)

**Backend (使用 testing/quick 或 gopter):**
- JWT 请求头验证 (Property 7, 8)
- 注册禁用返回 403 (Property 9)
- TOTP 验证 (Property 11)
- 用户数据隔离 (Property 12, 13)
- 用户启用状态影响登录 (Property 14)
- 管理员端点权限 (Property 15)
- 设置持久化 (Property 16)
- 密码加密 round-trip (Property 17)

### Integration Testing

- 完整登录流程（包括 2FA）
- 用户注册流程
- 连接 CRUD 与权限验证
- Pub/Sub WebSocket 消息流
- 管理员用户管理流程

### Test Configuration

- Property-based tests: 最少 100 次迭代
- 每个 property test 必须标注对应的 correctness property
- 格式: `**Feature: redis-web-manager-v2, Property {number}: {property_text}**`

