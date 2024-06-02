package transformer

import (
	"cmd/internal/encryptor"
	"cmd/internal/experiment"
	queryconstructorv2 "cmd/internal/queryconstructor_v2"
	"fmt"
	"strconv"
	"strings"
)

type Query struct {
	resolver   encryptor.Resolver
	keyStorage encryptor.KeyStorage
}

func NewQuery(resolver encryptor.Resolver, keyStorage encryptor.KeyStorage) Query {
	return Query{resolver: resolver, keyStorage: keyStorage}
}

func (q *Query) Transform(qq queryconstructorv2.Query, cfg *experiment.Case) (queryconstructorv2.Query, error) {
	query := qq.GetQuery()
	args := qq.GetArguments()

	for _, tr := range cfg.Transforms {
		var err error
		query, args, err = q.transform(query, args, tr)
		if err != nil {
			return nil, err
		}
	}
	qq = queryconstructorv2.NewSelectQuery(query, args)
	return qq, nil
}

// TODO: replace with query parser
func (q *Query) transform(query string, args []any, tr experiment.Transform) (string, []any, error) {
	mode := strings.Split(tr.Transform, ":")
	if len(mode) < 2 {
		return "", nil, fmt.Errorf("syntax error in transformation: %s", tr.Transform)
	}
	if mode[0] != string(experiment.Encrypt) && mode[0] != string(experiment.Decrypt) {
		return "", nil, fmt.Errorf("unsupported transformation: %s", tr.Transform)
	}
	enc, err := q.resolver.GetEncryptor(experiment.EncryptionMode(mode[1]))
	if err != nil {
		return "", nil, err
	}

	if pos, ok := strings.CutPrefix(tr.Object, "$"); ok {
		position, err := strconv.Atoi(pos)
		if err != nil {
			return "", nil, err
		}
		if position >= len(args)+1 {
			return "", nil, fmt.Errorf("wrong position: %d on %v", position, args)
		}
		key, ok := q.keyStorage.Get(tr.Attribute)
		if !ok {
			return "", nil, fmt.Errorf("key not found for %s", tr.Attribute)
		}
		data, err := (*enc).Encrypt(key, []byte(fmt.Sprintf("%v", args[position-1])))
		if err != nil {
			return "", nil, err
		}
		args[position-1] = data[0]
		return query, args, nil
	}

	columns := (*enc).GetInfo().Columns
	replacement := ""
	for _, column := range columns {
		replacement += fmt.Sprintf("%s%s, ", tr.Object, column.Name)
	}
	replacement = replacement[:len(replacement)-2]

	position := strings.Index(query, tr.Object)
	if position == -1 {
		return query, args, nil
	}
	return query[:position] + replacement + query[position+len(tr.Object):], args, nil
}
