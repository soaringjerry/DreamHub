-# DreamHub 后端开发综合计划
-
-基于 `README.md` 和 `PLAN.md` 的分析，制定以下后端开发综合计划，涵盖核心功能、基础架构、安全性、DevOps、开发者体验和未来增强等方面。
-
-## 计划概览图 (Mermaid)
-
-```mermaid
-graph TD
-    subgraph "核心功能 (Core Features)"
-        direction LR
-        F1[1.1 用户数据隔离]
-        F2[1.2 异步任务处理]
-        F3[1.3 对话历史与摘要]
-        F4[1.4 WebSocket 实时通信]
-    end
-
-    subgraph "基础架构与健壮性 (Foundation)"
-        direction LR
-        B1[2.1 统一错误处理]
-        B2[2.2 统一日志记录]
-        B3[2.3 外部API韧性]
-        B4[2.4 数据一致性]
-    end
-
-    subgraph "安全性 (Security)"
-        direction LR
-        S1[3.1 身份验证与授权]
-        S2[3.2 Secrets 管理]
-        S3[3.3 文件上传安全]
-    end
-
-    subgraph "DevOps 与生产就绪 (DevOps & Prod Ready)"
-        direction TB
-        D1[4.1 CI/CD 流程]
-        D2[4.2 容器化]
-        D3[4.3 监控与告警]
-        D4[4.4 部署策略]
-        D5[4.5 数据管理]
-        D6[4.6 合规性]
-    end
-
-    subgraph "开发者体验 (DevEx)"
-        direction LR
-        E1[5.1 本地环境]
-        E2[5.2 代码质量]
-        E3[5.3 自动化测试]
-        E4[5.4 辅助脚本]
-    end
-
-    subgraph "未来增强 (Future)"
-        direction LR
-        O1[6.1 插件化 RAG]
-        O2[6.2 微调流水线]
-        O3[6.3 可视化面板]
-    end
-
-    %% 主要依赖关系
-    F1 & F2 & F3 & F4 --> B1 & B2 & B3 & B4 & S1 & S2 & S3 & D1 & D2 & D3 & D4 & D5 & D6 & E1 & E2 & E3 & E4
-
-    classDef core fill:#cde4ff,stroke:#36c,stroke-width:2px;
-    classDef foundation fill:#ddf1d4,stroke:#383,stroke-width:2px;
-    classDef security fill:#fff0cc,stroke:#a50,stroke-width:2px;
-    classDef devops fill:#e4ccf9,stroke:#539,stroke-width:2px;
-    classDef devex fill:#f9d4e1,stroke:#b27,stroke-width:2px;
-    classDef future fill:#d4f1f9,stroke:#05a,stroke-width:2px;
-
-    class F1,F2,F3,F4 core;
-    class B1,B2,B3,B4 foundation;
-    class S1,S2,S3 security;
-    class D1,D2,D3,D4,D5,D6 devops;
-    class E1,E2,E3,E4 devex;
-    class O1,O2,O3 future;
-
-```
-
-## 1. 核心功能实现与完善 (Core Features)
-
-*   **1.1 用户数据隔离 (Data Isolation):**
-    *   在 Repository 层强制加入 `user_id` / `tenant_id` 过滤。
-    *   建立 `pkg/ctxutil` 包，规范 Context 传递 `user_id`, `trace_id` 等。
-    *   确保向量库和对话历史的严格隔离。
-*   **1.2 异步任务处理 (Async Task Processing):**
-    *   使用 Asynq + Redis 实现任务队列。
-    *   实现 Worker 处理 Embedding (分块、调用 LLM、存储)。
-    *   创建 `task` 实体和 `task_repo` 用于状态管理。
-    *   提供 `/api/v1/tasks/{task_id}/status` API。
-    *   实现任务幂等性 (基于 `file_hash` 或 `chunk_hash`)。
-    *   优化大文件处理 (分段事务或批处理)。
-    *   增加队列保护机制 (如 `MAXLEN`) 防止爆仓。
-*   **1.3 对话历史与摘要 (Conversation & Summarization):**
-    *   创建 `chat_repo` 并迁移数据库逻辑。
-    *   实现 `memory_service` 进行对话摘要。
-    *   设计数据库表存储摘要和管理长对话上下文。
-*   **1.4 WebSocket 实时通信 (Real-time Communication):**
-    *   搭建 WebSocket 服务器 (Hub 管理连接, 认证)。
-    *   Worker 通过 Hub 推送任务进度。
-    *   Chat Service 支持流式响应并通过 WebSocket 发送。(TODO 2025-04-26: 当前 `openai_provider.go` 中的流式实现是临时的，因 `langchaingo` API 问题待查，暂时使用非流式替代。)
-    *   考虑 Hub 的高可用性 (如 Redis Pub/Sub)。
-
-## 2. 基础架构与健壮性 (Foundation & Robustness)
-
-*   **2.1 统一错误处理 (Error Handling):**
-    *   完善 `pkg/apperr` 定义应用错误。
-    *   使用 Gin 中间件统一捕获和格式化错误响应。
-*   **2.2 统一日志记录 (Logging):**
-    *   完善 `pkg/logger` (如使用 slog/zap)。
-    *   实现结构化日志，并在关键路径添加 Trace ID。
-*   **2.3 外部 API 调用韧性 (External API Resilience):**
-    *   为 OpenAI 等调用添加指数退避重试、速率限制。
-    *   实现 API Key 轮换和额度监控与告警。
-*   **2.4 数据一致性 (Data Consistency):**
-    *   在 Service 层协调 PostgreSQL 和 PGVector 的写入，使用事务或 Outbox Pattern。
-
-## 3. 安全性 (Security)
-
-*   **3.1 身份验证与授权 (AuthN & AuthZ):**
-    *   实现基于 Token (如 JWT) 的用户认证。
-    *   在 API 和 WebSocket 入口强制执行认证。
-*   **3.2 Secrets 管理 (Secrets Management):**
-    *   使用专用工具 (Vault, AWS Secrets Manager 等) 或安全机制管理敏感信息。
-    *   实现密钥轮换。
-*   **3.3 文件上传安全 (Secure File Upload):**
-    *   实现 MIME 类型白名单过滤。
-    *   (推荐) 集成病毒扫描 (如 ClamAV)。
-
-## 4. DevOps 与生产就绪 (DevOps & Production Readiness)
-
-*   **4.1 CI/CD 流程 (CI/CD Pipeline):**
-    *   使用 GitHub Actions 自动化测试、构建 Docker 镜像、推送到仓库 (如 GHCR)。
-    *   实现手动触发部署到测试服务器。
-*   **4.2 容器化 (Containerization):**
-    *   为 `server` 和 `worker` 编写 Dockerfile。
-    *   创建 `docker-compose.yml` 用于本地开发和测试服务器部署。
-*   **4.3 监控与告警 (Monitoring & Alerting):**
-    *   (推荐) 集成 OpenTelemetry (Tracing) + Prometheus (Metrics) + Grafana (Dashboard)。
-    *   监控关键指标 (队列深度、延迟、错误率、资源使用、成本)。
-    *   设置告警规则 (如队列积压、额度不足)。
-*   **4.4 部署策略 (Deployment Strategy):**
-    *   为应用和数据添加版本管理。
-    *   (推荐) 采用蓝绿部署或 Canary 发布策略。
-*   **4.5 数据管理 (Data Management):**
-    *   (推荐) 建立数据库备份和恢复演练流程。
-    *   优化数据库性能 (索引、Vacuum)。
-*   **4.6 合规性 (Compliance):**
-    *   (如果需要) 实现 GDPR 等法规要求的数据删除流程。
-
-## 5. 开发者体验 (Developer Experience)
-
-*   **5.1 本地开发环境:** 确保 `docker-compose up` 能快速启动所有依赖。
-*   **5.2 代码质量:** (推荐) 使用 `pre-commit` hooks 运行 linters 和格式化工具。
-*   **5.3 自动化测试:** 编写单元测试，(推荐) 建立端到端测试流程。
-*   **5.4 辅助脚本:** 提供 `make dev-up`, `scripts/post-merge.sh` 等提高效率。
-
-## 6. (可选) 未来增强 (Future Enhancements)
-
-*   插件化 RAG 数据源。
-*   模型微调流水线。
-*   可视化监控面板 (任务、知识库增长)。

# DreamHub CI/CD 详细计划 (GitHub Actions)

## 目标

为 DreamHub 项目创建一个 GitHub Actions CI/CD 流程，该流程在每次推送到 `main` 分支时触发，执行以下操作：

1.  运行 Go 后端和 React 前端的 Linting（可选但推荐）。
2.  运行 Go 后端和 React 前端（使用 Vitest）的测试。
3.  构建一个包含 Go server、Go worker 和 React 前端静态文件的 Docker 镜像，并使用 Supervisor 管理 Go 进程。
4.  将构建好的 Docker 镜像推送到 GitHub Container Registry (GHCR)。
5.  通过 SSH 连接到目标服务器，拉取最新的 Docker 镜像，并使用 `docker compose up -d`（假设服务器上已有 `docker-compose.yml`）重新部署应用。

## 准备工作 (需要在本地完成)

1.  **添加前端测试框架 (Vitest):**
    *   进入 `frontend` 目录: `cd frontend`
    *   安装 Vitest: `npm install --save-dev vitest @vitest/ui @testing-library/react @testing-library/jest-dom jsdom`
    *   配置 Vitest: 在 `frontend/vite.config.ts` 中添加 Vitest 配置，或者创建一个 `frontend/vitest.config.ts` 文件。
        ```typescript
        // Example vite.config.ts modification
        import { defineConfig } from 'vite'
        import react from '@vitejs/plugin-react'
        import type { UserConfig } from 'vitest/config' // Import UserConfig type

        // https://vitejs.dev/config/
        export default defineConfig({
          plugins: [react()],
          test: { // Add this test configuration
            globals: true,
            environment: 'jsdom',
            setupFiles: './src/setupTests.ts', // Optional setup file
            css: true,
          } as UserConfig['test'], // Cast to UserConfig['test']
        })
        ```
    *   (可选) 创建 `frontend/src/setupTests.ts` 文件用于测试设置。
    *   在 `frontend/package.json` 的 `scripts` 部分添加 `test` 脚本:
        ```json
        "scripts": {
          // ... other scripts
          "test": "vitest",
          "test:ui": "vitest --ui", // Optional for local UI testing
          "coverage": "vitest run --coverage" // Optional for coverage
        },
        ```
    *   编写前端测试用例。

2.  **创建 Dockerfile:**
    *   在项目根目录 (`f:/Download/DreamHub`) 创建 `Dockerfile` 文件。
        ```dockerfile
        # Stage 1: Build Go Backend (Server & Worker)
        FROM golang:1.22-alpine AS builder
        WORKDIR /app
        COPY go.mod go.sum ./
        RUN go mod download
        COPY . .
        # Build server
        RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /server cmd/server/main.go
        # Build worker
        RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /worker cmd/worker/main.go

        # Stage 2: Build Frontend
        FROM node:20-alpine AS frontend-builder
        WORKDIR /app/frontend
        COPY frontend/package.json frontend/package-lock.json* ./
        # Use ci for potentially faster and more reliable installs in CI
        RUN npm ci
        COPY frontend/ ./
        # Add build-time args if needed, e.g., ARG VITE_API_URL
        # RUN npm run build -- --base=./ --mode production -- VITE_API_URL=$VITE_API_URL
        RUN npm run build

        # Stage 3: Final Image
        FROM alpine:latest
        # Install supervisor and any other runtime dependencies
        RUN apk update && apk add --no-cache supervisor ca-certificates tzdata && rm -rf /var/cache/apk/*
        WORKDIR /app
        # Copy Go binaries from builder stage
        COPY --from=builder /server /app/server
        COPY --from=builder /worker /app/worker
        # Copy frontend build from frontend-builder stage
        # Assuming the server serves frontend files from a 'public' or 'static' directory
        COPY --from=frontend-builder /app/frontend/dist /app/frontend/dist
        # Copy supervisor config
        COPY supervisord.conf /etc/supervisor/conf.d/supervisord.conf
        # Copy .env.example for reference, but actual .env should be mounted or managed externally on the server
        # COPY .env.example .env.example

        # Expose the port the server listens on (adjust if needed)
        EXPOSE 8080

        # Set the entrypoint to supervisor
        ENTRYPOINT ["/usr/bin/supervisord", "-c", "/etc/supervisor/conf.d/supervisord.conf"]
        # CMD is defined in supervisord.conf
        ```

3.  **创建 Supervisor 配置:**
    *   在项目根目录创建 `supervisord.conf` 文件，用于管理 server 和 worker 进程：
        ```ini
        [supervisord]
        nodaemon=true ; Run in foreground for Docker

        [program:server]
        command=/app/server
        autostart=true
        autorestart=true
        stderr_logfile=/dev/stderr
        stderr_logfile_maxbytes=0
        stdout_logfile=/dev/stdout
        stdout_logfile_maxbytes=0
        # Add environment variables if needed directly here, or manage via .env file mount
        # environment=VAR1="value1",VAR2="value2"

        [program:worker]
        command=/app/worker
        autostart=true
        autorestart=true
        stderr_logfile=/dev/stderr
        stderr_logfile_maxbytes=0
        stdout_logfile=/dev/stdout
        stdout_logfile_maxbytes=0
        # environment=VAR1="value1",VAR2="value2"
        ```

4.  **创建 `.dockerignore`:**
    *   在项目根目录创建 `.dockerignore` 文件，排除不需要复制到 Docker 镜像中的文件/目录：
        ```
        .git
        .vscode
        frontend/node_modules
        uploads/*
        *.md
        # Add other files/dirs to ignore
        ```

5.  **(可选) 创建 `docker-compose.yml`:**
    *   在项目根目录创建 `docker-compose.yml`，方便本地测试和服务器部署。这个文件应该与服务器上的文件保持一致或作为其基础。
        ```yaml
        version: '3.8'

        services:
          app:
            image: your-local-image-name:latest # Use local name for testing
            # build: . # Uncomment to build locally
            restart: always
            ports:
              - "8080:8080" # Map host port to container port
            env_file:
              - .env # Mount the .env file from the host
            volumes:
              - ./uploads:/app/uploads # Mount uploads directory if needed
            # Add dependencies like database if needed
            # depends_on:
            #   - db

        # Example database service (if needed)
        # db:
        #   image: postgres:15-alpine
        #   restart: always
        #   environment:
        #     POSTGRES_USER: ${DB_USER}
        #     POSTGRES_PASSWORD: ${DB_PASSWORD}
        #     POSTGRES_DB: ${DB_NAME}
        #   volumes:
        #     - postgres_data:/var/lib/postgresql/data
        #     - ./init_db.sql:/docker-entrypoint-initdb.d/init.sql # Optional init script
        #   ports:
        #     - "5432:5432"

        # volumes:
        #   postgres_data:
        ```

## GitHub Actions Workflow (`.github/workflows/ci-cd.yml`)

```yaml
name: DreamHub CI/CD

on:
  push:
    branches: [ main ]
  workflow_dispatch: # Allows manual triggering

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22' # Match your project's Go version

      - name: Run Go Lint # Replace with your preferred linter, e.g., golangci-lint
        run: go vet ./... # Example, use golangci-lint run ./... if configured

      - name: Set up Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20' # Match your project's Node version
          cache: 'npm'
          cache-dependency-path: frontend/package-lock.json

      - name: Install Frontend Dependencies
        run: npm ci --prefix frontend

      - name: Run Frontend Lint
        run: npm run lint --prefix frontend

  test:
    runs-on: ubuntu-latest
    needs: lint # Optional: Run tests only if lint passes
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Run Go Tests
        run: go test -v ./... # Add coverage flags if needed, e.g., -coverprofile=coverage.out

      # Optional: Upload Go coverage report
      # - name: Upload Go coverage to Codecov
      #   uses: codecov/codecov-action@v4
      #   with:
      #     token: ${{ secrets.CODECOV_TOKEN }} # Store Codecov token in secrets
      #     files: ./coverage.out
      #     flags: go

      - name: Set up Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
          cache-dependency-path: frontend/package-lock.json

      - name: Install Frontend Dependencies
        run: npm ci --prefix frontend

      - name: Run Frontend Tests (Vitest)
        run: npm test -- --run --coverage --prefix frontend # Use --run for non-watch mode in CI

      # Optional: Upload Vitest coverage report
      # - name: Upload Vitest coverage to Codecov
      #   uses: codecov/codecov-action@v4
      #   with:
      #     token: ${{ secrets.CODECOV_TOKEN }}
      #     flags: frontend # Flag to distinguish coverage reports
      #     working-directory: ./frontend # Specify working directory for coverage file search

  build_and_push_image:
    runs-on: ubuntu-latest
    needs: test # Run build only if tests pass
    permissions:
      contents: read
      packages: write # Needed to push to GHCR
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Log in to GitHub Container Registry
        uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }} # Use the default GITHUB_TOKEN

      - name: Build and push Docker image
        uses: docker/build-push-action@v6
        with:
          context: .
          file: ./Dockerfile
          push: true
          tags: |
            ghcr.io/${{ github.repository_owner }}/dreamhub:latest
            ghcr.io/${{ github.repository_owner }}/dreamhub:${{ github.sha }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
          # Add build args if needed, getting values from secrets
          # build-args: |
          #   VITE_API_URL=${{ secrets.VITE_API_URL }}

  deploy:
    runs-on: ubuntu-latest
    needs: build_and_push_image # Run deploy only after image is pushed
    environment: production # Optional: Define a GitHub environment for deployment secrets/rules
    steps:
      - name: Deploy to Server via SSH
        uses: appleboy/ssh-action@v1.0.3
        with:
          host: ${{ secrets.SSH_HOST }} # Server IP or hostname
          username: ${{ secrets.SSH_USER }} # SSH username
          key: ${{ secrets.SSH_PRIVATE_KEY }} # SSH private key
