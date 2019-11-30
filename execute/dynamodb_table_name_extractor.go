package main

import (
	"errors"
	"fmt"
	"strings"
)

var (
	errInsufficientDynamodbEventSourceArn = errors.New("insufficient dynamodb event source ARN has come")
)

func extractDynamodbTableNameFromEventSourceArn(eventSourceArn string) (string, error) {
	// example expected data: arn:aws:dynamodb:ap-northeast-1:0123456789:table/tableName/stream/2019-11-30T22:24:10.892

	leaves := strings.Split(eventSourceArn, ":")
	if len(leaves) < 6 {
		return "", fmt.Errorf("%w: %s", errInsufficientDynamodbEventSourceArn, eventSourceArn)
	}

	resources := strings.Split(leaves[5], "/")
	if len(resources) < 4 {
		return "", fmt.Errorf("%w: %s", errInsufficientDynamodbEventSourceArn, eventSourceArn)
	}

	return resources[1], nil
}
