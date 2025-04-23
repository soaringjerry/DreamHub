// src/components/FileUpload.tsx
import React, { useState, useCallback } from 'react';
import { useDropzone } from 'react-dropzone';
import { useChatStore } from '../store/chatStore'; // 导入 Zustand store

const FileUpload: React.FC = () => {
  const [selectedFile, setSelectedFile] = useState<File | null>(null);

  // --- Zustand Store Integration ---
  // 使用单独的选择器获取状态和 action
  const uploadFile = useChatStore((state) => state.uploadFile);
  const isUploading = useChatStore((state) => state.isUploading);
  const uploadError = useChatStore((state) => state.uploadError);
  const setUploadError = useChatStore((state) => state.setUploadError); // 获取 action

  // --- Event Handlers ---
  // 处理文件选择 (如果使用隐藏的 input)
  // const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
  //   if (event.target.files && event.target.files[0]) {
  //     setSelectedFile(event.target.files[0]);
  //     setUploadError(null); // 使用 action 清除错误
  //   }
  // };

  // 处理文件上传
  const handleUpload = async () => {
    if (selectedFile) {
      setUploadError(null); // 清除之前的错误再上传
      await uploadFile(selectedFile);
      // 上传完成后可以清空选择（如果需要）
      // setSelectedFile(null);
    }
  };

  // 配置 react-dropzone
  const onDrop = useCallback((acceptedFiles: File[]) => {
    if (acceptedFiles && acceptedFiles[0]) {
      setSelectedFile(acceptedFiles[0]);
      setUploadError(null); // 使用 action 清除错误
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [setUploadError]); // 依赖 setUploadError (虽然它通常不变，但符合 exhaustive-deps 规则)

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    accept: { // 可以限制接受的文件类型
      'text/plain': ['.txt'],
      'application/pdf': ['.pdf'],
      'application/vnd.openxmlformats-officedocument.wordprocessingml.document': ['.docx'],
      // 根据后端支持添加更多类型
    },
    multiple: false, // 一次只接受一个文件
  });

  // --- Render ---
  return (
    <div className="space-y-4">
      {/* Dropzone Area */}
      <div
        {...getRootProps()}
        className={`p-6 border-2 border-dashed rounded-lg cursor-pointer transition-colors duration-200 ease-in-out
                   ${isDragActive ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/30' : 'border-gray-300 dark:border-gray-600 hover:border-gray-400 dark:hover:border-gray-500'}
                   ${uploadError ? 'border-red-500 bg-red-50 dark:bg-red-900/30' : ''}`}
      >
        <input {...getInputProps()} />
        {isDragActive ? (
          <p className="text-center text-blue-600 dark:text-blue-400">将文件拖放到这里...</p>
        ) : (
          <p className="text-center text-gray-500 dark:text-gray-400">将文件拖放到这里，或点击选择文件</p>
        )}
      </div>

      {/* Display selected file name */}
      {selectedFile && (
        <div className="text-sm text-gray-700 dark:text-gray-300">
          已选择: <span className="font-medium">{selectedFile.name}</span>
        </div>
      )}

      {/* Upload Button */}
      <button
        onClick={handleUpload}
        disabled={!selectedFile || isUploading}
        className={`w-full px-4 py-2 rounded-md text-white transition-colors duration-200 ease-in-out
                   ${isUploading ? 'bg-gray-400 cursor-not-allowed' : 'bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 dark:focus:ring-offset-gray-800'}
                   disabled:opacity-50 disabled:cursor-not-allowed`}
      >
        {isUploading ? '上传中...' : '上传文件'}
      </button>

      {/* Display Upload Error */}
      {uploadError && (
        <p className="text-sm text-red-600 dark:text-red-400 mt-2">
          上传失败: {uploadError}
        </p>
      )}
    </div>
  );
};

export default FileUpload;