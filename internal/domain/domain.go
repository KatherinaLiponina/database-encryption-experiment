package domain

import (
	"cmd/internal/experiment"
	"strings"
)

type Case struct {
	Relations []Relation
}

type Relation struct {
	Name       string
	Size       uint64
	Attributes []Attribute
}

type Attribute struct {
	Name       string
	Type       experiment.Type
	Constraint string
	Encryption experiment.EncryptionMode
}

type EncryptorConfig map[string]map[string]experiment.EncryptionMode

func NewEncryptorConfig(cfg *experiment.Config) EncryptorConfig {
	encCfg := make(map[string]map[string]experiment.EncryptionMode)
	for _, enc := range cfg.Encryptions {
		rules := make(map[string]experiment.EncryptionMode)
		for _, r := range enc.Rules {
			rules[r.Attribute] = r.Encryption
		}
		encCfg[enc.Name] = rules
	}
	return encCfg
}

func NewCasesConfig(cfg *experiment.Config) []Case {
	encCfg := NewEncryptorConfig(cfg)

	var cases []Case
	for enc, rules := range encCfg {
		var domain Case
		for _, rel := range cfg.Relations {
			var relation = Relation{Name: rel.Name + "_" + enc, Size: uint64(rel.Size)}
			for _, attr := range rel.Attributes {
				var attribute = Attribute{Name: attr.Name, Type: attr.Type,
					Constraint: attr.Constraint}
				encryption, ok := rules[strings.ToLower(rel.Name+"."+attr.Name)]
				if !ok {
					encryption = experiment.None
				}
				attribute.Encryption = encryption
				relation.Attributes = append(relation.Attributes, attribute)
			}
			domain.Relations = append(domain.Relations, relation)
		}
		cases = append(cases, domain)
	}
	return cases
}
