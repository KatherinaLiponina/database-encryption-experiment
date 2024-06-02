package executor

import (
	"cmd/internal/database"
	"cmd/internal/domain"
	"cmd/internal/encryptor"
	"cmd/internal/experiment"
	"cmd/internal/generator"
	queryconstructorv2 "cmd/internal/queryconstructor_v2"
	"cmd/internal/transformer"
	"cmd/internal/watcher"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
)

type Experiment interface {
	Prepare(*experiment.Config) error
	Start(*experiment.Config) (*Conclusion, error)
	CleanUp(cfg *experiment.Config) error
}

type Conclusion struct {
	MemoryTaken uint64
	Results     []*QueryConclusion
}

type QueryConclusion struct {
	Query string
	Time  struct {
		Prepare        time.Duration
		Exec           time.Duration
		PostProcessing time.Duration
	}
	// memory
}

func NewExperiment(logger *zap.SugaredLogger, conn database.Connection, resolver *encryptor.Resolver) Experiment {
	ks := encryptor.NewKeyStorage()
	return &executor{conn: conn, generator: generator.New(),
		resolver: resolver, keyStorage: &ks, logger: logger, samples: make(map[string]*[]any)}
}

type executor struct {
	logger *zap.SugaredLogger

	conn database.Connection

	generator  generator.RandomGenerator
	resolver   *encryptor.Resolver
	keyStorage *encryptor.KeyStorage

	samples map[string]*[]any
}

func (e *executor) Prepare(cfg *experiment.Config) error {
	fmt.Println("preparing...")
	defer fmt.Println("...ready")
	return e.prepare(cfg, &e.samples)
}

func (e *executor) Start(cfg *experiment.Config) (*Conclusion, error) {
	w := watcher.New()
	queryTransformer := transformer.NewQuery(*e.resolver, *e.keyStorage)
	encCfg := domain.NewEncryptorConfig(cfg)
	resultTransformer := transformer.NewResults(encCfg, *e.resolver, *e.keyStorage)

	f, err := os.Open("test")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	for i, query := range cfg.Queries {
		fmt.Println("Evaluating query: ", query.Origin)
		// prepare arguments
		args, err := getArguments(query.Args, e.samples, e.generator)
		if err != nil {
			return nil, err
		}

		for _, enc := range cfg.Encryptions {
			fmt.Println("Case: ", enc.Name)
			w.Start()

			q := replaceTable(query.Origin, cfg, enc.Name)
			qq, err := queryTransformer.Transform(queryconstructorv2.NewSelectQuery(q, args),
				&enc.Cases[i])
			if err != nil {
				return nil, err
			}

			rows, err := e.conn.Select(qq.GetQuery(), qq.GetArguments()...)
			if err != nil {
				return nil, err
			}
			if rows.Err() != nil {
				return nil, rows.Err()
			}
			defer rows.Close()

			columns, err := rows.Columns()
			if err != nil {
				return nil, err
			}
			// columnTypes, err := rows.ColumnTypes()
			// if err != nil {
			// 	return nil, err
			// }
			var data = make([]any, len(columns))
			var dataHelper = make([]any, len(columns))
			for i := range columns {
				//data[i] = reflect.New(columnTypes[i].ScanType()).Interface()

				dataHelper[i] = &data[i]
			}
			for rows.Next() {
				err = rows.Scan(dataHelper...)
				if err != nil {
					return nil, err
				}
				decryptedData, err := resultTransformer.Transform(data, &query, enc.Name)
				if err != nil {
					return nil, err
				}
				for _, r := range decryptedData {
					f.WriteString(stringify(r) + " ")
				}
			}
			f.WriteString("\n")

			report := w.Stop()
			fmt.Println("took", report.Time())
		}
	}

	return nil, nil
}

func (e *executor) prepare(cfg *experiment.Config, sample *map[string]*[]any) error {
	// create tables
	cases := domain.NewCasesConfig(cfg)
	for _, cs := range cases {
		builder := queryconstructorv2.NewCreateTableBuilder(&cs, *e.resolver)
		for {
			query, err := builder.Next()
			if err != nil {
				if !errors.Is(err, queryconstructorv2.ErrNoMoreQueries) {
					return fmt.Errorf("cannot build create table queries: %w", err)
				}
				break
			}
			err = e.conn.Exec((*query).GetQuery())
			if err != nil {
				return fmt.Errorf("create table failed with error: %w", err)
			}
		}
	}

	// insert data
	for i, relation := range cfg.Relations {
		valueStorage := make(map[string]*[]any)
		builder := queryconstructorv2.NewInsertBuilder(&relation, &e.generator, &valueStorage)

		var newQueries = make([]string, len(cases))

		for k := 0; k < relation.Size; k++ {
			// build common arguments
			query, err := builder.Next()
			if err != nil {
				return fmt.Errorf("cannot build inserts: %s", err.Error())
			}
			// prepare for every case
			for l, cs := range cases {
				rel := cs.Relations[i]
				args := (*query).GetArguments()
				var newArgs = make([]any, 0, len(args))
				for j := 0; j < len(rel.Attributes); j++ {
					enc, err := (*e.resolver).GetEncryptor(rel.Attributes[j].Encryption)
					if err != nil {
						newArgs = append(newArgs, args[j])
						continue
					}
					key, ok := e.keyStorage.Get(strings.ToLower(relation.Name + "." + rel.Attributes[j].Name))
					if !ok {
						key, err = (*enc).GenerateKey()
						if err != nil {
							return err
						}
						e.keyStorage.Add(strings.ToLower(relation.Name+"."+rel.Attributes[j].Name), key)
					}
					encrypted, err := (*enc).Encrypt(key, []byte(fmt.Sprintf("%v", args[j])))
					if err != nil {
						return err
					}
					newArgs = append(newArgs, encrypted...)
				}
				if newQueries[l] == "" {
					newQueries[l] = queryconstructorv2.BuildInsert(rel.Name, len(newArgs))
				}

				err = e.conn.Insert(newQueries[l], newArgs...)
				if err != nil {
					return fmt.Errorf("cannot insert row: error: %w, row: %v, query: %s", err, (*query).GetArguments(), (*query).GetQuery())
				}
			}
		}
		for key, value := range valueStorage {
			(*sample)[key] = value
		}
	}
	return nil
}

func getArguments(args []string, sample map[string]*[]any, generator generator.RandomGenerator) ([]any, error) {
	var arguments []any
	for _, a := range args {
		array := sample[a]
		if array == nil || len(*array) == 0 {
			return nil, fmt.Errorf("no values for attribute: %s", a)
		}
		arguments = append(arguments, (*array)[generator.GenerateRandomInt()%uint64(len(*array)/2)])
	}
	return arguments, nil
}

func stringify(a any) string {
	switch v := a.(type) {
	case []byte:
		return string(v)
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

// TODO: rewrite with query parser
func replaceTable(query string, cfg *experiment.Config, encName string) string {
	var table string
	var index int
	for _, relation := range cfg.Relations {
		if index = strings.Index(query, relation.Name); index != -1 {
			table = strings.ToLower(relation.Name)
			break
		}
	}
	if len(table) == 0 {
		fmt.Println("WARN: no replaced was made")
		return query
	}
	return query[:index] + table + "_" + encName + query[index+len(table):]
}

func (e *executor) CleanUp(cfg *experiment.Config) error {
	cases := domain.NewCasesConfig(cfg)
	for _, cs := range cases {
		for _, rel := range cs.Relations {
			err := e.conn.Exec(fmt.Sprintf("DROP TABLE %s;", rel.Name))
			if err != nil {
				return err
			}
		}
	}
	return nil
}
