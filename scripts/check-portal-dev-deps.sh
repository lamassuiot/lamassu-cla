#!/usr/bin/env bash
# Checks that development-only tooling never reaches the portal (F-05 AC-05-12, decision D-25):
# the production dependency tree and the bundled modules of a source-mapped portal build must not
# include axe-core, caniuse-lite, Storybook, Playwright, or Vitest.
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"
forbidden='(^|/)(axe-core|caniuse-lite|storybook|@storybook/[^/]+|playwright|playwright-core|vitest|@vitest/[^/]+)(/|$)'
status=0

if hits=$(npm ls --omit=dev --all --parseable -w @lamassu-cla/web 2>/dev/null | grep -oE "node_modules/.*" | grep -E "$forbidden"); then
  echo "Production dependencies of @lamassu-cla/web include development-only tooling:"
  echo "$hits"
  status=1
fi

out="$(mktemp -d)"
trap 'rm -rf "$out"' EXIT
npx vite build apps/web --sourcemap --outDir "$out" --emptyOutDir --logLevel error
if hits=$(node -e '
  const fs = require("node:fs");
  const path = require("node:path");
  const pattern = new RegExp(process.argv[2]);
  const hits = new Set();
  const walk = (dir) => {
    for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
      const file = path.join(dir, entry.name);
      if (entry.isDirectory()) walk(file);
      else if (file.endsWith(".map")) {
        for (const source of JSON.parse(fs.readFileSync(file, "utf8")).sources ?? []) {
          const module = source.split("node_modules/").pop();
          if (source.includes("node_modules/") && pattern.test(module)) hits.add(module);
        }
      }
    }
  };
  walk(process.argv[1]);
  console.log([...hits].join("\n"));
  process.exit(hits.size ? 0 : 1);
' "$out" "$forbidden"); then
  echo "The portal bundle includes development-only modules:"
  echo "$hits"
  status=1
fi

[[ $status -eq 0 ]] && echo "The portal bundle excludes development-only tooling."
exit "$status"
