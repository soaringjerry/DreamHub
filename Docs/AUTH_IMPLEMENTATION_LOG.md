# 鉴权功能实现日志 (Auth Implementation Log)

**版本/日期:** 2025-05-01

**目标:** 为 DreamHub 应用添加基于用户名/密码和 JWT 的用户认证和授权机制。

## 1. 架构与方案

*   **认证方式:** 用户名和密码。
*   **会话管理:** 使用 JSON Web Tokens (JWT) 进行无状态会话管理。Token 在登录成功后生成，并在后续请求中通过 `Authorization: Bearer <token>` 请求头传递。
*   **用户存储:** 在 PostgreSQL 数据库中新建 `users` 表存储用户信息（ID, 用户名, 密码哈希, 时间戳）。
*   **密码处理:** 使用 `bcrypt` 算法对用户密码进行哈希存储和验证。
*   **技术栈:**
    *   后端: Go (Gin 框架, pgx/v5, golang-jwt/jwt/v5, bcrypt)
    *   前端: React (TypeScript, Zustand, Axios, react-router-dom)
    *   数据库: PostgreSQL
    *   部署: Docker, Docker Compose, GitHub Actions

## 2. 后端实现 (`internal/`, `pkg/`, `cmd/`)

*   **数据库迁移 (`migrations/001_create_users_table.sql`):**
    *   创建了 `users` 表，包含 `id` (UUID), `username` (Unique), `password_hash`, `created_at`, `updated_at` 字段。
    *   添加了 `username` 索引和 `updated_at` 自动更新触发器。
*   **实体 (`internal/entity/user.go`):**
    *   定义了 `User` 结构体映射数据库表。
    *   定义了 `SanitizedUser` 结构体用于 API 响应，排除了密码哈希。
*   **仓库 (`internal/repository/user_repo.go`, `internal/repository/postgres/user_repo_impl.go`):**
    *   定义了 `UserRepository` 接口 (`CreateUser`, `GetUserByUsername`, `GetUserByID`)。
    *   实现了 PostgreSQL 版本的仓库，处理数据库交互和错误（如用户名冲突 `apperr.CodeConflict`）。
*   **错误处理 (`pkg/apperr/error.go`):**
    *   使用了现有的自定义错误包 `apperr`，定义了 `AppError` 结构和相关错误码（`CodeUnauthenticated`, `CodeInvalidArgument`, `CodeConflict`, `CodeInternal` 等）。
*   **配置 (`pkg/config/config.go`, `.env.example`):**
    *   在 `Config` 结构体中添加了 `JWTSecret` 和 `JWTExpirationMinutes` 字段。
    *   更新了 `LoadConfig` 函数以从环境变量加载这些值。
    *   在 `.env.example` 中添加了 `JWT_SECRET` 和 `JWT_EXPIRATION_MINUTES` 示例。
*   **认证服务 (`internal/service/auth_service.go`, `internal/service/auth_service_impl.go`):**
    *   定义了 `AuthService` 接口 (`Register`, `Login`, `ValidateToken`)。
    *   实现了服务逻辑：
        *   `Register`: 验证输入，使用 `bcrypt` 哈希密码，调用 `UserRepository` 创建用户。
        *   `Login`: 调用 `UserRepository` 获取用户，使用 `bcrypt` 验证密码，成功后使用 `golang-jwt/jwt/v5` 生成 JWT（包含 `user_id` 和过期时间）。
        *   `ValidateToken`: 解析并验证 JWT 签名和过期时间，返回用户 ID。
*   **API 处理器 (`internal/api/auth_handler.go`):**
    *   创建了 `AuthHandler`。
    *   实现了 `/api/v1/auth/register` 和 `/api/v1/auth/login` 的处理逻辑，调用 `AuthService` 并返回 JSON 响应。
*   **认证中间件 (`internal/api/auth_middleware.go`):**
    *   创建了 `AuthMiddleware`。
    *   实现了 `Authenticate()` Gin 中间件：
        *   从 `Authorization: Bearer <token>` 头提取 token。
        *   调用 `AuthService.ValidateToken` 验证 token。
        *   验证成功后，将 `userID` 存入 Gin 上下文 (`c.Set(authorizationPayloadKey, userID)`)。
        *   验证失败则中止请求并返回 401 错误。
    *   添加了辅助函数 `GetUserIDFromContext(c *gin.Context)` 从上下文中安全地获取用户 ID。
*   **主程序入口 (`cmd/server/main.go`):**
    *   初始化 `UserRepository`, `AuthService`, `AuthHandler`, `AuthMiddleware`。
    *   注册了 `/api/v1/auth` 路由组（公开）。
    *   创建了受保护的路由组 `/api/v1/`，并将 `AuthMiddleware` 应用于此组。
    *   将原有的 `ChatHandler` 和 `FileHandler` 的路由注册移至受保护的路由组下。
*   **依赖 (`go.mod`, `go.sum`):**
    *   添加了 `github.com/golang-jwt/jwt/v5` 依赖。
    *   运行 `go mod tidy` 整理依赖。
*   **现有处理器更新 (`internal/api/chat_handler.go`, `internal/api/file_handler.go`):**
    *   移除了从请求体/表单/查询参数中获取 `user_id` 的逻辑。
    *   改为使用 `api.GetUserIDFromContext(c)` 从 Gin 上下文中获取认证后的用户 ID。

## 3. 前端实现 (`frontend/`)

*   **API 服务 (`src/services/api.ts`):**
    *   添加了 `loginUser` 和 `registerUser` 函数，用于调用新的认证 API。
    *   修改了 `sendMessage` 和 `uploadFile` 函数，移除了 `userId` 参数。
    *   创建了专用的 Axios 实例 (`apiClient`)。
    *   添加了 Axios 请求拦截器，自动从 `localStorage` 读取 `authToken` 并添加到非认证请求的 `Authorization` 头中。
    *   导出了 `SanitizedUser` 接口。
*   **状态管理 (`src/store/authStore.ts`):**
    *   创建了基于 Zustand 的 `authStore`。
    *   定义了状态 (`isAuthenticated`, `user`, `token`, `isLoading`, `error`)。
    *   实现了 `login`, `logout`, `register` action，这些 action 会调用 `api.ts` 中的函数并更新状态。
    *   使用 `persist` 中间件将 `isAuthenticated`, `user`, `token` 持久化到 `localStorage`，以保持登录状态。
*   **路由 (`src/App.tsx`, `src/components/ProtectedRoute.tsx`):**
    *   安装了 `react-router-dom` 依赖。
    *   在 `App.tsx` 中设置了 `BrowserRouter` 和 `Routes`。
    *   创建了 `ProtectedRoute` 组件，检查 `authStore` 中的 `isAuthenticated` 状态，未登录则重定向到 `/login`。
    *   配置了路由规则：`/login`, `/register` 指向对应页面，根路径 `/` 受 `ProtectedRoute` 保护，渲染主应用布局 (`MainLayout`)。
*   **页面 (`src/pages/LoginPage.tsx`, `src/pages/RegisterPage.tsx`):**
    *   创建了登录和注册表单页面。
    *   使用 `authStore` 的 state 和 action 来处理用户输入、加载状态、错误显示和提交逻辑。
    *   使用 `react-router-dom` 的 `useNavigate` 在成功操作后进行页面跳转。
*   **UI 更新 (`src/App.tsx`):**
    *   移除了旧的手动设置 `userId` 的 UI 和逻辑。
    *   在页眉根据 `isAuthenticated` 状态动态显示：
        *   已登录: 显示用户名和登出按钮。
        *   未登录: 显示登录和注册链接。
*   **组件更新 (`src/components/FileUpload.tsx`):**
    *   移除了对 `userId` 的依赖，现在通过 `authStore` 和 `api.ts` 的拦截器处理认证。

## 4. CI/CD 更新 (`.github/workflows/ci-cd.yml`, `Dockerfile`, `docker-entrypoint.sh`)

*   **数据库迁移自动化:**
    *   修改了 `Dockerfile`，在构建阶段安装 `golang-migrate/migrate` CLI，并在最终镜像中包含该工具和 `migrations` 目录。
    *   创建了 `docker-entrypoint.sh` 脚本，该脚本在容器启动时首先执行 `migrate ... up` 命令应用数据库迁移，然后才启动 `supervisord`。
    *   更新了 `Dockerfile` 的 `ENTRYPOINT` 和 `CMD` 以使用此脚本。
*   **环境变量检查:**
    *   修改了 `deploy` 作业中的 SSH 脚本，在 `docker compose up` 之前增加了对服务器上 `.env` 文件中 `JWT_SECRET` 变量存在且非空的检查。如果检查失败，部署将中止。

## 5. 配置与启动

*   **数据库:** 需要运行数据库迁移脚本 (`migrations/001_create_users_table.sql`) 来创建 `users` 表。对于 CI/CD 环境，这已通过 `docker-entrypoint.sh` 自动化。
*   **环境变量:** 必须在部署环境（服务器上的 `.env` 文件）中设置一个强随机字符串作为 `JWT_SECRET` 的值。CI/CD 流程会检查此变量是否存在。
*   **启动:** 重新构建 Docker 镜像并使用 Docker Compose 启动更新后的服务。