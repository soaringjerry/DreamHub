import axios from 'axios';

// 后端 API 的基础 URL。由于我们配置了 Vite 代理，
// 在开发环境中可以直接使用相对路径。
// 在生产构建中，可能需要配置一个绝对 URL 或环境变量。
const API_BASE_URL = '/api/v1';

// 定义上传 API 的响应类型 (根据 README.md)
interface UploadResponse {
  message: string;
  filename: string;
  chunks: number;
}

// 定义聊天 API 的响应类型 (根据 README.md)
interface ChatResponse {
  conversation_id: string;
  reply: string;
}

/**
 * 上传文件到后端进行处理。
 * @param file 要上传的文件对象
 * @returns Promise，包含上传结果
 */
export const uploadFile = async (file: File): Promise<UploadResponse> => {
  const formData = new FormData();
  formData.append('file', file); // 后端期望的字段名是 'file'

  try {
    const response = await axios.post<UploadResponse>(`${API_BASE_URL}/upload`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data', // 必须设置正确的 Content-Type
      },
    });
    return response.data;
  } catch (error) {
    console.error('Error uploading file:', error);
    // 可以根据需要进行更具体的错误处理
    if (axios.isAxiosError(error) && error.response) {
      throw new Error(`Upload failed: ${error.response.data?.message || error.message}`);
    }
    throw new Error('An unknown error occurred during file upload.');
  }
};

/**
 * 发送聊天消息到后端。
 * @param message 用户发送的消息内容
 * @param conversationId 可选的当前对话 ID，用于继续对话
 * @returns Promise，包含 AI 的回复和对话 ID
 */
export const sendMessage = async (message: string, conversationId?: string): Promise<ChatResponse> => {
  const payload: { message: string; conversation_id?: string } = { message };
  if (conversationId) {
    payload.conversation_id = conversationId;
  }

  try {
    const response = await axios.post<ChatResponse>(`${API_BASE_URL}/chat`, payload, {
      headers: {
        'Content-Type': 'application/json',
      },
    });
    return response.data;
  } catch (error) {
    console.error('Error sending message:', error);
    if (axios.isAxiosError(error) && error.response) {
      throw new Error(`Chat request failed: ${error.response.data?.message || error.message}`);
    }
    throw new Error('An unknown error occurred while sending the message.');
  }
};