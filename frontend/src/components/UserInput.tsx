// src/components/UserInput.tsx
import React, { useState } from 'react';
import { useChatStore } from '../store/chatStore';

const UserInput: React.FC = () => {
  const [message, setMessage] = useState('');

  // --- Zustand Store Integration ---
  // 使用单独的选择器获取 action 和状态
  const sendMessage = useChatStore((state) => state.sendMessage);
  const isLoading = useChatStore((state) => state.isLoading);

  const handleInputChange = (event: React.ChangeEvent<HTMLTextAreaElement>) => {
    setMessage(event.target.value);
  };

  const handleSend = () => {
    if (message.trim() && !isLoading) {
      sendMessage(message.trim());
      setMessage(''); // 清空输入框
    }
  };

  // 处理 Enter 键发送 (Shift+Enter 换行)
  const handleKeyDown = (event: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (event.key === 'Enter' && !event.shiftKey && !isLoading) {
      event.preventDefault(); // 阻止默认的换行行为
      handleSend();
    }
  };

  return (
    <div className="flex items-center space-x-2">
      <textarea
        value={message}
        onChange={handleInputChange}
        onKeyDown={handleKeyDown}
        placeholder="输入消息... (Shift+Enter 换行)"
        rows={1} // 初始行数，可以根据内容自适应高度
        className="flex-grow p-2 border border-gray-300 rounded-md resize-none focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-100 dark:focus:ring-offset-gray-800"
        style={{ maxHeight: '100px', overflowY: 'auto' }} // 限制最大高度并允许滚动
        disabled={isLoading}
      />
      <button
        onClick={handleSend}
        disabled={isLoading || !message.trim()}
        className={`px-4 py-2 rounded-md text-white transition-colors duration-200 ease-in-out
                   ${isLoading ? 'bg-gray-400 cursor-not-allowed' : 'bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 dark:focus:ring-offset-gray-800'}
                   disabled:opacity-50 disabled:cursor-not-allowed`}
      >
        发送
      </button>
    </div>
  );
};

export default UserInput;