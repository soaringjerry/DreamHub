import React from 'react';
import { useTranslation } from 'react-i18next';
import {
  useChatStore,
  useActiveConversationId,
  useSortedConversations, // Import the new selector
  useIsConversationListLoading, // Import loading state selector
  useConversationListError, // Import error state selector
} from '../store/chatStore';
import { MessageSquare, PlusSquare, Trash2, Loader2, AlertCircle } from 'lucide-react'; // Import icons

const ConversationList: React.FC = () => {
  const { t } = useTranslation();
  // Use the new selectors
  const sortedConversations = useSortedConversations();
  const isLoading = useIsConversationListLoading();
  const error = useConversationListError();
  const activeConversationId = useActiveConversationId();
  // Get actions
  const switchConversation = useChatStore((state) => state.switchConversation);
  const startNewConversation = useChatStore((state) => state.startNewConversation);
  const deleteConversation = useChatStore((state) => state.deleteConversation);
  const fetchConversations = useChatStore((state) => state.fetchConversations); // Get fetch action for retry

  const handleNewConversation = () => {
    startNewConversation();
  };

  const handleDelete = (e: React.MouseEvent, conversationId: string) => {
    e.stopPropagation(); // Prevent switching conversation when clicking delete
    if (window.confirm(t('deleteConversationConfirmation', 'Are you sure you want to delete this conversation?'))) {
      deleteConversation(conversationId);
    }
  };

  return (
    <div className="flex flex-col h-full bg-gray-50 dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700">
      {/* Header with New Chat Button */}
      <div className="p-3 border-b border-gray-200 dark:border-gray-700 flex justify-between items-center">
        <h2 className="text-sm font-semibold text-gray-700 dark:text-gray-200">{t('conversationsTitle', 'Conversations')}</h2>
        <button
          onClick={handleNewConversation}
          className="p-1.5 rounded-md text-gray-500 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 focus:outline-none focus:ring-1 focus:ring-inset focus:ring-primary-500"
          aria-label={t('newChatLabel', 'New Chat')}
          title={t('newChatLabel', 'New Chat')}
        >
          <PlusSquare size={16} />
        </button>
      </div>

      {/* Conversation List Area */}
      <div className="flex-grow overflow-y-auto p-2">
        {isLoading ? (
          <div className="flex items-center justify-center h-full">
            <Loader2 className="h-5 w-5 text-gray-400 animate-spin" />
            <span className="ml-2 text-xs text-gray-500">{t('loadingConversations', 'Loading...')}</span>
          </div>
        ) : error ? (
          <div className="p-3 text-center text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900/30 rounded border border-red-200 dark:border-red-700/50">
            <AlertCircle className="h-4 w-4 inline-block mr-1 mb-0.5" />
            <p className="text-xs font-medium mb-1">{t('errorLoadingConversationsTitle', 'Error Loading Conversations')}</p>
            <p className="text-xs mb-2">{error}</p>
            <button
              onClick={() => fetchConversations()} // Retry button
              className="text-xs px-2 py-0.5 bg-red-100 dark:bg-red-800/50 text-red-700 dark:text-red-300 rounded hover:bg-red-200 dark:hover:bg-red-700/50"
            >
              {t('retry', 'Retry')}
            </button>
          </div>
        ) : sortedConversations.length === 0 ? (
          <p className="pt-4 text-xs text-center text-gray-500 dark:text-gray-400">{t('noConversations', 'No conversations yet.')}</p>
        ) : (
          <ul className="space-y-1">
            {sortedConversations.map((conv) => (
              <li key={conv.id}>
                <button
                  onClick={() => switchConversation(conv.id)}
                  className={`w-full flex items-center justify-between p-2 rounded-md text-sm text-left transition-colors duration-150 group ${
                    activeConversationId === conv.id
                      ? 'bg-primary-100 dark:bg-primary-800/50 text-primary-700 dark:text-primary-300 font-medium'
                      : 'text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'
                  }`}
                >
                  <span className="flex items-center space-x-2 truncate">
                    <MessageSquare size={14} className="flex-shrink-0" />
                    <span className="truncate" title={conv.title}>{conv.title}</span>
                  </span>
                  {/* Delete Button (appears on hover) */}
                  <button
                    onClick={(e) => handleDelete(e, conv.id)}
                    className="p-1 rounded text-gray-400 dark:text-gray-500 opacity-0 group-hover:opacity-100 hover:text-red-500 dark:hover:text-red-400 focus:opacity-100 focus:text-red-500 transition-opacity duration-150"
                    aria-label={t('deleteConversationLabel', 'Delete conversation')}
                    title={t('deleteConversationLabel', 'Delete conversation')}
                  >
                    <Trash2 size={14} />
                  </button>
                </button>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
};

export default ConversationList;