# F-07: GitHub sign-in and sessions

| Field | Value |
| --- | --- |
| Status | Draft; approval needs D-06 (Security owner); sign-in for real people and contributor-record creation also need D-20 (Legal, DPO) |
| Phase | 4 |
| Owner | Repository owner; Security owner for the authentication model (D-06) |
| Depends on | F-02, F-03, F-06; D-02, D-06, D-20, D-28, D-29, D-30; [security](../security/security.md#31-people-github-sign-in), [data protection](../security/data-protection.md), [threat model](../security/threat-model.md) (T-15, T-16) |
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

- D-06 is decided by the Security owner. This draft follows the repository owner's proposal of
  2026-10-22: a GitHub **OAuth App** used only for sign-in, with no repository permissions and no
  email access, separate from the GitHub App that receives pull-request events and publishes
  Checks (F-14, D-08), which the same `lksnext` organization owns.
- D-02 is decided (2026-10-22): the portal origin is `https://cla.developers.lamassu.cloud`, the
  versioned API base is `https://cla.developers.lamassu.cloud/api/v1`, CloudFront routes `/api/*`
  to API Gateway, and the OAuth callback URL is
  `https://cla.developers.lamassu.cloud/api/auth/github/callback`, deliberately outside the
  versioned routes. The domain is not provisioned, and D-02 does not authorize deployment.
- D-20 is decided by Legal and the DPO (lawful basis and privacy notice). Until then, no sign-in by
  real people and no contributor-record creation happen in any environment (D-29); the code is
  built and tested with synthetic identities only.
- The OAuth App is owned by the `lksnext` GitHub organization, controlled by LKS S.Coop., the data
  controller (D-30, decided 2026-10-22). LKS organization admins register and manage it with
  exactly the callback URL above. Its client secret is stored in AWS Secrets Manager and readable
  only by the authentication backend. Nothing about the OAuth App is committed. The owning
  organization does not have to be the one that holds the repositories. Confirming the owner does
  not authorize production deployment or real sign-ins.
- Running sign-in in an environment needs F-03 step 5 (deployment). Until then, everything is
  built and tested without AWS: domain and use cases with fakes, adapters against local test
  servers, and Terraform plan-only with mocked providers.

## Main flow

1. A visitor chooses **Sign in with GitHub** in the portal header or on a page that needs sign-in.
   The portal navigates the browser to `GET /api/v1/auth/github/start?return_to=<relative path>`.
2. The API validates `return_to` against the relative-path allowlist (falling back to `/`),
   creates a single-use `state` and, where GitHub supports PKCE for OAuth Apps (security,
   section 3.1; confirmed in unit 2), a PKCE verifier, stores them hashed in the `ephemeral` table for
   10 minutes, sets a short-lived pre-authentication cookie bound to `state`, and redirects to
   GitHub's authorization page with the exact registered callback URL.
3. The person approves on GitHub. GitHub redirects to the registered callback,
   `GET /api/auth/github/callback`, with `code` and `state`. The callback stays outside the
   versioned API, so the registered URL does not change when the API version changes (D-02).
4. The API checks that `state` exists, is unexpired, unused, and matches the pre-authentication
   cookie; deletes it; exchanges `code` (with the PKCE verifier) for a user access token; reads the
   GitHub user ID and login; and discards the token. The token is never stored, logged, or sent to
   the browser.
5. The API creates or updates the minimal contributor record (D-29): the numeric GitHub user ID,
   which is the only stable identifier, and the current login, refreshed on every sign-in for
   display only (FR-CON-01). A new contributor starts as `PENDING`. The login is never used as an
   identifier or key. No email address is read or stored; the OAuth App requests no email access,
   and email lookup values (D-19) are added by the feature that resolves commit authors (F-15).
   Record creation is enabled only after D-20 is decided.
6. The API creates a session: a 256-bit random token, stored only as a hash in the `ephemeral`
   table with the GitHub user ID, creation time, last-seen time, idle expiry, absolute expiry, and
   a session-bound CSRF secret. It sets the `__Host-cla_session` cookie (`HttpOnly`, `Secure`,
   `SameSite=Lax`, `Path=/`), clears the pre-authentication cookie, writes an audit event, and
   redirects to the validated `return_to`.
7. The portal calls `GET /api/v1/session`, which returns the signed-in person (GitHub user ID and
   login), the session expiry, and the CSRF token. The header shows the login and a **Sign out**
   button.
8. Every authenticated request is checked by middleware: the session hash must exist and be
   within its idle and absolute expiry; the last-seen time is extended (throttled to at most one
   write per minute). State-changing requests also need a matching `X-CSRF-Token` header and an
   allowed `Origin`.
9. **Sign out** calls `POST /api/v1/session/sign-out` (CSRF-protected). The API deletes the session,
   expires the cookie, writes an audit event, and returns 204; the portal returns to the landing
   page in the signed-out state.

## Alternative flows

- **Already signed in.** Starting sign-in again replaces the session (the old one is deleted), so a
  session token is never reused across sign-ins (session rotation).
- **Session expires.** After 30 minutes without activity or 8 hours after sign-in, whichever comes
  first (D-06 proposal), the API answers 401; there is no silent renewal, and the person signs in
  again. The portal shows the signed-out state and offers sign-in with `return_to` set to the
  current page. Timeouts are configuration (security, section 3.1).
- **Suspended or revoked contributor.** Sign-in still succeeds (D-28, decided 2026-10-22).
  Authentication only confirms identity: it grants no permissions and does not restore the
  validity of any CLA. Coverage and authorization decisions keep using the contributor status.
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
- No GitHub token is persisted, logged, or returned. The OAuth App requests no scopes: it reads only
  the public numeric user ID and login, with no repository permissions and no email access (D-06
  proposal). It is separate from the GitHub App, so a sign-in token never carries repository or
  Checks permissions.
- Identity is the numeric GitHub user ID. The login can change and be reused by another account, so
  it is never used as an identifier, key, or authorization input (D-29).
- Signing in grants no permissions and does not restore an agreement's validity (D-28).
- `return_to` accepts only relative paths from an allowlist; protocol-relative and absolute URLs
  are rejected.
- Logs, metrics, and problem responses contain no session tokens, codes, `state` values, or
  personal data; the request ID correlates them. Audit events record the GitHub user ID as the
  actor, as the domain model allows.
- The API sets no CORS allowance for other origins; the portal and API share a site (ADR-0010).

## API changes

New operations in `specs/api/openapi.yaml` under the existing `auth` tag, added in unit 1 before
any handler; `sessionCookie` and `csrfToken` security schemes already exist. Paths below are
site-relative (D-02). Unit 1 also aligns the contract's server configuration with D-02 (versioned
operations under `/api/v1`, the callback at `/api/auth/github/callback`) and regenerates the API
client; the only existing operation, `GET /v1/health`, moves to `/api/v1/health`. No consumer
exists yet, so this is not a breaking change for any client.

| Operation | Purpose | Responses |
| --- | --- | --- |
| `GET /api/v1/auth/github/start` | Start sign-in; query `return_to` (relative path) | 302 to GitHub; 400 for malformed input |
| `GET /api/auth/github/callback` | Complete sign-in; query `code`, `state`, or `error`. Unversioned: it is the URL registered with GitHub (D-02) | 302 to `return_to` with the session cookie; 302 to a portal error route on failure |
| `GET /api/v1/session` | Current session (`sessionCookie`) | 200 with `github_user_id`, `github_login`, `expires_at`, `csrf_token`; 401 |
| `POST /api/v1/session/sign-out` | End the session (`sessionCookie` and `csrfToken`) | 204; 401; 403 |

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
- Session state comes from `GET /api/v1/session` through the generated API client. Using TanStack
  Query for server state (architecture, section 7) is a new runtime dependency and is proposed with
  evidence in unit 4, like A-27.

## Acceptance criteria

- **AC-07-1** Given a visitor who starts sign-in, the API redirects to GitHub's OAuth authorization
  page with a new single-use `state`, a PKCE challenge where supported, no requested scopes, and
  the exact registered callback URL, and sets a pre-authentication cookie bound to that `state`.
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

- **AC-07-11** Given a contributor whose status is `SUSPENDED` or `REVOKED`, sign-in succeeds, and
  the session carries no permission and changes no agreement or coverage state (D-28).
- **AC-07-12** Given two sign-ins by the same numeric GitHub user ID with different logins, one
  contributor record exists with the latest login; given a sign-in whose login matches another
  record's former login, no record is matched by login (D-29).

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
| F07-T12 | Sign-in of suspended and revoked contributors, and login changes and reuse, with fakes. | Unit | AC-07-11, AC-07-12 |

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
   `ephemeral` tables, the Secrets Manager secret (value set by LKS organization admins; readable only by the authentication backend), and
   least-privilege IAM, with plan-only tests (F07-T11). No apply until F-03 step 5.
4. **Portal.** Session state, header states, sign-in error route, privacy-notice link, and
   accessibility runs (F07-T10); the server-state library proposed with evidence.

Units 1 and 2 depend on D-06 (Security owner). Enabling contributor-record creation and any
sign-in by real people depends on D-20 (Legal, DPO); unit 4's privacy-notice link depends on the
approved notice. Registering the OAuth App (under `lksnext`, D-30) does not authorize deployment or real sign-ins.
Deployment stays blocked until F-03 step 5; D-02 does not authorize it.

## Open questions

- D-06: Security-owner decision on the repository owner's proposal (OAuth App for sign-in only;
  30 minutes idle and 8 hours absolute).
- D-20: Legal and DPO validation of the proposed lawful bases, the privacy notice, retention
  periods, recipients, and any international transfers.

Decided on 2026-10-22: D-02 (origins and callback URL), D-30 (the `lksnext` organization owns the
sign-in OAuth App), D-28 (suspended or revoked contributors may
sign in), and D-29 (minimal record at first sign-in, gated on D-20).
