package aesgcm

import (
	"cmd/internal/encryptor"
	"cmd/internal/experiment"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

type aesgcm_encryptor struct {
	ciphers map[string]*cipher.AEAD
}

func NewAES_GCM() encryptor.Encryptor {
	return &aesgcm_encryptor{ciphers: make(map[string]*cipher.AEAD)}
}

func (e *aesgcm_encryptor) GetInfo() encryptor.Info {
	return encryptor.Info{Columns: []encryptor.Column{
		{
			Name: "_enc",
			Type: experiment.ByteArray,
		},
		{
			Name: "_nonce",
			Type: experiment.ByteArray,
		},
	}}
}

func (e *aesgcm_encryptor) Encrypt(key any, data ...any) ([]any, error) {
	if len(data) != 1 {
		return nil, fmt.Errorf("wrong arguments: %v, expected two, got %d", data, len(data))
	}
	keyFormatted, ok := key.([]byte)
	if !ok {
		return nil, fmt.Errorf("cannot cast key to []byte: %v", key)
	}
	src, ok := data[0].([]byte)
	if !ok {
		return nil, fmt.Errorf("cannot cast data to []byte: %v", data[0])
	}

	enc, err := e.getCipher(string(keyFormatted))
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ciphertext := (*enc).Seal(nil, nonce, src, nil)

	return []any{ciphertext, nonce}, nil
}

func (e *aesgcm_encryptor) Decrypt(key any, data ...any) ([]any, error) {
	if len(data) != 2 {
		return nil, fmt.Errorf("wrong arguments: %v, expected two, got %d", data, len(data))
	}
	keyFormatted, ok := key.([]byte)
	if !ok {
		return nil, fmt.Errorf("cannot cast key to []byte: %v", key)
	}
	ciphertext, ok := data[0].([]byte)
	if !ok {
		return nil, fmt.Errorf("cannot cast data to []byte: %v", data[0])
	}
	nonce, ok := data[1].([]byte)
	if !ok {
		return nil, fmt.Errorf("cannot cast nonce to []byte: %v", data[1])
	}

	enc, err := e.getCipher(string(keyFormatted))
	if err != nil {
		return nil, err
	}

	plaintext, err := (*enc).Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return []any{plaintext}, nil
}

func (e *aesgcm_encryptor) GenerateKey() (any, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	return key, nil
}

func (e *aesgcm_encryptor) getCipher(key string) (*cipher.AEAD, error) {
	if enc, ok := e.ciphers[key]; ok {
		return enc, nil
	}
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	e.ciphers[key] = &aesgcm
	return &aesgcm, nil
}
