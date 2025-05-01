package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256" // Import sha256 for key derivation
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

// deriveKey generates a 32-byte key suitable for AES-256 from the input secret.
// It uses SHA-256 to hash the secret.
func deriveKey(secret string) []byte {
	hash := sha256.Sum256([]byte(secret))
	return hash[:] // Use the 32-byte hash directly
}

// EncryptString encrypts a string using AES-GCM with the provided secret.
// The secret is first derived into a 32-byte key using SHA-256.
// The returned byte slice contains the nonce prefixed to the ciphertext.
func EncryptString(plaintext string, secret string) ([]byte, error) {
	key := deriveKey(secret)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM cipher: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Seal encrypts and authenticates plaintext, authenticates the
	// additional data and appends the result to dst, returning the updated
	// slice. The nonce must be NonceSize() bytes long and unique for all
	// time, for a given key.
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return ciphertext, nil
}

// DecryptString decrypts a byte slice (nonce + ciphertext) using AES-GCM with the provided secret.
// The secret is first derived into a 32-byte key using SHA-256.
func DecryptString(ciphertext []byte, secret string) (string, error) {
	key := deriveKey(secret)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM cipher: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, encryptedMessage := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Open decrypts and authenticates ciphertext, authenticates the
	// additional data and, if successful, appends the resulting plaintext
	// to dst, returning the updated slice. The nonce must be NonceSize()
	// bytes long and both it and the additional data must match the
	// value passed to Seal.
	plaintext, err := gcm.Open(nil, nonce, encryptedMessage, nil)
	if err != nil {
		// It's important not to reveal specific crypto errors to the outside world usually,
		// but for internal debugging, logging the error might be useful.
		// For user-facing errors, a generic "decryption failed" is often better.
		return "", fmt.Errorf("failed to decrypt data: %w", err)
	}

	return string(plaintext), nil
}

// EncryptStringHex is a convenience function that encrypts and returns hex-encoded string.
func EncryptStringHex(plaintext string, secret string) (string, error) {
	encryptedBytes, err := EncryptString(plaintext, secret)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(encryptedBytes), nil
}

// DecryptStringHex is a convenience function that decrypts a hex-encoded string.
func DecryptStringHex(encryptedHex string, secret string) (string, error) {
	encryptedBytes, err := hex.DecodeString(encryptedHex)
	if err != nil {
		return "", fmt.Errorf("failed to decode hex string: %w", err)
	}
	return DecryptString(encryptedBytes, secret)
}
