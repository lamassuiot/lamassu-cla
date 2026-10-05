# Test-Scenario Catalogue

| Field | Value |
| --- | --- |
| Status | Draft — in review |
| Related | [Requirements](../requirements.md), [threat model](../security/threat-model.md) |

These are the minimum scenarios the platform must cover. Each feature specification maps its
acceptance criteria and automated tests to these IDs. Fixtures use synthetic identities and
documents only.

| ID | Scenario | Expected result | Requirements | Feature | Levels |
| --- | --- | --- | --- | --- | --- |
| TS-01 | Contributor signs the active ICLA successfully. | Agreement `ACTIVE`, documents stored with hashes, contributor `ACTIVE`, audit events written. | FR-ICLA-01 to 07, 10 | F-11 | Application, adapter contract, end-to-end |
| TS-02 | Contributor abandons signing. | No active agreement; session `ABANDONED` or `EXPIRED`; contributor can restart. | FR-ICLA-08 | F-11 | Application, end-to-end |
| TS-03 | Provider sends a duplicate callback. | Second callback has no effect; one set of audit events. | FR-ICLA-09, NFR-09 | F-11 | Handler, idempotency |
| TS-04 | Provider sends an invalid callback. | Rejected with `401`; nothing enqueued; metric incremented. | FR-ICLA-04 | F-10, F-11 | Handler, adapter contract |
| TS-05 | Signed-document download fails. | Agreement stays `VERIFYING`; retried; alarm after retry limit; never `ACTIVE`. | FR-ICLA-05, 06 | F-11 | Application, adapter |
| TS-06 | Agreement document stored but metadata update fails. | No `ACTIVE` state; retry completes the transaction idempotently; reconciliation reports orphans. | FR-ICLA-05, 06 | F-11 | Application, integration |
| TS-07 | Contributor with an active ICLA opens a PR. | Check Run `PASSED`. | FR-GH-05, 06 | F-15 | Application, end-to-end |
| TS-08 | Contributor with no CLA opens a PR. | Check Run `ACTION_REQUIRED` with signing link; no private data in output. | FR-GH-06 | F-15 | Application, end-to-end |
| TS-09 | Contributor covered by an active CCLA opens a PR. | Check Run `PASSED`. Variant: email domain matches the organization but the contributor is not on the list, so the result is `ACTION_REQUIRED`. | FR-CCLA-04, 06 | F-13, F-15 | Application |
| TS-10 | CCLA authorization is revoked. | Coverage ends; subsequent evaluations `ACTION_REQUIRED`; audit events written. | FR-CCLA-07 | F-13 | Application |
| TS-11 | GitHub webhook signature is invalid. | Rejected with `401`; not processed; metric incremented. | FR-GH-02 | F-14 | Handler |
| TS-12 | User requests another contributor's private documents. | `404`; access denial logged. | FR-CON-03 | F-08, F-11 | Authorization, handler |
| TS-13 | Maintainer approves a pending organization agreement. | CCLA and organization `ACTIVE`; audit events. | FR-CCLA-02 | F-12 | Application, end-to-end |
| TS-14 | Administrator revalidates a contributor. | `last_revalidated_at` updated; open PRs re-evaluated; audit event. | FR-CON-05, FR-GH-08 | F-13, F-15 | Application |
| TS-15 | Signing provider is unavailable. | Session creation returns `503` with `Retry-After`; no partial agreement; reconciliation recovers pending envelopes. | FR-SIGN-03 | F-10, F-11 | Adapter, application |
| TS-16 | Repository has CLA enforcement disabled. | No blocking Check Run (skipped or neutral per Phase 8 specification). | FR-GH-07, FR-POL-02 | F-16 | Application |
| TS-17 | Commit author cannot be resolved to a GitHub user. | Check Run `ACTION_REQUIRED`, never `PASSED`. | NFR-12 | F-15 | Application |
| TS-18 | Logs and Check Run output are inspected after flows TS-01 to TS-17. | No emails, names, tokens, or document contents present. | NFR-02 | All | Redaction tests |
