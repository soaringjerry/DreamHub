# CI/CD 配置日志 (GitHub Actions)

**日期:** 2025-04-26

**目标:** 为 DreamHub 项目设置一个完整的 CI/CD 流程，使用 GitHub Actions 实现自动化构建、测试、镜像推送和服务器部署。

## 主要步骤与挑战

1.  **需求确认:**
    *   明确了触发条件（推送 `main` 分支）、主要步骤（Lint, Test, Build Docker, Deploy）、部署目标（自有服务器 via SSH）和配置管理方式（服务器本地 `.env` 文件）。

2.  **前端测试框架集成 (Vitest):**
    *   安装了 Vitest 及相关依赖 (`@testing-library/react`, `jsdom` 等)。
    *   修改了 `vite.config.ts` 添加 Vitest 配置。
    *   在 `package.json` 中添加了 `test` 脚本。
    *   **挑战:**
        *   测试运行时缺少 `@vitest/coverage-v8` 依赖，已安装。
        *   测试因缺少 `setupTests.ts` 文件而失败，后创建该文件。
        *   测试因 `window.matchMedia` 在 jsdom 中未定义而失败，在 `setupTests.ts` 中添加了 mock 实现。
        *   修复了多处 ESLint 错误（未使用变量、未使用导入、`any` 类型警告等）。

3.  **Docker 配置:**
    *   创建了多阶段 `Dockerfile`，分别构建 Go 后端（server & worker）和 React 前端，最终使用 Alpine 基础镜像和 Supervisor 管理 Go 进程。
    *   创建了 `supervisord.conf` 配置文件。
    *   创建了 `.dockerignore` 文件优化构建。
    *   创建了 `docker-compose.yml` 文件，用于定义服务栈（app, postgres, redis）和配置（端口映射、卷、环境变量来源等）。
    *   **挑战:**
        *   `Dockerfile` 中 Go 版本与 `go.mod` 不匹配，已更新基础镜像。
        *   `docker-compose.yml` 中 `ankane/pgvector` 镜像标签错误，尝试了多个标签后最终确定使用 `latest`。
        *   服务器 5432 端口冲突，将 `docker-compose.yml` 中 PostgreSQL 的主机端口映射修改为 5433。

4.  **GitHub Actions Workflow (`.github/workflows/ci-cd.yml`):**
    *   定义了 `lint`, `test`, `build_and_push_image`, `deploy` 四个作业。
    *   配置了作业依赖关系 (`needs`)。
    *   配置了 Docker 镜像构建、登录 GHCR 并推送镜像。
    *   配置了部署步骤，使用 SSH 连接服务器。
    *   **挑战:**
        *   部署作业缺少 `actions/checkout` 步骤，导致找不到要复制的文件。
        *   文件复制步骤 (`appleboy/scp-action`) 反复失败，报错 `tar: empty archive`。尝试了拆分步骤、启用调试日志，最终替换为使用 `sshpass scp` 命令进行文件复制。
        *   Docker 镜像拉取因权限问题失败 (`denied: denied`)，确认镜像已设为 Public，并通过在服务器执行 `docker logout ghcr.io` 清除缓存凭证解决。
        *   部署脚本中关于是否启动 Docker PostgreSQL 的逻辑反复调整，最终确定为检查服务器 `.env` 文件中 `POSTGRES_USER/PASSWORD/DB` 是否都存在来决定是否激活 `docker-db` profile。

5.  **服务器端配置:**
    *   明确了部署脚本依赖服务器 `/root/dreamhub/` 目录下预先存在的 `.env` 文件来获取应用配置（数据库连接、API Key 等）。
    *   明确了需要手动将 `init_db.sql` 放置在服务器对应目录（后改为由 CI/CD 自动复制）。
    *   **挑战:**
        *   Go 应用容器内因 `.env` 文件缺少 `DATABASE_URL` 或 `DATABASE_URL` 配置错误（使用了 `localhost` 而不是服务名 `postgres`）导致无法连接数据库而启动失败。通过指导用户正确配置服务器上的 `.env` 文件解决。

6.  **后端静态文件服务:**
    *   检查发现 Go `server` 代码缺少提供前端静态文件的配置。
    *   修改了 `cmd/server/main.go`，添加了 `gin-contrib/static` 中间件和 `NoRoute` 处理器来服务 React 应用。
    *   使用 `go get` 添加了 `gin-contrib/static` 依赖。

## 最终成果

*   成功配置了 GitHub Actions CI/CD 流程。
*   实现了代码提交到 `main` 分支后自动进行 Lint、Test 和 Docker 镜像构建与推送。
*   实现了手动触发部署到服务器的功能。
*   部署流程能够根据服务器 `.env` 文件的配置，智能选择使用 Docker Compose 管理的数据库或连接外部数据库。
*   Go 后端配置为可以提供前端静态文件。

## 后续

*   需要用户在服务器 `/root/dreamhub/` 目录下正确配置 `.env` 文件。
*   需要用户在 GitHub Secrets 中配置 SSH 连接凭证。
*   建议编写更完善的前后端测试用例。