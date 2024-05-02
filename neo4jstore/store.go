package neo4jstore

import (
	"asset-relations/aws"
	"asset-relations/aws/awsfetcher"
	"asset-relations/config"
	"context"
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"log/slog"
	"strconv"
	"strings"
)

type Neo4jDataStore struct {
	logger *slog.Logger
	driver neo4j.DriverWithContext
}

func NewNeo4jDataStore(logger *slog.Logger, cfg config.Neo4jConfig) (aws.DataStore, error) {
	driver, err := neo4j.NewDriverWithContext(cfg.Uri, neo4j.BasicAuth(cfg.Username, cfg.Password, ""))
	if err != nil {
		return nil, fmt.Errorf("can't connect to Neo4j: %s", err.Error())
	}

	return &Neo4jDataStore{
		logger: logger,
		driver: driver,
	}, nil
}

const storeInstanceQuery = `
	MERGE(n_POS_:Ec2Instance {id: $_POS_.id}) SET n_POS_ = {
		id: $_POS_.id, 
		version: COALESCE(n_POS_.version, 0) + 1
	}
`

func (n *Neo4jDataStore) StoreInstances(ctx context.Context, instances []awsfetcher.Ec2Instance) error {
	n.logger.Info("Storing ec2 instances")
	b := strings.Builder{}
	params := make(map[string]any, len(instances))

	for idx, instance := range instances {
		pos := strconv.Itoa(idx)
		params[pos] = map[string]any{
			"id": instance.Id,
		}

		b.WriteString(strings.ReplaceAll(storeInstanceQuery, "_POS_", pos))
	}

	return n.write(ctx, b.String(), params)
}

func (n *Neo4jDataStore) StoreVPCRelatedInstances(ctx context.Context, instances map[string][]awsfetcher.Ec2Instance) error {
	return nil
}

func (n *Neo4jDataStore) Close(ctx context.Context) error {
	return n.driver.Close(ctx)
}
func (n *Neo4jDataStore) write(ctx context.Context, writeQuery string, params map[string]any) error {
	session := n.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		_, err := tx.Run(ctx, writeQuery, params)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})

	return err
}
