package internal

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type NACLClientImpl struct {
	EC2Srv *ec2.EC2
}

func NewNACLClientImpl(ec2Srv *ec2.EC2) *NACLClientImpl {
	return &NACLClientImpl{
		EC2Srv: ec2Srv,
	}
}

func (n *NACLClientImpl) ReleaseDenyingByNACL(networkACLID string, ruleNumber int64, ingressMode bool) error {
	_, err := n.EC2Srv.DeleteNetworkAclEntry(&ec2.DeleteNetworkAclEntryInput{
		Egress:       aws.Bool(ingressMode),
		NetworkAclId: aws.String(networkACLID),
		RuleNumber:   aws.Int64(ruleNumber),
	})
	if err != nil {
		return err
	}
	return nil
}

func (n *NACLClientImpl) DenyByNACL(cidr string, protocolNumber int64, networkACLID string, ruleNumber int64, portRange *ec2.PortRange, ingressMode bool) error {
	// TODO IPv6 supporting
	_, err := n.EC2Srv.CreateNetworkAclEntry(&ec2.CreateNetworkAclEntryInput{
		CidrBlock:    aws.String(cidr),
		Egress:       aws.Bool(ingressMode),
		NetworkAclId: aws.String(networkACLID),
		Protocol:     aws.String(fmt.Sprintf("%d", protocolNumber)),
		PortRange:    portRange,
		RuleAction:   aws.String(ec2.RuleActionDeny),
		RuleNumber:   aws.Int64(ruleNumber),
	})

	if err != nil {
		return err
	}
	return nil
}

func (n *NACLClientImpl) RetrieveNACLEntries(networkACLID string) (*ec2.DescribeNetworkAclsOutput, error) {
	return n.EC2Srv.DescribeNetworkAcls(&ec2.DescribeNetworkAclsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("network-acl-id"),
				Values: []*string{aws.String(networkACLID)},
			},
		},
	})
}
