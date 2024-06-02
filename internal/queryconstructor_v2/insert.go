package queryconstructorv2

import (
	"cmd/internal/experiment"
	"cmd/internal/generator"
	"fmt"
	"strings"
)

type insertBuilder struct {
	cfg *experiment.Relation

	query        string
	rowGenerator func() ([]any, error)

	generator *generator.RandomGenerator

	valueStorage *map[string]*[]any
}

func NewInsertBuilder(cfg *experiment.Relation, generator *generator.RandomGenerator,
	valueStorage *map[string]*[]any) QueryBuilder {
	return &insertBuilder{cfg: cfg, generator: generator, valueStorage: valueStorage}
}

func (b *insertBuilder) Next() (*Query, error) {
	if b.query != "" && b.rowGenerator != nil {
		args, err := b.rowGenerator()
		if err != nil {
			return nil, err
		}
		q := NewInsertQuery(b.query, args)
		return &q, nil
	}

	row, err := NewRowGenerator(b.cfg, b.generator, b.valueStorage)
	if err != nil {
		return nil, err
	}
	args, err := row.generate()
	if err != nil {
		return nil, err
	}

	columns := len(args)
	query := fmt.Sprintf("INSERT INTO %s VALUES\n", b.cfg.Name)
	placeholders := make([]string, 0, columns)
	for i := 0; i < columns; i++ {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
	}
	query += fmt.Sprintf("(%s);", strings.Join(placeholders, ", "))

	b.query = query
	b.rowGenerator = row.generate

	q := NewInsertQuery(query, args)
	return &q, nil
}

type row struct {
	rules     []rule
	queue     []func(r *rule, rg *rowGenerator)
	generator *generator.RandomGenerator
}

type rule struct {
	probability float64
	values      *[]any
}

func NewRowGenerator(cfg *experiment.Relation, generator *generator.RandomGenerator,
	valueStorage *map[string]*[]any) (row, error) {
	rules := make([]rule, 0, len(cfg.Attributes))
	queue := make([]func(r *rule, rg *rowGenerator), 0, len(cfg.Attributes))

	for i := 0; i < len(cfg.Attributes); i++ {
		attribute := cfg.Attributes[i]

		// set function for attribute generation
		switch attribute.Generation {
		case experiment.Probabilistic:
			queue = append(queue, func(r *rule, rg *rowGenerator) {
				rg.randomProbabilistic(attribute.Type, &r.probability, r.values)
			})
		case experiment.Unique:
			queue = append(queue, func(r *rule, rg *rowGenerator) {
				rg.randomUnique(attribute.Type, &r.probability, r.values)
			})
		case experiment.FromValues:
			queue = append(queue, func(r *rule, rg *rowGenerator) {
				rg.fromValues(*r.values)
			})
		default:
			return row{}, fmt.Errorf("unsupported generation type: %s", attribute.Generation)
		}

		// set a rule for generation
		if len(attribute.Values) != 0 {
			rules = append(rules, rule{probability: 1, values: &attribute.Values})
			(*valueStorage)[strings.ToLower(cfg.Name+"."+attribute.Name)] = &attribute.Values
		} else {
			var values = make([]any, 0)
			rules = append(rules, rule{probability: 1, values: &values})
			(*valueStorage)[strings.ToLower(cfg.Name+"."+attribute.Name)] = &values
		}

		// // set function for encryption if applicable
		// if attribute.Encryption == experiment.None {
		// 	continue
		// }
		// encryptor, err := resolver.GetEncryptor(attribute.Encryption)
		// if err != nil {
		// 	return row{}, fmt.Errorf("get encryptor failed: %w", err)
		// }
		// fullName := strings.ToLower(cfg.Name + "." + attribute.Name)
		// key, ok := keyStorage.Get(fullName)
		// if !ok {
		// 	key, err = (*encryptor).GenerateKey()
		// 	if err != nil {
		// 		return row{}, err
		// 	}
		// 	keyStorage.Add(fullName, key)
		// }
		// queue = append(queue, func(r *rule, rg *rowGenerator) {
		// 	rg.encryptLast(encryptor, key)
		// })
		// // set empty rule for encryptor
		// rules = append(rules, rule{})
	}

	return row{queue: queue, rules: rules, generator: generator}, nil
}

func (r *row) generate() ([]any, error) {
	rg := newRowGenerator(r.generator)
	for i := 0; i < len(r.queue); i++ {
		r.queue[i](&r.rules[i], rg)
	}
	return rg.commit()
}

// row generator is a struct which collects arguments until all rules are applied
// when ready, commit should be called to get results and possibly occured error
type rowGenerator struct {
	output    []any
	generator *generator.RandomGenerator

	err error
}

func newRowGenerator(generator *generator.RandomGenerator) *rowGenerator {
	return &rowGenerator{output: make([]any, 0), generator: generator}
}

func (g *rowGenerator) randomProbabilistic(t experiment.Type, probability *float64, values *[]any) {
	coin := float64(g.generator.GenerateRandomInt()%100) / 100
	var value any
	if coin <= *probability {
		*probability = *probability / 2
		value = g.generator.RandomByType(t)
		*values = append(*values, value)
	} else {
		value = (*values)[int(g.generator.GenerateRandomInt())%len(*values)]
	}
	g.output = append(g.output, value)
}

func (g *rowGenerator) randomUnique(t experiment.Type, probability *float64, values *[]any) {
	value := g.generator.RandomByType(t)
	coin := float64(g.generator.GenerateRandomInt()%100) / 100
	if coin <= *probability {
		*probability = *probability / 2
		*values = append(*values, value)
	}
	g.output = append(g.output, value)
}

func (g *rowGenerator) fromValues(values []any) {
	g.output = append(g.output, (values)[int(g.generator.GenerateRandomInt())%len(values)])
}

// func (g *rowGenerator) encryptLast(enc *encryptor.Encryptor, key any) {
// 	last := g.output[len(g.output)-1]
// 	g.output = g.output[:len(g.output)-1]

// 	data, err := (*enc).Encrypt(key, fmt.Sprintf("%v", last))
// 	g.output = append(g.output, data...)
// 	if err != nil {
// 		g.err = err
// 	}
// }

func (g *rowGenerator) commit() ([]any, error) {
	output := g.output
	g.output = []any{}
	err := g.err
	g.err = nil
	return output, err
}

func BuildInsert(table string, args int) string {
	query := fmt.Sprintf("INSERT INTO %s VALUES\n", table)
	placeholders := make([]string, 0, args)
	for i := 0; i < args; i++ {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
	}
	query += fmt.Sprintf("(%s);", strings.Join(placeholders, ", "))
	return query
}
