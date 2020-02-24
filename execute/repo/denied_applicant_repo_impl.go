package repo

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/moznion/aws-lambda-ddos-hangman/execute/data"
)

type DeniedApplicantRepoImpl struct {
	dyn       *dynamodb.DynamoDB
	tableName string
}

var (
	errNoDeniedApplicant = errors.New("there is no denied applicant")
)

func NewDeniedApplicantRepoImpl(dyn *dynamodb.DynamoDB, tableName string) *DeniedApplicantRepoImpl {
	return &DeniedApplicantRepoImpl{
		dyn:       dyn,
		tableName: tableName,
	}
}

func (r *DeniedApplicantRepoImpl) DeleteOldestDeniedApplicant() error {
	oldestDeniedApplicant, err := r.getOldestDeniedApplicant()
	if err != nil {
		if errors.Is(err, errNoDeniedApplicant) {
			return nil
		}
		return err
	}

	err = r.deleteDeniedApplicant(oldestDeniedApplicant)
	if err != nil {
		return err
	}

	return nil
}

func (r *DeniedApplicantRepoImpl) PutDeniedApplicant(deniedApplicant *data.DeniedApplicant) error {
	item, err := dynamodbattribute.MarshalMap(deniedApplicant)
	if err != nil {
		return err
	}

	_, err = r.dyn.PutItem(&dynamodb.PutItemInput{
		Item:         item,
		ReturnValues: aws.String(dynamodb.ReturnValueNone),
		TableName:    aws.String(r.tableName),
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *DeniedApplicantRepoImpl) UpdateACLRuleNumber(subject *data.Subject, aclRuleNumber int64) error {
	_, err := r.dyn.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"subject": {
				S: aws.String(subject.String()),
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

func (r *DeniedApplicantRepoImpl) getOldestDeniedApplicant() (*data.DeniedApplicant, error) {
	// NOTE:
	// a maximum number of Network ACL is not big (quota of the system is 20, special in case is 40),
	// so using Scan might not be a serious problem.
	//
	// Ref: https://docs.aws.amazon.com/vpc/latest/userguide/amazon-vpc-limits.html#vpc-limits-nacls
	res, err := r.dyn.Scan(&dynamodb.ScanInput{
		Limit:     aws.Int64(40),
		TableName: aws.String(r.tableName),
	})
	if err != nil {
		return nil, err
	}

	if *res.Count <= 0 {
		// nothing to do
		return nil, errNoDeniedApplicant
	}

	items := res.Items
	oldestItem := items[len(items)-1]

	var deniedApplicant data.DeniedApplicant
	err = dynamodbattribute.UnmarshalMap(oldestItem, &deniedApplicant)
	if err != nil {
		return nil, err
	}

	return &deniedApplicant, nil
}

func (r *DeniedApplicantRepoImpl) deleteDeniedApplicant(deniedApplicant *data.DeniedApplicant) error {
	_, err := r.dyn.DeleteItem(&dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"subject": {
				S: aws.String(deniedApplicant.Subject),
			},
		},
		ReturnValues: aws.String(dynamodb.ReturnValueNone),
		TableName:    aws.String(r.tableName),
	})
	if err != nil {
		return err
	}

	return nil
}
