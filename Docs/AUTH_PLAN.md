# 鉴权登录功能实施计划

基于用户需求：使用用户名/密码认证，用户信息存储在 PostgreSQL 数据库，创建新的登录页面，并添加注册功能。

## 1. 数据库层面 (Database - PostgreSQL)

*   **设计 `users` 表**:
    *   在 PostgreSQL 数据库中创建一个名为 `users` 的新表。
    *   表结构应至少包含以下字段：
        *   `id` (UUID 或 SERIAL, Primary Key)
        *   `username` (VARCHAR, Unique, Not Null)
        *   `password_hash` (VARCHAR, Not Null) - 存储哈希后的密码
        *   `created_at` (TIMESTAMP WITH TIME ZONE, Default NOW())
        *   `updated_at` (TIMESTAMP WITH TIME ZONE, Default NOW())
*   **创建迁移脚本**:
    *   编写 SQL 脚本 (`migrations/00X_create_users_table.sql` 或类似文件) 来创建 `users` 表。
    *   考虑使用数据库迁移工具（如 `golang-migrate/migrate`）来管理数据库结构变更。

## 2. 后端层面 (Backend - Go)

*   **实体 (Entity)**:
    *   在 `internal/entity/` 目录下创建 `user.go` 文件，定义 `User` 结构体，映射 `users` 表。
*   **仓库 (Repository)**:
    *   在 `internal/repository/` 目录下创建 `user_repo.go` 文件，定义 `UserRepository` 接口，包含 `CreateUser` 和 `GetUserByUsername` 等方法。
    *   在 `internal/repository/postgres/` 目录下创建 `user_repo_impl.go` 文件，实现 `UserRepository` 接口，使用 `pgxpool` 与数据库交互。
*   **服务 (Service)**:
    *   在 `internal/service/` 目录下创建 `auth_service.go` 文件，定义 `AuthService` 接口，包含 `Register`, `Login`, `ValidateToken` 等方法。
    *   创建 `auth_service_impl.go` 文件，实现 `AuthService` 接口：
        *   **注册逻辑**: 接收用户名和密码，检查用户名是否已存在，对密码进行哈希处理（推荐使用 `golang.org/x/crypto/bcrypt`），然后调用 `UserRepository.CreateUser` 保存用户信息。
        *   **登录逻辑**: 接收用户名和密码，调用 `UserRepository.GetUserByUsername` 获取用户信息，使用 `bcrypt.CompareHashAndPassword` 验证密码，验证成功后生成 JWT（推荐使用 `github.com/golang-jwt/jwt/v5`）。
        *   **JWT 生成与验证**: 定义 JWT 的 Claims（至少包含用户 ID 和过期时间），使用安全的密钥（从配置中读取）进行签名和验证。
*   **API 处理器 (Handler)**:
    *   在 `internal/api/` 目录下创建 `auth_handler.go` 文件，定义 `AuthHandler` 结构体。
    *   实现处理 `/api/v1/register` 和 `/api/v1/login` 的方法，接收请求参数，调用 `AuthService` 的相应方法，并返回 JSON 响应（例如，登录成功返回 JWT）。
    *   在 `main.go` 中注册这些新的路由。
*   **中间件 (Middleware)**:
    *   在 `internal/api/` 目录下创建 `auth_middleware.go` 文件。
    *   实现一个 Gin 中间件函数 `Authenticate()`，从请求头 (`Authorization: Bearer <token>`) 中提取 JWT，调用 `AuthService.ValidateToken` 进行验证。
    *   验证成功后，可以将用户信息（如用户 ID）存入 Gin 的 Context (`c.Set("userID", userID)`)，供后续的 Handler 使用。
    *   验证失败则返回 `401 Unauthorized` 错误。
    *   在 `main.go` 中，将此中间件应用到需要保护的 API 路由组上（例如 `/api/v1/chat`, `/api/v1/upload`）。
*   **配置 (Config)**:
    *   在 `pkg/config/config.go` 和 `.env.example` / `.env` 文件中添加 JWT 相关的配置，如密钥 (`JWT_SECRET`) 和过期时间 (`JWT_EXPIRATION_MINUTES`)。
*   **依赖注入**:
    *   在 `cmd/server/main.go` 中：
        *   初始化 `UserRepository`。
        *   初始化 `AuthService`，并将 `UserRepository` 和 JWT 配置注入。
        *   初始化 `AuthHandler`，并将 `AuthService` 注入。
        *   初始化 `AuthMiddleware`，并将 `AuthService` 注入。
        *   注册 `AuthHandler` 的路由。
        *   将 `AuthMiddleware` 应用到需要认证的路由组。
        *   修改现有的 `ChatHandler` 和 `FileHandler`，让它们从 Gin Context 中获取 `userID`，而不是从请求体中获取。

## 3. 前端层面 (Frontend - React + TypeScript)

*   **页面与组件**:
    *   在 `frontend/src/` 下创建 `pages` 目录（如果尚不存在）。
    *   创建 `LoginPage.tsx`：包含用户名、密码输入框和登录按钮，处理表单提交，调用 `loginUser` API。
    *   创建 `RegisterPage.tsx`：包含用户名、密码、确认密码输入框和注册按钮，处理表单提交，调用 `registerUser` API。
*   **路由**:
    *   使用 `react-router-dom`（假设项目已使用）。
    *   在 `App.tsx` 或专门的路由配置文件中添加 `/login` 和 `/register` 路由，分别指向 `LoginPage` 和 `RegisterPage`。
    *   创建 `ProtectedRoute` 组件：检查用户是否已登录（通过状态管理），如果未登录，则使用 `<Navigate to="/login" replace />` 重定向到登录页；如果已登录，则渲染子组件 (`<Outlet />`)。
    *   将需要登录才能访问的页面（如聊天界面）包裹在 `ProtectedRoute` 中。
*   **API 服务 (`frontend/src/services/api.ts`)**:
    *   添加 `loginUser(credentials)` 和 `registerUser(userInfo)` 函数，分别调用后端的 `/api/v1/login` 和 `/api/v1/register` API。
    *   修改 `axios` 实例或创建拦截器 (interceptor)，在发送需要认证的请求（如 `sendMessage`, `uploadFile`）时，自动从状态管理或 localStorage 读取 JWT，并添加到 `Authorization` 请求头中。
    *   修改 `sendMessage` 和 `uploadFile` 函数，移除 `userId` 参数，因为后端会从 JWT 中获取。
*   **状态管理**:
    *   使用项目现有的状态管理库（如 Zustand `frontend/src/store/chatStore.ts`，可以扩展或创建新的 `authStore.ts`）或 React Context API。
    *   创建 `authStore` 或类似的状态切片，包含 `isAuthenticated` (boolean), `user` (object | null), `token` (string | null) 等状态。
    *   提供 `login(token, userData)` 和 `logout()` action/reducer：
        *   `login`: 更新状态，并将 token 存储到 `localStorage` (为了持久化登录状态)。
        *   `logout`: 清除状态，并从 `localStorage` 移除 token。
    *   在应用初始化时（例如 `App.tsx` 的 `useEffect`），检查 `localStorage` 中是否有 token，如果有且有效（可以添加一个简单的验证 API 调用），则调用 `login` action 恢复登录状态。
*   **UI 修改**:
    *   在 `App.tsx` 或导航组件中，根据 `isAuthenticated` 状态显示不同的内容（例如，显示登录/注册链接或用户信息/登出按钮）。
    *   修改 `ChatInterface.tsx` 等组件，不再需要传递 `userId` prop，而是从 `authStore` 中获取用户信息（如果需要显示用户名等）。

## 4. Mermaid 流程图

```mermaid
graph TD
    subgraph Frontend (React + TS)
        A[LoginPage] -- Credentials --> B(api.ts: loginUser);
        C[RegisterPage] -- UserInfo --> D(api.ts: registerUser);
        E[ChatInterface] -- Message --> F(api.ts: sendMessage);
        G[FileUpload] -- File --> H(api.ts: uploadFile);
        I[Router] --> A;
        I --> C;
        I --> J{ProtectedRoute};
        J -- Logged In --> E;
        J -- Logged In --> G;
        J -- Not Logged In --> A;
        K[Auth State (Store/Context)] <--> A;
        K <--> C;
        K <--> E;
        K <--> G;
        B -- JWT & UserData --> L[Store JWT & User State];
        L --> K;
        F -- Needs Auth --> M{Add JWT Header from State/Storage};
        H -- Needs Auth --> M;
        M --> N[Backend API];
        O[Logout Button] --> P[Clear Auth State & Storage];
        P --> K;
    end

    subgraph Backend (Go + Gin)
        N -- POST /api/v1/login --> Q[AuthHandler: Login];
        N -- POST /api/v1/register --> R[AuthHandler: Register];
        N -- POST /api/v1/chat --> S{AuthMiddleware};
        N -- POST /api/v1/upload --> S;
        Q -- Credentials --> T[AuthService: Login];
        R -- UserInfo --> U[AuthService: Register];
        S -- Valid JWT --> V[Extract userID from JWT];
        S -- Invalid JWT --> W[Return 401 Unauthorized];
        V --> X[ChatHandler/FileHandler (Use userID from Context)];
        T -- Valid Credentials --> Y[Generate JWT];
        T -- Invalid Credentials --> Z[Return Error];
        U -- Success --> AA[Return Success];
        U -- Failure --> Z;
        Y --> Q;
        AA --> R;
        T --> BB[UserRepository: FindUser];
        U --> CC[UserRepository: CreateUser];
        BB --> DD[DB: users table];
        CC --> DD;
    end

    subgraph Database (PostgreSQL)
        DD[users table (id, username, password_hash, ...)];
    end