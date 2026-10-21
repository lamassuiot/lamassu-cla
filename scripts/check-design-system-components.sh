#!/usr/bin/env bash
# Checks the design-system component conventions (F-05 AC-05-1, AC-05-2): every component directory
# has its component, a story file, and an export from src/index.ts, and every CSS Module keeps all of
# its rules inside `@layer components`.
set -euo pipefail

root="${1:-packages/design-system}"
components="$root/src/components"
[[ -d "$components" ]] || exit 0

status=0
fail() {
  echo "$1"
  status=1
}

for dir in "$components"/*/; do
  name="$(basename "$dir")"
  [[ -f "$dir$name.tsx" ]] || fail "$dir: missing $name.tsx"
  [[ -f "$dir$name.stories.tsx" ]] || fail "$dir: missing $name.stories.tsx; every component needs a story"
  grep -qE "from '\./components/$name/$name(\.tsx)?'" "$root/src/index.ts" ||
    fail "$dir: not exported from $root/src/index.ts"
done

while IFS= read -r -d '' css; do
  if ! node -e '
    const text = require("node:fs").readFileSync(process.argv[1], "utf8").replace(/\/\*[\s\S]*?\*\//g, "");
    const open = text.indexOf("{");
    if (text.slice(0, open).trim() !== "@layer components") process.exit(1);
    let depth = 0;
    for (let i = open; i < text.length; i++) {
      if (text[i] === "{") depth++;
      if (text[i] === "}" && --depth === 0) process.exit(text.slice(i + 1).trim() === "" ? 0 : 1);
    }
    process.exit(1);
  ' "$css"; then
    fail "$css: all rules must be inside a single @layer components block"
  fi
done < <(find "$components" -name '*.module.css' -print0)

exit "$status"
