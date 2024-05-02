package main

import (
	"asset-relations/aws"
	"asset-relations/config"
	"asset-relations/neo4jstore"
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"
)

func main() {
	start := time.Now()

	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg, err := config.LoadConfig(logger)
	if err != nil {
		logger.Error("Couldn't initialize configuration: " + err.Error())
		return
	}

	logger = logger.With(slog.String("region", cfg.Aws.Region))

	store, err := neo4jstore.NewNeo4jDataStore(ctx, logger, cfg.Neo4j)
	if err != nil {
		logger.Error("Couldn't initialize Neo4j Data Store: " + err.Error())
		return
	}

	defer store.(*neo4jstore.Neo4jDataStore).Close(ctx)

	builder := aws.NewRelationBuilder(logger, cfg.Aws, store)

	if err := builder.Build(ctx); err != nil {
		logger.Error("Relation Builder exited with an error: " + err.Error())
		return
	}

	elapsed := time.Since(start)
	logger.Info(fmt.Sprintf("Elapsed time %s", elapsed))
}
