package aws

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/lamassuiot/lamassu-cla/services/api/internal/domain"
	"github.com/lamassuiot/lamassu-cla/services/api/internal/ports"
)

// Integration tests against DynamoDB Local (A-30). CLA_DYNAMODB_ENDPOINT selects the endpoint.
// When CLA_DYNAMODB_REQUIRED=1 (set by `make test-dynamodb` and CI), a missing endpoint or an
// unreachable service fails the tests instead of skipping them.

type seqRandom struct {
	mu sync.Mutex
	n  byte
}

func (r *seqRandom) Secret() (domain.Secret, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.n++
	s := make(domain.Secret, domain.SecretBytes)
	for i := range s {
		s[i] = r.n
	}
	return s, nil
}

func localClient(t *testing.T) *dynamodb.Client {
	t.Helper()
	endpoint := os.Getenv("CLA_DYNAMODB_ENDPOINT")
	required := os.Getenv("CLA_DYNAMODB_REQUIRED") == "1"
	if endpoint == "" {
		if required {
			t.Fatal("CLA_DYNAMODB_REQUIRED=1 but CLA_DYNAMODB_ENDPOINT is empty")
		}
		t.Skip("CLA_DYNAMODB_ENDPOINT not set; run `make test-dynamodb`")
	}
	client := NewDynamoDBClient(awssdk.Config{}, endpoint)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := client.ListTables(ctx, &dynamodb.ListTablesInput{}); err != nil {
		t.Fatalf("DynamoDB at %s is unreachable: %v", endpoint, err)
	}
	return client
}

func createTable(t *testing.T, c *dynamodb.Client, name string, withSK bool) {
	t.Helper()
	defs := []types.AttributeDefinition{{AttributeName: awssdk.String("PK"), AttributeType: types.ScalarAttributeTypeS}}
	keys := []types.KeySchemaElement{{AttributeName: awssdk.String("PK"), KeyType: types.KeyTypeHash}}
	if withSK {
		defs = append(defs, types.AttributeDefinition{AttributeName: awssdk.String("SK"), AttributeType: types.ScalarAttributeTypeS})
		keys = append(keys, types.KeySchemaElement{AttributeName: awssdk.String("SK"), KeyType: types.KeyTypeRange})
	}
	_, err := c.CreateTable(context.Background(), &dynamodb.CreateTableInput{
		TableName: awssdk.String(name), BillingMode: types.BillingModePayPerRequest,
		AttributeDefinitions: defs, KeySchema: keys,
	})
	if err != nil {
		t.Fatalf("create table %s: %v", name, err)
	}
	t.Cleanup(func() {
		_, _ = c.DeleteTable(context.Background(), &dynamodb.DeleteTableInput{TableName: awssdk.String(name)})
	})
}

func newStore(t *testing.T) (*DynamoDB, *dynamodb.Client, Tables) {
	t.Helper()
	c := localClient(t)
	suffix := time.Now().UTC().Format("150405.000000")
	tables := Tables{Ephemeral: "ephemeral-" + suffix, Core: "core-" + suffix, Audit: "audit-" + suffix}
	createTable(t, c, tables.Ephemeral, false)
	createTable(t, c, tables.Core, true)
	createTable(t, c, tables.Audit, true)
	d, err := NewDynamoDB(c, tables, &seqRandom{})
	if err != nil {
		t.Fatal(err)
	}
	return d, c, tables
}

func count(t *testing.T, c *dynamodb.Client, table string) int {
	t.Helper()
	out, err := c.Scan(context.Background(), &dynamodb.ScanInput{TableName: awssdk.String(table), Select: types.SelectCount})
	if err != nil {
		t.Fatal(err)
	}
	return int(out.Count)
}

var (
	t0       = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	identity = domain.GitHubIdentity{UserID: "1234567", Login: "synthetic-contributor"}
)

// F07-T2 (store part): consume is atomic and enforces expiry itself.
func TestDynamoDBConsume(t *testing.T) {
	d, _, _ := newStore(t)
	ctx := context.Background()
	attempt := domain.SignInAttempt{StateHash: "h1", Verifier: "v1", ReturnTo: "/agreements/", ExpiresAt: t0.Add(10 * time.Minute)}
	if err := d.Put(ctx, attempt); err != nil {
		t.Fatal(err)
	}
	const n = 16
	var wg sync.WaitGroup
	results := make([]error, n)
	got := make([]domain.SignInAttempt, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			got[i], results[i] = d.Consume(ctx, "h1", t0.Add(time.Minute))
		}(i)
	}
	wg.Wait()
	wins := 0
	for i, err := range results {
		switch {
		case err == nil:
			wins++
			if got[i] != attempt {
				t.Fatalf("returned attempt = %+v, want %+v", got[i], attempt)
			}
		case errors.Is(err, ports.ErrNotFound):
		default:
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if wins != 1 {
		t.Fatalf("winners = %d, want 1", wins)
	}
	// Expired before TTL removal: condition fails.
	if err := d.Put(ctx, domain.SignInAttempt{StateHash: "h2", Verifier: "v", ReturnTo: "/", ExpiresAt: t0}); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Consume(ctx, "h2", t0); !errors.Is(err, ports.ErrNotFound) {
		t.Fatalf("expired consume: %v", err)
	}
	// Delete of a superseded attempt is idempotent.
	if err := d.Delete(ctx, "h2"); err != nil {
		t.Fatal(err)
	}
	if err := d.Delete(ctx, "missing"); err != nil {
		t.Fatal(err)
	}
}

// F07-T3 and F07-T8 (store part): session replacement and deletion carry the audit event in one
// transaction; a cancelled transaction writes nothing.
func TestDynamoDBSessions(t *testing.T) {
	d, c, tables := newStore(t)
	ctx := context.Background()
	st := d.Sessions()
	policy := domain.DefaultSessionPolicy
	first := domain.NewSession("hash-1", "csrf-1", identity, policy, t0)
	if err := st.Replace(ctx, first, "", domain.AuditEvent{Type: domain.AuditSignedIn, ActorID: identity.UserID, OccurredAt: t0, RequestID: "r1"}); err != nil {
		t.Fatal(err)
	}
	got, err := st.Get(ctx, "hash-1")
	if err != nil || got != first {
		t.Fatalf("Get = %+v, %v; want %+v", got, err, first)
	}
	if count(t, c, tables.Audit) != 1 {
		t.Fatal("one audit event expected")
	}
	// Replacement deletes the old session and writes the second event atomically.
	second := domain.NewSession("hash-2", "csrf-2", identity, policy, t0.Add(time.Hour))
	if err := st.Replace(ctx, second, "hash-1", domain.AuditEvent{Type: domain.AuditSignedIn, ActorID: identity.UserID, OccurredAt: t0.Add(time.Hour), RequestID: "r2"}); err != nil {
		t.Fatal(err)
	}
	if _, err := st.Get(ctx, "hash-1"); !errors.Is(err, ports.ErrNotFound) {
		t.Fatal("old session must be gone")
	}
	if count(t, c, tables.Ephemeral) != 1 || count(t, c, tables.Audit) != 2 {
		t.Fatalf("ephemeral=%d audit=%d", count(t, c, tables.Ephemeral), count(t, c, tables.Audit))
	}
	// A replacement whose put condition fails (token hash collision) cancels the whole
	// transaction: no audit event is written.
	err = st.Replace(ctx, second, "", domain.AuditEvent{Type: domain.AuditSignedIn, ActorID: identity.UserID, OccurredAt: t0.Add(2 * time.Hour), RequestID: "r3"})
	if !errors.Is(err, ports.ErrNotFound) || count(t, c, tables.Audit) != 2 {
		t.Fatalf("cancelled transaction: err=%v audit=%d", err, count(t, c, tables.Audit))
	}
	// Touch: conditional on the session existing and being unexpired.
	touched, due := second.Touch(policy, t0.Add(time.Hour+2*time.Minute))
	if !due {
		t.Fatal("touch expected")
	}
	if err := st.Touch(ctx, touched); err != nil {
		t.Fatal(err)
	}
	got, _ = st.Get(ctx, "hash-2")
	if !got.IdleExpiresAt.Equal(touched.IdleExpiresAt) {
		t.Fatalf("idle expiry = %v, want %v", got.IdleExpiresAt, touched.IdleExpiresAt)
	}
	if err := st.Touch(ctx, domain.Session{TokenHash: "gone", LastSeenAt: t0, IdleExpiresAt: t0}); !errors.Is(err, ports.ErrNotFound) {
		t.Fatalf("touch of a missing session: %v", err)
	}
	// Delete writes the sign-out event with the deletion; a second delete finds nothing and
	// writes nothing.
	if err := st.Delete(ctx, "hash-2", domain.AuditEvent{Type: domain.AuditSignedOut, ActorID: identity.UserID, OccurredAt: t0.Add(3 * time.Hour), RequestID: "r4"}); err != nil {
		t.Fatal(err)
	}
	if count(t, c, tables.Ephemeral) != 0 || count(t, c, tables.Audit) != 3 {
		t.Fatalf("after delete: ephemeral=%d audit=%d", count(t, c, tables.Ephemeral), count(t, c, tables.Audit))
	}
	if err := st.Delete(ctx, "hash-2", domain.AuditEvent{Type: domain.AuditSignedOut, ActorID: identity.UserID, OccurredAt: t0, RequestID: "r5"}); !errors.Is(err, ports.ErrNotFound) {
		t.Fatalf("second delete: %v", err)
	}
	if count(t, c, tables.Audit) != 3 {
		t.Fatal("a failed delete must not write an audit event")
	}
}

// F07-T12 (store part): first sign-in creates PENDING; later sign-ins refresh only the login.
func TestDynamoDBContributors(t *testing.T) {
	d, c, tables := newStore(t)
	ctx := context.Background()
	repo := d.Contributors()
	if _, err := repo.Get(ctx, identity.UserID); !errors.Is(err, ports.ErrNotFound) {
		t.Fatalf("missing contributor: %v", err)
	}
	if err := repo.Put(ctx, domain.NewContributor(identity, t0)); err != nil {
		t.Fatal(err)
	}
	got, err := repo.Get(ctx, identity.UserID)
	if err != nil || got.Status != domain.ContributorPending || got.GitHubLogin != identity.Login || !got.CreatedAt.Equal(t0) {
		t.Fatalf("Get = %+v, %v", got, err)
	}
	// An administrative status set elsewhere survives a login refresh (AC-07-11, AC-07-12).
	_, err = c.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: awssdk.String(tables.Core), Key: contributorKey(identity.UserID),
		UpdateExpression: awssdk.String("SET #s = :s"), ExpressionAttributeNames: map[string]string{"#s": "status"},
		ExpressionAttributeValues: map[string]types.AttributeValue{":s": s("REVOKED")},
	})
	if err != nil {
		t.Fatal(err)
	}
	refreshed, _ := got.RefreshLogin(domain.GitHubIdentity{UserID: identity.UserID, Login: "new-login"}, t0.Add(time.Hour))
	refreshed.Status = domain.ContributorRevoked
	if err := repo.Put(ctx, refreshed); err != nil {
		t.Fatal(err)
	}
	got, _ = repo.Get(ctx, identity.UserID)
	if got.Status != domain.ContributorRevoked || got.GitHubLogin != "new-login" || !got.CreatedAt.Equal(t0) || !got.UpdatedAt.Equal(t0.Add(time.Hour)) {
		t.Fatalf("after refresh = %+v", got)
	}
}

func TestNewDynamoDBValidation(t *testing.T) {
	if _, err := NewDynamoDB(nil, Tables{}, nil); err == nil {
		t.Fatal("missing tables must be rejected")
	}
	if _, err := NewDynamoDB(nil, Tables{Ephemeral: "a", Core: "b", Audit: "c"}, &seqRandom{}); err == nil {
		t.Fatal("nil client must be rejected")
	}
}

func TestMapError(t *testing.T) {
	if !errors.Is(mapError(&types.ConditionalCheckFailedException{}), ports.ErrNotFound) {
		t.Fatal("conditional failure must map to not found")
	}
	if !errors.Is(mapError(&types.TransactionCanceledException{}), ports.ErrNotFound) {
		t.Fatal("cancelled transaction must map to not found")
	}
	if !errors.Is(mapError(context.DeadlineExceeded), ports.ErrUncertain) {
		t.Fatal("timeout must map to uncertain")
	}
	if !errors.Is(mapError(&types.ProvisionedThroughputExceededException{}), ports.ErrUnavailable) {
		t.Fatal("a service rejection must map to unavailable")
	}
	if !errors.Is(mapError(errors.New("connection reset")), ports.ErrUncertain) {
		t.Fatal("a transport error must map to uncertain")
	}
}
