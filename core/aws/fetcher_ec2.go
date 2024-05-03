package aws

import (
	"asset-relations/support/parallel"
	"asset-relations/support/ptr"
	"context"
	"fmt"
	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"log/slog"
)

var ec2MaxResultsPerPage = int32(100)
var ec2SecurityGroupRulesMaxGoroutines = 50

type Ec2InstancesFetcher struct {
	client *ec2.Client
	logger *slog.Logger
}

func NewEc2InstanceFetcher(awsCfg awssdk.Config, logger *slog.Logger) Ec2InstancesFetcher {
	return Ec2InstancesFetcher{
		client: ec2.NewFromConfig(awsCfg),
		logger: logger,
	}
}

func (e *Ec2InstancesFetcher) Fetch(ctx context.Context) ([]Ec2Instance, error) {
	instances, err := e.fetchInstances(ctx)
	if err != nil {
		return nil, err
	}

	instances, err = parallel.Map(ctx, instances, e.enrichSecurityGroup, ec2SecurityGroupRulesMaxGoroutines)

	if err != nil {
		return nil, err
	}

	return instances, nil
}

func (e *Ec2InstancesFetcher) fetchInstances(ctx context.Context) ([]Ec2Instance, error) {
	e.logger.Info("Fetching EC2 instances")

	params := ec2.DescribeInstancesInput{MaxResults: &ec2MaxResultsPerPage}
	instances := make([]Ec2Instance, 0, ec2MaxResultsPerPage*10)

	for {
		res, err := e.client.DescribeInstances(ctx, &params)
		if err != nil {
			return nil, err
		}
		instances = append(instances, extractInstances(res)...)

		if ptr.IsEmpty(res.NextToken) {
			break
		}

		params.NextToken = res.NextToken
	}

	e.logger.Info(fmt.Sprintf("Fetched %d instances", len(instances)))

	return instances, nil
}

func extractInstances(res *ec2.DescribeInstancesOutput) []Ec2Instance {
	if res == nil {
		return nil
	}

	instances := make([]Ec2Instance, 0, ec2MaxResultsPerPage)
	for _, reservation := range res.Reservations {
		for _, instance := range reservation.Instances {
			instances = append(instances, convertInstance(instance))
		}
	}

	return instances
}

func convertInstance(instance ec2types.Instance) Ec2Instance {
	secGroupIds := make([]string, 0, len(instance.SecurityGroups))
	for _, secGroup := range instance.SecurityGroups {
		secGroupIds = append(secGroupIds, ptr.Deref(secGroup.GroupId))
	}

	return Ec2Instance{
		Id:               ptr.Deref(instance.InstanceId),
		PrivateIP:        ptr.Deref(instance.PrivateIpAddress),
		PrivateDNS:       ptr.Deref(instance.PrivateDnsName),
		PublicIP:         instance.PublicIpAddress,
		PublicDNS:        instance.PublicDnsName,
		VPC:              ptr.Deref(instance.VpcId),
		SSHKeyPairName:   instance.KeyName,
		SecurityGroupIds: secGroupIds,
	}
}

func (e *Ec2InstancesFetcher) enrichSecurityGroup(ctx context.Context, instance Ec2Instance) (Ec2Instance, error) {
	e.logger.Info("Fetching Security Group Rules for EC2 instance "+instance.Id, slog.Any("rule-ids", instance.SecurityGroupIds))

	rules := make([]Ec2SecGroupRule, 0, len(instance.SecurityGroupIds)*5)

	params := ec2.DescribeSecurityGroupsInput{
		GroupIds: instance.SecurityGroupIds,
	}

	for {
		res, err := e.client.DescribeSecurityGroups(ctx, &params)
		if err != nil {
			return instance, err
		}
		rules = append(rules, extractSecGroup(res)...)

		if ptr.IsEmpty(res.NextToken) {
			break
		}

		params.NextToken = res.NextToken
	}

	instance.SecurityGroupRules = rules
	return instance, nil
}

func extractSecGroup(res *ec2.DescribeSecurityGroupsOutput) []Ec2SecGroupRule {
	if res == nil {
		return nil
	}

	ipPermissions := make([]Ec2SecGroupRule, 0, ec2MaxResultsPerPage)
	for _, group := range res.SecurityGroups {
		for _, ipPermission := range append(group.IpPermissions) {
			ipPermissions = append(ipPermissions, convertSecurityGroup(ipPermission, trafficDirectionIngress))
		}
		for _, ipPermission := range append(group.IpPermissionsEgress) {
			ipPermissions = append(ipPermissions, convertSecurityGroup(ipPermission, trafficDirectionEgress))
		}

	}

	return ipPermissions
}

func convertSecurityGroup(ipPermission ec2types.IpPermission, trafficDirection trafficDirection) Ec2SecGroupRule {
	return Ec2SecGroupRule{
		FromPort:         ptr.Deref(ipPermission.FromPort),
		ToPort:           ptr.Deref(ipPermission.ToPort),
		IpProtocol:       ptr.Deref(ipPermission.IpProtocol),
		IpRanges:         extractIpRanges(ipPermission.IpRanges),
		TrafficDirection: trafficDirection,
	}
}

func extractIpRanges(ranges []ec2types.IpRange) []string {
	cidrs := make([]string, 0, len(ranges))
	for _, ipRange := range ranges {
		cidrs = append(cidrs, ptr.Deref(ipRange.CidrIp))
	}

	return cidrs
}
