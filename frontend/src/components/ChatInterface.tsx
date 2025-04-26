// src/components/ChatInterface.tsx
import React from 'react';
import { useTranslation } from 'react-i18next';
// 导入新的选择器和 action
import {
  useChatStore,
  useActiveMessages,
  useActiveConversationStatus,
} from '../store/chatStore';
import MessageDisplay from './MessageDisplay';
import UserInput from './UserInput';
import { PlusSquare } from 'lucide-react'; // 导入新建图标

const ChatInterface: React.FC = () => {
  const { t } = useTranslation();
  // --- 使用新的选择器和 action ---
  // const messages = useActiveMessages(); // 获取活动对话的消息 (暂时移除未使用的变量)
  const { isLoading, error } = useActiveConversationStatus(); // 获取活动对话的状态
  const startNewConversation = useChatStore((state) => state.startNewConversation); // 获取新建对话 action

  // 处理新建聊天
  const handleNewChat = () => {
    // 可以选择性地添加确认，但通常新建操作不需要
    startNewConversation();
  };

  return (
    <div className="flex flex-col h-full bg-white dark:bg-gray-850 rounded-b-lg"> {/* Match panel background */}
      {/* Message Display Area: Adjusted padding, background */}
      <div className="flex-grow overflow-y-auto p-4 md:p-5 space-y-5 bg-white dark:bg-gray-850 relative"> {/* Removed gradient, adjusted padding/spacing */}
        <MessageDisplay /> {/* MessageDisplay 现在会使用 useActiveMessages */}

        {/* Top Right Controls: New Chat Button */}
        <div className="absolute top-3 right-3 flex items-center space-x-2">
          {/* New Chat Button */}
          <button
            onClick={handleNewChat}
            className="p-1.5 rounded-md bg-gray-100 hover:bg-gray-200 dark:bg-gray-700 dark:hover:bg-gray-600 text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 transition-colors duration-150 focus:outline-none focus:ring-1 focus:ring-gray-400"
            aria-label={t('newChatLabel', 'New Chat')} // 添加 i18n key
            title={t('newChatLabel', 'New Chat')} // 添加 i18n key
          >
            <PlusSquare size={14} /> {/* 使用新图标 */}
          </button>
        </div>
      </div>

      {/* Loading Indicator (uses isLoading from active conversation) */}
      {isLoading && (
        <div className="p-2.5 text-center text-xs border-t border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800">
          <div className="flex items-center justify-center space-x-1.5 text-gray-500 dark:text-gray-400">
            {/* Simple spinner */}
            <svg className="animate-spin h-3.5 w-3.5 text-primary-500" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            <span className="font-medium">{t('aiThinking')}</span>
          </div>
        </div>
      )}

      {/* Error Display (uses error from active conversation) */}
      {error && (
        <div className="p-2.5 border-t border-red-200 dark:border-red-700 bg-red-50 dark:bg-red-900/20">
          <div className="flex items-center justify-center text-xs">
            <svg className="w-4 h-4 text-red-500 dark:text-red-400 mr-1.5 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg"> {/* Adjusted size */}
              <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd"></path>
            </svg>
            <span className="font-medium text-red-600 dark:text-red-300">{error}</span>
          </div>
        </div>
      )}

      {/* User Input Area: Adjusted padding, background, border */}
      <div className="p-3 md:p-4 border-t border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 rounded-b-lg">
        <UserInput /> {/* 渲染用户输入组件 */}
      </div>
    </div>
  );
};

export default ChatInterface;