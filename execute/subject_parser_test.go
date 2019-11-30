package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSubjectWhenOnlyCidrAndProto(t *testing.T) {
	given := "192.0.2.1/32:6"

	cidr, protocolNumber, fromPort, toPort, err := parseSubject(given)
	assert.NoError(t, err)
	assert.EqualValues(t, "192.0.2.1/32", cidr)
	assert.EqualValues(t, 6, protocolNumber)
	assert.EqualValues(t, 0, fromPort)
	assert.EqualValues(t, 0, toPort)
}

func TestParseSubjectWhenCidrAndProtoWithPortRange(t *testing.T) {
	given := "192.0.2.1/32:17:22-123"

	cidr, protocolNumber, fromPort, toPort, err := parseSubject(given)
	assert.NoError(t, err)
	assert.EqualValues(t, "192.0.2.1/32", cidr)
	assert.EqualValues(t, 17, protocolNumber)
	assert.EqualValues(t, 22, fromPort)
	assert.EqualValues(t, 123, toPort)
}

func TestParseSubjectShouldFailWhenInvalidCidr(t *testing.T) {
	given := "192.0.2.1:17"

	cidr, fromPort, protocolNumber, toPort, err := parseSubject(given)
	assert.Equal(t, errInvalidSubjectFormat, errors.Unwrap(err))
	assert.EqualValues(t, "", cidr)
	assert.EqualValues(t, 0, protocolNumber)
	assert.EqualValues(t, 0, fromPort)
	assert.EqualValues(t, 0, toPort)
}

func TestParseSubjectShouldFailWhenInvalidPortNotation(t *testing.T) {
	given := "192.0.2.1/32:17:1-"

	cidr, fromPort, protocolNumber, toPort, err := parseSubject(given)
	assert.Equal(t, errInvalidSubjectFormat, errors.Unwrap(err))
	assert.EqualValues(t, "", cidr)
	assert.EqualValues(t, 0, protocolNumber)
	assert.EqualValues(t, 0, fromPort)
	assert.EqualValues(t, 0, toPort)
}
