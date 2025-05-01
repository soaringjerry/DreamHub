// src/components/MessageDisplay.tsx
import React, { useEffect, useRef } from 'react';
import { useTranslation } from 'react-i18next';
// 导入新的选择器和 Message 类型
import { useActiveMessages, Message } from '../store/chatStore'; // Import Message type
import { User, Bot, FileText } from 'lucide-react';
import ReactMarkdown from 'react-markdown';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { atomDark } from 'react-syntax-highlighter/dist/esm/styles/prism';

// 定义代码组件类型
interface CodeProps {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  node?: any; // TODO: Define more specific type based on Markdown AST node
  inline?: boolean;
  className?: string;
  children?: React.ReactNode;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  [key: string]: any; // Allow other props passed by react-markdown
}

const MessageDisplay: React.FC = () => {
  const { t } = useTranslation();
  // 使用新的选择器获取活动对话的消息
  const messages = useActiveMessages();
  const messagesEndRef = useRef<HTMLDivElement>(null); // 用于自动滚动到底部

  // 自动滚动到最新的消息
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]); // 依赖于消息列表的变化

  // 如果没有消息，显示欢迎信息
  if (!messages || messages.length === 0) {
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
      </div>
    );
  }

  // 检测消息内容是否包含代码块
  const hasCodeBlock = (content: string): boolean => {
    return /```[\s\S]*?```/.test(content);
  };

  // 判断是否为文件上传成功/失败消息 (可以扩展)
  const isStatusMessage = (content: string): boolean => {
    // 简单检查是否包含特定关键词，可以根据需要改进
    return content.includes(t('fileUploadSuccessPrefix')) || content.includes('文件上传失败');
  };

  return (
    <div className="space-y-4"> {/* Adjusted spacing */}
      {/* Ensure messages is an array and use the correct Message type */}
      {Array.isArray(messages) && messages.map((msg: Message) => ( // Use Message type
        <div
          // Use message ID as key, it should be unique from backend
          key={msg.id}
          className={`flex items-start gap-2.5 ${msg.role === 'user' ? 'justify-end' : 'justify-start'} animate-fadeIn`} // Use role
        >
          {/* Avatar for AI */}
          {msg.role === 'assistant' && ( // Use role
            <div className={`flex-shrink-0 w-8 h-8 rounded-full flex items-center justify-center shadow-sm ${
              isStatusMessage(msg.content)
                ? 'bg-green-100 dark:bg-green-800/50'
                : 'bg-primary-100 dark:bg-gray-700'
            }`}>
              {isStatusMessage(msg.content) ? (
                <FileText size={16} className="text-green-600 dark:text-green-400" />
              ) : (
                <Bot size={16} className="text-primary-600 dark:text-primary-400" />
              )}
            </div>
          )}

          {/* Message Bubble */}
          <div
            className={`max-w-sm md:max-w-md lg:max-w-lg px-3.5 py-2.5 rounded-lg shadow-xs text-sm leading-relaxed break-words
                       ${msg.role === 'user' // Use role
                         ? 'bg-primary-500 dark:bg-primary-600 text-white rounded-tr-md'
                         : isStatusMessage(msg.content)
                           ? 'bg-green-50 dark:bg-gray-750 border border-green-100 dark:border-gray-600 text-green-700 dark:text-green-300 rounded-tl-md font-medium'
                           : hasCodeBlock(msg.content)
                             ? 'bg-gray-800 dark:bg-gray-900 border border-gray-700 dark:border-gray-700 text-gray-100 rounded-tl-md p-0 overflow-hidden'
                             : 'bg-gray-100 dark:bg-gray-750 border border-gray-200 dark:border-gray-600 text-gray-800 dark:text-gray-100 rounded-tl-md'
                       }`}
          >
            {/* Code Block Rendering */}
            {msg.role === 'assistant' && hasCodeBlock(msg.content) ? ( // Use role
              <ReactMarkdown
                components={{
                  code({ node, inline, className, children, ...props }: CodeProps) {
                    const match = /language-(\w+)/.exec(className || '');
                    return !inline && match ? (
                      <SyntaxHighlighter
                        style={atomDark}
                        language={match[1]}
                        PreTag="div"
                        className="!bg-transparent !p-3"
                        {...props}
                      >
                        {String(children).replace(/\n$/, '')}
                      </SyntaxHighlighter>
                    ) : (
                      <code className={`bg-gray-200 dark:bg-gray-600 px-1 py-0.5 rounded text-xs ${className}`} {...props}>
                        {children}
                      </code>
                    );
                  },
                  p: ({node, ...props}) => <p className="mb-2 last:mb-0" {...props} />,
                }}
              >
                {msg.content}
              </ReactMarkdown>
            ) : isStatusMessage(msg.content) && msg.role === 'assistant' ? ( // Use role
              // Status Message Content
              <span>{msg.content}</span>
            ) : (
              // Default Text Message Content
              <div className="whitespace-pre-wrap">{msg.content}</div>
            )}
          </div>

          {/* Avatar for User */}
          {msg.role === 'user' && ( // Use role
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