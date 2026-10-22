#!/usr/bin/env bash
# Checks the built portal (F-04 AC-04-9): no bundled font files, and nothing that loads a resource
# from another origin (absolute or protocol-relative URLs in HTML src/href, CSS url() or @import,
# @font-face, or dynamic import()). URL strings that are never fetched, such as XML namespaces in
# React, are not load points and are allowed.
set -euo pipefail

dist="${1:-apps/web/dist}"
[[ -d "$dist" ]] || {
  echo "$dist does not exist; run the build first."
  exit 1
}

status=0
fail() {
  echo "$1"
  status=1
}

while IFS= read -r -d '' font; do
  fail "${font}: font file in the build output; fonts are pending D-12."
done < <(find "$dist" -type f \( -iname '*.woff' -o -iname '*.woff2' -o -iname '*.ttf' -o -iname '*.otf' -o -iname '*.eot' \) -print0)

remote='(https?:)?//'
check() {
  local glob="$1" pattern="$2" message="$3" hits
  # find succeeds with no output when no file matches the glob.
  if hits=$(find "$dist" -type f -name "$glob" -exec grep -oHiE "$pattern" {} +) && [[ -n "$hits" ]]; then
    while IFS= read -r hit; do fail "${hit%%:*}: $message (${hit#*:})"; done <<<"$hits"
  fi
}

check '*.html' "(src|href)=[\"']?${remote}[^\"' >]*" 'loads a resource from another origin'
check '*.css' "url\\([\"']?${remote}[^)]*\\)" 'loads a resource from another origin'
check '*.css' "@import[^;]*${remote}[^;]*" 'imports a stylesheet from another origin'
check '*.css' '@font-face' 'declares a font; fonts are pending D-12'
check '*.js' "import\\([\"'\`]${remote}[^)]*\\)" 'imports a module from another origin'

# F-06 AC-06-7: the HTML must work under a Content Security Policy without 'unsafe-inline'.
if hits=$(find "$dist" -type f -name '*.html' -exec grep -oHiE '<script[^>]*>' {} + | grep -viE '[[:space:]]src='); then
  while IFS= read -r hit; do fail "${hit%%:*}: has an inline script (${hit#*:})"; done <<<"$hits"
fi
check '*.html' '<[^>]*[[:space:]]on[a-z]+[[:space:]]*=' 'has an inline event handler'
check '*.html' '<[^>]*[[:space:]]style[[:space:]]*=' 'has a style attribute'
check '*.html' '<style[[:space:]>]' 'has an inline style element'

exit "$status"
