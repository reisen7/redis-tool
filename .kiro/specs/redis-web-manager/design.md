# Design Document

## Overview

Redis Web Manager 采用前后端分离架构，前端使用 Vue 3 + Vite + Naive UI 构建轻量级 SPA 应用，后端使用 Go + Gin 提供 REST API 服务。系统设计遵循 Vibe 风格原则：代码简洁无冗余、交互流畅无延迟、视觉清爽简约、部署轻量化。

核心设计理念：
- **无状态后端**：后端不存储任何连接配置，每次请求携带完整连接信息
- **前端持久化**：所有连接配置存储在浏览器 localStorage
- **按需连接**：后端为每个请求创建临时 Redis 连接，执行完毕立即释放
- **最小依赖**：前后端仅引入必要依赖，确保轻量高性能

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Browser (Frontend)                        │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │ Connection  │  │   Command   │  │      Key Manager        │  │
│  │   Panel     │  │   Console   │  │   (List/Detail/Edit)    │  │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘  │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                    localStorage (Connections)                ││
│  └─────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
                              │ HTTP REST API
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Go Backend (Gin)                            │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │   Router    │  │  Handlers   │  │    Redis Service        │  │
│  │  (Gin)      │──│  (API)      │──│    (go-redis)           │  │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                       Redis Server(s)                            │
└─────────────────────────────────────────────────────────────────┘
```

## Components and Interfaces

### Frontend Components

```
src/
├── components/
│   ├── ConnectionPanel.vue    # 连接管理面板（左侧）
│   ├── ConnectionForm.vue     # 连接配置表单（弹窗）
│   ├── CommandConsole.vue     # 命令输入与执行
│   ├── ResultDisplay.vue      # 结果格式化展示
│   ├── KeyList.vue            # Key 列表与搜索
│   ├── KeyDetail.vue          # Key 详情与编辑
│   ├── ServerInfo.vue         # 服务器信息展示
│   └── ThemeToggle.vue        # 深色模式切换
├── composables/
│   ├── useConnections.ts      # 连接管理逻辑（localStorage）
│   ├── useRedisApi.ts         # API 请求封装
│   └── useTheme.ts            # 主题管理
├── types/
│   └── index.ts               # TypeScript 类型定义
├── App.vue                    # 根组件
└── main.ts                    # 入口文件
```

### Backend Structure

```
backend/
├── main.go                    # 入口文件
├── router/
│   └── router.go              # 路由定义
├── handler/
│   ├── connection.go          # 连接测试处理
│   ├── command.go             # 命令执行处理
│   ├── keys.go                # Key 管理处理
│   └── info.go                # 服务器信息处理
├── service/
│   └── redis.go               # Redis 操作服务
└── model/
    └── types.go               # 数据结构定义
```

### API Interfaces

| Endpoint | Method | Description | Request Body | Response |
|----------|--------|-------------|--------------|----------|
| `/api/ping` | POST | 测试连接 | `ConnectionConfig` | `{success, message}` |
| `/api/execute` | POST | 执行命令 | `{connection, command}` | `{success, data, type}` |
| `/api/keys` | POST | 获取 Key 列表 | `{connection, pattern, cursor, count}` | `{keys, cursor, total}` |
| `/api/key` | POST | 获取 Key 详情 | `{connection, key}` | `{type, value, ttl}` |
| `/api/key/set` | POST | 设置 Key | `{connection, key, type, value, ttl}` | `{success}` |
| `/api/key/delete` | POST | 删除 Key | `{connection, keys}` | `{success, deleted}` |
| `/api/key/ttl` | POST | 设置 TTL | `{connection, key, ttl}` | `{success}` |
| `/api/info` | POST | 服务器信息 | `ConnectionConfig` | `{info}` |

## Data Models

### ConnectionConfig (Frontend & Backend)

```typescript
interface ConnectionConfig {
  id: string;           // UUID，唯一标识
  name: string;         // 连接名称
  host: string;         // Redis 主机地址
  port: number;         // Redis 端口，默认 6379
  password: string;     // 密码，可为空
  db: number;           // 数据库索引，默认 0
  timeout: number;      // 连接超时（秒），默认 5
}
```

### CommandRequest

```typescript
interface CommandRequest {
  connection: ConnectionConfig;
  command: string;      // 完整命令字符串，如 "GET mykey"
}
```

### CommandResponse

```typescript
interface CommandResponse {
  success: boolean;
  data: any;            // 命令执行结果
  type: string;         // 结果类型：string, array, hash, set, nil, error
  error?: string;       // 错误信息
}
```

### KeyInfo

```typescript
interface KeyInfo {
  key: string;
  type: 'string' | 'hash' | 'list' | 'set' | 'zset';
  ttl: number;          // -1 表示永不过期，-2 表示不存在
}
```

### KeyValue

```typescript
interface KeyValue {
  type: string;
  value: string | string[] | Record<string, string> | Array<{member: string, score: number}>;
  ttl: number;
}
```



## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

Based on the acceptance criteria analysis, the following correctness properties must be verified through property-based testing:

### Property 1: Connection Persistence Round-Trip

*For any* valid connection configuration, saving it to localStorage and then loading it back should produce an equivalent configuration object with all fields preserved (id, name, host, port, password, db, timeout).

**Validates: Requirements 1.1, 1.2, 1.5, 1.6**

### Property 2: Command String Parsing

*For any* valid Redis command string, parsing it into command name and arguments should produce components that, when joined back together, form an equivalent command string.

**Validates: Requirements 2.7**

### Property 3: Result Formatting Consistency

*For any* Redis command result (string, array, hash, set, nil), the formatting function should produce a displayable output that preserves all data and can be parsed back to the original structure.

**Validates: Requirements 2.3, 2.8**

### Property 4: Key Info Display Completeness

*For any* KeyInfo object, the rendered display should contain the key name, a type indicator, and TTL information.

**Validates: Requirements 3.3**

### Property 5: Type-Appropriate Value Rendering

*For any* Redis value of type String, Hash, List, Set, or ZSet, the rendering function should produce output that correctly represents all data elements according to the type's structure (fields for Hash, indices for List, scores for ZSet).

**Validates: Requirements 4.1, 4.2, 4.3, 4.4, 4.5**

### Property 6: Input Validation Rejects Invalid Formats

*For any* input string that does not conform to the expected format for a given Redis data type, the validation function should reject it and return an appropriate error message.

**Validates: Requirements 4.6**

### Property 7: Theme Preference Persistence

*For any* theme setting (dark or light), saving the preference and reloading should restore the same theme setting.

**Validates: Requirements 6.2**

### Property 8: CORS Headers Present

*For any* HTTP response from the backend, the response should include appropriate CORS headers (Access-Control-Allow-Origin, Access-Control-Allow-Methods, Access-Control-Allow-Headers).

**Validates: Requirements 7.3**

### Property 9: Error Response Structure

*For any* error condition in the backend, the response should follow a consistent structure with success=false, an error field containing the message, and appropriate HTTP status code.

**Validates: Requirements 7.4**

### Property 10: No Credential Persistence

*For any* request containing connection credentials, the backend should not write any credential data to disk, logs, or any persistent storage.

**Validates: Requirements 8.4**

## Error Handling

### Frontend Error Handling

| Error Type | Handling Strategy | User Feedback |
|------------|-------------------|---------------|
| Network Error | Catch in API layer, retry once | "网络连接失败，请检查后端服务" |
| Connection Failed | Display specific Redis error | "连接失败: [具体错误信息]" |
| Command Error | Display Redis error message | "命令执行错误: [错误信息]" |
| Invalid Input | Validate before submit | "输入格式无效: [字段名]" |
| localStorage Full | Catch quota error | "本地存储空间不足" |

### Backend Error Handling

| Error Type | HTTP Status | Response Format |
|------------|-------------|-----------------|
| Invalid Request | 400 | `{success: false, error: "参数错误: [详情]"}` |
| Connection Failed | 500 | `{success: false, error: "连接失败: [详情]"}` |
| Command Error | 200 | `{success: false, error: "[Redis错误信息]"}` |
| Timeout | 504 | `{success: false, error: "连接超时"}` |
| Internal Error | 500 | `{success: false, error: "服务器内部错误"}` |

## Testing Strategy

### Unit Testing

使用 Vitest 进行前端单元测试，使用 Go testing 包进行后端单元测试。

**Frontend Unit Tests:**
- `useConnections.ts` - 连接配置的 CRUD 操作
- `useRedisApi.ts` - API 请求封装和错误处理
- 工具函数 - 命令解析、结果格式化

**Backend Unit Tests:**
- `service/redis.go` - Redis 操作封装
- `handler/*.go` - 请求处理和响应格式化
- 命令解析和参数验证

### Property-Based Testing

**Frontend (使用 fast-check):**
- 连接配置持久化 round-trip
- 命令字符串解析
- 结果格式化
- 主题偏好持久化

**Backend (使用 testing/quick 或 gopter):**
- CORS 头部验证
- 错误响应结构验证
- 请求参数验证

### Integration Testing

- 前后端 API 集成测试
- Redis 连接和命令执行测试
- 端到端用户流程测试

### Test Configuration

- Property-based tests: 最少 100 次迭代
- 每个 property test 必须标注对应的 correctness property
- 格式: `**Feature: redis-web-manager, Property {number}: {property_text}**`
