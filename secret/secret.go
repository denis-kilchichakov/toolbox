package secret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
)

type WrappedSecret string
type UnwrappedSecret string

func WrapSecret(secret string, masterKey string) (WrappedSecret, error) {
	block, err := aes.NewCipher([]byte(masterKey))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(secret), nil)
	return WrappedSecret(hex.EncodeToString(ciphertext)), nil
}

func UnwrapSecret(wrappedSecret WrappedSecret, masterKey string) (UnwrappedSecret, error) {
	block, err := aes.NewCipher([]byte(masterKey))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	ciphertext, err := hex.DecodeString(string(wrappedSecret))
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return UnwrappedSecret(plaintext), nil
}
