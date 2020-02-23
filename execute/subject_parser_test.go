package main

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseSubjectWhenOnlyCidrAndProto(t *testing.T) {
	givenCreatedAtEpochMillis := uint64(time.Now().UnixNano() / int64(time.Millisecond))
	givenCidr := "192.0.2.1/32"
	givenProtocol := 6
	given := fmt.Sprintf("%d:%s:%d", givenCreatedAtEpochMillis, givenCidr, givenProtocol)

	createdAtEpochMillis, cidr, protocolNumber, fromPort, toPort, err := parseSubject(given)
	assert.NoError(t, err)
	assert.EqualValues(t, givenCreatedAtEpochMillis, createdAtEpochMillis)
	assert.EqualValues(t, givenCidr, cidr)
	assert.EqualValues(t, givenProtocol, protocolNumber)
	assert.EqualValues(t, 0, fromPort)
	assert.EqualValues(t, 0, toPort)
}

func TestParseSubjectWhenCidrAndProtoWithPortRange(t *testing.T) {
	givenCreatedAtEpochMillis := uint64(time.Now().UnixNano() / int64(time.Millisecond))
	givenCidr := "192.0.2.1/32"
	givenProtocol := 17
	givenFromPort := 22
	givenToPort := 123
	given := fmt.Sprintf("%d:%s:%d:%d-%d", givenCreatedAtEpochMillis, givenCidr, givenProtocol, givenFromPort, givenToPort)

	createdAtEpochMillis, cidr, protocolNumber, fromPort, toPort, err := parseSubject(given)
	assert.NoError(t, err)
	assert.EqualValues(t, givenCreatedAtEpochMillis, createdAtEpochMillis)
	assert.EqualValues(t, givenCidr, cidr)
	assert.EqualValues(t, givenProtocol, protocolNumber)
	assert.EqualValues(t, givenFromPort, fromPort)
	assert.EqualValues(t, givenToPort, toPort)
}

func TestParseSubjectShouldFailWhenInvalidCidr(t *testing.T) {
	given := "1582425243:192.0.2.1:17"

	createdAtEpochMillis, cidr, fromPort, protocolNumber, toPort, err := parseSubject(given)
	assert.Equal(t, errInvalidSubjectFormat, errors.Unwrap(err))
	assert.EqualValues(t, 0, createdAtEpochMillis)
	assert.EqualValues(t, "", cidr)
	assert.EqualValues(t, 0, protocolNumber)
	assert.EqualValues(t, 0, fromPort)
	assert.EqualValues(t, 0, toPort)
}

func TestParseSubjectShouldFailWhenInvalidPortNotation(t *testing.T) {
	given := "1582425243:192.0.2.1/32:17:1-"

	createdAtEpochMillis, cidr, fromPort, protocolNumber, toPort, err := parseSubject(given)
	assert.Equal(t, errInvalidSubjectFormat, errors.Unwrap(err))
	assert.EqualValues(t, 0, createdAtEpochMillis)
	assert.EqualValues(t, "", cidr)
	assert.EqualValues(t, 0, protocolNumber)
	assert.EqualValues(t, 0, fromPort)
	assert.EqualValues(t, 0, toPort)
}
