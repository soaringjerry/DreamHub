# DreamHub 后端 API 文档

本文档详细描述了与 DreamHub 后端服务交互的 API 请求格式。

## 1. 通用约定

*   **基础 URL**: 所有 API 端点的基础路径为 `/api/v1` (除了 `/health`)。
*   **数据格式**:
    *   对于需要发送数据的 `POST` 或 `PUT` 请求，除非另有说明，请求体应为 JSON 格式，`Content-Type` 头应设置为 `application/json`。
    *   文件上传使用 `multipart/form-data` 格式。
*   **响应格式**:
    *   成功的响应通常返回 JSON 格式的数据和 `2xx` 状态码。
    *   错误的响应返回 JSON 格式，包含一个 `error` 字段，其值为包含错误详情的对象（基于 `pkg/apperr.AppError` 结构），并返回相应的 `4xx` 或 `5xx` 状态码。
        ```json
        // 错误响应示例 (e.g., 400 Bad Request)
        {
          "error": {
            "Code": "INVALID_ARGUMENT", // 错误代码 (ErrorCode)
            "Message": "缺少 user_id 表单字段", // 用户友好的消息
            "Details": null, // 可选的详细信息列表
            "HTTPStatus": 400 // 对应的 HTTP 状态码
          }
        }
        ```
*   **用户认证**: 目前 API 依赖在请求中（表单字段或请求体）传递 `user_id`。**这是一个临时方案**，未来将通过认证中间件（如 JWT）来获取用户信息。

## 2. API 端点

### 2.1 文件上传 (`/upload`)

此端点用于上传文件到服务器进行异步处理（文本分割、向量化等）。

*   **方法**: `POST`
*   **路径**: `/api/v1/upload`
*   **请求头**:
    *   `Content-Type`: `multipart/form-data`
*   **请求体**: `multipart/form-data`
    *   **`user_id`** (string, required): 用户标识符。**(临时方案)**
    *   **`file`** (file, required): 要上传的文件。
*   **示例 (`curl`):**
    ```bash
    curl -X POST -F "file=@mydocument.pdf" -F "user_id=user_test_1" http://localhost:8080/api/v1/upload
    ```
*   **成功响应 (202 Accepted)**: 表示文件已接收并开始后台处理。
    *   `Content-Type`: `application/json`
    *   **Body**:
        ```json
        {
          "message": "文件上传成功，正在后台处理中...",
          "filename": "mydocument.pdf",
          "doc_id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx", // 文档数据库 ID (UUID)
          "task_id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"  // Asynq 任务 ID (字符串)
        }
        ```
*   **错误响应**:
    *   **400 Bad Request**: 请求格式错误、缺少字段等。
    *   **500 Internal Server Error**: 文件处理或任务入队失败。
    *   *(示例见通用约定)*

### 2.2 聊天交互 (`/chat`)

此端点用于与 AI 进行对话。

*   **方法**: `POST`
*   **路径**: `/api/v1/chat`
*   **请求头**:
    *   `Content-Type`: `application/json`
*   **请求体**: JSON 对象
    *   **`user_id`** (string, required): 用户标识符。**(临时方案)**
    *   **`message`** (string, required): 用户发送的消息。
    *   **`conversation_id`** (string, optional): 对话 ID (UUID 格式)。如果为空或未提供，则开始新对话。
    ```json
    // 开始新对话
    { "user_id": "user_test_1", "message": "你好！" }
    // 继续对话
    { "user_id": "user_test_1", "conversation_id": "zzzzzzzz-zzzz-zzzz-zzzz-zzzzzzzzzzzz", "message": "上个问题再说详细点。" }
    ```
*   **成功响应 (200 OK)**:
    *   `Content-Type`: `application/json`
    *   **Body**:
        ```json
        {
          "conversation_id": "zzzzzzzz-zzzz-zzzz-zzzz-zzzzzzzzzzzz", // 对话 ID (UUID)
          "reply": "AI 的回复内容。"
        }
        ```
*   **错误响应**:
    *   **400 Bad Request**: 请求体无效、`conversation_id` 格式错误。
    *   **500 Internal Server Error**: 获取历史记录失败、LLM 调用失败、保存消息失败等。
    *   **503 Service Unavailable**: LLM 服务不可用。
    *   *(示例见通用约定)*

### 2.3 获取对话消息 (`/chat/{conversation_id}/messages`)

获取指定对话的消息列表。

*   **方法**: `GET`
*   **路径**: `/api/v1/chat/{conversation_id}/messages`
*   **路径参数**:
    *   `conversation_id`: (string, required) 对话 ID (UUID 格式)。
*   **查询参数 (可选)**:
    *   `limit`: (int, default: 50) 返回消息数量上限。
    *   `offset`: (int, default: 0) 跳过的消息数量。
*   **示例 (`curl`):**
    ```bash
    curl "http://localhost:8080/api/v1/chat/zzzzzzzz-zzzz-zzzz-zzzz-zzzzzzzzzzzz/messages?limit=20"
    ```
*   **成功响应 (200 OK)**:
    *   `Content-Type`: `application/json`
    *   **Body**: `entity.Message` 结构体数组。
        ```json
        [
          { "id": "...", "conversation_id": "...", "user_id": "...", "sender_role": "user", "content": "...", "timestamp": "...", "metadata": null },
          { "id": "...", "conversation_id": "...", "user_id": "...", "sender_role": "ai", "content": "...", "timestamp": "...", "metadata": null }
        ]
        ```
*   **错误响应**:
    *   **400 Bad Request**: `conversation_id` 格式错误。
    *   **403 Forbidden / 404 Not Found**: (需要认证后) 如果用户无权访问该对话。
    *   **500 Internal Server Error**: 查询数据库失败。
    *   *(示例见通用约定)*

### 2.4 查询任务状态 (`/tasks/{task_id}/status`)

查询异步任务（如文件处理）的状态。

*   **方法**: `GET`
*   **路径**: `/api/v1/tasks/{task_id}/status`
*   **路径参数**:
    *   `task_id`: (string, required) 文件上传时返回的 Asynq 任务 ID。
*   **示例 (`curl`):**
    ```bash
    curl http://localhost:8080/api/v1/tasks/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx/status
    ```
*   **成功响应 (200 OK)**: 返回 `entity.Task` 结构的 JSON 对象 (如果实现了 TaskRepository 持久化)。
    ```json
    {
        "id": "...", // Task UUID (如果持久化)
        "type": "embedding:generate",
        "payload": {...},
        "status": "completed", // "pending", "processing", "completed", "failed"
        "user_id": "...",
        "file_id": "...",
        "original_filename": "...",
        "progress": 100.0,
        "result": null,
        "error_message": "",
        "retry_count": 0,
        "max_retries": 3,
        "created_at": "...",
        "started_at": "...",
        "completed_at": "...",
        "updated_at": "..."
    }
    ```
*   **错误响应**:
    *   **400 Bad Request**: `task_id` 格式无效 (如果期望 UUID 但传入非 UUID)。
    *   **404 Not Found**: 任务不存在。
    *   **501 Not Implemented**: 当前实现可能不支持按 Asynq ID 查询。
    *   **500 Internal Server Error**: 查询数据库失败。
    *   *(示例见通用约定)*

### 2.5 文档管理 (`/documents`)

#### 2.5.1 列出用户文档

*   **方法**: `GET`
*   **路径**: `/api/v1/documents`
*   **查询参数**:
    *   `user_id`: (string, required) 用户 ID。**(临时方案)**
    *   `limit`: (int, default: 20) 返回数量上限。
    *   `offset`: (int, default: 0) 跳过数量。
*   **示例 (`curl`):**
    ```bash
    curl "http://localhost:8080/api/v1/documents?user_id=user_test_1&limit=10"
    ```
*   **成功响应 (200 OK)**: 返回 `entity.Document` 结构体数组。
    ```json
    [
      { "id": "...", "user_id": "...", "original_filename": "...", ... },
      ...
    ]
    ```
*   **错误响应**:
    *   **400 Bad Request**: 缺少 `user_id`。
    *   **403 Forbidden**: (需要认证后) 无权访问该用户文档。
    *   **500 Internal Server Error**: 查询数据库失败。
    *   *(示例见通用约定)*

#### 2.5.2 获取文档详情

*   **方法**: `GET`
*   **路径**: `/api/v1/documents/{doc_id}`
*   **路径参数**:
    *   `doc_id`: (string, required) 文档 ID (UUID 格式)。
*   **示例 (`curl`):**
    ```bash
    curl http://localhost:8080/api/v1/documents/yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy
    ```
*   **成功响应 (200 OK)**: 返回单个 `entity.Document` 结构体。
    ```json
    { "id": "...", "user_id": "...", "original_filename": "...", ... }
    ```
*   **错误响应**:
    *   **400 Bad Request**: `doc_id` 格式错误。
    *   **404 Not Found**: 文档不存在或用户无权访问。
    *   **500 Internal Server Error**: 查询数据库失败。
    *   *(示例见通用约定)*

#### 2.5.3 删除文档

*   **方法**: `DELETE`
*   **路径**: `/api/v1/documents/{doc_id}`
*   **路径参数**:
    *   `doc_id`: (string, required) 文档 ID (UUID 格式)。
*   **示例 (`curl`):**
    ```bash
    curl -X DELETE http://localhost:8080/api/v1/documents/yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy
    ```
*   **成功响应 (200 OK)**:
    ```json
    {
      "message": "文档已成功删除"
    }
    ```
*   **错误响应**:
    *   **400 Bad Request**: `doc_id` 格式错误。
    *   **404 Not Found**: 文档不存在或用户无权删除。
    *   **500 Internal Server Error**: 删除过程中发生错误（删除文件、向量或元数据失败）。
    *   *(示例见通用约定)*

### 2.6 健康检查 (`/health`)

检查 API 服务器是否正在运行。

*   **方法**: `GET`
*   **路径**: `/health` (注意：不在 `/api/v1` 下)
*   **请求体**: 无
*   **成功响应 (200 OK)**:
    *   `Content-Type`: `application/json`
    *   **Body**:
        ```json
        {
          "status": "ok"
        }
        ```
*   **错误响应**: 通常不会有特定错误，失败表示服务器无法访问。