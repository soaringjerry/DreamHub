import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { BrowserRouter as Router, Routes, Route, Link, useNavigate } from 'react-router-dom'; // Import Router components
import i18n from './i18n';
import { Moon, Sun, Languages, User, LogOut } from 'lucide-react'; // Removed Save import
import { useAuthStore, useIsAuthenticated, useCurrentUser } from './store/authStore'; // Import auth store hooks
// Removed chatStore import for userId as it's handled by authStore now

import FileUpload from './components/FileUpload';
import ChatInterface from './components/ChatInterface';
import ConversationList from './components/ConversationList';
import LoginPage from './pages/LoginPage'; // Import Login Page
import RegisterPage from './pages/RegisterPage'; // Import Register Page
import ProtectedRoute from './components/ProtectedRoute'; // Import Protected Route

// Main application layout component
const MainLayout: React.FC = () => {
  const { t } = useTranslation();
  const isAuthenticated = useIsAuthenticated(); // Use auth state for conditional rendering

  return (
    <div className="flex flex-grow overflow-hidden"> {/* Horizontal layout, prevent overflow */}
      {/* Sidebar (Conversation List) - Only shown if authenticated */}
      {isAuthenticated && (
        <aside className="w-64 flex-shrink-0 overflow-y-auto bg-white dark:bg-gray-850 border-r border-gray-200 dark:border-gray-700">
          <ConversationList />
        </aside>
      )}

      {/* Main Chat Area */}
      <main className="flex-grow flex flex-col md:flex-row p-4 md:p-5 gap-4 max-w-none overflow-hidden bg-gray-100 dark:bg-gray-900">

        {/* Left Panel (File Upload) - Only shown if authenticated */}
        {isAuthenticated && (
          <section className="w-full md:w-1/3 lg:w-1/4 bg-white dark:bg-gray-850 rounded-lg shadow-sm p-4 overflow-y-auto border border-gray-200 dark:border-gray-700 flex-shrink-0 md:max-h-[calc(100vh-100px)]">
            <h2 className="text-base font-semibold mb-3 pb-2 border-b border-gray-200 dark:border-gray-700 text-gray-700 dark:text-gray-300">{t('fileUploadTitle')}</h2>
            <FileUpload /> {/* Render directly as ProtectedRoute handles auth check */}
          </section>
        )}

        {/* Right Panel (Chat Interface) - Only shown if authenticated */}
        {isAuthenticated && (
          <section className="w-full md:w-2/3 lg:w-3/4 bg-white dark:bg-gray-850 rounded-lg shadow-sm flex flex-col overflow-hidden border border-gray-200 dark:border-gray-700">
            <div className="p-3 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 flex-shrink-0">
              <h2 className="text-base font-semibold text-gray-700 dark:text-gray-300">{t('aiAssistantTitle')}</h2>
              <p className="text-xs text-gray-500 dark:text-gray-400">{t('aiAssistantSubtitle')}</p>
            </div>
            <div className="flex-grow overflow-hidden">
              <ChatInterface /> {/* Render directly */}
            </div>
          </section>
        )}

        {/* Placeholder for non-authenticated users on main route? Or rely on redirect */}
        {!isAuthenticated && (
           <div className="flex items-center justify-center w-full">
               <p>{t('auth.pleaseLoginPrompt')}</p> {/* Or redirect handled by ProtectedRoute */}
           </div>
        )}
      </main>
    </div>
  );
};


function App() {
  const { t } = useTranslation();
  const [darkMode, setDarkMode] = useState(false);
  const [currentLanguage, setCurrentLanguage] = useState(i18n.language);

  // Auth state and actions
  const isAuthenticated = useIsAuthenticated();
  const currentUser = useCurrentUser();
  const logout = useAuthStore((state) => state.logout);
  const navigate = useNavigate(); // Hook for programmatic navigation

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

  // Handle logout
  const handleLogout = () => {
    logout();
    navigate('/login'); // Redirect to login after logout
  };

  return (
    <div className="flex flex-col h-screen bg-gray-100 dark:bg-gray-900 text-gray-900 dark:text-gray-100 font-sans overflow-hidden">
      {/* Header */}
      <header className="bg-white dark:bg-gray-850 text-gray-800 dark:text-white p-3 border-b border-gray-200 dark:border-gray-700 shadow-sm relative z-10 flex-shrink-0">
        <div className="container mx-auto flex items-center justify-between px-4">
          <h1 className="text-lg font-semibold tracking-tight text-primary-600 dark:text-primary-400">{t('appTitle')}</h1>
          <div className="flex items-center space-x-2">
            {/* Online status (Keep as is) */}
            <span className="relative flex h-2.5 w-2.5">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
              <span className="relative inline-flex rounded-full h-2.5 w-2.5 bg-green-500"></span>
            </span>
            <span className="text-xs font-medium text-gray-600 dark:text-gray-400">{t('onlineStatus')}</span>

            {/* User Info / Logout Button */}
            {isAuthenticated && currentUser ? (
              <div className="flex items-center space-x-2 ml-3 border-l border-gray-300 dark:border-gray-600 pl-3">
                <User size={14} className="text-gray-500 dark:text-gray-400 flex-shrink-0" />
                <span className="text-xs font-medium text-gray-600 dark:text-gray-400 truncate" title={currentUser.username}>
                  {currentUser.username}
                </span>
                <button
                  onClick={handleLogout}
                  className="p-1 rounded-md text-gray-500 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 focus:outline-none focus:ring-1 focus:ring-inset focus:ring-primary-500 flex-shrink-0"
                  aria-label={t('auth.logoutButton')}
                >
                  <LogOut size={14} />
                </button>
              </div>
            ) : (
              // Optionally show Login/Register links if not authenticated
              <div className="flex items-center space-x-2 ml-3 border-l border-gray-300 dark:border-gray-600 pl-3">
                 <Link to="/login" className="text-xs font-medium text-indigo-600 hover:text-indigo-500">{t('auth.loginLink')}</Link>
                 <span className="text-gray-300 dark:text-gray-600">|</span>
                 <Link to="/register" className="text-xs font-medium text-indigo-600 hover:text-indigo-500">{t('auth.registerLink')}</Link>
              </div>
            )}

            {/* Language Selector */}
            <div className="relative ml-3">
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
              {darkMode ? <Sun size={16} className="text-yellow-400" /> : <Moon size={16} className="text-primary-600" />}
            </button>
          </div>
        </div>
      </header>

      {/* Main content area managed by Routes */}
      <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />
          <Route path="/" element={<ProtectedRoute />}>
              {/* Child route for the main authenticated layout */}
              <Route index element={<MainLayout />} />
              {/* Add other protected routes here if needed */}
          </Route>
          {/* Optional: Add a 404 Not Found route */}
          {/* <Route path="*" element={<NotFoundPage />} /> */}
      </Routes>

    </div>
  );
}

// Wrap App with Router
const AppWrapper: React.FC = () => (
  <Router>
    <App />
  </Router>
);


export default AppWrapper; // Export the wrapper
