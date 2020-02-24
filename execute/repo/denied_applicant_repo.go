package repo

import "github.com/moznion/aws-lambda-ddos-hangman/execute/data"

type DeniedApplicantRepo interface {
	DeleteOldestDeniedApplicant() error
	PutDeniedApplicant(deniedApplicant *data.DeniedApplicant) error
	UpdateACLRuleNumber(subject *data.Subject, aclRuleNumber int64) error
}
