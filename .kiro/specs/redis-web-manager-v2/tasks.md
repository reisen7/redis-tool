# Implementation Plan: Redis Web Manager V2

## Overview

本实现计划将 Redis Web Manager 升级到 V2 版本，引入用户认证系统、多 Tab 工作区、Pub/Sub 功能和管理员控制面板。实现采用增量方式，确保每个阶段都可以独立测试和验证。

## Tasks

## Phase 1: Database and Authentication Foundation

- [x] 1. Setup database models and migrations
  - [x] 1.1 Create user model and database schema
    - Define User struct with all fields (id, username, password_hash, role, totp_secret, etc.)
    - Implement GORM auto-migration
    - Create default admin user on first run
    - _Requirements: 10.2, 10.3_
  - [x] 1.2 Create connection model with user association
    - Define UserConnection struct with user_id foreign key
    - Implement GORM relationships
    - _Requirements: 6.1, 10.3_
  - [x] 1.3 Create system settings model
    - Define SystemSetting struct
    - Implement default settings initialization
    - _Requirements: 8.4, 10.3_
  - [x] 1.4 Create recovery codes model
    - Define RecoveryCode struct
    - _Requirements: 5.7_
  - [x] 1.5 Implement AES-256 encryption for connection passwords
    - Create encrypt/decrypt functions
    - _Requirements: 10.4_
  - [ ]* 1.6 Write property test for password encryption round-trip
    - **Property 17: Connection Password Encryption Round-Trip**
    - **Validates: Requirements 10.4**

- [x] 2. Implement authentication service
  - [x] 2.1 Implement password hashing with bcrypt
    - Create HashPassword and VerifyPassword functions
    - Use cost factor 12
    - _Requirements: 9.6_
  - [x] 2.2 Implement JWT token generation and validation
    - Create GenerateToken and ValidateToken functions
    - Set 24-hour expiration
    - _Requirements: 9.7, 3.2_
  - [x] 2.3 Implement TOTP generation and verification
    - Create GenerateTOTPSecret, GenerateQRCode, VerifyTOTP functions
    - _Requirements: 5.1, 5.2, 5.4_
  - [ ]* 2.4 Write property test for TOTP verification
    - **Property 11: Invalid TOTP Denied**
    - **Validates: Requirements 5.4**
  - [x] 2.5 Implement recovery code generation and verification
    - Generate 10 single-use codes
    - Hash codes before storage
    - _Requirements: 5.6, 5.7_

- [x] 3. Implement authentication middleware
  - [x] 3.1 Create JWT authentication middleware
    - Extract token from Authorization header
    - Validate token and set user context
    - Return 401 for invalid/missing token
    - _Requirements: 3.7, 9.5_
  - [ ]* 3.2 Write property test for protected endpoints
    - **Property 8: Protected Endpoints Require Valid JWT**
    - **Validates: Requirements 3.7**
  - [x] 3.3 Create admin role middleware
    - Check user role from context
    - Return 403 for non-admin users
    - _Requirements: 7.7_
  - [ ]* 3.4 Write property test for admin-only endpoints
    - **Property 15: Admin-Only Endpoints Enforce Role**
    - **Validates: Requirements 7.7**

- [x] 4. Checkpoint - Authentication foundation tests passing
  - Ensure all tests pass, ask the user if questions arise.

## Phase 2: Authentication API Handlers

- [x] 5. Implement authentication handlers
  - [x] 5.1 Implement login handler
    - Validate credentials
    - Check if user is enabled
    - Return JWT or 2FA required flag
    - _Requirements: 3.2, 3.3, 9.1_
  - [ ]* 5.2 Write property test for user enabled state
    - **Property 14: User Enabled State Affects Login**
    - **Validates: Requirements 7.3, 7.4**
  - [x] 5.3 Implement registration handler
    - Validate input and password strength
    - Check registration enabled setting
    - Create user account
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 9.2_
  - [ ]* 5.4 Write property test for registration disabled
    - **Property 9: Registration Disabled Returns 403**
    - **Validates: Requirements 4.2**
  - [ ]* 5.5 Write property test for password validation
    - **Property 10: Password Validation Rules**
    - **Validates: Requirements 4.5, 4.6**
  - [x] 5.6 Implement 2FA verification handler
    - Validate TOTP code
    - Issue JWT on success
    - _Requirements: 5.3, 5.4, 9.4_
  - [x] 5.7 Implement logout handler
    - Invalidate session (client-side token removal)
    - _Requirements: 3.5_

- [x] 6. Implement 2FA setup handlers
  - [x] 6.1 Implement 2FA setup initiation
    - Generate TOTP secret
    - Generate QR code
    - Generate recovery codes
    - _Requirements: 5.1, 5.7_
  - [x] 6.2 Implement 2FA activation
    - Verify initial TOTP code
    - Enable 2FA on account
    - _Requirements: 5.2_
  - [x] 6.3 Implement 2FA disable
    - Require password and TOTP verification
    - Disable 2FA on account
    - _Requirements: 5.5_

- [x] 7. Checkpoint - Authentication API tests passing
  - Ensure all tests pass, ask the user if questions arise.

## Phase 3: User and Connection Management

- [x] 8. Implement user service and handlers
  - [x] 8.1 Implement user CRUD service
    - Create, Read, Update, Delete operations
    - Password reset with temporary password
    - _Requirements: 7.2, 7.3, 7.4, 7.5, 7.6_
  - [x] 8.2 Implement admin user management handlers
    - GET /api/admin/users - list all users
    - POST /api/admin/users - create user
    - PUT /api/admin/users/:id - update user
    - DELETE /api/admin/users/:id - delete user
    - POST /api/admin/users/:id/reset-password
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 7.6_

- [x] 9. Implement connection service with user isolation
  - [x] 9.1 Modify connection service for user ownership
    - Add user_id to all queries
    - Validate ownership on all operations
    - _Requirements: 6.1, 6.2, 6.5_
  - [ ]* 9.2 Write property test for user data isolation
    - **Property 12: User Data Isolation**
    - **Validates: Requirements 6.2**
  - [ ]* 9.3 Write property test for cross-user access
    - **Property 13: Cross-User Access Denied**
    - **Validates: Requirements 6.3**
  - [x] 9.4 Implement connection handlers with auth
    - GET /api/connections - list user's connections
    - POST /api/connections - create connection
    - PUT /api/connections/:id - update connection
    - DELETE /api/connections/:id - delete connection
    - POST /api/connections/:id/test - test connection
    - _Requirements: 6.1, 6.2, 6.3_

- [x] 10. Implement settings service and handlers
  - [x] 10.1 Implement settings service
    - Get/Set settings
    - Default values initialization
    - _Requirements: 8.2, 8.3, 8.4_
  - [ ]* 10.2 Write property test for settings persistence
    - **Property 16: Settings Persistence and Immediate Effect**
    - **Validates: Requirements 8.1, 8.3**
  - [x] 10.3 Implement admin settings handlers
    - GET /api/admin/settings
    - PUT /api/admin/settings
    - _Requirements: 8.1, 8.2, 8.3_

- [x] 11. Checkpoint - User and connection management tests passing
  - Ensure all tests pass, ask the user if questions arise.

## Phase 4: Pub/Sub Backend Implementation

- [x] 12. Implement Pub/Sub WebSocket service
  - [x] 12.1 Create WebSocket hub for client management
    - Client connection/disconnection handling
    - Message broadcasting
    - _Requirements: 2.6_
  - [x] 12.2 Implement Redis subscription manager
    - Subscribe/unsubscribe to channels
    - Forward messages to WebSocket clients
    - _Requirements: 2.2, 2.5_
  - [x] 12.3 Implement WebSocket handler
    - Authenticate WebSocket connection
    - Handle subscribe/unsubscribe messages
    - Handle connection errors and reconnection
    - _Requirements: 2.6, 2.7_
  - [x] 12.4 Implement publish handler
    - POST /api/pubsub/publish
    - _Requirements: 2.4_

- [x] 13. Checkpoint - Pub/Sub backend tests passing
  - Ensure all tests pass, ask the user if questions arise.

## Phase 5: Frontend Authentication

- [x] 14. Setup frontend authentication infrastructure
  - [x] 14.1 Create auth store (Pinia)
    - User state, token management
    - Login/logout actions
    - _Requirements: 3.2, 3.5, 3.6_
  - [ ]* 14.2 Write property test for JWT in requests
    - **Property 7: JWT Included in Authenticated Requests**
    - **Validates: Requirements 3.6**
  - [x] 14.3 Create auth API module
    - Login, register, verify2FA, logout endpoints
    - _Requirements: 3.2, 4.3, 5.3_
  - [x] 14.4 Configure axios interceptor for JWT
    - Add Authorization header to all requests
    - Handle 401 responses
    - _Requirements: 3.6, 3.4_

- [x] 15. Implement authentication views
  - [x] 15.1 Create LoginView component
    - Username/password form
    - Error handling
    - _Requirements: 3.1, 3.2, 3.3_
  - [x] 15.2 Create RegisterView component
    - Registration form with validation
    - Password strength indicator
    - _Requirements: 4.1, 4.3, 4.4, 4.5_
  - [x] 15.3 Create TwoFactorView component
    - TOTP code input
    - Recovery code option
    - _Requirements: 5.3, 5.4, 5.6_
  - [x] 15.4 Create TotpSetup component
    - QR code display
    - Verification input
    - Recovery codes display
    - _Requirements: 5.1, 5.2, 5.7_

- [x] 16. Setup Vue Router with auth guards
  - [x] 16.1 Configure routes
    - Public routes: /login, /register
    - Protected routes: /app, /admin
    - _Requirements: 3.1_
  - [x] 16.2 Implement navigation guards
    - Redirect unauthenticated users to login
    - Redirect non-admin from admin routes
    - _Requirements: 3.1, 7.7_

- [x] 17. Checkpoint - Frontend authentication tests passing
  - Ensure all tests pass, ask the user if questions arise.

## Phase 6: Multi-Tab Workspace

- [-] 18. Implement tab management
  - [x] 18.1 Create tabs store (Pinia)
    - Tab state management
    - Active tab tracking
    - Tab state persistence
    - _Requirements: 1.1, 1.2, 1.3, 1.5_
  - [ ]* 18.2 Write property test for tab state persistence
    - **Property 1: Tab State Persistence**
    - **Validates: Requirements 1.5, 1.7**
  - [ ]* 18.3 Write property test for tab creation
    - **Property 2: Tab Creation Increases Count**
    - **Validates: Requirements 1.2**
  - [ ]* 18.4 Write property test for tab removal
    - **Property 3: Tab Removal Decreases Count**
    - **Validates: Requirements 1.3**

- [x] 19. Implement tab UI components
  - [x] 19.1 Create TabBar component
    - Tab list display
    - New tab button
    - Close tab button
    - Tab switching
    - _Requirements: 1.2, 1.3, 1.5_
  - [x] 19.2 Create TabContent component
    - Container for active tab content
    - Default welcome content when no tabs
    - _Requirements: 1.4_
  - [x] 19.3 Create WorkspacePanel component
    - Keys list (top section)
    - Terminal (bottom section)
    - Resizable divider
    - _Requirements: 1.6, 1.7_
  - [x] 19.4 Create ResizableDivider component
    - Drag to resize
    - Persist position per tab
    - _Requirements: 1.7_

- [x] 20. Integrate tabs with existing components
  - [x] 20.1 Modify KeyList for tab context
    - Use tab-specific state
    - _Requirements: 1.5_
  - [x] 20.2 Modify RedisTerminal for tab context
    - Use tab-specific history
    - _Requirements: 1.5_

- [x] 21. Checkpoint - Multi-tab workspace tests passing
  - Ensure all tests pass, ask the user if questions arise.

## Phase 7: Pub/Sub Frontend

- [-] 22. Implement Pub/Sub frontend
  - [x] 22.1 Create pubsub store (Pinia)
    - Subscriptions list
    - Message history
    - WebSocket connection state
    - _Requirements: 2.2, 2.3, 2.5, 2.8_
  - [ ]* 22.2 Write property test for subscription management
    - **Property 4: Subscription Management Consistency**
    - **Validates: Requirements 2.2, 2.5**
  - [ ]* 22.3 Write property test for message display
    - **Property 5: Message Display Completeness**
    - **Validates: Requirements 2.3**
  - [ ]* 22.4 Write property test for clear messages
    - **Property 6: Clear Messages Preserves Subscriptions**
    - **Validates: Requirements 2.8**
  - [x] 22.5 Create useWebSocket composable
    - WebSocket connection management
    - Auto-reconnection with backoff
    - _Requirements: 2.6, 2.7_

- [x] 23. Implement Pub/Sub UI components
  - [x] 23.1 Create PubSubPanel component
    - Main container for Pub/Sub UI
    - _Requirements: 2.1_
  - [x] 23.2 Create SubscriptionList component
    - Display active subscriptions
    - Subscribe/unsubscribe controls
    - _Requirements: 2.2, 2.5_
  - [x] 23.3 Create MessageHistory component
    - Display messages with channel, content, timestamp
    - Clear history button
    - _Requirements: 2.3, 2.8_
  - [x] 23.4 Create PublishForm component
    - Channel and message input
    - Publish button
    - _Requirements: 2.4_

- [x] 24. Checkpoint - Pub/Sub frontend tests passing
  - Ensure all tests pass, ask the user if questions arise.

## Phase 8: Admin Panel

- [x] 25. Implement admin panel
  - [x] 25.1 Create admin API module
    - User management endpoints
    - Settings endpoints
    - _Requirements: 7.1, 7.2, 8.2_
  - [x] 25.2 Create AdminView component
    - Admin panel layout
    - Navigation between sections
    - _Requirements: 7.1_
  - [x] 25.3 Create UserList component
    - Display all users
    - Status and role indicators
    - Action buttons
    - _Requirements: 7.1_
  - [x] 25.4 Create UserForm component
    - Create/edit user form
    - Role selection
    - _Requirements: 7.2_
  - [x] 25.5 Create SettingsPanel component
    - Registration toggle
    - Other system settings
    - _Requirements: 8.1, 8.2_

- [x] 26. Checkpoint - Admin panel tests passing
  - Ensure all tests pass, ask the user if questions arise.

## Phase 9: Integration and Final Testing

- [x] 27. Integration and polish
  - [x] 27.1 Update main App.vue layout
    - Integrate new components
    - Update navigation
    - _Requirements: 6.1_
  - [x] 27.2 Update connection sidebar for authenticated mode
    - Load connections from backend
    - _Requirements: 6.2_
  - [x] 27.3 Add user profile section
    - Display current user
    - 2FA settings link
    - Logout button
    - _Requirements: 3.5, 5.1_

- [x] 28. Final Checkpoint - All tests passing
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional property-based tests and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation
- Property tests validate universal correctness properties
- Unit tests validate specific examples and edge cases
- Backend uses Go with Gin framework
- Frontend uses Vue 3 with TypeScript and Pinia
- Property-based testing: fast-check (frontend), testing/quick (backend)
