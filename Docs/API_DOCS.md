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
    *   **`model_name`** (string, optional): 指定要使用的 LLM 模型名称 (例如 "gpt-4", "gpt-3.5-turbo")。如果为空或未提供，则使用服务器配置的默认模型。
    ```json
    // 开始新对话 (使用默认模型)
    { "user_id": "user_test_1", "message": "你好！" }
    // 继续对话 (使用默认模型)
    { "user_id": "user_test_1", "conversation_id": "zzzzzzzz-zzzz-zzzz-zzzz-zzzzzzzzzzzz", "message": "上个问题再说详细点。" }
    // 开始新对话 (指定模型)
    { "user_id": "user_test_1", "message": "用 GPT-4 回答我", "model_name": "gpt-4" }
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

### 2.6 用户配置 (`/users/me/config`)

管理当前登录用户的配置信息。**注意:** 这些端点依赖于有效的用户认证（例如 JWT Token），而不是临时传递 `user_id`。

#### 2.6.1 获取用户配置

*   **方法**: `GET`
*   **路径**: `/api/v1/users/me/config`
*   **认证**: 需要有效的用户认证 (JWT Token)。
*   **成功响应 (200 OK)**: 返回用户的配置对象 (`entity.UserConfig`)。
    *   `Content-Type`: `application/json`
    *   **Body**:
        ```json
        {
          "user_id": "user-uuid-string",
          "openai_api_key": "sk-...", // 注意：实际返回时可能部分屏蔽或不返回敏感信息
          "default_model": "gpt-4",
          "created_at": "2025-05-01T10:00:00Z",
          "updated_at": "2025-05-01T10:00:00Z"
        }
        ```
*   **错误响应**:
    *   **401 Unauthorized**: 未认证或认证无效。
    *   **404 Not Found**: 用户配置不存在（可能是新用户首次访问）。
    *   **500 Internal Server Error**: 查询数据库失败或解密密钥失败。
    *   *(示例见通用约定)*

#### 2.6.2 更新用户配置

*   **方法**: `PUT`
*   **路径**: `/api/v1/users/me/config`
*   **认证**: 需要有效的用户认证 (JWT Token)。
*   **请求头**:
    *   `Content-Type`: `application/json`
*   **请求体**: JSON 对象 (`dto.UpdateUserConfigDTO`)，包含要更新的字段。允许部分更新。
    ```json
    // 示例：只更新 API Key
    {
      "openai_api_key": "new-sk-..."
    }
    // 示例：只更新默认模型
    {
      "default_model": "gpt-4-turbo"
    }
    // 示例：同时更新
    {
      "openai_api_key": "another-sk-...",
      "default_model": "gpt-3.5-turbo"
    }
    ```
*   **成功响应 (200 OK)**: 返回更新后的用户配置对象 (`entity.UserConfig`)。
    *   `Content-Type`: `application/json`
    *   **Body**:
        ```json
        {
          "user_id": "user-uuid-string",
          "openai_api_key": "new-sk-...", // 注意：实际返回时可能部分屏蔽或不返回敏感信息
          "default_model": "gpt-4-turbo", // 已更新
          "created_at": "2025-05-01T10:00:00Z",
          "updated_at": "2025-05-01T11:20:30Z" // 更新时间已改变
        }
        ```
*   **错误响应**:
    *   **400 Bad Request**: 请求体无效（例如，字段类型错误）。
    *   **401 Unauthorized**: 未认证或认证无效。
    *   **500 Internal Server Error**: 更新数据库失败或加密密钥失败。
    *   *(示例见通用约定)*
### 2.6 健康检查 (`/health`)
### 2.7 获取对话列表 (`/conversations`)

获取当前登录用户的对话列表（通常只包含对话 ID 和标题/摘要）。

*   **方法**: `GET`
*   **路径**: `/api/v1/conversations`
*   **认证**: 需要有效的用户认证 (JWT Token)。
*   **查询参数 (可选)**:
    *   `limit`: (int, default: 50) 返回对话数量上限。
    *   `offset`: (int, default: 0) 跳过的对话数量。
*   **成功响应 (200 OK)**: 返回对话信息对象的数组。具体结构取决于后端实现（例如，可能包含 `id`, `title`, `last_updated_at` 等）。
    *   `Content-Type`: `application/json`
    *   **Body**:
        ```json
        [
          {
            "id": "conv-uuid-1",
            "title": "关于项目 A 的讨论", // 或首条消息摘要
            "last_message_timestamp": "2025-05-01T11:00:00Z"
          },
          {
            "id": "conv-uuid-2",
            "title": "API 设计思路",
            "last_message_timestamp": "2025-04-30T15:30:00Z"
          }
          // ...
        ]
        ```
*   **错误响应**:
    *   **401 Unauthorized**: 未认证或认证无效。
    *   **500 Internal Server Error**: 查询数据库失败。
    *   *(示例见通用约定)*

### 2.8 结构化记忆 (`/memory/structured`)

管理用户的结构化记忆条目（键值对）。

*   **认证**: 所有端点都需要有效的用户认证 (JWT Token)。

#### 2.8.1 创建或更新记忆条目

*   **方法**: `POST`
*   **路径**: `/api/v1/memory/structured`
*   **请求头**:
    *   `Content-Type`: `application/json`
*   **请求体**: JSON 对象 (`entity.StructuredMemory`)
    ```json
    {
      "key": "user_preference_theme",
      "value": { "mode": "dark", "accent_color": "#8844ee" } // value 可以是任何 JSON 结构
    }
    ```
*   **成功响应 (201 Created 或 200 OK)**: 返回创建或更新后的记忆条目。
    *   `Content-Type`: `application/json`
    *   **Body**:
        ```json
        {
          "id": "memory-uuid-1",
          "user_id": "user-uuid-string",
          "key": "user_preference_theme",
          "value": { "mode": "dark", "accent_color": "#8844ee" },
          "created_at": "2025-05-01T11:30:00Z",
          "updated_at": "2025-05-01T11:30:00Z"
        }
        ```
*   **错误响应**:
    *   **400 Bad Request**: 请求体无效（缺少 `key` 或 `value`）。
    *   **401 Unauthorized**: 未认证或认证无效。
    *   **500 Internal Server Error**: 数据库操作失败。
    *   *(示例见通用约定)*

#### 2.8.2 获取所有记忆条目

*   **方法**: `GET`
*   **路径**: `/api/v1/memory/structured`
*   **成功响应 (200 OK)**: 返回该用户的所有记忆条目数组。
    *   `Content-Type`: `application/json`
    *   **Body**: `entity.StructuredMemory` 数组
        ```json
        [
          { "id": "...", "user_id": "...", "key": "key1", "value": {...}, ... },
          { "id": "...", "user_id": "...", "key": "key2", "value": "...", ... }
        ]
        ```
*   **错误响应**:
    *   **401 Unauthorized**: 未认证或认证无效。
    *   **500 Internal Server Error**: 查询数据库失败。
    *   *(示例见通用约定)*

#### 2.8.3 获取特定记忆条目

*   **方法**: `GET`
*   **路径**: `/api/v1/memory/structured/{key}`
*   **路径参数**:
    *   `key`: (string, required) 要获取的记忆条目的键。
*   **成功响应 (200 OK)**: 返回匹配的记忆条目。
    *   `Content-Type`: `application/json`
    *   **Body**: `entity.StructuredMemory`
        ```json
        { "id": "...", "user_id": "...", "key": "user_preference_theme", "value": {...}, ... }
        ```
*   **错误响应**:
    *   **401 Unauthorized**: 未认证或认证无效。
    *   **404 Not Found**: 指定 `key` 的记忆条目不存在。
    *   **500 Internal Server Error**: 查询数据库失败。
    *   *(示例见通用约定)*

#### 2.8.4 更新特定记忆条目 (按 Key)

*   **方法**: `PUT`
*   **路径**: `/api/v1/memory/structured/{key}`
*   **路径参数**:
    *   `key`: (string, required) 要更新的记忆条目的键。
*   **请求头**:
    *   `Content-Type`: `application/json`
*   **请求体**: JSON 对象，只包含 `value` 字段。
    ```json
    {
      "value": { "mode": "light", "accent_color": "#00aabb" } // 新的 value
    }
    ```
*   **成功响应 (200 OK)**: 返回更新后的记忆条目。
    *   `Content-Type`: `application/json`
    *   **Body**: `entity.StructuredMemory`
        ```json
        { "id": "...", "user_id": "...", "key": "user_preference_theme", "value": { "mode": "light", ... }, "updated_at": "...", ... }
        ```
*   **错误响应**:
    *   **400 Bad Request**: 请求体无效（缺少 `value`）。
    *   **401 Unauthorized**: 未认证或认证无效。
    *   **404 Not Found**: 指定 `key` 的记忆条目不存在。
    *   **500 Internal Server Error**: 更新数据库失败。
    *   *(示例见通用约定)*

#### 2.8.5 删除特定记忆条目

*   **方法**: `DELETE`
*   **路径**: `/api/v1/memory/structured/{key}`
*   **路径参数**:
    *   `key`: (string, required) 要删除的记忆条目的键。
*   **成功响应 (204 No Content)**: 表示成功删除，响应体为空。
*   **错误响应**:
    *   **401 Unauthorized**: 未认证或认证无效。
    *   **404 Not Found**: 指定 `key` 的记忆条目不存在。
    *   **500 Internal Server Error**: 删除数据库记录失败。
    *   *(示例见通用约定)*
### 2.9 健康检查 (`/health`)

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