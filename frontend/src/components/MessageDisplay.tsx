// src/components/MessageDisplay.tsx
import React, { useEffect, useRef } from 'react';
import { useChatStore } from '../store/chatStore'; // 导入 store 来获取消息

// 定义 Message 类型 (可以考虑提取到共享的 types 文件)
interface Message {
  sender: 'user' | 'ai';
  content: string;
}

const MessageDisplay: React.FC = () => {
  const messages = useChatStore((state) => state.messages); // 从 store 获取消息列表
  const messagesEndRef = useRef<HTMLDivElement>(null); // 用于自动滚动到底部

  // 自动滚动到最新的消息
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]); // 依赖于消息列表的变化

  return (
    <div className="space-y-4">
      {messages.map((msg, index) => (
        <div
          key={index}
          className={`flex ${msg.sender === 'user' ? 'justify-end' : 'justify-start'}`}
        >
          <div
            className={`max-w-xs md:max-w-md lg:max-w-lg px-4 py-2 rounded-lg shadow
                       ${msg.sender === 'user'
                         ? 'bg-blue-500 text-white'
                         : 'bg-gray-200 text-gray-800 dark:bg-gray-600 dark:text-gray-100'
                       }`}
          >
            {/* 简单的文本渲染，后续可以支持 Markdown 或代码块 */}
            <p className="whitespace-pre-wrap">{msg.content}</p>
          </div>
        </div>
      ))}
      {/* 空 div 用于滚动定位 */}
      <div ref={messagesEndRef} />
    </div>
  );
};

export default MessageDisplay;