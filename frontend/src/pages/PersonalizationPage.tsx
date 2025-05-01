import React, { useState, useEffect, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import * as api from '../services/api'; // Import API functions

type Tab = 'memory' | 'knowledge' | 'prompt';

const PersonalizationPage: React.FC = () => {
  const { t } = useTranslation();
  const [activeTab, setActiveTab] = useState<Tab>('memory');

// --- State for Structured Memory Tab ---
  const [memories, setMemories] = useState<api.StructuredMemory[]>([]);
  const [memoryLoading, setMemoryLoading] = useState<boolean>(false);
  const [memoryError, setMemoryError] = useState<string | null>(null);
  const [newMemoryKey, setNewMemoryKey] = useState<string>('');
  const [newMemoryValue, setNewMemoryValue] = useState<string>('');
  const [isAddingMemory, setIsAddingMemory] = useState<boolean>(false);
  const [editingMemoryKey, setEditingMemoryKey] = useState<string | null>(null); // Key of the item being edited
  const [editingMemoryValue, setEditingMemoryValue] = useState<string>('');
  const [isSavingEdit, setIsSavingEdit] = useState<boolean>(false);
  const [isDeletingKey, setIsDeletingKey] = useState<string | null>(null);

  // --- State for Knowledge Base Tab ---
  const [documents, setDocuments] = useState<api.DocumentInfo[]>([]);
  const [docsLoading, setDocsLoading] = useState<boolean>(false);
  const [docsError, setDocsError] = useState<string | null>(null);
  const [isDeletingDocId, setIsDeletingDocId] = useState<string | null>(null);

  // --- State for Custom Prompt Tab ---
  const [customPrompt, setCustomPrompt] = useState<string>('');
  const [promptLoading, setPromptLoading] = useState<boolean>(false);
  const [promptSaving, setPromptSaving] = useState<boolean>(false);
  const [promptError, setPromptError] = useState<string | null>(null);
  const [promptSuccess, setPromptSuccess] = useState<string | null>(null);

  // --- Fetch Memories ---
  const fetchMemories = useCallback(async () => {
    setMemoryLoading(true);
    setMemoryError(null);
    try {
      const data = await api.getMemories();
      setMemories(data);
    } catch (err) {
      setMemoryError(err instanceof Error ? err.message : t('personalizationPage.memory.errorLoading'));
    } finally {
      setMemoryLoading(false);
    }
  }, [t]); // Add t to dependency array

  // --- Add Memory ---
   const handleAddMemory = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newMemoryKey || !newMemoryValue) {
        setMemoryError(t('personalizationPage.memory.errorEmptyFields'));
        return;
    }
    setIsAddingMemory(true);
    setMemoryError(null);
    try {
        // Attempt to parse value as JSON, otherwise keep as string
        let parsedValue: unknown;
        try {
            parsedValue = JSON.parse(newMemoryValue);
        } catch { // Removed unused jsonError
            parsedValue = newMemoryValue; // Keep as string if not valid JSON
        }
        await api.createMemory({ key: newMemoryKey, value: parsedValue });
        setNewMemoryKey('');
        setNewMemoryValue('');
        await fetchMemories(); // Refresh list after adding
    } catch (err) {
        setMemoryError(err instanceof Error ? err.message : t('personalizationPage.memory.errorAdding'));
    } finally {
        setIsAddingMemory(false);
    }
};

  // --- Edit Memory ---
  const handleEditClick = (memory: api.StructuredMemory) => {
    setEditingMemoryKey(memory.key);
    // Store value as string for editing, even if it's an object/array
    setEditingMemoryValue(typeof memory.value === 'object' ? JSON.stringify(memory.value, null, 2) : String(memory.value));
    setMemoryError(null); // Clear previous errors
  };

  const handleCancelEdit = () => {
    setEditingMemoryKey(null);
    setEditingMemoryValue('');
  };

  const handleSaveEdit = async () => {
    if (editingMemoryKey === null) return;

    setIsSavingEdit(true);
    setMemoryError(null);
    try {
        let parsedValue: unknown;
        try {
            parsedValue = JSON.parse(editingMemoryValue);
        } catch { // Removed unused jsonError
            parsedValue = editingMemoryValue; // Keep as string if not valid JSON
        }
        await api.updateMemory(editingMemoryKey, parsedValue);
        setEditingMemoryKey(null);
        setEditingMemoryValue('');
        await fetchMemories(); // Refresh list
    } catch (err) {
        setMemoryError(err instanceof Error ? err.message : t('personalizationPage.memory.errorUpdating'));
    } finally {
        setIsSavingEdit(false);
    }
  };

  // --- Delete Memory ---
  const handleDeleteMemory = async (key: string) => {
    if (window.confirm(t('personalizationPage.memory.deleteConfirmation', { key }))) {
        setIsDeletingKey(key);
        setMemoryError(null);
        try {
            await api.deleteMemory(key);
            await fetchMemories(); // Refresh list
        } catch (err) {
            setMemoryError(err instanceof Error ? err.message : t('personalizationPage.memory.errorDeleting'));
        } finally {
            setIsDeletingKey(null);
        }
    }
  };
  
    // --- Fetch Documents ---
    const fetchDocuments = useCallback(async () => {
      setDocsLoading(true);
      setDocsError(null);
      try {
        const data = await api.getDocuments();
        setDocuments(data);
      } catch (err) {
        setDocsError(err instanceof Error ? err.message : t('personalizationPage.knowledge.errorLoading'));
      } finally {
        setDocsLoading(false);
      }
    }, [t]); // Add t dependency
  
    // --- Delete Document ---
    const handleDeleteDocument = async (docId: string, filename: string) => {
      if (window.confirm(t('personalizationPage.knowledge.deleteConfirmation', { filename }))) {
          setIsDeletingDocId(docId);
          setDocsError(null);
          try {
              await api.deleteDocument(docId);
              await fetchDocuments(); // Refresh list
          } catch (err) {
              setDocsError(err instanceof Error ? err.message : t('personalizationPage.knowledge.errorDeleting'));
          } finally {
              setIsDeletingDocId(null);
          }
      }
    };
  
    // --- Fetch Custom Prompt ---
     const fetchCustomPrompt = useCallback(async () => {
      setPromptLoading(true);
      setPromptError(null);
      setPromptSuccess(null); // Clear success message on load
      try {
        const config = await api.getUserConfig();
        setCustomPrompt(config.custom_prompt || ''); // Set to empty string if null/undefined
      } catch (err) {
        setPromptError(err instanceof Error ? err.message : t('personalizationPage.prompt.errorLoading'));
      } finally {
        setPromptLoading(false);
      }
    }, [t]); // Add t dependency
  
    // --- Save Custom Prompt ---
    const handleSavePrompt = async () => {
      setPromptSaving(true);
      setPromptError(null);
      setPromptSuccess(null);
      try {
        await api.updateUserConfig({ custom_prompt: customPrompt });
        setPromptSuccess(t('personalizationPage.prompt.saveSuccess'));
        // Optionally refetch or just assume success
        // await fetchCustomPrompt();
      } catch (err) {
        setPromptError(err instanceof Error ? err.message : t('personalizationPage.prompt.errorSaving'));
      } finally {
        setPromptSaving(false);
      }
    };
  
  
    // --- Load data when tab becomes active ---
    useEffect(() => {
      if (activeTab === 'memory') {
        fetchMemories();
      } else if (activeTab === 'knowledge') {
        fetchDocuments();
      } else if (activeTab === 'prompt') {
        fetchCustomPrompt();
      }
    }, [activeTab, fetchMemories, fetchDocuments, fetchCustomPrompt]); // Add fetchCustomPrompt dependency

  const renderTabContent = () => {
    switch (activeTab) {
      case 'memory':
        return (
          <div>
            <h2 className="text-xl font-semibold mb-4">{t('personalizationPage.memory.title')}</h2>

            {/* Add New Memory Form */}
            <form onSubmit={handleAddMemory} className="mb-6 p-4 border rounded dark:border-gray-600">
              <h3 className="text-lg font-medium mb-3">{t('personalizationPage.memory.addTitle')}</h3>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-3">
                <input
                  type="text"
                  placeholder={t('personalizationPage.memory.keyPlaceholder')}
                  value={newMemoryKey}
                  onChange={(e) => setNewMemoryKey(e.target.value)}
                  className="p-2 border rounded dark:bg-gray-700 dark:border-gray-600"
                  required
                />
                <textarea
                  placeholder={t('personalizationPage.memory.valuePlaceholder')}
                  value={newMemoryValue}
                  onChange={(e) => setNewMemoryValue(e.target.value)}
                  className="p-2 border rounded dark:bg-gray-700 dark:border-gray-600 md:col-span-2"
                  rows={2}
                  required
                />
              </div>
              <button
                type="submit"
                disabled={isAddingMemory || memoryLoading}
                className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
              >
                {isAddingMemory ? t('personalizationPage.memory.addingButton') : t('personalizationPage.memory.addButton')}
              </button>
            </form>

            {/* Display Memory List */}
            {memoryLoading && <p>{t('personalizationPage.memory.loading')}</p>}
            {memoryError && <p className="text-red-500 mb-4">{memoryError}</p>}
            {!memoryLoading && !memoryError && memories.length === 0 && (
              <p>{t('personalizationPage.memory.noEntries')}</p>
            )}
            {!memoryLoading && memories.length > 0 && (
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
                  <thead className="bg-gray-50 dark:bg-gray-800">
                    <tr>
                      <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider dark:text-gray-400">
                        {t('personalizationPage.memory.tableHeaderKey')}
                      </th>
                      <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider dark:text-gray-400">
                        {t('personalizationPage.memory.tableHeaderValue')}
                      </th>
                      <th scope="col" className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider dark:text-gray-400">
                        {t('personalizationPage.memory.tableHeaderActions')}
                      </th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200 dark:bg-gray-900 dark:divide-gray-700">
                    {memories.map((memory) => (
                      <tr key={memory.id}>
                        <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-white align-top">
                          {memory.key}
                        </td>
                        <td className="px-6 py-4 whitespace-pre-wrap text-sm text-gray-500 dark:text-gray-300 align-top">
                          {editingMemoryKey === memory.key ? (
                            <textarea
                              value={editingMemoryValue}
                              onChange={(e) => setEditingMemoryValue(e.target.value)}
                              className="w-full p-2 border rounded dark:bg-gray-700 dark:border-gray-600 text-sm"
                              rows={3}
                              disabled={isSavingEdit}
                            />
                          ) : (
                            <div className="max-h-32 overflow-y-auto"> {/* Limit height and allow scroll */}
                              {typeof memory.value === 'object' ? JSON.stringify(memory.value, null, 2) : String(memory.value)}
                            </div>
                          )}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium align-top">
                          {editingMemoryKey === memory.key ? (
                            <>
                              <button
                                onClick={handleSaveEdit}
                                disabled={isSavingEdit}
                                className="text-green-600 hover:text-green-900 dark:text-green-400 dark:hover:text-green-200 mr-3 disabled:opacity-50"
                              >
                                {isSavingEdit ? t('personalizationPage.memory.savingButton') : t('personalizationPage.memory.saveButton')}
                              </button>
                              <button
                                onClick={handleCancelEdit}
                                disabled={isSavingEdit}
                                className="text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-200 disabled:opacity-50"
                              >
                                {t('personalizationPage.memory.cancelButton')}
                              </button>
                            </>
                          ) : (
                            <>
                              <button
                                onClick={() => handleEditClick(memory)}
                                disabled={!!editingMemoryKey || !!isDeletingKey} // Disable if any edit/delete is in progress
                                className="text-indigo-600 hover:text-indigo-900 dark:text-indigo-400 dark:hover:text-indigo-200 mr-3 disabled:opacity-50"
                              >
                                {t('personalizationPage.memory.editButton')}
                              </button>
                              <button
                                onClick={() => handleDeleteMemory(memory.key)}
                                disabled={!!editingMemoryKey || isDeletingKey === memory.key} // Disable if editing or this one is being deleted
                                className="text-red-600 hover:text-red-900 dark:text-red-400 dark:hover:text-red-200 disabled:opacity-50"
                              >
                                {isDeletingKey === memory.key ? t('personalizationPage.memory.deletingButton') : t('personalizationPage.memory.deleteButton')}
                              </button>
                            </>
                          )}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        );
      case 'knowledge':
        return (
          <div>
            <h2 className="text-xl font-semibold mb-4">{t('personalizationPage.knowledge.title')}</h2>

            {docsLoading && <p>{t('personalizationPage.knowledge.loading')}</p>}
            {docsError && <p className="text-red-500 mb-4">{docsError}</p>}
            {!docsLoading && !docsError && documents.length === 0 && (
              <p>{t('personalizationPage.knowledge.noDocuments')}</p>
            )}
            {!docsLoading && documents.length > 0 && (
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
                  <thead className="bg-gray-50 dark:bg-gray-800">
                    <tr>
                      <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider dark:text-gray-400">
                        {t('personalizationPage.knowledge.tableHeaderFilename')}
                      </th>
                      <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider dark:text-gray-400">
                        {t('personalizationPage.knowledge.tableHeaderStatus')}
                      </th>
                       <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider dark:text-gray-400">
                        {t('personalizationPage.knowledge.tableHeaderUploadedAt')}
                      </th>
                      <th scope="col" className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider dark:text-gray-400">
                        {t('personalizationPage.knowledge.tableHeaderActions')}
                      </th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200 dark:bg-gray-900 dark:divide-gray-700">
                    {documents.map((doc) => (
                      <tr key={doc.id}>
                        <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-white">{doc.filename}</td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-300">
                           <span className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${
                               doc.status === 'completed' ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200' :
                               doc.status === 'failed' ? 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200' :
                               'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200' // pending or processing
                           }`}>
                               {t(`personalizationPage.knowledge.status.${doc.status}`)}
                           </span>
                        </td>
                         <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-300">
                            {new Date(doc.created_at).toLocaleString()}
                         </td>
                        <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                          <button
                            onClick={() => handleDeleteDocument(doc.id, doc.filename)}
                            disabled={isDeletingDocId === doc.id}
                            className="text-red-600 hover:text-red-900 dark:text-red-400 dark:hover:text-red-200 disabled:opacity-50"
                          >
                            {isDeletingDocId === doc.id ? t('personalizationPage.knowledge.deletingButton') : t('personalizationPage.knowledge.deleteButton')}
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        );
      case 'prompt':
        return (
          <div>
            <h2 className="text-xl font-semibold mb-4">{t('personalizationPage.prompt.title')}</h2>
            <p className="text-sm text-gray-600 dark:text-gray-400 mb-4">
              {t('personalizationPage.prompt.description')}
            </p>

            {promptLoading && <p>{t('personalizationPage.prompt.loading')}</p>}
            {promptError && <p className="text-red-500 mb-4">{promptError}</p>}
            {promptSuccess && <p className="text-green-500 mb-4">{promptSuccess}</p>}

            {!promptLoading && (
              <>
                <textarea
                  value={customPrompt}
                  onChange={(e) => setCustomPrompt(e.target.value)}
                  placeholder={t('personalizationPage.prompt.placeholder')}
                  className="w-full p-3 border rounded dark:bg-gray-700 dark:border-gray-600 mb-4 h-48 resize-y" // Added h-48 and resize-y
                  disabled={promptSaving}
                />
                <button
                  onClick={handleSavePrompt}
                  disabled={promptSaving || promptLoading}
                  className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
                >
                  {promptSaving ? t('personalizationPage.prompt.savingButton') : t('personalizationPage.prompt.saveButton')}
                </button>
              </>
            )}
          </div>
        );
      default:
        return null;
    }
  };

  return (
    <div className="p-4 md:p-6 lg:p-8 max-w-4xl mx-auto">
      <h1 className="text-2xl md:text-3xl font-bold mb-6">{t('personalizationPage.title')}</h1>

      <div className="border-b border-gray-200 dark:border-gray-700 mb-6">
        <nav className="-mb-px flex space-x-6" aria-label="Tabs">
          <button
            onClick={() => setActiveTab('memory')}
            className={`whitespace-nowrap py-3 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'memory'
                ? 'border-indigo-500 text-indigo-600 dark:text-indigo-400'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-200 dark:hover:border-gray-500'
            }`}
          >
            {t('personalizationPage.tabs.memory')}
          </button>
          <button
            onClick={() => setActiveTab('knowledge')}
            className={`whitespace-nowrap py-3 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'knowledge'
                ? 'border-indigo-500 text-indigo-600 dark:text-indigo-400'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-200 dark:hover:border-gray-500'
            }`}
          >
            {t('personalizationPage.tabs.knowledge')}
          </button>
          <button
            onClick={() => setActiveTab('prompt')}
            className={`whitespace-nowrap py-3 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'prompt'
                ? 'border-indigo-500 text-indigo-600 dark:text-indigo-400'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-200 dark:hover:border-gray-500'
            }`}
          >
            {t('personalizationPage.tabs.prompt')}
          </button>
        </nav>
      </div>

      <div>
        {renderTabContent()}
      </div>
    </div>
  );
};

export default PersonalizationPage;