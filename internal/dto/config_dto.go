package dto

// UserConfigDTO represents the user configuration data transferred to/from the API.
// It masks the actual API key, only indicating if it's set.
type UserConfigDTO struct {
	ApiEndpoint *string `json:"api_endpoint"` // Use pointers to distinguish between empty string and not set
	ModelName   *string `json:"model_name"`
	ApiKeyIsSet bool    `json:"api_key_is_set"` // Indicates if the user has set a specific API key
}

// UpdateUserConfigDTO represents the data received for updating user configuration.
// Fields are pointers to allow partial updates.
// ApiKey is received as plaintext and will be encrypted by the service.
// Sending an empty string "" for ApiKey means clearing the existing key.
// Sending nil means no change to the ApiKey.
type UpdateUserConfigDTO struct {
	ApiEndpoint *string `json:"api_endpoint"`
	ModelName   *string `json:"model_name"`
	ApiKey      *string `json:"api_key"` // Plaintext API key (or "" to clear, nil to leave unchanged)
}
