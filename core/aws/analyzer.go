package aws

import (
	"context"
	"log/slog"
)

type DataStore interface {
	StoreInstances(ctx context.Context, instances []Ec2Instance) error
	StoreVPCRelatedInstances(ctx context.Context, instances map[string][]Ec2Instance) error
}

type analyzer struct {
	logger *slog.Logger
	store  DataStore
}

func newAnalyzer(logger *slog.Logger, store DataStore) *analyzer {
	return &analyzer{
		logger: logger,
		store:  store,
	}
}

func (a *analyzer) buildRelationsAndSave(ctx context.Context, ec2Instances []Ec2Instance) error {
	err := a.store.StoreInstances(ctx, ec2Instances)
	if err != nil {
		return err
	}

	err = a.store.StoreVPCRelatedInstances(ctx, groupInstancesByVPC(ec2Instances))
	if err != nil {
		return err
	}

	return nil
}

func groupInstancesByVPC(instances []Ec2Instance) map[string][]Ec2Instance {
	grouped := make(map[string][]Ec2Instance, len(instances))

	for _, inst := range instances {
		group, exists := grouped[inst.VPC]
		if !exists {
			group = make([]Ec2Instance, 0, len(instances))
		}

		group = append(group, inst)
		grouped[inst.VPC] = group
	}

	return grouped
}
