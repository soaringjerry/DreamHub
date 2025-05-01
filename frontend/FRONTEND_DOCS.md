# DreamHub Frontend Documentation

## 1. Overview

This document provides an overview of the DreamHub frontend application. It's built using React, TypeScript, and Vite, styled with Tailwind CSS. State management is handled by Zustand, and internationalization (i18n) is managed using i18next.

**Tech Stack:**

*   Framework: React 19
*   Language: TypeScript
*   Build Tool: Vite
*   Styling: Tailwind CSS
*   State Management: Zustand
*   Internationalization: i18next
*   API Client: Axios

## 2. Getting Started

1.  Navigate to the `frontend` directory: `cd frontend`
2.  Install dependencies: `npm install`
3.  Run the development server: `npm run dev`

The application will be available at `http://localhost:5173` (or another port if 5173 is busy).

## 3. Project Structure

```
frontend/
├── public/         # Static assets
├── src/            # Source code
│   ├── assets/     # Image/font assets
│   ├── components/ # Reusable React components
│   ├── services/   # API interaction logic (api.ts)
│   ├── store/      # Zustand state management (chatStore.ts)
│   ├── App.tsx     # Main application component, layout
│   ├── i18n.ts     # i18next configuration and translations
│   ├── main.tsx    # Application entry point
│   └── ...         # Other source files (CSS, types)
├── .env.example    # Environment variable example
├── index.html      # HTML entry point
├── package.json    # Project dependencies and scripts
├── tailwind.config.js # Tailwind configuration
├── tsconfig.json   # TypeScript configuration
└── vite.config.ts  # Vite configuration
```

## 4. State Management (Zustand - `src/store/`)

Global application state is managed using Zustand. State is likely split across multiple stores for better organization:

**`chatStore.ts` (Core Chat Logic):**

*   `conversations: Record<string, Conversation>`: An object storing chat conversations fetched from the backend, keyed by `conversationId`. Each `Conversation` object likely contains:
   *   `id: string`
   *   `title: string` (May be generated or user-defined)
    *   `title: string`
    *   `messages: Message[]` (Array of `Message` objects: `{ sender: 'user' | 'ai', content: string, timestamp: number }`)
    *   `createdAt: number`
    *   `lastUpdatedAt: number` (or string timestamp)
*   `activeConversationId: string | null`: The ID of the currently displayed conversation.
*   `conversationStatus: Record<string, { isLoading: boolean; error: string | null }>`: Tracks loading and error states per conversation.
*   `isUploading: boolean`: Global flag for file upload status.
*   `uploadError: string | null`: Global error message for file uploads.
*   `uploadedFiles: UploadedFile[]`: List of uploaded files (metadata, likely fetched from backend).
*   `conversationList: ConversationMeta[]`: Array holding metadata for the conversation list (e.g., `{ id: string, title: string, last_message_timestamp: string }`).

**Persistence:**

*   With cloud sync, `localStorage` persistence for `conversations` and `uploadedFiles` is likely **removed**. `activeConversationId` might still be persisted locally for UX.

**Key Actions (`ChatActions`):**

*   `fetchConversations`: Action to fetch the list of conversation metadata from the backend (`GET /conversations`).
*   `fetchConversationMessages`: Action to fetch messages for a specific conversation (`GET /chat/{convId}/messages`).
*   `startNewConversation`: May still exist locally, but likely triggers backend interaction upon first message.
*   `switchConversation`: Changes the `activeConversationId` and potentially triggers `fetchConversationMessages`.
*   `deleteConversation`: Calls the backend API to delete a conversation and updates the local list.
*   `renameConversation`: Calls the backend API to rename a conversation and updates the local list/details.
*   `addMessage`: Adds a message locally (optimistic update) and potentially updates based on API response.
*   `sendMessage`: Handles sending a user message (adds user message locally, calls API `POST /chat`, updates with AI response/error). Handles creating new conversations on the backend if `conversationId` is null.
*   `uploadFile`: Handles file upload (calls API `POST /upload`, updates UI based on response).
*   `fetchUploadedFiles`: Action to fetch the list of uploaded documents (`GET /documents`).
*   `deleteUploadedFile`: Calls the backend API (`DELETE /documents/{docId}`) and updates the local list.
*   `setConversationStatus`: Updates the loading/error status for a specific conversation ID.
*   `setUploading`, `setUploadError`: Manage global upload state.

**`authStore.ts` (Authentication):**

*   Manages user authentication state (e.g., JWT token, user info, login/logout status).
*   Provides actions for `login`, `register`, `logout`.
*   Likely persists the auth token securely (e.g., `localStorage` or `sessionStorage`).

**(Potential) `configStore.ts` (User Configuration):**

*   Manages user-specific settings fetched from `/users/me/config`.
*   State might include `openaiApiKey` (potentially masked), `defaultModel`, etc.
*   Actions: `fetchConfig`, `updateConfig`.

**(Potential) `memoryStore.ts` (Structured Memory):**

*   Manages structured memory entries fetched from `/memory/structured`.
*   State: `memories: Record<string, StructuredMemoryEntry>` or an array.
*   Actions: `fetchMemories`, `addOrUpdateMemory`, `deleteMemory`.

**Selectors:**

*   Selectors will exist within each store to provide optimized access to specific parts of the state (e.g., `useActiveConversationId`, `useActiveMessages` in `chatStore`, `useIsAuthenticated` in `authStore`, `useUserConfig` in `configStore`).

*Note: Selectors returning objects or arrays should still aim to return stable references when possible to prevent unnecessary re-renders.*


## 5. Components (`src/components/` and `src/pages/`)

*   **`App.tsx`**: The root component. Sets up the main layout (header, sidebar, main content area), handles theme switching, language switching, and User ID input/display. Integrates `ConversationList`, `FileUpload`, and `ChatInterface`.
*   **`ConversationList.tsx`**: Displays the list of conversations in the sidebar. Allows switching, creating, and deleting conversations. Uses `useMemo` to optimize rendering.
*   **`ChatInterface.tsx`**: The main container for the chat area. Integrates `MessageDisplay` and `UserInput`. Displays loading/error states for the active conversation and provides a "New Chat" button.
*   **`MessageDisplay.tsx`**: Renders the messages for the currently active conversation. Handles Markdown rendering, code highlighting, and distinguishes between user/AI messages and status messages. Automatically scrolls to the bottom.
*   **`UserInput.tsx`**: Provides the text area for user input, the send button, and quick prompt suggestions. Handles message sending logic and disables input during loading.
*   **`FileUpload.tsx`**: Handles the file drag-and-drop/selection UI and initiates the file upload process via the `uploadFile` action. Displays the list of uploaded files.

*   **`SettingsPage.tsx`**: (页面组件) 允许用户查看和修改他们的配置，例如 OpenAI API 密钥和默认模型。它与 `configStore` (如果存在) 或直接与 API 交互来获取和保存设置。
*   **`PersonalizationPage.tsx`**: (页面组件) 提供界面来管理用户的结构化记忆（键值对）。允许用户查看、添加、编辑和删除记忆条目。它与 `memoryStore` (如果存在) 或直接与 API 交互。
## 6. Key Features

*   **User Authentication:** Users can register and log in. Authentication is handled via JWT tokens, enabling secure access to user-specific data.
*   **Cloud-Synced Conversations:** Chat history is stored on the server and fetched dynamically, allowing access across devices. The conversation list is displayed in the sidebar.
*   **Document Management:** Users can upload documents for RAG. The list of uploaded documents can be viewed, and individual documents can be deleted.
*   **User Settings:** A dedicated settings page allows users to configure application aspects, such as their OpenAI API key and preferred AI model.
*   **Personalization (Structured Memory):** A dedicated page allows users to manage key-value pairs (structured memory) for personalization purposes.
*   **Markdown & Code Highlighting:** AI responses are rendered as Markdown, with support for syntax highlighting.
*   **Dark Mode & i18n:** Supports theme switching (light/dark) and language switching (English/Chinese).

## 7. API Interaction (`src/services/api.ts`)

All communication with the backend API is centralized in `api.ts`. It uses `axios` to make requests.

*   `login(credentials)`: Sends login request to `/api/v1/auth/login`.
*   `register(userInfo)`: Sends registration request to `/api/v1/auth/register`.
*   `fetchConversations()`: Fetches conversation list from `/api/v1/conversations`.
*   `fetchConversationMessages(conversationId)`: Fetches messages for a conversation from `/api/v1/chat/{conversationId}/messages`.
*   `sendMessage(message, conversationId)`: Sends a user message to `/api/v1/chat`. (No longer needs `userId`).
*   `uploadFile(file)`: Uploads a file to `/api/v1/upload`. (No longer needs `userId`).
*   `fetchDocuments()`: Fetches the list of uploaded documents from `/api/v1/documents`.
*   `deleteDocument(docId)`: Deletes a document via `/api/v1/documents/{docId}`.
*   `getUserConfig()`: Fetches user configuration from `/api/v1/users/me/config`.
*   `updateUserConfig(configData)`: Updates user configuration via `PUT /api/v1/users/me/config`.
*   `getStructuredMemories()`: Fetches all structured memories from `/api/v1/memory/structured`.
*   `getStructuredMemory(key)`: Fetches a specific memory by key from `/api/v1/memory/structured/{key}`.
*   `createOrUpdateStructuredMemory(memoryData)`: Creates/updates a memory via `POST /api/v1/memory/structured`.
*   `updateStructuredMemoryByKey(key, valueData)`: Updates a memory by key via `PUT /api/v1/memory/structured/{key}`.
*   `deleteStructuredMemory(key)`: Deletes a memory by key via `DELETE /api/v1/memory/structured/{key}`.

*Note: The `api.ts` file likely includes an Axios instance configured with interceptors to automatically attach the authentication token (JWT) to relevant requests.*

## 8. Internationalization (i18n - `src/i18n.ts`)

Uses `i18next` and `react-i18next`. Translations for English (`en`) and Chinese (`zh`) are defined directly within the `i18n.ts` configuration file. Components use the `useTranslation` hook to access translated strings via keys.

## 9. Styling

Tailwind CSS is used for styling, configured in `tailwind.config.js` and `postcss.config.js`. Global styles and Tailwind base layers are included via `App.css` (or similar entry CSS file). Dark mode is supported using Tailwind's `dark:` variant, toggled by adding/removing the `dark` class on the `<html>` element.