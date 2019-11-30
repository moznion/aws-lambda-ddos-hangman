package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractDynamodbTableNameFromEventSourceArnShouldBeSuccessful(t *testing.T) {
	given := "arn:aws:dynamodb:ap-northeast-1:0123456789:table/tableName/stream/2019-11-30T22:24:10.892"

	got, err := extractDynamodbTableNameFromEventSourceArn(given)
	assert.NoError(t, err)
	assert.EqualValues(t, "tableName", got)
}

func TestExtractDynamodbTableNameFromEventSourceArnShouldFailWhenArnMissingResource(t *testing.T) {
	given := "arn:aws:dynamodb:ap-northeast-1:0123456789"

	got, err := extractDynamodbTableNameFromEventSourceArn(given)
	assert.Equal(t, errInsufficientDynamodbEventSourceArn, errors.Unwrap(err))
	assert.EqualValues(t, "", got)
}

func TestExtractDynamodbTableNameFromEventSourceArnShouldFailWhenResourceIsInsufficient(t *testing.T) {
	given := "arn:aws:dynamodb:ap-northeast-1:0123456789:table/tableName"

	got, err := extractDynamodbTableNameFromEventSourceArn(given)
	assert.Equal(t, errInsufficientDynamodbEventSourceArn, errors.Unwrap(err))
	assert.EqualValues(t, "", got)
}
