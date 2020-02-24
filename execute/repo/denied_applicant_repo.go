package repo

import "github.com/moznion/aws-lambda-ddos-hangman/execute/data"

// DeniedApplicantRepo is a repository for denied applicants.
type DeniedApplicantRepo interface {
	DeleteOldestDeniedApplicant() error
	PutDeniedApplicant(deniedApplicant *data.DeniedApplicant) error
	UpdateACLRuleNumber(subject *data.Subject, aclRuleNumber int64) error
}
