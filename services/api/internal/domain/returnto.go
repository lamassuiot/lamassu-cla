package domain

import (
	"strings"
)

// DefaultReturnTo is where sign-in lands when return_to is absent or rejected (AC-07-4).
const DefaultReturnTo = "/"

// ReturnToPolicy validates post-sign-in return paths (F-07, security considerations; T-16).
// Allowed holds canonical absolute paths without query or fragment; a value matches when it
// equals an entry or is a descendant of an entry that ends with "/".
type ReturnToPolicy struct {
	Allowed []string
}

// Resolve returns the canonical path for raw, or DefaultReturnTo when raw is empty or rejected.
// Rejection rules: percent-decode once; reject a remaining "%", a scheme or authority, a leading
// "//" or "/\", any backslash or control character, or a value not starting with "/".
func (p ReturnToPolicy) Resolve(raw string) string {
	path, ok := canonicalReturnTo(raw)
	if !ok {
		return DefaultReturnTo
	}
	for _, allowed := range p.Allowed {
		if path == allowed {
			return path
		}
		// "/" matches only itself; otherwise every path would be its descendant.
		if allowed != "/" && strings.HasSuffix(allowed, "/") && strings.HasPrefix(path, allowed) {
			return path
		}
	}
	return DefaultReturnTo
}

// canonicalReturnTo decodes once and keeps only a plain absolute path made of safe characters.
// It uses no URL parser: the domain layer performs no I/O and imports no net packages, and the
// accepted grammar is deliberately narrower than RFC 3986.
func canonicalReturnTo(raw string) (string, bool) {
	if raw == "" || len(raw) > 2048 {
		return "", false
	}
	decoded, ok := percentDecodeOnce(raw)
	if !ok || strings.Contains(decoded, "%") {
		return "", false
	}
	if !strings.HasPrefix(decoded, "/") || strings.HasPrefix(decoded, "//") || strings.HasPrefix(decoded, "/\\") {
		return "", false
	}
	// Drop query and fragment; the allowlist is about paths.
	if i := strings.IndexAny(decoded, "?#"); i >= 0 {
		decoded = decoded[:i]
	}
	for i := 0; i < len(decoded); i++ {
		if !safePathByte(decoded[i]) {
			return "", false
		}
	}
	if strings.Contains(decoded, "/../") || strings.HasSuffix(decoded, "/..") ||
		strings.Contains(decoded, "/./") || strings.HasSuffix(decoded, "/.") {
		return "", false
	}
	return decoded, true
}

// safePathByte accepts RFC 3986 unreserved characters, sub-delims, "/", ":" and "@". It rejects
// backslashes, control characters, whitespace, and anything that needs encoding.
func safePathByte(c byte) bool {
	switch {
	case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9':
		return true
	}
	return strings.IndexByte("-._~!$&'()*+,;=/:@", c) >= 0
}

// percentDecodeOnce decodes %XX sequences exactly once; malformed escapes fail.
func percentDecodeOnce(s string) (string, bool) {
	if !strings.Contains(s, "%") {
		return s, true
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		if s[i] != '%' {
			b.WriteByte(s[i])
			continue
		}
		if i+2 >= len(s) {
			return "", false
		}
		hi, ok1 := hexValue(s[i+1])
		lo, ok2 := hexValue(s[i+2])
		if !ok1 || !ok2 {
			return "", false
		}
		b.WriteByte(hi<<4 | lo)
		i += 2
	}
	return b.String(), true
}

func hexValue(c byte) (byte, bool) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', true
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, true
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, true
	}
	return 0, false
}
