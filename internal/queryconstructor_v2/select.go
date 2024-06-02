package queryconstructorv2

// import (
// 	"cmd/internal/encryptor"
// 	"cmd/internal/experiment"
// 	"fmt"
// 	"strconv"
// 	"strings"
// )

// type selectBuilder struct {
// 	cfg *experiment.Query

// 	query string
// 	args  []any

// 	resolver   encryptor.Resolver
// 	keyStorage *encryptor.KeyStorage
// }

// func NewSelectBuilder(cfg *experiment.Query, query string, args []any,
// 	resolver encryptor.Resolver, keyStorage *encryptor.KeyStorage) QueryBuilder {
// 	return &selectBuilder{cfg: cfg, query: query, args: args,
// 		resolver: resolver, keyStorage: keyStorage}
// }

// func (b *selectBuilder) Next() (*Query, error) {
// 	query := b.query
// 	args := b.args
// 	for _, tr := range b.cfg.Transformations {
// 		var err error
// 		query, args, err = b.transform(query, args, tr)
// 		if err != nil {
// 			return nil, err
// 		}
// 	}
// 	q := NewSelectQuery(query, args)
// 	return &q, nil
// }

// // TODO: replace with query parser
// func (b *selectBuilder) transform(query string, args []any, tr experiment.Transform) (string, []any, error) {
// 	mode := strings.Split(tr.Transform, ":")
// 	if len(mode) < 2 {
// 		return "", nil, fmt.Errorf("syntax error in transformation: %s", tr.Transform)
// 	}
// 	if mode[0] != string(experiment.Encrypt) && mode[0] != string(experiment.Decrypt) {
// 		return "", nil, fmt.Errorf("unsupported transformation: %s", tr.Transform)
// 	}
// 	enc, err := b.resolver.GetEncryptor(experiment.EncryptionMode(mode[1]))
// 	if err != nil {
// 		return "", nil, err
// 	}

// 	if pos, ok := strings.CutPrefix(tr.Object, "$"); ok {
// 		position, err := strconv.Atoi(pos)
// 		if err != nil {
// 			return "", nil, err
// 		}
// 		if position >= len(args)+1 {
// 			return "", nil, fmt.Errorf("wrong position: %d on %v", position, args)
// 		}
// 		key, ok := b.keyStorage.Get(tr.Attribute)
// 		if !ok {
// 			return "", nil, fmt.Errorf("key not found for %s", tr.Attribute)
// 		}
// 		data, err := (*enc).Encrypt(key, args[position-1])
// 		if err != nil {
// 			return "", nil, err
// 		}
// 		args[position-1] = data[0]
// 		return query, args, nil
// 	}

// 	columns := (*enc).GetInfo().Columns
// 	replacement := ""
// 	for _, column := range columns {
// 		replacement += fmt.Sprintf("%s%s, ", tr.Object, column.Name)
// 	}
// 	replacement = replacement[:len(replacement)-2]

// 	position := strings.Index(query, tr.Object)
// 	if position == -1 {
// 		return query, args, nil
// 	}
// 	return query[:position] + replacement + query[position+len(tr.Object):], args, nil
// }
