package data

// DeniedApplicant is a structure that represents an applicant to deny inbound requests.
type DeniedApplicant struct {
	// Subject is a partition key of the DynamoDB table. The function applies denying rule for this target.
	//
	// it has to follow the following regexp pattern:
	// (?<created_at_epoch_millis>[0-9]+):(?<cidr>[0-9]{1,3}[.][0-9]{1,3}[.][0-9]{1,3}[.][0-9]{1,3}/[0-9]{1,2}):(?<protocol>[0-9]+)(?::(?<port_range>[0-9]{1,5}-[0-9]{1,5}))?
	//
	// NOTE: `protocol` means the protocol number (e.g. 6=TCP and 17=UDP)
	//
	// example:
	// 1582425243392:192.168.1.1/32:6
	// 1582425243392:192.168.1.1/32:6:22-80
	Subject string `json:"subject"`

	// NetworkACLID is an identifier of the target ACL to apply the access control rule.
	NetworkACLID string `json:"networkAclID"`

	// ACLRuleNumber is the number that represents the NACL rule number.
	ACLRuleNumber int64 `json:"aclRuleNumber"`
}

// NewDeniedApplicant creates new DeniedApplicant.
func NewDeniedApplicant(subject *Subject, networkACLID string, aclRuleNumber int64) *DeniedApplicant {
	return &DeniedApplicant{
		Subject:       subject.String(),
		NetworkACLID:  networkACLID,
		ACLRuleNumber: aclRuleNumber,
	}
}
