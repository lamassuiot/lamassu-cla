# Design Tokens and Design-System Guardrails

| Field | Value |
| --- | --- |
| Status | Draft — values proposed; final brand assets and fonts pending D-12 |
| Visual direction | Decision A-11 in the [decision log](../../docs/decisions/decision-log.md) |
| Implementation | `packages/design-system` (Phase 3, [F-04](../features/F-04-design-tokens-and-themes.md)) |

The visual language is a professional enterprise SaaS style: deep navy or blue-green backgrounds,
warm orange accents, white typography, rounded cards, subtle borders and gradients, strong
hierarchy, and consistent light and dark themes. It is inspired by Lamassu-style references and must
not copy third-party logos or proprietary branding.

## 1. Token architecture

Tokens have three tiers. Components use only semantic or component tokens, never primitives.

| Tier | Example | Used by |
| --- | --- | --- |
| Primitive | `color.navy.950` | Semantic tokens only. |
| Semantic | `color.bg.canvas`, `color.text.primary` | Components and layouts. Values change per theme. |
| Component | `button.primary.bg` | One component. Defined from semantic tokens. |

- Source of truth: typed token definitions in `packages/design-system/src/tokens/`.
- Output: CSS custom properties prefixed `--cla-`, for example `--cla-color-bg-canvas`, scoped under
  `[data-theme="dark"]` and `[data-theme="light"]`.
- Theme selection: user choice persisted locally, defaulting to the operating-system preference
  (`prefers-color-scheme`).

## 2. Color

### 2.1 Primitives

| Token | Value | Token | Value |
| --- | --- | --- | --- |
| `navy.950` | `#0B1622` | `orange.400` | `#F78C40` |
| `navy.900` | `#0F1F2E` | `orange.500` | `#F2711C` |
| `navy.800` | `#162B3D` | `orange.600` | `#D9600F` |
| `navy.700` | `#1D3448` | `orange.700` | `#B4500A` |
| `navy.600` | `#2A4760` | `gray.50` | `#F7F5F2` |
| `slate.500` | `#5C7A94` | `gray.100` | `#EEF1F4` |
| `teal.950` | `#0D2A2E` | `gray.200` | `#DDE2E8` |
| `teal.800` | `#145055` | `gray.300` | `#C9D2DC` |
| `white` | `#FFFFFF` | `gray.400` | `#B8C4D0` |
| | | `gray.500` | `#7A8896` |
| | | `gray.700` | `#4A5B6C` |
| | | `gray.800` | `#3D4F61` |

### 2.2 Semantic colors

| Semantic token | Dark theme | Light theme |
| --- | --- | --- |
| `bg.canvas` | `navy.950` | `gray.50` |
| `bg.surface` (cards) | `navy.800` | `white` |
| `bg.raised` (popovers, dialogs) | `navy.700` | `white` |
| `bg.subtle` | `navy.900` | `gray.100` |
| `bg.nav` | `navy.900` | `teal.950` |
| `text.primary` | `white` | `navy.950` |
| `text.secondary` | `gray.400` | `gray.700` |
| `text.heading` | `white` | `teal.950` |
| `text.on-nav` | `white` | `white` |
| `text.accent` | `orange.400` | `orange.700` |
| `text.on-accent` | `navy.950` | `navy.950` |
| `border.subtle` (decorative) | `navy.600` | `gray.300` |
| `border.control` (inputs) | `slate.500` | `gray.500` |
| `action.primary.bg` | `orange.500` | `orange.500` |
| `action.primary.bg-hover` | `orange.400` | `orange.600` |
| `focus.ring` | `orange.400` | `orange.700` |

Primary buttons use **navy text on orange**. White text on `orange.500` is only 2.94:1 and fails
WCAG AA, so it is not allowed.

### 2.3 Status colors

Status badges use a text color on a tinted background.

| Status | Dark text / background | Light text / background |
| --- | --- | --- |
| `success` | `#4CC38A` / `#10302A` | `#1E7A4F` / `#E6F4EC` |
| `warning` | `#F2B84B` / `#33280F` | `#8A5A00` / `#FDF3DC` |
| `danger` | `#F28B82` / `#3A1A1C` | `#B42318` / `#FDE8E6` |
| `info` | `#7DB4F0` / `#132C47` | `#1F5FA8` / `#E6EFFA` |
| `neutral` | `#B8C4D0` / `#1F3447` | `#3D4F61` / `#EAEEF2` |

Status is never conveyed by color alone; badges always include text and, where useful, an icon.

| Domain state | Status color |
| --- | --- |
| `ACTIVE`, `PASSED`, `PUBLISHED` | `success` |
| `PENDING_SIGNATURE`, `VERIFYING`, `PENDING_REVIEW`, `ACTION_REQUIRED` | `warning` |
| `FAILED`, `REVOKED`, `REJECTED`, `DECLINED` | `danger` |
| `DRAFT`, informational notices | `info` |
| `SUSPENDED`, `EXPIRED`, `DEPRECATED`, `ARCHIVED`, `SUPERSEDED`, `SKIPPED` | `neutral` |

### 2.4 Verified contrast

Ratios computed with the WCAG 2.2 relative-luminance formula. Text needs at least 4.5:1; control
borders and focus indicators need at least 3:1.

| Pair | Dark | Light |
| --- | --- | --- |
| `text.primary` on `bg.canvas` | 18.23 | 16.75 |
| `text.primary` on `bg.surface` | 14.51 | 18.23 |
| `text.secondary` on `bg.canvas` | 10.28 | 6.42 |
| `text.secondary` on `bg.surface` | 8.19 | 6.99 |
| `text.accent` on `bg.canvas` | 7.62 | 4.72 |
| `text.accent` on `bg.surface` | 6.07 | 5.13 |
| `text.on-accent` on `action.primary.bg` | 6.21 | 6.21 |
| `text.on-accent` on `action.primary.bg-hover` | 7.62 | 4.89 |
| `text.on-nav` on `bg.nav` | 16.72 | 15.13 |
| `border.control` on `bg.canvas` | 4.05 | 3.33 |
| `border.control` on `bg.surface` | 3.22 | 3.63 |
| `focus.ring` on `bg.canvas` | 7.62 | 4.72 |
| Status badges (lowest) | 6.41 | 4.69 |

`border.subtle` is decorative only and is never the sole boundary of an interactive control.

### 2.5 Gradients

Restrained, for hero areas and page headers only:

- `gradient.hero.dark`: `navy.950` to `teal.950`, 135 degrees.
- `gradient.hero.light`: `gray.50` to `gray.100`, 180 degrees.

Text over gradients must meet the contrast of the darkest or lightest stop, whichever is worse.

## 3. Typography

Font families are proposed pending D-12. Only open-licensed fonts may be bundled. Until D-12 is
confirmed, no font files are bundled and both font tokens resolve to their fallback stacks
([F-04](../features/F-04-design-tokens-and-themes.md)).

| Token | Value |
| --- | --- |
| `font.family.sans` | `Inter` (SIL Open Font License), then system UI fallback stack |
| `font.family.mono` | `JetBrains Mono` (SIL Open Font License), then system monospace fallback |

| Token | Size / line height | Weight | Use |
| --- | --- | --- | --- |
| `text.display` | 48 / 56 px | 700 | Landing hero only. |
| `text.h1` | 36 / 44 px | 700 | Page title. |
| `text.h2` | 28 / 36 px | 600 | Section heading. |
| `text.h3` | 22 / 30 px | 600 | Card heading. |
| `text.body-lg` | 18 / 28 px | 400 | Lead paragraphs. |
| `text.body` | 16 / 24 px | 400 | Default. |
| `text.body-sm` | 14 / 20 px | 400 | Secondary text, tables. |
| `text.label` | 12 / 16 px | 600, uppercase, letter spacing 0.08 em | Section markers, badges. |
| `text.code` | 14 / 20 px | 400 | Hashes, identifiers. |

Sizes are defined in `rem` in code. Text must reflow at 200% zoom without loss of content.

## 4. Spacing

Base unit 4 px. Allowed values only:

| Token | Value |
| --- | --- |
| `space.0` | 0 |
| `space.1` | 4 px |
| `space.2` | 8 px |
| `space.3` | 12 px |
| `space.4` | 16 px |
| `space.5` | 24 px |
| `space.6` | 32 px |
| `space.7` | 48 px |
| `space.8` | 64 px |
| `space.9` | 96 px |

## 5. Radius, borders, shadows

| Token | Value | Use |
| --- | --- | --- |
| `radius.sm` | 6 px | Inputs, badges. |
| `radius.md` | 10 px | Buttons, menus. |
| `radius.lg` | 16 px | Cards, dialogs. |
| `radius.pill` | 9999 px | Pills, avatars. |
| `border.width` | 1 px | Default. |
| `border.width-strong` | 2 px | Selected states. |

| Token | Dark theme | Light theme |
| --- | --- | --- |
| `shadow.sm` | `0 1px 2px rgb(0 0 0 / 0.40)` | `0 1px 2px rgb(11 22 34 / 0.08)` |
| `shadow.md` | `0 4px 12px rgb(0 0 0 / 0.45)` | `0 4px 12px rgb(11 22 34 / 0.10)` |
| `shadow.lg` | `0 12px 32px rgb(0 0 0 / 0.50)` | `0 12px 32px rgb(11 22 34 / 0.14)` |

## 6. Motion

| Token | Value |
| --- | --- |
| `motion.duration.fast` | 120 ms |
| `motion.duration.base` | 200 ms |
| `motion.duration.slow` | 320 ms |
| `motion.easing.standard` | `cubic-bezier(0.2, 0, 0, 1)` |
| `motion.easing.exit` | `cubic-bezier(0.3, 0, 1, 1)` |

With `prefers-reduced-motion: reduce`, transitions are limited to opacity and durations to
`motion.duration.fast` or less.

## 7. Breakpoints and layout

| Token | Minimum width |
| --- | --- |
| `bp.sm` | 640 px |
| `bp.md` | 768 px |
| `bp.lg` | 1024 px |
| `bp.xl` | 1280 px |

- Mobile-first. Navigation collapses to a drawer below `bp.lg`.
- Maximum content width 1200 px; reading width for agreement text 72 characters.

## 8. Focus

- Every interactive element shows a visible focus indicator: 2 px `focus.ring` outline with 2 px
  offset, using `:focus-visible`.
- Focus is never removed without a replacement. Focus order follows visual order.

## 9. Components

Each component is implemented once in `packages/design-system` and reused:

App shell, navigation, page header, section label, button (primary, secondary, ghost, danger),
input, select, checkbox, radio, textarea, form field, dialog, confirmation dialog, drawer, card,
table, pagination, status badge, timeline item, step indicator, empty state, error state,
permission-denied state, loading state, toast and live-region announcer, agreement status display,
audit-event display, and hash display.

## 10. Guardrails

| Rule | Enforcement |
| --- | --- |
| Components are implemented once and reused. Pages compose components and do not define one-off button, card, or form styles. | Code review; Biome rule and plugin forbidding inline `style` attributes in `apps/web` ([ADR-0011](../../docs/decisions/0011-biome-for-javascript-and-typescript.md)). |
| Only tokens are used for visual values. | Biome GritQL plugin `.biome/plugins/no-raw-css-values.grit` rejects raw colors, color functions, named colors, raw spacing, radius, border-width, font-size, and shadow lengths, and raw durations in every CSS file except the generated `tokens.css`; `scripts/check-css-suppressions.sh` rejects Biome suppressions in CSS ([F-04](../features/F-04-design-tokens-and-themes.md), D-23, A-17). Biome also forbids inline `style` attributes in `apps/web`. |
| A new color requires a token change and an update to this specification with a contrast check. | Change control in [the SDD workflow](../sdd-workflow.md#3-change-control). |
| A new spacing value requires written justification in the PR and an update to this specification. | Code review. |
| Every component has a preview story in light and dark themes, including loading, empty, error, and disabled states where applicable. | Storybook coverage check in CI. |
| Accessibility is tested per component. | Automated axe checks in Storybook and Playwright. |
| Visual drift is detected. | Visual regression snapshots for the primary component set. |
