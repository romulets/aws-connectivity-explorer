package aws

import (
	"asset-relations/support/ptr"
	"strings"
)

type Ec2Instance struct {
	Id               string
	PrivateIP        string
	PublicIP         *string
	PublicDNS        *string
	VPC              string
	SecurityGroupIds []string
	EgressSecRules   []Ec2SecGroupRule
	IngressSecRules  []Ec2SecGroupRule
	PrivateDNS       string
	SSHKeyPairName   *string
}

const (
	sshPort int32 = 22
	rdpPort int32 = 3389
)

type Ec2SecGroupRule struct {
	FromPort   int32
	ToPort     int32
	IpProtocol string
	IpRanges   []string
}

func (e *Ec2Instance) IsOpenToInternet() bool {
	return ptr.IsEmpty(e.PublicIP)
}

func (e *Ec2Instance) HasSSHPortOpen() bool {
	_, isOpen := e.findIngressRules(sshPort)
	return isOpen
}

func (e *Ec2Instance) GetOpenIngressPorts() []int32 {
	return extractPorts(e.IngressSecRules)
}

func (e *Ec2Instance) GetOpenEgressPorts() []int32 {
	return extractPorts(e.EgressSecRules)
}

func extractPorts(rules []Ec2SecGroupRule) []int32 {
	ports := make([]int32, 0, len(rules))
	for _, rule := range rules {
		if rule.FromPort == rule.ToPort && rule.FromPort > 1 {
			ports = append(ports, rule.FromPort)
			continue
		}

		for port := rule.FromPort; port <= rule.ToPort; port++ {
			if port < 1 {
				continue
			}

			ports = append(ports, port)
		}
	}

	return ports
}

func (e *Ec2Instance) GetSSHOpenToIpRanges() *string {
	rule, isOpen := e.findIngressRules(sshPort)
	if !isOpen {
		return nil
	}

	return ptr.Ref(strings.Join(rule.IpRanges, ","))
}

func (e *Ec2Instance) HasRDPPortOpen() bool {
	_, isOpen := e.findIngressRules(rdpPort)
	return isOpen
}

func (e *Ec2Instance) GetRDPOpenToIpRanges() *string {
	rule, isOpen := e.findIngressRules(rdpPort)
	if !isOpen {
		return nil
	}

	return ptr.Ref(strings.Join(rule.IpRanges, ","))
}

func (e *Ec2Instance) findIngressRules(port int32) (Ec2SecGroupRule, bool) {
	for _, rule := range e.IngressSecRules {
		if rule.FromPort <= port && rule.ToPort >= port {
			return rule, true
		}
	}

	return Ec2SecGroupRule{}, false
}
