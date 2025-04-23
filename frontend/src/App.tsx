// import { useState } from 'react' // No longer needed
// import reactLogo from './assets/react.svg' // No longer needed
// import viteLogo from '/vite.svg' // No longer needed
// import './App.css' // We'll use index.css with Tailwind

// Import the components we created
import FileUpload from './components/FileUpload';
import ChatInterface from './components/ChatInterface';

function App() {
  // const [count, setCount] = useState(0) // No longer needed

  return (
    // Main container using Tailwind classes
    <div className="flex flex-col h-screen bg-gray-100 dark:bg-gray-900 text-gray-900 dark:text-gray-100">
      {/* Header (Optional) */}
      <header className="bg-gray-800 text-white p-4 shadow-md">
        <h1 className="text-xl font-semibold">DreamHub</h1>
      </header>

      {/* Main content area */}
      <main className="flex-grow flex flex-col md:flex-row overflow-hidden p-4 gap-4">

        {/* Left Panel (File Upload) */}
        <section className="w-full md:w-1/4 bg-white dark:bg-gray-800 rounded-lg shadow p-4 overflow-y-auto">
          <h2 className="text-lg font-semibold mb-4 border-b pb-2 dark:border-gray-600">文件上传</h2>
          {/* Render FileUpload component */}
          <FileUpload />
        </section>

        {/* Right Panel (Chat Interface) */}
        {/* Ensure this section takes remaining space and allows ChatInterface to manage its internal height */}
        <section className="w-full md:w-3/4 bg-white dark:bg-gray-800 rounded-lg shadow flex flex-col overflow-hidden">
          {/* Render ChatInterface component */}
          {/* ChatInterface is designed to fill height (h-full), so it should work within this flex container */}
          <ChatInterface />
        </section>

      </main>

      {/* Footer (Optional) */}
      {/* <footer className="bg-gray-200 dark:bg-gray-700 p-2 text-center text-sm">
        Status Bar or Footer Info
      </footer> */}
    </div>
  )
}

export default App
