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

## 4. State Management (Zustand - `src/store/chatStore.ts`)

Global application state is managed using Zustand. The store (`chatStore.ts`) holds the core application data and logic related to chat functionality.

**State Structure (`ChatState`):**

*   `userId: string | null`: The current user's identifier. Set manually via the UI and stored in `localStorage`.
*   `conversations: Record<string, Conversation>`: An object storing all chat conversations, keyed by `conversationId`. Each `Conversation` object contains:
    *   `id: string`
    *   `title: string`
    *   `messages: Message[]` (Array of `Message` objects: `{ sender: 'user' | 'ai', content: string, timestamp: number }`)
    *   `createdAt: number`
    *   `lastUpdatedAt: number`
*   `activeConversationId: string | null`: The ID of the currently displayed conversation.
*   `conversationStatus: Record<string, { isLoading: boolean; error: string | null }>`: Tracks loading and error states per conversation.
*   `isUploading: boolean`: Global flag for file upload status.
*   `uploadError: string | null`: Global error message for file uploads.
*   `uploadedFiles: UploadedFile[]`: List of successfully uploaded files (metadata).

**Persistence:**

*   The store uses the `persist` middleware to save parts of the state (`userId`, `conversations`, `activeConversationId`, `uploadedFiles`) to `localStorage` under the key `chat-storage`. This allows data to persist across browser sessions.

**Key Actions (`ChatActions`):**

*   `setUserId`: Sets the user ID.
*   `startNewConversation`: Creates a new empty conversation and sets it as active.
*   `switchConversation`: Changes the `activeConversationId`.
*   `deleteConversation`: Removes a conversation.
*   `renameConversation`: Updates a conversation's title.
*   `addMessage`: Adds a message to a specific conversation.
*   `sendMessage`: Handles sending a user message (adds user message, calls API, adds AI response/error to the active conversation). Creates a new conversation if none is active.
*   `uploadFile`: Handles file upload (calls API, adds status messages to the active conversation).
*   `setConversationStatus`: Updates the loading/error status for a specific conversation.
*   `setUploading`, `setUploadError`, `addUploadedFile`: Manage global upload state.

**Selectors:**

*   `useActiveConversationId`: Returns the ID of the active conversation.
*   `useActiveConversation`: Returns the full data object for the active conversation.
*   `useActiveMessages`: Returns the message array for the active conversation (returns a stable empty array reference if no active conversation or messages).
*   `useConversationStatus`: Returns the loading/error status for a specific conversation ID.
*   `useActiveConversationStatus`: Returns the loading/error status for the active conversation (returns a stable default status object reference if needed).

*Note: Selectors returning objects or arrays are designed to return stable references when their underlying data hasn't changed meaningfully, preventing unnecessary component re-renders.*

## 5. Components (`src/components/`)

*   **`App.tsx`**: The root component. Sets up the main layout (header, sidebar, main content area), handles theme switching, language switching, and User ID input/display. Integrates `ConversationList`, `FileUpload`, and `ChatInterface`.
*   **`ConversationList.tsx`**: Displays the list of conversations in the sidebar. Allows switching, creating, and deleting conversations. Uses `useMemo` to optimize rendering.
*   **`ChatInterface.tsx`**: The main container for the chat area. Integrates `MessageDisplay` and `UserInput`. Displays loading/error states for the active conversation and provides a "New Chat" button.
*   **`MessageDisplay.tsx`**: Renders the messages for the currently active conversation. Handles Markdown rendering, code highlighting, and distinguishes between user/AI messages and status messages. Automatically scrolls to the bottom.
*   **`UserInput.tsx`**: Provides the text area for user input, the send button, and quick prompt suggestions. Handles message sending logic and disables input during loading.
*   **`FileUpload.tsx`**: Handles the file drag-and-drop/selection UI and initiates the file upload process via the `uploadFile` action. Displays the list of uploaded files.

## 6. Key Features

*   **User ID Management:** Users can manually set a User ID in the header. This ID is stored in `localStorage` and used in API calls. File uploads are disabled until a User ID is set.
*   **Multi-Conversation Chat:** Users can create multiple independent conversations. The conversation list is displayed in a sidebar, allowing users to switch between them. Conversation data is persisted in `localStorage`.
*   **File Upload:** Users can upload documents (TXT, PDF, DOCX) which are sent to the backend for processing. Upload status is reflected globally and success/error messages are added to the active conversation.
*   **Markdown & Code Highlighting:** AI responses are rendered as Markdown, with support for syntax highlighting in code blocks.
*   **Dark Mode & i18n:** Supports theme switching (light/dark) and language switching (English/Chinese).

## 7. API Interaction (`src/services/api.ts`)

All communication with the backend API is centralized in `api.ts`. It uses `axios` to make requests.

*   `sendMessage(message, conversationId, userId)`: Sends a user message to the `/api/v1/chat` endpoint.
*   `uploadFile(file, userId)`: Uploads a file to the `/api/v1/upload` endpoint.

*Note: Currently, conversation history relies solely on frontend `localStorage`. Backend integration is required for cloud synchronization.*

## 8. Internationalization (i18n - `src/i18n.ts`)

Uses `i18next` and `react-i18next`. Translations for English (`en`) and Chinese (`zh`) are defined directly within the `i18n.ts` configuration file. Components use the `useTranslation` hook to access translated strings via keys.

## 9. Styling

Tailwind CSS is used for styling, configured in `tailwind.config.js` and `postcss.config.js`. Global styles and Tailwind base layers are included via `App.css` (or similar entry CSS file). Dark mode is supported using Tailwind's `dark:` variant, toggled by adding/removing the `dark` class on the `<html>` element.