import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios';

// 后端 API 的基础 URL
const API_BASE_URL = '/api/v1';

// --- Interfaces based on backend definitions ---

// User info safe to expose (matches entity.SanitizedUser)
export interface SanitizedUser { // Add export keyword
	id: string; // UUID as string
	username: string;
	created_at: string; // Assuming ISO string format from Go time.Time
  updated_at: string;
}

// Payload for registration (matches service.RegisterPayload)
export interface RegisterPayload {
  username: string;
  password?: string; // Make password optional here if validation is done elsewhere, or required
}

// Payload for login (matches service.LoginCredentials)
export interface LoginCredentials {
  username: string;
  password?: string; // Make password optional here if validation is done elsewhere, or required
}

// Response after successful login (matches service.LoginResponse)
export interface LoginResponse {
  token: string;
  user: SanitizedUser;
}

// Response after successful registration (backend returns SanitizedUser)
export type RegisterResponse = SanitizedUser;

// Existing response types
interface UploadResponse {
  message: string;
  filename: string;
  doc_id: string; // Added based on backend handler response
  task_id: string;
}

interface ChatResponse {
  conversation_id: string;
  reply: string;
}

// --- Axios Instance and Interceptor ---

// Create a dedicated Axios instance
const apiClient = axios.create({
  baseURL: API_BASE_URL,
});

// Request interceptor to add JWT token
apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig): InternalAxiosRequestConfig => {
    // Only add token to requests that are not for auth endpoints
    if (config.url && !config.url.startsWith('/auth/')) {
      const token = localStorage.getItem('authToken'); // Assuming token is stored in localStorage
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
    }
    return config;
  },
  (error: AxiosError): Promise<AxiosError> => {
    // Handle request error
    return Promise.reject(error);
  }
);

// --- Helper for Error Handling ---
const handleApiError = (error: unknown, defaultMessage: string): Error => {
  console.error('API Error:', error);
  if (axios.isAxiosError(error) && error.response) {
    // Try to get error message from backend response structure { error: { message: '...' } }
    const backendErrorMessage = error.response.data?.error?.message;
    return new Error(`${defaultMessage}: ${backendErrorMessage || error.message}`);
  }
  return new Error(`${defaultMessage}: An unknown error occurred.`);
};


// --- API Functions ---

/**
 * Registers a new user.
 * @param payload Registration data (username, password)
 * @returns Promise containing the registered user's info (SanitizedUser)
 */
export const registerUser = async (payload: RegisterPayload): Promise<RegisterResponse> => {
  try {
    const response = await apiClient.post<RegisterResponse>('/auth/register', payload);
    return response.data;
  } catch (error) {
    throw handleApiError(error, 'Registration failed');
  }
};

/**
 * Logs in a user.
 * @param credentials Login data (username, password)
 * @returns Promise containing the JWT and user info
 */
export const loginUser = async (credentials: LoginCredentials): Promise<LoginResponse> => {
  try {
    const response = await apiClient.post<LoginResponse>('/auth/login', credentials);
    // Store token upon successful login (consider moving this logic to where loginUser is called, e.g., auth store/hook)
    // localStorage.setItem('authToken', response.data.token);
    return response.data;
  } catch (error) {
    // Clear token if login fails? Maybe not here.
    // localStorage.removeItem('authToken');
    throw handleApiError(error, 'Login failed');
  }
};

/**
 * Uploads a file to the backend for processing.
 * User ID is now handled by the backend via JWT.
 * @param file The file object to upload
 * @returns Promise containing the upload result
 */
export const uploadFile = async (file: File): Promise<UploadResponse> => {
  const formData = new FormData();
  formData.append('file', file); // Backend expects 'file' field

  try {
    // Use the configured apiClient which includes the auth header
    const response = await apiClient.post<UploadResponse>('/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  } catch (error) {
    throw handleApiError(error, 'Upload failed');
  }
};

/**
 * Sends a chat message to the backend.
 * User ID is now handled by the backend via JWT.
 * @param message The message content
 * @param conversationId Optional conversation ID
 * @returns Promise containing the AI reply and conversation ID
 */
export const sendMessage = async (message: string, conversationId?: string): Promise<ChatResponse> => {
  // Payload no longer includes user_id
  const payload: { message: string; conversation_id?: string } = {
    message,
  };
  if (conversationId) {
    payload.conversation_id = conversationId;
  }

  try {
    // Use the configured apiClient which includes the auth header
    const response = await apiClient.post<ChatResponse>('/chat', payload, {
      headers: {
        'Content-Type': 'application/json',
      },
    });
    return response.data;
  } catch (error) {
    throw handleApiError(error, 'Chat request failed');
  }
};

// TODO: Add functions for other endpoints if needed (e.g., listDocuments, getTaskStatus)
// Remember to use `apiClient` for these calls as well.
// Example:
// export const listDocuments = async (limit = 20, offset = 0): Promise<Document[]> => {
//   try {
//     const response = await apiClient.get<Document[]>('/documents', { params: { limit, offset } });
//     return response.data;
//   } catch (error) {
//     throw handleApiError(error, 'Failed to fetch documents');
//   }
// };