package runstore

import "regexp"

// Secret-shaped patterns are redacted from every persisted request, response,
// and error before it touches disk. The run bundle is triage evidence, not a
// credential store: even though the host session never shares its tokens, the
// prompts and model answers can quote code that embeds secrets.
var (
	jwtRegex         = regexp.MustCompile(`eyJ[A-Za-z0-9_-]{10,}\.eyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{8,}`)
	knownSecretRegex = regexp.MustCompile(`(?i)\b(?:AKIA[0-9A-Z]{16}|gh[ps]_[A-Za-z0-9_]{20,}|glpat-[A-Za-z0-9\-_]{20,}|sk_live_[A-Za-z0-9]{16,}|SG\.[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,})\b`)
	namedSecretRegex = regexp.MustCompile("(?i)((?:password|passwd|token|secret|api[_-]?key|jwt)[^\\n:=]{0,40}\\s*[:=]\\s*)[`\"']?[^`\"'\\s]{8,}([`\"']?)")
)

// Redact returns text with secret-shaped substrings replaced by stable
// placeholders. It is non-reversible and deterministic.
func Redact(text string) string {
	text = jwtRegex.ReplaceAllString(text, "[REDACTED_JWT]")
	text = knownSecretRegex.ReplaceAllString(text, "[REDACTED_SECRET]")
	text = namedSecretRegex.ReplaceAllString(text, "${1}[REDACTED]")
	return text
}
