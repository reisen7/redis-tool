# Requirements Document

## Introduction

Redis Web Manager 是一款基于 Vue 3（前端）+ Go（后端）技术栈的 Redis 可视化管理工具。该工具旨在提供媲美 Redis 官方工具（redis-cli/Redis Insight）的功能体验，同时严格遵循 Vibe 编程风格——简洁高效、交互流畅、视觉美观、代码无冗余、轻量高性能。核心特性包括连接配置本地持久化、全面的 Redis 操作支持、以及优雅的深色模式界面。

## Glossary

- **Redis Web Manager**: 本项目开发的 Redis 可视化 Web 管理工具
- **Connection Configuration**: Redis 连接配置，包含主机、端口、密码、数据库索引等信息
- **Key**: Redis 中存储数据的键名
- **Value**: Redis 中与 Key 关联的数据值
- **Data Type**: Redis 支持的数据类型，包括 String、Hash、List、Set、ZSet
- **TTL (Time To Live)**: Key 的过期时间，单位为秒
- **DB Index**: Redis 数据库索引，默认 0-15
- **Command**: Redis 命令，如 GET、SET、HGETALL 等
- **localStorage**: 浏览器本地存储 API，用于持久化连接配置
- **Vibe Style**: 一种编程风格，强调简洁优雅、交互流畅、视觉清爽、轻量高性能

## Requirements

### Requirement 1: Connection Management

**User Story:** As a developer, I want to manage multiple Redis connection configurations, so that I can easily switch between different Redis instances (development/testing/production).

#### Acceptance Criteria

1. WHEN a user creates a new connection THEN Redis Web Manager SHALL save the connection configuration (name, host, port, password, db index, timeout) to browser localStorage
2. WHEN a user refreshes the page or restarts the browser THEN Redis Web Manager SHALL restore all previously saved connection configurations from localStorage
3. WHEN a user clicks the connect button THEN Redis Web Manager SHALL establish a connection to the specified Redis server and display the connection status (connected/disconnected/connecting)
4. WHEN a connection attempt fails THEN Redis Web Manager SHALL display a specific error message indicating the failure reason (password incorrect, host unreachable, timeout)
5. WHEN a user edits an existing connection THEN Redis Web Manager SHALL update the connection configuration in localStorage and reflect changes immediately
6. WHEN a user deletes a connection THEN Redis Web Manager SHALL remove the connection configuration from localStorage and update the connection list
7. WHEN a user switches database index on an active connection THEN Redis Web Manager SHALL execute SELECT command and update the current database context

### Requirement 2: Command Execution

**User Story:** As a developer, I want to execute Redis commands manually, so that I can perform custom operations and debug Redis data.

#### Acceptance Criteria

1. WHEN a user types a Redis command in the input field THEN Redis Web Manager SHALL display syntax highlighting for the command
2. WHEN a user presses Ctrl+Enter or clicks the execute button THEN Redis Web Manager SHALL send the command to the backend and display a loading indicator
3. WHEN the backend returns a successful result THEN Redis Web Manager SHALL format and display the result based on its data type (string, array, hash, set)
4. WHEN the backend returns an error THEN Redis Web Manager SHALL display the error message in a distinct error format
5. WHEN a user clicks the copy button on a result THEN Redis Web Manager SHALL copy the result content to the clipboard
6. WHEN displaying complex results (arrays, hashes) THEN Redis Web Manager SHALL provide collapse/expand functionality for better readability
7. WHEN a command is executed THEN Redis Web Manager SHALL parse the command string and serialize it to JSON for backend transmission
8. WHEN displaying command results THEN Redis Web Manager SHALL pretty-print JSON data for readability

### Requirement 3: Key Management

**User Story:** As a developer, I want to browse and manage Redis keys visually, so that I can efficiently inspect and modify data without memorizing commands.

#### Acceptance Criteria

1. WHEN a user views the key list THEN Redis Web Manager SHALL display all keys in the current database with pagination (default 100 keys per page)
2. WHEN a user enters a search pattern THEN Redis Web Manager SHALL filter keys using the SCAN command with the specified pattern
3. WHEN displaying a key THEN Redis Web Manager SHALL show the key name, data type icon, and TTL information
4. WHEN a user clicks on a key THEN Redis Web Manager SHALL fetch and display the key's value in a type-appropriate format
5. WHEN a user creates a new key THEN Redis Web Manager SHALL provide a form with type selection (String/Hash/List/Set/ZSet) and value input
6. WHEN a user deletes a key THEN Redis Web Manager SHALL confirm the action and execute the DEL command
7. WHEN a user modifies a key's TTL THEN Redis Web Manager SHALL execute EXPIRE or PERSIST command based on user input
8. WHEN a user selects multiple keys THEN Redis Web Manager SHALL enable batch delete operation

### Requirement 4: Data Type Operations

**User Story:** As a developer, I want type-specific operations for different Redis data types, so that I can manipulate complex data structures efficiently.

#### Acceptance Criteria

1. WHEN viewing a String key THEN Redis Web Manager SHALL display the value and provide edit/delete operations
2. WHEN viewing a Hash key THEN Redis Web Manager SHALL display all fields in a table format with add/edit/delete field operations
3. WHEN viewing a List key THEN Redis Web Manager SHALL display elements with index numbers and provide push/pop/remove operations
4. WHEN viewing a Set key THEN Redis Web Manager SHALL display all members and provide add/remove member operations
5. WHEN viewing a ZSet key THEN Redis Web Manager SHALL display members with scores in a sortable table and provide add/update/remove operations
6. WHEN editing a value THEN Redis Web Manager SHALL validate the input format before submission

### Requirement 5: Server Information

**User Story:** As a developer, I want to view Redis server statistics, so that I can monitor server health and resource usage.

#### Acceptance Criteria

1. WHEN a user opens the server info panel THEN Redis Web Manager SHALL display server version, memory usage, connected clients count, and uptime
2. WHEN a user views database statistics THEN Redis Web Manager SHALL display the number of keys in each database
3. WHEN server info is displayed THEN Redis Web Manager SHALL provide a refresh button to update the statistics

### Requirement 6: User Interface

**User Story:** As a developer, I want a clean and responsive interface with dark mode support, so that I can work comfortably for extended periods.

#### Acceptance Criteria

1. WHEN the application loads THEN Redis Web Manager SHALL display a layout with connection panel on the left and operation panel on the right
2. WHEN a user toggles dark mode THEN Redis Web Manager SHALL switch the entire interface to dark theme and persist the preference
3. WHEN performing any operation THEN Redis Web Manager SHALL provide visual feedback (loading spinners, success/error notifications)
4. WHEN the window is resized THEN Redis Web Manager SHALL adapt the layout responsively for desktop and tablet screens
5. WHEN displaying notifications THEN Redis Web Manager SHALL auto-dismiss success messages after 3 seconds and keep error messages until dismissed

### Requirement 7: Backend API

**User Story:** As a frontend application, I want a reliable REST API backend, so that I can execute Redis commands securely and efficiently.

#### Acceptance Criteria

1. WHEN receiving a connection test request THEN the backend SHALL validate connection parameters and return success or detailed error message
2. WHEN receiving a command execution request THEN the backend SHALL create a temporary Redis client, execute the command, and return the formatted result
3. WHEN processing requests THEN the backend SHALL handle CORS headers to allow frontend cross-origin requests
4. WHEN an invalid command is received THEN the backend SHALL return a structured error response with error type and message
5. WHEN the Redis server is unreachable THEN the backend SHALL return a timeout error within 5 seconds
6. WHEN receiving key listing requests THEN the backend SHALL use SCAN command with cursor-based pagination to avoid blocking

### Requirement 8: Performance and Security

**User Story:** As a system administrator, I want the application to be lightweight and secure, so that I can deploy it on resource-constrained environments safely.

#### Acceptance Criteria

1. WHILE the frontend is running THEN Redis Web Manager SHALL maintain memory usage below 50MB in typical usage scenarios
2. WHILE the backend is running THEN the backend service SHALL maintain memory usage below 50MB with no active connections
3. WHEN building the frontend THEN the build process SHALL produce a bundle size under 500KB (gzipped)
4. WHEN the backend receives a request THEN the backend SHALL not persist any connection credentials to disk or logs
5. WHEN establishing Redis connections THEN the backend SHALL use connection pooling with a maximum of 10 connections per request context
