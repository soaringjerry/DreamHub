// src/components/ChatInterface.tsx
import React from 'react';
import { useTranslation } from 'react-i18next'; // 导入 useTranslation
import { useChatStore } from '../store/chatStore';
import MessageDisplay from './MessageDisplay'; // 导入消息显示组件
import UserInput from './UserInput';       // 导入用户输入组件
import { RotateCcw } from 'lucide-react'; // 导入重置图标

const ChatInterface: React.FC = () => {
  const { t } = useTranslation(); // 初始化 useTranslation
  // --- Zustand Store Integration ---
  // 使用单独的选择器获取状态和操作
  const isLoading = useChatStore((state) => state.isLoading);
  const error = useChatStore((state) => state.error);
  const messages = useChatStore((state) => state.messages);
  const resetChat = useChatStore((state) => state.resetChat);
  const conversationId = useChatStore((state) => state.conversationId); // 获取 conversationId

  // 处理重置聊天
  const handleResetChat = () => {
    if (window.confirm(t('resetChatConfirmation'))) {
      resetChat();
    }
  };

  return (
    <div className="flex flex-col h-full bg-white dark:bg-gray-850 rounded-b-lg"> {/* Match panel background */}
      {/* Message Display Area: Adjusted padding, background */}
      <div className="flex-grow overflow-y-auto p-4 md:p-5 space-y-5 bg-white dark:bg-gray-850 relative"> {/* Removed gradient, adjusted padding/spacing */}
        <MessageDisplay /> {/* 渲染消息显示组件 */}

        {/* Top Right Controls: Reset Button */}
        <div className="absolute top-3 right-3 flex items-center space-x-2">
          {/* Reset Button */}
          {messages.length > 0 && (
            <button
              onClick={handleResetChat}
              className="p-1.5 rounded-md bg-gray-100 hover:bg-gray-200 dark:bg-gray-700 dark:hover:bg-gray-600 text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 transition-colors duration-150 focus:outline-none focus:ring-1 focus:ring-gray-400"
              aria-label={t('resetChatLabel')}
              title={t('resetChatLabel')}
            >
              <RotateCcw size={14} /> {/* Adjusted size */}
            </button>
          )}
        </div>
      </div>

      {/* Loading Indicator: Refined style */}
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

      {/* Error Display: Refined style */}
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