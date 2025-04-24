# 后端 API 请求格式文档

本文档详细描述了与 DreamHub 后端服务交互的 API 请求格式。

## 1. 通用约定

*   **基础 URL**: 所有 API 端点的基础路径为 `/api/v1`。
*   **数据格式**:
    *   对于需要发送数据的 POST 请求，除非另有说明，请求体应为 JSON 格式，`Content-Type` 头应设置为 `application/json`。
    *   文件上传使用 `multipart/form-data` 格式。
*   **响应格式**: 成功的响应通常返回 JSON 格式的数据。错误响应也可能返回 JSON，包含一个 `message` 字段描述错误。

## 2. API 端点

### 2.1 文件上传 (`/upload`)

此端点用于上传文件到服务器进行处理（例如，文档解析）。

*   **方法**: `POST`
*   **路径**: `/api/v1/upload`
*   **请求头**:
    *   `Content-Type`: `multipart/form-data` (由浏览器或客户端自动设置边界)
*   **请求体**: `multipart/form-data`
    *   包含以下两个字段：
        *   **`file`** (file, required): 要上传的文件内容。
        *   **`user_id`** (string, required): 标识上传用户的唯一 ID。
    ```
    ------WebKitFormBoundary...
    Content-Disposition: form-data; name="user_id"

    user_12345
    ------WebKitFormBoundary...
    Content-Disposition: form-data; name="file"; filename="example.txt"
    Content-Type: text/plain

    (文件内容)
    ------WebKitFormBoundary...--
    ```
*   **成功响应 (202 Accepted)**:
    *   `Content-Type`: `application/json`
    *   **Body**:
        ```json
        {
          "message": "File upload accepted, processing in background.", // 描述操作结果的消息
          "filename": "example.txt", // 上传的文件名
          "task_id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" // 后台处理任务的 ID
        }
        ```
*   **错误响应**:
    *   **4xx/5xx**: 可能返回 JSON 格式的错误信息。
        ```json
        {
          "message": "Upload failed: Missing file part" // 具体的错误描述
        }
        ```
    *   也可能返回其他 HTTP 错误状态码，具体取决于错误类型（例如，服务器内部错误）。

### 2.2 聊天交互 (`/chat`)

此端点用于向 AI 发送消息并获取回复。

*   **方法**: `POST`
*   **路径**: `/api/v1/chat`
*   **请求头**:
    *   `Content-Type`: `application/json`
*   **请求体**: JSON 对象
    *   **`user_id`** (string, required): 标识当前用户的唯一 ID。用于隔离用户数据和对话历史。
    *   **`message`** (string, required): 用户发送的聊天消息内容。
    *   **`conversation_id`** (string, optional): 当前对话的唯一标识符 (UUID 格式)。如果提供，服务器将尝试在现有对话上下文中继续（并验证用户 ID 是否匹配）；如果不提供或服务器找不到对应 ID，则会开始新的对话并返回新的对话 ID。
    ```json
    // 开始新对话 (必须提供 user_id)
    {
      "user_id": "user_abcde",
      "message": "你好，请介绍一下你自己。"
    }

    // 继续现有对话 (必须提供 user_id 和 conversation_id)
    {
      "user_id": "user_abcde",
      "message": "关于上一个问题，你能详细说明吗？",
      "conversation_id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" // UUID 格式
    }
    ```
*   **成功响应 (200 OK)**:
    *   `Content-Type`: `application/json`
    *   **Body**:
        ```json
        {
          "conversation_id": "conv_12345abcde", // 本次交互所属的对话 ID (可能是新的或沿用旧的)
          "reply": "我是 DreamHub AI 助手，我可以帮助你处理文档和回答问题。" // AI 的回复内容
        }
        ```
*   **错误响应**:
    *   **4xx/5xx**: 可能返回 JSON 格式的错误信息。
        ```json
        {
          "message": "Chat request failed: Invalid input message" // 具体的错误描述
        }
        ```
    *   也可能返回其他 HTTP 错误状态码。
### 2.3 健康检查 (`/ping`)

此端点用于检查服务器是否正在运行。注意：此端点不在 `/api/v1` 路径下。

*   **方法**: `GET`
*   **路径**: `/ping`
*   **请求体**: 无
*   **成功响应 (200 OK)**:
    *   `Content-Type`: `application/json`
    *   **Body**:
        ```json
        {
          "message": "pong"
        }
        ```
*   **错误响应**: 通常不会有特定错误，失败表示服务器无法访问或内部错误。