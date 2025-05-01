import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
// import LanguageDetector from 'i18next-browser-languagedetector'; // Removed unused import

i18n
  // 将 i18n 实例传递给 react-i18next
  .use(initReactI18next)
  // 初始化 i18next
  .init({
    lng: 'en', // 设置默认语言为英文
    fallbackLng: 'en', // 如果某个键在当前语言中不存在，则回退到英文
    debug: true, // 开发环境中开启 debug 模式
    interpolation: {
      escapeValue: false, // react 已经处理了 XSS，所以禁用转义
    },
    // 可以在这里直接定义资源，或者之后使用 backend 加载
    resources: {
      en: {
        translation: {
          // 在这里添加英文翻译
          greeting: "Hello!",
          appTitle: "DreamHub",
          onlineStatus: "Online",
          switchToLightMode: "Switch to Light Mode",
          switchToDarkMode: "Switch to Dark Mode",
          fileUploadTitle: "File Upload",
          aiAssistantTitle: "AI Assistant",
          aiAssistantSubtitle: "Chat with your documents",
          footerText: "© {{year}} DreamHub",
          // FileUpload translations
          processedFiles: "Processed Files",
          fileDropzoneAreaLabel: "File dropzone area",
          dropFileToUpload: "Drop file here to upload",
          fileReady: "File ready",
          dragOrClickToSelect: "Drag & drop or click to select",
          supportedFormats: "Supports TXT, PDF, DOCX",
          removeFileLabel: "Remove file",
          uploadSuccessMessage: "Upload successful! File processed.",
          processingButton: "Processing...",
          doneButton: "Done",
          fileReadyButton: "File Ready",
          uploadAndProcessButton: "Upload & Process",
          uploadErrorTitle: "Upload Error",
          // ChatInterface translations
          resetChatConfirmation: "Are you sure you want to clear the current conversation? This action cannot be undone.",
          resetChatLabel: "Reset Conversation",
          aiThinking: "AI is thinking...",
          // MessageDisplay translations
          welcomeTitle: "Welcome to DreamHub",
          welcomeMessage: "After uploading documents, you can start asking questions, and the AI will answer based on the document content.",
          fileUploadSuccessPrefix: "File upload successful: ", // Note the space at the end
          // UserInput translations
          showQuickPromptsLabel: "Show quick prompts",
          quickPromptsTitle: "Quick Prompts",
          inputPlaceholderReady: "Ask a question to explore your documents...",
          inputPlaceholderUpload: "Upload documents to start asking questions...",
          sendMessageLabel: "Send message",
          quickPromptSummarize: "Summarize the main content of the document",
          quickPromptExtract: "Extract key information from the document",
          quickPromptExplain: "Explain technical terms in the document",
          quickPromptFindData: "Find important data in the document",
          quickPromptAnalyze: "Analyze the structure and logic of the document",
          filesUploadedHint_one: "{{count}} file uploaded, ready for questions",
          filesUploadedHint_other: "{{count}} files uploaded, ready for questions",
          uploadFirstHint: "Please upload documents before asking questions",
          // App component translations
          languageSelectorLabel: "Select Language",
          userIdNotSet: "ID Not Set",
          setUserIdPlaceholder: "Set User ID",
          saveUserIdButton: "Save User ID",
          setUserIdToUpload: "Please set a User ID before uploading files.",
          // Header translations (besides appTitle)
          "header.settingsLink": "Settings",
          "header.personalizationLink": "Personalization", // Added
          // ConversationList translations
          conversationsTitle: "Conversations",
          newChatLabel: "New Chat",
          noConversations: "No conversations yet.",
          deleteConversationConfirmation: "Are you sure you want to delete this conversation?",
          deleteConversationLabel: "Delete conversation",
          loadingConversations: "Loading...", // Added
          errorLoadingConversationsTitle: "Error Loading Conversations", // Added
          retry: "Retry", // Added
// Auth related
          "auth.usernameLabel": "Username",
          "auth.passwordLabel": "Password",
          "auth.logoutButton": "Logout",
          "auth.loginLink": "Login",
          "auth.registerLink": "Register",
          "auth.pleaseLoginPrompt": "Please log in to access the chat.",
          // Login Page
          "login.title": "Login to DreamHub",
          "login.loadingButton": "Logging in...",
          "login.submitButton": "Login",
          "login.registerPrompt": "Don't have an account?",
          "login.registerLink": "Register here",
          "login.validation.missingCredentials": "Please enter both username and password.",
          // Register Page
          "register.title": "Register for DreamHub",
          "register.confirmPasswordLabel": "Confirm Password",
          "register.loadingButton": "Registering...",
          "register.submitButton": "Register",
          "register.loginPrompt": "Already have an account?",
          "register.loginLink": "Login here",
          "register.validation.allFieldsRequired": "Please fill in all fields.",
          "register.validation.passwordsMismatch": "Passwords do not match.",
          "register.successMessage": "Registration successful! Please login.",
          // Settings Page
          "settings.title": "Settings",
          "settings.description": "Configure your application settings here.", // Added from initial component
          "settings.apiEndpointLabel": "API Endpoint",
          "settings.apiEndpointPlaceholder": "e.g., https://api.example.com/v1",
          "settings.apiEndpointHint": "Optional. Leave blank to use the default.",
          "settings.defaultModelLabel": "Default Model Name",
          "settings.defaultModelPlaceholder": "e.g., gpt-4o",
          "settings.defaultModelHint": "Optional. The default model to use for chats.",
          "settings.apiKeyLabel": "API Key",
          "settings.apiKeyPlaceholder": "Enter your API key",
          "settings.apiKeyPlaceholderUpdate": "Enter new key to update, leave blank to keep current",
         "settings.apiKeyHint": "Optional. Your personal API key for the service.",
          "settings.apiKeyHintUpdate": "An API key is currently set. Enter a new key to replace it, or leave blank to keep the existing one.",
         "settings.resetButton": "Reset",
         "settings.saveButton": "Save Changes",
          "settings.saving": "Saving...",
          "settings.saveSuccess": "Settings saved successfully!",
          "settings.saveError": "Error saving settings",
          "settings.loadingConfig": "Loading configuration...",
          "settings.loadingError": "Error loading configuration",
          // Personalization Page
          "personalizationPage.title": "Personalization (Memory)",
          "personalizationPage.tabs.memory": "Structured Memory",
          "personalizationPage.tabs.knowledge": "Knowledge Base",
          "personalizationPage.tabs.prompt": "Custom Prompt",
          // Structured Memory Tab
          "personalizationPage.memory.title": "Manage Structured Memory",
          "personalizationPage.memory.addTitle": "Add New Entry",
          "personalizationPage.memory.keyPlaceholder": "Enter Key",
          "personalizationPage.memory.valuePlaceholder": "Enter Value (string or JSON)",
          "personalizationPage.memory.addButton": "Add Entry",
          "personalizationPage.memory.addingButton": "Adding...",
          "personalizationPage.memory.loading": "Loading memory entries...",
          "personalizationPage.memory.errorLoading": "Error loading memory entries.",
          "personalizationPage.memory.errorAdding": "Error adding memory entry.",
          "personalizationPage.memory.errorEmptyFields": "Key and Value cannot be empty.",
          "personalizationPage.memory.noEntries": "No memory entries found.",
          "personalizationPage.memory.tableHeaderKey": "Key",
          "personalizationPage.memory.tableHeaderValue": "Value",
          "personalizationPage.memory.tableHeaderActions": "Actions",
          "personalizationPage.memory.editButton": "Edit",
          "personalizationPage.memory.deleteButton": "Delete",
          "personalizationPage.memory.saveButton": "Save",
          "personalizationPage.memory.savingButton": "Saving...",
          "personalizationPage.memory.cancelButton": "Cancel",
          "personalizationPage.memory.deletingButton": "Deleting...",
          "personalizationPage.memory.errorUpdating": "Error updating memory entry.",
          "personalizationPage.memory.errorDeleting": "Error deleting memory entry.",
          "personalizationPage.memory.deleteConfirmation": "Are you sure you want to delete the memory entry with key \"{{key}}\"?",
          // Knowledge Base Tab
          "personalizationPage.knowledge.title": "Manage Knowledge Base",
          "personalizationPage.knowledge.loading": "Loading documents...",
          "personalizationPage.knowledge.errorLoading": "Error loading documents.",
          "personalizationPage.knowledge.noDocuments": "No documents found.",
          "personalizationPage.knowledge.tableHeaderFilename": "Filename",
          "personalizationPage.knowledge.tableHeaderStatus": "Status",
          "personalizationPage.knowledge.tableHeaderUploadedAt": "Uploaded At",
          "personalizationPage.knowledge.tableHeaderActions": "Actions",
          "personalizationPage.knowledge.status.pending": "Pending",
          "personalizationPage.knowledge.status.processing": "Processing",
          "personalizationPage.knowledge.status.completed": "Completed",
          "personalizationPage.knowledge.status.failed": "Failed",
          "personalizationPage.knowledge.deleteButton": "Delete",
          "personalizationPage.knowledge.deletingButton": "Deleting...",
          "personalizationPage.knowledge.errorDeleting": "Error deleting document.",
          "personalizationPage.knowledge.deleteConfirmation": "Are you sure you want to delete the document \"{{filename}}\"?",
          // Custom Prompt Tab
          "personalizationPage.prompt.title": "Customize System Prompt",
          "personalizationPage.prompt.description": "Enter a custom system prompt that will be used by the AI assistant during conversations. Leave blank to use the default.",
          "personalizationPage.prompt.placeholder": "Enter your custom prompt here...",
          "personalizationPage.prompt.loading": "Loading custom prompt...",
          "personalizationPage.prompt.errorLoading": "Error loading custom prompt.",
          "personalizationPage.prompt.saveButton": "Save Prompt",
          "personalizationPage.prompt.savingButton": "Saving...",
          "personalizationPage.prompt.saveSuccess": "Custom prompt saved successfully!",
          "personalizationPage.prompt.errorSaving": "Error saving custom prompt.",
       }
     },
     zh: {
        translation: {
          // 在这里添加中文翻译
          greeting: "你好！",
          appTitle: "DreamHub",
          onlineStatus: "在线",
          switchToLightMode: "切换到亮色模式",
          switchToDarkMode: "切换到深色模式",
          fileUploadTitle: "文件上传",
          aiAssistantTitle: "AI 助手",
          aiAssistantSubtitle: "与您的文档进行对话",
          footerText: "© {{year}} DreamHub",
          // FileUpload translations
          processedFiles: "已处理文件",
          fileDropzoneAreaLabel: "文件拖放区域",
          dropFileToUpload: "释放文件以上传",
          fileReady: "文件已就绪",
          dragOrClickToSelect: "拖放文件或点击选择",
          supportedFormats: "支持 TXT, PDF, DOCX",
          removeFileLabel: "移除文件",
          uploadSuccessMessage: "上传成功！文件已处理。",
          processingButton: "处理中...",
          doneButton: "完成",
          fileReadyButton: "文件已就绪",
          uploadAndProcessButton: "上传并处理",
          uploadErrorTitle: "上传出错",
          // ChatInterface translations
          resetChatConfirmation: "确定要清空当前对话吗？此操作不可撤销。",
          resetChatLabel: "重置对话",
          aiThinking: "AI 正在思考...",
          // MessageDisplay translations
          welcomeTitle: "欢迎使用 DreamHub",
          welcomeMessage: "上传文档后，您可以开始提问，AI 将基于文档内容进行回答。",
          fileUploadSuccessPrefix: "文件上传成功： ", // 注意末尾的空格
          // UserInput translations
          showQuickPromptsLabel: "显示快捷提示",
          quickPromptsTitle: "快捷提示",
          inputPlaceholderReady: "输入问题，探索您的文档内容...",
          inputPlaceholderUpload: "上传文档后开始提问...",
          sendMessageLabel: "发送消息",
          quickPromptSummarize: "总结文档的主要内容",
          quickPromptExtract: "提取文档中的关键信息",
          quickPromptExplain: "解释文档中的专业术语",
          quickPromptFindData: "查找文档中的重要数据",
          quickPromptAnalyze: "分析文档的结构和逻辑",
          filesUploadedHint_one: "已上传 {{count}} 个文件，可以开始提问",
          filesUploadedHint_other: "已上传 {{count}} 个文件，可以开始提问",
          uploadFirstHint: "请先上传文档再开始提问",
          // App component translations
          languageSelectorLabel: "选择语言",
          userIdNotSet: "ID 未设置",
          setUserIdPlaceholder: "设置用户ID",
          saveUserIdButton: "保存用户ID",
          setUserIdToUpload: "请先设置用户ID才能上传文件。",
          // 页眉翻译 (除了 appTitle)
          "header.settingsLink": "设置",
          "header.personalizationLink": "个性化", // Added
          // ConversationList translations
          conversationsTitle: "对话列表",
          newChatLabel: "新建对话",
          noConversations: "暂无对话。",
          deleteConversationConfirmation: "确定要删除此对话吗？",
          deleteConversationLabel: "删除对话",
          loadingConversations: "加载中...", // Added
          errorLoadingConversationsTitle: "加载对话出错", // Added
          retry: "重试", // Added
// Auth related
          "auth.usernameLabel": "用户名",
          "auth.passwordLabel": "密码",
          "auth.logoutButton": "登出",
          "auth.loginLink": "登录",
          "auth.registerLink": "注册",
          "auth.pleaseLoginPrompt": "请登录以访问聊天。",
          // Login Page
          "login.title": "登录 DreamHub",
          "login.loadingButton": "登录中...",
          "login.submitButton": "登录",
          "login.registerPrompt": "还没有账户？",
          "login.registerLink": "在此注册",
          "login.validation.missingCredentials": "请输入用户名和密码。",
          // Register Page
          "register.title": "注册 DreamHub",
          "register.confirmPasswordLabel": "确认密码",
          "register.loadingButton": "注册中...",
          "register.submitButton": "注册",
          "register.loginPrompt": "已经有账户了？",
          "register.loginLink": "在此登录",
          "register.validation.allFieldsRequired": "请填写所有字段。",
          "register.validation.passwordsMismatch": "密码不匹配。",
          "register.successMessage": "注册成功！请登录。",
          // 设置页面
          "settings.title": "设置",
          "settings.description": "在此配置您的应用程序设置。", // 从初始组件添加
          "settings.apiEndpointLabel": "API 端点",
          "settings.apiEndpointPlaceholder": "例如：https://api.example.com/v1",
          "settings.apiEndpointHint": "可选。留空以使用默认值。",
          "settings.defaultModelLabel": "默认模型名称",
          "settings.defaultModelPlaceholder": "例如：gpt-4o",
          "settings.defaultModelHint": "可选。用于聊天的默认模型。",
          "settings.apiKeyLabel": "API 密钥",
          "settings.apiKeyPlaceholder": "输入您的 API 密钥",
          "settings.apiKeyPlaceholderUpdate": "输入新密钥以更新，留空以保留当前密钥",
         "settings.apiKeyHint": "可选。您用于该服务的个人 API 密钥。",
          "settings.apiKeyHintUpdate": "当前已设置 API 密钥。输入新密钥以替换它，或留空以保留现有密钥。",
         "settings.resetButton": "重置",
         "settings.saveButton": "保存更改",
          "settings.saving": "保存中...",
          "settings.saveSuccess": "设置已成功保存！",
          "settings.saveError": "保存设置时出错",
          "settings.loadingConfig": "正在加载配置...",
          "settings.loadingError": "加载配置时出错",
          // 个性化页面
          "personalizationPage.title": "个性化（记忆）",
          "personalizationPage.tabs.memory": "结构化记忆",
          "personalizationPage.tabs.knowledge": "知识库",
          "personalizationPage.tabs.prompt": "自定义指令",
          // 结构化记忆选项卡
          "personalizationPage.memory.title": "管理结构化记忆",
          "personalizationPage.memory.addTitle": "添加新条目",
          "personalizationPage.memory.keyPlaceholder": "输入键",
          "personalizationPage.memory.valuePlaceholder": "输入值（字符串或 JSON）",
          "personalizationPage.memory.addButton": "添加条目",
          "personalizationPage.memory.addingButton": "添加中...",
          "personalizationPage.memory.loading": "正在加载记忆条目...",
          "personalizationPage.memory.errorLoading": "加载记忆条目时出错。",
          "personalizationPage.memory.errorAdding": "添加记忆条目时出错。",
          "personalizationPage.memory.errorEmptyFields": "键和值不能为空。",
          "personalizationPage.memory.noEntries": "未找到记忆条目。",
          "personalizationPage.memory.tableHeaderKey": "键",
          "personalizationPage.memory.tableHeaderValue": "值",
          "personalizationPage.memory.tableHeaderActions": "操作",
          "personalizationPage.memory.editButton": "编辑",
          "personalizationPage.memory.deleteButton": "删除",
          "personalizationPage.memory.saveButton": "保存",
          "personalizationPage.memory.savingButton": "保存中...",
          "personalizationPage.memory.cancelButton": "取消",
          "personalizationPage.memory.deletingButton": "删除中...",
          "personalizationPage.memory.errorUpdating": "更新记忆条目时出错。",
          "personalizationPage.memory.errorDeleting": "删除记忆条目时出错。",
          "personalizationPage.memory.deleteConfirmation": "确定要删除键为 \"{{key}}\" 的记忆条目吗？",
          // 知识库选项卡
          "personalizationPage.knowledge.title": "管理知识库",
          "personalizationPage.knowledge.loading": "正在加载文档...",
          "personalizationPage.knowledge.errorLoading": "加载文档时出错。",
          "personalizationPage.knowledge.noDocuments": "未找到文档。",
          "personalizationPage.knowledge.tableHeaderFilename": "文件名",
          "personalizationPage.knowledge.tableHeaderStatus": "状态",
          "personalizationPage.knowledge.tableHeaderUploadedAt": "上传时间",
          "personalizationPage.knowledge.tableHeaderActions": "操作",
          "personalizationPage.knowledge.status.pending": "待处理",
          "personalizationPage.knowledge.status.processing": "处理中",
          "personalizationPage.knowledge.status.completed": "已完成",
          "personalizationPage.knowledge.status.failed": "失败",
          "personalizationPage.knowledge.deleteButton": "删除",
          "personalizationPage.knowledge.deletingButton": "删除中...",
          "personalizationPage.knowledge.errorDeleting": "删除文档时出错。",
          "personalizationPage.knowledge.deleteConfirmation": "确定要删除文档 \"{{filename}}\" 吗？",
          // 自定义指令选项卡
          "personalizationPage.prompt.title": "自定义系统指令",
          "personalizationPage.prompt.description": "输入自定义的系统指令，AI 助手将在对话中使用该指令。留空则使用默认指令。",
          "personalizationPage.prompt.placeholder": "在此输入您的自定义指令...",
          "personalizationPage.prompt.loading": "正在加载自定义指令...",
          "personalizationPage.prompt.errorLoading": "加载自定义指令时出错。",
          "personalizationPage.prompt.saveButton": "保存指令",
          "personalizationPage.prompt.savingButton": "保存中...",
          "personalizationPage.prompt.saveSuccess": "自定义指令已成功保存！",
          "personalizationPage.prompt.errorSaving": "保存自定义指令时出错。",
       }
     }
   }
  });

export default i18n;