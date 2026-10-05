# Using this template

This guide is for the people creating a new Lamassu open-source repository from this template.
Agents follow [AGENTS.md](./AGENTS.md), which contains the full policy. This file and
`scripts/configure-repository.sh` are template-only: delete both after instantiation. The
`Validate repository template` job in `.github/workflows/ci.yml` fails in generated repositories
while either file remains.

Never guess repository names, descriptions, contacts, versions, URLs, or legal information. Ask
the responsible owner and leave the placeholder in place until the value is confirmed.

## 1. Create the repository

Create the repository from this template in the intended organization, for example:

```bash
gh repo create lamassuiot/<repository-name> --public \
  --template lamassuiot/lamassu-open-source-repo-template --clone
```

Rulesets and private-repository security features require a public repository or a paid plan.

## 2. Replace placeholders

Search for `<REPLACE_WITH_` and replace every match with a confirmed value.

| Placeholder | Files | Confirmed by |
| --- | --- | --- |
| `<REPLACE_WITH_REPOSITORY_NAME>` | `README.md`, `NOTICE` | Repository owner |
| `<REPLACE_WITH_REPOSITORY_DESCRIPTION>` | `README.md` | Repository owner |
| `<REPLACE_WITH_PROJECT_OVERVIEW>` | `README.md` | Repository owner |
| `<REPLACE_WITH_PREREQUISITES>` | `README.md` | Maintainers |
| `<REPLACE_WITH_INSTALLATION_AND_USAGE>` | `README.md` | Maintainers |
| `<REPLACE_WITH_SECURITY_CONTACT>` | `SECURITY.md` | Security owner |
| `<REPLACE_WITH_SUPPORTED_VERSION_POLICY>` | `SECURITY.md` | Maintainers |
| `<REPLACE_WITH_CONDUCT_CONTACT>` | `CODE_OF_CONDUCT.md` | Community or governance owner |
| `<REPLACE_WITH_COPYRIGHT_HOLDER>` | `README.md`, `NOTICE` | Legal |
| `<REPLACE_WITH_COPYRIGHT_YEARS>` | `NOTICE` | Legal |
| `<REPLACE_WITH_PROJECT_HOMEPAGE_URL>` | `NOTICE` | Repository owner |

Delete `NOTICE` instead if the repository bundles no third-party material that requires
attribution.

## 3. Adapt repository files

- **Changelog and releases:** `cliff.toml`, `CHANGELOG.md`, and `.github/workflows/release.yml`
  assume the repository publishes tagged `vX.Y.Z` releases. If this repository will not publish
  versioned releases, delete those three files instead of configuring them.
- **CODEOWNERS:** keep `@lamassuiot/lamassu-maintainers` only if the team owns this repository,
  and give the team explicit write access, otherwise Code Owner review can never be satisfied:

  ```bash
  gh api -X PUT orgs/lamassuiot/teams/lamassu-maintainers/repos/lamassuiot/<repository-name> \
    -f permission=push
  ```

- **Default branch:** if it is not `main`, update the `push.branches` filters in
  `.github/workflows/ci.yml` and `.github/workflows/scorecard.yml`.
- **Dependabot:** keep the `github-actions` entry in `.github/dependabot.yml` and add entries only
  for dependency manifests that exist in the repository.
- **Dev Container:** add `.devcontainer/` only when the technology stack is known, and add it to
  CODEOWNERS.

## 4. Configure GitHub settings

Run the setup script from a clone with an authenticated `gh` session that has administrator
access to the new repository:

```bash
scripts/configure-repository.sh lamassuiot/<repository-name>
```

The script:

- Allows squash merging only, uses the PR title as the squash commit title, and deletes head
  branches after merge.
- Enables private vulnerability reporting, Dependabot alerts and security updates, secret
  scanning, and push protection.
- Creates or updates the `Protect default branch` ruleset: no deletion or force push, signed
  commits, pull requests with one approval and Code Owner review, squash merges only, and the
  template workflows as required status checks.
- Verifies the `bug`, `enhancement`, and `needs-triage` labels, which come from the organization
  default labels, and the organization Issue Types `Bug`, `Feature`, and `Task`.

The script is safe to run again. Dependency review can fail on the first pull request right after
the script enables Dependabot alerts, while GitHub builds the dependency graph; re-run the job.

Configure these manually:

- Contributions use the CLA process only. If the organization's DCO app is installed for the new
  repository, an organization owner must remove the repository from the app's repository access.
- Rulesets for protected release branches, when applicable.
- GitHub Discussions, if the project uses them.
- Domain automation for GitHub Projects.

## 5. Finish

1. Delete `TEMPLATE.md` and `scripts/configure-repository.sh`.
2. Open a pull request with a Conventional Commit title, for example
   `chore: instantiate repository from template`.
3. Confirm that every required status check passes.

## Checklist

- [ ] Repository name, description, and owning organization confirmed.
- [ ] No `<REPLACE_WITH_...>` placeholders remain.
- [ ] Security contact, conduct contact, and supported-version policy confirmed.
- [ ] Copyright holder confirmed by legal.
- [ ] Maintainer team has write access and matches CODEOWNERS.
- [ ] `scripts/configure-repository.sh` completed without failures.
- [ ] Private vulnerability reporting tested.
- [ ] Issue chooser shows the Bug, Feature, and Task forms.
- [ ] The [Lamassu CLA portal](https://cla.developers.lamassu.cloud/) is reachable.
- [ ] Template-only files deleted.
