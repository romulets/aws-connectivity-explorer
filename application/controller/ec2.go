package controller

import (
	"asset-relations/core/aws"
	"asset-relations/core/neo4jstore"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
)

type Ec2Controller struct {
	logger          *slog.Logger
	store           *neo4jstore.Neo4jDataStore
	relationBuilder *aws.RelationBuilder
}

func NewEc2Controller(logger *slog.Logger, store *neo4jstore.Neo4jDataStore, relationBuilder *aws.RelationBuilder) *Ec2Controller {
	return &Ec2Controller{
		logger:          logger,
		store:           store,
		relationBuilder: relationBuilder,
	}
}

func (e *Ec2Controller) GetInstancesOpenToInternet(ctx context.Context) JSONResponse {
	instances, err := e.store.GetInstancesOpenToTheInternet(ctx)
	if err != nil {
		e.logger.Error("Couldn't get Instances open to internet: " + err.Error())
		msg := []byte(fmt.Sprintf(`{"error":"%s"}`, err.Error()))
		return jsonRes(500, msg)
	}

	data, err := json.Marshal(instances)
	if err != nil {
		e.logger.Error("Couldn't convert Instances to json: " + err.Error())
		msg := []byte(fmt.Sprintf(`{"error":"%s"}`, err.Error()))
		return jsonRes(500, msg)
	}

	return jsonRes(200, data)
}

func (e *Ec2Controller) FetchInstancesGraph(ctx context.Context) JSONResponse {
	start := time.Now()
	defer func() {
		elapsed := time.Since(start)
		e.logger.Info(fmt.Sprintf("Elapsed time %s", elapsed))
	}()

	if err := e.relationBuilder.Build(ctx); err != nil {
		e.logger.Error("Relation Builder exited with an error: " + err.Error())
		msg := []byte(fmt.Sprintf(`{"error":"%s"}`, err.Error()))
		return jsonRes(500, msg)
	}

	return jsonRes(202, nil)
}
