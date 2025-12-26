# Requirements Document

## Introduction

Redis Web Manager V2 是对现有 Redis 可视化管理工具的重大升级。本次升级引入多 Tab 工作区界面、Redis 发布订阅功能、完整的用户认证系统（包括双重认证）、以及管理员控制面板。系统设计确保每个用户的数据完全隔离，管理员可以管理用户和系统设置。

## Glossary

- **Tab**: 工作区标签页，每个 Tab 包含独立的 Keys 列表和终端
- **Pub/Sub**: Redis 发布订阅机制，用于实时消息传递
- **Channel**: Pub/Sub 中的消息频道
- **Subscriber**: 订阅频道的客户端
- **Publisher**: 向频道发布消息的客户端
- **2FA (Two-Factor Authentication)**: 双重认证，使用 TOTP 算法
- **TOTP**: 基于时间的一次性密码算法
- **JWT**: JSON Web Token，用于用户会话管理
- **Admin**: 系统管理员，拥有用户管理权限
- **User**: 普通用户，只能访问自己的数据
- **Connection**: 用户保存的 Redis 连接配置
- **Workspace**: 用户的工作区，包含多个 Tab

## Requirements

### Requirement 1: Multi-Tab Workspace Interface

**User Story:** As a developer, I want to open multiple tabs for a single Redis connection, so that I can work on different tasks simultaneously without switching contexts.

#### Acceptance Criteria

1. WHEN a user connects to a Redis server THEN the System SHALL create a default Tab displaying the Keys list and Terminal
2. WHEN a user clicks the "New Tab" button THEN the System SHALL create a new Tab with an independent Keys list and Terminal instance
3. WHEN a user closes a Tab THEN the System SHALL remove the Tab and release associated resources
4. WHEN all Tabs are closed THEN the System SHALL display a default welcome content with connection information
5. WHEN a user switches between Tabs THEN the System SHALL preserve the state of each Tab including scroll position and selected keys
6. WHILE a Tab is active THEN the System SHALL display the Keys list in the upper section and Terminal in the lower section with a resizable divider
7. WHEN a user resizes the divider between Keys list and Terminal THEN the System SHALL persist the divider position for that Tab

### Requirement 2: Redis Pub/Sub Functionality

**User Story:** As a developer, I want to subscribe to Redis channels and publish messages, so that I can test and debug real-time messaging features.

#### Acceptance Criteria

1. WHEN a user opens the Pub/Sub panel THEN the System SHALL display a channel subscription interface and message history
2. WHEN a user subscribes to a channel pattern THEN the System SHALL establish a persistent subscription and display incoming messages in real-time
3. WHEN a message is received on a subscribed channel THEN the System SHALL display the channel name, message content, and timestamp
4. WHEN a user publishes a message to a channel THEN the System SHALL send the PUBLISH command and display confirmation
5. WHEN a user unsubscribes from a channel THEN the System SHALL terminate the subscription and stop receiving messages
6. WHILE subscribed to channels THEN the System SHALL maintain a WebSocket connection for real-time message delivery
7. WHEN the WebSocket connection is lost THEN the System SHALL attempt automatic reconnection and notify the user of connection status
8. WHEN a user clears the message history THEN the System SHALL remove all displayed messages while maintaining active subscriptions

### Requirement 3: User Authentication

**User Story:** As a system administrator, I want users to authenticate before accessing the system, so that I can ensure data security and user accountability.

#### Acceptance Criteria

1. WHEN an unauthenticated user accesses the application THEN the System SHALL redirect to the login page
2. WHEN a user submits valid credentials THEN the System SHALL issue a JWT token and redirect to the main application
3. WHEN a user submits invalid credentials THEN the System SHALL display an error message without revealing which field is incorrect
4. WHEN a JWT token expires THEN the System SHALL redirect the user to login and clear local session data
5. WHEN a user clicks logout THEN the System SHALL invalidate the session and redirect to the login page
6. WHILE a user is authenticated THEN the System SHALL include the JWT token in all API requests
7. WHEN the backend receives a request without valid JWT THEN the Backend SHALL return 401 Unauthorized status

### Requirement 4: User Registration

**User Story:** As a new user, I want to register an account, so that I can access the Redis management features.

#### Acceptance Criteria

1. WHEN registration is enabled AND a user visits the registration page THEN the System SHALL display a registration form
2. WHEN registration is disabled THEN the System SHALL hide the registration link and return 403 on registration attempts
3. WHEN a user submits a registration form with valid data THEN the System SHALL create the user account and redirect to login
4. WHEN a user submits a registration form with an existing username THEN the System SHALL display "Username already exists" error
5. WHEN a user submits a registration form with weak password THEN the System SHALL display password requirements
6. THE System SHALL require passwords to contain at least 8 characters, one uppercase letter, one lowercase letter, and one number

### Requirement 5: Two-Factor Authentication (2FA)

**User Story:** As a security-conscious user, I want to enable two-factor authentication, so that my account is protected even if my password is compromised.

#### Acceptance Criteria

1. WHEN a user enables 2FA THEN the System SHALL generate a TOTP secret and display a QR code for authenticator app setup
2. WHEN a user scans the QR code and enters a valid TOTP code THEN the System SHALL activate 2FA for the account
3. WHEN a user with 2FA enabled logs in with correct password THEN the System SHALL prompt for TOTP code before granting access
4. WHEN a user enters an invalid TOTP code THEN the System SHALL deny access and display an error message
5. WHEN a user disables 2FA THEN the System SHALL require current password and TOTP code verification before disabling
6. WHEN a user loses access to their authenticator THEN the System SHALL provide recovery codes generated during 2FA setup
7. THE System SHALL generate 10 single-use recovery codes when 2FA is enabled

### Requirement 6: User Data Isolation

**User Story:** As a user, I want my Redis connections and settings to be private, so that other users cannot access my data.

#### Acceptance Criteria

1. WHEN a user creates a connection THEN the System SHALL associate the connection with the user's account
2. WHEN a user queries connections THEN the System SHALL return only connections owned by that user
3. WHEN a user attempts to access another user's connection THEN the System SHALL return 403 Forbidden
4. WHEN a user deletes their account THEN the System SHALL remove all associated connections and settings
5. THE Backend SHALL validate user ownership for all connection-related operations

### Requirement 7: Admin User Management

**User Story:** As an administrator, I want to manage user accounts, so that I can maintain system security and user access.

#### Acceptance Criteria

1. WHEN an admin accesses the admin panel THEN the System SHALL display a list of all users with their status and role
2. WHEN an admin creates a new user THEN the System SHALL create the account with specified role and send credentials
3. WHEN an admin disables a user account THEN the System SHALL prevent that user from logging in
4. WHEN an admin enables a disabled user account THEN the System SHALL restore login capability
5. WHEN an admin resets a user's password THEN the System SHALL generate a temporary password and require change on next login
6. WHEN an admin deletes a user THEN the System SHALL remove the user and all associated data
7. WHEN a non-admin user attempts to access admin endpoints THEN the System SHALL return 403 Forbidden

### Requirement 8: Admin System Settings

**User Story:** As an administrator, I want to control system-wide settings, so that I can manage registration and security policies.

#### Acceptance Criteria

1. WHEN an admin toggles registration setting THEN the System SHALL enable or disable new user registration immediately
2. WHEN an admin views system settings THEN the System SHALL display current registration status and other configurable options
3. WHEN an admin changes a system setting THEN the System SHALL persist the change and apply it immediately
4. THE System SHALL store system settings in the database with appropriate defaults

### Requirement 9: Backend Authentication API

**User Story:** As the frontend application, I want secure authentication endpoints, so that I can implement user login and registration flows.

#### Acceptance Criteria

1. WHEN receiving a login request THEN the Backend SHALL validate credentials and return JWT token or error
2. WHEN receiving a registration request THEN the Backend SHALL validate input, create user, and return success or error
3. WHEN receiving a 2FA setup request THEN the Backend SHALL generate TOTP secret and return QR code data
4. WHEN receiving a 2FA verification request THEN the Backend SHALL validate TOTP code and complete authentication
5. WHEN processing authenticated requests THEN the Backend SHALL extract and validate JWT from Authorization header
6. THE Backend SHALL use bcrypt for password hashing with cost factor of 12
7. THE Backend SHALL set JWT expiration to 24 hours

### Requirement 10: Database Persistence

**User Story:** As a system, I want to persist user data and settings, so that data survives server restarts.

#### Acceptance Criteria

1. THE Backend SHALL use SQLite for data persistence in single-file deployment
2. WHEN the application starts THEN the Backend SHALL automatically create required database tables if they do not exist
3. THE Backend SHALL store users, connections, and system settings in separate tables
4. WHEN storing sensitive data THEN the Backend SHALL encrypt connection passwords using AES-256

