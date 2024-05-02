package neo4jstore

import (
	"asset-relations/aws"
	"asset-relations/aws/awsfetcher"
	"asset-relations/config"
	"context"
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"log/slog"
	"strings"
)

type Neo4jDataStore struct {
	logger *slog.Logger
	driver neo4j.DriverWithContext
}

func NewNeo4jDataStore(ctx context.Context, logger *slog.Logger, cfg config.Neo4jConfig) (aws.DataStore, error) {
	driver, err := neo4j.NewDriverWithContext(cfg.Uri, neo4j.BasicAuth(cfg.Username, cfg.Password, ""))
	if err != nil {
		return nil, fmt.Errorf("can't connect to Neo4j: %s", err.Error())
	}

	store := &Neo4jDataStore{
		logger: logger,
		driver: driver,
	}

	if err := store.initDB(ctx); err != nil {
		return nil, err
	}

	return store, nil
}

const initQuery = `
	CREATE INDEX ec2Id IF NOT EXISTS FOR (n:Ec2Instance) ON (n.id)
`

func (n *Neo4jDataStore) initDB(ctx context.Context) error {
	n.logger.Info("Initializing DB")
	return n.write(ctx, initQuery, nil)
}

// Use MERGE as create or update statement
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
		pos := fmt.Sprintf("v%d", idx)
		params[pos] = map[string]any{
			"id": instance.Id,
		}

		b.WriteString(strings.ReplaceAll(storeInstanceQuery, "_POS_", pos))
	}

	return n.write(ctx, b.String(), params)
}

// Use MERGE as create or update statement
const storeVPCRelationQuery = `
	MATCH (from_POS_:Ec2Instance {id: $fromId})
	MATCH (to_POS_:Ec2Instance {id: $toId})
	MERGE (from_POS_)-[r_POS_:IN_VPC {vpcId: $vpcId}]->(to_POS_) WITH r_POS_
	FINISH
`

func (n *Neo4jDataStore) StoreVPCRelatedInstances(ctx context.Context, groupedInstances map[string][]awsfetcher.Ec2Instance) error {
	n.logger.Info("Storing VPC related instances")

	params := make(map[string]any, len(groupedInstances)*40)

	vpcIdAcc := 0
	for vpcId, instances := range groupedInstances {
		for fromIdx, fromInst := range instances {
			for toIdx, toInst := range instances {
				if fromInst.Id == toInst.Id {
					continue
				}

				pos := buildVPCArg(fromInst, toInst, vpcIdAcc, fromIdx, toIdx)
				if _, exists := params[pos]; exists {
					continue
				}

				params[pos] = map[string]any{
					"fromId": fromInst.Id,
					"toId":   toInst.Id,
					"vpcId":  vpcId,
				}

				// Writing multiple times instead of one big query because it became too slow
				err := n.write(ctx, strings.ReplaceAll(storeVPCRelationQuery, "_POS_", pos), params[pos].(map[string]any))
				if err != nil {
					return err
				}
			}
		}

		vpcIdAcc++
	}

	n.logger.Info(fmt.Sprintf("Created %d relationships", len(params)))

	return nil
}

func buildVPCArg(fromInst, toInst awsfetcher.Ec2Instance, vpcIdAcc, fromIdx, toIdx int) string {
	if fromInst.Id > toInst.Id {
		return fmt.Sprintf("v%d_%d_%d", vpcIdAcc, fromIdx, toIdx)
	} else {
		return fmt.Sprintf("v%d_%d_%d", vpcIdAcc, toIdx, fromIdx)
	}
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
