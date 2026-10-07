# ADR-0011: Biome for JavaScript and TypeScript formatting and linting

- Status: Accepted
- Date: 2026-10-07
- Decision-log entry: A-17 (amends A-13, A-15, A-16, and D-23)

## Context

F-01 used ESLint 9 (`@eslint/js`, `typescript-eslint` strict, `eslint-plugin-jsx-a11y` strict,
`eslint-plugin-react-hooks`, `eslint-plugin-react-refresh`) and Prettier. These packages added 247
entries to `package-lock.json` and constrained the ESLint and TypeScript versions through plugin
peer ranges (A-13, A-15, A-16). The repository owner decided to replace both tools with Biome,
provided the existing controls can be preserved.

A bounded check inside the Dev Container compared the 123 ESLint rules active on portal files with
Biome 2.5.15. Each rule was triggered with a synthetic fixture and verified against both tools.

## Decision

- Biome replaces ESLint and Prettier for JavaScript, TypeScript, JSX, JSON, CSS, and HTML. The
  version is pinned exactly (`@biomejs/biome` 2.5.15), because Biome patch releases can change
  results.
- `biome.json` is the single configuration. It enables the `recommended` preset and sets every
  rule that has an ESLint counterpart to `error`. `make lint` runs `biome ci --error-on-warnings`.
- The formatter keeps the previous Prettier style: two spaces, 100 columns, single quotes,
  trailing commas, and semicolons. `package.json` files are always expanded to match npm output.
- `tsc --noEmit` remains the type check and Vitest remains the test runner. Go, Terraform, OpenAPI,
  and CodeQL tooling are unchanged.
- Inline `style` attributes in `apps/web` fail lint through two controls:
  - `nursery/noInlineStyles` (error), which covers DOM elements and `createElement`.
  - The GritQL plugin `.biome/plugins/no-inline-style-attribute.grit`, which also covers
    components such as `<Card style={...} />`. `noInlineStyles` misses this case.
- Rules without a Biome equivalent are covered as follows:

  | Former rule | Replacement |
  | --- | --- |
  | `react-hooks` React Compiler rules | `nursery/useReactCompiler` (error) |
  | `@typescript-eslint/no-dynamic-delete` | GritQL plugin `no-dynamic-delete.grit` |
  | `@typescript-eslint/no-non-null-asserted-nullish-coalescing` | GritQL plugin `no-non-null-asserted-nullish-coalescing.grit` |
  | `@typescript-eslint/triple-slash-reference`, `ban-ts-comment` (`@ts-nocheck`, undescribed `@ts-expect-error`) | `scripts/check-ts-directives.sh`, with tests |
  | `no-delete-var`, `no-octal`, `no-require-imports` (`import x = require`) | `tsc` (TS1102, TS1121, TS1202) |

- Accepted gaps: `no-invalid-regexp` (constructor strings are checked at runtime by tests),
  `no-unexpected-multiline` (made moot by enforced semicolons and the format check), and the
  `react-hooks` `config`, `gating`, and `unsupported-syntax` rules. The first two `react-hooks`
  rules apply only to React Compiler build options, which the project does not use. The third is
  partly covered by `noGlobalEval`.

## Consequences

- `package-lock.json` drops from 389 to 151 entries. The ESLint peer-range constraints on
  TypeScript and ESLint no longer apply. TypeScript stays pinned to 6.0.3 under A-13 until a
  separate decision changes it.
- The lint and format check is faster: about 1.3 s for `biome ci` against about 9 s for ESLint plus
  Prettier on the same loaded host.
- Biome is stricter in two places, with no options to relax them:
  - `noEmptyBlockStatements` flags empty arrow functions.
  - `noNoninteractiveElementInteractions` flags `onLoad` and `onError` on `img` and `iframe`.
- `noInlineStyles` and `useReactCompiler` are nursery rules. Nursery rules carry no semantic
  versioning guarantee and can change or disappear in any release. The exact version pin and the
  GritQL plugin limit this risk. Every Biome upgrade must re-run `make check`.
- Biome GritQL plugin `includes` patterns must start with `**/`. Other patterns silently match
  nothing.
- Biome has no equivalent of the `jsx-a11y` component-mapping settings. Only a few a11y rules
  accept component names, for example `noLabelWithoutControl`. Custom design-system components
  may need those options in F-05.
- Token-only CSS validation stays in F-04 (D-23). Biome CSS GritQL plugins worked for named
  properties, but no general raw-value pattern was demonstrated.

## Alternatives considered

- **Keep ESLint and Prettier.** Rejected by the repository owner. ESLint plugin peer ranges
  constrained the TypeScript and ESLint upgrade paths, and the dependency surface was large.
- **Biome with a small ESLint layer for `react-hooks`.** Rejected. The layer would add back about
  140 packages and the `typescript-eslint` peer range on TypeScript. `useReactCompiler` covered 11
  of the 11 testable React Compiler fixtures.
