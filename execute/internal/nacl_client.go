package internal

import "github.com/aws/aws-sdk-go/service/ec2"

// NACLClient is a API client for EC2 network ACL.
type NACLClient interface {
	ReleaseDenyingByNACL(networkACLID string, ruleNumber int64, ingressMode bool) error
	DenyByNACL(cidr string, protocolNumber int64, networkACLID string, ruleNumber int64, portRange *ec2.PortRange, ingressMode bool) error
	RetrieveNACLEntries(networkACLID string) (*ec2.DescribeNetworkAclsOutput, error)
}
