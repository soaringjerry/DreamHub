import { create } from 'zustand'; // Removed unused StateCreator
import { persist, createJSONStorage, PersistOptions } from 'zustand/middleware'; // Import PersistOptions
import { useAuthStore } from './authStore'; // Import auth store
// Removed unused uuidv4 import
import {
  sendMessage as sendMessageApi,
  uploadFile as uploadFileApi,
  getUserConversations,
  getConversationMessages,
  ConversationInfo,
  Message as ApiMessage
} from '../services/api';

// --- Type Definitions (Aligned with backend/api.ts) ---

export interface Message {
  id: string;
  conversation_id: string;
  role: 'user' | 'assistant';
  content: string;
  created_at: string; // ISO string format
}

export interface Conversation {
  id: string;
  title: string;
  messages: Message[];
  createdAt: string; // ISO string format
  updatedAt: string; // ISO string format
}

interface UploadedFile {
  name: string;
  size: number;
  id: string; // task_id
}

// Store State Type
interface ChatState {
  userId: string | null;
  conversations: Record<string, Conversation>;
  activeConversationId: string | null;
  isUploading: boolean;
  uploadError: string | null;
  uploadedFiles: UploadedFile[];
  conversationStatus: Record<string, { isLoading: boolean; error: string | null; hasLoadedMessages: boolean }>;
  isConversationListLoading: boolean;
  conversationListError: string | null;
}

// Store Actions Type
interface ChatActions {
  setUserId: (userId: string | null) => void;
  startNewConversation: () => void;
  switchConversation: (conversationId: string | null) => void;
  deleteConversation: (conversationId: string) => void;
  renameConversation: (conversationId: string, newTitle: string) => void;
  addMessage: (conversationId: string, message: Message) => void;
  sendMessage: (userMessage: string) => Promise<void>;
  uploadFile: (file: File) => Promise<void>;
  setConversationStatus: (conversationId: string, status: Partial<{ isLoading: boolean; error: string | null; hasLoadedMessages: boolean }>) => void;
  setIsConversationListLoading: (isLoading: boolean) => void;
  setConversationListError: (error: string | null) => void;
  fetchConversations: () => Promise<void>;
  fetchMessages: (conversationId: string) => Promise<void>;
  setUploading: (uploading: boolean) => void;
  setUploadError: (error: string | null) => void;
  addUploadedFile: (file: UploadedFile) => void;
  clearAllConversations: () => void;
}

// Combine State and Actions for the store type
type ChatStore = ChatState & ChatActions;

// Removed unused ChatStateCreator type definition

// Define the shape of the state that will be persisted
interface PersistedChatState {
  userId: string | null;
  activeConversationId: string | null;
  uploadedFiles: UploadedFile[];
}

// Define persist options with explicit type for the persisted state shape
const persistOptions: PersistOptions<ChatStore, PersistedChatState> = {
  name: 'chat-storage', // localStorage key
  storage: createJSONStorage(() => localStorage),
  // Persist only non-volatile state
  partialize: (state: ChatStore): PersistedChatState => ({ // Return the defined persisted shape
    userId: state.userId,
    activeConversationId: state.activeConversationId,
    uploadedFiles: state.uploadedFiles,
  }),
};

// --- Zustand Store Implementation ---
export const useChatStore = create<ChatStore>()( // Use combined type
  persist(
    (set, get): ChatStore => ({ // Ensure return type matches ChatStore
      // --- Initial State (implements ChatState) ---
      userId: null,
      conversations: {},
      activeConversationId: null,
      conversationStatus: {},
      isUploading: false,
      uploadError: null,
      uploadedFiles: [],
      isConversationListLoading: false,
      conversationListError: null,

      // --- Actions Implementation (implements ChatActions) ---
      setUserId: (userId: string | null) => set({ userId: userId }), // Add type

      startNewConversation: () => {
          set({ activeConversationId: null });
          console.log("Set state for new conversation (will be created on first message)");
      },

      switchConversation: (conversationId: string | null) => { // Add type
        const { activeConversationId, conversations, conversationStatus } = get(); // fetchMessages is an action, access via get() if needed inside
        if (conversationId === activeConversationId) return;

        if (conversationId === null || conversations[conversationId]) {
          set({ activeConversationId: conversationId });
          console.log("Switched to conversation:", conversationId);

          if (conversationId && (!conversationStatus[conversationId]?.hasLoadedMessages || conversations[conversationId]?.messages.length === 0)) {
             if (!conversationStatus[conversationId]?.isLoading) {
                console.log(`Messages for ${conversationId} not loaded, fetching...`);
                get().fetchMessages(conversationId); // Call fetchMessages via get()
             }
          }
        } else {
          console.warn(`Attempted to switch to non-existent conversation: ${conversationId}`);
          set({ activeConversationId: null });
        }
      },

      deleteConversation: (conversationId: string) => { // Add type
        console.warn("Frontend deleteConversation called, but backend deletion is not implemented yet.");
        set((state: ChatStore) => { // Use combined type
          // 使用 delete 操作符直接删除属性
          const remainingConversations = { ...state.conversations };
          const remainingStatus = { ...state.conversationStatus };
          delete remainingConversations[conversationId];
          delete remainingStatus[conversationId];
          let newActiveId = state.activeConversationId;
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
        console.log("Deleted conversation locally:", conversationId);
      },

       renameConversation: (conversationId: string, newTitle: string) => { // Add types
         console.warn("Frontend renameConversation called, but backend renaming is not implemented yet.");
         set((state: ChatStore) => { // Use combined type
           const conversation = state.conversations[conversationId];
           if (conversation && newTitle.trim()) {
             return {
               conversations: {
                 ...state.conversations,
                 [conversationId]: {
                   ...conversation,
                   title: newTitle.trim(),
                   updatedAt: new Date().toISOString(),
                 }
               }
             }
           }
           return {};
         });
       },

       addMessage: (conversationId: string, message: Message) => { // Add type
         set((state: ChatStore) => { // Use combined type
           const targetConversation = state.conversations[conversationId];
           if (!targetConversation) {
             console.error(`Conversation ${conversationId} not found when adding message.`);
             return {};
           }
           if (targetConversation.messages.some((m: Message) => m.id === message.id)) {
             console.log(`Message ${message.id} already exists in conversation ${conversationId}. Skipping add.`);
             return {};
           }
           const updatedMessages = [...targetConversation.messages, message].sort(
             (a: Message, b: Message) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
           );
           const latestTimestamp = message.created_at > targetConversation.updatedAt
             ? message.created_at
             : targetConversation.updatedAt;
           const updatedConversation: Conversation = {
             ...targetConversation,
             messages: updatedMessages,
             updatedAt: latestTimestamp,
           };
           return {
             conversations: {
               ...state.conversations,
               [conversationId]: updatedConversation,
             },
           };
         });
       },

       setConversationStatus: (conversationId: string, statusUpdate: Partial<{ isLoading: boolean; error: string | null; hasLoadedMessages: boolean }>) => { // Add type
         set((state: ChatStore) => { // Use combined type
           const currentStatus = state.conversationStatus[conversationId] || { isLoading: false, error: null, hasLoadedMessages: false };
           return {
             conversationStatus: {
               ...state.conversationStatus,
               [conversationId]: { ...currentStatus, ...statusUpdate },
             },
           };
         });
       },

       setIsConversationListLoading: (isLoading: boolean) => set({ isConversationListLoading: isLoading }), // Add type
       setConversationListError: (error: string | null) => set({ conversationListError: error }), // Add type

       setUploading: (uploading: boolean) => set({ isUploading: uploading }), // Add type
       setUploadError: (error: string | null) => set({ uploadError: error }), // Add type
       addUploadedFile: (file: UploadedFile) => set((state: ChatStore) => ({ // Add type and use combined type
         uploadedFiles: [...state.uploadedFiles, file]
       })),

      clearAllConversations: () => {
          console.warn("Frontend clearAllConversations called, but backend clearing is not implemented yet.");
          set({
              conversations: {},
              conversationStatus: {},
              activeConversationId: null,
              conversationListError: null,
              isConversationListLoading: false,
          });
          console.log("Cleared local conversations.");
      },

      fetchConversations: async () => {
        // Access actions via get()
        get().setIsConversationListLoading(true);
        get().setConversationListError(null);
        console.log("Fetching conversations...");
        try {
          const conversationInfos: ConversationInfo[] = await getUserConversations();
          console.log("API返回的对话列表:", conversationInfos);
          console.log("API返回的对话列表类型:", typeof conversationInfos, Array.isArray(conversationInfos));
          
          const newConversations: Record<string, Conversation> = {};
          const newConversationStatus: Record<string, { isLoading: boolean; error: string | null; hasLoadedMessages: boolean }> = {};

          // 确保conversationInfos是数组
          if (Array.isArray(conversationInfos)) {
            conversationInfos.forEach((info: ConversationInfo) => {
              newConversations[info.id] = {
                id: info.id,
                title: info.title,
                messages: [],
                createdAt: info.created_at,
                updatedAt: info.updated_at,
              };
              newConversationStatus[info.id] = { isLoading: false, error: null, hasLoadedMessages: false };
            });
          } else {
            console.error("API返回的对话列表不是数组:", conversationInfos);
          }

          console.log("转换后的conversations对象:", newConversations);
          console.log("转换后的conversations类型:", typeof newConversations);

          set((state: ChatStore) => { // Use combined type
            let newActiveId = state.activeConversationId;
            // Check if the current active ID is still valid in the newly fetched list
            if (newActiveId && !newConversations[newActiveId]) {
              console.warn(`Active conversation ${newActiveId} not found after fetch. Resetting activeConversationId to null.`);
              newActiveId = null; // Reset to null
            }
            
            // 确保返回的conversations始终是一个对象
            const updatedConversations = typeof newConversations === 'object' && newConversations !== null
              ? newConversations
              : {};
              
            return {
              conversations: updatedConversations,
              conversationStatus: { ...state.conversationStatus, ...newConversationStatus },
              isConversationListLoading: false,
              activeConversationId: newActiveId, // Update activeConversationId if needed
            };
          });
          console.log(`Fetched and updated ${Object.keys(newConversations).length} conversations.`);

        } catch (error) {
          console.error("Failed to fetch conversations:", error);
          const errorMsg = error instanceof Error ? error.message : "Unknown error fetching conversations";
          get().setConversationListError(errorMsg); // Use get()
          get().setIsConversationListLoading(false); // Use get()
        }
      },

      fetchMessages: async (conversationId: string) => { // Add type
        get().setConversationStatus(conversationId, { isLoading: true, error: null }); // Use get()
        console.log(`Fetching messages for conversation ${conversationId}...`);
        try {
          const apiMessages: ApiMessage[] = await getConversationMessages(conversationId);
          const localMessages: Message[] = apiMessages.map((m: ApiMessage) => ({
            id: m.id,
            conversation_id: m.conversation_id,
            role: m.role,
            content: m.content,
            created_at: m.created_at,
          }));

          set((state: ChatStore) => { // Use combined type
              const targetConversation = state.conversations[conversationId];
              if (!targetConversation) {
                  console.warn(`Conversation ${conversationId} not found in state during fetchMessages completion.`);
                  return {};
              }
              const latestMessageTime = localMessages.reduce((latest, msg) => {
                  return msg.created_at > latest ? msg.created_at : latest;
              }, targetConversation.updatedAt);

              return {
                  conversations: {
                      ...state.conversations,
                      [conversationId]: {
                          ...targetConversation,
                          messages: localMessages,
                          updatedAt: latestMessageTime,
                      }
                  }
              }
          })

          get().setConversationStatus(conversationId, { isLoading: false, hasLoadedMessages: true }); // Use get()
          console.log(`Fetched ${localMessages.length} messages for conversation ${conversationId}`);

        } catch (error) {
          console.error(`Failed to fetch messages for ${conversationId}:`, error);
          const errorMsg = error instanceof Error ? error.message : "Unknown error fetching messages";
          get().setConversationStatus(conversationId, { isLoading: false, error: errorMsg, hasLoadedMessages: false }); // Use get()
        }
      },

      sendMessage: async (userMessage: string) => { // Add type
        // Get userId from authStore, not chatStore's potentially stale state
        const userId = useAuthStore.getState().user?.id;
        const { activeConversationId } = get(); // Keep getting activeConversationId from chatStore

        if (!userId) {
          console.error("User ID is missing from authStore. Cannot send message.");
          return;
        }

        const currentConversationId = activeConversationId;

        if (currentConversationId) {
             get().setConversationStatus(currentConversationId, { isLoading: true, error: null }); // Use get()
             console.log(`Sending message to existing conversation: ${currentConversationId}`);
        } else {
            console.log("Sending first message for a new conversation...");
        }

        try {
          const response = await sendMessageApi(userMessage, currentConversationId ?? undefined);
          const returnedConversationId = response.conversation_id;
          console.log(`API responded for conversation: ${returnedConversationId}`);

          if (!currentConversationId) {
            console.log(`New conversation ${returnedConversationId} created by backend.`);
            await get().fetchConversations(); // Use get()
            set({ activeConversationId: returnedConversationId });
            await get().fetchMessages(returnedConversationId); // Use get()
          } else {
            if (returnedConversationId !== currentConversationId) {
              console.warn(`API returned different conversation ID (${returnedConversationId}) than expected (${currentConversationId}). Refreshing list and switching.`);
              await get().fetchConversations(); // Use get()
              set({ activeConversationId: returnedConversationId });
              await get().fetchMessages(returnedConversationId); // Use get()
            } else {
              console.log(`Fetching updated messages for ${currentConversationId}...`);
              await get().fetchMessages(currentConversationId); // Use get()
            }
          }

        } catch (err) {
          console.error("Error in sendMessage store action:", err);
          const errorMsg = err instanceof Error ? err.message : '发送消息时发生未知错误';
          if (currentConversationId) {
            get().setConversationStatus(currentConversationId, { error: errorMsg, isLoading: false }); // Use get()
          } else {
              get().setConversationListError(`Failed to start conversation: ${errorMsg}`); // Use get()
              console.error("Error sending first message:", errorMsg);
          }
        } finally {
           const finalActiveId = get().activeConversationId;
           if (finalActiveId && get().conversationStatus[finalActiveId]) {
               get().setConversationStatus(finalActiveId, { isLoading: false }); // Use get()
           }
           console.log("sendMessage finished.");
        }
      },

      uploadFile: async (file: File) => { // Add type
        const { userId, activeConversationId } = get(); // Get state parts needed initially

        if (!userId) {
          get().setUploadError("User ID is missing. Cannot upload file."); // Use get()
          return;
        }

        const targetConversationId = activeConversationId;

        get().setUploading(true); // Use get()
        get().setUploadError(null); // Use get()
        console.log(`Uploading file: ${file.name}`);

        try {
          const response = await uploadFileApi(file);
          console.log("File upload API success:", response);

          get().addUploadedFile({ // Use get()
            name: file.name,
            size: file.size,
            id: response.task_id
          });

          const successMsg = `文件 "${response.filename}" 已接受处理。 ${response.message}`;
          if (targetConversationId) {
            console.log(`Upload successful. Refreshing messages for conversation ${targetConversationId} to show status.`);
            await get().fetchMessages(targetConversationId); // Use get()
          } else {
            console.log("Global upload success (no active conversation):", successMsg);
          }

        } catch (err) {
          console.error("Error in uploadFile store action:", err);
          const errorMsg = err instanceof Error ? err.message : '上传文件时发生未知错误';
          get().setUploadError(errorMsg); // Use get()

          const failureMsg = `文件上传失败: ${errorMsg}`;
           if (targetConversationId) {
            console.log(`Upload failed. Refreshing messages for conversation ${targetConversationId} in case of error message.`);
             await get().fetchMessages(targetConversationId); // Use get()
             get().setConversationStatus(targetConversationId, { error: `Upload failed: ${errorMsg}` }); // Use get()
          } else {
             console.log("Global upload failure (no active conversation):", failureMsg);
          }
        } finally {
          get().setUploading(false); // Use get()
          console.log("uploadFile finished.");
        }
      },
    }),
    persistOptions // Pass typed persist options
  )
);

// --- Selectors ---

const EMPTY_MESSAGES: Message[] = [];
const DEFAULT_CONVERSATION_STATUS = { isLoading: false, error: null, hasLoadedMessages: false };
const EMPTY_CONVERSATIONS: Conversation[] = []; // Stable empty array reference

export const useUserId = () => useChatStore((state) => state.userId);
export const useActiveConversationId = () => useChatStore((state) => state.activeConversationId);
export const useActiveConversation = () => useChatStore((state: ChatStore) =>
  state.activeConversationId ? state.conversations[state.activeConversationId] : null
);
export const useActiveMessages = () => useChatStore((state: ChatStore) => {
  console.log("useActiveMessages - activeConversationId:", state.activeConversationId);
  console.log("useActiveMessages - conversations:", state.conversations);
  console.log("useActiveMessages - conversations type:", typeof state.conversations);
  
  // 防御性检查，确保conversations是一个对象
  if (!state.conversations || typeof state.conversations !== 'object') {
    console.error("conversations不是一个有效的对象:", state.conversations);
    return EMPTY_MESSAGES;
  }
  
  if (!state.activeConversationId) {
    return EMPTY_MESSAGES;
  }
  
  const activeConversation = state.conversations[state.activeConversationId];
  console.log("useActiveMessages - activeConversation:", activeConversation);
  
  if (!activeConversation) {
    console.warn(`找不到ID为${state.activeConversationId}的对话`);
    return EMPTY_MESSAGES;
  }
  
  if (!Array.isArray(activeConversation.messages)) {
    console.error("对话的messages不是一个数组:", activeConversation.messages);
    return EMPTY_MESSAGES;
  }
  
  return activeConversation.messages;
});
export const useConversationStatus = (conversationId: string | null) => useChatStore((state: ChatStore) =>
  conversationId
    ? state.conversationStatus[conversationId] ?? DEFAULT_CONVERSATION_STATUS
    : DEFAULT_CONVERSATION_STATUS
);
export const useActiveConversationStatus = () => useChatStore((state: ChatStore) =>
    state.activeConversationId
        ? state.conversationStatus[state.activeConversationId] ?? DEFAULT_CONVERSATION_STATUS
        : DEFAULT_CONVERSATION_STATUS
);
export const useIsConversationListLoading = () => useChatStore((state) => state.isConversationListLoading);
export const useConversationListError = () => useChatStore((state) => state.conversationListError);
// Selector that returns a sorted array of conversations
export const useSortedConversations = () => useChatStore((state: ChatStore) => {
    // console.log("useSortedConversations selector running");
    if (!state.conversations || typeof state.conversations !== 'object') {
      // console.error("useSortedConversations: conversations is not a valid object:", state.conversations);
      return EMPTY_CONVERSATIONS; // Return stable reference
    }
    try {
      const values = Object.values(state.conversations);
      return values.sort((a: Conversation, b: Conversation) =>
        new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime()
      );
    } catch (error) {
      console.error("Error sorting conversations in useSortedConversations selector:", error);
      return EMPTY_CONVERSATIONS; // Return stable reference
    }
});
export const useIsUploading = () => useChatStore((state) => state.isUploading);
export const useUploadError = () => useChatStore((state) => state.uploadError);
export const useUploadedFiles = () => useChatStore((state) => state.uploadedFiles);