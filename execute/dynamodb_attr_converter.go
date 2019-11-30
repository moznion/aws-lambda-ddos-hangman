package main

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func convertAttrValueMap(image map[string]events.DynamoDBAttributeValue) (map[string]*dynamodb.AttributeValue, error) {
	dbAttrMap := make(map[string]*dynamodb.AttributeValue)
	for k, v := range image {
		var dbAttr dynamodb.AttributeValue

		bytes, err := v.MarshalJSON()
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(bytes, &dbAttr); err != nil {
			return nil, err
		}
		dbAttrMap[k] = &dbAttr
	}

	return dbAttrMap, nil
}
