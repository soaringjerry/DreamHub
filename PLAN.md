# DreamHub 项目计划

## 1. 项目名称

DreamHub

## 2. 项目目标

构建一个全能的 AI 驱动工作站/工作面板，旨在通过集成文档管理、自动化工作流和 AI Agent 能力，提高个人和团队的工作效率，并具备接近人类的、能够自动判断存储方式的混合记忆系统。

## 3. 核心功能领域

*   **智能文档中心 (Intelligent Document Hub):** 统一管理、智能搜索、AI 驱动的洞察（摘要、问答、分类）。
*   **自动化工作流引擎 (Automation Workflow Engine):** 规则驱动自动化、集成外部服务（邮件、日历等）、定时任务。
*   **AI Agent 协作平台 (AI Agent Collaboration Platform):** 复杂任务委派、工具调用、人机协作、Agent 定制。
*   **统一工作面板 (Unified Workspace Dashboard):** 个性化视图、集中信息展示、通知中心。
*   **高级混合记忆系统:** 结合对话历史、向量知识库、结构化知识存储，并探索智能路由。

## 4. AI 记忆系统策略 (逐步实现)

目标是构建一个接近人类记忆方式的混合记忆系统，能够智能地存储和检索信息，并逐步实现自动判断存储类型的能力。

*   **短期记忆 (Working Memory):**
    *   **实现:** 通过管理传递给 LLM 的 Prompt 上下文窗口实现。
    *   **工具:** LangChainGo 框架的上下文管理。
*   **长期记忆 - 对话历史 (Episodic Memory):**
    *   **实现:** 将用户和 AI 的每轮交互存储在 PostgreSQL 的 `conversation_history` 表中。
    *   **作用:** 支持连贯的多轮对话。
    *   **状态:** **第一阶段实现基础版本。**
    *   **数据库表结构 (PostgreSQL):**
        ```sql
        CREATE TABLE IF NOT EXISTS conversation_history (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- 使用 UUID 作为主键
            conversation_id UUID NOT NULL,                 -- 标识对话会话
            sender_role VARCHAR(10) NOT NULL CHECK (sender_role IN ('user', 'ai')), -- 限制角色
            message_content TEXT NOT NULL,                 -- 消息内容
            timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),  -- 带时区的时间戳
            metadata JSONB                                 -- 可选的元数据
        );

        -- (可选) 为 conversation_id 和 timestamp 创建索引以加速查询
        CREATE INDEX IF NOT EXISTS idx_conversation_history_conv_id_ts
        ON conversation_history (conversation_id, timestamp);
        ```
*   **长期记忆 - 知识库 (Semantic Memory - Unstructured / Vector):**
    *   **实现:** 通过 RAG 实现。处理文档，分块，向量化后存入 pgvector。
    *   **作用:** 让 AI 能基于非结构化的文档内容进行语义理解和问答。
    *   **状态:** **第一阶段实现基础版本。**
*   **长期记忆 - 结构化知识 (Semantic Memory - Structured):**
    *   **实现:** (未来阶段) 让 LLM 执行信息提取任务 (Named Entity Recognition, Relation Extraction)，将识别出的关键实体、关系、事实存入 PostgreSQL 的标准 SQL 表中（例如项目表、联系人表、笔记表等）。
    *   **作用:** 支持更精确的查询、推理、知识关联和自动化任务。
    *   **状态:** **未来阶段规划。**
*   **区分硬性偏好与软性记忆点 (Handling Constraints vs. Preferences):**
    *   **挑战:** 需要区分必须严格遵守的规则/约束（硬性偏好，如“我不吃海鲜”、“项目截止日期是周五”）和一般性的知识点/用户风格/事实（软性记忆点，可通过向量或结构化知识检索）。将硬性规则放入向量库进行检索是不可靠的。
    *   **策略:**
        *   **硬性偏好/规则:** 应识别并存储在更可靠、易于精确检索的位置（例如：专门的用户配置表、规则库）。在构建 Prompt 时，这些规则需要被**优先、显式地**插入，确保模型严格遵守。
        *   **软性记忆点:** 可以作为结构化知识（如摘要、关键词、实体关系）存储在向量元数据或独立的结构化表中，并在 RAG 检索或 Prompt 构建时作为上下文信息提供给模型。
    *   **状态:** **未来阶段规划。**
*   **记忆路由/决策 (Memory Routing / Automatic Judgment):**
    *   **实现:** (未来/高级阶段) 设计机制（规则、启发式或 LLM 驱动的 Agent）来决定：
        *   **存储时:** 新信息应该存入哪个或哪些记忆系统。
        *   **检索时:** 根据用户查询的意图和内容，应该优先或组合查询哪些记忆系统。
    *   **作用:** 实现更智能、更自动化的记忆管理，逼近人类记忆的灵活性。
    *   **状态:** **未来/高级阶段规划。**

## 5. 第一阶段重点 (MVP)

快速搭建可用基础，为后续迭代铺路：

1.  **核心框架搭建:**
    *   后端: Go (Gin)。
    *   前端: (待定，例如 React/Vue)。
    *   数据库: PostgreSQL (启用 `pgvector` 扩展)。
    *   文件存储: 本地 `uploads` 目录。
    *   用户认证授权 (基础)。
    *   通过 `.env` 文件管理配置 (使用 `godotenv`)。
2.  **基础文档处理与向量存储 (知识库记忆):**
    *   实现 `/api/v1/upload` API。
    *   文件保存到本地。
    *   读取文件内容 (TXT)。
    *   使用 LangChainGo `RecursiveCharacterTextSplitter` 进行文本分块。
    *   使用 OpenAI Embedding 模型 (`text-embedding-3-large`) 进行向量化。
    *   将文本块、向量、元数据存入 pgvector (使用 `AddDocuments`)。
3.  **基础 RAG 问答:**
    *   实现 `/api/v1/chat` API。
    *   使用 LangChainGo `pgvector.Store` 的 `SimilaritySearch` 检索相关文档块。
    *   构建包含检索上下文的 Prompt。
    *   调用 OpenAI LLM (`gpt-4o`) 生成回复。
4.  **基础对话历史记忆:**
    *   创建 `conversation_history` 表 (如上定义)。
    *   修改 `/api/v1/chat` API：
        *   接收/生成 `conversation_id`。
        *   查询最近的对话历史。
        *   将历史记录加入 Prompt。
        *   将当前交互存入历史表。
5.  **统一工作面板 (极简):**
    *   提供基础的前端界面用于文件上传和聊天交互。

## 6. 技术栈选型

*   **后端:** Go (Gin)
*   **前端:** React (使用 Vite, TypeScript 推荐)
*   **UI 库:** Tailwind CSS
*   **状态管理:** Zustand
*   **数据库:** PostgreSQL + `pgvector` 扩展
*   **文档解析:** (初期) Go 标准库 (TXT), (后续) Go 开源库 (PDF, DOCX 等)
*   **AI 核心:** LangChainGo, OpenAI API (LLM: gpt-4o, Embedding: text-embedding-3-large)
*   **文件存储:** 本地文件系统 (`uploads` 目录)
*   **配置管理:** `.env` 文件 (使用 `godotenv` 库)

## 7. 开发理念

*   **快速迭代:** 优先交付核心价值，小步快跑。
*   **拥抱开源:** 最大化利用成熟的开源项目。
*   **混合记忆策略:** 逐步构建结合对话历史、向量、结构化知识的记忆系统。
*   **模块化设计:** 组件解耦。
*   **用户中心:** 关注实际场景。

## 8. 阶段一简化架构图

```mermaid
graph TD
    User --> FrontendUI(Web UI - React, Tailwind, Zustand)
    FrontendUI <-- REST API --> BackendAPI(Go + Gin + LangChainGo)

    BackendAPI --> Database[(PostgreSQL + pgvector + History Table)]
    BackendAPI --> FileStorage[(Local File Storage)]
    BackendAPI --> OpenAI_API[OpenAI API (LLM & Embeddings)]

    subgraph Core Infrastructure (Phase 1)
        BackendAPI
        Database
        FileStorage
        OpenAI_API
    end

    subgraph User Interface (Phase 1)
        FrontendUI
    end

## 前端开发计划 (React)

**目标:** 构建一个基于 React 的前端界面，用于与 DreamHub 后端服务交互，实现文件上传、AI 聊天和对话历史展示功能，采用简洁科技风格，并为未来扩展到 React Native 奠定基础。

**1. 项目初始化与设置**

*   在项目根目录下创建 `frontend` 文件夹。
*   使用 Vite 初始化 React 项目 (推荐使用 TypeScript):
    ```bash
    # 使用 TypeScript (推荐)
    npm create vite@latest frontend -- --template react-ts
    # 或者使用 JavaScript
    # npm create vite@latest frontend -- --template react
    cd frontend
    npm install
    ```
*   安装 `axios`, `zustand`, 和 `tailwindcss`:
    ```bash
    npm install axios zustand
    npm install -D tailwindcss postcss autoprefixer
    npx tailwindcss init -p
    ```
*   配置 Tailwind CSS (参照官方 Vite 指南)。

**2. 核心组件设计 (React)**

```mermaid
graph TD
    A[App.tsx (主应用)] --> B(FileUpload.tsx 文件上传);
    A --> C(ChatInterface.tsx 聊天界面);
    C --> D(MessageDisplay.tsx 消息展示);
    C --> E(UserInput.tsx 用户输入);
    A --> F(HistoryPanel.tsx 对话历史面板);
```

*   **`App.tsx`**: 应用根组件，负责布局和路由。
*   **`FileUpload.tsx`**: 处理文件上传逻辑和 UI。
*   **`ChatInterface.tsx`**: 核心聊天交互区。
    *   **`MessageDisplay.tsx`**: 展示消息流。
    *   **`UserInput.tsx`**: 用户输入框和发送按钮。
*   **`HistoryPanel.tsx`**: (可选) 展示和管理对话历史。

**3. API 通信层 (`src/services/api.ts`)**

*   封装 `uploadFile(file)` 和 `sendMessage(message, conversationId)` 函数。

**4. 状态管理 (Zustand)**

*   创建 store (`src/store/chatStore.ts`) 管理全局状态。

**5. UI 实现与风格 (Tailwind CSS)**

*   使用 React 组件和 JSX 构建界面。
*   利用 Tailwind CSS 原子类实现“简洁科技风”。

**6. 功能实现流程**

*   **文件上传**: 组件处理 -> 调用 API -> 更新 Store/状态 -> UI 反映。
*   **聊天**: 用户输入 -> 调用 API -> 更新 Store -> UI 反映。
*   **对话历史**: 组件从 Store 读取并渲染。

**7. 开发服务器与代理**

*   使用 `npm run dev` 启动开发服务器。
*   配置 `frontend/vite.config.ts` 代理 `/api/v1/*` 到后端。

**开发计划概览 (Mermaid Gantt):**

```mermaid
gantt
    dateFormat  YYYY-MM-DD
    title DreamHub 前端开发计划 (React)

    section 设置与基础
    项目初始化与配置 (React+Vite) :a1, 2025-04-25, 1d
    安装依赖 (Zustand, Tailwind)  :a2, after a1, 1d
    API 服务封装                 :a3, after a1, 1d
    Vite 代理与 Tailwind 配置     :a4, after a1, 1d

    section 核心功能开发
    文件上传组件 (UI)             :b1, after a2, 1d
    文件上传逻辑 (API)            :b2, after b1 a3, 1d
    聊天界面布局 (UI)             :b3, after a2, 1d
    消息展示组件 (UI)             :b4, after b3, 1d
    用户输入组件 (UI)             :b5, after b3, 1d
    聊天逻辑 (API & State)        :b6, after b4 b5 a3, 2d
    对话历史展示 (UI)             :b7, after b6, 1d

    section 样式与完善
    整体样式调整 (Tailwind)       :c1, after b7, 2d
    交互细节优化                 :c2, after c1, 1d
    测试与 Bug 修复              :c3, after c2, 2d