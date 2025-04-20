# DreamHub 项目计划

## 1. 项目名称

DreamHub

## 2. 项目目标

构建一个全能的 AI 驱动工作站/工作面板，旨在通过集成文档管理、自动化工作流和 AI Agent 能力，提高个人和团队的工作效率。

## 3. 核心功能领域

*   **智能文档中心 (Intelligent Document Hub):** 统一管理、智能搜索、AI 驱动的洞察（摘要、问答、分类）。
*   **自动化工作流引擎 (Automation Workflow Engine):** 规则驱动自动化、集成外部服务（邮件、日历等）、定时任务。
*   **AI Agent 协作平台 (AI Agent Collaboration Platform):** 复杂任务委派、工具调用、人机协作、Agent 定制。
*   **统一工作面板 (Unified Workspace Dashboard):** 个性化视图、集中信息展示、通知中心。

## 4. 第一阶段重点 (MVP)

快速搭建可用基础，为后续迭代铺路：

1.  **核心框架搭建:**
    *   后端: Go (使用 Gin 或 Echo 框架)。
    *   前端: React (Next.js) 或 Vue (Nuxt.js) (配合开源 UI 库)。
    *   数据库: PostgreSQL (启用 `pgvector` 扩展)。
    *   文件存储: 本地存储或 MinIO。
    *   用户认证授权。
2.  **基础文档管理:**
    *   文件上传、列表、基础组织（文件夹/标签）。
    *   集成 Go 开源库进行常见格式 (PDF, DOCX) 的文本提取。
3.  **向量化与存储:**
    *   实现文本分块 (Chunking) 策略。
    *   使用开源 Embedding 模型生成文本块向量。
    *   将向量及对应文本块存入 PostgreSQL 的 `pgvector`。
4.  **统一工作面板:**
    *   使用开源 UI 库构建基础仪表板界面。
    *   包含快速访问、最近文件等模块。
5.  **基础 RAG (文档问答) 与记忆机制:**
    *   **集成 LangChainGo** 用于流程编排和记忆管理。
    *   实现类似 ChatGPT 的交互界面（关联文件上下文）。
    *   后端利用 LangChainGo 编排流程：结合 PostgreSQL 标准表进行元数据过滤（可选，精确查找），然后使用 `pgvector` 进行语义相似性搜索以召回相关文档块（语义召回），构建 Prompt 并调用 LLM 生成答案。
    *   利用 LangChainGo 的 Memory 模块管理短期对话记忆（可考虑对接 PostgreSQL 存储长期历史）。

## 5. 技术栈选型

*   **后端:** Go (Gin / Echo)
*   **前端:** React (Next.js) / Vue (Nuxt.js)
*   **UI 库:** (待选) Material UI (MUI), Ant Design, Chakra UI 等
*   **数据库:** PostgreSQL (同时用于结构化数据如元数据、对话历史、配置，并启用 `pgvector` 扩展进行向量存储与搜索)
*   **文档解析:** Go 开源库 (如 `pdfcpu`, `go-docx`, 或其他，需调研选型)
*   **AI 核心:** LangChainGo (用于 RAG 流程编排、Memory 管理、Agent 构建), 开源 Embedding 模型, LLM API (OpenAI, Gemini 等)
*   **文件存储:** MinIO / 本地文件系统 / 云对象存储

## 6. 开发理念

*   **快速迭代:** 优先交付核心价值，小步快跑。
*   **拥抱开源:** 最大化利用成熟的开源项目 (框架、库、工具)，避免重复造轮子，降低工作量。
*   **混合记忆策略:** 结合结构化存储（精确查找）和向量存储（语义召回）解决 AI 记忆和信息调用问题。
*   **模块化设计:** 各组件尽可能解耦，方便独立开发、测试和扩展。
*   **用户中心:** 关注实际用户场景和体验。

## 7. 阶段一简化架构图 (保持不变，细节体现在实现中)

```mermaid
graph TD
    User --> FrontendUI(Web UI - Dashboard, Docs, Chat Input)
    FrontendUI <-- REST API --> BackendAPI(Go + LangChainGo)

    BackendAPI --> Database[(PostgreSQL + pgvector)]
    BackendAPI --> FileStorage[(File Storage)]
    BackendAPI --> EmbeddingModel(Embedding Model API/Service)
    BackendAPI --> LLM(LLM API)

    subgraph Core Infrastructure (Phase 1)
        BackendAPI
        Database
        FileStorage
        EmbeddingModel
        LLM
    end

    subgraph User Interface (Phase 1)
        FrontendUI
    end