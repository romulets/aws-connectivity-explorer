package aws

import (
	"asset-relations/ptr"
	"strings"
)

type Ec2Instance struct {
	Id                 string
	PrivateIP          string
	PublicIP           *string
	PublicDNS          *string
	VPC                string
	SecurityGroupIds   []string
	SecurityGroupRules []Ec2SecGroupRule
	PrivateDNS         string
	SSHKeyPairName     *string
}

type trafficDirection string

const (
	trafficDirectionIngress = "INGRESS"
	trafficDirectionEgress  = "EGRESS"
)

type Ec2SecGroupRule struct {
	FromPort         int32
	ToPort           int32
	IpProtocol       string
	IpRanges         []string
	TrafficDirection trafficDirection
}

func (e *Ec2Instance) IsOpenToInternet() bool {
	return ptr.IsEmpty(e.PublicIP)
}

func (e *Ec2Instance) HasSSHPortOpen() bool {
	_, isOpen := e.findSSHRule()
	return isOpen
}

func (e *Ec2Instance) GetSSHOpenToIpRanges() *string {
	rule, isOpen := e.findSSHRule()
	if !isOpen {
		return nil
	}

	return ptr.Ref(strings.Join(rule.IpRanges, ","))
}

func (e *Ec2Instance) findSSHRule() (Ec2SecGroupRule, bool) {
	for _, rule := range e.SecurityGroupRules {
		if rule.TrafficDirection == trafficDirectionIngress && rule.FromPort <= 22 && rule.ToPort >= 22 {
			return rule, true
		}
	}

	return Ec2SecGroupRule{}, false
}
