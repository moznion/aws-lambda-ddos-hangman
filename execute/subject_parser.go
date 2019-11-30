package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

var subjectPattern = regexp.MustCompile("^([0-9]{1,3}[.][0-9]{1,3}[.][0-9]{1,3}[.][0-9]{1,3}/[0-9]{1,2}):([0-9]+)(?::([0-9]{1,5})-([0-9]{1,5}))?$")

var (
	errInvalidSubjectFormat = errors.New("invalid subject string has come")
)

func parseSubject(subjectString string) (string, int64, int64, int64, error) {
	subjectSubmatch := subjectPattern.FindStringSubmatch(subjectString)
	if len(subjectSubmatch) < 5 {
		return "", 0, 0, 0, fmt.Errorf("%w: %s (expected pattern: %s)", errInvalidSubjectFormat, subjectString, subjectPattern)
	}
	cidr := subjectSubmatch[1]

	protocolNumber, _ := strconv.ParseInt(subjectSubmatch[2], 10, 64)

	var fromPort int64
	fromPortStr := subjectSubmatch[3]
	if fromPortStr != "" {
		fromPort, _ = strconv.ParseInt(fromPortStr, 10, 64)
	}

	var toPort int64
	toPortStr := subjectSubmatch[4]
	if toPortStr != "" {
		toPort, _ = strconv.ParseInt(toPortStr, 10, 64)
	}

	return cidr, protocolNumber, fromPort, toPort, nil
}
