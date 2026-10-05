# API Conventions

| Field | Value |
| --- | --- |
| Status | Draft — in review |
| Contract | [`openapi.yaml`](openapi.yaml) (OpenAPI 3.1, source of truth) |

These conventions apply to every operation. Feature specifications add operations to the contract
following them. Changing a convention is a contract change ([change control](../sdd-workflow.md#3-change-control)).

## 1. Contract-first

- `openapi.yaml` is changed **before** handlers. CI lints the contract and fails when the generated
  TypeScript client or Go server types are out of date.
- Every operation declares `operationId`, tags, security, request and response schemas, and every
  error response it can return.
- Every operation declares its required roles with the `x-required-roles` extension. An empty list
  means any authenticated user; operations without security are public.

## 2. Versioning

- The major version is in the path: `/v1/...`.
- Within a major version, only additive changes are allowed: new operations, new optional request
  fields, new response fields, new enum values in responses (clients must tolerate unknown values).
- Breaking changes require a new major version, a deprecation period for the old one, and a
  Conventional Commit with `!`.
- Deprecated operations return `Deprecation` and `Sunset` headers.

## 3. Representation

- JSON (`application/json`) with UTF-8. Field names in `snake_case`.
- Enums are `UPPER_SNAKE_CASE`.
- Timestamps are RFC 3339 UTC strings; dates are `YYYY-MM-DD`.
- Identifiers are strings. Clients treat them as opaque.
- Unknown request fields are rejected with `400`.
- Request bodies are limited in size; the limit is defined per operation, with a global default of
  64 KiB.

## 4. Errors

Errors use Problem Details (RFC 9457) with `application/problem+json`:

```json
{
  "type": "https://example.org/problems/validation-error",
  "title": "Validation failed",
  "status": 400,
  "code": "VALIDATION_FAILED",
  "detail": "One or more fields are invalid.",
  "request_id": "01J9Z3K4Q8X7M2N5P6R7S8T9V0",
  "errors": [
    { "field": "language", "code": "UNSUPPORTED_VALUE" }
  ]
}
```

- `code` is a stable, documented machine-readable value. Clients branch on `code`, not on `title`.
- The `type` URI base is open until the production URL is confirmed (D-02). The example uses the
  reserved `example.org` domain.
- Error responses never include stack traces, SQL or DynamoDB details, provider responses, secrets,
  or another user's data.
- `404` is returned instead of `403` when revealing existence would leak private information, for
  example another contributor's agreement.

| Status | Use |
| --- | --- |
| `400` | Malformed request or validation failure. |
| `401` | Missing or invalid session. |
| `403` | Authenticated but not permitted. |
| `404` | Not found or not visible to the caller. |
| `409` | State conflict, for example an active ICLA already exists. |
| `412` | `If-Match` precondition failed (optimistic concurrency). |
| `422` | Idempotency key reused with a different request. |
| `429` | Rate limit exceeded, with `Retry-After`. |
| `503` | Dependency unavailable (for example the signing provider), with `Retry-After`. |

## 5. Pagination, filtering, and sorting

- Cursor-based pagination: `limit` (default 25, maximum 100) and `cursor` (opaque).
- List responses: `{ "items": [...], "next_cursor": "..." }`. `next_cursor` is absent on the last
  page.
- Filters are explicit, documented query parameters per operation. No generic query language.
- Sorting uses `sort` with an allowlisted field, prefixed by `-` for descending, for example
  `sort=-updated_at`.

## 6. Idempotency

- `POST` operations that create resources or start workflows require an `Idempotency-Key` header
  (UUID or ULID).
- The platform stores the key, the caller, a hash of the request, and the response for 24 hours.
- Same key and same request: the original response is replayed.
- Same key and different request: `422` with code `IDEMPOTENCY_KEY_REUSED`.
- `PUT`, `PATCH`, and `DELETE` use `If-Match` with the resource `etag` for optimistic concurrency.

## 7. Authentication and authorization

- Portal users authenticate with a session cookie set by the API after GitHub sign-in (model: D-06).
  The cookie is `HttpOnly`, `Secure`, `SameSite=Lax`, and uses the `__Host-` prefix.
- State-changing requests require the `X-CSRF-Token` header matching the session's CSRF token, and
  an allowed `Origin`.
- Webhooks authenticate with signatures, not sessions (section 9).
- Authorization is enforced in the application layer for every operation, independent of the UI.

## 8. Correlation and rate limits

- Clients may send `X-Request-Id` (ULID or UUID). Otherwise the API generates one. It is returned
  on every response and logged with every event.
- API Gateway applies stage-level and route-level throttling. Values are set per environment in
  Terraform. Exceeding limits returns `429` with `Retry-After`.
- Authentication and signing-session creation have stricter limits than read operations.

## 9. Webhooks

| Aspect | GitHub | Signing provider |
| --- | --- | --- |
| Verification | HMAC-SHA256 of the raw body with the webhook secret, compared in constant time with `X-Hub-Signature-256`. | Adapter-specific (for DocuSign Connect: HMAC signature headers). The platform then re-reads envelope status from the provider API. |
| Deduplication | `X-GitHub-Delivery` stored in the `ephemeral` table. | Hash of provider event ID and envelope ID. |
| Response | `202` after verification and enqueueing; `401` on invalid signature; `400` on malformed payload. | `200` after verification and enqueueing; `401` on failed verification. |
| Replay protection | Delivery ID deduplication. Payloads for unknown installations are ignored. | Deduplication plus authoritative status re-read; stale events cannot regress state. |
| Retries | GitHub does not retry failed deliveries automatically. A scheduled job may redeliver failed deliveries through the GitHub API, and a reconciliation job re-evaluates open pull requests. | Provider retries per its policy. Processing is idempotent. |
| Body handling | Raw body preserved for signature verification; maximum size enforced. | Same. |
