package main

import (
	"asset-relations/aws"
	"asset-relations/config"
	"context"
	"log/slog"
	"os"
)

func main() {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg, err := config.LoadConfig(logger)
	if err != nil {
		logger.Error("Couldn't initialize configuration: " + err.Error())
		return
	}

	logger = logger.With(slog.String("region", cfg.Aws.Region))

	if err := aws.BuildRelation(ctx, logger, cfg); err != nil {
		logger.Error("Relation Builder exited with an error: " + err.Error())
		return
	}
}
