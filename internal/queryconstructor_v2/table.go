package queryconstructorv2

import (
	"cmd/internal/domain"
	"cmd/internal/encryptor"
	"cmd/internal/experiment"
	"fmt"
	"strings"
)

type createTableBuilder struct {
	cfg     *domain.Case
	pointer int

	resolver encryptor.Resolver
}

func NewCreateTableBuilder(cfg *domain.Case, resolver encryptor.Resolver) QueryBuilder {
	return &createTableBuilder{cfg: cfg, resolver: resolver}
}

func (b *createTableBuilder) Next() (*Query, error) {
	if b.pointer >= len(b.cfg.Relations) {
		return nil, ErrNoMoreQueries
	}
	relation := b.cfg.Relations[b.pointer]
	b.pointer++

	var query string
	query = fmt.Sprintf("CREATE TABLE %s (\n", relation.Name)
	attributes := make([]string, 0, len(relation.Attributes))
	for _, attribute := range relation.Attributes {
		if attribute.Encryption == experiment.None || len(attribute.Encryption) == 0 {
			if attribute.Constraint == "" {
				attributes = append(attributes, fmt.Sprintf("\t%s %s", attribute.Name, attribute.Type))
			} else {
				attributes = append(attributes, fmt.Sprintf("\t%s %s %s", attribute.Name, attribute.Type, attribute.Constraint))
			}
			continue
		}

		encryptor, err := b.resolver.GetEncryptor(attribute.Encryption)
		if err != nil {
			return nil, err
		}
		columns := (*encryptor).GetInfo().Columns
		for _, column := range columns {
			attributes = append(attributes, fmt.Sprintf("\t%s%s %s", attribute.Name, column.Name, column.Type))
		}
	}

	query += strings.Join(attributes, ",\n") + "\n);"
	result := NewExecQuery(query)

	return &result, nil
}

type dropTableBuilder struct {
	cfg     *experiment.Config
	pointer int
}

func NewDropTableBuilder(cfg *experiment.Config) QueryBuilder {
	return &dropTableBuilder{cfg: cfg}
}

func (b *dropTableBuilder) Next() (*Query, error) {
	if b.pointer >= len(b.cfg.Relations) {
		return nil, ErrNoMoreQueries
	}
	relation := b.cfg.Relations[b.pointer]
	b.pointer++

	q := NewExecQuery(fmt.Sprintf("DROP TABLE %s;", relation.Name))
	return &q, nil
}
