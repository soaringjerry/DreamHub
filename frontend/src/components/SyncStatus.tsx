import React, { useEffect } from 'react';
import { useSyncStore } from '../store/syncStore';
import { useAuthStore } from '../store/authStore';
import { Cloud, CloudOff, Refresh, AlertCircle, X } from 'lucide-react';

export const SyncStatus: React.FC = () => {
  const {
    syncStatus,
    isSyncing,
    lastSyncError,
    conflicts,
    initializeSync,
    performSync,
    startAutoSync,
    stopAutoSync,
    resolveAllConflicts
  } = useSyncStore();

  const isAuthenticated = useAuthStore(state => state.isAuthenticated);

  // Initialize sync when authenticated
  useEffect(() => {
    if (isAuthenticated) {
      initializeSync().then(() => {
        startAutoSync();
      });
    } else {
      stopAutoSync();
    }

    return () => {
      stopAutoSync();
    };
  }, [isAuthenticated]);

  if (!isAuthenticated) {
    return null;
  }

  const handleManualSync = () => {
    if (!isSyncing) {
      performSync();
    }
  };

  const formatLastSyncTime = () => {
    if (!syncStatus?.last_sync_at) return '从未同步';
    
    const lastSync = new Date(syncStatus.last_sync_at);
    const now = new Date();
    const diffMs = now.getTime() - lastSync.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    
    if (diffMins < 1) return '刚刚';
    if (diffMins < 60) return `${diffMins}分钟前`;
    if (diffMins < 1440) return `${Math.floor(diffMins / 60)}小时前`;
    return `${Math.floor(diffMins / 1440)}天前`;
  };

  return (
    <div className="relative">
      {/* Sync Status Indicator */}
      <div className="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400">
        <button
          onClick={handleManualSync}
          disabled={isSyncing}
          className="flex items-center gap-1 px-2 py-1 rounded hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
          title={isSyncing ? '同步中...' : '点击同步'}
        >
          {isSyncing ? (
            <Refresh className="w-4 h-4 animate-spin" />
          ) : lastSyncError ? (
            <CloudOff className="w-4 h-4 text-red-500" />
          ) : (
            <Cloud className="w-4 h-4 text-green-500" />
          )}
          <span>{isSyncing ? '同步中' : formatLastSyncTime()}</span>
        </button>
      </div>

      {/* Error Message */}
      {lastSyncError && !isSyncing && (
        <div className="absolute top-full right-0 mt-1 p-2 bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 text-xs rounded shadow-lg z-10">
          <div className="flex items-start gap-1">
            <AlertCircle className="w-3 h-3 mt-0.5 flex-shrink-0" />
            <span>{lastSyncError}</span>
          </div>
        </div>
      )}

      {/* Conflicts Modal */}
      {conflicts.length > 0 && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl p-6 max-w-md w-full mx-4">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold">同步冲突</h3>
              <button
                onClick={() => resolveAllConflicts('remote')}
                className="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
              >
                <X className="w-5 h-5" />
              </button>
            </div>
            
            <p className="text-sm text-gray-600 dark:text-gray-400 mb-4">
              检测到 {conflicts.length} 个冲突。请选择如何解决：
            </p>

            <div className="space-y-2">
              <button
                onClick={() => resolveAllConflicts('local')}
                className="w-full px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors"
              >
                使用本地版本
              </button>
              <button
                onClick={() => resolveAllConflicts('remote')}
                className="w-full px-4 py-2 bg-gray-600 text-white rounded hover:bg-gray-700 transition-colors"
              >
                使用云端版本
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};