# SDD Discovery Report

| Field | Value |
| --- | --- |
| Status | Approved on 2026-10-05 by the repository owner |
| SDD phase | Phase 1: Discovery and specification foundation |
| Date | 2026-10-05 |
| Base commit | `89f9d40` (`main`) |

This report records the current state of the repository before any specification or production
code is written. It is the entry point for the Specification-Driven Development (SDD) workflow:

```text
requirements → architecture → feature specification → acceptance criteria → implementation plan
  → implementation → tests → security review → documentation update → pull request
```

No production code is proposed or written in this phase.

## 1. Repository inventory

The repository was instantiated from the Lamassu open-source repository template (PRs #2 and #3).
It contains **governance and community-health files only**. There is no application code, no
dependency manifest, no infrastructure code, and no specification yet.

| Area | Present | Notes |
| --- | --- | --- |
| Governance | `AGENTS.md`, `CODE_OF_CONDUCT.md`, `CONTRIBUTING.md`, `SECURITY.md`, `LICENSE`, `NOTICE` | `AGENTS.md` still describes the repository as a reusable template. |
| GitHub config | `.github/CODEOWNERS`, `dependabot.yml`, issue forms, PR template, `copilot-instructions.md` | CODEOWNERS: `@lamassuiot/lamassu-maintainers`. |
| Workflows | `ci.yml`, `lint-pr-title.yml`, `release.yml`, `scorecard.yml` | All actions are pinned by SHA. |
| Release tooling | `cliff.toml`, `CHANGELOG.md` | Tagged `vX.Y.Z` releases, squash-merge, Conventional Commits. |
| Editor/lint | `.editorconfig`, `.markdownlint.json`, `.markdownlintignore` | 2-space indent, LF, markdownlint with `MD013` disabled. |
| Application code | None | No `apps/`, `services/`, `packages/`, `infra/`. |
| Specifications | None | No `docs/` or `specs/` before this report. |
| Dev Container | None | Correct per `AGENTS.md` until the stack exists. |

### Repository facts verified with GitHub

| Fact | Value |
| --- | --- |
| Slug | `lamassuiot/lamassu-cla` |
| Visibility | Public |
| Default branch | `main` |
| Template repository | No |
| Issues | Enabled |
| Discussions | Disabled |
| Homepage URL | Not set |
| License | GNU AGPLv3 |
| Copyright holder | LKS S. Coop. |

### Local toolchain observed on the maintainer workstation

| Tool | Version | Implication |
| --- | --- | --- |
| Go | 1.24.1 | Pin the Go version in `go.mod` and CI. |
| Node.js | 23.10.0 | Odd-numbered, non-LTS. CI uses Node 24. Pin one version in `.nvmrc`/`engines`. |
| pnpm | Not installed | Package-manager choice is open (see D-07). |
| Terraform | 1.16.1 | Pin `required_version`. |
| TFLint | 0.61.0 | Usable for Terraform linting. |
| Docker | Installed | Enables local DynamoDB/S3 emulators and a Dev Container. |
| AWS CLI | Not installed | Needed only for manual operations. |

## 2. Existing constraints the new platform must respect

These constraints come from files already in the repository. New work must comply with them or
change them explicitly through a reviewed decision.

### 2.1 CI (`ci.yml`)

- **Current status:** the `CI` workflow failed on `main` for commits `80d4bfc` and `89f9d40`
  because the instantiated contact addresses in `CODE_OF_CONDUCT.md` and `SECURITY.md` were bare
  URLs (markdownlint `MD034`). The same PR as this report wraps them as Markdown autolinks without
  changing the contacts.
- All `*.yml` and `*.yaml` files must parse. OpenAPI documents in YAML are included.
- Workflows, issue forms, and `dependabot.yml` must pass JSON-schema validation.
- All Markdown must pass `markdownlint`. This applies to every specification file.
- **Private-URL scan.** CI fails on any file containing `localhost` followed by a colon and a
  port number, or the words `internal`, `intranet`, or `corp` adjacent to a dot. This conflicts
  with the planned Go package layout, where documentation and code reference `internal` packages
  with dotted identifiers, and with local development docs and Vite configuration that use
  `localhost` with a port. See D-01.
- Placeholder scan fails on any `<REPLACE_WITH_...>` outside `AGENTS.md`/`TEMPLATE.md`.
- `gitleaks` scans full history. Test fixtures containing fake keys or webhook secrets need an
  explicit, reviewed allowlist.
- `dependency-review` runs on PRs because the repository is public.
- The job is named `Validate repository template`. New jobs for backend, frontend, Terraform, and
  OpenAPI must be added as separate workflows or jobs with stable names for required checks.

### 2.2 Governance (`AGENTS.md`, `CONTRIBUTING.md`, PR template)

- PR titles must follow Conventional Commits. PRs are squash-merged.
- Agents must not invent URLs, contacts, maintainers, versions, legal text, or licenses.
- Agents must not copy ICLA/CCLA text, invent legal wording, or link to the private
  `lamassu-platform` legal directory.
- Signed agreements and legal evidence must never enter the public repository, issues, or PRs.
- CLA verification is documented as **manual** and performed by maintainers.
- Agents may sign commits only with already-configured keys.
- Dependabot entries may be added only for manifests that actually exist.
- A Dev Container may be added only once the stack exists, and `.devcontainer/` must then be added
  to CODEOWNERS.

### 2.3 Public-repository implications

Because the repository is public:

- Threat model, security specification, and operations guides will be public. They must not include
  account IDs, ARNs, bucket names, internal hostnames, secrets, or incident details.
- Environment-specific Terraform values (account IDs, domains, role ARNs) must be supplied outside
  the repository, for example through CI variables or a private configuration repository.
- Test fixtures must use synthetic identities only.

## 3. Gap analysis against the requested target

| Target capability | Current state | Gap |
| --- | --- | --- |
| SDD artifacts (PRD, architecture, domain model, ADRs, threat model) | None | All to be created. |
| Orchestrator workflow and task-state file | None | To be created. `AGENTS.md` must be rewritten for the platform. |
| Monorepo structure | None | Full scaffold needed in Phase 2. |
| Go backend on Lambda | None | Full build-out. |
| Vite/React frontend and design system | None | Full build-out. Visual reference images were **not** attached (see D-12). |
| OpenAPI contract | None | To be created as source of truth. |
| Terraform modules and environments | None | Full build-out. Remote-state backend unknown (see D-05). |
| GitHub App | None | App registration is an organization-admin action outside the repository. |
| DocuSign adapter | None | Requires DocuSign account, integration key, and legal decision on signature level. |
| CI/CD for code and infrastructure | Governance CI only | Backend, frontend, Terraform, OpenAPI, SAST, deploy workflows needed. |
| Dependabot ecosystems | `github-actions` only | Add `gomod`, `npm`, `terraform` once manifests exist. |
| Dev Container | None | Add after Phase 2, plus CODEOWNERS entry. |

## 4. Conflicts and risks identified

| ID | Conflict or risk | Impact | Proposed handling |
| --- | --- | --- | --- |
| R-01 | `AGENTS.md` describes a reusable template, not this platform. | Agents receive wrong instructions. | Rewrite in Phase 1, preserving the integrity, legal, CLA, and Conventional Commit rules. |
| R-02 | CI private-URL regex blocks dotted `internal` identifiers and `localhost` with a port. | Go layout and dev docs fail CI. | Decide in D-01 before Phase 2. |
| R-03 | Requirement "public repository may contain canonical CLA text" vs. `AGENTS.md` "must not copy the full ICLA or CCLA". | Legal and governance conflict. | Decide in D-03. Until then, no CLA text is added. |
| R-04 | `CONTRIBUTING.md` says verification is manual and links to `https://cla.developers.lamassu.cloud/`; this platform would automate verification and may serve that URL. | Docs become inaccurate at go-live. | Keep the manual process until the platform is in production. Update docs as part of the release that enables enforcement. |
| R-05 | Bootstrapping: this repository requires a CLA, but the CLA platform is not built yet. | Contributions to the platform itself need the manual process. | Keep the manual process for this repository until production. |
| R-06 | Domain-to-company assumption is explicitly forbidden for CCLA. | Authorization model must not use email-domain matching. | Encode as a domain invariant and test scenario. |
| R-07 | "Qualified" signature claims depend on provider configuration and jurisdiction (likely eIDAS). | Legal exposure. | Signature level is configuration, never assumed. Legal must confirm the required level (D-09). |
| R-08 | Storing personal data (GitHub identity, emails, representatives) and legal evidence. | GDPR obligations: retention, erasure vs. legal hold, access requests. | Data-protection notes and retention policy require Legal/DPO input (D-10). |
| R-09 | Object Lock compliance mode is irreversible for the retention period. | Data can't be erased even when legally required, and storage costs can't be reduced. | Decide mode (governance vs. compliance) and period with Legal (D-10). |
| R-10 | Public CI could leak environment details through Terraform plan output. | Information disclosure. | Run plans with OIDC in protected environments. Don't post full plans to public PRs. |
| R-11 | GitHub Check Runs on repositories in other organizations, or on forks. | Scope creep and permissions. | Scope the GitHub App to `lamassuiot` initially (D-08). |
| R-12 | Node 23 locally vs. Node 24 in CI. | Inconsistent builds. | Pin a single LTS version. |

## 5. Unresolved decisions

These decisions must be confirmed by the named owner before the dependent work starts. None of
them has been decided in this report.

| ID | Decision | Blocks | Owner to confirm |
| --- | --- | --- | --- |
| D-01 | How to reconcile the CI private-URL scan with `internal/` Go packages and local dev URLs (narrow the regex, add path exclusions, or avoid the patterns). | Phase 2 | Maintainers |
| D-02 | Is `https://cla.developers.lamassu.cloud/` the production URL of this platform, or a separate static site? | Phase 4–6 | Repository owner |
| D-03 | May the public repository contain the canonical ICLA/CCLA text, and who supplies it? | Phase 5 | Legal |
| D-04 | AWS account structure: one account per environment or shared; AWS region(s). | Phase 2 (infra), Phase 9 | Platform/infrastructure owner |
| D-05 | Terraform remote-state backend (S3 bucket, DynamoDB/S3-native locking) and who provisions it. | Phase 2 (infra) | Platform/infrastructure owner |
| D-06 | Human authentication: GitHub OAuth App vs. GitHub App user-to-server tokens; session model (server-side session vs. signed cookie/JWT). | Phase 4 | Security owner |
| D-07 | Frontend package manager (`npm` vs. `pnpm`) and workspace tool. | Phase 2 | Maintainers |
| D-08 | GitHub App scope: `lamassuiot` only, or additional organizations; who registers the App. | Phase 8 | Organization admins |
| D-09 | Required electronic-signature level (simple, advanced, qualified) and evidence package per agreement type and jurisdiction. | Phase 6 | Legal |
| D-10 | Retention periods, Object Lock mode, erasure policy, and data minimization (email vs. email hash). | Phase 5–6 | Legal / DPO |
| D-11 | DocuSign account, environment (demo vs. production), and integration-key ownership. | Phase 6 | Repository owner |
| D-12 | Visual reference images were mentioned but not provided. Brand assets and fonts that may be used. | Phase 3 | Design / repository owner |
| D-13 | Role assignment: who are maintainers, legal administrators, and platform administrators, and how roles are granted (GitHub team membership vs. application-managed). | Phase 4 | Security owner |
| D-14 | Co-author trailers: are co-authors required to be covered, and how is an unverified trailer email handled? | Phase 8 | Maintainers / Legal |
| D-15 | Whether existing manually verified signatories must be migrated into the platform, and from what source. | Phase 6–7 | Repository owner / Legal |
| D-16 | Custom domains, certificates, and WAF enablement per environment. | Phase 9 | Platform/infrastructure owner |

## 6. Proposed Phase 1 deliverables

Phase 1 produces specifications only. Each item is a separate, reviewable PR, in this order:

1. **Orchestration foundation**
   - Rewrite `AGENTS.md` for the platform, keeping existing governance rules.
   - Add the SDD orchestration rules and a task-state file (`specs/STATE.md` or equivalent).
   - Add the decision log (`docs/decisions/`) seeded with D-01 to D-16.
2. **Product requirements** — `specs/requirements.md`: personas, roles, capabilities, non-goals,
   success criteria.
3. **Architecture specification** — `specs/architecture.md`: hexagonal layering, Lambda topology,
   API Gateway routes, data flows, environment model, ADRs for major choices.
4. **Domain model** — entities, invariants, state machines (agreement, contributor, organization,
   signing session), DynamoDB keys, GSIs, transactional boundaries.
5. **Initial OpenAPI structure** — `specs/api/openapi.yaml` skeleton: conventions, error format,
   pagination, idempotency, security schemes, versioning. No endpoints implemented.
6. **Security specification and threat model** — `specs/security/`.
7. **Design-token specification** — `specs/ux/design-tokens.md` (blocked on D-12 for final values).

Phase 2 (repository and tooling) starts only after items 1–5 are reviewed and D-01, D-04, D-05,
and D-07 are resolved.

## 7. Approval

Approved on 2026-10-05 by the repository owner, together with the Phase 1 deliverable order and
decisions A-01 to A-11. Resolutions of R-01 to R-12 and the current status of every decision are
maintained in the [decision log](../docs/decisions/decision-log.md). This report is kept as a
snapshot of the repository at commit `89f9d40` and is not updated further.
