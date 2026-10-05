# ADR-0004: Provider-neutral signing and configurable signature level

- Status: Accepted
- Date: 2026-10-05
- Decision-log entries: A-08; related open decisions D-09, D-11

## Context

Electronic signing is delegated to external providers, starting with DocuSign. The legal effect of
an electronic signature depends on the signature level (for example simple, advanced, or
qualified under eIDAS), the provider's service, and the jurisdiction. The platform must be able to
replace the provider without changing the domain logic.

## Decision

- The domain and application layers depend only on a provider-neutral `SigningProvider` port. Adapters
  (DocuSign first) live under `services/api/internal/adapters/signing/`.
- The platform never implements electronic-signature cryptography.
- The **required signature level** and **required evidence** (signed document, certificate of
  completion or audit trail, signer authentication method) are configuration per agreement type,
  stored as a versioned signing policy.
- Each provider adapter **declares** the signature levels its configured service actually
  delivers. Activating an agreement requires the delivered level to meet the required level.
- The platform records the delivered level as reported by the provider configuration. It displays
  "qualified" only when the configured provider service is a qualified service. Otherwise it uses
  the neutral term "electronic signature".
- The required levels are set by Legal (D-09). Until then, the configuration defaults to the
  provider's standard electronic signature and no qualified claim is made anywhere.

## Consequences

- Switching providers needs a new adapter and configuration, not domain changes.
- Contract tests per adapter prove it meets the port's behavior.
- UI copy, API responses, and evidence metadata must use the recorded level, never a hard-coded
  claim.
- A Legal change to the required level can make in-flight signing sessions invalid. The feature
  specification for Phase 6 defines how they are handled.

## Alternatives considered

- **Call the DocuSign SDK directly from application services.** Rejected: it couples the domain to
  one vendor.
- **Fixed signature level in code.** Rejected: the level is a legal decision that can change.
