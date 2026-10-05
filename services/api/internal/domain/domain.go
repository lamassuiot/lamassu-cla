// Package domain holds entities, value objects, state machines, and invariants.
// It performs no I/O and imports only the standard library and other domain packages.
package domain

// AgreementType identifies the kind of contributor license agreement.
type AgreementType string

const (
	AgreementTypeICLA AgreementType = "ICLA"
	AgreementTypeCCLA AgreementType = "CCLA"
)

// Valid reports whether t is a known agreement type.
func (t AgreementType) Valid() bool {
	return t == AgreementTypeICLA || t == AgreementTypeCCLA
}
