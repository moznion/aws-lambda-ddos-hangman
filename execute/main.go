package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/moznion/aws-lambda-ddos-hangman/execute/internal/data"
)

const (
	regionEnvVarKey          = "REGION"
	beginRuleNumberEnvVarKey = "BEGIN_RULE_NUMBER"
	ignoreErrorEnvVarKey     = "IGNORE_ERROR"
)
const ruleNumberUpperLimit = 32766

var region = os.Getenv(regionEnvVarKey)
var beginRuleNumber = os.Getenv(beginRuleNumberEnvVarKey)
var shouldIgnoreError = os.Getenv(ignoreErrorEnvVarKey) != ""

var ingressMode = aws.Bool(false)
var sess = session.Must(session.NewSessionWithOptions(session.Options{
	Config: aws.Config{
		Region: aws.String(region),
	},
}))
var ec2Srv = ec2.New(sess)
var dynamodbSrv = dynamodb.New(sess)

func handler(ctx context.Context, event events.DynamoDBEvent) (string, error) {
	for _, record := range event.Records {
		err := handleRecord(record)
		if err != nil {
			log.Printf("[error] %s\n", err)
			if shouldIgnoreError {
				log.Print("[info] continued\n")
				continue
			}
			return "", err
		}
	}
	return "OK", nil
}

func handleRecord(record events.DynamoDBEventRecord) error {
	tableName, err := extractDynamodbTableNameFromEventSourceArn(record.EventSourceArn)
	if err != nil {
		return err
	}

	switch events.DynamoDBOperationType(record.EventName) {
	case events.DynamoDBOperationTypeInsert:
		image, err := convertAttrValueMap(record.Change.NewImage)
		if err != nil {
			return err
		}
		var deniedApplicant data.DeniedApplicant
		err = dynamodbattribute.UnmarshalMap(image, &deniedApplicant)
		if err != nil {
			return err
		}
		log.Printf("[info] inserted denied applicant: %#v\n", deniedApplicant)

		subject := deniedApplicant.Subject
		cidr, protocolNumber, fromPort, toPort, err := parseSubject(subject)
		if err != nil {
			return err
		}

		aclRuleNumber, err := denyByNACL(deniedApplicant.NetworkACLID, protocolNumber, cidr, fromPort, toPort)
		if err != nil {
			return err
		}

		err = markACLRuleNumberOnDynamodbTable(tableName, subject, aclRuleNumber)
		if err != nil {
			return err
		}
	case events.DynamoDBOperationTypeRemove:
		image, err := convertAttrValueMap(record.Change.OldImage)
		if err != nil {
			return err
		}
		var deniedApplicant data.DeniedApplicant
		err = dynamodbattribute.UnmarshalMap(image, &deniedApplicant)
		if err != nil {
			return err
		}
		log.Printf("[info] removed denied applicant: %#v\n", deniedApplicant)

		if deniedApplicant.ACLRuleNumber != 0 {
			err = releaseDenyingByNACL(deniedApplicant.NetworkACLID, deniedApplicant.ACLRuleNumber)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func denyByNACL(networkACLID string, protocolNumber int64, cidrBlock string, fromPort int64, toPort int64) (int64, error) {
	var portRange *ec2.PortRange
	if fromPort != 0 && toPort != 0 {
		portRange = &ec2.PortRange{
			From: aws.Int64(fromPort),
			To:   aws.Int64(toPort),
		}
	}

	ruleNumber, err := strconv.ParseInt(beginRuleNumber, 10, 64)
	if err != nil {
		return 0, err
	}

	acls, err := ec2Srv.DescribeNetworkAcls(&ec2.DescribeNetworkAclsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("network-acl-id"),
				Values: []*string{aws.String(networkACLID)},
			},
		},
	})
	if err != nil {
		return 0, err
	}

	ruleNumBag := make(map[int64]bool)
	for _, acl := range acls.NetworkAcls {
		for _, entries := range acl.Entries {
			ruleNumber := entries.RuleNumber
			if ruleNumber != nil {
				ruleNumBag[*ruleNumber] = true
			}
		}
	}

	for {
		if ruleNumber > ruleNumberUpperLimit {
			return 0, errors.New("there is no available rule number (upper limit exceeded)")
		}

		if !ruleNumBag[ruleNumber] {
			break
		}

		ruleNumber++
	}

	for {
		if ruleNumber > ruleNumberUpperLimit {
			return 0, errors.New("there is no available rule number (upper limit exceeded)")
		}

		// TODO IPv6 supporting
		_, err = ec2Srv.CreateNetworkAclEntry(&ec2.CreateNetworkAclEntryInput{
			CidrBlock:    aws.String(cidrBlock),
			Egress:       ingressMode,
			NetworkAclId: aws.String(networkACLID),
			Protocol:     aws.String(fmt.Sprintf("%d", protocolNumber)),
			PortRange:    portRange,
			RuleAction:   aws.String(ec2.RuleActionDeny),
			RuleNumber:   aws.Int64(ruleNumber),
		})
		if err == nil {
			break
		}

		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == "NetworkAclEntryAlreadyExists" {
				// retry with increment rule number
				ruleNumber++
				continue
			}
		}
		return 0, err
	}

	return ruleNumber, nil
}

func markACLRuleNumberOnDynamodbTable(tableName string, subject string, aclRuleNumber int64) error {
	_, err := dynamodbSrv.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"subject": {
				S: aws.String(subject),
			},
		},
		UpdateExpression: aws.String("SET #ACL_RULE_NUMBER = :aclRuleNumber"),
		ExpressionAttributeNames: map[string]*string{
			"#ACL_RULE_NUMBER": aws.String("aclRuleNumber"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":aclRuleNumber": {N: aws.String(fmt.Sprintf("%d", aclRuleNumber))},
		},
		ReturnValues: aws.String(dynamodb.ReturnConsumedCapacityNone),
	})

	if err != nil {
		return err
	}
	return nil
}

func releaseDenyingByNACL(networkACLID string, ruleNumber int64) error {
	_, err := ec2Srv.DeleteNetworkAclEntry(&ec2.DeleteNetworkAclEntryInput{
		Egress:       ingressMode,
		NetworkAclId: aws.String(networkACLID),
		RuleNumber:   aws.Int64(ruleNumber),
	})
	if err != nil {
		return err
	}
	return nil
}

func main() {
	lambda.Start(handler)
}
