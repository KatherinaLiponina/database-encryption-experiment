package encryptor

import (
	"cmd/internal/experiment"
	"fmt"
)

type Resolver interface {
	RegisterEncryptor(mode experiment.EncryptionMode, encryptor *Encryptor)
	GetEncryptor(mode experiment.EncryptionMode) (*Encryptor, error)
}

type resolver struct {
	encryptors map[experiment.EncryptionMode]*Encryptor
}

func NewResolver(encryptors map[experiment.EncryptionMode]*Encryptor) Resolver {
	return &resolver{encryptors: encryptors}
}

func (r *resolver) RegisterEncryptor(mode experiment.EncryptionMode, encryptor *Encryptor) {
	r.encryptors[mode] = encryptor
}

func (r *resolver) GetEncryptor(mode experiment.EncryptionMode) (*Encryptor, error) {
	if encryptor, ok := r.encryptors[mode]; ok {
		return encryptor, nil
	}
	return nil, fmt.Errorf("no such encryption mode")
}

type Encryptor interface {
	GetInfo() Info
	Encrypt(key any, data ...any) ([]any, error)
	Decrypt(key any, data ...any) ([]any, error)
	GenerateKey() (any, error)
}

type Info struct {
	Columns []Column
}

type Column struct {
	Name string
	Type experiment.Type
}

type KeyStorage struct {
	storage map[string]any
}

func NewKeyStorage() KeyStorage {
	return KeyStorage{storage: make(map[string]any)}
}

func (s *KeyStorage) Get(name string) (any, bool) {
	if key, ok := s.storage[name]; ok {
		return key, ok
	}
	return nil, false
}
func (s *KeyStorage) Add(name string, key any) {
	s.storage[name] = key
}
