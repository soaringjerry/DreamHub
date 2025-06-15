import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { syncService, SyncStatus, Conflict } from '../services/syncService';

interface SyncState {
  // Sync status
  syncStatus: SyncStatus | null;
  isSyncing: boolean;
  lastSyncError: string | null;
  conflicts: Conflict[];
  
  // Sync queue for offline changes
  offlineQueue: {
    conversations: any[];
    messages: any[];
    structured_memory: any[];
    user_config: any | null;
    deleted_items: {
      conversation_ids: string[];
      message_ids: string[];
      structured_memory_keys: string[];
    };
  };

  // Actions
  initializeSync: () => Promise<void>;
  performSync: () => Promise<void>;
  addToOfflineQueue: (type: string, data: any) => void;
  clearOfflineQueue: () => void;
  resolveConflict: (conflict: Conflict, resolution: 'local' | 'remote' | 'merge') => Promise<void>;
  resolveAllConflicts: (resolution: 'local' | 'remote') => Promise<void>;
  startAutoSync: () => void;
  stopAutoSync: () => void;
}

export const useSyncStore = create<SyncState>()(
  persist(
    (set, get) => ({
      syncStatus: null,
      isSyncing: false,
      lastSyncError: null,
      conflicts: [],
      offlineQueue: {
        conversations: [],
        messages: [],
        structured_memory: [],
        user_config: null,
        deleted_items: {
          conversation_ids: [],
          message_ids: [],
          structured_memory_keys: []
        }
      },

      initializeSync: async () => {
        try {
          const status = await syncService.getSyncStatus();
          set({ syncStatus: status });
        } catch (error) {
          console.error('Failed to initialize sync:', error);
          set({ lastSyncError: 'Failed to initialize sync' });
        }
      },

      performSync: async () => {
        const state = get();
        if (state.isSyncing) return;

        set({ isSyncing: true, lastSyncError: null });
        
        try {
          // Get offline changes
          const hasOfflineChanges = 
            state.offlineQueue.conversations.length > 0 ||
            state.offlineQueue.messages.length > 0 ||
            state.offlineQueue.structured_memory.length > 0 ||
            state.offlineQueue.user_config !== null ||
            state.offlineQueue.deleted_items.conversation_ids.length > 0 ||
            state.offlineQueue.deleted_items.message_ids.length > 0 ||
            state.offlineQueue.deleted_items.structured_memory_keys.length > 0;

          const localChanges = hasOfflineChanges ? state.offlineQueue : undefined;

          // Perform sync
          const result = await syncService.performSync(localChanges);

          // Handle pulled changes
          if (result.pulled.changes) {
            // TODO: Apply pulled changes to local stores (chat, memory, config)
            // This would involve updating the respective stores with server data
          }

          // Handle conflicts
          const conflicts: Conflict[] = [];
          if (result.pulled.conflicts) {
            conflicts.push(...result.pulled.conflicts);
          }
          if (result.pushed?.conflicts) {
            conflicts.push(...result.pushed.conflicts);
          }

          // Clear offline queue if push was successful
          if (result.pushed?.success) {
            set({ 
              offlineQueue: {
                conversations: [],
                messages: [],
                structured_memory: [],
                user_config: null,
                deleted_items: {
                  conversation_ids: [],
                  message_ids: [],
                  structured_memory_keys: []
                }
              }
            });
          }

          // Update sync status
          const newStatus = await syncService.getSyncStatus();
          set({ 
            syncStatus: newStatus,
            conflicts,
            isSyncing: false 
          });

        } catch (error: any) {
          console.error('Sync failed:', error);
          set({ 
            lastSyncError: error.message || 'Sync failed',
            isSyncing: false 
          });
        }
      },

      addToOfflineQueue: (type: string, data: any) => {
        set((state) => {
          const newQueue = { ...state.offlineQueue };
          
          switch (type) {
            case 'conversation':
              newQueue.conversations.push(data);
              break;
            case 'message':
              newQueue.messages.push(data);
              break;
            case 'structured_memory':
              newQueue.structured_memory.push(data);
              break;
            case 'user_config':
              newQueue.user_config = data;
              break;
            case 'deleted_conversation':
              newQueue.deleted_items.conversation_ids.push(data);
              break;
            case 'deleted_message':
              newQueue.deleted_items.message_ids.push(data);
              break;
            case 'deleted_memory':
              newQueue.deleted_items.structured_memory_keys.push(data);
              break;
          }
          
          return { offlineQueue: newQueue };
        });
      },

      clearOfflineQueue: () => {
        set({
          offlineQueue: {
            conversations: [],
            messages: [],
            structured_memory: [],
            user_config: null,
            deleted_items: {
              conversation_ids: [],
              message_ids: [],
              structured_memory_keys: []
            }
          }
        });
      },

      resolveConflict: async (conflict: Conflict, resolution: 'local' | 'remote' | 'merge') => {
        conflict.resolution = resolution;
        await syncService.resolveConflicts([conflict]);
        
        // Remove resolved conflict from list
        set((state) => ({
          conflicts: state.conflicts.filter(c => 
            c.entity_type !== conflict.entity_type || 
            c.entity_id !== conflict.entity_id
          )
        }));
      },

      resolveAllConflicts: async (resolution: 'local' | 'remote') => {
        const state = get();
        const resolvedConflicts = state.conflicts.map(c => ({
          ...c,
          resolution
        }));
        
        await syncService.resolveConflicts(resolvedConflicts);
        set({ conflicts: [] });
      },

      startAutoSync: () => {
        syncService.startPeriodicSync(30000, (result) => {
          // Handle sync result
          console.log('Auto sync completed:', result);
        });
      },

      stopAutoSync: () => {
        syncService.stopPeriodicSync();
      }
    }),
    {
      name: 'sync-storage',
      partialize: (state) => ({
        offlineQueue: state.offlineQueue
      })
    }
  )
);