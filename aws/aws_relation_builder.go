package aws

import (
	"asset-relations/aws/awsfetcher"
	"asset-relations/config"
	"context"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"log/slog"
)

func BuildRelation(ctx context.Context, logger *slog.Logger, cfg *config.Config) error {
	logger.Info("Building Relationship")
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}

	awsCfg.Region = cfg.Aws.Region

	ec2F := awsfetcher.NewEc2InstanceFetcher(awsCfg, logger)
	instances, err := ec2F.Fetch(ctx)
	if err != nil {
		return err
	}

	a := newAnalyzer(logger)
	a.analyze(instances)

	return nil
}
