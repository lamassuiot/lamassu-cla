# F-06: App shell and static screens

| Field | Value |
| --- | --- |
| Status | Done; approved; unit 1 (router and shell) merged in PR #57, with A-27 decided; unit 2 (landing page and portal accessibility runs) merged in PR #59, with D-27 decided |
| Phase | 3 |
| Owner | Repository owner |
| Depends on | F-05; D-12, D-21, D-26, D-27; [architecture](../architecture.md#7-frontend-architecture), [ADR-0010](../../docs/decisions/0010-cloudfront-frontend-hosting.md) |
| Approved by | Repository owner, 2026-10-21 (specification, PR #56) |

## Problem statement

F-05 provides the components, but the portal (`apps/web`) still renders a placeholder heading. The
screens of later features (sign-in in F-07, the catalogue in F-09, and the dashboards and flows
after them) need one running application to plug into: client-side routing, the app shell with
navigation and theme switching, page titles and focus handling on navigation, and screens for
"page not found" and unexpected errors. F-06 builds that frame and one public landing page, without
any API calls, so later features add routes instead of reworking the application.

## User roles

- Visitors (unauthenticated): they reach the landing page, the not-found screen, and the error
  screen.

No authorization roles are involved. F-06 adds no route guards; F-07 and F-08 add authentication
and role-aware guards (which improve UX only; the API enforces authorization).

## Preconditions

- F-05 is done: every component, including the app shell, navigation, page header, and theme
  toggle, is exported from `@lamassu-cla/design-system`.
- The portal language is English only until D-21 is decided. All user-visible text lives in one
  copy module, so other languages can be added later without touching screens.
- No logo or brand font is used until D-12 is confirmed; the brand is shown as text.

## Main flow

1. `apps/web` adds a client-side router (React Router, per the
   [architecture](../architecture.md#7-frontend-architecture)). The exact version is evaluated and
   recorded as a proposed decision in unit 1, with the same evidence as A-23: exact pin, no
   `overrides`, `npm audit`, install scripts, licenses, and bundle-size change.
2. Every route renders inside the F-05 `AppShell`: skip link, header with the brand text and the
   `ThemeToggle`, and `Navigation` (sidebar from `bp.lg`, drawer below it).
3. Each route is loaded on demand (route-level code splitting), shows a `LoadingState` while its
   code loads, and renders a `PageHeader` with the page's only `h1`.
4. On every navigation, the document title becomes `<page title> – <brand>`, and focus moves to the
   page's `h1` so that screen-reader and keyboard users start at the new content. The first page
   load keeps the browser's default focus.
5. `Navigation` links move between routes without a full page load. Modified clicks (Ctrl, Meta,
   Shift, or the middle button) and links to other origins keep the browser's default behavior.
   This needs an additive change in `Navigation` (the click event is passed to `onNavigate`, which
   can prevent the default); the F-05 link restrictions stay.
6. Routes in F-06:

   | Path | Screen | Content |
   | --- | --- | --- |
   | `/` | Landing | Page header and a short description of the platform (D-27). No call to action until F-07 adds sign-in. |
   | any other path | Not found | `EmptyState` with a link back to the landing page. |

7. All user-visible text, including page titles and accessible names, comes from the copy module
   `apps/web/src/copy/en.ts`. Screens contain no literal text.

## Alternative flows

- **Theme.** The `ThemeToggle` in the header offers light, dark, and system; the choice persists
  through the F-04 theme module.
- **Narrow screens and zoom.** At 320 CSS px and at 200% zoom, every screen reflows without
  horizontal scrolling (F-05 AC-05-11).
- **Deep links.** CloudFront serves `index.html` for unknown paths
  ([ADR-0010](../../docs/decisions/0010-cloudfront-frontend-hosting.md)), so the router, not the
  server, shows the not-found screen. Its HTTP status is therefore 200; F-06 does not change the
  hosting configuration.

## Error cases

| Case | Detection | System behavior | User-visible result |
| --- | --- | --- | --- |
| Unknown path | No route matches | Renders the not-found screen inside the shell | "Page not found" with a link to the landing page |
| A route's code fails to load (for example after a deployment replaced the chunk) | Router error boundary | Renders the error screen inside the shell; no details of the error are shown | Error state with a reload action |
| A screen throws while rendering | Router error boundary | Same as above; the error is logged to the browser console only | Error state with a reload action |

## Security considerations

- No API calls, cookies, tokens, or personal data. Nothing is stored except the F-04 theme
  preference.
- React renders all text; `dangerouslySetInnerHTML` stays prohibited. The error screen never shows
  error messages or stack traces.
- Navigation links keep the F-05 restriction to relative paths and `https:` URLs.
- The portal build must stay compatible with a strict Content Security Policy without
  `'unsafe-inline'`: no inline scripts, inline event handlers, or `style` attributes in the HTML
  output. The policy itself is set by CloudFront
  ([ADR-0010](../../docs/decisions/0010-cloudfront-frontend-hosting.md)) and is not part of F-06.
- The new router dependency follows the existing controls: exact pin, `npm audit`, the required
  `Dependency review` check, and no runtime dependency beyond what the evaluation records.

## API changes

None.

## Data-model changes

None.

## UI changes

- `apps/web`: router, `AppShell` frame, landing page, not-found screen, error screen, copy module,
  and the title and focus handling on navigation.
- `packages/design-system`: `Navigation` passes the click event and the href to `onNavigate`, and
  `AppShell` forwards a caller's `onNavigate` to the sidebar and the drawer; the state panels accept
  `headingLevel={1}`, so a full-page state can be the page's only `h1` (AC-06-1). All three changes
  are additive; existing callers keep working. No new component and no new token.
- Whether the shell links to the portal's source code (AGPL-3.0 section 13) is decided in D-26 and
  is not part of F-06 until then.

## Acceptance criteria

- **AC-06-1** Given any route, the page renders inside the `AppShell`, with the skip link as the
  first tab stop and exactly one `h1`.
- **AC-06-2** Given a click on a `Navigation` link to another route, the route changes without a
  full page load, the document title becomes `<page title> – <brand>`, focus moves to the new
  page's `h1`, and the link to the current route has `aria-current="page"`. A Ctrl-, Meta-, or
  Shift-click, or a middle click, keeps the browser's default behavior.
- **AC-06-3** Given an unknown path, the not-found screen is shown inside the shell with a link to
  the landing page.
- **AC-06-4** Given a route whose code fails to load or that throws while rendering, the error
  screen is shown inside the shell with a reload action and without error details.
- **AC-06-5** Given every screen, in both themes and at 320 px and 640 px (200% zoom of a 1280 px
  window), the automated accessibility check reports no WCAG 2.2 A or AA violations, and the page
  does not scroll horizontally.
- **AC-06-6** Screens and `apps/web` components contain no literal user-visible text; all of it
  comes from the copy module.
- **AC-06-7** The portal build has no inline scripts, inline event handlers, or `style` attributes
  in its HTML output, still passes the F-04 build-output check and the F-05 development-dependency
  check, and the router is the only new runtime dependency.
- **AC-06-8** The theme toggle in the header switches the theme, and the choice survives a reload.

## Automated test scenarios

| ID | Scenario | Level | Covers |
| --- | --- | --- | --- |
| F06-T1 | Each route renders inside the shell with one `h1`; the skip link is the first tab stop. | Component | AC-06-1 |
| F06-T2 | Navigation between routes: no full reload, title, focus on `h1`, `aria-current`; modified clicks are not intercepted. | Component (browser) | AC-06-2 |
| F06-T3 | Unknown path shows the not-found screen. | Component | AC-06-3 |
| F06-T4 | A failing lazy route and a throwing screen show the error screen without details. | Component | AC-06-4 |
| F06-T5 | axe and the no-horizontal-scroll check on every screen in both themes at the default width, 320 px, and 640 px. | Component (browser), CI | AC-06-5 |
| F06-T6 | The F-05 copy plugin also covers `apps/web/src` (except the copy module and tests). | Lint | AC-06-6 |
| F06-T7 | The build-output check rejects inline scripts, inline event handlers, and `style` attributes; the dependency check confirms the runtime dependencies. | Script | AC-06-7 |
| F06-T8 | Switching the theme updates `data-theme` and persists across a remount. | Component | AC-06-8 |

## Observability requirements

None. F-06 has no server interaction. Client-side error reporting is out of scope.

## Rollback considerations

Front-end source and one runtime dependency only; revert the PRs. No data, API, or infrastructure
is affected, and nothing is deployed until an environment exists (F-03 step 5).

## Implementation plan

1. **Router and shell.** Evaluate and add the router (recorded as a proposed decision with
   evidence, approved before merge); the copy module; the `AppShell` frame with theme toggle; the
   not-found and error screens; title and focus handling; the additive `Navigation` change; the
   copy-plugin and build-output check extensions (F06-T1 to F06-T4, F06-T6 to F06-T8).
2. **Landing page and portal accessibility runs.** The landing page with the approved copy (D-27),
   and browser tests of every portal screen in both themes at the default width, 320 px, and
   640 px, in the pinned Playwright image in CI (F06-T5). Depends on unit 1 and D-27.

Each unit is one pull request.

Implementation notes (unit 1: router and shell):

- React Router 8.5.0 (`react-router`, MIT) is added with an exact pin as the portal's only new
  runtime dependency, recorded as A-27 (proposed; approval needed before merge). Evidence: it adds
  `react-router` 8.5.0, `cookie-es` 3.1.1, and `@remix-run/route-pattern` 0.22.1, all MIT, none
  with an install script; no existing package version changed; no `overrides`; `npm audit` reports
  0 vulnerabilities. The main chunk grows from 220 KB to 323 KB (68 KB to 102 KB gzip); the landing
  and not-found screens load as separate chunks.
- `apps/web/src/routes.tsx` builds the tree with `portalRoutes(pages)`: `PortalShell` (the
  `AppShell` frame with the theme toggle and navigation) wraps a pathless route whose
  `errorElement` is the error screen, so the shell stays visible when a page's chunk fails to load
  or a screen throws. Tests reuse `portalRoutes` to inject failing pages. `App` creates the browser
  router once, inside the component.
- Each screen sets the document title in a layout effect (`usePageTitle`), so the error and
  not-found screens set it too. After each navigation (not the first load), a layout effect moves
  focus to the page's `h1`, made focusable with `tabIndex = -1`. A passive effect moved focus only
  after the new page was already visible, which the browser test caught.
- `shouldRoute` routes only plain left clicks on same-origin paths; Ctrl-, Meta-, Shift-, and
  Alt-clicks keep the browser's default.
- `scripts/check-web-build.sh` now rejects inline scripts, inline event handlers, `style`
  attributes, and `style` elements in the built HTML (seven new test cases), and
  `scripts/check-portal-dev-deps.sh` fails unless the portal's runtime dependencies are exactly
  `@lamassu-cla/design-system`, `react`, `react-dom`, and `react-router`. The copy plugin now covers
  `apps/web/src` except `copy/` and tests (four new test cases).
- Portal browser tests (`apps/web/vitest.browser.config.ts`, 7 tests) run in `make test-browser`
  and in the CI browser job. Mutations that failed them: routing modified clicks, not moving focus,
  removing the error element, showing the error message, and not marking the current link.

Implementation notes (unit 2: landing page and portal accessibility runs):

- The landing page shows the D-27 text, decided by the repository owner on 2026-10-22: the title
  "Lamassu Contributor Agreements" and the description "Complete and review the contributor
  agreements required for your Lamassu contributions." The title differs from the brand, so the
  document title is `Lamassu Contributor Agreements – Lamassu CLA`.
- `apps/web/src/accessibility.browser.test.tsx` (F06-T5) renders the landing, not-found, and error
  screens in both themes at 1280, 640 (200% zoom of a 1280 px window), and 320 px: 18 runs. Each
  runs axe with the WCAG 2.0 to 2.2 A and AA tags on the whole document and fails on any
  violation, then fails if the page scrolls horizontally. `axe-core` 4.14.0 is a portal development
  dependency, the version approved in D-25; the portal dependency check still passes.
- A page header that could not reflow failed the 320 px and 640 px runs, and an unnamed link
  failed axe (`link-name`) on every not-found run.
- Portal browser tests: 25 (7 routing and shell tests and 18 accessibility runs).

## Open questions

- D-12: brand assets; the shell shows the brand as text until confirmed.
- D-21: portal languages; English only, with all text in one copy module.
- D-26: whether the portal links to its source code (AGPL-3.0 section 13), and where.
- D-27: the landing-page copy (decided 2026-10-22).
