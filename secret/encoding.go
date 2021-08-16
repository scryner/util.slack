package secret

import (
	"encoding/base64"

	"github.com/scryner/util.slack/internal/crypto"
)

func Encode(b []byte) (string, error) {
	// encrypt
	encrypted, err := crypto.Encrypt(b)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func Decode(s string) ([]byte, error) {
	if s != "" {
		// base64 decoding
		ciphertext, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return nil, err
		}

		// decrypt
		decrypted, err := crypto.Decrypt([]byte(ciphertext))
		if err != nil {
			return nil, err
		}

		return decrypted, nil
	}

	return nil, nil
}
