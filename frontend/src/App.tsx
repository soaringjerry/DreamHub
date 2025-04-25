import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import i18n from './i18n';
import { Moon, Sun, Languages, User, Save } from 'lucide-react';
import { useChatStore } from './store/chatStore';

import FileUpload from './components/FileUpload';
import ChatInterface from './components/ChatInterface';
import ConversationList from './components/ConversationList'; // 导入对话列表组件

function App() {
  const { t } = useTranslation();
  const [darkMode, setDarkMode] = useState(false);
  const [currentLanguage, setCurrentLanguage] = useState(i18n.language);

  const currentUserId = useChatStore((state) => state.userId);
  const setUserId = useChatStore((state) => state.setUserId);
  const [userIdInput, setUserIdInput] = useState('');

  // Load userId from localStorage
  useEffect(() => {
    const storedUserId = localStorage.getItem('userId');
    if (storedUserId && storedUserId !== currentUserId) {
      setUserId(storedUserId);
    } else if (!storedUserId && currentUserId !== null) {
      setUserId(null);
    }
    setUserIdInput(storedUserId || '');
  }, [setUserId, currentUserId]);

  // Handle setting User ID
  const handleSetUserId = () => {
    const trimmedUserId = userIdInput.trim();
    if (trimmedUserId) {
      if (trimmedUserId !== currentUserId) {
        localStorage.setItem('userId', trimmedUserId);
        setUserId(trimmedUserId);
      }
    } else {
      console.warn("User ID input is empty.");
    }
  };

  // Theme initialization
  useEffect(() => {
    const savedTheme = localStorage.getItem('theme');
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
    if (savedTheme === 'dark' || (!savedTheme && prefersDark)) {
      setDarkMode(true);
      document.documentElement.classList.add('dark');
    } else {
      setDarkMode(false);
      document.documentElement.classList.remove('dark');
    }
  }, []);

  // Toggle theme
  const toggleDarkMode = () => {
    const newDarkMode = !darkMode;
    setDarkMode(newDarkMode);
    if (newDarkMode) {
      document.documentElement.classList.add('dark');
      localStorage.setItem('theme', 'dark');
    } else {
      document.documentElement.classList.remove('dark');
      localStorage.setItem('theme', 'light');
    }
  };

  // Change language
  const changeLanguage = (lng: string) => {
    i18n.changeLanguage(lng);
    setCurrentLanguage(lng);
  };

  return (
    <div className="flex flex-col h-screen bg-gray-100 dark:bg-gray-900 text-gray-900 dark:text-gray-100 font-sans overflow-hidden"> {/* Prevent body scroll */}
      {/* Header */}
      <header className="bg-white dark:bg-gray-850 text-gray-800 dark:text-white p-3 border-b border-gray-200 dark:border-gray-700 shadow-sm relative z-10 flex-shrink-0"> {/* Adjusted padding */}
        <div className="container mx-auto flex items-center justify-between px-4"> {/* Added padding */}
          <h1 className="text-lg font-semibold tracking-tight text-primary-600 dark:text-primary-400">{t('appTitle')}</h1>
          <div className="flex items-center space-x-2">
            {/* Online status */}
            <span className="relative flex h-2.5 w-2.5">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
              <span className="relative inline-flex rounded-full h-2.5 w-2.5 bg-green-500"></span>
            </span>
            <span className="text-xs font-medium text-gray-600 dark:text-gray-400">{t('onlineStatus')}</span>

            {/* User ID Display and Input */}
            <div className="flex items-center space-x-2 ml-3 border-l border-gray-300 dark:border-gray-600 pl-3"> {/* Adjusted margin/padding */}
              <User size={14} className="text-gray-500 dark:text-gray-400 flex-shrink-0" />
              {currentUserId ? (
                <span className="text-xs font-medium text-gray-600 dark:text-gray-400 truncate" title={currentUserId}>
                  ID: {currentUserId}
                </span>
              ) : (
                <span className="text-xs text-red-500 flex-shrink-0">{t('userIdNotSet', 'ID 未设置')}</span>
              )}
              <input
                type="text"
                value={userIdInput}
                onChange={(e) => setUserIdInput(e.target.value)}
                placeholder={t('setUserIdPlaceholder', '设置用户ID')}
                className="px-2 py-0.5 border border-gray-300 dark:border-gray-600 rounded-md text-xs bg-white dark:bg-gray-700 text-gray-700 dark:text-gray-200 focus:outline-none focus:ring-1 focus:ring-primary-500 focus:border-primary-500 w-24"
                aria-label={t('setUserIdPlaceholder', '设置用户ID')}
              />
              <button
                onClick={handleSetUserId}
                className="p-1 rounded-md text-gray-500 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 focus:outline-none focus:ring-1 focus:ring-inset focus:ring-primary-500 flex-shrink-0"
                aria-label={t('saveUserIdButton', '保存用户ID')}
              >
                <Save size={14} />
              </button>
            </div>

            {/* Language Selector */}
            <div className="relative ml-3"> {/* Adjusted margin */}
              <select
                value={currentLanguage}
                onChange={(e) => changeLanguage(e.target.value)}
                className="appearance-none bg-transparent border border-gray-300 dark:border-gray-600 rounded-md py-1 pl-2 pr-7 text-xs text-gray-600 dark:text-gray-300 focus:outline-none focus:ring-1 focus:ring-primary-500 focus:border-primary-500"
                aria-label={t('languageSelectorLabel')}
              >
                <option value="en">English</option>
                <option value="zh">中文</option>
              </select>
              <Languages size={12} className="absolute right-2 top-1/2 transform -translate-y-1/2 text-gray-400 pointer-events-none" />
            </div>

            {/* Theme Toggle Button */}
            <button
              onClick={toggleDarkMode}
              className="ml-2 p-1.5 rounded-md text-gray-500 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-inset focus:ring-primary-500"
              aria-label={darkMode ? t('switchToLightMode') : t('switchToDarkMode')}
            >
              {darkMode ? <Sun size={16} className="text-yellow-400" /> : <Moon size={16} className="text-primary-600" />} {/* Adjusted size */}
            </button>
          </div>
        </div>
      </header>

      {/* Main content area with Sidebar */}
      <div className="flex flex-grow overflow-hidden"> {/* Horizontal layout, prevent overflow */}

        {/* Sidebar (Conversation List) */}
        <aside className="w-64 flex-shrink-0 overflow-y-auto bg-white dark:bg-gray-850 border-r border-gray-200 dark:border-gray-700"> {/* Added background and border */}
          <ConversationList />
        </aside>

        {/* Main Chat Area */}
        <main className="flex-grow flex flex-col md:flex-row p-4 md:p-5 gap-4 max-w-none overflow-hidden bg-gray-100 dark:bg-gray-900"> {/* Added background */}

          {/* Left Panel (File Upload) - Now inside main chat area */}
          <section className="w-full md:w-1/3 lg:w-1/4 bg-white dark:bg-gray-850 rounded-lg shadow-sm p-4 overflow-y-auto border border-gray-200 dark:border-gray-700 flex-shrink-0 md:max-h-[calc(100vh-100px)]"> {/* Adjusted padding, max height, reduced header height assumption */}
            <h2 className="text-base font-semibold mb-3 pb-2 border-b border-gray-200 dark:border-gray-700 text-gray-700 dark:text-gray-300">{t('fileUploadTitle')}</h2>
            {currentUserId ? (
               <FileUpload />
            ) : (
               <p className="text-sm text-yellow-600 dark:text-yellow-400">{t('setUserIdToUpload', '请先设置用户ID才能上传文件。')}</p>
            )}
          </section>

          {/* Right Panel (Chat Interface) - Now inside main chat area */}
          <section className="w-full md:w-2/3 lg:w-3/4 bg-white dark:bg-gray-850 rounded-lg shadow-sm flex flex-col overflow-hidden border border-gray-200 dark:border-gray-700">
            {/* Chat header remains */}
            <div className="p-3 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 flex-shrink-0">
              <h2 className="text-base font-semibold text-gray-700 dark:text-gray-300">{t('aiAssistantTitle')}</h2>
              <p className="text-xs text-gray-500 dark:text-gray-400">{t('aiAssistantSubtitle')}</p>
            </div>
            {/* ChatInterface takes remaining space */}
            <div className="flex-grow overflow-hidden"> {/* Add wrapper for ChatInterface */}
                <ChatInterface />
            </div>
          </section>
        </main>
      </div>

      {/* Footer (Optional, might be removed with sidebar layout) */}
      {/*
      <footer className="bg-gray-50 dark:bg-gray-850 p-3 text-center text-xs text-gray-500 dark:text-gray-400 border-t border-gray-200 dark:border-gray-700 flex-shrink-0">
        <p>{t('footerText', { year: new Date().getFullYear() })}</p>
      </footer>
      */}
    </div>
  );
}

export default App;
