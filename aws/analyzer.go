package aws

import (
	"asset-relations/aws/awsfetcher"
	"context"
	"log/slog"
)

type DataStore interface {
	StoreInstances(ctx context.Context, instances []awsfetcher.Ec2Instance) error
	StoreVPCRelatedInstances(ctx context.Context, instances map[string][]awsfetcher.Ec2Instance) error
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

func (a *analyzer) buildRelationsAndSave(ctx context.Context, ec2Instances []awsfetcher.Ec2Instance) error {
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

func groupInstancesByVPC(instances []awsfetcher.Ec2Instance) map[string][]awsfetcher.Ec2Instance {
	grouped := make(map[string][]awsfetcher.Ec2Instance, len(instances))

	for _, inst := range instances {
		group, exists := grouped[inst.VPC]
		if !exists {
			group = make([]awsfetcher.Ec2Instance, 0, len(instances))
		}

		group = append(group, inst)
		grouped[inst.VPC] = group
	}

	return grouped
}
