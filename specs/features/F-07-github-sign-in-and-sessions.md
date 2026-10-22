# F-07: GitHub sign-in and sessions

| Field | Value |
| --- | --- |
| Status | Draft; blocked on D-02, D-06, and D-20 before approval |
| Phase | 4 |
| Owner | Repository owner; Security owner for the authentication model (D-06) |
| Depends on | F-02, F-03, F-06; D-02, D-06, D-08 (if D-06 selects a GitHub App), D-20, D-28, D-29; [security](../security/security.md#31-people-github-sign-in), [data protection](../security/data-protection.md), [threat model](../security/threat-model.md) (T-15, T-16) |
| Approved by | Pending |

## Problem statement

Every personal action on the platform (signing an ICLA, managing an organization, reviewing
agreements) needs to know which GitHub account the person controls (FR-CON-01, FR-ICLA-01). The
portal has no sign-in yet, and the API has no handlers. F-07 adds GitHub sign-in, server-side
sessions, sign-out, and the request authentication and CSRF protection that later features build
on. It does not decide what a signed-in person may do; roles and authorization are F-08.

## User roles

- Visitors: they start sign-in.
- Signed-in people (any GitHub account): they see who they are signed in as and can sign out.

No role-specific behavior. F-08 maps people to maintainer and administrator roles (D-13).

## Preconditions

- D-06 is decided: GitHub App user authorization (recommended in the decision log) or a GitHub
  OAuth App, and the session model. This draft follows the recommendation; the API, data model,
  and acceptance criteria hold for either choice unless marked otherwise.
- D-02 is decided, so the callback URL can be registered exactly (no wildcards).
- D-20 is decided, so the portal can show the privacy notice before sign-in.
- If D-06 selects a GitHub App, the App is registered (D-08) and its client ID and client secret
  are stored in Secrets Manager by the platform owner. Nothing about the App is committed.
- Running sign-in in an environment needs F-03 step 5 (deployment). Until then, everything is
  built and tested without AWS: domain and use cases with fakes, adapters against local test
  servers, and Terraform plan-only with mocked providers.

## Main flow

1. A visitor chooses **Sign in with GitHub** in the portal header or on a page that needs sign-in.
   The portal navigates the browser to `GET /v1/auth/github/start?return_to=<relative path>`.
2. The API validates `return_to` against the relative-path allowlist (falling back to `/`),
   creates a single-use `state` and a PKCE verifier, stores them hashed in the `ephemeral` table for
   10 minutes, sets a short-lived pre-authentication cookie bound to `state`, and redirects to
   GitHub's authorization page with the exact registered callback URL.
3. The person approves on GitHub. GitHub redirects to `GET /v1/auth/github/callback` with `code`
   and `state`.
4. The API checks that `state` exists, is unexpired, unused, and matches the pre-authentication
   cookie; deletes it; exchanges `code` (with the PKCE verifier) for a user access token; reads the
   GitHub user ID and login; and discards the token. The token is never stored, logged, or sent to
   the browser.
5. The API creates or updates the contributor record (`github_user_id`; `github_login` refreshed,
   FR-CON-01). A new contributor starts as `PENDING` (proposed; D-29). No email address is read or stored in F-07;
   email lookup values (D-19) are added by the feature that resolves commit authors (F-15).
6. The API creates a session: a 256-bit random token, stored only as a hash in the `ephemeral`
   table with the GitHub user ID, creation time, last-seen time, idle expiry, absolute expiry, and
   a session-bound CSRF secret. It sets the `__Host-cla_session` cookie (`HttpOnly`, `Secure`,
   `SameSite=Lax`, `Path=/`), clears the pre-authentication cookie, writes an audit event, and
   redirects to the validated `return_to`.
7. The portal calls `GET /v1/session`, which returns the signed-in person (GitHub user ID and
   login), the session expiry, and the CSRF token. The header shows the login and a **Sign out**
   button.
8. Every authenticated request is checked by middleware: the session hash must exist and be
   within its idle and absolute expiry; the last-seen time is extended (throttled to at most one
   write per minute). State-changing requests also need a matching `X-CSRF-Token` header and an
   allowed `Origin`.
9. **Sign out** calls `POST /v1/session/sign-out` (CSRF-protected). The API deletes the session,
   expires the cookie, writes an audit event, and returns 204; the portal returns to the landing
   page in the signed-out state.

## Alternative flows

- **Already signed in.** Starting sign-in again replaces the session (the old one is deleted), so a
  session token is never reused across sign-ins (session rotation).
- **Session expires.** After the idle timeout (proposed 30 minutes) or the absolute timeout
  (proposed 8 hours), the API answers 401; the portal shows the signed-out state and offers sign-in
  with `return_to` set to the current page. Timeouts are configuration (security, section 3.1);
  final values come with D-06.
- **Suspended or revoked contributor.** Sign-in still succeeds: it proves identity, not coverage.
  Coverage and later authorization decisions use the contributor status. Open question D-28.
- **GitHub login changed.** The GitHub user ID is the identity; the stored login is refreshed on
  every sign-in.

## Error cases

| Case | Detection | System behavior | User-visible result |
| --- | --- | --- | --- |
| Missing, expired, reused, or mismatched `state` | `ephemeral` lookup and pre-authentication cookie check | No token exchange; audit-free; logged as a security event without personal data | Error page: sign-in could not be completed, with a link to try again |
| Person declines on GitHub | `error=access_denied` on the callback | No session | Signed-out portal with a message that sign-in was cancelled |
| GitHub unavailable or code exchange fails | Upstream error or timeout | No session; no retry loop | Error page with a link to try again |
| `return_to` not allowed | Allowlist check | Falls back to `/` | Lands on the landing page |
| Missing, unknown, or expired session | Middleware | 401 problem response | Signed-out state with a sign-in prompt |
| Missing or wrong CSRF token, or disallowed `Origin` | Middleware on state-changing requests | 403 problem response; nothing changes | Error message; reloading the session fixes a stale token |
| Session storage unavailable | Store error | 503 problem response; no session created or extended | Error state with retry |

## Security considerations

- Follows the [security specification](../security/security.md#31-people-github-sign-in), section
  3.1, and threats T-15 (session theft or CSRF) and T-16 (open redirect).
- `state` and PKCE verifier are single use, expire after 10 minutes, and are stored hashed; the
  pre-authentication cookie binds them to the browser that started sign-in.
- Session tokens are 256-bit random values from a cryptographic source, stored only as SHA-256
  hashes; comparisons of secrets are constant time. The CSRF token is bound to the session.
- No GitHub token is persisted, logged, or returned. Requested GitHub permissions are the minimum
  for reading the user ID and login (no email scope in F-07).
- `return_to` accepts only relative paths from an allowlist; protocol-relative and absolute URLs
  are rejected.
- Logs, metrics, and problem responses contain no session tokens, codes, `state` values, or
  personal data; the request ID correlates them. Audit events record the GitHub user ID as the
  actor, as the domain model allows.
- The API sets no CORS allowance for other origins; the portal and API share a site (ADR-0010).

## API changes

New operations in `specs/api/openapi.yaml` under the existing `auth` tag, added in unit 1 before
any handler. None is breaking; `sessionCookie` and `csrfToken` security schemes already exist.

| Operation | Purpose | Responses |
| --- | --- | --- |
| `GET /v1/auth/github/start` | Start sign-in; query `return_to` (relative path) | 302 to GitHub; 400 for malformed input |
| `GET /v1/auth/github/callback` | Complete sign-in; query `code`, `state`, or `error` | 302 to `return_to` with the session cookie; 302 to a portal error route on failure |
| `GET /v1/session` | Current session (`sessionCookie`) | 200 with `github_user_id`, `github_login`, `expires_at`, `csrf_token`; 401 |
| `POST /v1/session/sign-out` | End the session (`sessionCookie` and `csrfToken`) | 204; 401; 403 |

## Data-model changes

- `ephemeral` table (domain model, section 5.3): the existing `OAUTH#{state_hash}` record gains the
  PKCE verifier and the validated `return_to`; the existing `SESSION#{session_token_hash}` record is
  specified (GitHub user ID, `created_at`, `last_seen_at`, `idle_expires_at`, `expires_at`, CSRF
  secret hash). Both use TTL on `expires_at`, and readers treat expired items as absent.
- `core` table: contributor records are created on first sign-in (`PENDING`), and `github_login`
  is refreshed on later sign-ins. No new attributes.
- `audit` table: events `auth.signed_in` and `auth.signed_out`, written in the same transaction as
  the session change where the stores allow it.

## UI changes

- Header action: **Sign in with GitHub** when signed out; the GitHub login and **Sign out** when
  signed in. Built from existing components; no new tokens.
- A sign-in error route shows the error cases above with an `ErrorState` and a retry link.
- The privacy notice from D-20 is linked next to the sign-in action.
- Session state comes from `GET /v1/session` through the generated API client. Using TanStack
  Query for server state (architecture, section 7) is a new runtime dependency and is proposed with
  evidence in unit 4, like A-27.

## Acceptance criteria

- **AC-07-1** Given a visitor who starts sign-in, the API redirects to GitHub with a new single-use
  `state`, a PKCE challenge, and the exact registered callback URL, and sets a pre-authentication
  cookie bound to that `state`.
- **AC-07-2** Given a callback whose `state` is missing, expired, already used, or not bound to the
  browser's pre-authentication cookie, no code exchange happens and no session is created.
- **AC-07-3** Given a successful callback, a contributor record exists for the GitHub user ID with
  the current login, a session exists whose token is stored only as a hash, the `__Host-cla_session`
  cookie is `HttpOnly`, `Secure`, `SameSite=Lax`, and `Path=/`, and the browser is redirected to the
  validated `return_to`.
- **AC-07-4** Given any `return_to` that is not an allowed relative path (absolute, protocol-relative,
  backslash, or encoded variants), the redirect goes to `/`.
- **AC-07-5** No GitHub token, authorization code, `state`, session token, or email address appears
  in storage, logs, metrics, audit events, or responses.
- **AC-07-6** Given a session past its idle or absolute expiry, or an unknown token, authenticated
  operations return 401.
- **AC-07-7** Given a state-changing request without the session's CSRF token or with a disallowed
  `Origin`, the API returns 403 and changes nothing.
- **AC-07-8** Given sign-out, the session is deleted, the cookie is expired, and the old token no
  longer authenticates.
- **AC-07-9** Signing in again replaces the previous session (rotation).
- **AC-07-10** The portal shows the signed-out and signed-in header states and the sign-in error
  route, with no WCAG 2.2 A or AA violations in either theme at the F-06 widths.

## Automated test scenarios

| ID | Scenario | Level | Covers |
| --- | --- | --- | --- |
| F07-T1 | Sign-in start: redirect parameters, stored hashed `state` and verifier, pre-authentication cookie. | Unit (use case with fakes), handler | AC-07-1 |
| F07-T2 | Callback with missing, expired, reused, and mismatched `state`: no exchange, no session. | Unit, handler | AC-07-2 |
| F07-T3 | Successful callback against a local fake of GitHub's endpoints: contributor upsert, hashed session, cookie attributes, redirect. | Adapter and handler (local test server) | AC-07-3, AC-07-9 |
| F07-T4 | `return_to` table test with relative, absolute, protocol-relative, backslash, and encoded inputs. | Unit | AC-07-4 |
| F07-T5 | Logs, audit events, stored items, and responses from the flows above are scanned for the token, code, `state`, and session token. | Unit, handler | AC-07-5 |
| F07-T6 | Middleware: idle and absolute expiry with a fake clock; unknown token. | Unit | AC-07-6 |
| F07-T7 | CSRF and `Origin` checks on a state-changing route. | Handler | AC-07-7 |
| F07-T8 | Sign-out deletes the session and expires the cookie; the old token returns 401. | Handler | AC-07-8 |
| F07-T9 | OpenAPI contract tests: every new operation's responses validate against `openapi.yaml`. | Contract | All API criteria |
| F07-T10 | Portal header states and sign-in error route, axe and reflow at the F-06 widths, with the API mocked. | Component (browser) | AC-07-10 |
| F07-T11 | Terraform plan-only tests with mocked providers for the new API, Lambda, tables, and secret. | Infrastructure | Rollback and least privilege |

## Observability requirements

- Structured log events (no personal data, request ID included): `auth.start`, `auth.callback`
  with outcome (`success`, `state_invalid`, `denied`, `upstream_error`), `auth.sign_out`,
  `session.rejected` with reason (`unknown`, `idle_expired`, `expired`, `csrf`, `origin`).
- Metrics: sign-in outcomes by reason; GitHub exchange latency and errors; session store errors.
- Alarm proposals (finalized in F-18): sustained GitHub exchange errors; spikes in `state_invalid`
  or CSRF rejections.
- Audit events: `auth.signed_in`, `auth.signed_out`.

## Rollback considerations

Revert the PRs. Sessions and OAuth state are ephemeral and expire by TTL; contributor records
created by sign-in contain only the GitHub user ID and login and stay valid. Infrastructure is not
applied until F-03 step 5; after that, removing the API routes leaves the tables in place (they are
retained, not destroyed, by default).

## Implementation plan

1. **Contract, domain, and use cases.** OpenAPI operations; domain types (session, sign-in attempt,
   `return_to` policy); ports (identity provider, session store, contributor repository, audit
   writer, clock, random source); use cases with fakes (F07-T1, F07-T2, F07-T4 to F07-T6).
2. **Adapters and HTTP.** GitHub identity adapter tested against a local fake server; DynamoDB
   adapters (test approach proposed with evidence, for example a local DynamoDB container in the
   pinned CI image); handlers, middleware, Lambda wiring, and a local development server (F07-T3,
   F07-T7 to F07-T9).
3. **Infrastructure (plan-only).** API Gateway HTTP API, the `api` Lambda, the `core`, `audit`, and
   `ephemeral` tables, the Secrets Manager secret (value set by the platform owner), and
   least-privilege IAM, with plan-only tests (F07-T11). No apply until F-03 step 5.
4. **Portal.** Session state, header states, sign-in error route, privacy-notice link, and
   accessibility runs (F07-T10); the server-state library proposed with evidence.

Units 1 and 2 depend on D-06; unit 4 depends on D-20; registering the callback URL depends on D-02.

## Open questions

- D-02: production URL, needed for the exact callback URL.
- D-06: GitHub App user authorization or OAuth App, and session timeouts (proposed 30 minutes idle,
  8 hours absolute).
- D-08: GitHub App registration, if D-06 selects a GitHub App.
- D-20: lawful basis and privacy notice shown at sign-in.
- D-28: may suspended or revoked contributors sign in?
- D-29: does the first sign-in create a contributor record, or only starting an agreement?
