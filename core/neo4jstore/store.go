package neo4jstore

import (
	"asset-relations/core/aws"
	"asset-relations/support/config"
	"context"
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
	"log/slog"
	"strings"
)

type Neo4jDataStore struct {
	logger *slog.Logger
	driver neo4j.DriverWithContext
}

func NewNeo4jDataStore(ctx context.Context, logger *slog.Logger, cfg config.Neo4jConfig) (*Neo4jDataStore, error) {
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
const mergeInstanceQuery = `
	MERGE(n_POS_:Ec2Instance {id: $_POS_.id}) SET n_POS_ = {
		id: 				$_POS_.id, 
		isOpenToInternet: 	$_POS_.isOpenToInternet, 
		hasSSHPortOpen: 	$_POS_.hasSSHPortOpen, 
		SSHOpenToIps: 		$_POS_.SSHOpenToIps,
		hasRDPPortOpen: 	$_POS_.hasRDPPortOpen, 
		RDPOpenToIps: 		$_POS_.RDPOpenToIps, 
		VPCId: 				$_POS_.VPCId,
		openIngressPorts: 	$_POS_.openIngressPorts,
        openEgressPorts:	$_POS_.openEgressPorts,
		version: COALESCE(n_POS_.version, 0) + 1
	}
`

func (n *Neo4jDataStore) StoreInstances(ctx context.Context, instances []aws.Ec2Instance) error {
	n.logger.Info("Storing ec2 instances")
	b := strings.Builder{}
	params := make(map[string]any, len(instances))

	for idx, inst := range instances {
		pos := fmt.Sprintf("v%d", idx)
		params[pos] = map[string]any{
			"id":               inst.Id,
			"isOpenToInternet": inst.IsOpenToInternet(),
			"hasSSHPortOpen":   inst.HasSSHPortOpen(),
			"SSHOpenToIps":     inst.GetSSHOpenToIpRanges(),
			"hasRDPPortOpen":   inst.HasRDPPortOpen(),
			"RDPOpenToIps":     inst.GetRDPOpenToIpRanges(),
			"VPCId":            inst.VPC,
			"openIngressPorts": inst.GetOpenIngressPorts(),
			"openEgressPorts":  inst.GetOpenEgressPorts(),
		}

		b.WriteString(strings.ReplaceAll(mergeInstanceQuery, "_POS_", pos))
	}

	return n.write(ctx, b.String(), params)
}

// Use MERGE as create or update statement
const mergeVPCRelationQuery = `
	MATCH (from_POS_:Ec2Instance), (to_POS_:Ec2Instance)
	WHERE from_POS_.id = $fromId AND to_POS_.id = $toId 
	MERGE (from_POS_)-[r_POS_:IN_VPC {vpcId: $vpcId}]->(to_POS_) WITH r_POS_
	FINISH
`

func (n *Neo4jDataStore) StoreVPCRelatedInstances(ctx context.Context, groupedInstances map[string][]aws.Ec2Instance) error {
	n.logger.Info("Storing VPC related instances")

	queryParams := make(map[string]map[string]any, len(groupedInstances)*40)

	vpcIdAcc := 0
	for vpcId, instances := range groupedInstances {
		for fromIdx, fromInst := range instances {
			for toIdx, toInst := range instances {
				if fromInst.Id == toInst.Id {
					continue
				}

				pos := buildVPCArg(fromInst, toInst, vpcIdAcc, fromIdx, toIdx)
				query := strings.ReplaceAll(mergeVPCRelationQuery, "_POS_", pos)
				if _, exists := queryParams[query]; exists {
					continue
				}

				queryParams[query] = map[string]any{
					"fromId": fromInst.Id,
					"toId":   toInst.Id,
					"vpcId":  vpcId,
				}
			}
		}

		vpcIdAcc++
	}

	// Writing multiple times instead of one big query because it became too slow
	err := n.writeMultiple(ctx, queryParams)
	if err != nil {
		return err
	}

	n.logger.Info(fmt.Sprintf("Created %d relationships", len(queryParams)))

	return nil
}

func buildVPCArg(fromInst, toInst aws.Ec2Instance, vpcIdAcc, fromIdx, toIdx int) string {
	if fromInst.Id > toInst.Id {
		return fmt.Sprintf("v%d_%d_%d", vpcIdAcc, fromIdx, toIdx)
	} else {
		return fmt.Sprintf("v%d_%d_%d", vpcIdAcc, toIdx, fromIdx)
	}
}

const matchInstancesOpenToTheInternetQuery = `
	MATCH(n:Ec2Instance) 
	WHERE 
		n.isOpenToInternet = true 
		AND n.hasSSHPortOpen = true  
		AND n.SSHOpenToIps = '0.0.0.0/0' 
	RETURN(n)
`

func (n *Neo4jDataStore) GetInstancesWithOpenSSH(ctx context.Context) ([]map[string]any, error) {
	records, err := n.read(ctx, matchInstancesOpenToTheInternetQuery, nil)
	if err != nil {
		return nil, err
	}

	return extractPropsFromNodes(records, "n"), nil
}

const matchInstancesPartiallyOpenToTheInternetQuery = `
	MATCH(n:Ec2Instance) 
	WHERE 
		n.isOpenToInternet = true 
		AND n.hasSSHPortOpen = true  
		AND n.SSHOpenToIps <> '0.0.0.0/0' 
	RETURN(n)
`

func (n *Neo4jDataStore) GetInstancesWithPartiallyOpenSSH(ctx context.Context) ([]map[string]any, error) {
	records, err := n.read(ctx, matchInstancesPartiallyOpenToTheInternetQuery, nil)
	if err != nil {
		return nil, err
	}

	return extractPropsFromNodes(records, "n"), nil
}

const matchInstancesInVPCQuery = `
	MATCH(n:Ec2Instance)-[:IN_VPC]->(o:Ec2Instance)
	WHERE
		n.id = $id
	RETURN o
`

func (n *Neo4jDataStore) GetInstancesInVPC(ctx context.Context, id string) ([]map[string]any, error) {
	records, err := n.read(ctx, matchInstancesInVPCQuery, map[string]any{"id": id})
	if err != nil {
		return nil, err
	}

	return extractPropsFromNodes(records, "o"), nil
}

func extractPropsFromNodes(records []*neo4j.Record, variableName string) []map[string]any {
	response := make([]map[string]any, 0, len(records))
	for _, record := range records {
		props, exist := record.Get(variableName)
		if !exist {
			continue
		}

		response = append(response, props.(dbtype.Node).Props)
	}

	return response
}

func (n *Neo4jDataStore) Close(ctx context.Context) error {
	return n.driver.Close(ctx)
}

func (n *Neo4jDataStore) write(ctx context.Context, writeQuery string, params map[string]any) error {
	return n.writeMultiple(ctx, map[string]map[string]any{
		writeQuery: params,
	})
}

func (n *Neo4jDataStore) writeMultiple(ctx context.Context, queryParams map[string]map[string]any) error {
	session := n.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		for query, params := range queryParams {
			if _, err := tx.Run(ctx, query, params); err != nil {
				return nil, err
			}
		}

		return nil, nil
	})

	return err
}

func (n *Neo4jDataStore) read(ctx context.Context, readQuery string, params map[string]any) ([]*neo4j.Record, error) {
	session := n.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	records, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, readQuery, params)
		if err != nil {
			return nil, err
		}

		records, err := result.Collect(ctx)
		if err != nil {
			return nil, err
		}

		return records, nil
	})

	if err != nil {
		return nil, err
	}

	return records.([]*neo4j.Record), nil
}
