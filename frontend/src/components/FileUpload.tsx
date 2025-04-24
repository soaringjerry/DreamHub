// src/components/FileUpload.tsx
import React, { useState, useCallback } from 'react';
import { useDropzone } from 'react-dropzone';
import { useChatStore } from '../store/chatStore'; // 导入 Zustand store
import { UploadCloud, X, FileText, File, FileType, FileCode, CheckCircle2 } from 'lucide-react'; // 添加成功图标

// 定义上传文件接口以匹配store中的类型
interface UploadedFile {
  name: string;
  size: number;
  chunks: number;
  id: string;
}

const FileUpload: React.FC = () => {
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [dropActive, setDropActive] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0); // 添加上传进度状态

  // --- Zustand Store Integration ---
  // 使用单独的选择器获取状态和 action
  const uploadFile = useChatStore((state) => state.uploadFile);
  const isUploading = useChatStore((state) => state.isUploading);
  const uploadError = useChatStore((state) => state.uploadError);
  const setUploadError = useChatStore((state) => state.setUploadError);
  const uploadedFiles = useChatStore((state) => state.uploadedFiles as UploadedFile[]); // 获取已上传文件列表

  // --- Helper Function ---
  // 改进的文件类型图标选择
  const getFileTypeIcon = (fileName: string | undefined) => {
    if (!fileName) return <FileText size={20} className="mr-2 flex-shrink-0 text-gray-500" />;
    const extension = fileName.split('.').pop()?.toLowerCase();
    
    switch (extension) {
      case 'pdf':
        return <FileType size={20} className="mr-2 flex-shrink-0 text-red-500" />;
      case 'docx':
      case 'doc':
        return <FileText size={20} className="mr-2 flex-shrink-0 text-blue-500" />;
      case 'txt':
        return <FileText size={20} className="mr-2 flex-shrink-0 text-gray-500" />;
      case 'js':
      case 'jsx':
      case 'ts':
      case 'tsx':
        return <FileCode size={20} className="mr-2 flex-shrink-0 text-yellow-500" />;
      default:
        return <File size={20} className="mr-2 flex-shrink-0 text-gray-500" />;
    }
  };

  // 处理文件上传
  const handleUpload = async () => {
    if (selectedFile) {
      setUploadError(null);
      setUploadProgress(0); // 重置进度
      
      // 模拟上传进度 (实际上传中应从后端获取进度)
      const progressInterval = setInterval(() => {
        setUploadProgress(prev => {
          // 随机增加进度，最大到95%（最后5%在上传完成后设置）
          const nextProgress = Math.min(prev + Math.random() * 15, 95);
          return nextProgress;
        });
      }, 400);
      
      try {
        await uploadFile(selectedFile);
        clearInterval(progressInterval);
        setUploadProgress(100); // 上传完成
        
        // 成功上传后3秒重置UI
        setTimeout(() => {
          setSelectedFile(null);
          setUploadProgress(0);
        }, 3000);
      } catch (error) {
        clearInterval(progressInterval);
        setUploadProgress(0);
      }
    }
  };

  // 配置 react-dropzone
  const onDrop = useCallback((acceptedFiles: File[]) => {
    setDropActive(false);
    if (acceptedFiles && acceptedFiles[0]) {
      setSelectedFile(acceptedFiles[0]);
      setUploadError(null);
    }
  }, [setUploadError]);

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    onDragEnter: () => setDropActive(true),
    onDragLeave: () => setDropActive(false),
    accept: { 
      'text/plain': ['.txt'],
      'application/pdf': ['.pdf'],
      'application/vnd.openxmlformats-officedocument.wordprocessingml.document': ['.docx'],
      'application/msword': ['.doc'],
    },
    multiple: false,
  });

  // 检查文件是否已上传
  const isFileAlreadyUploaded = selectedFile && uploadedFiles?.some(
    (file: UploadedFile) => file.name === selectedFile.name && file.size === selectedFile.size
  );

  // --- Render ---
  return (
    <div className="space-y-4">
      {/* 已上传文件列表 */}
      {uploadedFiles && uploadedFiles.length > 0 && (
        <div className="mb-4">
          <h3 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2 flex items-center">
            <CheckCircle2 size={16} className="mr-1.5 text-green-500" />
            已上传文件
          </h3>
          <div className="space-y-2 max-h-[120px] overflow-y-auto pr-1">
            {uploadedFiles.map((file: UploadedFile, idx: number) => (
              <div key={idx} className="flex items-center p-2 rounded bg-green-50 dark:bg-green-900/20 border border-green-100 dark:border-green-800 text-xs">
                {getFileTypeIcon(file.name)}
                <span className="truncate">{file.name}</span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Dropzone Area with Animation */}
      <div
        {...getRootProps()}
        className={`flex flex-col items-center justify-center p-6 border-2 border-dashed rounded-xl cursor-pointer transition-all duration-300 ease-in-out text-center
                   ${dropActive || isDragActive
                     ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/20 scale-105 shadow-md' 
                     : isFileAlreadyUploaded 
                       ? 'border-green-500 bg-green-50 dark:bg-green-900/20'
                       : 'border-gray-300 dark:border-gray-600 hover:border-gray-400 dark:hover:border-gray-500'}
                   ${uploadError ? 'border-red-500 bg-red-50 dark:bg-red-900/20' : ''}`}
      >
        <input {...getInputProps()} aria-label="选择文件上传区域" />
        <div className={`transition-all duration-300 ${dropActive || isDragActive ? 'scale-110' : ''}`}>
          <div className={`w-16 h-16 mx-auto mb-3 rounded-full flex items-center justify-center
                        ${isFileAlreadyUploaded
                          ? 'bg-green-100 dark:bg-green-800/30 text-green-600 dark:text-green-400'
                          : dropActive || isDragActive 
                            ? 'bg-primary-100 dark:bg-primary-800/30 text-primary-600 dark:text-primary-400'
                            : 'bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400'}`}>
            <UploadCloud size={30} className={`${dropActive || isDragActive ? 'animate-bounce' : ''}`} />
          </div>
          
          {dropActive || isDragActive ? (
            <p className="text-sm font-semibold text-primary-600 dark:text-primary-400">释放文件以上传...</p>
          ) : isFileAlreadyUploaded ? (
            <p className="text-sm font-semibold text-green-600 dark:text-green-400">文件已准备就绪</p>
          ) : (
            <div>
              <p className="text-sm font-semibold text-gray-700 dark:text-gray-300">将文件拖放到这里</p>
              <p className="text-xs text-gray-500 dark:text-gray-400">或点击选择文件</p>
              <p className="text-xs text-gray-400 dark:text-gray-500 mt-2 px-3 py-1 rounded-full bg-gray-100 dark:bg-gray-700 inline-block">
                支持 TXT, PDF, DOCX
              </p>
            </div>
          )}
        </div>
      </div>

      {/* Display selected file name with remove button */}
      {selectedFile && (
        <div className="flex items-center justify-between p-3 rounded-lg bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 shadow-sm transition-all duration-200 hover:shadow-md">
           <div className="flex items-center overflow-hidden mr-2 flex-grow">
             {getFileTypeIcon(selectedFile?.name)}
             <div className="truncate">
               <p className="text-sm font-medium text-gray-800 dark:text-gray-200 truncate" title={selectedFile.name}>
                 {selectedFile.name}
               </p>
               <p className="text-xs text-gray-500 dark:text-gray-400">
                 {(selectedFile.size / 1024).toFixed(1)} KB
               </p>
             </div>
           </div>
          <button
            onClick={(e) => {
              e.stopPropagation();
              setSelectedFile(null);
              setUploadProgress(0);
            }}
            className="p-1.5 rounded-full text-gray-500 hover:bg-gray-200 dark:hover:bg-gray-600 focus:outline-none transition-colors duration-150"
            aria-label="移除文件"
            disabled={isUploading}
          >
            <X size={16} />
          </button>
        </div>
      )}

      {/* 上传进度条 */}
      {uploadProgress > 0 && uploadProgress < 100 && (
        <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2.5 mb-4 overflow-hidden">
          <div 
            className="bg-primary-600 h-2.5 rounded-full transition-all duration-300 ease-out"
            style={{ width: `${uploadProgress}%` }}
            role="progressbar" 
            aria-valuenow={uploadProgress} 
            aria-valuemin={0} 
            aria-valuemax={100}
          ></div>
        </div>
      )}
      
      {/* Upload success indicator */}
      {uploadProgress === 100 && (
        <div className="flex items-center p-2 rounded-lg bg-green-50 dark:bg-green-900/20 text-green-800 dark:text-green-300 border border-green-200 dark:border-green-800 animate-fadeIn">
          <CheckCircle2 size={18} className="mr-2 text-green-500" />
          <span className="text-sm font-medium">上传成功</span>
        </div>
      )}

      {/* Upload Button */}
      <button
        onClick={handleUpload}
        disabled={!!(!selectedFile || isUploading || uploadProgress === 100 || isFileAlreadyUploaded)}
        className={`w-full px-4 py-2.5 rounded-lg text-white transition-all duration-200 ease-in-out
                   ${isUploading ? 'bg-gray-400 cursor-not-allowed' 
                               : !selectedFile || isFileAlreadyUploaded ? 'bg-gray-300 dark:bg-gray-700 cursor-not-allowed'
                               : 'bg-gradient-to-r from-primary-600 to-primary-500 hover:from-primary-700 hover:to-primary-600 shadow-sm hover:shadow active:scale-98'}
                   disabled:opacity-60`}
      >
        <div className="flex items-center justify-center">
          {isUploading ? (
            <>
              <svg className="animate-spin h-5 w-5 mr-2" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              <span>上传中...</span>
            </>
          ) : uploadProgress === 100 ? (
            <span>上传完成</span>
          ) : isFileAlreadyUploaded ? (
            <span>文件已上传</span>
          ) : (
            <>
              <UploadCloud size={18} className="mr-2" />
              <span>上传文件</span>
            </>
          )}
        </div>
      </button>

      {/* Upload Error */}
      {uploadError && (
        <div className="flex items-center p-3 text-sm rounded-lg bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-800 animate-fadeIn" role="alert">
          <svg className="flex-shrink-0 inline w-5 h-5 mr-3 text-red-500 dark:text-red-400" fill="currentColor" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg">
            <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zm-1 9a1 1 0 102 0V11a1 1 0 10-2 0v4z" clipRule="evenodd"></path>
          </svg>
          <div>
            <span className="font-medium text-red-800 dark:text-red-300">上传失败</span> 
            <p className="text-red-600 dark:text-red-400">{uploadError}</p>
          </div>
        </div>
      )}
    </div>
  );
};

export default FileUpload;