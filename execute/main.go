package main

import (
	"context"
	"errors"
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
	"github.com/moznion/aws-lambda-ddos-hangman/execute/data"
	"github.com/moznion/aws-lambda-ddos-hangman/execute/internal"
	"github.com/moznion/aws-lambda-ddos-hangman/execute/repo"
)

const (
	regionEnvVarKey          = "REGION"
	beginRuleNumberEnvVarKey = "BEGIN_RULE_NUMBER"
	ignoreErrorEnvVarKey     = "IGNORE_ERROR"
	tableNameEnvVarKey       = "TABLE_NAME"
)
const ruleNumberUpperLimit = 32766

var region = os.Getenv(regionEnvVarKey)
var beginRuleNumber = os.Getenv(beginRuleNumberEnvVarKey)
var shouldIgnoreError = os.Getenv(ignoreErrorEnvVarKey) != ""
var tableName = os.Getenv(tableNameEnvVarKey)

var sess = session.Must(session.NewSessionWithOptions(session.Options{
	Config: aws.Config{
		Region: aws.String(region),
	},
}))
var ec2Srv = ec2.New(sess)
var dynamodbSrv = dynamodb.New(sess)

var deniedApplicantRepo repo.DeniedApplicantRepo = repo.NewDeniedApplicantRepoImpl(dynamodbSrv, tableName)
var naclClient internal.NACLClient = internal.NewNACLClientImpl(ec2Srv)

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
	switch events.DynamoDBOperationType(record.EventName) {
	case events.DynamoDBOperationTypeInsert:
		image, err := internal.ConvertAttrValueMap(record.Change.NewImage)
		if err != nil {
			return err
		}
		var deniedApplicant data.DeniedApplicant
		err = dynamodbattribute.UnmarshalMap(image, &deniedApplicant)
		if err != nil {
			return err
		}
		log.Printf("[info] inserted denied applicant: %#v\n", deniedApplicant)

		subjectString := deniedApplicant.Subject
		subject, err := data.ParseSubjectString(subjectString)
		if err != nil {
			return err
		}

		aclRuleNumber, err := denyByNACL(deniedApplicant.NetworkACLID, subject)
		if err != nil {
			return err
		}

		err = deniedApplicantRepo.UpdateACLRuleNumber(subject, aclRuleNumber)
		if err != nil {
			return err
		}
	case events.DynamoDBOperationTypeRemove:
		image, err := internal.ConvertAttrValueMap(record.Change.OldImage)
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
			err = naclClient.ReleaseDenyingByNACL(deniedApplicant.NetworkACLID, deniedApplicant.ACLRuleNumber, false)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func denyByNACL(networkACLID string, subject *data.Subject) (int64, error) {
	ruleNumber, err := strconv.ParseInt(beginRuleNumber, 10, 64)
	if err != nil {
		return 0, err
	}

	acls, err := naclClient.RetrieveNACLEntries(networkACLID)
	if err != nil {
		return 0, err
	}

	ruleNumSet := convertNetworkACLsToRuleNumberSet(acls.NetworkAcls)

	for {
		if ruleNumber > ruleNumberUpperLimit {
			return 0, errors.New("there is no available rule number (upper limit exceeded)")
		}

		if !ruleNumSet[ruleNumber] {
			break
		}

		ruleNumber++
	}

	portRange := subject.PortRange()
	for {
		if ruleNumber > ruleNumberUpperLimit {
			return 0, errors.New("there is no available rule number (upper limit exceeded)")
		}

		err = naclClient.DenyByNACL(subject.CIDR, subject.ProtocolNumber, networkACLID, ruleNumber, portRange, false)
		if err == nil {
			break
		}

		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == "NetworkAclEntryLimitExceeded" {
				// Remove the oldest denied applicant like FIFO.
				err = deniedApplicantRepo.DeleteOldestDeniedApplicant()
				if err != nil {
					return 0, err
				}
				continue
			}

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

func main() {
	lambda.Start(handler)
}

func convertNetworkACLsToRuleNumberSet(acls []*ec2.NetworkAcl) map[int64]bool {
	ruleNumSet := make(map[int64]bool)
	for _, acl := range acls {
		for _, entries := range acl.Entries {
			ruleNumber := entries.RuleNumber
			if ruleNumber != nil {
				ruleNumSet[*ruleNumber] = true
			}
		}
	}
	return ruleNumSet
}
