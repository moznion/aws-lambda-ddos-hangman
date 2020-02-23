package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

var subjectPattern = regexp.MustCompile("^([0-9]+):([0-9]{1,3}[.][0-9]{1,3}[.][0-9]{1,3}[.][0-9]{1,3}/[0-9]{1,2}):([0-9]+)(?::([0-9]{1,5})-([0-9]{1,5}))?$")

var (
	errInvalidSubjectFormat = errors.New("invalid subject string has come")
)

func parseSubject(subjectString string) (uint64, string, int64, int64, int64, error) {
	subjectSubmatch := subjectPattern.FindStringSubmatch(subjectString)
	if len(subjectSubmatch) < 6 {
		return 0, "", 0, 0, 0, fmt.Errorf("%w: %s (expected pattern: %s)", errInvalidSubjectFormat, subjectString, subjectPattern)
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

	return createdAtEpochMillis, cidr, protocolNumber, fromPort, toPort, nil
}
