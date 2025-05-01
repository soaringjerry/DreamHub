import React, { useState, useEffect, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { getUserConfig, updateUserConfig, UserConfigResponse, UpdateUserConfigRequest } from '../services/api';

const SettingsPage: React.FC = () => {
  const { t } = useTranslation();

  const [apiEndpoint, setApiEndpoint] = useState<string>('');
  const [defaultModel, setDefaultModel] = useState<string>('');
  const [apiKey, setApiKey] = useState<string>(''); // Only for entering *new* key
  const [apiKeyIsSet, setApiKeyIsSet] = useState<boolean>(false); // Track if key exists on backend

  const [isLoading, setIsLoading] = useState<boolean>(true); // Start loading initially
  const [isSaving, setIsSaving] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  // Function to fetch config
  const fetchConfig = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    setSuccessMessage(null);
    try {
      const config = await getUserConfig();
      setApiEndpoint(config.api_endpoint || '');
      setDefaultModel(config.default_model || '');
      setApiKeyIsSet(config.api_key_encrypted);
      setApiKey(''); // Always clear API key input on load/reset
    } catch (err) {
      setError(t('settings.loadingError') + ': ' + (err as Error).message);
    } finally {
      setIsLoading(false);
    }
  }, [t]); // Add t to dependency array

  // Fetch config on component mount
  useEffect(() => {
    fetchConfig();
  }, [fetchConfig]);

  const handleSave = async () => {
    setIsSaving(true);
    setError(null);
    setSuccessMessage(null);

    const payload: UpdateUserConfigRequest = {
      // Use null if string is empty, otherwise send the string
      api_endpoint: apiEndpoint.trim() === '' ? null : apiEndpoint.trim(),
      default_model: defaultModel.trim() === '' ? null : defaultModel.trim(),
    };

    // Only include api_key in payload if user actually entered something
    if (apiKey.trim() !== '') {
      payload.api_key = apiKey.trim();
    }

    try {
      const updatedConfig = await updateUserConfig(payload);
      // Update state with response from backend
      setApiEndpoint(updatedConfig.api_endpoint || '');
      setDefaultModel(updatedConfig.default_model || '');
      setApiKeyIsSet(updatedConfig.api_key_encrypted);
      setApiKey(''); // Clear API key input after successful save
      setSuccessMessage(t('settings.saveSuccess'));
    } catch (err) {
      setError(t('settings.saveError') + ': ' + (err as Error).message);
    } finally {
      setIsSaving(false);
    }
  };

  const handleReset = () => {
    // Refetch the original configuration
    fetchConfig();
  };

  return (
    <div className="p-6 max-w-lg mx-auto bg-white rounded-lg shadow-md">
      <h1 className="text-2xl font-semibold mb-6 text-gray-800">{t('settings.title')}</h1>

      {isLoading && <p className="text-blue-600 mb-4">{t('settings.loadingConfig')}</p>}

      <form onSubmit={(e) => { e.preventDefault(); handleSave(); }}>
        {/* API Endpoint */}
        <div className="mb-4">
          <label htmlFor="apiEndpoint" className="block text-sm font-medium text-gray-700 mb-1">
            {t('settings.apiEndpointLabel')}
          </label>
          <input
            type="text"
            id="apiEndpoint"
            value={apiEndpoint}
            onChange={(e) => setApiEndpoint(e.target.value)}
            placeholder={t('settings.apiEndpointPlaceholder')}
            className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm disabled:bg-gray-100"
            disabled={isLoading || isSaving}
          />
           <p className="mt-1 text-xs text-gray-500">{t('settings.apiEndpointHint')}</p>
        </div>

        {/* Default Model */}
        <div className="mb-4">
          <label htmlFor="defaultModel" className="block text-sm font-medium text-gray-700 mb-1">
            {t('settings.defaultModelLabel')}
          </label>
          <input
            type="text"
            id="defaultModel"
            value={defaultModel}
            onChange={(e) => setDefaultModel(e.target.value)}
            placeholder={t('settings.defaultModelPlaceholder')}
            className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm disabled:bg-gray-100"
            disabled={isLoading || isSaving}
          />
           <p className="mt-1 text-xs text-gray-500">{t('settings.defaultModelHint')}</p>
        </div>

        {/* API Key */}
        <div className="mb-6">
          <label htmlFor="apiKey" className="block text-sm font-medium text-gray-700 mb-1">
            {t('settings.apiKeyLabel')}
          </label>
          <input
            type="password"
            id="apiKey"
            value={apiKey}
            onChange={(e) => setApiKey(e.target.value)}
            placeholder={apiKeyIsSet ? t('settings.apiKeyPlaceholderUpdate') : t('settings.apiKeyPlaceholder')}
            className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm disabled:bg-gray-100"
            autoComplete="new-password" // Prevent browser autofill
            disabled={isLoading || isSaving}
          />
           <p className="mt-1 text-xs text-gray-500">
             {apiKeyIsSet ? t('settings.apiKeyHintUpdate') : t('settings.apiKeyHint')}
           </p>
        </div>

        {/* Status Messages */}
        <div className="mb-4 h-5"> {/* Reserve space for messages */}
          {isSaving && <p className="text-blue-600 text-sm">{t('settings.saving')}</p>}
          {error && <p className="text-red-600 text-sm">{error}</p>}
          {successMessage && <p className="text-green-600 text-sm">{successMessage}</p>}
        </div>

        {/* Action Buttons */}
        <div className="flex justify-end space-x-3">
          <button
            type="button"
            onClick={handleReset}
            disabled={isLoading || isSaving}
            className="px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {t('settings.resetButton')}
          </button>
          <button
            type="submit"
            disabled={isLoading || isSaving}
            className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isSaving ? t('settings.saving') : t('settings.saveButton')}
          </button>
        </div>
      </form>
    </div>
  );
};

export default SettingsPage;