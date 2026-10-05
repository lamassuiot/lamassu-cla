# lamassu-cla

<!--
Optional: badges commonly used across Lamassu Open Source repos, e.g. license, build status,
latest release. Remove this block if not applicable to this repository.

[![License](https://img.shields.io/badge/license-AGPL--3.0-blue.svg)](./LICENSE)
-->

<REPLACE_WITH_REPOSITORY_DESCRIPTION>

---

## 🚀 Overview

Secure CLA management platform for Lamassu, including ICLA/CCLA workflows, GitHub App
pull-request validation, electronic-signature integrations, and serverless AWS infrastructure.

---

## 📦 Getting Started

### Prerequisites

There are no technical prerequisites for using or contributing to the repository.
Environment-specific requirements are documented in the development and deployment guides.

### Installation and usage

Clone the repository and follow the setup and contribution instructions in
[CONTRIBUTING.md](./CONTRIBUTING.md). Configure the required environment variables, then use the
provided commands to run, test, and validate the frontend and backend. Infrastructure and
application deployments are performed through the documented Terraform and GitHub Actions
workflows.

---

## 📚 Documentation

<!-- Optional: link to project-specific docs (public URLs only). Remove if not applicable. -->

---

## 🤝 Contributing

We welcome contributions!

Please read [CONTRIBUTING.md](./CONTRIBUTING.md) before submitting a Pull Request.

Please create or link an issue before starting implementation work.

This project follows our [Code of Conduct](./CODE_OF_CONDUCT.md).

---

## 🧭 Development Workflow

This repository follows the Lamassu Open Source engineering workflow.

### 🧠 Key Principles

* Strategic planning is managed centrally.
* Implementation work is created in this repository as **Features**, **Tasks**, or **Bugs**.
* All changes are submitted through **Pull Requests**.
* Every PR should reference the issue it addresses, for example:

  ```text
  Closes #<issue>
  ```

* Keep changes **small, focused, and reviewable**.

### 👥 Ownership

* Issues are assigned to **individuals**.
* GitHub **teams (see [CODEOWNERS](./.github/CODEOWNERS))** define code ownership and review context.

> Team owns the code; person owns the work.

<!--
Optional: if this project has publicly accessible reference documentation for its workflow
or conventions, link it here using only public URLs. Do not link to private repositories,
internal wikis, or non-public documentation.
-->

---

## ❓ Getting Help

* For questions and general discussion, use [GitHub Discussions](../../discussions) if enabled for this repository, or see [CONTRIBUTING.md](./CONTRIBUTING.md).
* For bugs, feature requests, and tasks, please [open an issue](../../issues/new/choose).
* To report a security vulnerability, do **not** open a public issue — follow [SECURITY.md](./SECURITY.md) instead.

---

## 🧾 Changelog

See [CHANGELOG.md](./CHANGELOG.md) for release history. Entries are generated from
[Conventional Commits](https://www.conventionalcommits.org/) via `git-cliff`.

---

## 📄 License

This project is licensed under the terms of the [LICENSE](./LICENSE) file (GNU AGPLv3).

Copyright holder: LKS S. Coop.

Contributions require the applicable ICLA or CCLA. Read the [current Lamassu CLA
agreements and signing instructions](https://cla.developers.lamassu.cloud/).

Third-party attributions, where applicable, are recorded in [NOTICE](./NOTICE).
