package crypto

import (
	"bytes"
	"testing"
)

func TestCryptoService(t *testing.T) {
	key := "12345678901234567890123456789012" // 32 bytes
	service, err := New(key)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	plaintext := []byte("my secret password")

	ciphertext, err := service.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	if len(ciphertext) == 0 {
		t.Fatal("ciphertext is empty")
	}

	if bytes.Equal(plaintext, ciphertext) {
		t.Fatal("ciphertext equals plaintext")
	}

	decrypted, err := service.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Fatalf("expected %q, got %q", plaintext, decrypted)
	}
}

func TestCryptoService_InvalidKey(t *testing.T) {
	_, err := New("short")
	if err == nil {
		t.Fatal("expected error for short key")
	}
}
