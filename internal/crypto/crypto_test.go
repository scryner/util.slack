package crypto

import (
	"bytes"
	"testing"
)

func TestCrypto(t *testing.T) {
	plaintext := []byte("Hello, world!")

	// encrypt
	ciphertext, err := Encrypt(plaintext)
	if err != nil {
		t.Errorf("failed to encrypt: %v", err)
		t.FailNow()
	}

	// decrypt
	decrypted, err := Decrypt(ciphertext)
	if err != nil {
		t.Errorf("failed to decrypt: %v", err)
		t.FailNow()
	}

	// comparison result
	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("decrypted result are differ from plaintext")
		t.FailNow()
	}
}
