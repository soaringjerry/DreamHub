// src/components/FileUpload.tsx
import React, { useState, useCallback } from 'react';
import { useTranslation } from 'react-i18next'; // 导入 useTranslation
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
  const { t } = useTranslation(); // 初始化 useTranslation
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
      } catch (error) { // eslint-disable-line @typescript-eslint/no-unused-vars
        // TODO: Handle upload error (e.g., show notification to user)
        console.error("Upload failed:", error); // Add basic error logging
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
    <div className="space-y-5"> {/* Increased spacing */}
      {/* 已上传文件列表: Improved styling */}
      {uploadedFiles && uploadedFiles.length > 0 && (
        <div className="mb-4">
          <h3 className="text-xs font-semibold text-gray-500 dark:text-gray-400 mb-2 uppercase tracking-wider flex items-center">
            <CheckCircle2 size={14} className="mr-1.5 text-green-500" />
            {t('processedFiles')}
          </h3>
          <div className="space-y-1.5 max-h-[150px] overflow-y-auto pr-1"> {/* Increased max height */}
            {uploadedFiles.map((file: UploadedFile, idx: number) => (
              <div key={idx} className="flex items-center p-2 rounded-md bg-green-50 dark:bg-gray-700 border border-green-100 dark:border-gray-600 text-xs shadow-sm">
                {getFileTypeIcon(file.name)}
                <span className="truncate text-green-800 dark:text-green-300 font-medium">{file.name}</span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Dropzone Area: Refined styles */}
      <div
        {...getRootProps()}
        className={`flex flex-col items-center justify-center p-6 border border-dashed rounded-lg cursor-pointer transition-all duration-200 ease-in-out text-center
                   ${dropActive || isDragActive
                     ? 'border-primary-400 bg-primary-50 dark:bg-primary-900/10 ring-2 ring-primary-300 ring-offset-1 dark:ring-offset-gray-850' // Enhanced active state
                     : isFileAlreadyUploaded
                       ? 'border-green-400 bg-green-50 dark:bg-gray-700' // Success state
                       : 'border-gray-300 dark:border-gray-600 hover:border-gray-400 dark:hover:border-gray-500 bg-gray-50 dark:bg-gray-800'}
                   ${uploadError ? 'border-red-400 bg-red-50 dark:bg-red-900/10' : ''}`} // Error state
      >
        <input {...getInputProps()} aria-label={t('fileDropzoneAreaLabel')} />
        <div className={`transition-transform duration-200 ${dropActive || isDragActive ? 'scale-105' : ''}`}>
          {/* Icon styling */}
          <div className={`w-12 h-12 mx-auto mb-3 rounded-lg flex items-center justify-center transition-colors duration-200
                        ${isFileAlreadyUploaded
                          ? 'bg-green-100 dark:bg-green-800/50 text-green-600 dark:text-green-400'
                          : dropActive || isDragActive
                            ? 'bg-primary-100 dark:bg-primary-800/50 text-primary-600 dark:text-primary-400'
                            : 'bg-gray-100 dark:bg-gray-700 text-gray-400 dark:text-gray-500'}`}>
            <UploadCloud size={24} /> {/* Adjusted icon size */}
          </div>

          {/* Text styling */}
          {dropActive || isDragActive ? (
            <p className="text-sm font-medium text-primary-600 dark:text-primary-400">{t('dropFileToUpload')}</p>
          ) : isFileAlreadyUploaded ? (
            <p className="text-sm font-medium text-green-600 dark:text-green-400">{t('fileReady')}</p>
          ) : (
            <div>
              <p className="text-sm font-medium text-gray-700 dark:text-gray-300">{t('dragOrClickToSelect')}</p>
              <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">{t('supportedFormats')}</p>
            </div>
          )}
        </div>
      </div>

      {/* Display selected file name: Refined styling */}
      {selectedFile && !isFileAlreadyUploaded && uploadProgress === 0 && ( // Only show if not uploaded and not in progress
        <div className="flex items-center justify-between p-2.5 rounded-md bg-gray-100 dark:bg-gray-750 border border-gray-200 dark:border-gray-600 shadow-sm">
           <div className="flex items-center overflow-hidden mr-2 flex-grow min-w-0"> {/* Added min-w-0 */}
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
            className="p-1 rounded-full text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-600 hover:text-gray-600 dark:hover:text-gray-200 focus:outline-none transition-colors duration-150 flex-shrink-0" // Added flex-shrink-0
            aria-label={t('removeFileLabel')}
            disabled={isUploading}
          >
            <X size={16} />
          </button>
        </div>
      )}

      {/* 上传进度条: Refined styling */}
      {uploadProgress > 0 && uploadProgress < 100 && (
        <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2 overflow-hidden"> {/* Adjusted height */}
          <div
            className="bg-primary-500 h-2 rounded-full transition-all duration-300 ease-out" // Adjusted color and height
            style={{ width: `${uploadProgress}%` }}
            role="progressbar"
            aria-valuenow={uploadProgress}
            aria-valuemin={0}
            aria-valuemax={100}
          ></div>
        </div>
      )}

      {/* Upload success indicator: Refined styling */}
      {uploadProgress === 100 && (
        <div className="flex items-center p-2.5 rounded-md bg-green-50 dark:bg-green-900/30 text-green-700 dark:text-green-300 border border-green-200 dark:border-green-700 animate-fadeIn">
          <CheckCircle2 size={16} className="mr-2 flex-shrink-0" /> {/* Adjusted size */}
          <span className="text-sm font-medium">{t('uploadSuccessMessage')}</span>
        </div>
      )}

      {/* Upload Button: Refined styling and states */}
      <button
        onClick={handleUpload}
        disabled={!!(!selectedFile || isUploading || uploadProgress === 100 || isFileAlreadyUploaded)}
        className={`w-full px-4 py-2 rounded-md text-sm font-medium transition-all duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-offset-2 dark:focus:ring-offset-gray-850
                   ${isUploading
                     ? 'bg-gray-400 dark:bg-gray-600 text-white cursor-wait' // Uploading state
                     : !selectedFile || isFileAlreadyUploaded || uploadProgress === 100
                       ? 'bg-gray-300 dark:bg-gray-700 text-gray-500 dark:text-gray-400 cursor-not-allowed' // Disabled state
                       : 'bg-primary-600 hover:bg-primary-700 text-white shadow-sm focus:ring-primary-500 active:bg-primary-800' // Active state
                   } disabled:opacity-70`} // General disabled opacity
      >
        <div className="flex items-center justify-center">
          {isUploading ? (
            <>
              <svg className="animate-spin h-4 w-4 mr-2" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              <span>{t('processingButton')}</span>
            </>
          ) : uploadProgress === 100 ? (
            <span className="flex items-center"><CheckCircle2 size={16} className="mr-1.5"/>{t('doneButton')}</span>
          ) : isFileAlreadyUploaded ? (
            <span>{t('fileReadyButton')}</span>
          ) : (
            <>
              <UploadCloud size={16} className="mr-1.5" /> {/* Adjusted size */}
              <span>{t('uploadAndProcessButton')}</span>
            </>
          )}
        </div>
      </button>

      {/* Upload Error: Refined styling */}
      {uploadError && (
        <div className="flex items-start p-3 text-sm rounded-md bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-700 animate-fadeIn" role="alert"> {/* Changed to items-start */}
          <svg className="flex-shrink-0 inline w-4 h-4 mr-2 mt-0.5 text-red-500 dark:text-red-400" fill="currentColor" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg"> {/* Adjusted size and margin */}
            <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zm-1 9a1 1 0 102 0V11a1 1 0 10-2 0v4z" clipRule="evenodd"></path>
          </svg>
          <div>
            <span className="font-semibold text-red-700 dark:text-red-300">{t('uploadErrorTitle')}</span>
            <p className="text-red-600 dark:text-red-400 mt-0.5">{uploadError}</p> {/* Added margin */}
          </div>
        </div>
      )}
    </div>
  );
};

export default FileUpload;