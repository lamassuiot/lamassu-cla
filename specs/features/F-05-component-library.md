# F-05: Component library and preview

| Field | Value |
| --- | --- |
| Status | In review (specification only) |
| Phase | 3 |
| Owner | Repository owner; Design for final visual values (D-12) |
| Depends on | F-04; A-11, A-17, D-12, D-21; [design tokens](../ux/design-tokens.md), [ADR-0004](../../docs/decisions/0004-provider-neutral-signing-and-signature-level.md), [ADR-0011](../../docs/decisions/0011-biome-for-javascript-and-typescript.md) |
| Approved by | Pending |

## Problem statement

F-04 provides tokens, themes, and base styles, but no components. Screens (F-06 onward) must
compose shared components only; pages may not define one-off button, card, or form styles
([design tokens](../ux/design-tokens.md#10-guardrails)). The portal must meet WCAG 2.2 level AA
(NFR-04), handle loading, empty, error, and permission-denied states in every data-driven view
(FR-UI-02), and support light and dark themes (FR-UI-04) and small to wide screens (FR-UI-05).
F-05 builds the component set once in `packages/design-system`, with a preview of every component
and automated accessibility and visual-drift checks.

## User roles

- All portal visitors: they use the components through screens built in later features.
- Contributors and maintainers (developer experience): they build screens only from these
  components and preview them in light and dark themes.

No authorization roles are involved; components never decide access (the API enforces it).

## Preconditions

- F-04 is done: tokens, `tokens.css`, `base.css` with the `components` cascade layer, the theme
  module, and the token-only CSS check are merged.
- The visual values are the proposed values (A-11). Final brand assets, logos, and fonts are pending
  D-12; components use tokens only, so a D-12 change is a token change.
- Portal languages are pending D-21; components contain no built-in user-visible text.

## Main flow

1. Each component in section 9 of the [design tokens](../ux/design-tokens.md#9-components) is
   implemented once in `packages/design-system/src/components/<name>/`, as a typed React component
   with a CSS Module (`<name>.module.css`, supported by Vite without a new dependency).
2. Component CSS is declared inside `@layer components`, uses only `var(--cla-…)` values (enforced
   by the F-04 token-only check), and defines no new token. A new token follows the change-control
   rules in the design tokens.
3. Components prefer native elements and semantics (`button`, `input`, `select`, `dialog`,
   `table`) over ARIA re-implementations. Every interactive element is keyboard operable and shows
   the F-04 focus ring.
4. All user-visible text, including button labels, empty-state messages, and accessible names, is
   passed in through props. Components contain no built-in copy (D-21).
5. Every component has preview stories in light and dark themes, covering loading, empty, error,
   and disabled states where applicable. A coverage check fails when an exported component has no
   story.
6. Automated accessibility checks run against every story in both themes; visual regression
   snapshots cover the primary set (button, form field, card, status badge, dialog, and table).
7. `packages/design-system` exports the components from its entry point; `apps/web` consumes them
   in F-06.

## Alternative flows

- **Reduced motion.** Under `prefers-reduced-motion: reduce`, component transitions follow the F-04
  rules (opacity only, `motion.duration.fast` or less). Components use motion tokens only.
- **Narrow screens and zoom.** Components reflow at a 320 CSS px viewport and at 200% zoom without
  loss of content or horizontal scrolling, except tables, which scroll horizontally inside their
  own container.
- **Theme toggle.** The theme toggle offers `light`, `dark`, and `system` and calls the F-04
  `setThemePreference()`; it does not read or write storage directly.

## Error cases

| Case | Detection | System behavior | User-visible result |
| --- | --- | --- | --- |
| Status badge receives an unknown domain state | Lookup in the state-to-status map | Renders the `neutral` status with the supplied label | Neutral badge with text |
| A dialog is opened without an accessible name | Development-time check and story test | Fails the component test | Not shipped |
| Copy to clipboard fails or is not permitted | Clipboard API rejects | Announces the failure through the live region; the value stays selectable | Message that the value can be copied manually |
| A component is exported without a story | Story coverage check | `make check` and CI fail | Component named |
| A story has an accessibility violation | Automated accessibility check | CI fails | Rule, element, and story named |
| A primary component's rendering changes | Visual regression comparison | CI fails until the snapshot is reviewed and updated | Image difference |

## Security considerations

- Components render text through React only; `dangerouslySetInnerHTML` stays prohibited
  ([security](../security/security.md)). Links accept only relative paths or `https:` URLs, and
  links that open a new tab set `rel="noopener noreferrer"`.
- No inline `<script>` or `style` attributes; the Content Security Policy is unaffected.
- The agreement status display shows the signature level exactly as supplied by the data and never
  labels a signature as qualified on its own
  ([ADR-0004](../../docs/decisions/0004-provider-neutral-signing-and-signature-level.md)).
- The audit-event display and hash display render only the fields they are given; they never fetch
  data, and stories and fixtures use synthetic identities and values only.
- Preview, accessibility, and visual-regression tooling are development dependencies. They are not
  part of the portal build, which the F-04 build-output check keeps free of third-party load
  points. The preview is not deployed; any hosted preview needs a separate decision.
- New dependencies follow the existing controls: exact pins, `npm audit`, the required
  `Dependency review` check, and no `dependency-review` exception without a reviewed decision.
  Tool telemetry is disabled.

## API changes

None.

## Data-model changes

None. The status badge uses the domain states listed in section 2.3 of the design tokens; it does
not change them.

## UI changes

The component set in section 9 of the design tokens, grouped by implementation unit:

| Group | Components |
| --- | --- |
| Foundations | Button (primary, secondary, ghost, danger), section label, card, status badge, theme toggle |
| Forms | Form field (label, hint, error), input, select, checkbox, radio, textarea |
| States and feedback | Loading state, empty state, error state, permission-denied state, toast, live-region announcer |
| Overlays | Dialog, confirmation dialog, drawer |
| Data display | Table, pagination, timeline item, step indicator, hash display, agreement status display, audit-event display |
| Layout | App shell, navigation, page header |

The theme toggle is part of F-05 ([F-04](F-04-design-tokens-and-themes.md#alternative-flows)).
F-06 composes the app shell, navigation, and page header into the running portal.

## Acceptance criteria

- **AC-05-1** Every component listed in section 9 of the design tokens, and the theme toggle, is
  exported from `@lamassu-cla/design-system`. Given an exported component without a story, the
  story coverage check fails.
- **AC-05-2** Every component's CSS is inside `@layer components`, passes the token-only check
  without exceptions, and defines no token; `apps/web` and stories use no inline `style`.
- **AC-05-3** Every interactive component is operable with the keyboard alone (Tab, Shift+Tab,
  Enter, Space, Escape, and arrow keys where the native element supports them) and shows the
  visible focus ring. Component tests cover each interaction.
- **AC-05-4** Given every story, in both themes, the automated accessibility check reports no
  violations at WCAG 2.2 level A or AA.
- **AC-05-5** Given an open dialog, confirmation dialog, or drawer, focus moves into it, stays
  inside it, Escape closes it, focus returns to the element that opened it, and the rest of the
  page is inert. A confirmation dialog for a destructive action requires an explicit confirming
  action; Escape and the cancel action do not confirm.
- **AC-05-6** The status badge maps every domain state in section 2.3 of the design tokens to the
  specified status color and always shows its text label; an unknown state renders as `neutral`.
- **AC-05-7** The hash display shows the full value in the monospace font with an accessible name;
  copying announces success or failure through the live-region announcer.
- **AC-05-8** The agreement status display renders the signature level only from its input and
  never adds "qualified" or equivalent wording.
- **AC-05-9** Components contain no built-in user-visible text; every label, message, and
  accessible name comes from props.
- **AC-05-10** Visual regression snapshots of the primary set (button, form field, card, status
  badge, dialog, table) in both themes run in CI in a pinned environment; an unreviewed difference
  fails the check.
- **AC-05-11** At a 320 CSS px viewport and at 200% zoom, components reflow without loss of
  content; only tables scroll horizontally, inside their container.
- **AC-05-12** The portal build contains no preview or test tooling; `make check` passes; and
  `docs/development/local-setup.md` documents how to run the preview and the checks.

## Automated test scenarios

| ID | Scenario | Level | Covers |
| --- | --- | --- | --- |
| F05-T1 | Story coverage check: removing a story for an exported component fails; the committed set passes. | Script | AC-05-1 |
| F05-T2 | Token-only check, layer check, and inline-style check on all component CSS and stories. | Lint | AC-05-2 |
| F05-T3 | Keyboard interaction tests per interactive component. | Component | AC-05-3 |
| F05-T4 | Automated accessibility check over every story in light and dark themes. | Component, CI | AC-05-4 |
| F05-T5 | Focus management tests for dialog, confirmation dialog, and drawer. | Component | AC-05-5 |
| F05-T6 | Status badge mapping test for every domain state and an unknown state. | Unit | AC-05-6 |
| F05-T7 | Hash display copy success and failure announcements. | Component | AC-05-7 |
| F05-T8 | Agreement status display renders only the supplied signature level. | Component | AC-05-8 |
| F05-T9 | A check that component source contains no string literals rendered as text. | Script | AC-05-9 |
| F05-T10 | Visual regression comparison in a pinned container; a deliberate change fails (verified once). | CI | AC-05-10 |
| F05-T11 | Stories rendered at 320 px and at 200% zoom in the visual and accessibility runs. | CI | AC-05-11 |
| F05-T12 | The F-04 build-output check and a dependency check confirm preview tooling is not in the portal build. | Script | AC-05-12 |

## Observability requirements

None. Components have no server-observable behavior.

## Rollback considerations

Front-end source and development tooling only; revert the PR. No data, API, or infrastructure is
affected. Screens do not use the components until F-06.

## Implementation plan

1. **Tooling evaluation and decision.** Evaluate, with the pinned Node.js 24, TypeScript 6.0.3,
   React 19, and Vite versions: a component preview (Storybook, per the
   [architecture](../architecture.md#7-frontend-architecture)), an automated accessibility engine,
   keyboard-interaction testing, and visual regression in a pinned container. For each candidate,
   record installation without `overrides`, `npm audit`, install scripts, licenses, telemetry, and
   whether CI can run it offline after installation. Record the result as a proposed decision; no
   component work starts until it is approved.
2. **Foundations and forms.** Button, section label, card, status badge, theme toggle, form field,
   input, select, checkbox, radio, and textarea, with the story coverage check (F05-T1, F05-T2,
   F05-T3, F05-T6, F05-T9).
3. **States and feedback.** Loading, empty, error, and permission-denied states, toast, and the
   live-region announcer.
4. **Overlays.** Dialog, confirmation dialog, and drawer (F05-T5).
5. **Data display.** Table, pagination, timeline item, step indicator, hash display, agreement
   status display, and audit-event display (F05-T7, F05-T8).
6. **Layout and visual baseline.** App shell, navigation, and page header; accessibility and
   visual regression runs in CI for the primary set (F05-T4, F05-T10, F05-T11, F05-T12).

Each unit is one pull request and depends on the previous one. Units 2 to 6 add stories and tests
for their own components.

## Open questions

- D-12: final brand assets, logos, and fonts (the app shell shows no logo until it is confirmed).
- D-21: portal languages; components take all text through props, so no translation mechanism is
  chosen in F-05.
- The preview, accessibility, and visual-regression tools are chosen in unit 1 and recorded as a
  decision before adoption.
