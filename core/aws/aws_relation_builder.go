package aws

import (
	"asset-relations/support/config"
	"context"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"log/slog"
)

type RelationBuilder struct {
	logger   *slog.Logger
	cfg      config.AwsConfig
	analyzer *analyzer
}

func NewRelationBuilder(logger *slog.Logger, cfg config.AwsConfig, store DataStore) *RelationBuilder {
	return &RelationBuilder{
		logger:   logger,
		cfg:      cfg,
		analyzer: newAnalyzer(logger, store),
	}
}

func (r *RelationBuilder) Build(ctx context.Context) error {
	r.logger.Info("Building Relationship")
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}

	awsCfg.Region = r.cfg.Region

	ec2F := NewEc2InstanceFetcher(awsCfg, r.logger)
	instances, err := ec2F.Fetch(ctx)
	if err != nil {
		return err
	}

	return r.analyzer.buildRelationsAndSave(ctx, instances)
}
