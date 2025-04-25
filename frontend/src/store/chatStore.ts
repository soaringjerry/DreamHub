import { create } from 'zustand';
import { sendMessage as sendMessageApi, uploadFile as uploadFileApi } from '../services/api'; // 从 API 服务导入

// 定义单条消息的类型
interface Message {
  sender: 'user' | 'ai';
  content: string;
}

// 定义上传文件的类型 (移除 chunks)
// TODO: Decide if chunks info is needed later (e.g., via task status polling)
interface UploadedFile {
  name: string;
  size: number;
  // chunks: number; // Removed as it's not directly available on upload
  id: string; // Represents the task_id or a generated ID
}

// 定义 Store 的状态类型
interface ChatState {
  conversationId: string | null;
  userId: string | null; // 添加 userId 状态
  messages: Message[];
  isLoading: boolean;
  error: string | null;
  isUploading: boolean; // 添加上传状态
  uploadError: string | null; // 添加上传错误状态
  uploadedFiles: UploadedFile[]; // 添加已上传文件列表
}

// 定义 Store 的操作类型
interface ChatActions {
  addMessage: (message: Message) => void;
  setUserId: (userId: string | null) => void; // 添加设置 userId 的 action
  sendMessage: (userMessage: string) => Promise<void>;
  uploadFile: (file: File) => Promise<void>; // 添加上传文件的 action
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  resetChat: () => void; // 添加重置聊天的 action
  setUploading: (uploading: boolean) => void; // 添加设置上传状态的 action
  setUploadError: (error: string | null) => void; // 添加设置上传错误的 action
  addUploadedFile: (file: UploadedFile) => void; // 添加上传文件到列表
}

// 创建 Zustand Store
export const useChatStore = create<ChatState & ChatActions>((set, get) => ({
  // 初始状态
  conversationId: null,
  userId: null, // 初始化 userId
  messages: [],
  isLoading: false,
  error: null,
  isUploading: false,
  uploadError: null,
  uploadedFiles: [], // 初始化空数组

  // 操作实现
  addMessage: (message) => set((state) => ({ messages: [...state.messages, message] })),

  setUserId: (userId) => set({ userId: userId }), // 实现 setUserId

  setLoading: (loading) => set({ isLoading: loading }),

  setError: (error) => set({ error: error }),

  setUploading: (uploading) => set({ isUploading: uploading }),

  setUploadError: (error) => set({ uploadError: error }),

  addUploadedFile: (file) => set((state) => ({
    uploadedFiles: [...state.uploadedFiles, file]
  })),

  sendMessage: async (userMessage) => {
    const { addMessage, setLoading, setError, conversationId, userId } = get(); // 获取 userId

    // 检查 userId 是否存在
    if (!userId) {
      setError("User ID is missing. Cannot send message.");
      addMessage({ sender: 'ai', content: '消息发送失败：缺少用户ID。' });
      return; // 阻止执行
    }

    // 1. 添加用户消息到列表
    addMessage({ sender: 'user', content: userMessage });
    setLoading(true);
    setError(null); // 清除之前的错误

    try {
      // 2. 调用 API 发送消息
      const response = await sendMessageApi(userMessage, conversationId ?? undefined, userId); // 传递 userId (已确认需要)

      // 3. 添加 AI 回复到列表
      addMessage({ sender: 'ai', content: response.reply });

      // 4. 更新 conversationId (如果是新对话)
      if (!conversationId) {
        set({ conversationId: response.conversation_id });
      }
    } catch (err) {
      console.error("Error in sendMessage store action:", err);
      setError(err instanceof Error ? err.message : '发送消息时发生未知错误');
    } finally {
      setLoading(false);
    }
  },

  uploadFile: async (file) => {
    const { setUploading, setUploadError, addMessage, addUploadedFile, userId } = get(); // 获取 userId

    // 检查 userId 是否存在
    if (!userId) {
      setUploadError("User ID is missing. Cannot upload file.");
      addMessage({ sender: 'ai', content: '文件上传失败：缺少用户ID。' });
      return; // 阻止执行
    }
    setUploading(true);
    setUploadError(null);

    try {
      const response = await uploadFileApi(file, userId); // 传递 userId (已确认需要)

      // 添加到已上传文件列表 (移除 chunks)
      // 注意：这里的 UploadedFile 接口也需要更新，暂时先移除 chunks
      // TODO: Update UploadedFile interface if needed, or remove chunks property
      addUploadedFile({
        name: file.name,
        size: file.size,
        // chunks: 0, // Removed this line to fix TS error
        id: response.task_id // Use task_id from response if available, or generate one
      });

      // 上传成功后添加一条系统消息 (使用 API 返回的消息)
      addMessage({ sender: 'ai', content: `文件 "${response.filename}" 已接受处理。 ${response.message}` });
    } catch (err) {
      console.error("Error in uploadFile store action:", err);
      setUploadError(err instanceof Error ? err.message : '上传文件时发生未知错误');
      // 上传失败后添加一条系统消息
      addMessage({ sender: 'ai', content: `文件上传失败: ${err instanceof Error ? err.message : '未知错误'}` });
    } finally {
      setUploading(false);
    }
  },

  resetChat: () => set({
    conversationId: null,
    // userId: null, // 通常重置聊天不应重置用户ID，除非需要重新登录等场景
    messages: [],
    isLoading: false,
    error: null,
    isUploading: false,
    uploadError: null,
    uploadedFiles: [], // 清空已上传文件列表
  }),
}));