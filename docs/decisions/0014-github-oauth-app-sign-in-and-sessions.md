# ADR-0014: GitHub OAuth App sign-in and server-side sessions

- Status: Accepted
- Date: 2026-11-27
- Decision-log entry: D-06; related D-08, D-19, D-28, D-29, D-30

## Context

People sign in to the portal with GitHub. The [security specification](../../specs/security/security.md#31-people-github-sign-in)
left the model choice open (D-06): a GitHub OAuth App, or GitHub App user-to-server tokens, and
the session model. The repository owner proposed a model on 2026-10-22 and the Security owner
reviewed it. The platform handles legal evidence and personal data, so the sign-in path must
request the least identity data that still identifies a contributor stably.

## Decision

The Security owner confirmed the proposal on 2026-11-27 (confirmation reported by the repository
owner):

1. **App type.** A dedicated GitHub **OAuth App**, used only for portal sign-in. GitHub App
   user-to-server tokens are not used for people.
2. **Scope.** The OAuth App has no repository permissions and no email access. During sign-in the
   API reads only the numeric GitHub user ID and the current login, then discards the OAuth token.
   The numeric ID is the stable identity; the login is display data (D-29).
3. **Separation.** A different GitHub App, installed on repositories, receives pull-request events
   and publishes Checks with the minimum permissions (D-08). Both apps are owned by the `lksnext`
   organization (D-30).
4. **Session model.** Sessions are server-side. The browser holds an opaque 256-bit random token in
   a `__Host-` cookie (`HttpOnly`, `Secure`, `SameSite=Lax`); DynamoDB stores only the token's
   hash. A session expires after 30 minutes of inactivity or 8 hours after sign-in, whichever comes
   first, with no silent renewal; the person signs in again. State-changing requests require the
   session-bound `X-CSRF-Token` header and an `Origin` check.

The authorization-code flow, `state` handling, exact redirect-URI registration, and the
`return_to` allowlist in the security specification are unchanged. PKCE is used where GitHub
supports it for OAuth Apps. The timeout values are held as configuration with the decided
defaults.

## Consequences

- F-07 can be approved and implemented against a fixed authentication model; D-20 (Legal, DPO)
  still gates sign-in by real people and contributor-record creation.
- Because the OAuth App has no email scope, verified emails cannot be read at sign-in. The
  [data-protection notes](../../specs/security/data-protection.md#4-email-lookup-values-d-19)
  (section 4.1 and section 4.4 step 3) assume that they are. This is recorded as contradiction
  C-03 in `specs/STATE.md`; its resolution (a later, explicit step that requests the email scope,
  for example when a contributor starts an agreement or links commit emails) needs the Security
  owner and the DPO (D-19) and does not block F-07.
- The threat model entries T-15 (session theft, CSRF) and T-16 (open redirect) are unchanged.

## Alternatives considered

- **GitHub App user-to-server tokens.** One app for people and repositories. Rejected: it couples
  sign-in to the repository-facing app's permissions and installation lifecycle, and offers no
  benefit for an identity-only flow.
- **Stateless signed session tokens (JWT in a cookie).** Rejected: server-side revocation on
  sign-out and role change would need a denylist anyway, and the hash-only store keeps no secret
  material at rest.
- **Silent session renewal.** Rejected by the proposal and confirmed: fixed idle and absolute
  limits bound the exposure of a stolen cookie (T-15).
