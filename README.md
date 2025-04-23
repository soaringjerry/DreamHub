# DreamHub

DreamHub 是一个 AI 驱动的工作站/工作面板后端服务，旨在通过集成个人知识库和对话记忆，提高信息处理和交互效率。

当前版本实现了以下核心功能：
*   **文件上传与处理:** 接收上传的文件，自动进行文本分块和向量化，并将 `user_id` 存入元数据。
*   **个人知识库 (RAG):** 将处理后的文档存入向量数据库 (PostgreSQL + pgvector)，支持基于文件内容的智能问答。**(注意：当前的向量搜索用户隔离过滤存在已知问题，详见 `debug.md`)**
*   **对话历史记忆:** 记录多轮对话上下文（基于 `conversation_id`），实现更连贯的 AI 交互。
*   **基础 API:** 提供文件上传和聊天交互的 API 端点，支持临时的 `user_id` 进行数据关联。

## 设置与运行

### 前提条件

1.  **Go:** 安装 Go 1.23 或更高版本。
2.  **Docker & Docker Desktop:** 用于运行 PostgreSQL + pgvector 数据库容器。
3.  **OpenAI API Key:** 需要一个有效的 OpenAI API 密钥。

### 步骤

1.  **克隆仓库 (如果需要):**
    ```bash
    git clone <your-repo-url>
    cd DreamHub
    ```

2.  **运行 PostgreSQL + pgvector 容器:**
    打开终端，运行以下命令 (请将 `mysecretpassword` 替换为您选择的强密码):
    ```bash
    docker run --name dreamhub-db -e POSTGRES_PASSWORD=mysecretpassword -e POSTGRES_DB=dreamhub_db -p 5432:5432 -d ankane/pgvector
    ```
    **注意:** 如果容器已存在，您可能需要先使用 `docker stop dreamhub-db && docker rm dreamhub-db` 来停止并删除旧容器。

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
    ```
    *(注意: 添加了 `conversation_history` 表的创建)*

4.  **创建 `.env` 文件:**
    在项目根目录创建 `.env` 文件，并填入以下内容，替换占位符：
    ```dotenv
    # .env
    OPENAI_API_KEY=sk-YOUR_OPENAI_API_KEY_HERE
    DATABASE_URL=postgres://postgres:mysecretpassword@localhost:5432/dreamhub_db
    ```
    *(将 `mysecretpassword` 替换为您设置的真实密码)*

5.  **安装 Go 依赖:**
    ```bash
    go mod tidy
    ```

6.  **运行服务器:**
    ```bash
    go run cmd/server/main.go
    ```
    服务器将在 `http://localhost:8080` 上启动。

## API 用法示例

使用 `curl` 或 Postman 等工具与 API 交互。

### 1. 上传文件 (构建知识库)

需要提供 `user_id` (表单字段) 和 `file`。将 `your_document.txt` 替换为实际文件名，`user_A` 替换为用户标识。

```bash
# PowerShell / bash / zsh
curl -X POST -F "file=@your_document.txt" -F "user_id=user_A" http://localhost:8080/api/v1/upload

# Windows cmd
# (curl 在 cmd 中上传文件可能需要不同语法，建议使用 PowerShell 或其他工具)
```
成功响应示例:
```json
{
  "message": "文件上传、分块、向量化并存储成功",
  "filename": "your_document.txt",
  "chunks": 5
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

*   **对话历史隔离:** 代码层面已为 `conversation_history` 表添加 `user_id` 支持以实现用户隔离。**注意：需要手动在数据库中为该表添加 `user_id` 字段才能生效。**

## 后续开发计划 (参考 PLAN.md)

*   **修复向量搜索过滤问题。**
*   **为对话历史添加用户隔离。**
*   实现更复杂的文档解析 (PDF, DOCX)。
*   优化 RAG 检索策略和 Prompt 工程。
*   实现结构化知识提取与存储。
*   构建前端用户界面。
*   开发 AI Agent 功能。
*   添加自动化工作流。