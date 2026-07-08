package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// securecodingGuidelines contains mandatory secure coding rules adapted from
// SecureCoder's securecoder_generation skill. These guidelines are exposed as
// an MCP resource so any AI client (Claude, Cursor, Antigravity) can read them.
//
// Source: SecureCoder plugin skills/securecoder_generation/SKILL.md (553 lines)
// Adapted: Removed SecureCoder-specific branding, added AITriage context.
const securecodingGuidelines = `# AITriage Secure Coding Guidelines

> **CRITICAL**: You MUST follow these rules for ALL code generation and security remediation tasks.
> These rules cover database access, file handling, session management, and frontend rendering
> to prevent critical vulnerabilities. Do NOT ignore any section.

## General Principles

- **Input Validation & Sanitization**: Treat all external data as untrusted.
  Validate inputs against an allow-list of expected types, lengths, and formats.
- **Output Encoding**: Convert data into a safe format before sending to the client.
- **Authentication & Authorization**: Require strong authentication for non-public pages.
  Use OAuth 2.0 or OpenID Connect. Never hardcode JWT secrets, API keys, or fallback strings.
- **Password Management**: Use Argon2 or bcrypt with unique per-user salts. Never plaintext.
- **Least Privilege**: Grant minimum permissions necessary. Use RBAC.
- **Secure Sessions**: Cryptographically strong random session IDs on server.
  HttpOnly, Secure flags on cookies. Short inactivity timeouts.
- **Path & File Security**: Never trust user input in file paths. Use path.basename()
  to strip traversal sequences. Enforce trailing slash in directory boundary checks.
- **Command Execution**: Never pass unvalidated user input to exec/spawn.
  Validate binary paths and arguments against a strict hardcoded allow-list.
- **Error Handling**: Generic messages to users, detailed logs for developers.
  Logs MUST NOT contain passwords, tokens, or session IDs.
- **Encryption**: TLS 1.2+ in transit, encrypt at rest.
- **Fail Safe**: When something fails, deny access.
- **Cryptography**: Use established libraries, authenticated encryption, secure PRNG.
- **Deserialization**: Never use insecure deserialization formats.
- **Code Loading**: Never remotely load code without verifying origin.

## Testing

- Servers **MUST** listen on localhost or 127.0.0.1 when testing.
- Servers **MUST NOT** listen on 0.0.0.0.

---

## XSS Prevention

### Framework-Native Escaping
- **MUST** escape untrusted data in all outgoing: HTML, JS, CSS, HTTP headers.
- **MUST** rely on framework auto-escaping (React JSX, Angular interpolation).
- **MUST** quote HTML attributes with variables: ` + "`" + `<div class="{{ var }}">` + "`" + `
- **MUST NOT** use dangerouslySetInnerHTML or bypassSecurityTrustHtml without DOMPurify.
- **MUST** use DOMPurify when rendering unavoidable raw HTML.

### Vanilla JS DOM
- **MUST NOT** use innerHTML, outerHTML, document.write, insertAdjacentHTML.
- **MUST** use textContent or innerText for text insertion.
- **MUST** use createElement/setAttribute/appendChild for DOM structure.
- **MUST** use element.replaceChildren() to clear content.
- **MUST** use DOMParser for static structures (SVGs).

---

## Storage & Session

- **MUST NOT** store auth tokens in localStorage/sessionStorage (XSS exposure).
- **MUST** use HttpOnly, Secure, SameSite=Lax cookies for session management.
- **MUST** clear client state on logout. Trigger full page reload.

---

## Content Security Policy (CSP)

- **MUST** implement strict CSP. Use nonces for inline scripts.
- **MUST NOT** use unsafe-inline or unsafe-eval without explicit security review.
- **MUST** use SRI hashes for non-first-party CDN assets.
- **MUST** configure X-Frame-Options: DENY and CSP frame-ancestors.

---

## Data Handling

- **MUST NOT** surface full PII in UI (mask as ***-***-1234).
- **MUST NOT** log structured user objects or tokens via console.log.
- **MUST NOT** use alert()/confirm()/prompt() in production.
- **MUST** verify API communication uses HTTPS.

---

## Backend Session Management

### Passwords
- **MUST** validate password strength (min 8 chars, 12+ recommended).
- **MUST** store with memory-hard hashing (Argon2, scrypt) + unique salts.
- **MUST** implement CSRF tokens for login/logout/signup endpoints.
- **MUST NOT** send credentials in URL parameters.
- **MUST NOT** log credentials, even for failed attempts.

### Sessions
- **MUST** use framework built-in session management.
- **MUST** set session expiration. No infinite sessions.
- **MUST** invalidate all sessions on logout/account deletion/org removal.

---

## Authentication & Authorization

- **MUST** authenticate all APIs. Rate limit all APIs.
- **MUST** harden cookies: __Host- prefix, SameSite, Secure, HttpOnly.
- **MUST** for JWT: reject 'none' algo, hardcode expected algo, use crypto RNG for secrets, validate exp.
- **MUST** implement CSRF for all state-changing requests (POST, PUT, DELETE, PATCH).
- **MUST NOT** disable framework CSRF protection (@csrf_exempt in Django = CRITICAL violation).
- **MUST NOT** store secrets in code. No hardcoded literals or literal fallbacks.
- **MUST** authenticate server-side, not client-side.
- **MUST** validate resource ownership on every request.

### HTTP Headers
- **MUST** use allow-list of HTTP methods. Disable TRACE, PUT, DELETE if unused.
- **MUST** set strict CSP, X-Content-Type-Options: nosniff, X-Frame-Options: DENY.
- **MUST** disable unused browser features (camera, microphone, geolocation).
- **MUST** use strict CORS policy. No wildcard origins (*).

---

## File Uploads

- **MUST** validate extension AND content (magic bytes).
- **MUST** use allow-list of permitted file types.
- **MUST** impose size limits (1-10MB).
- **MUST** generate unique filenames (UUID/hash). Store originals in DB.
- **MUST** store outside web root in non-executable directory.
- **MUST** serve with Content-Disposition: attachment, X-Content-Type-Options: nosniff.
- **MUST** validate zip paths for directory traversal.
- **MUST** harden XML parsing: disable external entities, DTD, XInclude, network requests.

---

## Database Security

### SQL Injection
- **MUST NOT** use string concatenation for SQL queries.
- **MUST** use parameterized queries, prepared statements, or ORMs.
- **MUST NOT** trust strings from database — escape before sending to client.
- **MUST NOT** expose SQL errors to users.

### Database Configuration
- **MUST** use least-privilege DB user (SELECT only if read-only).
- **MUST NOT** use root/admin accounts for web apps.
- **MUST** use mTLS for DB connections.
- **MUST** isolate databases per application.

---

## Security Review & Planning

- **MUST** add a security section to every Verification Plan.
- **MUST NOT** omit security measures without explaining why (use TODO(security) comments).
`

func registerGuidelinesResource(srv *mcp.Server) {
	srv.AddResource(&mcp.Resource{
		URI:         "aitriage://secure-coding-guidelines",
		Name:        "Secure Coding Guidelines",
		Description: "Mandatory secure coding rules for web applications: XSS, CSRF, SQLi, sessions, auth, file uploads, DB security. Read before generating or reviewing code.",
		MIMEType:    "text/markdown",
	}, func(ctx context.Context, request *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      "aitriage://secure-coding-guidelines",
					MIMEType: "text/markdown",
					Text:     securecodingGuidelines,
				},
			},
		}, nil
	})
}
