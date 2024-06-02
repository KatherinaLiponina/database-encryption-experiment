package main

import (
	"cmd/internal/database"
	"cmd/internal/encryptor"
	"cmd/internal/encryptor/aescbc"
	"cmd/internal/encryptor/aesgcm"
	"cmd/internal/executor"
	"cmd/internal/experiment"
	"cmd/internal/parser"
	"fmt"
	"os"

	"go.uber.org/zap"
)

func main() {
	l, _ := zap.NewProduction()
	logger := l.Sugar()

	args := os.Args[1:]
	if len(args) == 0 {
		logger.Fatal("No config file provided!")
	}

	appConfig, err := NewConfigFromEnv()
	if err != nil {
		logger.Fatalf("Cannot get AppConfig: %w", err)
	}

	cfg, err := parser.ParseExperimentConfig(args[0])
	if err != nil {
		logger.Fatalf("Cannot parse config: %w", err)
	}

	connect, err := database.NewConnection(database.ConnectionConfig{
		Host:         appConfig.Database.Host,
		Port:         appConfig.Database.Port,
		User:         appConfig.Database.User,
		Password:     appConfig.Database.Password,
		DatabaseName: appConfig.Database.Name,
	})
	if err != nil {
		logger.Fatalf("Cannot establish database connection: %w", err)
	}
	_ = connect

	aes_cbc := aescbc.NewAES_CBC()
	aes_gcm := aesgcm.NewAES_GCM()
	resolver := encryptor.NewResolver(map[experiment.EncryptionMode]*encryptor.Encryptor{
		experiment.AES_CBC: &aes_cbc,
		experiment.AES_GCM: &aes_gcm,
	})

	expr := executor.NewExperiment(logger, connect, &resolver)
	err = expr.Prepare(cfg)
	if err != nil {
		expr.CleanUp(cfg)
		logger.Fatalf("Failed to prepare an experiment: %w", err)
	}
	conclusion, err := expr.Start(cfg)
	if err != nil {
		expr.CleanUp(cfg)
		logger.Fatalf("Experiment failed with error: %w", err)
	}
	fmt.Println(conclusion)
}
