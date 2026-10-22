# F-05: Component library and preview

| Field | Value |
| --- | --- |
| Status | Done; approved; unit 1 (tooling evaluation) done; A-23 and D-25 decided; unit 2 (tooling, foundations, and forms) completed in PR #47 and PR #48; unit 3 (states and feedback) implemented in PR #49; unit 4 (overlays) implemented in PR #50; unit 5 (data display) implemented: table, pagination, timeline item, and step indicator in PR #51; hash display, agreement status display, and audit-event display in PR #52; unit 6 (layout and visual baseline) done in PR #54 and PR #55; done |
| Phase | 3 |
| Owner | Repository owner; Design for final visual values (D-12) |
| Depends on | F-04; A-11, A-17, A-23, D-12, D-21, D-25; [design tokens](../ux/design-tokens.md), [ADR-0004](../../docs/decisions/0004-provider-neutral-signing-and-signature-level.md), [ADR-0011](../../docs/decisions/0011-biome-for-javascript-and-typescript.md) |
| Approved by | Repository owner, 2026-10-21 (specification, PR #43) |

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
- **AC-05-5** Given an open dialog, confirmation dialog, or drawer, focus moves into it, Escape
  closes it, focus returns to the element that opened it, and the rest of the page is inert. While
  it is open, focus cannot reach any element of the page content behind it; the browser's native
  behavior, such as moving focus to the browser's own controls, is allowed (A-25). A confirmation
  dialog for a destructive action requires an explicit confirming
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

Implementation notes (unit 1):

- The evaluation and the choices are recorded as A-23 in the
  [decision log](../../docs/decisions/decision-log.md), approved on 2026-10-21 subject to D-25. D-25
  was decided by the repository owner on 2026-10-21: `axe-core` (MPL-2.0) and `caniuse-lite`
  (CC-BY-4.0) may be added as development-only dependencies, with their license notices kept,
  caniuse.com attributed if an artifact including its data is redistributed, and both excluded
  from the portal bundle. D-25 is reopened if Storybook, the development container, or that data
  starts being distributed.
- After the browser-test CI job added in unit 2 has run successfully, its check name is added to
  the `main` ruleset.
- Proposed: Storybook 10.6.3 for previews; `@storybook/addon-a11y` with axe-core run on every story
  in both themes through `@storybook/addon-vitest`; Vitest browser mode with Playwright 1.64.1
  (Chromium) for keyboard and focus tests; and `toMatchScreenshot` baselines created only in the
  digest-pinned Playwright 1.64.1 image.
- Focus and keyboard tests (AC-05-3, AC-05-5) use real key presses in browser mode. Storybook's
  simulated `userEvent` misreported focus in a native modal `dialog`, and jsdom implements neither
  modal `dialog` nor `inert`.
- Visual baselines differ between a developer host and the container, so they are created and
  compared only in the pinned image. Each screenshot test cleans up the previous render.

Implementation notes (unit 2, first pull request: tooling and button):

- Unit 2 is split into two pull requests to keep each reviewable: this one adds the A-23 tooling,
  the checks, and the button; the next adds the other foundations and forms.
- `packages/design-system` gains the A-23 development dependencies (exact pins) and declares React
  19 as a peer dependency. `npm audit` reports 0 vulnerabilities; no existing package version
  changed; no new install script.
- `.storybook/` configures Storybook with telemetry disabled. A `theme` global switches themes in
  the preview. `vitest.browser.config.ts` runs every story in two projects, `stories-light` and
  `stories-dark` (axe checks fail on any WCAG 2.0 to 2.2 A or AA violation), and the
  `*.browser.test.tsx` component tests with real key presses. `vitest.config.ts` keeps the Node
  unit tests separate. Verified: each project renders its theme (canvas `rgb(247, 245, 242)` and
  `rgb(11, 22, 34)`), and a deliberate contrast regression failed `color-contrast` in both.
- `scripts/check-design-system-components.sh` (8-case test) requires, for each component directory,
  the component, a story, an export from `src/index.ts`, and CSS Modules with every rule inside a
  single `@layer components` block (F05-T1, AC-05-2). Biome's inline-style rules now also cover
  `packages/design-system`.
- `scripts/check-portal-dev-deps.sh` (F05-T12, D-25) fails if the portal's production dependencies,
  or the modules in a source-mapped portal build, include `axe-core`, `caniuse-lite`, Storybook,
  Playwright, or Vitest. It runs in `make build` and the Frontend workflow. Verified: importing
  either package into the portal fails the check.
- The `Design system browser tests` job runs in the digest-pinned Playwright 1.64.1 image and
  fails if the installed `playwright` version differs from the image. Locally, `make test-browser`
  runs the same tests (Chromium from `npx playwright install chromium`).
- Button: `primary`, `secondary`, `ghost`, and `danger` variants from semantic tokens only; a token-
  styled disabled state; `type="button"` by default; no `className` or `style` props.

Implementation notes (unit 2, second pull request: foundations and forms):

- Section label, card, status badge, theme toggle, form field, input, select, checkbox, radio
  group, and textarea, each with stories in both themes and no `className` or `style` props.
- The status badge's state map is a pure module (`status.ts`) with a unit test copied from section
  2.3 of the design tokens; unknown states, including `Object` prototype names, render as
  `neutral` (F05-T6). The badge always renders its label, and accepts either a domain `state` or an
  explicit `status` for informational notices.
- `FormField` renders the label, hint, and error and passes `id`, `required`,
  `aria-describedby`, and `aria-invalid` to its control through a render function. Input, select,
  and textarea share one stylesheet; an invalid control gets a strong danger border, and the error
  text uses `status.danger.text` (contrast at least 5.80:1 on canvas, surface, and subtle
  backgrounds in both themes, computed during implementation).
- Radios are rendered by `RadioGroup` (a `fieldset` with a `legend`), so native arrow-key and Tab
  behavior applies. Hints sit outside the `<label>` and are referenced by `aria-describedby`; a
  keyboard test found that a hint inside the label became part of the accessible name, which axe
  does not report.
- The theme toggle offers `light`, `dark`, and `system` through `RadioGroup` and the F-04
  `setThemePreference()`; it does not touch storage directly.
- `.biome/plugins/no-component-copy.grit` rejects literal JSX text, string or template children, and
  literal `alt`, `title`, `placeholder`, and `aria-*` text in component source (stories and tests
  excluded), so all user-visible text comes from props (F05-T9, AC-05-9).
  `scripts/check-component-copy.test.sh` (12 cases) runs it with the real configuration.
- The Frontend workflow now runs the component-convention check, its test, and the copy-plugin
  test. Browser tests: 78 (22 story files in two themes, 4 component test files); deliberately
  removing the form-field description wiring or the theme-toggle update failed the tests.

Implementation notes (unit 3: states and feedback):

- `LiveRegionProvider` renders one polite (`role="status"`) and one assertive (`role="alert"`)
  region before any message is sent, because screen readers ignore regions that are inserted
  together with their content. `useAnnounce` sends a message to either region. A repeated message
  alternates a trailing non-breaking space so that it is announced again. The regions are visually
  hidden but stay in the accessibility tree. Using `useAnnounce` outside the provider throws.
- `LoadingState` is a polite, busy (`aria-busy`) status region. `EmptyState`, `ErrorState`, and
  `PermissionDeniedState` are labelled sections with an optional description and action. The
  heading level is a prop (default 2) so the state fits the surrounding page structure.
- `ToastProvider` takes its region and dismiss labels as props and requires a surrounding
  `LiveRegionProvider`. `useToast()` returns a function that shows a toast. Toasts are announced
  (danger toasts assertively), never take focus, and never disappear on a timer (WCAG 2.2.1);
  they stay until dismissed. Each dismiss button is described by its message. Dismissing a toast
  that holds focus moves focus to the next toast, the previous toast, or the toast region, so
  keyboard focus is never dropped to the document.
- Browser tests: 111 (14 story files, each run in both themes, and 7 component test files).
  Injecting a 5-second
  auto-dismiss or removing the focus hand-off failed the tests. The timer test fakes timers before
  the toast is shown and advances them inside `act`; without both, it did not detect the injected
  auto-dismiss.

Implementation notes (unit 4: overlays):

- `Dialog`, `ConfirmationDialog`, and `Drawer` render a native `dialog` opened with `showModal()`.
  The browser makes the rest of the page inert, focuses the first focusable element, and keeps
  focus out of the page; no custom focus trap is added (A-25). A shared `useModalDialog` hook syncs
  the element with the controlled `open` prop. Escape calls `onClose` (or `onCancel`) and the caller
  closes by setting `open` to false. Focus returns to the opener on close, and also when an open
  overlay is unmounted.
- The title is the accessible name (`aria-labelledby`); the optional description is linked with
  `aria-describedby`. An empty title throws in development (error case "dialog without an
  accessible name").
- `Dialog` requires `actions`, so there is always a pointer and touch way to close it.
  `ConfirmationDialog` places cancel first, so it takes the initial focus and Enter right away
  cancels; Escape also cancels; only the confirm button confirms, and it uses the `danger` variant
  when `destructive` is set. `Drawer` has a labelled close button in its header and opens from the
  inline end by default or from the start.
- The backdrop uses the new `overlay.scrim` token (A-24, [design tokens](../ux/design-tokens.md#26-overlay)).
  Dialog and drawer widths are `60ch` and `40ch`, capped by the viewport; these are layout widths,
  not spacing values.
- Browser tests: 135 (17 story files, each run in both themes, and 10 component test files).
  Removing `showModal()`, the focus return on unmount, the name check, or the cancel-first order, or
  wiring Escape to confirm, failed the tests. Chromium already returns focus to the opener on
  `close()`, so the explicit return is kept for browsers that do not.

Implementation notes (unit 5, first pull request: generic data display):

- Unit 5 is split into two pull requests: this one adds the table, pagination, timeline item, and
  step indicator; the next adds the hash display, agreement status display, and audit-event display
  (F05-T7, F05-T8).
- `Table` is a native `table` with a visible caption that also names its container. Column headers
  have `scope="col"`; a column marked `rowHeader` renders `th scope="row"`. Only the table scrolls
  horizontally (AC-05-11). Its container becomes a tab stop only while the table overflows it, so
  keyboard users can scroll it, and it is not an empty tab stop otherwise. Overflow is measured with
  a `ResizeObserver` on the container and the table.
- `Pagination` follows the API's cursor pagination (`next_cursor` only): previous and next buttons
  in a labelled `nav`, with no page numbers; the caller keeps earlier cursors. When the pressed
  button becomes disabled, focus moves to the other button instead of the document.
- `Timeline` (a labelled `ol`) and `TimelineItem` render a title, a `time` element with a
  machine-readable `dateTime` and a caller-formatted label (D-21), and an optional description. The
  status color is on a decorative marker only; the title states the outcome.
- `StepIndicator` is a labelled `ol`; the current step has `aria-current="step"`, and completed
  steps add a visually hidden `completedLabel`, so progress is not shown by style alone.
- Browser tests: 160 (21 story files, each run in both themes, and 14 component test files).
  Removing the pagination focus move, or making the table container never or always a tab stop,
  failed the tests. The table and container observers overlap in practice; the tests cannot tell
  them apart.

Implementation notes (unit 5, second pull request: domain data display):

- `HashDisplay` is a `figure` named by its visible `figcaption` (also linked with
  `aria-labelledby`), with the full value in a `code` element in the monospace font. The value wraps
  instead of being truncated and is selected with one click (`user-select: all`), so it can be
  copied manually. The copy button is described by the label. A successful copy is announced
  politely; a rejected copy, or a missing Clipboard API outside secure contexts, is announced
  assertively with the caller's failure message (AC-05-7, error case "copy to clipboard fails").
  It requires a surrounding `LiveRegionProvider`.
- `AgreementStatusDisplay` is a named `section` with a heading, a status badge for the domain state,
  and a description list. The signature level is the caller's text, rendered exactly as supplied;
  the component has no signature-level vocabulary of its own, so it cannot add or upgrade a
  "qualified" claim (ADR-0004, AC-05-8). Other details are optional term and description pairs.
- `AuditEventDisplay` is a named `article` with the event title, the dotted event type, a `time`
  element, and only the fields it is given, in order; it never fetches data. Callers must keep
  personal data out of the fields (domain model, audit event metadata).
- Stories and tests use synthetic values only; the example hash is the public SHA-256 of empty
  input.
- The agreement state `VOIDED` (domain model, section 3.2) had no status color in the design
  tokens (section 2.3), so the badge rendered it as `neutral` through the unknown-state fallback
  (C-02). The repository owner approved adding it to the `neutral` row beside `EXPIRED` on
  2026-10-22; the mapping and its spec-copied test now list it explicitly.
- Browser tests: 181 (24 story files, each run in both themes, and 17 component test files).
  Removing either copy announcement, announcing the failure politely, removing `user-select`, or
  appending a qualified claim to the signature level failed the tests.

Implementation notes (unit 6, first pull request: layout):

- Unit 6 is split into two pull requests: this one adds the app shell, navigation, and page header;
  the next adds the visual-regression baselines (F05-T10), the 320 px and 200% zoom runs (F05-T11),
  and the local-setup documentation (AC-05-12).
- `Navigation` is a labelled `nav` with a list of links; the current page has
  `aria-current="page"` and is marked by weight and a bar, not by color alone. Links accept only a
  relative path (not `//` or `/\`) or an `https:` URL; anything else throws (security
  considerations). Links are plain anchors; router integration is left to F-06.
- `focus.ring` in the light theme (`orange.700`) is only 2.95:1 on `bg.nav`, below the 3:1 minimum
  for focus indicators. The new `focus.ring-on-nav` token (`orange.400`, 6.99:1 dark, 6.33:1 light)
  is used for links on `bg.nav` (A-26, [design tokens](../ux/design-tokens.md#8-focus)). The token
  test checks the ratio.
- `AppShell` has a skip link to `main` as the first tab stop, a header (`banner`) with the brand and
  optional actions, and the navigation. From `bp.lg` the navigation is a sidebar; below it, a menu
  button opens it in a start-side `Drawer`, and following a link closes the drawer. The shell renders
  `Navigation` from its props, so the drawer can close without click delegation. The breakpoint is
  written as `1024px` in the media query, because custom properties cannot be used there.
- `PageHeader` renders the page's `h1`, an optional description, and actions on the hero gradient;
  text contrast holds against the worse stop (light: `text.heading` 13.35:1 and `text.secondary`
  6.17:1 on `gray.100`; dark: 15.13:1 and 8.54:1 on `teal.950`).
- The menu-open story sets Storybook's `mobile1` viewport global, which the Vitest addon applies,
  so the drawer is checked by the accessibility run.
- Browser tests: 209 (27 story files, each run in both themes, and 19 component test files).
  Removing the navigation focus-ring rule, allowing protocol-relative or non-`https:` URLs, keeping
  the drawer open after a link, removing the breakpoint, or pointing the skip link elsewhere failed
  the tests.

Implementation notes (unit 6, second pull request: visual baseline):

- **Narrow widths (F05-T11).** Two more story projects run every story in the light theme at the
  Storybook viewports `reflow` (320 px) and `zoom200` (640 px, a 1280 px window at 200% zoom),
  defined in `.storybook/preview.tsx`. The width does not change colors, so the narrow runs use one
  theme. Story widths come from the viewport global, not from the Vitest browser instance, which
  the Storybook addon overrides. A Storybook `afterEach` hook fails any story whose page scrolls
  horizontally (AC-05-11); tables scroll inside their own container. Removing the hash display's
  wrapping made the hash story fail at 320 px (551 px of content).
- **Visual regression (F05-T10).** `src/visual/primary.visual.test.tsx` screenshots the primary set
  (button, form field, card, status badge, dialog, table) in both themes at 1280, 640, and 320 px:
  36 baselines (about 420 KB) under `src/visual/__screenshots__/`. The `visual` project exists only
  with `CLA_VISUAL=1`, so host runs never compare against container baselines. `make test-visual`
  and `make test-visual-update` run it in the digest-pinned Playwright 1.64.1 image with Docker,
  `--network none`, and the caller's user ID; CI runs it in the same image. Failure images go to
  `.vitest/attachments` (ignored), which CI uploads as an artifact with `actions/upload-artifact`
  v7.0.2, pinned by commit.
- The comparison uses `pixelmatch` with `threshold: 0`. With the default threshold (0.1), changing
  `orange.500` to the `orange.600` value still passed. With 0, two consecutive runs matched, and
  changing `orange.500` by one unit failed 12 screenshots (button and dialog, both themes, three
  widths); the deliberate-change check was done once, as F05-T10 requires.
- F05-T12 is covered by the unit 2 checks (`scripts/check-web-build.sh` and
  `scripts/check-portal-dev-deps.sh` in `make build`).
- `docs/development/local-setup.md` documents the preview (`npm run storybook`), `make test-browser`,
  `make test-visual`, and `make test-visual-update` (AC-05-12).
- Browser tests: 347 (27 story files in four runs: light, dark, 320 px, and 640 px; plus 19
  component test files), and 36 visual comparisons.

## Open questions

- D-12: final brand assets, logos, and fonts (the app shell shows no logo until it is confirmed).
- D-21: portal languages; components take all text through props, so no translation mechanism is
  chosen in F-05.
- The preview, accessibility, and visual-regression tools: A-23, with D-25 (licenses) decided on
  2026-10-21.
