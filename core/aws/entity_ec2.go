package aws

import (
	"asset-relations/support/ptr"
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
	trafficDirectionIngress trafficDirection = "INGRESS"
	trafficDirectionEgress  trafficDirection = "EGRESS"
)

const (
	sshPort int32 = 22
	rdpPort int32 = 3389
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
	_, isOpen := e.findRuleByPort(sshPort)
	return isOpen
}

func (e *Ec2Instance) GetSSHOpenToIpRanges() *string {
	rule, isOpen := e.findRuleByPort(sshPort)
	if !isOpen {
		return nil
	}

	return ptr.Ref(strings.Join(rule.IpRanges, ","))
}

func (e *Ec2Instance) HasRDPPortOpen() bool {
	_, isOpen := e.findRuleByPort(sshPort)
	return isOpen
}

func (e *Ec2Instance) GetRDPOpenToIpRanges() *string {
	rule, isOpen := e.findRuleByPort(sshPort)
	if !isOpen {
		return nil
	}

	return ptr.Ref(strings.Join(rule.IpRanges, ","))
}

func (e *Ec2Instance) findRuleByPort(port int32) (Ec2SecGroupRule, bool) {
	for _, rule := range e.SecurityGroupRules {
		if rule.TrafficDirection == trafficDirectionIngress && rule.FromPort <= port && rule.ToPort >= port {
			return rule, true
		}
	}

	return Ec2SecGroupRule{}, false
}
