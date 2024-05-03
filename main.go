package main

import (
	"asset-relations/application/controller"
	"asset-relations/application/http"
	"asset-relations/core/aws"
	"asset-relations/core/neo4jstore"
	"asset-relations/support/config"
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

	store, err := neo4jstore.NewNeo4jDataStore(ctx, logger, cfg.Neo4j)
	if err != nil {
		logger.Error("Couldn't initialize Neo4j Data Store: " + err.Error())
		return
	}

	defer store.Close(ctx)

	builder := aws.NewRelationBuilder(logger, cfg.Aws, store)
	ec2Controller := controller.NewEc2Controller(logger, store, builder)
	server := http.NewServer(ec2Controller, logger, cfg.Http)

	server.ListenAndServe()
}
