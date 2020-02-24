package repo

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/moznion/aws-lambda-ddos-hangman/execute/data"
	"github.com/stretchr/testify/assert"
)

const tableName = "test-denied-applicants"
const dynamodbEndpoint = "http://127.0.0.1"
const dynamodbPort = 8000
const subjectKeyName = "subject"

var d *dynamodb.DynamoDB
var repo *DeniedApplicantRepoImpl

func init() {
	conf := &aws.Config{
		Region:   aws.String("us-west-2"),
		Endpoint: aws.String(fmt.Sprintf("%s:%d", dynamodbEndpoint, dynamodbPort)),
		HTTPClient: &http.Client{
			Timeout: 1 * time.Second,
		},
	}
	sess, err := session.NewSession(conf)
	if err != nil {
		panic(err)
	}
	d = dynamodb.New(sess)

	repo = NewDeniedApplicantRepoImpl(d, tableName)
}

func setupTestTable(d *dynamodb.DynamoDB) {
	_, _ = d.DeleteTable(&dynamodb.DeleteTableInput{
		TableName: aws.String(tableName),
	})
	_, err := d.CreateTable(&dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String(subjectKeyName),
				AttributeType: aws.String(dynamodb.ScalarAttributeTypeS),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String(subjectKeyName),
				KeyType:       aws.String(dynamodb.KeyTypeHash),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(1),
			WriteCapacityUnits: aws.Int64(1),
		},
		TableName: aws.String(tableName),
	})

	if err != nil {
		panic(err)
	}
}

func TestRemoveOldestDeniedApplicant(t *testing.T) {
	setupTestTable(d)

	const createdAt1 uint64 = 1582423132830
	err := repo.PutDeniedApplicant(data.NewDeniedApplicant(&data.Subject{
		CreatedAtEpochMillis: createdAt1,
		CIDR:                 "192.168.1.1/32",
		ProtocolNumber:       6,
		FromPort:             22,
		ToPort:               80,
	}, "acl-foo", 100))
	assert.NoError(t, err)

	const createdAt2 uint64 = 2582423132830
	err = repo.PutDeniedApplicant(data.NewDeniedApplicant(&data.Subject{
		CreatedAtEpochMillis: createdAt2,
		CIDR:                 "192.168.0.2/32",
		ProtocolNumber:       6,
		FromPort:             22,
		ToPort:               80,
	}, "acl-bar", 101))
	assert.NoError(t, err)

	err = repo.DeleteOldestDeniedApplicant()
	assert.NoError(t, err)

	applicant, err := repo.getOldestDeniedApplicant()
	assert.NoError(t, err)
	assert.EqualValues(t, 101, applicant.ACLRuleNumber)

	err = repo.DeleteOldestDeniedApplicant()
	assert.NoError(t, err)

	applicant, err = repo.getOldestDeniedApplicant()
	assert.EqualError(t, err, errNoDeniedApplicant.Error())
}

func TestUpdateACLRuleNumber(t *testing.T) {
	setupTestTable(d)

	subject := &data.Subject{
		CreatedAtEpochMillis: uint64(time.Now().UnixNano() / int64(time.Millisecond)),
		CIDR:                 "192.168.1.1/32",
		ProtocolNumber:       6,
		FromPort:             22,
		ToPort:               80,
	}

	err := repo.PutDeniedApplicant(data.NewDeniedApplicant(subject, "acl-foo", 100))
	assert.NoError(t, err)

	applicant, err := repo.getOldestDeniedApplicant()
	assert.NoError(t, err)
	assert.EqualValues(t, 100, applicant.ACLRuleNumber)

	err = repo.UpdateACLRuleNumber(subject, 200)
	assert.NoError(t, err)

	applicant, err = repo.getOldestDeniedApplicant()
	assert.NoError(t, err)
	assert.EqualValues(t, 200, applicant.ACLRuleNumber)
}
