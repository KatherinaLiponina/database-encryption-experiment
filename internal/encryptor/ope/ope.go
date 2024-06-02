package ope

import (
	"cmd/internal/encryptor"
	"cmd/internal/experiment"
)

type ope_encryptor struct {
}

func NewOPE() encryptor.Encryptor {
	return &ope_encryptor{}
}

func (e *ope_encryptor) GetInfo() encryptor.Info {
	return encryptor.Info{
		Columns: []encryptor.Column{
			{
				Name: "_enc",
				Type: experiment.Integer,
			},
		},
	}
}
func (e *ope_encryptor) Encrypt(key any, data ...any) ([]any, error) {
	return []any{data[0]}, nil
}
func (e *ope_encryptor) Decrypt(key any, data ...any) ([]any, error) {
	return []any{data[0]}, nil
}
func (e *ope_encryptor) GenerateKey() (any, error) {
	return 10, nil
}
