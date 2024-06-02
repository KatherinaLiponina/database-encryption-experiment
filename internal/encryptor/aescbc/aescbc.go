package aescbc

import (
	"cmd/internal/encryptor"
	"cmd/internal/experiment"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

type aescbc_encryptor struct {
	ciphers map[string]*cipher.Block
}

func NewAES_CBC() encryptor.Encryptor {
	return &aescbc_encryptor{ciphers: make(map[string]*cipher.Block)}
}

func (e *aescbc_encryptor) GetInfo() encryptor.Info {
	return encryptor.Info{Columns: []encryptor.Column{
		{
			Name: "_enc",
			Type: experiment.ByteArray,
		},
	}}
}

func (e *aescbc_encryptor) Encrypt(key any, data ...any) ([]any, error) {
	formattedKey, src, err := e.checkArguments(key, data...)
	if err != nil {
		return nil, err
	}

	enc, err := e.getCipher(string(formattedKey))
	if err != nil {
		return nil, err
	}

	plaintext := src
	for len(plaintext)%aes.BlockSize != 0 {
		plaintext = append(plaintext, 0)
	}

	dst := make([]byte, aes.BlockSize)
	(*enc).Encrypt(dst, plaintext)

	return []any{dst}, nil
}

func (e *aescbc_encryptor) Decrypt(key any, data ...any) ([]any, error) {
	formattedKey, src, err := e.checkArguments(key, data...)
	if err != nil {
		return nil, err
	}

	enc, err := e.getCipher(string(formattedKey))
	if err != nil {
		return nil, err
	}

	dst := make([]byte, aes.BlockSize)
	(*enc).Decrypt(dst, src)

	// todo: remove zeroes in end

	return []any{dst}, nil
}

func (e *aescbc_encryptor) GenerateKey() (any, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	return key, nil
}

func (e *aescbc_encryptor) checkArguments(key any, data ...any) ([]byte, []byte, error) {
	if len(data) != 1 {
		return nil, nil, fmt.Errorf("wrong arguments: %v, expected two, got %d", data, len(data))
	}
	formattedKey, ok := key.([]byte)
	if !ok {
		return nil, nil, fmt.Errorf("cannot cast key to []byte: %v", key)
	}
	src, ok := data[0].([]byte)
	if !ok {
		return nil, nil, fmt.Errorf("cannot cast data to []byte: %v", data[0])
	}
	return formattedKey, src, nil
}

func (e *aescbc_encryptor) getCipher(key string) (*cipher.Block, error) {
	enc, ok := e.ciphers[string(key)]
	if ok {
		return enc, nil
	}
	// todo: check is cipher cbc or cbc
	enc_s, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}
	e.ciphers[string(key)] = &enc_s
	return &enc_s, nil
}
