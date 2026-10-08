# F-04: Design tokens and themes

| Field | Value |
| --- | --- |
| Status | Approved; unit 1 merged (PR #26); unit 2 (token-only CSS check) in progress |
| Phase | 3 |
| Owner | Repository owner; Design for final visual values (D-12) |
| Depends on | F-01; A-11, A-17, D-12, D-23; [design tokens](../ux/design-tokens.md), [ADR-0011](../../docs/decisions/0011-biome-for-javascript-and-typescript.md) |
| Approved by | Repository owner, 2026-10-08 (Phase 3 start; specification, PR #25) |

## Problem statement

The portal has no visual foundation. `packages/design-system` exports only the `--cla-` prefix, and
`apps/web` renders unstyled HTML. Components (F-05) and screens (F-06) need typed design tokens,
light and dark themes, and an automated check that CSS uses only tokens, so that pages cannot
introduce one-off visual values (FR-UI-04, NFR-04, D-23).

## User roles

- All portal visitors, including anonymous visitors: they see the selected or system theme.
- Contributors and maintainers (developer experience): they consume tokens and are bound by the
  token-only CSS check.

No authorization roles are involved.

## Preconditions

- F-01 is done: npm workspaces, TypeScript 6.0.3, Biome 2.5.15, and Vitest are in place.
- The token values in [design tokens](../ux/design-tokens.md) are the proposed values (A-11).
  Final brand assets, logos, and fonts are pending D-12. F-04 implements the proposed values; a
  later D-12 decision changes them through a token change.

## Main flow

1. Token definitions are written as typed TypeScript objects in
   `packages/design-system/src/tokens/`, exactly as listed in sections 2 to 7 of the
   [design tokens](../ux/design-tokens.md). The design tokens define no component tokens yet;
   F-05 adds them with the components that use them.
2. A generator script, run with Node.js 24 type stripping and no new dependency, renders the tokens
   to `packages/design-system/src/styles/tokens.css` as CSS custom properties prefixed `--cla-`.
   The generated file is committed.
3. The CSS defines the light theme on `:root` and `[data-theme="light"]`, the dark theme on
   `[data-theme="dark"]`, and the dark theme on `:root:not([data-theme])` inside
   `@media (prefers-color-scheme: dark)`. Semantic colors and gradients reference color primitives
   through `var()`. Color primitives, including the status colors, and the tokens without a
   primitive tier (shadows, typography, spacing, radius, border width, and motion) contain raw
   values. Breakpoints are exported only as TypeScript constants, because CSS custom properties
   cannot be used in media queries.
4. `packages/design-system` exports the token objects, the generated CSS, and a theme module with
   `getThemePreference()`, `setThemePreference(theme)`, and `applyTheme()`.
5. At start-up, `apps/web/src/main.tsx` imports the token CSS and the base styles, then calls
   `applyTheme()` before the first render. `applyTheme()` reads the stored preference and sets
   `data-theme` on `<html>`, or removes it so that the operating-system preference applies.
6. Base styles in the design system set the document background, text color, font family, body
   text size, a global `:focus-visible` ring (2 px `focus.ring`, 2 px offset), and reduced motion
   under `prefers-reduced-motion: reduce`. They use only `var(--cla-…)` values.
7. A token-only CSS check runs in `make lint` and rejects raw visual values in every CSS file except
   the generated `tokens.css`.

## Alternative flows

- **Stored preference.** The visitor previously chose `light` or `dark`; `applyTheme()` sets
  `data-theme` to that value. Selecting `system` removes the stored value and the attribute.
- **Preference changed at runtime.** `setThemePreference()` stores the value and re-applies the
  theme without a reload. The theme toggle control itself is a component in F-05.
- **Fonts before D-12.** No font files are bundled and no font is loaded from a third-party
  service. `font.family.sans` and `font.family.mono` resolve to the system fallback stacks until
  D-12 confirms the font families. Adding font files later is a token change.

## Error cases

| Case | Detection | System behavior | User-visible result |
| --- | --- | --- | --- |
| Storage unavailable (blocked, private mode, quota) | `localStorage` access throws | Preference is not read or written; the system theme applies; no error is raised | System theme; the choice is not remembered |
| Stored value is not `light` or `dark` | Allowlist check in `getThemePreference()` | Value is ignored and treated as `system` | System theme |
| Raw color, spacing, radius, shadow, or duration in CSS | Token-only CSS check | `make lint` and CI fail | File, line, and rule named |
| Generated `tokens.css` differs from the token definitions | Drift check | `make lint` and CI fail | Instruction to run the generator |
| A semantic token pair falls below its contrast threshold | Contrast unit test | `make test` and CI fail | Pair, theme, ratio, and threshold named |

## Security considerations

- The strict Content Security Policy (security specification section 5, T-15) must not be
  weakened. F-04 adds no inline `<script>` or `<style>` elements and no `style` attributes; the
  theme is applied from the bundled module.
- The stored theme value is checked against an allowlist before it is applied.
- `localStorage` holds only the key `cla-theme` with the value `light` or `dark`. It contains no
  identifier and no personal data, and nothing is sent to the server.
- No fonts, stylesheets, or other assets are loaded from third-party origins.
- No new runtime dependency is added. If Biome cannot express the token-only check, Stylelint is
  added only as D-23 allows: with a patched `braces`, or with a reviewed, time-limited
  `dependency-review` exception recorded in the decision log.

## API changes

None.

## Data-model changes

None.

## UI changes

No new screens. The placeholder `App` renders on the canvas background with the primary text
color and the default font tokens in both themes. Components, previews, and the theme toggle are
F-05; the app shell is F-06.

## Acceptance criteria

- **AC-04-1** Given the token definitions, every primitive, semantic, status, gradient, typography,
  spacing, radius, border, shadow, motion, and breakpoint token in the
  [design tokens](../ux/design-tokens.md) exists with the specified value, and no other token
  exists.
- **AC-04-2** Every CSS custom property in `tokens.css` starts with `--cla-`. Semantic colors and
  gradients reference only color primitives through `var()` and contain no raw color.
- **AC-04-3** Given each theme, every pair in section 2.4 of the design tokens meets its WCAG 2.2
  threshold (4.5:1 for text, 3:1 for control borders and focus indicators), and its computed ratio
  is within 0.01 of the table value.
- **AC-04-4** Given no stored preference, the operating-system color-scheme preference selects the
  theme. Given a stored `light` or `dark`, that theme applies regardless of the system preference.
  Given `system`, the stored value and the `data-theme` attribute are removed.
- **AC-04-5** Given storage that throws, or a stored value other than `light` or `dark`, the system
  theme applies and no error is thrown.
- **AC-04-6** Given a CSS file other than the generated `tokens.css`, when it contains a raw color
  (hex, `rgb()`, `hsl()`, or named color), a raw length in a spacing, radius, border-width, or
  font-size property, a raw shadow, or a raw duration, then `make lint` fails. Values built only
  from `var(--cla-…)`, `0`, and CSS-wide keywords pass. Suppressing the check is not allowed.
- **AC-04-7** Given a change to the token definitions without regenerating `tokens.css`,
  `make lint` fails.
- **AC-04-8** Under `prefers-reduced-motion: reduce`, transitions are limited to opacity and to
  `motion.duration.fast` or less.
- **AC-04-9** No font file is bundled and no request is made to a third-party origin until D-12 is
  confirmed.
- **AC-04-10** `make check` passes, and `docs/development/local-setup.md` documents the generator
  and the token-only check.

## Automated test scenarios

| ID | Scenario | Level | Covers |
| --- | --- | --- | --- |
| F04-T1 | The token objects match the values listed in the design tokens, and no extra token exists. | Unit | AC-04-1 |
| F04-T2 | Parsing the generated CSS: every property has the `--cla-` prefix, and semantic colors and gradients use only `var()` references to color primitives. | Unit | AC-04-2 |
| F04-T3 | WCAG contrast is computed for every pair in section 2.4 in both themes. | Unit | AC-04-3 |
| F04-T4 | Theme module with mocked storage and `matchMedia`: stored, system, invalid, and throwing-storage cases. | Unit | AC-04-4, AC-04-5 |
| F04-T5 | Token-only check fixtures: at least one violation per value category fails with the expected rule, and token-only fixtures pass. A suppression comment fails. | Lint, Script | AC-04-6 |
| F04-T6 | The drift check fails after a token edit without regeneration (verified once during implementation) and passes on the committed files. | Script | AC-04-7 |
| F04-T7 | The reduced-motion rule exists in the base styles and uses only motion tokens. | Unit | AC-04-8 |
| F04-T8 | The build output contains no font files and no third-party URLs. | Script | AC-04-9 |
| F04-T9 | CI runs `make check` (Frontend and Dev Container jobs). | CI | AC-04-10 |

## Observability requirements

None. F-04 has no runtime behavior that the server can observe.

## Rollback considerations

Front-end source and tooling only; revert the PR. No data, API, or infrastructure is affected.
The `cla-theme` value left in a visitor's browser is ignored by older builds.

## Implementation plan

1. **Tokens and generation.** Typed token definitions, the generator, the committed `tokens.css`,
   the drift check, and tests F04-T1 to F04-T3 and F04-T6.
2. **Token-only CSS check.** Evaluate Biome CSS rules and GritQL plugins first (D-23, A-17), with
   the fixtures of F04-T5. If Biome cannot express every category in AC-04-6, record the result and
   propose Stylelint under the D-23 conditions before adding it.
3. **Themes and base styles.** Theme module, base styles, the `apps/web` start-up integration, and
   tests F04-T4, F04-T7, and F04-T8.

Each unit is one pull request and depends on the previous one.

Implementation notes (unit 1):

- Tokens live in `primitives.ts`, `semantic.ts`, and `scales.ts`; `css.ts` renders them, and
  `scripts/generate-tokens.ts` writes or checks `src/styles/tokens.css`
  (`npm run tokens` and `npm run tokens:check` in `@lamassu-cla/design-system`).
- CSS names follow the token path: `bg.canvas` becomes `--cla-color-bg-canvas`, a status primitive
  becomes `--cla-color-status-success-dark-text`, and typography styles expand to `-size`,
  `-line-height`, `-weight`, `-letter-spacing`, and `-transform` properties.
- Typography sizes are converted from the specified pixels to `rem` with a 16 px root.
- Biome formatting is disabled for the generated `tokens.css` only (`biome.json` override); Biome
  still lints it, and the drift check guards its content.
- `make lint` and the Frontend workflow run the drift check. The package's `tsconfig.json` enables
  `allowImportingTsExtensions` (required for Node.js type stripping) and the `node` types.
- The contrast test recomputed every pair in section 2.4; all match the table within 0.01.

Implementation notes (unit 2):

- Biome 2.5.15 expresses every category in AC-04-6, so Stylelint is not added (D-23). The plugin
  `.biome/plugins/no-raw-css-values.grit` applies to `**/*.css` except the generated `tokens.css`
  (`biome.json`). It reports:
  - every hex color (`CssColor`) and color function (`rgb()`, `hsl()`, `hwb()`, `lab()`, `lch()`,
    `oklab()`, `oklch()`, `color()`, `color-mix()`, `light-dark()`, `device-cmyk()`), anywhere;
  - the 148 CSS named colors in color, background, border, outline, decoration, shadow, `fill`,
    `stroke`, and custom properties;
  - raw lengths in margin, padding, inset, gap, position offsets, border width and radius,
    outline, column rule, `font`, `font-size`, `line-height`, `letter-spacing`, shadows, and custom
    properties;
  - raw durations (`ms`, `s`) anywhere.
- Named colors and lengths are scoped by property, because checking every identifier or dimension
  flagged grid areas, animation names, font families, `max-width: 72ch`, and media-query
  breakpoints. Layout sizes outside the token categories (for example `width`) are not checked.
- Biome suppression comments disable plugin diagnostics in every form (`biome-ignore`,
  `-all`, `-start`/`-end`), so `scripts/check-css-suppressions.sh` rejects any `biome-ignore`
  comment in CSS (`make lint`).
- Biome GritQL regexes match the whole text, and capturing groups `( )` silently match nothing;
  the plugin uses only non-capturing groups `(?: )`. A generic `$property: $value` snippet does not
  compile for CSS; the plugin matches `CssGenericProperty(name, value)` nodes instead.
- `scripts/check-css-tokens.test.sh` (29 cases) lints each case in a temporary repository with a
  copy of the real `biome.json` and `.biome/`, and runs in `make test` and the Frontend workflow.

## Open questions

- D-12: final brand assets, logos, and font families. Until it is confirmed, fonts use the system
  fallback stacks and the color values remain the proposed values.
