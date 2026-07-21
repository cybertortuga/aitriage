package engine

import (
	"regexp"
	"strings"
)

// This file implements deterministic detection of hardcoded secrets and
// credential/data leaks. It combines two strategies:
//
//  1. High-precision signatures for well-known credential formats (AWS keys,
//     GitHub/Slack/Stripe/OpenAI/… tokens, private-key blocks, JWTs, and
//     connection strings that embed a password). These fire regardless of the
//     surrounding variable name or entropy, because the format itself is proof.
//
//  2. A keyword-guarded assignment heuristic (`api_key = "…"`, `password: "…"`)
//     for random-looking secrets that do not match a known vendor format.
//
// All detected values are collected (not just the first per file), so multiple
// secrets in one file are all reported.

// secretHit is a single detected secret occurrence.
type secretHit struct {
	id       string
	name     string
	line     int
	evidence string
	severity string
	varName  string // best-effort identifier for leak-to-logger correlation
}

type secretSignature struct {
	id       string
	name     string
	re       *regexp.Regexp
	severity string
}

// secretSignatures are ordered most-specific first so overlapping matches
// (e.g. sk-ant- vs sk-) are attributed to the more precise vendor.
var secretSignatures = []secretSignature{
	{"SECRET-PRIVATE-KEY", "Private Key Block", regexp.MustCompile(`-----BEGIN (?:RSA |EC |DSA |OPENSSH |PGP )?PRIVATE KEY-----`), "CRITICAL"},
	{"SECRET-AWS-AKID", "AWS Access Key ID", regexp.MustCompile(`\b(?:A3T[A-Z0-9]|AKIA|ASIA|ABIA|ACCA)[0-9A-Z]{16}\b`), "CRITICAL"},
	{"SECRET-GH-PAT", "GitHub Fine-grained PAT", regexp.MustCompile(`\bgithub_pat_[0-9A-Za-z_]{40,}`), "CRITICAL"},
	{"SECRET-GH-TOKEN", "GitHub Token", regexp.MustCompile(`\bgh[pousr]_[0-9A-Za-z]{36}\b`), "CRITICAL"},
	{"SECRET-SLACK-TOKEN", "Slack Token", regexp.MustCompile(`\bxox[baprs]-[0-9A-Za-z-]{10,48}\b`), "CRITICAL"},
	{"SECRET-SLACK-WEBHOOK", "Slack Webhook URL", regexp.MustCompile(`https://hooks\.slack\.com/services/[A-Za-z0-9_/]{20,}`), "HIGH"},
	{"SECRET-ANTHROPIC", "Anthropic API Key", regexp.MustCompile(`\bsk-ant-[A-Za-z0-9_\-]{20,}`), "CRITICAL"},
	{"SECRET-OPENAI", "OpenAI API Key", regexp.MustCompile(`\bsk-(?:proj-)?[A-Za-z0-9_\-]{20,}`), "CRITICAL"},
	{"SECRET-STRIPE", "Stripe Secret Key", regexp.MustCompile(`\b[rs]k_live_[0-9A-Za-z]{16,}\b`), "CRITICAL"},
	{"SECRET-SENDGRID", "SendGrid API Key", regexp.MustCompile(`\bSG\.[A-Za-z0-9_\-]{16,}\.[A-Za-z0-9_\-]{16,}\b`), "CRITICAL"},
	{"SECRET-GOOGLE-API", "Google API Key", regexp.MustCompile(`\bAIza[0-9A-Za-z_\-]{30,45}`), "HIGH"},
	{"SECRET-NPM", "npm Token", regexp.MustCompile(`\bnpm_[0-9A-Za-z]{36}\b`), "CRITICAL"},
	{"SECRET-TWILIO", "Twilio API Key", regexp.MustCompile(`\bSK[0-9a-fA-F]{32}\b`), "CRITICAL"},
	{"SECRET-JWT", "JSON Web Token", regexp.MustCompile(`\beyJ[A-Za-z0-9_\-]{10,}\.eyJ[A-Za-z0-9_\-]{10,}\.[A-Za-z0-9_\-]{10,}`), "HIGH"},
	{"SECRET-CONN-STRING", "Credentialed Connection String", regexp.MustCompile(`\b(?:postgres|postgresql|mysql|mongodb(?:\+srv)?|redis|rediss|amqps?|mssql)://[^:@\s/"']+:([^@\s/"']+)@`), "CRITICAL"},
}

// credentialAssignRegex matches `secretName = "value"` / `secretName: "value"`.
var credentialAssignRegex = regexp.MustCompile(
	`(?i)(?:^|[\s,({])(api[_-]?key|secret|token|password|passwd|passphrase|private[_-]?key|auth[_-]?key|access[_-]?key|secret[_-]?key|client[_-]?secret|signing[_-]?key|db[_-]?pass(?:word)?|credentials?)\s*[=:]\s*["']([^"'\n]{6,})["']`,
)

// placeholderValueRegex matches values that are clearly not real secrets:
// template interpolation, env lookups, or obvious dummy words. It intentionally
// does NOT reject bare `$`/`%`/`{` characters, since strong passwords contain
// those — only actual interpolation/format syntax.
var placeholderValueRegex = regexp.MustCompile(`(?i)(\$\{|\{\{|%\(|%[svd]\b|process\.env|os\.(getenv|environ)|getenv\(|^(your[_-]|change[_-]?me|replace[_-]|example|placeholder|dummy|sample|xxx+|test|todo|none|null|undefined|foobar|lorem)$)`)

// detectSecretHits returns every secret occurrence in src. Line numbers are
// 1-based. It does not apply suppression or file-type gating — callers do that.
//
// includeAssign controls the keyword-guarded assignment heuristic: callers pass
// true only for code and .env files, where `name = "value"` assignments are
// meaningful, and false for structured config where that heuristic is noisier
// (vendor-format signatures still run there).
func detectSecretHits(src string, includeAssign bool) []secretHit {
	var hits []secretHit
	seen := map[string]bool{} // dedup by "line:matchedText"

	add := func(id, name, matched, severity string, idx int, varName string) {
		line := strings.Count(src[:idx], "\n") + 1
		key := matched
		if len(key) > 24 {
			key = key[:24]
		}
		dedup := itoa(line) + ":" + key
		if seen[dedup] {
			return
		}
		seen[dedup] = true
		hits = append(hits, secretHit{
			id:       id,
			name:     name,
			line:     line,
			evidence: redactSecret(matched),
			severity: severity,
			varName:  varName,
		})
	}

	// 1. Known-format signatures.
	for _, sig := range secretSignatures {
		for _, loc := range sig.re.FindAllStringIndex(src, -1) {
			matched := src[loc[0]:loc[1]]
			add(sig.id, sig.name, matched, sig.severity, loc[0], "")
		}
	}

	// 2. Keyword-guarded assignment heuristic (code / .env only).
	if !includeAssign {
		return hits
	}
	for _, loc := range credentialAssignRegex.FindAllStringSubmatchIndex(src, -1) {
		if len(loc) < 6 {
			continue
		}
		varName := src[loc[2]:loc[3]]
		val := src[loc[4]:loc[5]]
		if !isLikelySecretValue(val) {
			continue
		}
		line := strings.Count(src[:loc[4]], "\n") + 1
		key := val
		if len(key) > 24 {
			key = key[:24]
		}
		dedup := itoa(line) + ":" + key
		if seen[dedup] { // already reported by a vendor-format signature
			continue
		}
		seen[dedup] = true
		hits = append(hits, secretHit{
			id:       "SECRET-ASSIGN",
			name:     "Hardcoded Secret Assignment",
			line:     line,
			evidence: redactSecret(val),
			severity: "CRITICAL",
			varName:  varName,
		})
	}
	return hits
}

// isLikelySecretValue decides whether a value assigned to a secret-named
// variable is a real secret rather than a placeholder or env reference.
func isLikelySecretValue(val string) bool {
	v := strings.TrimSpace(val)
	if len(v) < 8 {
		return false
	}
	if placeholderValueRegex.MatchString(v) {
		return false
	}
	// A plain URL without embedded credentials is not itself a secret;
	// credentialed connection strings are caught by a dedicated signature.
	if (strings.HasPrefix(v, "http://") || strings.HasPrefix(v, "https://")) && !strings.Contains(v, "@") {
		return false
	}
	ent := CalculateShannonEntropy(v)
	// High-entropy random blob (API key / token) or a reasonably complex
	// password. Structured vendor tokens are already caught by signatures.
	if len(v) >= 16 && ent > 3.5 {
		return true
	}
	if len(v) >= 12 && ent > 4.0 {
		return true
	}
	return false
}

// redactSecret returns a safe, non-reversible preview of a detected secret so
// evidence never re-emits the full credential.
func redactSecret(s string) string {
	s = strings.TrimSpace(s)
	if len(s) <= 4 {
		return "****"
	}
	prefix := s
	if len(prefix) > 4 {
		prefix = prefix[:4]
	}
	return prefix + "…(redacted, " + itoa(len(s)) + " chars)"
}

// itoa is a tiny strconv.Itoa to avoid an extra import churn in this file.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
