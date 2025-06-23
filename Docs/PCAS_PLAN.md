# DreamHub 生态核心: PCAS 架构计划 (v6.0 - Actionable Roadmap)

## 1. 顶层设计与核心原则 (v5.0 Consensus)

本章节总结了我们在此前版本中达成的核心共识。

### 1.1. 核心原则: 数据绝对私有，计算灵活调度
*   **数据绝对私有**: PCAS作为软件，部署在用户的私有环境中。
*   **计算灵活调度**: PCAS内置策略引擎，允许用户在本地计算和云端API计算之间进行选择与平衡。

### 1.2. 三大计算模式
1.  **本地模式 (Local Compute):** 极致隐私。
2.  **混合模式 (Hybrid Compute):** 兼顾隐私与性能。
3.  **云端模式 (Cloud Compute):** 最强算力。

### 1.3. 终极愿景: 从数据熔炉到个人AI
PCAS是一个**数据熔炉**，旨在帮助用户沉淀私有数据集，用以微调或训练个人AI模型。

### 1.4. 社会理想: 建立开放生态与标准
我们致力于构建一个开源、社区驱动的生态系统，并为个人AI领域定义开放标准。

---

## 2. 关键技术决策与落地细则 (o3 Review Integration)

本章节旨在吸收社区核心成员(o3)的专业评审，将高层架构细化为具体的工程决策。

### 2.1. 事件协议 (Protocol Selection)
*   **决策:** 为避免“再造轮子”，我们的事件协议将**兼容 CloudEvents v1.0 核心字段** (`id`, `source`, `specversion`, `type`, `datacontenttype`, `subject`)，并在此基础上增加PCAS特定的扩展字段。所有事件结构将使用 **Protobuf** 进行定义，并包含 `version` 字段。

### 2.2. 策略引擎 (Policy Engine Implementation)
*   **决策:** 采用分阶段落地策略。
    *   **V1:** 实现一个基于YAML配置的**静态规则引擎**，支持基于`数据敏感度Tag`和`LLM token数`等阈值的决策。
    *   **V2:** 探索引入一个轻量级的**领域特定语言(DSL)** 或独立的策略服务，以支持更动态的策略编程。

### 2.3. 本地模型栈 (Local Model Stack)
*   **决策:** PCAS将定义一套标准的**本地模型推理接口**，以支持不同的模型格式（如GGUF, ONNX）。我们将提供一个清晰的**模型生命周期管理流程**（下载-缓存-版本核对-更新），以避免“版本漂移”导致的推理差异。初期将以集成 **Ollama** 作为参考实现。

### 2.4. 核心记忆模型 (Memory & Persistence)
*   **决策:** 采用基于`nodes/edges`表的**统一图模型**，并在V1阶段提供**可插拔的`StorageProvider`**（默认SQLite，可选PostgreSQL）。所有流入“数据熔炉”的决策日志，都必须包含一个**可追溯的`trace_id`**，并接入 **OpenTelemetry** 标准。

### 2.5. 安全与权限 (Security Implementation)
*   **决策:**
    *   **V1:** 采用 **JWT + Scope列表** 作为D-App能力令牌的最小可行实现。
    *   **V2:** 探索引入 **WebAssembly (WASM)** 沙箱来运行高风险的第三方D-App。

### 2.6. 社区与治理 (Governance)
*   **决策:** 为支撑开放生态，我们将建立**技术治理三件套**：
    1.  **代码规范:** 引入严格的Linting规则和CI检查。
    2.  **变更提案流程:** 建立一个轻量级的DIP（DreamHub Improvement Proposal）或ADR（Architecture Decision Record）流程。
    3.  **决策委员会:** 明确项目的核心维护者（Core Maintainers）及其职责。

---

## 3. 前瞻性架构方向 (Advanced Directions)

本章节列出了将使DreamHub保持长期领先性的探索方向。

*   **事件层回滚与补偿:** 在事件协议中加入`compensation`描述（Saga模式），为高风险的写操作提供撤销能力。
*   **隐私等级自动标注:** 训练一个本地分类器，自动为流入的数据打上隐私等级标签，供策略引擎使用。
*   **D-App生命周期管理:** 借鉴VS Code插件模式，为D-App设计包含版本、权限、依赖等信息的`manifest.json`，并由PCAS内的`Extension Host`负责热加载与管理。
*   **端到端加密与备份:** 默认对整个本地数据库进行AES-GCM加密（由用户自管密钥），并提供一键加密备份到用户自托管对象存储的功能。

---

## 4. 下一步落地清单 (Actionable Checklist)

**核心原则: 跑通一条链，再上楼。**

| 时间 | 目标 | 里程碑交付物 |
| :--- | :--- | :--- |
| **+2 周** | **最小事件总线 & CLI** | `pcas serve`, `pcas emit` 命令；带`trace_id`；基于JSON Schema的初步校验。 |
| **+1 月** | **Policy v0 + Providers** | `policy.yaml`静态规则；`OpenAIProvider` & `MockLocalProvider` 走通；集成Prometheus指标。 |
| **+2 月** | **可解释决策 + Graph存储** | `LLM-decide()`产出行动计划+决策日志；SQLite两表落库；提供`pcas replay`回放CLI。 |
| **+3 月** | **SDK & 三个示例D-App** | 发布Go/TS SDK；发布Scheduler, Communicator, Knowledge三个核心D-App。 |
| **+4 月** | **Preview Release & 社区开启** | 发布GitHub Beta Tag；建立文档站；启动RFC流程，邀请首批贡献者。 |