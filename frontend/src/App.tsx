// import { useState } from 'react' // No longer needed
// import reactLogo from './assets/react.svg' // No longer needed
// import viteLogo from '/vite.svg' // No longer needed
// import './App.css' // We'll use index.css with Tailwind
import { useState, useEffect } from 'react';
import { Moon, Sun } from 'lucide-react';

// Import the components we created
import FileUpload from './components/FileUpload';
import ChatInterface from './components/ChatInterface';

function App() {
  // 添加深色模式状态管理
  const [darkMode, setDarkMode] = useState(false);

  // 监听系统主题变化并初始化
  useEffect(() => {
    // 检查本地存储中的主题设置
    const savedTheme = localStorage.getItem('theme');
    // 检查系统偏好
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
    
    // 设置初始主题
    if (savedTheme === 'dark' || (!savedTheme && prefersDark)) {
      setDarkMode(true);
      document.documentElement.classList.add('dark');
    } else {
      setDarkMode(false);
      document.documentElement.classList.remove('dark');
    }
  }, []);

  // 切换主题
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

  return (
    // Main container using Tailwind classes
    <div className="flex flex-col h-screen bg-gradient-to-br from-gray-50 to-gray-100 dark:from-gray-900 dark:to-gray-800 text-gray-900 dark:text-gray-100">
      {/* Header with gradient and shadow */}
      <header className="bg-gradient-to-r from-primary-600 to-primary-700 text-white p-4 shadow-md relative z-10">
        <div className="container mx-auto flex items-center justify-between">
          <h1 className="text-2xl font-display font-bold tracking-tight">DreamHub</h1>
          <div className="flex items-center space-x-2">
            <span className="inline-flex h-2 w-2 rounded-full bg-green-400 mr-2 animate-pulse"></span>
            <span className="text-sm font-medium text-white/90">在线</span>
          </div>
        </div>
      </header>

      {/* Main content area */}
      <main className="flex-grow flex flex-col md:flex-row container mx-auto p-4 md:p-8 gap-6 max-w-7xl">
        {/* Left Panel (File Upload) */}
        <section className="w-full md:w-1/3 lg:w-1/4 bg-white dark:bg-gray-800 rounded-xl shadow-soft p-5 transition-all duration-300 hover:shadow-lg overflow-y-auto border border-gray-100 dark:border-gray-700"> 
          <h2 className="text-lg font-display font-semibold mb-4 border-b pb-2 dark:border-gray-600 text-primary-800 dark:text-primary-300">文件上传</h2>
          {/* Render FileUpload component */}
          <FileUpload />
        </section>

        {/* Right Panel (Chat Interface) */}
        <section className="w-full md:w-2/3 lg:w-3/4 bg-white dark:bg-gray-800 rounded-xl shadow-soft flex flex-col overflow-hidden transition-all duration-300 hover:shadow-lg border border-gray-100 dark:border-gray-700">
          {/* Chat header */}
          <div className="p-4 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-750">
            <h2 className="text-lg font-display font-semibold text-primary-800 dark:text-primary-300">AI 助手</h2>
            <p className="text-sm text-gray-500 dark:text-gray-400">针对您的文档提问任何问题</p>
          </div>
          
          {/* Render ChatInterface component */}
          <ChatInterface />
        </section>
      </main>

      {/* Footer */}
      <footer className="bg-white dark:bg-gray-800 p-4 text-center text-sm text-gray-500 dark:text-gray-400 border-t border-gray-200 dark:border-gray-700 shadow-inner-lg">
        <p>© 2023 DreamHub - 智能文档助手</p>
      </footer>

      {/* 浮动主题切换按钮 */}
      <button
        onClick={toggleDarkMode}
        className="fixed bottom-6 right-6 p-3 rounded-full bg-white dark:bg-gray-700 shadow-lg hover:shadow-xl transition-all duration-300 focus:outline-none focus:ring-2 focus:ring-primary-500 dark:focus:ring-primary-400"
        aria-label={darkMode ? '切换到亮色模式' : '切换到深色模式'}
      >
        {darkMode ? (
          <Sun size={24} className="text-yellow-500" />
        ) : (
          <Moon size={24} className="text-primary-700" />
        )}
      </button>
    </div>
  )
}

export default App
