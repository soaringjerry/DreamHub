// src/components/ChatInterface.tsx
import React from 'react';
import { useChatStore } from '../store/chatStore';
import MessageDisplay from './MessageDisplay'; // 导入消息显示组件
import UserInput from './UserInput';       // 导入用户输入组件

const ChatInterface: React.FC = () => {
  // --- Zustand Store Integration ---
  // 使用单独的选择器获取状态
  const isLoading = useChatStore((state) => state.isLoading);
  const error = useChatStore((state) => state.error);

  return (
    <div className="flex flex-col h-full"> {/* 让聊天界面填满其容器高度 */}
      {/* Message Display Area */}
      <div className="flex-grow overflow-y-auto p-4 space-y-4 bg-gray-50 dark:bg-gray-700 rounded-t-lg">
        <MessageDisplay /> {/* 渲染消息显示组件 */}
      </div>

      {/* Loading Indicator */}
      {isLoading && (
        <div className="p-2 text-center text-sm text-gray-500 dark:text-gray-400 border-t border-gray-200 dark:border-gray-600">
          AI 正在思考...
        </div>
      )}

      {/* Error Display */}
      {error && (
        <div className="p-2 text-center text-sm text-red-600 bg-red-100 dark:bg-red-900/50 dark:text-red-300 border-t border-red-200 dark:border-red-700">
          错误: {error}
        </div>
      )}

      {/* User Input Area */}
      {/* 确保即使在加载或错误时，底部圆角也能正确应用 */}
      <div className={`p-4 border-t border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-800 rounded-b-lg`}>
        <UserInput /> {/* 渲染用户输入组件 */}
      </div>
    </div>
  );
};

export default ChatInterface;