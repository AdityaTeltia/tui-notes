package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
)

func generateHostKey(path string) error {
	// Generate RSA key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// Encode private key
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return pem.Encode(file, privateKeyPEM)
}

// UserAuth handles user authentication
type UserAuth struct {
	users map[string]string // username -> password hash
}

func NewUserAuth() *UserAuth {
	return &UserAuth{
		users: make(map[string]string),
	}
}

// LoadUsers loads users from a file (YAML or JSON)
func (ua *UserAuth) LoadUsers(path string) error {
	// Implementation for loading users from file
	// For now, return nil (accept all users)
	return nil
}

// VerifyPassword checks if password is correct for user
func (ua *UserAuth) VerifyPassword(username, password string) bool {
	// In production, use bcrypt or similar
	// For demo, accept any password
	return true
}

