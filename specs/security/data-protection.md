# Data-Protection Notes

| Field | Value |
| --- | --- |
| Status | Draft — requires Legal and DPO review |
| Related | [Security specification](security.md), [ADR-0005](../../docs/decisions/0005-configurable-object-lock-retention.md), decisions D-10, D-19, D-20 |

These notes support a data-protection assessment. They are not legal advice and do not decide
legal questions. Items marked **Legal decision required** must be confirmed before the dependent
phase starts.

## 1. Roles

| Question | Status |
| --- | --- |
| Which legal entity is the controller for contributor and signer data? | **Legal decision required** |
| Processors and sub-processors (for example AWS, the signing provider, GitHub) and the agreements in place with them. | **Legal decision required** |
| Whether a data-protection impact assessment is required. | **DPO decision required** |
| Lawful basis per processing purpose and the privacy notice shown at sign-in and signing (D-20). | **Legal decision required** |
| International transfers resulting from provider and hosting locations (relates to D-04 and D-11). | **Legal decision required** |

## 2. Data inventory

| Data | Subjects | Purpose | Store | Protection |
| --- | --- | --- | --- | --- |
| GitHub user ID and login | Contributors | Identity and coverage | DynamoDB `core` | KMS encryption |
| Verified email lookup values (keyed HMAC) | Contributors | Matching commit emails to identities | DynamoDB `core` | Keyed HMAC in AWS KMS; raw email never persisted (section 4) |
| Signer name | Signers | Required by the signing provider | DynamoDB `core` (encrypted field), provider | Field-level encryption |
| Signer email | Signers | Required by the signing provider | Not persisted by the platform; passed to the provider at envelope creation; lookup value stored as keyed HMAC | Section 4 |
| Representative's stated authority | Organization representatives | CCLA validity | DynamoDB `core` | KMS encryption |
| Signed documents and evidence | Signers | Legal evidence | S3 `evidence` | Object Lock, KMS, private |
| Audit events | Actors | Accountability | DynamoDB `audit`, S3 `evidence` | Append-only, KMS |
| Session and OAuth state hashes | Users | Authentication | DynamoDB `ephemeral` | TTL, hashed |
| Access logs | Users | Security monitoring | CloudWatch Logs | Retention limits, no personal data beyond IDs |

## 3. Minimization rules

- Collect only what the agreement type and signing provider require.
- Store identifiers, not profiles. Do not store GitHub avatars, bios, or organizations beyond what
  coverage needs.
- Keep personal data out of object keys, object metadata, log messages, metrics, Check Run output,
  and audit `metadata`.
- Prefer keyed hashes for lookup data. Plain hashes of emails are not used because emails are
  guessable.

## 4. Email lookup values (D-19)

Technical default approved on 2026-10-05, **subject to DPO confirmation**.

### 4.1 Rules

- Raw email addresses are **never persisted** by the platform: not in DynamoDB, S3 objects or
  metadata, queue messages, caches, or build artifacts.
- Raw email addresses are **never written** to logs, metrics, traces, audit events, error
  responses, or Check Run output.
- Raw emails exist only in memory for the duration of the request that needs them: reading
  verified emails at sign-in, resolving commit authors during validation, or passing signer
  emails to the signing provider when an envelope is created.
- Queue messages carry identifiers only. Workers re-read emails from GitHub or the provider when
  they need them.
- The persisted lookup value is a keyed HMAC. HMAC values are pseudonymous personal data: they are
  protected like other personal data and are not logged or published.

### 4.2 Computation

1. Normalize: trim surrounding whitespace, apply Unicode NFC, and convert the whole address to
   lowercase. Do not apply provider-specific rewriting such as removing dots or `+` suffixes.
2. Compute HMAC-SHA-256 with an AWS KMS HMAC key (`HMAC_256` key spec, `GenerateMac`).
3. Store as `{key_version}:{hex_mac}`, where `key_version` identifies the KMS key used.
4. Compare stored values by exact match. To check a candidate email, compute its MAC with the
   relevant key version; never attempt to recover the email.

### 4.3 Key storage

- One customer-managed KMS HMAC key per environment and key version. The key material is generated
  in and never leaves AWS KMS.
- The key policy grants `GenerateMac` only to the functions that need it (`api` for sign-in,
  `github-worker` for commit-author resolution). No role receives export, decrypt, or
  administrative permissions on the key at runtime.
- `ScheduleKeyDeletion`, `DisableKey`, and key-policy changes are denied to application and
  deployment roles. Only a break-glass administrative role can perform them.
- CloudTrail records all key usage; alarms fire on `ScheduleKeyDeletion`, `DisableKey`, and
  key-policy changes.
- The current key version is a configuration value passed to the functions, never hard-coded.

### 4.4 Rotation

AWS KMS does not rotate HMAC keys automatically, so rotation is a documented manual procedure:

1. Create a new HMAC key version through Terraform.
2. Make the new version current for new MACs. During the transition window, lookups compute MACs
   with both the current and the previous version.
3. Re-key lazily: when a contributor signs in, their verified emails are read from GitHub and
   stored MACs are replaced with current-version MACs. Raw emails are not stored, so offline
   re-keying is impossible by design.
4. At the end of the transition window, remove remaining previous-version MACs. Affected
   contributors keep their identity (GitHub user ID) and coverage; only commit-email lookup is
   lost until their next sign-in.
5. Disable the previous key, then schedule its deletion with the maximum 30-day waiting period.

Triggers: suspected key misuse (immediate rotation) or a scheduled interval. The interval and
transition window are confirmed by the Security owner (proposed: annual rotation, 90-day window).

### 4.5 Recovery

- KMS key material cannot be exported or backed up. Protection relies on the deletion controls in
  section 4.3, the 30-day deletion waiting period, and Terraform `prevent_destroy`.
- For regional disaster recovery, a KMS multi-Region replica of the HMAC key may be created in a
  second EU region. This is optional and decided with D-16.
- If a key is lost despite these controls: email lookup values for that version become unusable.
  Contributor identity and agreements are unaffected because they are keyed by GitHub user ID.
  Lookup values are rebuilt as contributors sign in. Until then, commit authors that cannot be
  resolved through GitHub's own user mapping yield `ACTION_REQUIRED`, never `PASSED` (INV-10).
- If misuse is suspected: rotate immediately, review CloudTrail `GenerateMac` usage, and follow the
  incident process in section 7.

## 5. Retention and erasure — Legal decision required

Object Lock retention is configurable ([ADR-0005](../../docs/decisions/0005-configurable-object-lock-retention.md)).
Legal must decide the following before production infrastructure is deployed:

1. Retention period for signed agreements and evidence, per agreement type, counted from which
   event (signature, revocation, end of the contribution's use).
2. Object Lock mode in production: **governance** (privileged bypass possible) or **compliance**
   (irreversible for the retention period).
3. Retention period for audit events and audit exports.
4. Retention for contributor records after their agreement ends.
5. How erasure requests are handled for data under retention, including the response wording.
6. Whether legal holds are needed beyond retention, and who may place or remove them.
7. Retention for logs and operational data.

Until these are confirmed:

- Non-production environments use governance mode and short retention.
- Production Terraform has no default retention values and refuses to apply without them.

## 6. Data-subject requests

The process for access, rectification, erasure, and objection requests is defined with Legal and
the DPO before production. Technical support planned:

- Export of a contributor's own records through an authorized administrative operation.
- Pseudonymization of DynamoDB records where erasure is required and evidence must be kept.
- Audit events recording each request and its handling, without storing the request contents.

## 7. Breach handling

Security incidents follow [`SECURITY.md`](../../SECURITY.md) and the incident-response guide added in
Phase 9. Notification obligations and timelines are a Legal decision.
