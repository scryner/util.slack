package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

var (
	block cipher.Block
)

func init() {
	// make random AES-256 key
	var k [32]byte
	_, err := rand.Read(k[:])
	if err != nil {
		// never reached
		panic(fmt.Sprintf("failed to make random crypto(AES-256) key: %v", err))
	}

	// make block
	block, err = aes.NewCipher(k[:])
	if err != nil {
		// never reached
		panic(fmt.Sprintf("failed to make AES-256 cipher block: %v", err))
	}
}

func Encrypt(plain []byte) ([]byte, error) {
	// make stream
	var iv [aes.BlockSize]byte
	stream := cipher.NewOFB(block, iv[:])

	// make stream writer
	out := new(bytes.Buffer)
	writer := &cipher.StreamWriter{S: stream, W: out}

	// write plaintext to stream writer
	if _, err := io.Copy(writer, bytes.NewReader(plain)); err != nil {
		return nil, fmt.Errorf("failed to encrypt: %v", err)
	}

	return out.Bytes(), nil
}

func Decrypt(ciphertext []byte) ([]byte, error) {
	// make stream
	var iv [aes.BlockSize]byte
	stream := cipher.NewOFB(block, iv[:])

	// make stream reader
	reader := &cipher.StreamReader{S: stream, R: bytes.NewReader(ciphertext)}

	// read ciphertext from stream reader
	out := new(bytes.Buffer)
	if _, err := io.Copy(out, reader); err != nil {
		return nil, fmt.Errorf("failed to decrypt: %v", err)
	}

	return out.Bytes(), nil
}
