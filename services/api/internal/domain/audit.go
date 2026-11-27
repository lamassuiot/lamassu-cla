package domain

import "time"

// AuditEventType names the audit events F-07 writes (domain model, audit table).
type AuditEventType string

const (
	AuditSignedIn  AuditEventType = "auth.signed_in"
	AuditSignedOut AuditEventType = "auth.signed_out"
)

// AuditEvent is written in the same transaction as the session change it describes
// (F-07, data-model changes). The actor is the GitHub user ID; it carries no other personal data.
type AuditEvent struct {
	Type       AuditEventType
	ActorID    string
	OccurredAt time.Time
	RequestID  string
}
