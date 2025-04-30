# DreamHub

DreamHub 是一个 AI 驱动的工作站/工作面板后端服务，旨在通过集成个人知识库和对话记忆，提高信息处理和交互效率。

当前版本（经过重构）包含两个主要组件：**API Server** 和 **Worker**。

## 核心功能

*   **文件上传与异步处理:** API Server 接收上传的文件并保存（目前为本地存储），然后将处理任务（文本分块、向量化）放入 Asynq 任务队列 (Redis)。Worker 进程负责后台消费队列并处理这些任务。API Server 立即返回文档 ID 和任务 ID。
*   **个人知识库 (RAG):** Worker 将处理后的文档块向量存入向量数据库 (PostgreSQL + pgvector)。Repository 层在查询时强制进行用户隔离过滤（基于 `cmetadata` 中的 `user_id`），支持基于用户文件内容的智能问答。
*   **对话历史记忆:** 记录多轮对话上下文（基于 `conversation_id` 和 `user_id`），存储在 PostgreSQL 中，通过 `ChatRepository` 进行管理，实现更连贯的 AI 交互。
*   **基础 API:** 提供文件上传 (`/api/v1/upload`)、聊天交互 (`/api/v1/chat`)、文档管理 (`/api/v1/documents`) 和任务状态查询 (`/api/v1/tasks/{task_id}/status`) 的 API 端点。详细 API 文档请参考 [API_DOCS.md](Docs/API_DOCS.md)。
*   **前端界面:** 提供了一个基于 React 的用户界面，用于文件上传、聊天交互和对话管理。详细前端文档请参考 [frontend/FRONTEND_DOCS.md](frontend/FRONTEND_DOCS.md)。

## 文档说明

本项目的所有详细文档，包括 API 文档、前端文档、开发计划、进度日志等，均已整理并移动到根目录下的 `Docs/` 文件夹中。请在该目录查找最新的详细信息。
## 系统架构

```mermaid
graph TD
    subgraph "用户端"
        Client[前端 React App]
    end

    subgraph "后端服务"
        direction LR
        subgraph "API Server (Gin)"
            Router[路由 /api/v1] --> Middleware[中间件 Log/Err/Auth?]
            Middleware --> Handler[API Handlers]
            Handler --> Service[业务逻辑 Services]
        end
        subgraph "Async Worker (Asynq)"
            WorkerServer[Asynq Server] --> TaskHandler[任务处理器 Embedding]
            TaskHandler --> ServiceWorker[业务逻辑 Services]
        end
    end

    subgraph "核心服务层"
        Service --> Repo[数据仓库 Repositories]
        Service --> TaskQueue[(Asynq Client)]
        Service --> FileStore[文件存储 Local/S3]
        Service --> LLMProvider[LLM Provider OpenAI]
        ServiceWorker --> Repo
        ServiceWorker --> FileStore
        ServiceWorker --> EmbeddingProvider[Embedding Provider OpenAI]
        %% ServiceWorker --> Notifier[通知服务 WS/SSE?]
    end

    subgraph "数据与依赖"
        Repo --> DB[(PostgreSQL)]
        Repo --> VectorDB[(PGVector)]
        TaskQueue --> Redis[(Redis)]
        LLMProvider --> OpenAI_API[OpenAI API]
        EmbeddingProvider --> OpenAI_API
    end

    %% 主要流程连线
    Client -- HTTP Request --> Router
    %% Client <- WebSocket -> Notifier

    %% 样式
    classDef frontend fill:#d4f1f9,stroke:#05a,stroke-width:2px;
    classDef backend fill:#ddf1d4,stroke:#383,stroke-width:2px;
    classDef service fill:#e8f5e9,stroke:#2e7d32,stroke-width:1px;
    classDef storage fill:#f9e4cc,stroke:#a50,stroke-width:2px;
    classDef external fill:#e4ccf9,stroke:#539,stroke-width:2px;
    classDef queue fill:#fff9c4,stroke:#f57f17,stroke-width:2px;

    class Client frontend;
    class Router,Middleware,Handler,WorkerServer,TaskHandler backend;
    class Service,ServiceWorker,LLMProvider,EmbeddingProvider,Repo,FileStore service;
    class DB,VectorDB,Redis storage;
    class TaskQueue queue;
    class OpenAI_API external;

```

**主要数据流:**

1.  前端通过 HTTP API 上传文件 (`/upload`) 或发送聊天请求 (`/chat`)。
2.  API Server (通过 `FileService`) 将文件保存到存储，记录元数据到 PostgreSQL，并将 Embedding 任务信息（包含 `document_id`, `user_id`, `file_path`）放入 Redis 任务队列。API Server 立即返回 `doc_id` 和 `task_id`。
3.  Worker 消费任务队列，`EmbeddingTaskHandler` 处理任务：读取文件、文本分块、调用 OpenAI 生成向量。
4.  Worker (通过 `VectorRepository`) 将文档块和向量存入 PGVector (包含 `user_id` 等元数据)，并更新 PostgreSQL 中对应文档的状态为 `completed` 或 `failed`。
5.  聊天请求 (`/chat`) 由 `ChatService` 处理：保存用户消息，获取对话历史（根据 `user_id` 和 `conversation_id`），调用 OpenAI 生成回复，保存 AI 回复。
6.  (未来) RAG 流程：`ChatService` 调用 `RAGService`，后者使用用户查询文本生成向量，通过 `VectorRepository` 在 PGVector 中搜索相似文档块（根据 `user_id` 过滤），并将检索到的内容加入 LLM 上下文。
7.  (未来) WebSocket 流程：Worker 处理任务时，通过 WebSocket 推送进度更新。`ChatService` 的流式接口通过 WebSocket 返回 AI 回复块。

## 设置与运行

### 前提条件

1.  **Go:** 安装 Go 1.23 或更高版本。
2.  **Docker & Docker Desktop:** 用于运行 PostgreSQL + pgvector 和 Redis 容器。
3.  **OpenAI API Key:** 需要一个有效的 OpenAI API 密钥。
4.  **(可选) 数据库客户端:** 如 DBeaver, Navicat, psql 等，用于连接数据库和执行初始化脚本。
5.  **(可选) `psql` 命令行工具:** 如果希望通过命令行执行数据库初始化脚本。

### 步骤

1.  **克隆仓库 (如果需要):**
    ```bash
    git clone <your-repo-url>
    cd DreamHub
    ```

2.  **运行依赖容器 (PostgreSQL + pgvector, Redis):**
    打开终端，运行以下命令 (请将 `mysecretpassword` 替换为您选择的强密码):
    ```bash
    # 启动 PostgreSQL + pgvector
    docker run --name dreamhub-db -e POSTGRES_PASSWORD=mysecretpassword -e POSTGRES_DB=dreamhub_db -p 5432:5432 -d ankane/pgvector

    # 启动 Redis (用于 Asynq 任务队列)
    docker run --name dreamhub-redis -p 6379:6379 -d redis
    # 如果 Redis 需要密码: docker run --name dreamhub-redis -p 6379:6379 -d redis redis-server --requirepass your_redis_password
    ```
    **注意:** 如果容器已存在，您可能需要先使用 `docker stop <container_name> && docker rm <container_name>` 来停止并删除旧容器。

3.  **初始化数据库:**
    使用数据库客户端 (如 DBeaver, Navicat, psql) 连接到数据库:
    *   主机: `localhost`
    *   端口: `5432`
    *   数据库: `dreamhub_db`
    *   用户名: `postgres`
    *   密码: (您在步骤 2 中设置的密码)
    打开并执行项目根目录下的 `init_db.sql` 文件中的所有 SQL 语句。这个脚本会启用 `vector` 扩展并创建所有必需的表和索引（包括 `conversation_history`, `documents`, `tasks`, 和 `langchain_pg_embedding`）。
    **警告:** 该脚本会执行 `DROP TABLE IF EXISTS langchain_pg_embedding;` 来确保向量维度正确，请在已有重要数据的数据库上谨慎操作或先备份！

4.  **创建 `.env` 文件:**
    复制根目录下的 `.env.example` 文件为 `.env`，并根据您的环境填入必要的值，特别是 `OPENAI_API_KEY` 和 `DATABASE_URL` (确保密码正确)。
    ```bash
    cp .env.example .env
    # 然后编辑 .env 文件
    ```
    至少需要设置：
    ```dotenv
    OPENAI_API_KEY=sk-YOUR_OPENAI_API_KEY_HERE
    DATABASE_URL=postgres://postgres:mysecretpassword@localhost:5432/dreamhub_db?sslmode=disable
    # REDIS_ADDR=localhost:6379 # 如果 Redis 不在默认地址，请取消注释并修改
    # REDIS_PASSWORD=your_redis_password # 如果您的 Redis 服务器需要密码
    ```

5.  **安装 Go 依赖:**
    ```bash
    go mod tidy
    ```

6.  **编译应用:**
    ```bash
    # 编译 API Server
    go build -o bin/server ./cmd/server
    # 编译 Worker
    go build -o bin/worker ./cmd/worker
    ```
    *(在 Windows 上，输出文件名可能是 `server.exe` 和 `worker.exe`)*

7.  **运行应用 (需要两个终端):**
    确保 `.env` 文件已配置好。
    *   **终端 1 (运行 API Server):**
        ```bash
        # Windows
        .\bin\server.exe
        # Linux / macOS
        ./bin/server
        ```
        API 服务器将在 `.env` 文件中配置的端口（默认为 8080）上启动。
    *   **终端 2 (运行 Worker):**
        ```bash
        # Windows
        .\bin\worker.exe
        # Linux / macOS
        ./bin/worker
        ```
        Worker 将连接到 Redis 并开始处理 `embedding:generate` 队列中的任务。

## API 用法示例

使用 `curl` 或 Postman 等工具与 API 交互。所有 API 路径均以 `/api/v1` 为前缀。完整 API 文档请参考 [API_DOCS.md](Docs/API_DOCS.md)。

### 1. 上传文件 (触发 Embedding)

*   **请求:** `POST /api/v1/upload`
*   **类型:** `multipart/form-data`
*   **表单字段:**
    *   `user_id`: (string, required) 用户标识符。**(临时方案，未来应通过认证获取)**
    *   `file`: (file, required) 要上传的文件。
*   **示例 (`curl`):**
    ```bash
    curl -X POST -F "file=@mydocument.pdf" -F "user_id=user_test_1" http://localhost:8080/api/v1/upload
    ```
*   **成功响应 (HTTP 202 Accepted):**
    ```json
    {
      "message": "文件上传成功，正在后台处理中...",
      "filename": "mydocument.pdf",
      "doc_id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx", // 文档数据库 ID
      "task_id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"  // Asynq 任务 ID
    }
    ```

### 2. 查询异步任务状态

*   **请求:** `GET /api/v1/tasks/{task_id}/status`
*   **路径参数:**
    *   `task_id`: (string, required) 文件上传时返回的 Asynq 任务 ID。
*   **示例 (`curl`):**
    ```bash
    curl http://localhost:8080/api/v1/tasks/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx/status
    ```
*   **成功响应 (HTTP 200 OK):** 返回 `entity.Task` 结构的 JSON 对象 (如果实现了 TaskRepository 持久化)。
    ```json
    {
        "id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx", // Task UUID (如果持久化)
        "type": "embedding:generate",
        "payload": {"user_id":"user_test_1", ...}, // Base64 编码或 JSON
        "status": "completed", // "pending", "processing", "completed", "failed"
        "user_id": "user_test_1",
        "file_id": "yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy",
        "original_filename": "mydocument.pdf",
        "progress": 100.0,
        "result": null, // 或处理结果
        "error_message": "",
        "retry_count": 0,
        "max_retries": 3,
        "created_at": "2025-04-26T07:00:00Z",
        "started_at": "2025-04-26T07:00:05Z",
        "completed_at": "2025-04-26T07:00:30Z",
        "updated_at": "2025-04-26T07:00:30Z"
    }
    ```
    **注意:** 当前 `FileService.GetTaskStatus` 尚未完全实现按 Asynq Task ID 查询，此 API 可能返回未实现错误。

### 3. 开始新对话

*   **请求:** `POST /api/v1/chat`
*   **类型:** `application/json`
*   **请求体:**
    ```json
    {
      "user_id": "user_test_1", // (临时方案，未来应通过认证获取)
      "message": "你好！"
    }
    ```
*   **示例 (`curl`):**
    ```bash
    curl -X POST -H "Content-Type: application/json" -d '{"user_id":"user_test_1","message":"你好！"}' http://localhost:8080/api/v1/chat
    ```
*   **成功响应 (HTTP 200 OK):**
    ```json
    {
      "conversation_id": "zzzzzzzz-zzzz-zzzz-zzzz-zzzzzzzzzzzz", // 新生成的对话 ID
      "reply": "你好！有什么可以帮您的吗？"
    }
    ```

### 4. 继续对话

*   **请求:** `POST /api/v1/chat`
*   **类型:** `application/json`
*   **请求体:**
    ```json
    {
      "user_id": "user_test_1", // (临时方案)
      "conversation_id": "zzzzzzzz-zzzz-zzzz-zzzz-zzzzzzzzzzzz", // 上一步获取的 ID
      "message": "请根据我上传的文件总结一下主要内容。"
    }
    ```
*   **示例 (`curl`):**
    ```bash
    curl -X POST -H "Content-Type: application/json" -d '{"user_id":"user_test_1","conversation_id":"zzzzzzzz-zzzz-zzzz-zzzz-zzzzzzzzzzzz","message":"请根据我上传的文件总结一下主要内容。"}' http://localhost:8080/api/v1/chat
    ```
*   **成功响应 (HTTP 200 OK):** (回复内容取决于 LLM 和 RAG 结果)
    ```json
    {
      "conversation_id": "zzzzzzzz-zzzz-zzzz-zzzz-zzzzzzzzzzzz",
      "reply": "根据您上传的文档，主要内容是关于..."
    }
    ```

### 5. 获取对话消息列表

*   **请求:** `GET /api/v1/chat/{conversation_id}/messages`
*   **路径参数:**
    *   `conversation_id`: (string, required) 要获取消息的对话 ID。
*   **查询参数 (可选):**
    *   `limit`: (int, default: 50) 返回消息数量上限。
    *   `offset`: (int, default: 0) 跳过的消息数量。
*   **示例 (`curl`):**
    ```bash
    curl "http://localhost:8080/api/v1/chat/zzzzzzzz-zzzz-zzzz-zzzz-zzzzzzzzzzzz/messages?limit=20&offset=0"
    ```
*   **成功响应 (HTTP 200 OK):** 返回 `entity.Message` 数组。
    ```json
    [
      {
        "id": "...", "conversation_id": "...", "user_id": "user_test_1", "sender_role": "user", "content": "你好！", "timestamp": "...", "metadata": null
      },
      {
        "id": "...", "conversation_id": "...", "user_id": "user_test_1", "sender_role": "ai", "content": "你好！有什么可以帮您的吗？", "timestamp": "...", "metadata": null
      },
      ...
    ]
    ```

### 6. 列出用户上传的文档

*   **请求:** `GET /api/v1/documents`
*   **查询参数:**
    *   `user_id`: (string, required) 要查询的用户 ID。**(临时方案)**
    *   `limit`: (int, default: 20) 返回文档数量上限。
    *   `offset`: (int, default: 0) 跳过的文档数量。
*   **示例 (`curl`):**
    ```bash
    curl "http://localhost:8080/api/v1/documents?user_id=user_test_1&limit=10"
    ```
*   **成功响应 (HTTP 200 OK):** 返回 `entity.Document` 数组。
    ```json
    [
      {
        "id": "yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy",
        "user_id": "user_test_1",
        "original_filename": "mydocument.pdf",
        "stored_path": "./uploads/user_test_1/...",
        "file_size": 123456,
        "content_type": "application/pdf",
        "upload_time": "...",
        "processing_status": "completed",
        "processing_task_id": null, // 或关联的 Task UUID
        "error_message": ""
      },
      ...
    ]
    ```

## 已知问题与待办事项

*   **用户认证与授权:** 当前 API 依赖请求中显式传递 `user_id`，缺乏安全的认证机制 (如 JWT)。需要实现认证中间件。
*   **用户 ID 传递:** 虽然 Context 用于在内部传递 `user_id`，但从 API 入口到 Context 的设置依赖临时方案。
*   **任务状态查询:** `/tasks/{task_id}/status` API 依赖的 `TaskRepository` 按 Asynq Task ID 查询的功能尚未完全实现。
*   **LLM 流式处理:** `ChatService` 中的流式接口依赖的 `LLMProvider.GenerateContentStream` 可能因 `langchaingo` API 问题无法正常工作，需要进一步调试或寻找替代方案。
*   **RAG 与对话摘要:** 相关 Service (`RAGService`, `MemoryService`) 尚未实现。
*   **WebSocket:** 实时通信功能尚未实现。
*   **文本分割器:** `EmbeddingTaskHandler` 中目前使用的是临时的简单分割逻辑，需要替换为 `langchaingo/textsplitter`。
*   **错误处理与韧性:** 统一错误处理中间件已添加，但 Service 和 Repository 中的错误处理、重试、限流等机制尚不完善。
*   **配置:** 部分参数（如文本分割器参数、历史消息数）仍为硬编码，需要移至配置。
*   **测试:** 缺乏单元测试和集成测试。
*   **向量插入问题 (服务器环境):** 在服务器环境部署时，使用 `pgx.CopyFrom` 批量插入向量到 `langchain_pg_embedding` 表时，即使数据库模式 (`vector(1536)`) 和向量维度 (1536) 均正确，仍可能遇到 `ERROR: vector cannot have more than 16000 dimensions` 的运行时错误。本地环境（相同版本）可能不会出现此问题。根本原因推测与服务器环境下 `pgx` 对 `vector` 类型的处理（尤其是在 `CopyFrom` 的二进制协议中，可能与类型注册失败有关）存在问题。**临时解决方案:** 已将 `internal/repository/pgvector/vector_repo_impl.go` 中的 `AddChunks` 函数修改为使用逐条 `INSERT` 语句，这解决了运行时错误，但可能会带来性能影响，尤其是在处理包含大量块的文档时。未来可考虑升级 `pgvector-go` 和 `pgx` 依赖，尝试恢复使用 `CopyFrom`。

## 后续开发计划

详细的开发计划、架构决策和检查清单请参考 [DETAILED_PLAN.md](Docs/DETAILED_PLAN.md) 文件。开发过程中的具体步骤和成果记录在 [PROGRESS_LOG.md](Docs/PROGRESS_LOG.md) 文件中。