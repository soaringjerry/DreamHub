// src/components/MessageDisplay.tsx
import React, { useEffect, useRef } from 'react';
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
  const messages = useChatStore((state) => state.messages); // 从 store 获取消息列表
  const messagesEndRef = useRef<HTMLDivElement>(null); // 用于自动滚动到底部

  // 自动滚动到最新的消息
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]); // 依赖于消息列表的变化

  // 如果没有消息，显示欢迎信息
  if (messages.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-full text-center p-4 space-y-4 opacity-80">
        <div className="w-16 h-16 rounded-full bg-primary-100 dark:bg-primary-900/30 flex items-center justify-center">
          <Bot size={28} className="text-primary-600 dark:text-primary-400" />
        </div>
        <h3 className="text-xl font-display font-semibold text-gray-800 dark:text-gray-200">欢迎使用 DreamHub</h3>
        <p className="text-gray-500 dark:text-gray-400 max-w-sm">
          上传文档并开始提问，AI 将帮助您理解和分析文档内容。
        </p>
        <div className="mt-2 flex flex-col items-center text-sm text-gray-400 dark:text-gray-500">
          <p>您可以尝试以下问题：</p>
          <ul className="list-disc text-left mt-2 space-y-1 text-primary-600 dark:text-primary-400">
            <li>总结文档的主要内容</li>
            <li>解释文档中的某个概念</li>
            <li>提取文档中的关键信息</li>
          </ul>
        </div>
      </div>
    );
  }

  // 检测消息内容是否包含代码块
  const hasCodeBlock = (content: string): boolean => {
    return content.includes('```');
  };

  // 判断是否为文件上传成功消息
  const isFileUploadMessage = (content: string): boolean => {
    return content.includes('上传成功') && content.includes('个块');
  };

  return (
    <div className="space-y-6">
      {messages.map((msg, index) => (
        <div
          key={index}
          className={`flex items-start gap-3 ${msg.sender === 'user' ? 'justify-end' : 'justify-start'} animate-fadeIn`}
        >
          {/* Avatar for AI - only show on left side */}
          {msg.sender === 'ai' && (
            <div className="flex-shrink-0 w-10 h-10 rounded-full bg-gradient-to-br from-primary-600 to-primary-700 shadow-md flex items-center justify-center">
              {isFileUploadMessage(msg.content) ? (
                <FileText size={20} className="text-white" />
              ) : (
                <Bot size={20} className="text-white" />
              )}
            </div>
          )}

          {/* Message Bubble */}
          <div
            className={`max-w-xs md:max-w-md lg:max-w-lg px-4 py-3 rounded-2xl shadow-soft backdrop-blur-sm
                       ${msg.sender === 'user'
                         ? 'bg-gradient-to-r from-primary-500 to-primary-600 text-white rounded-tr-none' // User message style
                         : hasCodeBlock(msg.content)
                           ? 'bg-gray-800 dark:bg-gray-900 border border-gray-700 text-gray-100 rounded-tl-none' // AI code message style
                           : 'bg-white dark:bg-gray-800 border border-gray-100 dark:border-gray-700 text-gray-800 dark:text-gray-100 rounded-tl-none' // AI message style
                       }`}
          >
            {msg.sender === 'ai' && hasCodeBlock(msg.content) ? (
              <div className="prose dark:prose-invert prose-sm max-w-none">
                <ReactMarkdown
                  components={{
                    code({node, inline, className, children, ...props}: CodeProps) {
                      const match = /language-(\w+)/.exec(className || '');
                      return !inline && match ? (
                        <SyntaxHighlighter
                          style={atomDark}
                          language={match[1]}
                          PreTag="div"
                          {...props}
                        >
                          {String(children).replace(/\n$/, '')}
                        </SyntaxHighlighter>
                      ) : (
                        <code className={className} {...props}>
                          {children}
                        </code>
                      );
                    }
                  }}
                >
                  {msg.content}
                </ReactMarkdown>
              </div>
            ) : isFileUploadMessage(msg.content) && msg.sender === 'ai' ? (
              <div className="flex items-center text-green-600 dark:text-green-400 font-medium">
                <FileText size={16} className="mr-2" />
                <span>{msg.content}</span>
              </div>
            ) : (
              <div className="whitespace-pre-wrap break-words">{msg.content}</div>
            )}
          </div>

          {/* Avatar for User - only show on right side */}
          {msg.sender === 'user' && (
            <div className="flex-shrink-0 w-10 h-10 rounded-full bg-gradient-to-br from-secondary-500 to-secondary-600 shadow-md flex items-center justify-center">
              <User size={20} className="text-white" />
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