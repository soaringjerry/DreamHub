// src/components/UserInput.tsx
import React, { useState, useRef, useEffect } from 'react';
import { useTranslation } from 'react-i18next'; // 导入 useTranslation
import { useChatStore } from '../store/chatStore';
import { Send, Sparkles, Lightbulb } from 'lucide-react'; // 导入图标

// 预定义的快速提示键 (将在组件内部使用 t 函数获取实际文本)
const QUICK_PROMPT_KEYS = [
  "quickPromptSummarize",
  "quickPromptExtract",
  "quickPromptExplain",
  "quickPromptFindData",
  "quickPromptAnalyze",
];

const UserInput: React.FC = () => {
  const { t } = useTranslation(); // 初始化 useTranslation
  const [message, setMessage] = useState('');
  const [isFocused, setIsFocused] = useState(false);
  const [showPrompts, setShowPrompts] = useState(false);
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const promptsRef = useRef<HTMLDivElement>(null);

  // --- Zustand Store Integration ---
  // 使用单独的选择器获取 action 和状态
  const sendMessage = useChatStore((state) => state.sendMessage);
  const isLoading = useChatStore((state) => state.isLoading);
  const uploadedFiles = useChatStore((state) => state.uploadedFiles);

  // 点击外部关闭提示框
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (promptsRef.current && !promptsRef.current.contains(event.target as Node)) {
        setShowPrompts(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, []);

  const handleInputChange = (event: React.ChangeEvent<HTMLTextAreaElement>) => {
    setMessage(event.target.value);
  };

  // 自动调整 textarea 高度
  useEffect(() => {
    const textarea = textareaRef.current;
    if (textarea) {
      textarea.style.height = 'auto'; // Reset height
      const scrollHeight = textarea.scrollHeight;
      // 设置一个最大高度，例如 120px (约 5 行)
      const maxHeight = 120;
      textarea.style.height = `${Math.min(scrollHeight, maxHeight)}px`;
      // 如果内容超过最大高度，显示滚动条
      textarea.style.overflowY = scrollHeight > maxHeight ? 'auto' : 'hidden';
    }
  }, [message]);


  const handleSend = () => {
    if (message.trim() && !isLoading) {
      sendMessage(message.trim());
      setMessage(''); // 清空输入框
      
      // 自动聚焦回输入框
      setTimeout(() => {
        textareaRef.current?.focus();
      }, 10);
    }
  };

  // 处理 Enter 键发送 (Shift+Enter 换行)
  const handleKeyDown = (event: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (event.key === 'Enter' && !event.shiftKey && !isLoading) {
      event.preventDefault(); // 阻止默认的换行行为
      handleSend();
    }
  };

  // 处理快速提示选择
  const handlePromptSelect = (prompt: string) => {
    setMessage(prompt);
    setShowPrompts(false);
    textareaRef.current?.focus();
  };

  // 切换提示面板
  const togglePrompts = () => {
    setShowPrompts(prev => !prev);
  };

  return (
    <div className="relative">
      <div className={`flex items-center bg-gray-50 dark:bg-gray-700 rounded-xl transition-all duration-200 border-2 
        ${isFocused 
          ? 'border-primary-400 dark:border-primary-500 shadow-md' 
          : 'border-gray-200 dark:border-gray-600'}`}>
        
        {/* 辅助图标 - 快捷提示按钮 */}
        <button
          type="button"
          onClick={togglePrompts}
          className={`ml-3 p-1.5 rounded-full transition-all duration-300 ${
            showPrompts 
              ? 'bg-primary-100 dark:bg-primary-800/40 text-primary-600 dark:text-primary-400' 
              : 'text-gray-400 dark:text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-600'
            }`}
            aria-label={t('showQuickPromptsLabel')}
            title={t('quickPromptsTitle')}
        >
          <Lightbulb size={18} className={showPrompts ? 'text-primary-500 animate-pulse' : ''} />
        </button>
        
        {/* 文本输入区域 */}
        <textarea
          value={message}
          ref={textareaRef}
          onChange={handleInputChange}
          onKeyDown={handleKeyDown}
          onFocus={() => setIsFocused(true)}
          onBlur={() => setIsFocused(false)}
          placeholder={uploadedFiles?.length > 0
            ? t('inputPlaceholderReady')
            : t('inputPlaceholderUpload')}
          rows={1} // 初始为 1 行，将根据内容自动调整
          className="flex-grow p-3 bg-transparent border-0 rounded-lg resize-none focus:outline-none
                     text-gray-800 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500"
          style={{ minHeight: '48px', maxHeight: '120px', overflowY: 'hidden' }}
          disabled={isLoading}
        />
        
        {/* 辅助图标 */}
        <span className="mr-2 text-gray-400 dark:text-gray-500">
          <Sparkles size={18} className={`transition-all duration-300 ${isFocused ? 'text-primary-500 rotate-12' : ''}`} />
        </span>
        
        {/* 发送按钮 */}
        <button
          onClick={handleSend}
          disabled={isLoading || !message.trim()}
          className={`flex items-center justify-center m-1 p-2.5 rounded-lg text-white transition-all duration-200 ease-in-out
                     ${isLoading 
                       ? 'bg-gray-400 cursor-not-allowed' 
                       : message.trim() 
                         ? 'bg-primary-600 hover:bg-primary-700 active:scale-95' 
                         : 'bg-gray-300 dark:bg-gray-600 cursor-not-allowed'}
                     disabled:opacity-50`}
          aria-label={t('sendMessageLabel')}
        >
          {isLoading ? (
            <svg className="animate-spin h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
          ) : (
            <Send size={20} className="transition-transform group-hover:translate-x-1" />
          )}
        </button>
      </div>
      
      {/* 快捷提示下拉面板 */}
      {showPrompts && (
        <div 
          ref={promptsRef}
          className="absolute left-0 bottom-full mb-2 w-full bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 py-2 animate-fadeIn z-10"
        >
          <h3 className="px-4 py-1 text-xs font-semibold text-gray-500 dark:text-gray-400 border-b border-gray-100 dark:border-gray-700">
            {t('quickPromptsTitle')}
          </h3>
          <div className="max-h-48 overflow-y-auto p-1">
            {QUICK_PROMPT_KEYS.map((promptKey, index) => (
              <button
                key={index}
                onClick={() => handlePromptSelect(t(promptKey))} // 使用 t 函数获取文本
                className="w-full text-left px-3 py-2 text-sm rounded-md hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors duration-150"
              >
                {t(promptKey)} {/* 显示翻译后的文本 */}
              </button>
            ))}
          </div>
        </div>
      )}
      
      {/* 底部提示信息 */}
      <div className="text-xs text-center mt-2 text-gray-400 dark:text-gray-500">
        {uploadedFiles?.length > 0
          ? t('filesUploadedHint', { count: uploadedFiles.length })
          : t('uploadFirstHint')
        }
      </div>
    </div>
  );
};

export default UserInput;