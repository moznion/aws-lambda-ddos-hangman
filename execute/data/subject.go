package data

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type Subject struct {
	CreatedAtEpochMillis uint64
	CIDR                 string
	ProtocolNumber       int64
	FromPort             int64
	ToPort               int64
}

var subjectPattern = regexp.MustCompile("^([0-9]+):([0-9]{1,3}[.][0-9]{1,3}[.][0-9]{1,3}[.][0-9]{1,3}/[0-9]{1,2}):([0-9]+)(?::([0-9]{1,5})-([0-9]{1,5}))?$")

var (
	errInvalidSubjectFormat = errors.New("invalid subject string has come")
)

func ParseSubjectString(subjectString string) (*Subject, error) {
	subjectSubmatch := subjectPattern.FindStringSubmatch(subjectString)
	if len(subjectSubmatch) < 6 {
		return &Subject{}, fmt.Errorf("%w: %s (expected pattern: %s)", errInvalidSubjectFormat, subjectString, subjectPattern)
	}

	createdAtEpochMillis, _ := strconv.ParseUint(subjectSubmatch[1], 10, 64)
	cidr := subjectSubmatch[2]

	protocolNumber, _ := strconv.ParseInt(subjectSubmatch[3], 10, 64)

	var fromPort int64
	fromPortStr := subjectSubmatch[4]
	if fromPortStr != "" {
		fromPort, _ = strconv.ParseInt(fromPortStr, 10, 64)
	}

	var toPort int64
	toPortStr := subjectSubmatch[5]
	if toPortStr != "" {
		toPort, _ = strconv.ParseInt(toPortStr, 10, 64)
	}

	return &Subject{
		CreatedAtEpochMillis: createdAtEpochMillis,
		CIDR:                 cidr,
		ProtocolNumber:       protocolNumber,
		FromPort:             fromPort,
		ToPort:               toPort,
	}, nil
}

func (s *Subject) String() string {
	if s.FromPort == 0 || s.ToPort == 0 {
		return fmt.Sprintf("%d:%s:%d", s.CreatedAtEpochMillis, s.CIDR, s.ProtocolNumber)
	}
	return fmt.Sprintf("%d:%s:%d:%d-%d", s.CreatedAtEpochMillis, s.CIDR, s.ProtocolNumber, s.FromPort, s.ToPort)
}

func (s *Subject) PortRange() *ec2.PortRange {
	if s.FromPort != 0 && s.ToPort != 0 {
		return &ec2.PortRange{
			From: aws.Int64(s.FromPort),
			To:   aws.Int64(s.ToPort),
		}
	}
	return nil
}
