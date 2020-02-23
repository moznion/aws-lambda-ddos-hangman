package data

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseSubjectStringWhenOnlyCidrAndProto(t *testing.T) {
	givenCreatedAtEpochMillis := uint64(time.Now().UnixNano() / int64(time.Millisecond))
	givenCidr := "192.0.2.1/32"
	givenProtocol := 6
	given := fmt.Sprintf("%d:%s:%d", givenCreatedAtEpochMillis, givenCidr, givenProtocol)

	subject, err := ParseSubjectString(given)
	assert.NoError(t, err)
	assert.EqualValues(t, givenCreatedAtEpochMillis, subject.CreatedAtEpochMillis)
	assert.EqualValues(t, givenCidr, subject.CIDR)
	assert.EqualValues(t, givenProtocol, subject.ProtocolNumber)
	assert.EqualValues(t, 0, subject.FromPort)
	assert.EqualValues(t, 0, subject.ToPort)
}

func TestParseSubjectStringWhenCidrAndProtoWithPortRange(t *testing.T) {
	givenCreatedAtEpochMillis := uint64(time.Now().UnixNano() / int64(time.Millisecond))
	givenCidr := "192.0.2.1/32"
	givenProtocol := 17
	givenFromPort := 22
	givenToPort := 123
	given := fmt.Sprintf("%d:%s:%d:%d-%d", givenCreatedAtEpochMillis, givenCidr, givenProtocol, givenFromPort, givenToPort)

	subject, err := ParseSubjectString(given)
	assert.NoError(t, err)
	assert.EqualValues(t, givenCreatedAtEpochMillis, subject.CreatedAtEpochMillis)
	assert.EqualValues(t, givenCidr, subject.CIDR)
	assert.EqualValues(t, givenProtocol, subject.ProtocolNumber)
	assert.EqualValues(t, givenFromPort, subject.FromPort)
	assert.EqualValues(t, givenToPort, subject.ToPort)
}

func TestParseSubjectStringShouldFailWhenInvalidCidr(t *testing.T) {
	given := "1582425243:192.0.2.1:17"

	subject, err := ParseSubjectString(given)
	assert.Equal(t, errInvalidSubjectFormat, errors.Unwrap(err))
	assert.EqualValues(t, 0, subject.CreatedAtEpochMillis)
	assert.EqualValues(t, "", subject.CIDR)
	assert.EqualValues(t, 0, subject.ProtocolNumber)
	assert.EqualValues(t, 0, subject.FromPort)
	assert.EqualValues(t, 0, subject.ToPort)
}

func TestParseSubjectStringShouldFailWhenInvalidPortNotation(t *testing.T) {
	given := "1582425243:192.0.2.1/32:17:1-"

	subject, err := ParseSubjectString(given)
	assert.Equal(t, errInvalidSubjectFormat, errors.Unwrap(err))
	assert.EqualValues(t, 0, subject.CreatedAtEpochMillis)
	assert.EqualValues(t, "", subject.CIDR)
	assert.EqualValues(t, 0, subject.ProtocolNumber)
	assert.EqualValues(t, 0, subject.FromPort)
	assert.EqualValues(t, 0, subject.ToPort)
}

func TestSubjectStringify(t *testing.T) {
	createdAtEpochMillis := time.Now().UnixNano() / int64(time.Millisecond)
	cidr := "192.0.2.1/32"
	proto := int64(17)

	subject := &Subject{
		CreatedAtEpochMillis: uint64(createdAtEpochMillis),
		CIDR:                 cidr,
		ProtocolNumber:       proto,
	}
	assert.Equal(t, fmt.Sprintf("%d:%s:%d", createdAtEpochMillis, cidr, proto), subject.String())

	subject = &Subject{
		CreatedAtEpochMillis: uint64(createdAtEpochMillis),
		CIDR:                 cidr,
		ProtocolNumber:       proto,
		FromPort:             0,
		ToPort:               22,
	}
	assert.Equal(t, fmt.Sprintf("%d:%s:%d", createdAtEpochMillis, cidr, proto), subject.String())

	subject = &Subject{
		CreatedAtEpochMillis: uint64(createdAtEpochMillis),
		CIDR:                 cidr,
		ProtocolNumber:       proto,
		FromPort:             22,
		ToPort:               0,
	}
	assert.Equal(t, fmt.Sprintf("%d:%s:%d", createdAtEpochMillis, cidr, proto), subject.String())

	fromPort := int64(22)
	toPort := int64(123)
	subject = &Subject{
		CreatedAtEpochMillis: uint64(createdAtEpochMillis),
		CIDR:                 cidr,
		ProtocolNumber:       proto,
		FromPort:             fromPort,
		ToPort:               toPort,
	}
	assert.Equal(t, fmt.Sprintf("%d:%s:%d:%d-%d", createdAtEpochMillis, cidr, proto, fromPort, toPort), subject.String())
}
