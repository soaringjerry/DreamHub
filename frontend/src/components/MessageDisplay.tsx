// src/components/MessageDisplay.tsx
import React, { useEffect, useRef } from 'react';
import { useTranslation } from 'react-i18next'; // 导入 useTranslation
import { useChatStore } from '../store/chatStore'; // 导入 store 来获取消息
import { User, Bot, FileText } from 'lucide-react'; // 添加文件图标
import ReactMarkdown from 'react-markdown'; // 添加Markdown支持
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'; // 添加代码高亮
import { atomDark } from 'react-syntax-highlighter/dist/esm/styles/prism'; // 代码高亮样式

// (类型定义已移除，因为它未被使用，可能由 chatStore 提供)

// 定义代码组件类型
interface CodeProps {
  node?: any;
  inline?: boolean;
  className?: string;
  children?: React.ReactNode;
  [key: string]: any;
}

const MessageDisplay: React.FC = () => {
  const { t } = useTranslation(); // 初始化 useTranslation
  const messages = useChatStore((state) => state.messages); // 从 store 获取消息列表
  const messagesEndRef = useRef<HTMLDivElement>(null); // 用于自动滚动到底部

  // 自动滚动到最新的消息
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]); // 依赖于消息列表的变化

  // 如果没有消息，显示欢迎信息: Refined styling
  if (messages.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-full text-center p-6 space-y-3 text-gray-500 dark:text-gray-400">
        {/* Icon */}
        <div className="w-14 h-14 rounded-full bg-primary-100 dark:bg-gray-700 flex items-center justify-center mb-2">
          <Bot size={26} className="text-primary-600 dark:text-primary-400" />
        </div>
        {/* Title */}
        <h3 className="text-lg font-semibold text-gray-700 dark:text-gray-200">{t('welcomeTitle')}</h3>
        {/* Description */}
        <p className="text-sm max-w-xs">
          {t('welcomeMessage')}
        </p>
        {/* Example Questions (Optional) */}
        {/*
        <div className="mt-3 pt-3 border-t border-gray-200 dark:border-gray-700 w-full max-w-xs">
          <p className="text-xs text-gray-400 dark:text-gray-500 mb-1.5">尝试提问:</p>
          <ul className="text-xs space-y-1 text-primary-600 dark:text-primary-400">
            <li>总结文档的主要观点</li>
            <li>解释 [某个概念]</li>
            <li>提取关键日期和事件</li>
          </ul>
        </div>
        */}
      </div>
    );
  }

  // 检测消息内容是否包含代码块
  const hasCodeBlock = (content: string): boolean => {
    // More robust check for code blocks
    return /```[\s\S]*?```/.test(content);
  };

  // 判断是否为文件上传成功消息
  const isFileUploadMessage = (content: string): boolean => {
    // Check against translated string
    return content.startsWith(t('fileUploadSuccessPrefix'));
  };

  return (
    <div className="space-y-4"> {/* Adjusted spacing */}
      {messages.map((msg, index) => (
        <div
          key={index}
          className={`flex items-start gap-2.5 ${msg.sender === 'user' ? 'justify-end' : 'justify-start'} animate-fadeIn`} // Adjusted gap
        >
          {/* Avatar for AI */}
          {msg.sender === 'ai' && (
            <div className={`flex-shrink-0 w-8 h-8 rounded-full flex items-center justify-center shadow-sm ${
              isFileUploadMessage(msg.content)
                ? 'bg-green-100 dark:bg-green-800/50' // File upload icon background
                : 'bg-primary-100 dark:bg-gray-700' // AI icon background
            }`}>
              {isFileUploadMessage(msg.content) ? (
                <FileText size={16} className="text-green-600 dark:text-green-400" />
              ) : (
                <Bot size={16} className="text-primary-600 dark:text-primary-400" />
              )}
            </div>
          )}

          {/* Message Bubble: Refined styles */}
          <div
            className={`max-w-sm md:max-w-md lg:max-w-lg px-3.5 py-2.5 rounded-lg shadow-xs text-sm leading-relaxed break-words
                       ${msg.sender === 'user'
                         ? 'bg-primary-500 dark:bg-primary-600 text-white rounded-tr-md' // User message style (softer corners)
                         : isFileUploadMessage(msg.content)
                           ? 'bg-green-50 dark:bg-gray-750 border border-green-100 dark:border-gray-600 text-green-700 dark:text-green-300 rounded-tl-md font-medium' // File upload message style
                           : hasCodeBlock(msg.content)
                             ? 'bg-gray-800 dark:bg-gray-900 border border-gray-700 dark:border-gray-700 text-gray-100 rounded-tl-md p-0 overflow-hidden' // AI code message style (remove padding for highlighter)
                             : 'bg-gray-100 dark:bg-gray-750 border border-gray-200 dark:border-gray-600 text-gray-800 dark:text-gray-100 rounded-tl-md' // AI default message style
                       }`}
          >
            {/* Code Block Rendering */}
            {msg.sender === 'ai' && hasCodeBlock(msg.content) ? (
              <ReactMarkdown
                // Removed className from here to fix TS error
                components={{
                  code({ node, inline, className, children, ...props }: CodeProps) {
                    const match = /language-(\w+)/.exec(className || '');
                    return !inline && match ? (
                      <SyntaxHighlighter
                        style={atomDark} // Or choose another theme like oneDark, materialDark, etc.
                        language={match[1]}
                        PreTag="div"
                        className="!bg-transparent !p-3" // Override background and padding
                        {...props}
                      >
                        {String(children).replace(/\n$/, '')}
                      </SyntaxHighlighter>
                    ) : (
                      // Inline code style
                      <code className={`bg-gray-200 dark:bg-gray-600 px-1 py-0.5 rounded text-xs ${className}`} {...props}>
                        {children}
                      </code>
                    );
                  },
                  // Optional: Style other markdown elements like p, ul, etc. if needed inside code messages
                  p: ({node, ...props}) => <p className="mb-2 last:mb-0" {...props} />,
                }}
              >
                {msg.content}
              </ReactMarkdown>
            ) : isFileUploadMessage(msg.content) && msg.sender === 'ai' ? (
              // File Upload Message Content (already styled by bubble)
              // Display the translated part if it matches, otherwise the original content
              <span>{isFileUploadMessage(msg.content) ? t('fileUploadSuccessPrefix') + msg.content.substring(t('fileUploadSuccessPrefix').length) : msg.content}</span>
            ) : (
              // Default Text Message Content
              <div className="whitespace-pre-wrap">{msg.content}</div>
            )}
          </div>

          {/* Avatar for User */}
          {msg.sender === 'user' && (
            <div className="flex-shrink-0 w-8 h-8 rounded-full bg-secondary-500 dark:bg-secondary-600 shadow-sm flex items-center justify-center">
              <User size={16} className="text-white" />
            </div>
          )}
        </div>
      ))}

      {/* Empty div for scroll positioning */}
      <div ref={messagesEndRef} />
    </div>
  );
};

export default MessageDisplay;