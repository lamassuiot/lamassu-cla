// Package aws implements ports with AWS services. Adapters are the only layer that imports SDKs.
//
// DynamoDB items follow specs/domain-model.md section 5: the ephemeral table (OAUTH# and
// SESSION# items keyed by hash), the core table (CONTRIB#{id} / PROFILE), and the audit table
// (ENTITY#... / {timestamp}#{event_id}). Session changes and their audit events are written in one
// TransactWriteItems call (F-07, data-model changes; INV-08).
package aws

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/smithy-go"

	"github.com/lamassuiot/lamassu-cla/services/api/internal/domain"
	"github.com/lamassuiot/lamassu-cla/services/api/internal/ports"
)

// Tables names the three DynamoDB tables (ADR-0009).
type Tables struct {
	Ephemeral string
	Core      string
	Audit     string
}

// Validate reports missing table names.
func (t Tables) Validate() error {
	if t.Ephemeral == "" || t.Core == "" || t.Audit == "" {
		return errors.New("dynamodb: ephemeral, core, and audit table names are required")
	}
	return nil
}

// DynamoDB implements the F-07 store ports.
type DynamoDB struct {
	client *dynamodb.Client
	tables Tables
	// newEventID supplies audit event IDs; injected so that tests are deterministic.
	newEventID func() (string, error)
}

// NewDynamoDB wires the stores. random supplies audit event IDs.
func NewDynamoDB(client *dynamodb.Client, tables Tables, random ports.RandomSource) (*DynamoDB, error) {
	if err := tables.Validate(); err != nil {
		return nil, err
	}
	if client == nil || random == nil {
		return nil, errors.New("dynamodb: client and random source are required")
	}
	return &DynamoDB{client: client, tables: tables, newEventID: func() (string, error) {
		s, err := random.Secret()
		if err != nil {
			return "", err
		}
		return s.Encode(), nil
	}}, nil
}

func s(v string) *types.AttributeValueMemberS { return &types.AttributeValueMemberS{Value: v} }
func n(v int64) *types.AttributeValueMemberN {
	return &types.AttributeValueMemberN{Value: strconv.FormatInt(v, 10)}
}

func getS(item map[string]types.AttributeValue, key string) string {
	if v, ok := item[key].(*types.AttributeValueMemberS); ok {
		return v.Value
	}
	return ""
}

func getTime(item map[string]types.AttributeValue, key string) time.Time {
	if v, ok := item[key].(*types.AttributeValueMemberN); ok {
		if secs, err := strconv.ParseInt(v.Value, 10, 64); err == nil {
			return time.Unix(secs, 0).UTC()
		}
	}
	return time.Time{}
}

// mapError translates SDK failures into the port's sentinel errors. Conditional failures are
// reported by the caller; everything else is either unavailable (never reached the service, or the
// service said no) or uncertain (the request may have been applied).
func mapError(err error) error {
	var cfe *types.ConditionalCheckFailedException
	if errors.As(err, &cfe) {
		return ports.ErrNotFound
	}
	var tce *types.TransactionCanceledException
	if errors.As(err, &tce) {
		// A cancelled transaction applied nothing; a condition inside it failed.
		return ports.ErrNotFound
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return fmt.Errorf("%w: %w", ports.ErrUncertain, err)
	}
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		// The service answered: the write was rejected, not applied.
		return fmt.Errorf("%w: %s", ports.ErrUnavailable, apiErr.ErrorCode())
	}
	// Transport failures after the request was sent: the outcome is unknown.
	return fmt.Errorf("%w: %w", ports.ErrUncertain, err)
}

// --- SignInAttemptStore -----------------------------------------------------------------------

func attemptKey(stateHash string) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{"PK": s("OAUTH#" + stateHash)}
}

// Put stores the attempt with TTL on expires_at.
func (d *DynamoDB) Put(ctx context.Context, a domain.SignInAttempt) error {
	_, err := d.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: awssdk.String(d.tables.Ephemeral),
		Item: map[string]types.AttributeValue{
			"PK":         s("OAUTH#" + a.StateHash),
			"verifier":   s(a.Verifier),
			"return_to":  s(a.ReturnTo),
			"expires_at": n(a.ExpiresAt.Unix()),
		},
	})
	if err != nil {
		return mapError(err)
	}
	return nil
}

// Consume deletes the item if it exists and is unexpired, returning it (F-07, step 4.3).
func (d *DynamoDB) Consume(ctx context.Context, stateHash string, now time.Time) (domain.SignInAttempt, error) {
	out, err := d.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName:                 awssdk.String(d.tables.Ephemeral),
		Key:                       attemptKey(stateHash),
		ConditionExpression:       awssdk.String("attribute_exists(PK) AND expires_at > :now"),
		ExpressionAttributeValues: map[string]types.AttributeValue{":now": n(now.Unix())},
		ReturnValues:              types.ReturnValueAllOld,
	})
	if err != nil {
		return domain.SignInAttempt{}, mapError(err)
	}
	return domain.SignInAttempt{
		StateHash: stateHash,
		Verifier:  getS(out.Attributes, "verifier"),
		ReturnTo:  getS(out.Attributes, "return_to"),
		ExpiresAt: getTime(out.Attributes, "expires_at"),
	}, nil
}

// Delete removes a superseded attempt; a missing item is not an error.
func (d *DynamoDB) Delete(ctx context.Context, stateHash string) error {
	_, err := d.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{TableName: awssdk.String(d.tables.Ephemeral), Key: attemptKey(stateHash)})
	if err != nil {
		return mapError(err)
	}
	return nil
}

// --- SessionStore -------------------------------------------------------------------------------

// Sessions returns the ports.SessionStore view of the adapter.
func (d *DynamoDB) Sessions() ports.SessionStore { return sessionStore{d} }

type sessionStore struct{ d *DynamoDB }

func sessionKey(tokenHash string) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{"PK": s("SESSION#" + tokenHash)}
}

func sessionItem(sess domain.Session) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"PK":              s("SESSION#" + sess.TokenHash),
		"github_user_id":  s(sess.GitHubUserID),
		"github_login":    s(sess.GitHubLogin),
		"csrf_token":      s(sess.CSRFToken),
		"created_at":      n(sess.CreatedAt.Unix()),
		"last_seen_at":    n(sess.LastSeenAt.Unix()),
		"idle_expires_at": n(sess.IdleExpiresAt.Unix()),
		"expires_at":      n(sess.ExpiresAt.Unix()),
	}
}

func sessionFromItem(tokenHash string, item map[string]types.AttributeValue) domain.Session {
	return domain.Session{
		TokenHash:     tokenHash,
		GitHubUserID:  getS(item, "github_user_id"),
		GitHubLogin:   getS(item, "github_login"),
		CSRFToken:     getS(item, "csrf_token"),
		CreatedAt:     getTime(item, "created_at"),
		LastSeenAt:    getTime(item, "last_seen_at"),
		IdleExpiresAt: getTime(item, "idle_expires_at"),
		ExpiresAt:     getTime(item, "expires_at"),
	}
}

func (st sessionStore) Get(ctx context.Context, tokenHash string) (domain.Session, error) {
	out, err := st.d.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: awssdk.String(st.d.tables.Ephemeral), Key: sessionKey(tokenHash), ConsistentRead: awssdk.Bool(true),
	})
	if err != nil {
		return domain.Session{}, mapError(err)
	}
	if out.Item == nil {
		return domain.Session{}, ports.ErrNotFound
	}
	return sessionFromItem(tokenHash, out.Item), nil
}

// Replace writes the session, deletes the previous one, and writes the audit event atomically.
func (st sessionStore) Replace(ctx context.Context, sess domain.Session, previousTokenHash string, event domain.AuditEvent) error {
	audit, err := st.d.auditPut(event)
	if err != nil {
		return err
	}
	items := []types.TransactWriteItem{
		{Put: &types.Put{
			TableName: awssdk.String(st.d.tables.Ephemeral), Item: sessionItem(sess),
			ConditionExpression: awssdk.String("attribute_not_exists(PK)"),
		}},
		{Put: audit},
	}
	if previousTokenHash != "" && previousTokenHash != sess.TokenHash {
		items = append(items, types.TransactWriteItem{Delete: &types.Delete{
			TableName: awssdk.String(st.d.tables.Ephemeral), Key: sessionKey(previousTokenHash),
		}})
	}
	if _, err := st.d.client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{TransactItems: items}); err != nil {
		return mapError(err)
	}
	return nil
}

// Touch persists the slid idle expiry, conditional on the session still existing and unexpired.
func (st sessionStore) Touch(ctx context.Context, sess domain.Session) error {
	_, err := st.d.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:           awssdk.String(st.d.tables.Ephemeral),
		Key:                 sessionKey(sess.TokenHash),
		UpdateExpression:    awssdk.String("SET last_seen_at = :seen, idle_expires_at = :idle"),
		ConditionExpression: awssdk.String("attribute_exists(PK) AND idle_expires_at > :seen AND expires_at > :seen"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":seen": n(sess.LastSeenAt.Unix()),
			":idle": n(sess.IdleExpiresAt.Unix()),
		},
	})
	if err != nil {
		return mapError(err)
	}
	return nil
}

// Delete removes the session and writes the audit event atomically; a missing session is
// reported as ErrNotFound without writing the event.
func (st sessionStore) Delete(ctx context.Context, tokenHash string, event domain.AuditEvent) error {
	audit, err := st.d.auditPut(event)
	if err != nil {
		return err
	}
	_, err = st.d.client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{TransactItems: []types.TransactWriteItem{
		{Delete: &types.Delete{
			TableName: awssdk.String(st.d.tables.Ephemeral), Key: sessionKey(tokenHash),
			ConditionExpression: awssdk.String("attribute_exists(PK)"),
		}},
		{Put: audit},
	}})
	if err != nil {
		return mapError(err)
	}
	return nil
}

// auditPut builds the audit item (domain model, sections 2.12 and 5.2). F-07 events carry no
// personal data beyond the GitHub user ID as actor.
func (d *DynamoDB) auditPut(event domain.AuditEvent) (*types.Put, error) {
	id, err := d.newEventID()
	if err != nil {
		return nil, fmt.Errorf("audit event id: %w", err)
	}
	ts := event.OccurredAt.UTC().Format(time.RFC3339)
	return &types.Put{
		TableName: awssdk.String(d.tables.Audit),
		Item: map[string]types.AttributeValue{
			"PK":          s("ENTITY#CONTRIBUTOR#" + event.ActorID),
			"SK":          s(ts + "#" + id),
			"GSI1PK":      s("ACTOR#CONTRIBUTOR#" + event.ActorID),
			"GSI1SK":      s(ts),
			"GSI2PK":      s("DAY#" + event.OccurredAt.UTC().Format("2006-01-02")),
			"GSI2SK":      s(ts),
			"event_id":    s(id),
			"event_type":  s(string(event.Type)),
			"actor_type":  s("CONTRIBUTOR"),
			"actor_id":    s(event.ActorID),
			"entity_type": s("CONTRIBUTOR"),
			"entity_id":   s(event.ActorID),
			"timestamp":   s(ts),
			"request_id":  s(event.RequestID),
		},
		// Audit events are never overwritten (INV-07).
		ConditionExpression: awssdk.String("attribute_not_exists(PK)"),
	}, nil
}

// --- ContributorRepository ----------------------------------------------------------------------

// Contributors returns the ports.ContributorRepository view of the adapter.
func (d *DynamoDB) Contributors() ports.ContributorRepository { return contributorRepo{d} }

type contributorRepo struct{ d *DynamoDB }

func contributorKey(id string) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{"PK": s("CONTRIB#" + id), "SK": s("PROFILE")}
}

func (r contributorRepo) Get(ctx context.Context, githubUserID string) (domain.Contributor, error) {
	out, err := r.d.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: awssdk.String(r.d.tables.Core), Key: contributorKey(githubUserID), ConsistentRead: awssdk.Bool(true),
	})
	if err != nil {
		return domain.Contributor{}, mapError(err)
	}
	if out.Item == nil {
		return domain.Contributor{}, ports.ErrNotFound
	}
	return domain.Contributor{
		GitHubUserID: githubUserID,
		GitHubLogin:  getS(out.Item, "github_login"),
		Status:       domain.ContributorStatus(getS(out.Item, "status")),
		CreatedAt:    getTime(out.Item, "created_at"),
		UpdatedAt:    getTime(out.Item, "updated_at"),
	}, nil
}

// Put writes the profile. Only sign-in's attributes are set; agreement links and lookup values
// are separate items and untouched (AC-07-12).
func (r contributorRepo) Put(ctx context.Context, c domain.Contributor) error {
	_, err := r.d.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: awssdk.String(r.d.tables.Core),
		Key:       contributorKey(c.GitHubUserID),
		UpdateExpression: awssdk.String(
			"SET github_login = :login, #status = if_not_exists(#status, :status), " +
				"created_at = if_not_exists(created_at, :created), updated_at = :updated, GSI3PK = :gsi3"),
		ExpressionAttributeNames: map[string]string{"#status": "status"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":login":   s(c.GitHubLogin),
			":status":  s(string(c.Status)),
			":created": n(c.CreatedAt.Unix()),
			":updated": n(c.UpdatedAt.Unix()),
			":gsi3":    s("LOGIN#" + lower(c.GitHubLogin)),
		},
	})
	if err != nil {
		return mapError(err)
	}
	return nil
}

func lower(v string) string {
	b := []byte(v)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 'a' - 'A'
		}
	}
	return string(b)
}
