# Agent Instructions

## Purpose of this repository

This repository is a reusable template for Lamassu open-source repositories. It contains
community-health files, contribution guidance, issue forms, security guidance, and validation
workflows that can be instantiated for a specific Lamassu project.

Do not treat this repository as a finished project. Before publishing a repository created from
this template, replace or confirm every repository-specific value.

`TEMPLATE.md` is the human-facing instantiation guide, and `scripts/configure-repository.sh`
applies the baseline GitHub settings. Both files are template-only and must be deleted from the
generated repository after instantiation; the `Validate repository template` job in
`.github/workflows/ci.yml` fails while they remain.

The generic template must not contain a technology-specific `.devcontainer/devcontainer.json`.
Dev Container configuration is generated only after repository instantiation. The generated
repository must add `.devcontainer/` to the normal CODEOWNERS review scope.

## Information integrity

Never invent or guess any of the following:

- Repository names or slugs
- Repository descriptions
- URLs
- Maintainers or owning teams
- Email addresses or security contacts
- Supported versions or support policies
- Licenses, copyright holders, or other legal information

If required information is missing, ask the repository owner for it. If the work must proceed
without an answer, leave an explicit `TODO` or a clearly named placeholder such as
`<REPLACE_WITH_REPOSITORY_NAME>`. Do not silently substitute plausible-looking information.

## Repository tooling

Use the Serena MCP tools for repository navigation, symbol search, code understanding, and related
analysis when Serena is available in the agent environment. If Serena is unavailable or does not
support the required operation, use the repository's standard tools, such as `rg`, `git`, and the
available build or validation commands. Do not assume that Serena or any other optional MCP server
is installed.

## Instantiation workflow

When creating a repository from this template, complete these steps in order:

1. **Repository identity**
   - Confirm the repository name and GitHub slug.
   - Replace `<REPLACE_WITH_REPOSITORY_NAME>` only with the verified value.
   - Confirm the repository visibility and that it belongs to the intended Lamassu organization.

2. **Description**
   - Obtain an approved one-sentence description of the project.
   - Replace `<REPLACE_WITH_REPOSITORY_DESCRIPTION>` only with verified project wording.
   - Do not infer the project purpose from its name alone.

3. **Owning Lamassu domain**
   - Confirm whether the repository belongs to the IoT, SaaS, infrastructure, platform, AWS
     Serverless, or another Lamassu domain.
   - Use the verified domain for ownership, review, and project metadata.
   - Domain is calculated from the source repository by automation; agents must not guess or
     hard-code it in GitHub Project fields.

4. **Maintainers**
   - Confirm the repository’s maintainer team and its GitHub handle.
   - Use the verified `lamassu-maintainers` or domain-specific team only when the repository
     owner confirms that choice.
   - Replace all maintainer placeholders before relying on CODEOWNERS or review enforcement.

5. **Default branch**
   - Verify the repository’s actual default branch in GitHub.
   - Update branch-specific links and workflow assumptions only after confirmation.
   - Do not assume `main`, `master`, or any other branch name.
   - The template workflows trigger on pushes to `main`. If the confirmed default branch differs,
     update the `push.branches` filters in `.github/workflows/ci.yml` and
     `.github/workflows/scorecard.yml`.

6. **Security configuration**
   - Complete the root-level `SECURITY.md` as described below.
   - Verify whether GitHub Security Advisories are enabled for the repository.
   - Confirm the private fallback reporting contact and replace
     `<REPLACE_WITH_SECURITY_CONTACT>` only with an approved, real contact.

7. **Labels and Issue Types**
   - Confirm that the repository has the required labels: `bug`, `enhancement`, and
     `needs-triage`.
   - Confirm that the organization Issue Types are named exactly `Bug`, `Feature`, and `Task`.
   - Remember that issue forms do not create missing labels automatically.
   - Do not add a `task` label unless the repository owner explicitly confirms that it exists.

8. **Validation**
   - Validate all YAML files and GitHub Issue Forms.
   - Review links, placeholders, workflows, CODEOWNERS, and community-health files.
   - Test the issue-template chooser in the instantiated repository.
   - Review the final diff and confirm that unrelated template or user changes were preserved.
   - Delete the template-only files `TEMPLATE.md` and `scripts/configure-repository.sh`.

## CODEOWNERS and review protection

This template is intended primarily for public repositories in the `lamassuiot` organization.
The default code owner is:

```text
@lamassuiot/lamassu-maintainers
```

The documented default rule is:

```text
* @lamassuiot/lamassu-maintainers
```

When instantiating the template, verify that:

- The `lamassu-maintainers` team exists.
- The team has explicit write access to the generated repository.
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

CODEOWNERS only requests or identifies reviewers. It does not by itself prevent merging. After
creating the repository, configure a branch ruleset for the repository's protected default branch
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

After repository creation, configure repository settings or rulesets to require:

- Pull Requests.
- Required status checks.
- Required Code Owner approval.
- Required signed commits on protected branches.
- The PR-title lint workflow as a required status check.
- Squash merging only, with the Pull Request title as the default squash commit message, so the
  default branch history follows Conventional Commits for `git-cliff`.

`scripts/configure-repository.sh` applies these settings to the default branch and enables private
vulnerability reporting, secret scanning, and push protection. Rulesets for release branches remain
manual.

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
service. Contributor-facing CLA documentation belongs in `CONTRIBUTING.md`; public repositories
should contain only references and contributor instructions. The legal text and signing process are
centrally managed.

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
- CLA verification is currently manual and is performed by Lamassu maintainers.
- Signed agreements, private legal evidence, employment documents, and contributor records must
  never be added to Pull Requests, Issues, public repositories, public project boards, or public
  comments.
- Checking the CLA boxes in a Pull Request is only a contributor declaration; it does not replace
  maintainer verification.

Agents must not:

- Copy the full ICLA or CCLA into target repositories.
- Create or modify signed legal agreements.
- Invent legal wording.
- Link to the private `lamassu-platform` legal directory.
- Claim that a contributor is legally covered without maintainer verification.
- Replace the manual CLA process with an automated CLA service unless explicitly instructed.

## Placeholder conventions

Use explicit, searchable placeholders when verified information is unavailable:

- `<REPLACE_WITH_REPOSITORY_NAME>`
- `<REPLACE_WITH_REPOSITORY_DESCRIPTION>`
- `<REPLACE_WITH_PROJECT_OVERVIEW>`
- `<REPLACE_WITH_PREREQUISITES>`
- `<REPLACE_WITH_INSTALLATION_AND_USAGE>`
- `<REPLACE_WITH_SECURITY_CONTACT>`
- `<REPLACE_WITH_CONDUCT_CONTACT>`
- `<REPLACE_WITH_SUPPORTED_VERSION_POLICY>`
- `<REPLACE_WITH_COPYRIGHT_HOLDER>`
- `<REPLACE_WITH_COPYRIGHT_YEARS>`
- `<REPLACE_WITH_PROJECT_HOMEPAGE_URL>`
- `<REPLACE_WITH_DEFAULT_BRANCH>`
- `<REPLACE_WITH_MAINTAINER_TEAM>`

Always use the `<REPLACE_WITH_...>` form; do not use HTML comments or plausible-looking sample
content as placeholders, because they are invisible or misleading once rendered.

A placeholder is not complete configuration. Report every remaining placeholder and the person or
team that must confirm it. The `Validate repository template` job in `.github/workflows/ci.yml`
fails in generated (non-template) repositories while any `<REPLACE_WITH_...>` placeholder remains
outside `AGENTS.md` and `TEMPLATE.md`.

## Completing SECURITY.md

Keep `SECURITY.md` at the repository root. Before publishing the instantiated repository:

- Verify that GitHub Security Advisories are enabled and that the repository’s private reporting
  path works.
- Replace the supported-version placeholder with the project’s approved support policy. Do not
  invent versions or claim support that has not been confirmed.
- Configure a real private fallback contact when Security Advisories are unavailable, replacing
  `<REPLACE_WITH_SECURITY_CONTACT>` only after confirmation.
- Define the repository scope accurately, including the source code, configuration, workflows, or
  other maintained assets that are actually covered.
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
- Keep security guidance portable by referring to the instantiated repository’s local
  `SECURITY.md`.
- Preserve the confidentiality warnings and do not request secrets or private legal information
  in public issues.

Issue forms do not create missing labels automatically. Labels must be configured in the
repository or supplied as organization default labels before the forms can apply them.

## Dependabot configuration

Dependabot configuration is repository-specific. Do not copy a single configuration blindly
between Go, frontend, infrastructure, container, or other repositories.

When instantiating a repository from this template:

1. Inspect the generated repository for dependency manifests before creating a configuration.
2. Detect ecosystems from files that actually exist; do not infer an ecosystem from the project
   name, owning domain, or expected technology.
3. Keep the template's `github-actions` entry in `.github/dependabot.yml`, because every generated
   repository inherits the template workflows. Add entries only for other ecosystems whose
   supported dependency manifests actually exist.
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

The generic template ships only a `github-actions` Dependabot entry, because its workflows are
inherited by every generated repository. Entries for other ecosystems are added only after
repository instantiation and must match the actual technology stack. Security alerts and security
updates can still be enabled independently in GitHub settings.

## Development container

Dev Container configuration is repository-specific. Do not hardcode it into this generic template,
because generated repositories may use Go, frontend technologies, Python, Terraform, or other
technology stacks.

When instantiating a repository from this template:

1. Inspect the generated repository before creating a Dev Container.
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

## Portable configuration

`config.yml` must remain portable in this template. Do not add repository-specific Discussions,
Security Advisory, contact, or other absolute links to it. Add those links only after a concrete
repository exists and its verified URL is known.

## Legal and licensing boundaries

Do not modify legal documents, licenses, CLA files, copyright notices, or other legal information
unless the user explicitly requests that change and provides or confirms the authoritative content.
Do not invent legal contacts, license terms, copyright holders, agreement text, or supported legal
policies.

## Final validation checklist

Before declaring an instantiated repository complete, verify:

- No unresolved placeholders remain, or every remaining `TODO` has an owner and explicit follow-up.
- No contact information, maintainer, URL, version, license, or legal information was invented.
- `SECURITY.md` is present at the repository root and its reporting process is configured.
- GitHub Security Advisories and the private fallback contact have been verified.
- Required labels exist: `bug`, `enhancement`, and `needs-triage`.
- Issue Types `Bug`, `Feature`, and `Task` exist and match the forms exactly.
- All YAML files parse successfully.
- All issue forms are visible and selectable in the GitHub issue chooser.
- Reusable forms and `config.yml` contain no repository-specific links.
- Domain automation is configured to calculate Domain from the source repository.
- No unintended legal, license, CLA, copyright, or community-health changes were introduced.
- Template-only files `TEMPLATE.md` and `scripts/configure-repository.sh` were deleted.
- The final diff contains only the intended instantiation changes.

## Completion report

The agent completing an instantiation must report:

- Files changed.
- Placeholders replaced.
- Information still requiring confirmation.
- Labels and Issue Types verified.
- Validation performed.
- Remaining manual configuration, including GitHub settings, Security Advisories, repository
  access, CODEOWNERS, default branch, labels, Projects, and automation.
