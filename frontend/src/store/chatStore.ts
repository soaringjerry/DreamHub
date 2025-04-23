import { create } from 'zustand';
import { sendMessage as sendMessageApi, uploadFile as uploadFileApi } from '../services/api'; // 从 API 服务导入

// 定义单条消息的类型
interface Message {
  sender: 'user' | 'ai';
  content: string;
}

// 定义 Store 的状态类型
interface ChatState {
  conversationId: string | null;
  messages: Message[];
  isLoading: boolean;
  error: string | null;
  isUploading: boolean; // 添加上传状态
  uploadError: string | null; // 添加上传错误状态
}

// 定义 Store 的操作类型
interface ChatActions {
  addMessage: (message: Message) => void;
  sendMessage: (userMessage: string) => Promise<void>;
  uploadFile: (file: File) => Promise<void>; // 添加上传文件的 action
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  resetChat: () => void; // 添加重置聊天的 action
  setUploading: (uploading: boolean) => void; // 添加设置上传状态的 action
  setUploadError: (error: string | null) => void; // 添加设置上传错误的 action
}

// 创建 Zustand Store
export const useChatStore = create<ChatState & ChatActions>((set, get) => ({
  // 初始状态
  conversationId: null,
  messages: [],
  isLoading: false,
  error: null,
  isUploading: false,
  uploadError: null,

  // 操作实现
  addMessage: (message) => set((state) => ({ messages: [...state.messages, message] })),

  setLoading: (loading) => set({ isLoading: loading }),

  setError: (error) => set({ error: error }),

  setUploading: (uploading) => set({ isUploading: uploading }),

  setUploadError: (error) => set({ uploadError: error }),

  sendMessage: async (userMessage) => {
    const { addMessage, setLoading, setError, conversationId } = get();

    // 1. 添加用户消息到列表
    addMessage({ sender: 'user', content: userMessage });
    setLoading(true);
    setError(null); // 清除之前的错误

    try {
      // 2. 调用 API 发送消息
      const response = await sendMessageApi(userMessage, conversationId ?? undefined); // 使用 ?? 确保传递 undefined 而不是 null

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
    const { setUploading, setUploadError, addMessage } = get();
    setUploading(true);
    setUploadError(null);

    try {
      const response = await uploadFileApi(file);
      // 可选：上传成功后添加一条系统消息
      addMessage({ sender: 'ai', content: `文件 "${response.filename}" 上传成功，包含 ${response.chunks} 个块。` });
    } catch (err) {
      console.error("Error in uploadFile store action:", err);
      setUploadError(err instanceof Error ? err.message : '上传文件时发生未知错误');
       // 可选：上传失败后添加一条系统消息
       addMessage({ sender: 'ai', content: `文件上传失败: ${err instanceof Error ? err.message : '未知错误'}` });
    } finally {
      setUploading(false);
    }
  },

  resetChat: () => set({
    conversationId: null,
    messages: [],
    isLoading: false,
    error: null,
    isUploading: false, // 同时重置上传状态
    uploadError: null,
  }),
}));