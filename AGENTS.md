# Agent Instructions

## Purpose of this repository

This repository (`lamassuiot/lamassu-cla`, public, AGPLv3) contains Lamassu's CLA Management
Platform. The platform publishes CLA versions, collects and stores signed ICLAs and CCLAs, manages
contributor and organization coverage, and validates GitHub pull requests through a GitHub App.

The platform handles **legal evidence and personal data**. Treat every change to agreement
handling, coverage rules, authentication, authorization, storage, or retention as high risk.

The repository was instantiated from the Lamassu open-source repository template. The
instantiation is complete and the template-only files were removed.

## Read first

Before planning or making changes, read in this order:

1. This file.
2. [`specs/sdd-workflow.md`](specs/sdd-workflow.md) — the mandatory Specification-Driven
   Development workflow, approval gates, and change control.
3. [`specs/STATE.md`](specs/STATE.md) — current phase, approvals, blocked items, contradictions.
4. [`docs/decisions/decision-log.md`](docs/decisions/decision-log.md) — settled and open decisions.
5. The specifications relevant to the task under `specs/`.

## Repository layout

| Path | Contents |
| --- | --- |
| `specs/` | Normative specifications: requirements, architecture, domain model, `api/`, `security/`, `ux/`, `features/`, `test-scenarios/`. |
| `docs/` | Implementation and operational guides. |
| `docs/decisions/` | Decision log and architecture decision records (ADRs). |
| `apps/web/` | Portal (Vite, React, TypeScript). Added in Phase 2. |
| `services/api/` | Go backend for AWS Lambda. Added in Phase 2. |
| `packages/` | `api-client`, `design-system`, `shared-types`. Added in Phase 2. |
| `infra/` | Terraform modules, environments, bootstrap. Added in Phase 2. |

Keep this layout. Changing it requires an ADR.

## Information integrity

Never invent or guess any of the following:

- Repository names or slugs
- Repository descriptions
- URLs
- Maintainers or owning teams
- Email addresses or security contacts
- Supported versions or support policies
- Licenses, copyright holders, or other legal information

In addition, never invent or guess:

- AWS account IDs, regions, resource names, ARNs, or domains.
- Agreement text, signature levels, retention periods, or other legal policy.
- Provider accounts, integration keys, or GitHub App identifiers.

If required information is missing, ask the repository owner for it and record the question in the
[decision log](docs/decisions/decision-log.md) with the person or team that must confirm it. If the
work must proceed without an answer, use a placeholder as described in
[Placeholder conventions](#placeholder-conventions). Do not silently substitute plausible-looking
information.

## Repository tooling

Use the Serena MCP tools for repository navigation, symbol search, code understanding, and related
analysis when Serena is available in the agent environment. If Serena is unavailable or does not
support the required operation, use the repository's standard tools, such as `rg`, `git`, and the
available build or validation commands. Do not assume that Serena or any other optional MCP server
is installed.

## Specification-Driven Development

All work follows [`specs/sdd-workflow.md`](specs/sdd-workflow.md). In particular:

- Do not implement a feature until its specification and acceptance criteria exist and approval is
  recorded in [`specs/STATE.md`](specs/STATE.md).
- Never silently change an API contract, the data model, a security assumption, legal behavior, or
  a design token. Follow the change-control rules.
- Update specifications and ADRs in the same pull request as the implementation they describe.
- When code and specifications disagree, stop and record the contradiction in `specs/STATE.md`.
- Run the relevant tests and validations after every implementation unit.
- Keep pull requests small, focused, and reviewable.

## Architecture guardrails

- The backend follows the hexagonal layering in [`specs/architecture.md`](specs/architecture.md).
  Domain code has no I/O and no SDK imports; only adapters import AWS, GitHub, or provider SDKs.
- `specs/api/openapi.yaml` is the API source of truth. Change the contract before the handlers.
- Signing is accessed only through the provider-neutral signing port
  ([ADR-0004](docs/decisions/0004-provider-neutral-signing-and-signature-level.md)). Never
  implement signature cryptography.
- All infrastructure is managed with Terraform. Document any manual configuration explicitly.
- Deployments use GitHub OIDC. Never add long-lived AWS access keys to GitHub.

## Security and privacy rules

This repository is public. Agents must:

- Never commit secrets, credentials, tokens, private keys, personal data, signed agreements, legal
  evidence, or real contributor records, including in tests, fixtures, logs, or examples.
- Use synthetic identities and synthetic documents in fixtures.
- Never commit environment identifiers such as account IDs, ARNs, bucket names, or private
  hostnames. Supply them through CI environment variables or private configuration.
- Keep personal data out of logs, metrics, object keys, Check Run output, and audit metadata.
- Never claim that a signature is qualified unless the configured provider service is qualified
  ([ADR-0004](docs/decisions/0004-provider-neutral-signing-and-signature-level.md)).
- Never infer CCLA authorization or coverage from an email domain
  ([ADR-0006](docs/decisions/0006-explicit-ccla-authorization.md)).
- Keep S3 Object Lock retention configurable and never set production retention values without
  Legal confirmation ([ADR-0005](docs/decisions/0005-configurable-object-lock-retention.md)).
- Follow [`specs/security/security.md`](specs/security/security.md) and complete its pull-request
  checklist for code and infrastructure changes.

## Design-system guardrails

- Use only the tokens and components defined in
  [`specs/ux/design-tokens.md`](specs/ux/design-tokens.md) and `packages/design-system`.
- Pages must not define one-off button, card, or form styles.
- New colors require a token change, a contrast check, and a specification update. New spacing
  values require written justification.

## CI constraints

The `CI` workflow enforces, among other checks:

- YAML syntax and GitHub configuration schemas.
- Markdown lint for every Markdown file. Wrap bare URLs and email addresses in angle brackets.
- No unresolved `<REPLACE_WITH_...>` placeholders outside this file.
- A private-URL scan (`scripts/check-private-urls.sh`, [ADR-0003](docs/decisions/0003-refine-private-url-scan.md)).
  It fails when a URL's host contains an `internal`, `intranet`, or `corp` label. Go `internal/`
  packages and documented `localhost` or `127.0.0.1` development URLs are allowed. Add exceptions
  only to `.github/private-url-allowlist.txt`, each with a justification comment, through review.
- Secret scanning of the full Git history with gitleaks.

`make check` additionally runs Biome (`biome ci`) for JavaScript, TypeScript, JSX, JSON, CSS, and
HTML formatting and linting, `tsc --noEmit`, Vitest, Go tooling, and the repository scripts in
`scripts/`. Do not reintroduce ESLint or Prettier, and do not weaken a Biome rule, plugin in
`.biome/plugins/`, or repository check, without an ADR
([ADR-0011](docs/decisions/0011-biome-for-javascript-and-typescript.md)).

The default branch is `main`. Workflows trigger on pushes to `main`.

## CODEOWNERS and review protection

This repository belongs to the `lamassuiot` organization. The code owner is:

```text
@lamassuiot/lamassu-maintainers
```

The documented default rule is:

```text
* @lamassuiot/lamassu-maintainers
```

Verify that:

- The `lamassu-maintainers` team exists.
- The team has explicit write access to this repository.
- The repository belongs to the expected organization.
- The `CODEOWNERS` file is located at a GitHub-supported path, such as
  `.github/CODEOWNERS`, `CODEOWNERS`, or `docs/CODEOWNERS`.

If the repository belongs to another organization, replace the owner with a verified team from
that organization. Never invent a team name, handle, or access assignment. CODEOWNERS paths are
case-sensitive.

The maintainer team owns general repository changes and governance files, including:

- `.github/`, including `.github/CODEOWNERS`
- `AGENTS.md`
- `CODE_OF_CONDUCT.md`
- `CONTRIBUTING.md`
- `SECURITY.md`
- `LICENSE`
- `NOTICE`
- `cliff.toml`

When a `.devcontainer/` directory is added, add it to the CODEOWNERS review scope in the same pull
request.

CODEOWNERS only requests or identifies reviewers. It does not by itself prevent merging. Configure
a branch ruleset for the repository's protected default branch
and protected release branches, when applicable, that requires:

- Pull requests.
- Required status checks where applicable.
- Approval from Code Owners before merging.

Verify that the branch ruleset is configured after repository creation. Report any missing team
access, missing branch protection, or unresolved organization-specific configuration.

### Commit signing and Pull Request titles

Pull Request titles must follow the Conventional Commits format. This policy applies to Pull
Request titles; it does not necessarily require every commit message to use that format.

Use this format:

```text
<type>[optional scope][!]: <description>
```

Accepted types are:

- `feat`
- `fix`
- `docs`
- `style`
- `refactor`
- `perf`
- `test`
- `build`
- `ci`
- `chore`
- `revert`

Optional scopes are allowed. Breaking changes may use `!`, for example:

```text
feat(api)!: change enrollment contract
```

Agents must create Pull Request titles that follow this format. Agents must not generate, copy,
store, or expose private signing keys. Agents may use commit signing only when it is already
configured in the contributor's environment. Agents must not automatically create a new signing
key.

Signed commits must be enforced through GitHub branch protection or repository rulesets, not only
through agent instructions. The policy should apply primarily to the repository's protected default
branch and protected release branches, when applicable.

Repository settings or rulesets must require:

- Pull Requests.
- Required status checks.
- Required Code Owner approval.
- Required signed commits on protected branches.
- The PR-title lint workflow as a required status check.
- Squash merging only, with the Pull Request title as the default squash commit message, so the
  default branch history follows Conventional Commits for `git-cliff`.

These settings, private vulnerability reporting, secret scanning, and push protection were applied
to the default branch during instantiation. Verify them rather than assuming they are still in
place. Rulesets for release branches remain manual.

Conventional Commit Pull Request titles are enforced by this required status check. Configure the
lint workflow's stable check name in the applicable GitHub ruleset for the protected default branch
and any protected release branches, when applicable. The title-lint workflow is fork-safe because
it uses `pull_request_target` without checking out or executing contributor code.

Repository settings and rulesets cannot be enforced only by files committed to the repository.
They require administrator configuration. Verify that the ruleset targets the protected default
branch and any protected release branches, and report any missing manual configuration.

## Contributor License Agreement Documentation

The Pull Request template must remain concise. It must not contain the full ICLA or CCLA text,
private repositories, private signing records, legal evidence, or references to an automated CLA
service. Contributor-facing CLA documentation belongs in `CONTRIBUTING.md`. Other Lamassu
repositories contain only references and contributor instructions. The legal text and signing
process are centrally managed.

### Canonical agreement text in this repository

As the CLA platform, this repository **may** contain the canonical ICLA and CCLA text, their version
history, and signing instructions
([ADR-0002](docs/decisions/0002-public-cla-text-and-private-evidence.md)). Agents must:

- Add or change agreement text only when Legal supplies or confirms the exact content (D-03).
- Never draft, paraphrase, translate, or "improve" agreement text.
- Treat a published version as immutable. Corrections are new versions.

Signed agreements, personal data, and audit evidence remain exclusively in the private service.

The current legal documents and maintainer process are centrally managed through the
[Lamassu CLA portal](https://cla.developers.lamassu.cloud/). Agents must verify the authoritative
portal or public legal location before adding contributor-facing links. If the authoritative
location is unavailable or unclear, do not claim that a link is valid and do not invent a replacement
path. Report the uncertainty and leave the contributor-facing link unchanged until it is confirmed.

The CLA process is:

- Individual contributors must use the ICLA when contributing personally.
- Contributors acting on behalf of an organization must be covered by the applicable CCLA.
- A person whose contribution is covered by a CCLA must not also submit the same contribution under
  the ICLA.
- CLA verification is currently manual and is performed by Lamassu maintainers. `CONTRIBUTING.md`
  keeps describing it as manual until the platform is operational in production; switching is a
  separate, reviewed change.
- Signed agreements, private legal evidence, employment documents, and contributor records must
  never be added to Pull Requests, Issues, public repositories, public project boards, or public
  comments.
- Checking the CLA boxes in a Pull Request is only a contributor declaration; it does not replace
  maintainer verification.

Agents must not:

- Copy the full ICLA or CCLA into other repositories.
- Create or modify signed legal agreements.
- Invent legal wording.
- Link to the private `lamassu-platform` legal directory.
- Claim that a contributor is legally covered without maintainer verification.
- Replace the manual CLA process with an automated CLA service unless explicitly instructed.

## Placeholder conventions

When verified information is unavailable, use an explicit, searchable placeholder in the
`<REPLACE_WITH_...>` form, for example `<REPLACE_WITH_AWS_REGION>`, and add a matching entry to the
[decision log](docs/decisions/decision-log.md). Do not use HTML comments or plausible-looking sample
content as placeholders, because they are invisible or misleading once rendered.

In documentation, prefer referring to the decision-log entry (for example "pending D-04") over a
placeholder.

A placeholder is not complete configuration. Report every remaining placeholder and the person or
team that must confirm it. The `Validate repository template` job in `.github/workflows/ci.yml`
fails while any `<REPLACE_WITH_...>` placeholder remains outside `AGENTS.md`, so placeholders
block merging until resolved.

## Maintaining SECURITY.md

Keep `SECURITY.md` at the repository root and:

- Verify that GitHub Security Advisories are enabled and that the repository’s private reporting
  path works.
- Keep the supported-version policy aligned with the approved policy. Do not invent versions or
  claim support that has not been confirmed.
- Change the private fallback contact only after confirmation by the security owner.
- Keep the repository scope accurate as source code, infrastructure, and workflows are added.
- Keep instructions clear that vulnerabilities must not be reported through public issues,
  discussions, or pull requests.

## Issue forms

The reusable forms are stored in `.github/ISSUE_TEMPLATE/`:

- `bug.yml` — reports reproducible defects and uses Issue Type `Bug`.
- `feature.yml` — proposes capabilities and uses Issue Type `Feature`.
- `task.yml` — defines actionable engineering work and uses Issue Type `Task`.

Expected behavior:

- Public submissions through all three forms receive the `needs-triage` label.
- Bug reports also use `bug`; feature requests also use `enhancement`.
- Do not use a `task` label unless it exists and its use has been explicitly approved.
- Do not add repository-specific URLs or contact links to these reusable forms.
- Keep security guidance portable by referring to the repository’s local
  `SECURITY.md`.
- Preserve the confidentiality warnings and do not request secrets or private legal information
  in public issues.

Issue forms do not create missing labels automatically. Labels must be configured in the
repository or supplied as organization default labels before the forms can apply them. The
required labels are `bug`, `enhancement`, and `needs-triage`; the organization Issue Types must be
named exactly `Bug`, `Feature`, and `Task`.

## Dependabot configuration

Dependabot configuration is repository-specific. Do not copy a single configuration blindly
between Go, frontend, infrastructure, container, or other repositories.

When adding or changing dependency manifests:

1. Inspect the repository for dependency manifests before changing the configuration.
2. Detect ecosystems from files that actually exist; do not infer an ecosystem from the project
   name, owning domain, or expected technology.
3. Keep the `github-actions` entry in `.github/dependabot.yml`. Add entries only for other
   ecosystems whose supported dependency manifests actually exist, in the same pull request that
   adds the manifest.
4. Configure the correct directory for each detected manifest. Use `/` for a root-level manifest
   and the manifest's actual subdirectory for nested projects or workspaces.
5. Use the matching Dependabot ecosystem when applicable, including:
   - `go.mod` → `gomod`
   - `package.json` → `npm`
   - `Dockerfile` → `docker`
   - `requirements.txt` or `pyproject.toml` → `pip`
   - Terraform files or `.terraform.lock.hcl` → `terraform`
   - `.github/workflows/` → `github-actions`
   - Other ecosystems supported by Dependabot when their manifests are detected.
6. Use a weekly update schedule unless the repository owner justifies another frequency.
7. Set a reasonable `open-pull-requests-limit`, such as `5`, for each update configuration.
8. Group compatible dependency updates when doing so is appropriate for the repository.
9. Do not hard-code registries, reviewers, assignees, target branches, repository-specific URLs,
   or organization-specific secrets without explicit confirmation.
10. Do not use the `needs-triage` label for Dependabot pull requests. Use the default
    `dependencies` label, or another label only when it already exists or is explicitly created.
11. Configure Dependabot pull-request titles to comply with the repository's Conventional Commit
    policy. For example:

    ```yaml
    commit-message:
      prefix: "chore"
    ```

12. Dependabot alerts and security updates are separate settings. Enable them separately in the
    repository or organization settings when required; a `dependabot.yml` file alone does not
    enable those features.
13. Validate the generated `.github/dependabot.yml` after creating it, including YAML syntax,
    ecosystem names, manifest directories, schedules, limits, grouping, and labels.

The planned stack uses `gomod`, `npm`, and `terraform`. Add each entry only when its manifest
exists. Security alerts and security updates can still be enabled independently in GitHub settings.

## Development container

Dev Container configuration must match the actual technology stack. Add it only after the Phase 2
scaffolding exists.

When creating or changing the Dev Container:

1. Inspect the repository before creating a Dev Container.
2. Detect the language, framework, package manager, build tools, and required services from the
   actual repository files.
3. Create `.devcontainer/devcontainer.json` only when a useful project-specific configuration can
   be defined.
4. Use official Dev Container Templates and Features where appropriate.
5. Do not assume Go, Node.js, Python, Terraform, or any other technology.
6. Do not add unnecessary tools or personal editor preferences.
7. Never include credentials, tokens, private keys, customer data, or other secrets.
8. Pin container images and Feature versions where practical.
9. Keep Dockerfiles, compose files, and setup scripts next to the related `devcontainer.json`.
10. Validate the configuration by building the development container.
11. If the technology stack cannot be determined reliably, do not create a Dev Container
    configuration; document that decision instead.
12. Report the generated configuration and any manual setup still required.

## GitHub Projects and automation

GitHub Projects and automation manage the following project-management fields:

- Status
- Domain
- Priority
- Iteration
- Start Date
- Target Date
- Assignee
- Sub-issue progress

Do not add these as fields to the issue forms. Domain is calculated from the source repository by
automation. Agents must not duplicate, guess, or manually hard-code Domain values in issue forms
or repository documentation unless explicitly instructed by the project owner.

## Issue chooser configuration

Add Discussions, Security Advisory, contact, or other links to `.github/ISSUE_TEMPLATE/config.yml`
only with verified URLs. Discussions are currently disabled for this repository.

## Legal and licensing boundaries

Do not modify legal documents, licenses, CLA files, copyright notices, or other legal information
unless the user explicitly requests that change and provides or confirms the authoritative content.
Do not invent legal contacts, license terms, copyright holders, agreement text, or supported legal
policies.

## Validation checklist

Before declaring a unit of work complete, verify:

- The relevant specification, acceptance criteria, and approval exist.
- No unresolved placeholders remain, or every remaining one has a decision-log entry and owner.
- No contact information, maintainer, URL, version, license, legal, or environment information was
  invented.
- All YAML files parse, and the OpenAPI contract lints successfully.
- Markdown lint, the private-URL scan, and the placeholder scan pass locally.
- Tests for the changed code pass.
- Specifications, ADRs, `specs/STATE.md`, and the decision log reflect the change.
- No unintended legal, license, CLA, copyright, or community-health changes were introduced.
- The final diff contains only the intended changes, and unrelated user changes were preserved.

## Work report

At the end of each unit of work, report:

- Files changed.
- Specifications and decisions updated.
- Information still requiring confirmation, with the owner.
- Validation performed and its results.
- Remaining manual configuration, such as GitHub settings, AWS resources, provider accounts, or
  repository access.
