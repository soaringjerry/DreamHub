// src/components/ChatInterface.tsx
import React from 'react';
import { useChatStore } from '../store/chatStore';
import MessageDisplay from './MessageDisplay'; // 导入消息显示组件
import UserInput from './UserInput';       // 导入用户输入组件
import { RotateCcw } from 'lucide-react'; // 导入重置图标

const ChatInterface: React.FC = () => {
  // --- Zustand Store Integration ---
  // 使用单独的选择器获取状态和操作
  const isLoading = useChatStore((state) => state.isLoading);
  const error = useChatStore((state) => state.error);
  const messages = useChatStore((state) => state.messages);
  const resetChat = useChatStore((state) => state.resetChat);

  // 处理重置聊天
  const handleResetChat = () => {
    if (window.confirm('确定要清空当前对话吗？此操作不可撤销。')) {
      resetChat();
    }
  };

  return (
    <div className="flex flex-col h-full"> {/* 让聊天界面填满其容器高度 */}
      {/* Message Display Area with Reset Button when messages exist */}
      <div className="flex-grow overflow-y-auto p-4 md:p-6 space-y-6 bg-gradient-to-b from-gray-50 to-white dark:from-gray-700 dark:to-gray-750 rounded-t-lg relative">
        <MessageDisplay /> {/* 渲染消息显示组件 */}
        
        {/* 重置按钮 - 仅当有消息时显示 */}
        {messages.length > 0 && (
          <button 
            onClick={handleResetChat}
            className="absolute top-4 right-4 p-2 rounded-full bg-gray-200 hover:bg-gray-300 dark:bg-gray-600 dark:hover:bg-gray-500 text-gray-700 dark:text-gray-300 transition-colors duration-200"
            aria-label="重置对话"
            title="重置对话"
          >
            <RotateCcw size={16} />
          </button>
        )}
      </div>

      {/* Loading Indicator */}
      {isLoading && (
        <div className="p-3 text-center text-sm text-primary-700 dark:text-primary-300 border-t border-gray-200 dark:border-gray-600 bg-gray-50 dark:bg-gray-700 backdrop-blur-sm">
          <div className="flex items-center justify-center space-x-2">
            <div className="animate-pulse flex space-x-1">
              <div className="h-2 w-2 bg-primary-500 rounded-full"></div>
              <div className="h-2 w-2 bg-primary-500 rounded-full animation-delay-150"></div>
              <div className="h-2 w-2 bg-primary-500 rounded-full animation-delay-300"></div>
            </div>
            <span className="font-medium">AI 正在思考中...</span>
          </div>
        </div>
      )}

      {/* Error Display */}
      {error && (
        <div className="p-3 text-center bg-red-50 dark:bg-red-900/30 border-t border-red-200 dark:border-red-700">
          <div className="flex items-center justify-center">
            <svg className="w-5 h-5 text-red-500 mr-2" fill="currentColor" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg">
              <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd"></path>
            </svg>
            <span className="text-sm font-medium text-red-600 dark:text-red-300">{error}</span>
          </div>
        </div>
      )}

      {/* User Input Area */}
      <div className="p-4 border-t border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-800 rounded-b-lg shadow-inner-lg">
        <UserInput /> {/* 渲染用户输入组件 */}
      </div>
    </div>
  );
};

export default ChatInterface;