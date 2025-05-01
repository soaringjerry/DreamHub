import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware'; // 导入 persist 中间件
import { v4 as uuidv4 } from 'uuid';
import { sendMessage as sendMessageApi, uploadFile as uploadFileApi } from '../services/api';

// --- 类型定义 ---

// 单条消息
interface Message {
  sender: 'user' | 'ai';
  content: string;
  timestamp: number; // 添加时间戳
}

// 单个对话的数据结构
interface Conversation {
  id: string;
  title: string;
  messages: Message[];
  createdAt: number;
  lastUpdatedAt: number; // 添加最后更新时间戳，用于排序
  // 可以添加其他元数据，如关联的文件 ID 等
}

// 上传的文件信息 (保持不变)
interface UploadedFile {
  name: string;
  size: number;
  id: string; // task_id
}

// Store 状态类型
interface ChatState {
  userId: string | null;
  conversations: Record<string, Conversation>; // 存储所有对话 { [conversationId]: Conversation }
  activeConversationId: string | null; // 当前活动的对话 ID
  // 将 loading 和 error 移到 conversation 内部，实现更精细控制
  // isLoading: boolean;
  // error: string | null;
  isUploading: boolean; // 上传状态是全局的
  uploadError: string | null; // 上传错误是全局的
  uploadedFiles: UploadedFile[]; // 已上传文件列表是全局的
  // 用于跟踪特定对话的加载和错误状态
  conversationStatus: Record<string, { isLoading: boolean; error: string | null }>;
}

// Store 操作类型
interface ChatActions {
  setUserId: (userId: string | null) => void;
  startNewConversation: (title?: string) => string; // 创建新对话，可选标题，返回 ID
  switchConversation: (conversationId: string | null) => void; // 切换活动对话
  deleteConversation: (conversationId: string) => void; // 删除对话
  renameConversation: (conversationId: string, newTitle: string) => void; // 重命名对话
  addMessage: (conversationId: string, message: Omit<Message, 'timestamp'>) => void; // 向指定对话添加消息
  sendMessage: (userMessage: string) => Promise<void>; // 发送消息（处理活动对话）
  uploadFile: (file: File) => Promise<void>; // 上传文件
  setConversationStatus: (conversationId: string, status: Partial<{ isLoading: boolean; error: string | null }>) => void; // 设置特定对话的状态
  // 全局状态设置
  setUploading: (uploading: boolean) => void;
  setUploadError: (error: string | null) => void;
  addUploadedFile: (file: UploadedFile) => void;
  clearAllConversations: () => void; // 清空所有对话记录 (可选)
}

// --- Zustand Store 实现 (使用 persist 中间件) ---
export const useChatStore = create<ChatState & ChatActions>()(
  persist(
    (set, get) => ({
      // --- 初始状态 ---
      userId: null,
      conversations: {},
      activeConversationId: null,
      conversationStatus: {}, // 初始化为空对象
      isUploading: false,
      uploadError: null,
      uploadedFiles: [],

      // --- 操作实现 ---
      setUserId: (userId) => set({ userId: userId }),

      startNewConversation: (title) => {
        const newConversationId = uuidv4();
        const now = Date.now();
        const newConversation: Conversation = {
          id: newConversationId,
          title: title || `New Chat ${Object.keys(get().conversations).length + 1}`,
          messages: [],
          createdAt: now,
          lastUpdatedAt: now,
        };
        set((state) => ({
          conversations: {
            ...state.conversations,
            [newConversationId]: newConversation,
          },
          conversationStatus: {
            ...state.conversationStatus,
            [newConversationId]: { isLoading: false, error: null }, // 初始化状态
          },
          activeConversationId: newConversationId, // 设为活动对话
        }));
        console.log("Started new conversation:", newConversationId);
        return newConversationId;
      },

      switchConversation: (conversationId) => {
        if (conversationId === get().activeConversationId) return; // 如果已经是活动对话，则不处理

        if (conversationId === null || get().conversations[conversationId]) {
          set({ activeConversationId: conversationId });
          console.log("Switched to conversation:", conversationId);
        } else {
          console.warn(`Attempted to switch to non-existent conversation: ${conversationId}`);
          // 如果尝试切换到一个不存在的 ID，可以选择切换到 null 或第一个对话
          set({ activeConversationId: null });
        }
      },

      deleteConversation: (conversationId) => {
        set((state) => {
          // eslint-disable-next-line @typescript-eslint/no-unused-vars
          const { [conversationId]: _, ...remainingConversations } = state.conversations;
          // eslint-disable-next-line @typescript-eslint/no-unused-vars
          const { [conversationId]: __, ...remainingStatus } = state.conversationStatus;
          let newActiveId = state.activeConversationId;
          // 如果删除的是当前活动对话，则切换到第一个或 null
          if (state.activeConversationId === conversationId) {
            const remainingIds = Object.keys(remainingConversations);
            newActiveId = remainingIds.length > 0 ? remainingIds[0] : null;
          }
          return {
            conversations: remainingConversations,
            conversationStatus: remainingStatus,
            activeConversationId: newActiveId,
          };
        });
        console.log("Deleted conversation:", conversationId);
      },

       renameConversation: (conversationId, newTitle) => {
         set(state => {
           const conversation = state.conversations[conversationId];
           if (conversation && newTitle.trim()) {
             return {
               conversations: {
                 ...state.conversations,
                 [conversationId]: {
                   ...conversation,
                   title: newTitle.trim(),
                   lastUpdatedAt: Date.now(),
                 }
               }
             }
           }
           return {}; // No change if conversation not found or title is empty
         });
       },

      addMessage: (conversationId, messageData) => {
        set((state) => {
          const targetConversation = state.conversations[conversationId];
          if (!targetConversation) {
            console.error(`Conversation ${conversationId} not found when adding message.`);
            return {};
          }
          const newMessage: Message = {
            ...messageData,
            timestamp: Date.now(),
          };
          const updatedConversation: Conversation = {
            ...targetConversation,
            messages: [...targetConversation.messages, newMessage],
            lastUpdatedAt: newMessage.timestamp, // 更新最后活动时间
          };
          return {
            conversations: {
              ...state.conversations,
              [conversationId]: updatedConversation,
            },
          };
        });
      },

      setConversationStatus: (conversationId, statusUpdate) => {
        set((state) => {
          const currentStatus = state.conversationStatus[conversationId] || { isLoading: false, error: null };
          return {
            conversationStatus: {
              ...state.conversationStatus,
              [conversationId]: { ...currentStatus, ...statusUpdate },
            },
          };
        });
      },

      // 全局状态设置
      setUploading: (uploading) => set({ isUploading: uploading }),
      setUploadError: (error) => set({ uploadError: error }),
      addUploadedFile: (file) => set((state) => ({
        uploadedFiles: [...state.uploadedFiles, file]
      })),

      clearAllConversations: () => set({
          conversations: {},
          conversationStatus: {},
          activeConversationId: null,
      }),

      // --- sendMessage (重构后) ---
      sendMessage: async (userMessage) => {
        const { userId, activeConversationId, startNewConversation, addMessage, setConversationStatus, renameConversation } = get();

        if (!userId) {
          console.error("User ID is missing. Cannot send message.");
          // 如何处理错误？可以设置一个全局错误状态或特定对话的错误
          // 暂时不设置，依赖 UI 层检查 userId
          return;
        }

        let targetConversationId = activeConversationId;
        let isNewConversation = false;

        // 如果没有活动对话，创建一个新的
        if (!targetConversationId) {
          targetConversationId = startNewConversation(); // 创建并设为活动
          isNewConversation = true;
          console.log("No active conversation, started new one:", targetConversationId);
        }

        // 确保 ID 有效
        if (!targetConversationId) {
          console.error("Failed to get or create a conversation ID.");
          // 设置全局错误？
          return;
        }

        // 如果是新对话，用第一条消息设置标题
        if (isNewConversation) {
            const newTitle = userMessage.substring(0, 30) + (userMessage.length > 30 ? '...' : '');
            renameConversation(targetConversationId, newTitle);
        }


        // 1. 添加用户消息 & 设置状态
        addMessage(targetConversationId, { sender: 'user', content: userMessage });
        setConversationStatus(targetConversationId, { isLoading: true, error: null });

        try {
          // 2. 调用 API (Remove userId argument)
          const response = await sendMessageApi(userMessage, targetConversationId);

          // 确认返回的 ID (理论上应该匹配)
          if (response.conversation_id !== targetConversationId) {
            console.warn(`API returned different conversation ID: expected ${targetConversationId}, got ${response.conversation_id}`);
            // 可以选择信任 API 返回的 ID
            // targetConversationId = response.conversation_id;
          }

          // 3. 添加 AI 回复
          addMessage(targetConversationId, { sender: 'ai', content: response.reply });

        } catch (err) {
          console.error("Error in sendMessage store action:", err);
          const errorMsg = err instanceof Error ? err.message : '发送消息时发生未知错误';
          setConversationStatus(targetConversationId, { error: errorMsg });
          // 添加错误消息到 UI
          addMessage(targetConversationId, { sender: 'ai', content: `错误: ${errorMsg}` });
        } finally {
          setConversationStatus(targetConversationId, { isLoading: false });
        }
      },

      // --- uploadFile (重构后) ---
      uploadFile: async (file) => {
        const { userId, activeConversationId, addMessage, setUploading, setUploadError, addUploadedFile, startNewConversation } = get();

        if (!userId) {
          setUploadError("User ID is missing. Cannot upload file.");
          return;
        }

        let targetConversationId = activeConversationId;
        // 如果没有活动对话，创建一个新的用于显示上传状态消息
        if (!targetConversationId) {
            targetConversationId = startNewConversation("File Upload Status"); // 给个默认标题
            console.log("No active conversation for upload message, started new one:", targetConversationId);
        }

        if (!targetConversationId) {
            console.error("Failed to get or create a conversation ID for upload message.");
            setUploadError("Failed to determine conversation for status message.");
            return;
        }

        setUploading(true);
        setUploadError(null);

        try {
          // Call uploadFileApi (Remove userId argument)
          const response = await uploadFileApi(file);

          addUploadedFile({
            name: file.name,
            size: file.size,
            id: response.task_id
          });

          // 上传成功消息添加到目标对话
          addMessage(targetConversationId, { sender: 'ai', content: `文件 "${response.filename}" 已接受处理。 ${response.message}` });

        } catch (err) {
          console.error("Error in uploadFile store action:", err);
          const errorMsg = err instanceof Error ? err.message : '上传文件时发生未知错误';
          setUploadError(errorMsg); // 设置全局上传错误
          // 上传失败消息添加到目标对话
          addMessage(targetConversationId, { sender: 'ai', content: `文件上传失败: ${errorMsg}` });
        } finally {
          setUploading(false);
        }
      },
    }),
    {
      name: 'chat-storage', // localStorage 中的 key
      storage: createJSONStorage(() => localStorage), // 使用 localStorage
      // 可以选择性地持久化部分状态
      partialize: (state) => ({
        userId: state.userId,
        conversations: state.conversations,
        activeConversationId: state.activeConversationId,
        uploadedFiles: state.uploadedFiles, // 可能也需要持久化已上传文件列表
      }),
    }
  )
);

// --- 选择器 (Selectors) ---

// 定义一个稳定的空数组引用
const EMPTY_MESSAGES: Message[] = [];
// 定义一个稳定的默认状态对象引用
const DEFAULT_CONVERSATION_STATUS = { isLoading: false, error: null };

// 获取当前活动对话的 ID
export const useActiveConversationId = () => useChatStore((state) => state.activeConversationId);

// 获取当前活动对话的完整数据
export const useActiveConversation = () => useChatStore((state) =>
  state.activeConversationId ? state.conversations[state.activeConversationId] : null
);

// 获取当前活动对话的消息列表
export const useActiveMessages = () => useChatStore((state) =>
  state.activeConversationId
    ? state.conversations[state.activeConversationId]?.messages ?? EMPTY_MESSAGES // 返回常量空数组
    : EMPTY_MESSAGES // 返回常量空数组
);

// 获取特定对话的状态 (Loading 和 Error)
export const useConversationStatus = (conversationId: string | null) => useChatStore((state) =>
  conversationId ? state.conversationStatus[conversationId] ?? { isLoading: false, error: null } : { isLoading: false, error: null }
);

// 获取当前活动对话的状态
export const useActiveConversationStatus = () => useChatStore((state) =>
    state.activeConversationId
        ? state.conversationStatus[state.activeConversationId] ?? DEFAULT_CONVERSATION_STATUS // 返回常量默认状态
        : DEFAULT_CONVERSATION_STATUS // 返回常量默认状态
);