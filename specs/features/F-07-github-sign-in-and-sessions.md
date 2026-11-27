# F-07: GitHub sign-in and sessions

| Field | Value |
| --- | --- |
| Status | Approved by the repository owner on 2026-11-27 (commit `5185972` of PR #86, with the F07-T8 clarification); D-06 decided (Security owner, 2026-11-27, [ADR-0014](../../docs/decisions/0014-github-oauth-app-sign-in-and-sessions.md)); sign-in for real people and contributor-record creation also need D-20 (Legal, DPO) |
| Phase | 4 |
| Owner | Repository owner; Security owner for the authentication model (D-06) |
| Depends on | F-02, F-03, F-06; D-02, D-06, D-20, D-28, D-29, D-30; [security](../security/security.md#31-people-github-sign-in), [data protection](../security/data-protection.md), [threat model](../security/threat-model.md) (T-15, T-16) |
| Approved by | Repository owner, 2026-11-27 (units 1 and 2 may start; deferred tool decisions, deployment restrictions, the D-20 gate, and the C-03 follow-up stay as documented) |

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

- D-06 is decided by the Security owner (2026-11-27,
  [ADR-0014](../../docs/decisions/0014-github-oauth-app-sign-in-and-sessions.md)): a GitHub
  **OAuth App** used only for sign-in, with no repository permissions and no
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
   Sign-in is behind a configuration flag, off by default; while it is off (until D-20 and F-03
   step 5), `start` shows "sign-in is not yet available" and never redirects to GitHub.
2. The API validates `return_to` against the relative-path allowlist (falling back to `/`),
   creates a single-use `state` and a PKCE verifier (`S256`). If the browser still carries a
   pre-authentication cookie from an earlier attempt, that attempt's item is deleted first: a new
   attempt supersedes the previous one. It stores one `ephemeral` item for 10
   minutes, keyed by the hash of `state`, holding the verifier in plaintext (the code exchange
   needs the original value; the table is encrypted at rest with its KMS key) and the validated
   `return_to`. It sets the pre-authentication cookie `__Host-cla_oauth` (`HttpOnly`, `Secure`,
   `SameSite=Lax`, `Path=/`, no `Domain`, 10 minutes) whose value is bound to `state`, and
   redirects to GitHub's authorization page with `client_id`, the exact registered callback URL
   as `redirect_uri`, `state`, `code_challenge`, `code_challenge_method=S256`, and an explicitly
   empty `scope`. Any existing portal session is left untouched.
3. The person approves on GitHub. GitHub redirects to the registered callback,
   `GET /api/auth/github/callback`, with `code` and `state`, or with `error` and `state`. The
   callback stays outside the
   versioned API, so the registered URL does not change when the API version changes (D-02).
4. On every callback, success or error, the API processes the sign-in attempt in this order:
   1. Reject malformed input: duplicate parameters, both `code` and `error`, neither, or a
      missing `state`.
   2. Check the browser binding: the `__Host-cla_oauth` cookie must be present and bound to the
      `state` in the query. On a mismatch nothing is consumed, so a third party who learns a
      `state` cannot spend someone else's attempt.
   3. **Consume** `state` in one atomic operation: a conditional delete of `OAUTH#{state_hash}`
      that succeeds only if the item exists and `expires_at` is in the future (expiry is enforced
      here, not by TTL deletion, which may lag), and that returns the item. A failed delete
      (missing, expired, or already used) ends the attempt. Two concurrent callbacks with the same
      `state` complete at most one sign-in. Once consumed, any later failure requires a new
      attempt.
   4. Clear the pre-authentication cookie, on every outcome from here on.
   5. For `error`, show the cancelled or error result. For `code`, exchange `code` with
      `code_verifier`, the backend-held `client_secret`, and the same `redirect_uri` for a user
      access token. Validate that the token response explicitly reports an empty `scope`. Call
      authenticated `GET /user`, verify that its `X-OAuth-Scopes` response header is empty, and
      validate and extract `id` and `login`; missing or malformed scope metadata, or a missing or
      invalid `id` or `login`, fails the sign-in. Then discard the access token and any
      `refresh_token`, after a best-effort revocation (`DELETE /applications/{client_id}/token`)
      whose failure does not fail the sign-in. A non-empty granted scope ends sign-in without a
      session. No token is stored, logged, or sent to the browser.
5. The API creates or updates the minimal contributor record (D-29): the numeric GitHub user ID,
   which is the only stable identifier, and the current login, refreshed on every sign-in for
   display only (FR-CON-01). A new contributor starts as `PENDING`; an existing record keeps its
   status and agreement history, only `github_login` changes. The login is never used as an
   identifier or key. Only `id` and `login` are extracted from the `GET /user` response; other
   fields, including any public email address it contains, are ignored and are not persisted or
   logged, and no email endpoint is called. Email lookup values (D-19) are added by the feature
   that resolves commit authors (F-15, C-03).
6. The API creates the session **only now**: a 256-bit random token, stored only as a hash in the
   `ephemeral` table with the GitHub user ID, `created_at`, `last_seen_at`, `idle_expires_at`,
   `expires_at`, and a session-bound 256-bit random CSRF token, stored as is so that
   `GET /api/v1/session` can return the same value for the life of the session (it is useless
   without the session cookie, whose value is stored only hashed). In one transaction it writes the new session,
   deletes the browser's previous session if the request carried a valid one (rotation), and
   writes the `auth.signed_in` audit event; if the transaction does not commit, nothing changes and
   the previous session stays valid. It then sets the `__Host-cla_session` cookie (`HttpOnly`,
   `Secure`, `SameSite=Lax`, `Path=/`) and redirects to the validated `return_to`.
7. The portal calls `GET /api/v1/session`, which returns the signed-in person (GitHub user ID and
   login), the session expiry, and the CSRF token, with `Cache-Control: no-store`. The header shows
   the login and a **Sign out** button. This read does not count as activity (step 8).
8. Every authenticated request is checked by middleware: the session hash must exist and `now`
   must be before both `idle_expires_at` and `expires_at`. State-changing requests also need a
   matching `X-CSRF-Token` header and an allowed `Origin`. After these checks pass, an accepted
   request other than `GET /api/v1/session` may set `last_seen_at = now` and
   `idle_expires_at = min(now + 30 minutes, expires_at)`; the update is conditional on the session
   still existing and being unexpired, so an in-flight update cannot recreate a session after
   sign-out, and it is throttled to one write per minute, so persisted activity may lag by up to a
   minute.
9. **Sign out** calls `POST /api/v1/session/sign-out` (CSRF- and Origin-protected). The API deletes
   the session and writes the `auth.signed_out` audit event in one transaction, expires the
   cookie, and returns 204; the portal returns to the landing page in the signed-out state. Local
   sign-out ends the portal session only. It makes no further GitHub revocation request (the
   access token was already revoked best-effort in step 4) and does not remove the person's
   authorization grant for the OAuth App on GitHub.

## Alternative flows

- **Already signed in.** Starting sign-in leaves the current session untouched. On a successful
  callback the new session replaces the old one atomically (step 6). Cancellation and failures
  before session replacement preserve the existing valid session. If replacement commits but its
  result or the response is lost, the previous session may already be invalidated and the person
  may need to sign in again. A session token is never reused across sign-ins.
- **Second attempt in another tab.** The new attempt supersedes the first: the old tab's callback
  fails the binding check and shows "sign-in attempt expired or replaced".
- **Session expires.** After 30 minutes without activity or 8 hours after sign-in, whichever comes
  first (D-06), the API answers 401 and expires the cookie. "No silent renewal" means no automatic
  re-authentication after expiry; sliding idle expiry during activity (step 8) is intended. The
  person signs in again; the portal shows the signed-out state and offers sign-in with `return_to`
  set to the current page. Timeouts are configuration (security, section 3.1).
- **Suspended or revoked contributor.** Sign-in still succeeds (D-28, decided 2026-10-22).
  Authentication only confirms identity: it grants no permissions and does not restore the
  validity of any CLA. Every protected operation evaluates the current contributor status and CLA
  rules itself.
- **GitHub login changed.** The GitHub user ID is the identity; the stored login is refreshed on
  every sign-in, and status and agreement history are unchanged.
- **Sign-out without a session.** Returns 204 and expires any cookie (idempotent); the `Origin`
  check still applies. With a valid session, the matching CSRF token is required.

## Error cases

"No new session" below means that an existing valid session is preserved, except where the session
replacement transaction committed but its result was lost (see that row).

| Case | Detection | System behavior | User-visible result |
| --- | --- | --- | --- |
| Sign-in disabled by configuration | Flag check in `start` and the callback | No redirect to GitHub; no session | "Sign-in is not yet available" |
| Malformed callback: duplicate parameters, both `code` and `error`, neither, or no `state` | Query parsing | Rejected before any lookup; no exchange; no new session | Generic sign-in failure with a link to start a new attempt |
| Missing or mismatched pre-authentication cookie | Binding check before consumption | Nothing consumed; no exchange; no new session; sanitized security event | "Sign-in attempt expired or replaced", with a link to start a new attempt |
| Missing, expired, or reused `state`, on a success or an error callback | Atomic consume of `OAUTH#{state_hash}` | No exchange; no cancelled message; no business audit event; sanitized security event | Generic sign-in failure with a link to start a new attempt |
| Person declines on GitHub | `error=access_denied`, after binding check and consumption | Cookie cleared; no new session | Cancellation message; the existing signed-in or signed-out state is preserved |
| Granted scope is not empty, or scope metadata is missing or malformed | `scope` in the token response, then `X-OAuth-Scopes` on the `GET /user` response | Token discarded (best-effort revoked); no new session; security event `scope_unexpected` | Generic failure with a support reference |
| GitHub authorization, token exchange, or `GET /user` fails, or `id` or `login` is missing or invalid | Upstream error, timeout, or response validation | No new session; received credentials discarded; no automatic callback replay | Failure page with a link to start a new attempt |
| `return_to` not allowed | Canonicalization and allowlist check | Falls back to `/` | Lands on the landing page |
| Contributor upsert or session transaction fails or its outcome is uncertain | Store error or timeout | No session cookie is set; the previous session is preserved unless the transaction committed, in which case the person simply signs in again | Error page; restart sign-in when the service recovers |
| Missing, unknown, or expired session | Middleware | 401 problem response; a stale session cookie is expired | Signed-out state with a sign-in prompt |
| Missing or wrong CSRF token, or missing, `null`, or disallowed `Origin` | Middleware on state-changing requests | 403 problem response with a distinct `code` for each reason; no business operation and no activity extension | Error message; re-reading the session fixes a stale CSRF token, not an `Origin` rejection |
| Session storage unavailable | Store error | 503 problem response; the protected operation does not run; the error is never read as "no session" | Temporary service error; the portal keeps its last known authentication state |
| Sign-out deletion fails or its outcome is uncertain | Store error or timeout | No 204; 503 problem response; the session may still exist | "Could not complete sign-out. Try again." |

## Security considerations

- Follows the [security specification](../security/security.md#31-people-github-sign-in), section
  3.1, and threats T-15 (session theft or CSRF) and T-16 (open redirect).
- PKCE with `S256` is always used (GitHub accepts only `S256`): `code_challenge` and
  `code_challenge_method` on the authorization request, `code_verifier` on the exchange.
- `state` is single use, expires after 10 minutes, and is stored only as a hash, because it is
  only compared. The browser binding is checked first and the item is then consumed atomically
  (conditional delete with the expiry condition, returning the item) before any other callback
  processing, on success and error callbacks alike, so a replayed or forged callback can neither
  complete a sign-in nor show a misleading result, and a leaked `state` cannot spend another
  browser's attempt. The PKCE verifier cannot be hashed: the code exchange needs the original value. It
  is stored in plaintext inside the same encrypted, TTL-bounded `ephemeral` item, which is deleted
  before the exchange, and it never appears in logs, responses, or cookies. On its own it is
  useless without the authorization `code`, which only transits in the browser redirect.
- The pre-authentication cookie `__Host-cla_oauth` is `HttpOnly`, `Secure`, `SameSite=Lax`,
  `Path=/`, without `Domain`, and lives 10 minutes. `Lax` lets the browser send it on GitHub's
  top-level `GET` redirect to the callback. It is cleared on every outcome of the attempt.
- Session tokens are 256-bit random values from a cryptographic source, stored only as SHA-256
  hashes; comparisons of secrets are constant time. The CSRF token is a separate 256-bit random
  value bound to the session and stored in the encrypted session item as is, because the portal
  must be able to read it again; on its own it grants nothing, since the session cookie is only
  stored hashed. It changes whenever the session is replaced. A new
  session is created, and the old one deleted, only after authentication and the required writes
  have succeeded, in one transaction (session rotation on authentication).
- No GitHub token is persisted, logged, or returned, and any `refresh_token` is discarded. The
  access token is revoked best-effort right after `GET /user`. The OAuth App sends an explicitly
  empty `scope` and extracts only `id` and `login` from `GET /user`, which needs no scope; other
  fields, including any public email address, are ignored, and no email endpoint is called. It
  has no repository permissions and no email access (D-06, ADR-0014). Because GitHub grants
  previously authorized scopes when a request provides no `scope`, the empty parameter is not
  relied on: the API verifies that the granted scope is empty and refuses the sign-in otherwise.
  It is separate from the GitHub App, so a sign-in token never carries repository or Checks
  permissions.
- The registered callback URL is sent unchanged as `redirect_uri` in both the authorization and
  the token-exchange request; GitHub rejects a mismatch. The OAuth App registration must have
  callback-URL wildcard matching disabled (a per-app GitHub setting; D-30 admins).
- Identity is the numeric GitHub user ID. The login can change and be reused by another account, so
  it is never used as an identifier, key, or authorization input (D-29).
- Signing in grants no permissions and does not restore an agreement's validity (D-28). A session
  by itself authorizes only `GET /api/v1/session` and `POST /api/v1/session/sign-out`. Every other
  operation is authorized by F-08 roles or by the object-relationship checks in the security
  specification (section 2), never by the mere existence of a session. A `PENDING` contributor
  may use their own signing flow (F-11); maintainer and administrator functions stay closed until
  F-08 assigns a role.
- `return_to` accepts only relative paths from an allowlist. The value is percent-decoded once and
  rejected if it still contains a percent sign, has a scheme or authority, starts with `//` or
  `/\`, contains a backslash or a control character, or does not start with `/`; the canonical
  path is then compared with the allowlist. Anything rejected falls back to `/`.
- Session and authentication responses carry `Cache-Control: no-store`. Logs, traces, metrics, and
  problem responses contain no callback query strings, codes, `state` values, session tokens, CSRF
  tokens, cookies, or personal data; the request ID correlates them. API Gateway access logs must
  not record query strings (unit 3). Audit events record the GitHub user ID as the
  actor, as the domain model allows.
- Audit events are written in the same transaction as the session change; a sign-in or sign-out
  whose audit write fails is not reported as successful.
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
| `GET /api/v1/auth/github/start` | Start sign-in; query `return_to` (relative path) | 302 to GitHub; 400 for malformed input; 503 while sign-in is disabled or when the `ephemeral` store is unavailable |
| `GET /api/auth/github/callback` | Complete sign-in; query `code`, `state`, or `error`. Unversioned: it is the URL registered with GitHub (D-02) | 302 to `return_to` with the session cookie; 302 to a portal error route on every failure, including store failures |
| `GET /api/v1/session` | Current session (`sessionCookie`); `Cache-Control: no-store`; does not extend idle expiry | 200 with `github_user_id`, `github_login`, `expires_at`, `csrf_token`; 401; 503 when the store is unavailable |
| `POST /api/v1/session/sign-out` | End the session (`sessionCookie` and `csrfToken`; `Origin` checked) | 204 (also without a session); 403; 503 when the deletion fails or is uncertain |

## Data-model changes

- `ephemeral` table (domain model, section 5.3): the existing `OAUTH#{state_hash}` record gains the
  PKCE verifier (plaintext; the table is encrypted at rest and the item is deleted on use) and the
  validated `return_to`; the existing `SESSION#{session_token_hash}` record is
  specified (GitHub user ID, `created_at`, `last_seen_at`, `idle_expires_at`, `expires_at`, CSRF
  token in plaintext). Both use TTL on `expires_at` as cleanup only; expiry is enforced in the conditional
  operations and readers treat expired items as absent. Session creation, deletion of the replaced
  session, and the audit event are one `TransactWriteItems`.
- `core` table: contributor records are created on first sign-in (`PENDING`), and `github_login`
  is refreshed on later sign-ins; status and agreement history are never changed by sign-in. No
  new attributes.
- `audit` table: events `auth.signed_in` and `auth.signed_out`, written in the same transaction as
  the session change; a failed audit write fails the sign-in or sign-out.

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
  page with a new single-use `state`, a `S256` PKCE challenge, an explicitly empty
  `scope`, and the exact registered callback URL as `redirect_uri`, and sets the `__Host-cla_oauth`
  cookie (`HttpOnly`, `Secure`, `SameSite=Lax`, `Path=/`, 10 minutes) bound to that `state`,
  leaving any existing session untouched.
- **AC-07-2** Given a callback, success or error, whose `state` is missing, expired, already used,
  or not bound to the browser's `__Host-cla_oauth` cookie, or whose query is malformed, no code
  exchange happens, no session is created, no cancelled message is shown, and any existing session
  is preserved; a binding mismatch consumes nothing; two concurrent callbacks with the same
  `state` complete at most one sign-in; the cookie is cleared on every outcome after consumption.
- **AC-07-3** Given a successful callback, a contributor record exists for the GitHub user ID with
  the current login, a session exists whose token is stored only as a hash, the `__Host-cla_session`
  cookie is `HttpOnly`, `Secure`, `SameSite=Lax`, and `Path=/`, and the browser is redirected to the
  validated `return_to`.
- **AC-07-4** Given any `return_to` that is not an allowed relative path (absolute, protocol-relative,
  backslash, double-encoded, or other encoded variants), the redirect goes to `/`.
- **AC-07-5** No GitHub access token, refresh token, authorization code, or email address appears in
  storage, logs, metrics, audit events, response bodies, or cookies. The session token appears
  only in the `Set-Cookie` header of the callback response and in the session cookie, and is
  stored only as a hash. `state` appears only in the redirect to GitHub, in the callback query, and
  in the pre-authentication cookie binding, and is stored only as a hash. The PKCE verifier appears
  only in the `OAUTH#{state_hash}` item until it is used. The CSRF token appears only in the
  `GET /api/v1/session` body, the `X-CSRF-Token` request header, and the session item. None of these
  values appears in logs, metrics, traces, or audit events.
- **AC-07-6** Given a session past its idle or absolute expiry, or an unknown token, authenticated
  operations return 401 and expire the cookie. Given an accepted request other than
  `GET /api/v1/session`, `idle_expires_at` becomes `min(now + 30 minutes, expires_at)` at most once
  per minute, and never for a session that was deleted meanwhile.
- **AC-07-7** Given a state-changing request without the session's CSRF token or with a missing,
  `null`, or disallowed `Origin`, the API returns 403 with a distinct problem `code` per reason,
  runs no business operation, and extends no session.
- **AC-07-8** Given sign-out, the session is deleted and the `auth.signed_out` audit event is
  created in the same transaction, the cookie is
  expired, and the old token no longer authenticates; if the deletion fails or is uncertain, the
  API returns 503, not 204; without a session, sign-out returns 204. Audit events are never
  deleted.
- **AC-07-9** Given a signed-in browser that completes a new sign-in, the new session replaces the
  old one in one transaction and the old token no longer authenticates; given an attempt that is
  cancelled or fails before session replacement, the old session stays valid.
- **AC-07-10** The portal shows the signed-out and signed-in header states and the sign-in error
  route, with no WCAG 2.2 A or AA violations in either theme at the F-06 widths.

- **AC-07-11** Given a contributor whose status is `SUSPENDED` or `REVOKED`, sign-in succeeds, and
  the session carries no permission and changes no agreement or coverage state (D-28).
- **AC-07-12** Given two sign-ins by the same numeric GitHub user ID with different logins, one
  contributor record exists with the latest login; given a sign-in whose login matches another
  record's former login, no record is matched by login (D-29).
- **AC-07-13** Given a token response whose `scope` is not an explicit empty value, or a `GET /user`
  response whose `X-OAuth-Scopes` header is missing or not empty, the token is discarded, no
  session is created, and a `scope_unexpected` security event is logged without personal data.
- **AC-07-14** Given a `GET /user` response, only `id` and `login` are extracted; every other field,
  including a public `email`, is absent from storage, logs, and audit events; a missing or invalid
  `id` or `login` fails the sign-in; any `refresh_token` is discarded; the access token is revoked
  best-effort.
- **AC-07-15** Given the sign-in flag is off, `start` returns 503 without redirecting to GitHub and
  the callback creates no session.
- **AC-07-16** Given session storage is unavailable, protected operations return 503 and are not
  executed, and the response is never a 401.

## Automated test scenarios

| ID | Scenario | Level | Covers |
| --- | --- | --- | --- |
| F07-T1 | Sign-in start: redirect parameters, stored hashed `state`, stored verifier and `return_to`, pre-authentication cookie. | Unit (use case with fakes), handler | AC-07-1 |
| F07-T2 | Callback with missing, expired, reused, and mismatched `state`, on success and error callbacks, and two concurrent callbacks with the same `state` against a fake store with conditional delete: at most one exchange, no session otherwise. | Unit, handler | AC-07-2 |
| F07-T3 | Successful callback against a local fake of GitHub's endpoints: `S256` challenge and verifier pair, `redirect_uri` on both requests, contributor upsert, hashed session, cookie attributes, pre-authentication cookie cleared, redirect; with an existing session, the old token stops working only after success; with a cancelled attempt, it keeps working. | Adapter and handler (local test server) | AC-07-3, AC-07-9 |
| F07-T4 | `return_to` table test with relative, absolute, protocol-relative, backslash, control-character, single- and double-encoded inputs. | Unit | AC-07-4 |
| F07-T5 | Logs, audit events, stored items, response bodies, and cookies from the flows above are scanned for the access token, refresh token, code, `state`, PKCE verifier, session token, and CSRF token; each may appear only where AC-07-5 allows (the session token only in `Set-Cookie`, `state` only in the GitHub redirect and callback, the CSRF token only in the session body and header); the `OAUTH#{state_hash}` item is gone after the callback. | Unit, handler | AC-07-5 |
| F07-T6 | Middleware: idle and absolute expiry with a fake clock; unknown token; sliding `idle_expires_at` capped by `expires_at`, throttled to one write per minute, skipped for `GET /api/v1/session`, and not applied to a session deleted between check and update. | Unit | AC-07-6 |
| F07-T7 | CSRF and `Origin` checks on a state-changing route. | Handler | AC-07-7 |
| F07-T8 | Sign-out deletes the session and creates `auth.signed_out` in one transaction and expires the cookie; the old token returns 401; the audit event remains. A confirmed transaction failure yields 503, preserves the session, and creates no audit event. A transaction that committed before its response was lost (faked timeout after commit) yields 503 while the session is deleted and the audit event exists. Sign-out without a session yields 204. | Handler | AC-07-8 |
| F07-T9 | OpenAPI contract tests: every new operation's responses validate against `openapi.yaml`. | Contract | All API criteria |
| F07-T10 | Portal header states and sign-in error route, axe and reflow at the F-06 widths, with the API mocked. | Component (browser) | AC-07-10 |
| F07-T11 | Terraform plan-only tests with mocked providers for the new API, Lambda, tables, and secret. | Infrastructure | Rollback and least privilege |
| F07-T12 | Sign-in of suspended and revoked contributors, and login changes and reuse, with fakes. | Unit | AC-07-11, AC-07-12 |
| F07-T13 | Token exchange against the local GitHub fake returning a non-empty `scope` or `X-OAuth-Scopes`: token discarded, no session, `scope_unexpected` logged; the fake also asserts that `redirect_uri` and an empty `scope` were sent. | Adapter and handler (local test server) | AC-07-1, AC-07-13 |
| F07-T14 | `GET /user` fake returning a public `email`, extra fields, a missing `id`, a non-numeric `id`, and a `refresh_token`: only `id` and `login` are kept, invalid identity fails, the refresh token is never stored, the revocation call is made and its failure is tolerated. | Adapter (local test server) | AC-07-14 |
| F07-T15 | Sign-in flag off: `start` returns 503 and the callback refuses; flag on: normal flow. | Handler | AC-07-15 |
| F07-T16 | Store faked as unavailable: protected operations return 503, not 401, and do not run. | Handler | AC-07-16 |
| F07-T17 | Malformed callbacks (duplicate parameters, `code` and `error` together, neither, no `state`) are rejected before any store access. | Handler | AC-07-2 |

## Observability requirements

- Structured log events (no personal data, request ID included): `auth.start`, `auth.callback`
  with outcome (`success`, `state_invalid`, `denied`, `scope_unexpected`, `upstream_error`), `auth.sign_out`,
  `session.rejected` with reason (`unknown`, `idle_expired`, `expired`, `csrf`, `origin`,
  `store_unavailable`); `auth.callback` outcomes also include `binding_mismatch`, `malformed`, and
  `disabled`. Callback query strings are never logged.
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
   *Delivered:* `specs/api/openapi.yaml` (server base `/api`, four operations, `Session` schema)
   and the regenerated client; `services/api/internal/domain` (`secret.go`, `session.go`,
   `contributor.go`, `returnto.go`, `audit.go`); `services/api/internal/ports/auth.go`;
   `services/api/internal/application` (`StartSignIn`, `CompleteSignIn`, `Authenticate`,
   `VerifyCSRF`, `SignOut`). The audit event travels with the session change through the
   session-store port rather than a separate audit-writer port, so that one transaction covers
   both (step 6). Tests with fakes cover F07-T1, F07-T2 (including 16 concurrent callbacks),
   F07-T4, F07-T5, F07-T6, F07-T12, F07-T15, F07-T16, F07-T17, and the use-case parts of F07-T3,
   F07-T8, F07-T13, and F07-T14; the adapter and handler parts follow in unit 2.
2. **Adapters and HTTP.** GitHub identity adapter tested against a local fake server; DynamoDB
   adapters (test approach proposed with evidence, for example a local DynamoDB container in the
   pinned CI image); handlers, middleware, Lambda wiring, and a local development server (F07-T3,
   F07-T7 to F07-T9).
3. **Infrastructure (plan-only).** API Gateway HTTP API, the `api` Lambda, the `core`, `audit`, and
   `ephemeral` tables, the Secrets Manager secret (value set by LKS organization admins; readable only by the authentication backend), and
   least-privilege IAM, with plan-only tests (F07-T11). No apply until F-03 step 5.
4. **Portal.** Session state, header states, sign-in error route, privacy-notice link, and
   accessibility runs (F07-T10); the server-state library proposed with evidence.

Units 1 and 2 depend on this specification's approval by the repository owner (D-06 is decided). Enabling contributor-record creation and any
sign-in by real people depends on D-20 (Legal, DPO); unit 4's privacy-notice link depends on the
approved notice. Registering the OAuth App (under `lksnext`, D-30) does not authorize deployment or real sign-ins.
Deployment stays blocked until F-03 step 5; D-02 does not authorize it.

## Open questions

- D-20: Legal and DPO validation of the proposed lawful bases, the privacy notice, retention
  periods, recipients, and any international transfers.
- C-03 (`specs/STATE.md`): where verified emails are read, now that sign-in has no email scope.
  Out of scope for F-07; it concerns the agreement and validation flows.

Decided on 2026-11-27: D-06 (sign-in-only OAuth App; 30 minutes idle and 8 hours absolute;
[ADR-0014](../../docs/decisions/0014-github-oauth-app-sign-in-and-sessions.md)).
Decided on 2026-10-22: D-02 (origins and callback URL), D-30 (the `lksnext` organization owns the
sign-in OAuth App), D-28 (suspended or revoked contributors may
sign in), and D-29 (minimal record at first sign-in, gated on D-20).
