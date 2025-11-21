package models

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

var encryptionKey []byte // In production, derive from user password

func initEncryption(key []byte) {
	encryptionKey = key
}

func encryptNote(content string) (string, error) {
	if len(encryptionKey) == 0 {
		return content, errors.New("encryption key not set")
	}
	
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}
	
	// Create GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	
	// Create nonce
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	
	// Encrypt
	ciphertext := aesGCM.Seal(nonce, nonce, []byte(content), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func decryptNote(encryptedContent string) (string, error) {
	if len(encryptionKey) == 0 {
		return encryptedContent, errors.New("encryption key not set")
	}
	
	// Decode base64
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedContent)
	if err != nil {
		return "", err
	}
	
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}
	
	// Create GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	
	// Extract nonce
	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	
	// Decrypt
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	
	return string(plaintext), nil
}

