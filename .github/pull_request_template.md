## Pull Request title

Pull Request titles must follow the Conventional Commits format. Valid examples include:

- `feat: add device discovery`
- `fix(api): handle expired certificate`
- `docs: update installation guide`

The title is validated automatically, and the title-lint workflow must pass before merging.

## Summary

<!-- Describe what this PR changes and why. -->

## Related issue

<!--
Reference the Feature, Task, or Bug this PR addresses, e.g. "Closes #123".
If this work relates to a Strategic Epic tracked centrally, link it for context only —
do not use a closing keyword for it, since Strategic Epics are not closed from domain repository PRs.
-->

## Breaking change

- [ ] This PR introduces a breaking change. The title uses `!` (for example,
  `feat(api)!: change enrollment contract`), and the summary describes the impact and migration
  steps.

## Testing performed

<!-- Describe how the change was tested (unit tests, manual verification, etc.). -->

## Documentation impact

- [ ] No documentation changes needed
- [ ] Documentation updated in this PR
- [ ] Follow-up documentation issue created

## Security or privacy impact

<!-- Describe any security or privacy implications, or state "None". -->

## Checklist

- [ ] This PR does not include credentials, tokens, private keys, or other secrets
- [ ] This PR does not include confidential, customer-specific, or commercial licensing information
- [ ] The related issue is referenced above
- [ ] Tests and/or documentation have been updated as needed

## CLA verification

- [ ] I have completed the applicable Lamassu CLA process described in [CONTRIBUTING.md](../blob/HEAD/CONTRIBUTING.md#contributor-license-agreement).
- [ ] I am contributing under the correct capacity: ICLA or CCLA.
- [ ] If covered by a CCLA, I am not also submitting this contribution under the ICLA.
- [ ] All contributors and co-authors are identified.
- [ ] No private agreements or legal evidence are included in this pull request.

> These declarations do not replace the manual verification performed by Lamassu maintainers.
