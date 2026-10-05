# ADR-0008: Hexagonal Go backend on Lambda with asynchronous workers

- Status: Accepted
- Date: 2026-10-05
- Decision-log entry: D-17

## Context

The reference architecture requires a Go backend on AWS Lambda behind API Gateway HTTP APIs, with
domain logic independent of AWS SDK details. The backend receives three kinds of traffic:

- Interactive requests from the portal.
- GitHub webhooks, which GitHub times out after 10 seconds.
- Signing-provider callbacks, after which signed documents must be downloaded and stored.

Downloading documents, writing to S3, and creating GitHub Check Runs can be slow or fail
transiently. Doing them inside the webhook request risks timeouts and duplicate retries.

## Decision

Accepted:

- **Hexagonal architecture.** `domain` (entities, invariants, state machines) and `application`
  (use cases) depend only on interfaces in `ports`. Adapters for AWS, GitHub, and signing providers
  implement the ports. `handlers` translate HTTP or queue events into use-case calls.
- **Lambda functions split by trust boundary**, all built from one Go module:

  | Function | Trigger | Purpose |
  | --- | --- | --- |
  | `api` | API Gateway HTTP API | Portal, public catalogue, and administrative APIs. |
  | `github-webhook` | API Gateway | Verify signature, record delivery, enqueue. Respond quickly. |
  | `signing-webhook` | API Gateway | Verify callback, record event, enqueue. Respond quickly. |
  | `signing-worker` | SQS | Fetch envelope status, download documents, store evidence, activate agreements. |
  | `github-worker` | SQS | Evaluate CLA coverage and create or update Check Runs. |
  | `scheduler` | EventBridge Scheduler | Revalidation, signing-session expiry, reconciliation. |

- **Amazon SQS** queues with dead-letter queues between webhooks and workers. Not in the original
  reference architecture; approved for reliability.
- **Retries and visibility timeouts.** Each worker queue's visibility timeout is at least six times
  the consuming function's timeout. Messages move to the dead-letter queue after a configured
  maximum receive count. Workers report partial batch failures so successful messages are not
  retried. Exact values are set in the feature specifications and Terraform variables.
- Each function has its own IAM role with least privilege.

## Consequences

- Webhooks return within GitHub's and the provider's timeouts regardless of downstream latency.
- Retries are handled by SQS. Processing must be idempotent because SQS delivers at least once.
- Dead-letter queues need alarms and a documented redrive procedure.
- More functions mean more Terraform resources, mitigated by a shared Lambda module.

## Alternatives considered

- **Single Lambda for everything.** Simpler, but one IAM role holds every permission and webhook
  latency depends on downstream calls.
- **Function per endpoint.** Fine-grained IAM, but many deployables and cold starts.
- **AWS Step Functions for signing workflows.** Useful for long multi-step flows; reconsider in Phase
  7 if CCLA multi-signer flows need it.
