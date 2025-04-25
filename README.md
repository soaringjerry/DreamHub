# DreamHub

DreamHub 是一个 AI 驱动的工作站/工作面板后端服务，旨在通过集成个人知识库和对话记忆，提高信息处理和交互效率。

当前版本（经过重构）包含两个主要组件：**API Server** 和 **Worker**。

核心功能：
*   **文件上传与异步处理:** API Server 接收上传的文件并保存，然后将处理任务（文本分块、向量化）放入队列。Worker 进程负责后台处理这些任务。API Server 立即返回任务 ID。
*   **个人知识库 (RAG):** Worker 将处理后的文档向量存入向量数据库 (PostgreSQL + pgvector)。Repository 层强制进行用户隔离过滤（基于 `user_id` 元数据），支持基于用户文件内容的智能问答。**(注意：目前 `user_id` 的传递机制尚待完善)**
*   **对话历史记忆:** 记录多轮对话上下文（基于 `conversation_id` 和 `user_id`），实现更连贯的 AI 交互。**(注意：历史记录的读写目前仍在 Service 层，待重构至 Repository)**
*   **基础 API:** 提供文件上传 (`/upload`) 和聊天交互 (`/chat`) 的 API 端点。详细 API 文档请参考 [API_DOCS.md](API_DOCS.md)。

## 系统架构

```mermaid
graph TD
    %% 前端
    Client[前端客户端] -->|HTTP请求| API[API服务器]
    Client <-->|WebSocket| WS[WebSocket服务器]
    
    %% 后端服务
    API -->|任务入队| Redis[(Redis队列)]
    API -->|查询/存储| DB[(PostgreSQL)]
    API -->|存储文件| Files[(文件存储)]
    
    %% WebSocket服务
    WS -->|查询数据| DB
    WS -->|订阅通知| Redis
    
    %% Worker处理
    Worker[异步Worker] -->|消费任务| Redis
    Worker -->|生成向量| OpenAI[OpenAI API]
    Worker -->|存储向量| Vector[(PGVector)]
    Worker -->|读取文件| Files
    Worker -->|状态更新| WS
    
    %% 数据流说明
    DB <--> Vector
    
    %% 样式
    classDef frontend fill:#d4f1f9,stroke:#05a,stroke-width:2px;
    classDef backend fill:#ddf1d4,stroke:#383,stroke-width:2px;
    classDef storage fill:#f9e4cc,stroke:#a50,stroke-width:2px;
    classDef external fill:#e4ccf9,stroke:#539,stroke-width:2px;
    classDef websocket fill:#f9d4e1,stroke:#b27,stroke-width:2px;
    
    class Client frontend;
    class API,Worker backend;
    class DB,Vector,Redis,Files storage;
    class OpenAI external;
    class WS websocket;
```

主要数据流:
1. 前端通过HTTP API上传文件或发送聊天请求
2. API服务器将文件处理任务入队到Redis
3. Worker消费任务队列，处理文件(分块、调用OpenAI生成向量)
4. Worker将向量存入PGVector，并通过WebSocket发送进度更新
5. 前端通过WebSocket接收实时进度和流式AI回复
6. 聊天请求通过RAG检索相关文档向量，增强AI回复质量

## 设置与运行

### 前提条件

1.  **Go:** 安装 Go 1.23 或更高版本。
2.  **Docker & Docker Desktop:** 用于运行 PostgreSQL + pgvector 和 Redis 容器。
3.  **OpenAI API Key:** 需要一个有效的 OpenAI API 密钥。
4.  **(可选) 数据库客户端:** 如 Navicat, DBeaver, psql 等，用于初始化数据库。

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
    ```
    **注意:** 如果容器已存在，您可能需要先使用 `docker stop <container_name> && docker rm <container_name>` 来停止并删除旧容器。

3.  **启用 pgvector 扩展并创建表:**
    使用数据库客户端 (如 Navicat, DBeaver, psql) 连接到数据库:
    *   主机: `localhost`
    *   端口: `5432`
    *   数据库: `dreamhub_db`
    *   用户名: `postgres`
    *   密码: (您在上面设置的密码)
    执行以下 SQL:
    ```sql
    CREATE EXTENSION IF NOT EXISTS vector;

    CREATE TABLE IF NOT EXISTS conversation_history (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        conversation_id UUID NOT NULL,
        user_id VARCHAR(255), -- Added user_id for isolation (adjust type/length if needed)
        sender_role VARCHAR(10) NOT NULL CHECK (sender_role IN ('user', 'ai')),
        message_content TEXT NOT NULL,
        timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        metadata JSONB
    );
    -- Updated index to include user_id for better query performance
    CREATE INDEX IF NOT EXISTS idx_conversation_history_user_conv_id_ts
    ON conversation_history (user_id, conversation_id, timestamp);

    -- (重要性能优化) 为 embedding 元数据创建 GIN 索引
    -- 这对于高效地按 user_id 或其他元数据过滤向量至关重要
    -- 注意: langchaingo 可能默认创建 cmetadata 列为 json 类型。
    --       在某些 PostgreSQL 环境下，json 类型可能缺少必要的 GIN 操作符类 (json_ops)，导致索引创建失败。
    -- 推荐解决方案:
    -- 步骤 1: 将 cmetadata 列转换为 jsonb 类型 (jsonb 通常性能更好且索引支持更完善)
    ALTER TABLE langchain_pg_embedding ALTER COLUMN cmetadata TYPE jsonb USING cmetadata::jsonb;
    -- 步骤 2: 为转换后的 jsonb 列创建 GIN 索引 (通常不需要指定操作符类)
    CREATE INDEX IF NOT EXISTS idx_gin_embedding_metadata
    ON langchain_pg_embedding USING GIN (cmetadata);
    ```
    *(注意: 添加了 `conversation_history` 表的创建)*

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
        Worker 将连接到 Redis 并开始处理任务。

## API 用法示例

使用 `curl` 或 Postman 等工具与 API 交互。完整 API 文档请参考 [API_DOCS.md](API_DOCS.md)。

### 1. 上传文件 (构建知识库)

需要提供 `user_id` (表单字段) 和 `file`。将 `your_document.txt` 替换为实际文件名，`user_A` 替换为用户标识。

```bash
# PowerShell / bash / zsh
curl -X POST -F "file=@your_document.txt" -F "user_id=user_A" http://localhost:8080/api/v1/upload

# Windows cmd
# (curl 在 cmd 中上传文件可能需要不同语法，建议使用 PowerShell 或其他工具)
```
成功响应示例 (HTTP 状态码 202 Accepted):
```json
{
  "message": "文件上传成功，正在后台处理中...",
  "filename": "your_document.txt",
  "task_id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}
```

### 2. 开始新对话

需要提供 `user_id`。

```bash
# Windows cmd (注意 JSON 转义)
curl -X POST -H "Content-Type: application/json" -d "{\"user_id\":\"user_A\",\"message\":\"你好！\"}" http://localhost:8080/api/v1/chat

# PowerShell / bash / zsh
# curl -X POST -H "Content-Type: application/json" -d '{"user_id":"user_A","message":"你好！"}' http://localhost:8080/api/v1/chat
```
成功响应示例 (记下 `conversation_id`):
```json
{
  "conversation_id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  "reply": "你好！有什么可以帮您的吗？"
}
```

### 3. 继续对话 (使用 `conversation_id` 和 `user_id`)

将 `YOUR_CONVERSATION_ID` 替换为上一步获取的 ID，`user_A` 替换为对应的用户 ID。

```bash
# Windows cmd
curl -X POST -H "Content-Type: application/json" -d "{\"conversation_id\":\"YOUR_CONVERSATION_ID\",\"user_id\":\"user_A\",\"message\":\"请根据我上传的文件总结一下主要内容。\"}" http://localhost:8080/api/v1/chat

# PowerShell / bash / zsh
# curl -X POST -H "Content-Type: application/json" -d '{"conversation_id":"YOUR_CONVERSATION_ID","user_id":"user_A","message":"请根据我上传的文件总结一下主要内容。"}' http://localhost:8080/api/v1/chat
```
成功响应示例 (回复会基于该用户的 RAG 上下文和该对话的历史):
```json
{
  "conversation_id": "YOUR_CONVERSATION_ID",
  "reply": "根据您上传的文件，主要内容是关于..."
}
```

## 已知问题

*   **用户 ID 传递:** 当前向量库过滤和 Worker 处理依赖的 `user_id` 传递机制尚不完善（例如，通过 Context 传递）。
*   **对话历史 Repository:** 对话历史的读写逻辑尚未抽象到独立的 Repository 层。
*   **任务状态查询:** 尚未实现查询异步任务处理状态的 API。

## 后续开发计划

详细的开发计划、架构决策和检查清单请参考 [PLAN.md](PLAN.md) 文件。