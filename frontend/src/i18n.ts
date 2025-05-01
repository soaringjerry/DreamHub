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
          // ConversationList translations
          conversationsTitle: "Conversations",
          newChatLabel: "New Chat", // Reusing from ChatInterface potentially
          noConversations: "No conversations yet.",
          deleteConversationConfirmation: "Are you sure you want to delete this conversation?",
          deleteConversationLabel: "Delete conversation",
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
          // ConversationList translations
          conversationsTitle: "对话列表",
          newChatLabel: "新建对话", // Reusing from ChatInterface potentially
          noConversations: "暂无对话。",
          deleteConversationConfirmation: "确定要删除此对话吗？",
          deleteConversationLabel: "删除对话",
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
        }
      }
    }
  });

export default i18n;