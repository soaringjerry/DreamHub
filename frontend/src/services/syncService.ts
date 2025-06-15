import { api } from './api';

// Types for sync functionality
export interface SyncStatus {
  id: string;
  user_id: string;
  device_id: string;
  last_sync_at: string;
  sync_version: number;
  created_at: string;
  updated_at: string;
}

export interface SyncRequest {
  device_id: string;
  last_sync_at?: string;
  sync_version?: number;
}

export interface SyncChanges {
  conversations?: any[];
  messages?: any[];
  structured_memory?: any[];
  user_config?: any;
  deleted_items?: {
    conversation_ids?: string[];
    message_ids?: string[];
    structured_memory_keys?: string[];
  };
}

export interface SyncResponse {
  sync_version: number;
  changes: SyncChanges;
  conflicts?: Conflict[];
  server_time: string;
}

export interface SyncPushRequest {
  device_id: string;
  sync_version: number;
  changes: SyncChanges;
}

export interface SyncPushResponse {
  success: boolean;
  sync_version: number;
  conflicts?: Conflict[];
  server_time: string;
}

export interface Conflict {
  entity_type: string;
  entity_id: string;
  local_value: any;
  remote_value: any;
  resolution: 'local' | 'remote' | 'merge';
}

// Generate or get device ID
const getDeviceId = (): string => {
  let deviceId = localStorage.getItem('device_id');
  if (!deviceId) {
    // Generate a unique device ID
    deviceId = `web-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
    localStorage.setItem('device_id', deviceId);
  }
  return deviceId;
};

// Sync service class
class SyncService {
  private deviceId: string;
  private syncInProgress: boolean = false;
  private syncInterval: number | null = null;

  constructor() {
    this.deviceId = getDeviceId();
  }

  // Get sync status
  async getSyncStatus(): Promise<SyncStatus> {
    const response = await api.get<SyncStatus>('/sync/status', {
      params: { device_id: this.deviceId }
    });
    return response.data;
  }

  // Pull changes from server
  async pullChanges(): Promise<SyncResponse> {
    const lastSyncData = localStorage.getItem('last_sync');
    const lastSync = lastSyncData ? JSON.parse(lastSyncData) : {};

    const request: SyncRequest = {
      device_id: this.deviceId,
      last_sync_at: lastSync.last_sync_at,
      sync_version: lastSync.sync_version || 0
    };

    const response = await api.post<SyncResponse>('/sync/pull', request);
    
    // Update last sync info
    if (response.data.sync_version) {
      localStorage.setItem('last_sync', JSON.stringify({
        last_sync_at: response.data.server_time,
        sync_version: response.data.sync_version
      }));
    }

    return response.data;
  }

  // Push changes to server
  async pushChanges(changes: SyncChanges): Promise<SyncPushResponse> {
    const lastSyncData = localStorage.getItem('last_sync');
    const lastSync = lastSyncData ? JSON.parse(lastSyncData) : {};

    const request: SyncPushRequest = {
      device_id: this.deviceId,
      sync_version: lastSync.sync_version || 0,
      changes
    };

    const response = await api.post<SyncPushResponse>('/sync/push', request);
    
    // Update sync version if successful
    if (response.data.success && response.data.sync_version) {
      localStorage.setItem('last_sync', JSON.stringify({
        last_sync_at: response.data.server_time,
        sync_version: response.data.sync_version
      }));
    }

    return response.data;
  }

  // Resolve conflicts
  async resolveConflicts(conflicts: Conflict[]): Promise<void> {
    await api.post('/sync/conflicts/resolve', conflicts);
  }

  // Perform a full sync (pull then push)
  async performSync(localChanges?: SyncChanges): Promise<{
    pulled: SyncResponse;
    pushed?: SyncPushResponse;
    hasConflicts: boolean;
  }> {
    if (this.syncInProgress) {
      throw new Error('Sync already in progress');
    }

    this.syncInProgress = true;
    try {
      // First, pull changes from server
      const pulled = await this.pullChanges();

      // If we have local changes, push them
      let pushed: SyncPushResponse | undefined;
      if (localChanges && Object.keys(localChanges).length > 0) {
        pushed = await this.pushChanges(localChanges);
      }

      const hasConflicts = (pulled.conflicts && pulled.conflicts.length > 0) ||
                          (pushed?.conflicts && pushed.conflicts.length > 0);

      return { pulled, pushed, hasConflicts };
    } finally {
      this.syncInProgress = false;
    }
  }

  // Start periodic sync
  startPeriodicSync(intervalMs: number = 30000, onSync?: (result: any) => void): void {
    if (this.syncInterval) {
      this.stopPeriodicSync();
    }

    // Perform initial sync
    this.performSync().then(onSync).catch(console.error);

    // Set up periodic sync
    this.syncInterval = window.setInterval(async () => {
      try {
        const result = await this.performSync();
        if (onSync) {
          onSync(result);
        }
      } catch (error) {
        console.error('Periodic sync failed:', error);
      }
    }, intervalMs);
  }

  // Stop periodic sync
  stopPeriodicSync(): void {
    if (this.syncInterval) {
      clearInterval(this.syncInterval);
      this.syncInterval = null;
    }
  }

  // Check if sync is in progress
  isSyncing(): boolean {
    return this.syncInProgress;
  }
}

// Export singleton instance
export const syncService = new SyncService();