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

// --- Config API Interfaces ---

// Matches dto.UserConfigResponse (excluding sensitive data like raw API key)
export interface UserConfigResponse {
  api_endpoint: string | null;
  default_model: string | null;
  api_key_encrypted: boolean;
  custom_prompt?: string | null; // Add custom prompt field
  created_at: string;
  updated_at: string;
}

// Matches dto.UpdateUserConfigRequest
export interface UpdateUserConfigRequest {
  api_endpoint?: string | null;
  default_model?: string | null;
  api_key?: string | null;
  custom_prompt?: string | null; // Add custom prompt field
}

// --- Chat History API Interfaces ---

// Matches backend entity.ConversationInfo (adjust based on actual backend response)
export interface ConversationInfo {
  id: string;
  user_id: string; // May not be needed if backend filters by authenticated user
  title: string;
  created_at: string; // ISO string format
  updated_at: string; // ISO string format
}

// Matches backend entity.Message (adjust based on actual backend response)
export interface Message {
  id: string;
  conversation_id: string;
  role: 'user' | 'assistant';
  content: string;
  created_at: string; // ISO string format
  // Add other fields if present, e.g., model, metadata
}

// --- Structured Memory API Interfaces ---

// Matches backend entity.StructuredMemory
export interface StructuredMemory {
  id: string; // UUID
  user_id: string; // UUID
  key: string;
  value: unknown; // JSONB in backend, can be any valid JSON type
  created_at: string; // ISO string format
  updated_at: string; // ISO string format
}

// Payload for creating/updating memory
export interface MemoryPayload {
  key: string;
  value: unknown;
}

// --- Document API Interfaces ---

// Matches backend entity.Document (adjust based on actual backend response)
export interface DocumentInfo {
    id: string; // UUID
    user_id: string; // UUID
    filename: string;
    filepath: string; // Or maybe just filename is exposed? Check backend.
    status: 'pending' | 'processing' | 'completed' | 'failed';
    created_at: string; // ISO string format
    updated_at: string; // ISO string format
    // Add other relevant fields like error_message if available
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
    // Only add token to requests that are not for auth endpoints
    if (config.url && !config.url.startsWith('/auth/')) {
      const authStorage = localStorage.getItem('auth-storage'); // Read the persisted zustand state
      if (authStorage) {
        try {
          const authData = JSON.parse(authStorage);
          const token = authData?.state?.token; // Access the token within the persisted state object
          if (token) {
            config.headers.Authorization = `Bearer ${token}`;
            console.log('Attaching token to request:', config.url); // Added for debugging
          } else {
             console.warn('Token not found in auth-storage for request:', config.url); // Added for debugging
          }
        } catch (e) {
          console.error('Failed to parse auth-storage from localStorage', e);
        }
      } else {
         console.warn('auth-storage not found in localStorage for request:', config.url); // Added for debugging
      }
    }
    return config;
  },
  (error: AxiosError): Promise<AxiosError> => {
    // Handle request error
    return Promise.reject(error);
  }
);

// Response interceptor to handle errors, specifically 401 for logout
apiClient.interceptors.response.use(
  (response) => response, // Pass through successful responses
  async (error: AxiosError) => {
    const originalRequest = error.config; // Get the original request config

    // Check if it's a 401 error and not a request to the login/register endpoint itself
    if (error.response?.status === 401 && originalRequest && !originalRequest.url?.includes('/auth/')) {
      console.error('API Error: 401 Unauthorized. Logging out.');
      try {
        // Dynamically import the store to avoid circular dependencies if api.ts is imported in store
        const { useAuthStore } = await import('../store/authStore');
        // Get the logout function outside of a component context
        const logout = useAuthStore.getState().logout;
        logout(); // Call the logout action

        // Redirect to login page after logout
        // Check if running in a browser environment before redirecting
        if (typeof window !== 'undefined') {
           window.location.href = '/login'; // Force reload to clear state
        }

        // Optionally, you might want to prevent the original error from propagating further
        // by returning a resolved promise or a specific error object.
        // However, rejecting might be desired to let the calling code know the request failed.
        return Promise.reject(new Error('Session expired. Please log in again.'));

      } catch (logoutError) {
        console.error('Failed to execute logout action:', logoutError);
        // Still reject the original error even if logout fails
        return Promise.reject(error);
      }
    }

    // For other errors, just reject the promise
    return Promise.reject(error);
  }
);


// --- Helper for Error Handling ---
// This helper might still be useful for non-401 errors or if you want more specific messages
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

// --- User Configuration API Functions ---

/**
 * Fetches the current user's configuration.
 * Assumes user is authenticated (token handled by interceptor).
 * @returns Promise containing the user's configuration
 */
export const getUserConfig = async (): Promise<UserConfigResponse> => {
  try {
    const response = await apiClient.get<UserConfigResponse>('/users/me/config');
    return response.data;
  } catch (error) {
    throw handleApiError(error, 'Failed to fetch user configuration');
  }
};

// --- Chat History API Functions ---

/**
 * Fetches the list of conversations for the current user.
 * Assumes user is authenticated (token handled by interceptor).
 * @returns Promise containing an array of ConversationInfo objects
 */
export const getUserConversations = async (): Promise<ConversationInfo[]> => {
  try {
    const response = await apiClient.get<ConversationInfo[]>('/conversations');
    return response.data;
  } catch (error) {
    throw handleApiError(error, 'Failed to fetch conversations');
  }
};

/**
 * Fetches the messages for a specific conversation.
 * Assumes user is authenticated (token handled by interceptor).
 * @param conversationId The ID of the conversation to fetch messages for.
 * @returns Promise containing an array of Message objects
 */
export const getConversationMessages = async (conversationId: string): Promise<Message[]> => {
  if (!conversationId) {
    // Return an empty array or reject? Rejecting seems better for explicit error handling.
    return Promise.reject(new Error('Conversation ID is required'));
  }
  try {
    // Ensure the URL is correctly formed, e.g., /api/v1/chat/{conversationId}/messages
    const response = await apiClient.get<Message[]>(`/chat/${conversationId}/messages`);
    return response.data;
  } catch (error) {
    throw handleApiError(error, `Failed to fetch messages for conversation ${conversationId}`);
  }
};

/**
 * Updates the current user's configuration.
 * Assumes user is authenticated (token handled by interceptor).
 * @param configData The configuration data to update. Fields can be omitted for partial updates.
 *                   Send null to clear optional fields like api_key.
 * @returns Promise containing the updated user configuration
 */
export const updateUserConfig = async (configData: UpdateUserConfigRequest): Promise<UserConfigResponse> => {
  try {
    // Backend expects PUT request with JSON body
    const response = await apiClient.put<UserConfigResponse>('/users/me/config', configData, {
      headers: {
        'Content-Type': 'application/json',
      },
    });
    return response.data;
  } catch (error) {
    throw handleApiError(error, 'Failed to update user configuration');
  }
};

// --- Structured Memory API Functions ---

/**
 * Creates a new structured memory entry.
 * @param payload The key-value pair to store.
 * @returns Promise containing the created memory entry.
 */
export const createMemory = async (payload: MemoryPayload): Promise<StructuredMemory> => {
  try {
    const response = await apiClient.post<StructuredMemory>('/memory/structured', payload);
    return response.data;
  } catch (error) {
    throw handleApiError(error, 'Failed to create memory entry');
  }
};

/**
 * Fetches all structured memory entries for the current user.
 * @returns Promise containing an array of memory entries.
 */
export const getMemories = async (): Promise<StructuredMemory[]> => {
  try {
    const response = await apiClient.get<StructuredMemory[]>('/memory/structured');
    return response.data;
  } catch (error) {
    throw handleApiError(error, 'Failed to fetch memory entries');
  }
};

/**
 * Fetches a specific structured memory entry by key.
 * @param key The key of the memory entry to fetch.
 * @returns Promise containing the memory entry.
 */
export const getMemoryByKey = async (key: string): Promise<StructuredMemory> => {
  try {
    const response = await apiClient.get<StructuredMemory>(`/memory/structured/${encodeURIComponent(key)}`);
    return response.data;
  } catch (error) {
    throw handleApiError(error, `Failed to fetch memory entry with key: ${key}`);
  }
};

/**
 * Updates an existing structured memory entry.
 * @param key The key of the memory entry to update.
 * @param value The new value for the entry.
 * @returns Promise containing the updated memory entry.
 */
export const updateMemory = async (key: string, value: unknown): Promise<StructuredMemory> => {
    const payload: Partial<MemoryPayload> = { value }; // Backend expects value in body for PUT
    try {
      const response = await apiClient.put<StructuredMemory>(`/memory/structured/${encodeURIComponent(key)}`, payload);
      return response.data;
    } catch (error) {
      throw handleApiError(error, `Failed to update memory entry with key: ${key}`);
    }
  };

/**
 * Deletes a structured memory entry by key.
 * @param key The key of the memory entry to delete.
 * @returns Promise that resolves when deletion is successful.
 */
export const deleteMemory = async (key: string): Promise<void> => {
  try {
    await apiClient.delete(`/memory/structured/${encodeURIComponent(key)}`);
  } catch (error) {
    throw handleApiError(error, `Failed to delete memory entry with key: ${key}`);
  }
};

// --- Document API Functions ---

/**
 * Fetches the list of documents for the current user.
 * @returns Promise containing an array of DocumentInfo objects.
 */
export const getDocuments = async (): Promise<DocumentInfo[]> => {
  try {
    // Assuming the backend returns an array directly under /documents
    const response = await apiClient.get<DocumentInfo[]>('/documents');
    return response.data;
  } catch (error) {
    throw handleApiError(error, 'Failed to fetch documents');
  }
};

/**
 * Deletes a document by its ID.
 * @param docId The ID of the document to delete.
 * @returns Promise that resolves when deletion is successful.
 */
export const deleteDocument = async (docId: string): Promise<void> => {
  try {
    await apiClient.delete(`/documents/${docId}`);
  } catch (error) {
    throw handleApiError(error, `Failed to delete document with ID: ${docId}`);
  }
};


// TODO: Add functions for other endpoints if needed (e.g., getTaskStatus)
// Remember to use `apiClient` for these calls as well.