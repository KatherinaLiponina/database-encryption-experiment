package parser

import (
	"cmd/internal/experiment"
	"encoding/json"
	"os"
)

func ParseExperimentConfig(file string) (*experiment.Config, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var config experiment.Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

