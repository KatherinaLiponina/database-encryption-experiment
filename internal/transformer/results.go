package transformer

import (
	"cmd/internal/domain"
	"cmd/internal/encryptor"
	"cmd/internal/experiment"
	"fmt"
)

type Results struct {
	cfg domain.EncryptorConfig

	resolver   encryptor.Resolver
	keyStorage encryptor.KeyStorage
}

func NewResults(cfg domain.EncryptorConfig, resolver encryptor.Resolver,
	keyStorage encryptor.KeyStorage) Results {
	return Results{cfg: cfg, resolver: resolver, keyStorage: keyStorage}
}

func (r *Results) Transform(result []any, cfg *experiment.Query, cs string) ([]any, error) {
	index := 0
	var decrypted = make([]any, 0, len(result))
	for _, res := range cfg.Results {
		mp, ok := r.cfg[cs]
		if !ok {
			decrypted = append(decrypted, result[index])
			index++
			continue
		}
		enc, ok := mp[res]
		if !ok {
			decrypted = append(decrypted, result[index])
			index++
			continue
		}
		encrptr, err := r.resolver.GetEncryptor(enc)
		if err != nil {
			return nil, err
		}

		columns := (*encrptr).GetInfo().Columns
		var decryptData []any
		for range columns {
			decryptData = append(decryptData, result[index])
			index++
		}

		key, ok := r.keyStorage.Get(res)
		if !ok {
			return nil, fmt.Errorf("no key found for %s", res)
		}

		answer, err := (*encrptr).Decrypt(key, decryptData...)
		if err != nil {
			return nil, err
		}
		decrypted = append(decrypted, answer...)
	}
	return decrypted, nil
}
