# RULES EXPANSION PLAN: From 69 to 200+ Rules

## Current State Audit

| Stack | Current Rules | Gap Analysis |
|---|---|---|
| Universal | 17 | Missing: SSRF, NoSQL injection, insecure randomness, insecure deserialization, assert-as-guard, unsafe YAML/XML, logging sensitive data |
| Next.js/React | 21 | Missing: unsafe href (javascript:), Math.random for crypto, eval/Function constructor, NoSQL injection, SSRF, missing helmet |
| FastAPI | 13 | Missing: subprocess shell=True, pickle/marshal, yaml.load, assert-as-guard, SSRF, mass assignment, insecure tempfile |
| Flask | 7 | Missing: SSRF, pickle, IDOR, logging PII, insecure redirect, session fixation |
| Go | 4 | Missing: SSRF, command injection (os/exec), crypto/rand vs math/rand, unsafe pointer, deferred unlock, TLS skip verify, race condition patterns, error swallowing |
| ASP.NET Core | 2 | Missing: XXE, deserialization, path traversal, LDAP injection, hardcoded connection strings |
| Django | 0 (detected but no rules!) | Full gap: DEBUG, ALLOWED_HOSTS, SECRET_KEY, csrf_exempt, mark_safe, raw SQL, mass assignment, pickle |
| Express.js | 0 (detected but no rules!) | Full gap: helmet, NoSQL injection, SSRF, session config, eval, child_process, MongoDB injection |
| LLM/AI Security | 4 (AI residue only) | Missing: OWASP LLM Top 10 patterns — prompt injection markers, insecure output handling, excessive agency, system prompt leakage |
| Docker/IaC | 0 (engine has deploy audit but no YAML rules) | Missing: Dockerfile USER, privileged, host network, hardcoded secrets in compose |
| Python (universal) | 0 | Missing: subprocess shell=True, yaml.load, pickle, assert-as-guard, insecure tempfile, insecure random |

## Research Sources

- OWASP Top 10:2025 (Web) — A01-A10
- OWASP Top 10 for LLM Applications 2025 — LLM01-LLM10
- Semgrep Registry (p/default, p/django, p/expressjs, p/golang)
- Gosec rules (G1xx-G7xx)
- Bandit rules (B1xx-B6xx)
- Brakeman patterns (Rails)
- Real CVE patterns (Log4Shell, prototype pollution, SSRF bypasses)

---

## PHASE 1: Universal Rules (cross-stack, highest ROI)

### 1.1 Python Universal Security (NEW: `rules/python/security.yaml`)
Rules that apply to ALL Python stacks (FastAPI, Flask, Django).

| ID | Name | Pattern | Severity |
|---|---|---|---|
| PY-SUBPROCESS | Unsafe Subprocess | `subprocess.*shell\s*=\s*True` | CRITICAL |
| PY-YAML-LOAD | Unsafe YAML Load | `yaml\.load\(` without SafeLoader | CRITICAL |
| PY-PICKLE | Insecure Deserialization | `pickle\.load\|pickle\.loads\|marshal\.load` | CRITICAL |
| PY-ASSERT-GUARD | Assert as Security Guard | `assert\s+.*\b(user\|admin\|auth\|perm\|role)` | HIGH |
| PY-TEMPFILE | Insecure Temp File | `tempfile\.mktemp\(` | MEDIUM |
| PY-RANDOM | Insecure Randomness | `random\.(random\|randint\|choice)` in security context | HIGH |
| PY-EXEC | Dynamic Code Execution | `exec\(\|eval\(` (outside of known-safe wrappers) | CRITICAL |
| PY-LOGGING-SENSITIVE | Logging Sensitive Data | `logging\.\w+\(.*password\|token\|secret\|key` | HIGH |
| PY-TARFILE | Zip Slip / Path Traversal | `tarfile\.open\|zipfile\.extractall` without path validation | HIGH |
| PY-HARDCODED-BIND | Binding All Interfaces | `\.run\(.*host\s*=\s*['"]0\.0\.0\.0` | MEDIUM |
| PY-HASHLIB-WEAK | Weak Hash Algorithm | `hashlib\.(md5\|sha1)\(` | HIGH |
| PY-REQUESTS-NOVERIFY | Disabled SSL Verification | `verify\s*=\s*False` | HIGH |

### 1.2 Universal Extended (additions to `rules/universal/entropy.yaml`)

| ID | Name | Pattern | Severity |
|---|---|---|---|
| ENTR-SSRF | SSRF Risk | User input in URL fetch calls | HIGH |
| ENTR-NOSQL-INJECTION | NoSQL Injection | `\$where\|\$ne\|\$gt` from user input | CRITICAL |
| ENTR-INSECURE-RANDOM | Math.random for Security | `Math\.random\(\)` in token/key/session context | HIGH |
| ENTR-EVAL-FUNCTION | Dynamic Code via Function() | `new\s+Function\(` | CRITICAL |
| ENTR-INNERHTML | Direct innerHTML Assignment | `\.innerHTML\s*=` | HIGH |
| ENTR-PRIVATE-KEY | Private Key in Source | `-----BEGIN.*PRIVATE KEY-----` | CRITICAL |
| ENTR-TODO-FIXME | Unfinished Security Logic | `TODO.*secur\|FIXME.*auth\|HACK.*bypass` | MEDIUM |
| ENTR-CORS-WILDCARD | CORS Wildcard Origin | `Access-Control-Allow-Origin.*\*` | HIGH |
| ENTR-SENSITIVE-LOG | Sensitive Data in Logs | `console\.log\|fmt\.Print.*password\|token\|secret\|api.key` | HIGH |
| ENTR-JWT-NONE-ALG | JWT None Algorithm | `algorithm.*none\|alg.*none` | CRITICAL |

---

## PHASE 2: Django Rules (NEW STACK: `rules/django/security.yaml`)

Django is DETECTED by the engine but has ZERO rules. This is the biggest gap.

| ID | Name | Pattern | Severity |
|---|---|---|---|
| DJANGO-DEBUG | Debug Mode Enabled | `DEBUG\s*=\s*True` in settings | CRITICAL |
| DJANGO-HOSTS | Empty ALLOWED_HOSTS | `ALLOWED_HOSTS\s*=\s*\[\s*\]` | CRITICAL |
| DJANGO-SECRET | Hardcoded SECRET_KEY | `SECRET_KEY\s*=\s*['"]` (not os.environ) | CRITICAL |
| DJANGO-CSRF-EXEMPT | CSRF Exemption | `@csrf_exempt` | HIGH |
| DJANGO-MARK-SAFE | XSS via mark_safe | `mark_safe\(` | HIGH |
| DJANGO-RAW-SQL | Raw SQL Queries | `\.raw\(\|\.extra\(\|cursor\.execute\(` with string interpolation | CRITICAL |
| DJANGO-MASS-ASSIGN | Mass Assignment | `fields\s*=\s*['"]__all__['"]` | HIGH |
| DJANGO-MIDDLEWARE | Missing Security Middleware | Checks for SecurityMiddleware in settings | HIGH |
| DJANGO-CLICKJACK | Missing Clickjack Protection | Missing XFrameOptionsMiddleware | MEDIUM |
| DJANGO-SESSION | Insecure Session Config | `SESSION_COOKIE_SECURE\s*=\s*False` | HIGH |
| DJANGO-AUTH | Missing Authentication | Missing login_required/permission_required patterns | HIGH |
| DJANGO-LOGGING | Missing Security Logging | required_pattern for logging config | MEDIUM |

---

## PHASE 3: Express.js Rules (NEW STACK: `rules/express/security.yaml`)

Express is DETECTED by the engine but has ZERO rules.

| ID | Name | Pattern | Severity |
|---|---|---|---|
| EXPRESS-HELMET | Missing Helmet | required_pattern for helmet middleware | HIGH |
| EXPRESS-NOSQL | NoSQL Injection | `\.find\(\|\.findOne\(` with raw req.body | CRITICAL |
| EXPRESS-SESSION | Insecure Session | Missing secure/httpOnly/sameSite cookie flags | HIGH |
| EXPRESS-EVAL | Server-Side Eval | `eval\(\|new Function\(` | CRITICAL |
| EXPRESS-CHILD-PROC | Unsafe Child Process | `child_process\.\w+\(.*\$\{` template interpolation | CRITICAL |
| EXPRESS-SSRF | SSRF Risk | `axios\|fetch\|http\.get` with user-controlled URL | HIGH |
| EXPRESS-SQLI | SQL Injection | Raw query string concatenation | CRITICAL |
| EXPRESS-XSS | Direct Response XSS | `res\.send\(.*req\.\|res\.write\(.*req\.` | HIGH |
| EXPRESS-CORS | Missing CORS | required_pattern for cors middleware | MEDIUM |
| EXPRESS-RATELIMIT | Missing Rate Limiting | required_pattern for rate-limit middleware | HIGH |
| EXPRESS-AUTH | Missing Authentication | required_pattern for passport/jwt | HIGH |
| EXPRESS-BODYPARSER | Missing Body Parser Limit | `bodyParser\.json\(\)` without size limit | MEDIUM |

---

## PHASE 4: Go Extended (additions to `rules/golang/security.yaml`)

| ID | Name | Pattern | Severity |
|---|---|---|---|
| GO-SSRF | SSRF Risk | `http\.Get\|http\.Post` with variable URL from request | HIGH |
| GO-CMD-INJECTION | Command Injection | `exec\.Command.*\+\|fmt\.Sprintf.*exec` | CRITICAL |
| GO-MATH-RAND | Insecure Randomness | `math/rand` instead of `crypto/rand` | HIGH |
| GO-TLS-SKIP | Disabled TLS Verification | `InsecureSkipVerify:\s*true` | HIGH |
| GO-UNSAFE-PTR | Unsafe Pointer Usage | `unsafe\.Pointer` | MEDIUM |
| GO-ERR-BLANK | Ignored Error | `_\s*=\s*\w+\.\w+\(` (error in blank identifier) | HIGH |
| GO-HARDCODED-CREDS | Hardcoded Credentials | `password\|passwd\|secret.*=\s*"[^"]{8,}"` | HIGH |
| GO-DEFER-UNLOCK | Missing Defer Unlock | `\.Lock\(\)` without deferred `.Unlock()` | MEDIUM |
| GO-EMBED-SECRET | Embedded Secrets in Binary | `//go:embed.*\.env\|//go:embed.*secret` | CRITICAL |
| GO-WEAK-TLS | Weak TLS Configuration | `tls\.VersionTLS10\|tls\.VersionTLS11` | HIGH |

---

## PHASE 5: LLM/AI Security Rules (NEW: `rules/llm/security.yaml`)
Based on OWASP Top 10 for LLM Applications 2025.

| ID | Name | OWASP LLM | Pattern | Severity |
|---|---|---|---|---|
| LLM-PROMPT-INJECTION | Prompt Injection Markers | LLM01 | System prompt patterns with user input concatenation | CRITICAL |
| LLM-OUTPUT-EXEC | Insecure Output Handling | LLM05 | LLM output passed to eval/exec/sql/shell | CRITICAL |
| LLM-SYSTEM-PROMPT-LEAK | System Prompt in Client Code | LLM07 | System prompt strings in frontend/client code | HIGH |
| LLM-EXCESSIVE-AGENCY | Unchecked Tool Execution | LLM06 | LLM output directly used for file/db/api operations | HIGH |
| LLM-NO-GUARDRAILS | Missing Output Validation | LLM05 | LLM response used without schema validation | HIGH |
| LLM-SENSITIVE-IN-PROMPT | Sensitive Data in Prompts | LLM02 | Passwords/keys/PII in prompt templates | CRITICAL |
| LLM-NO-RATE-LIMIT | Missing LLM Rate Limiting | LLM10 | API calls without token/request limits | MEDIUM |
| LLM-HARDCODED-MODEL | Hardcoded Model Config | LLM03 | Model name/version hardcoded instead of configurable | LOW |

---

## PHASE 6: Docker & IaC Rules (NEW: `rules/docker/security.yaml`)

| ID | Name | Pattern | Severity |
|---|---|---|---|
| DOCKER-ROOT | Running as Root | Missing USER instruction in Dockerfile | HIGH |
| DOCKER-LATEST | Using :latest Tag | `FROM.*:latest` | MEDIUM |
| DOCKER-ADD | Using ADD Instead of COPY | `^ADD\s` (unless URL) | MEDIUM |
| DOCKER-SECRET-ENV | Secrets in ENV | `ENV.*PASSWORD\|ENV.*SECRET\|ENV.*API_KEY` | CRITICAL |
| DOCKER-PRIVILEGED | Privileged Container | `privileged:\s*true` in docker-compose | CRITICAL |
| DOCKER-HOST-NET | Host Network Mode | `network_mode:\s*host` | HIGH |
| DOCKER-NO-HEALTHCHECK | Missing Healthcheck | Missing HEALTHCHECK instruction | LOW |
| DOCKER-CURL-PIPE | Curl Pipe to Shell | `curl.*\|.*sh\|wget.*\|.*bash` | HIGH |

---

## PHASE 7: ASP.NET Core Extended

| ID | Name | Pattern | Severity |
|---|---|---|---|
| ASP-XXE | XML External Entity | `XmlReader\.Create\|XmlDocument` without DtdProcessing.Prohibit | CRITICAL |
| ASP-DESER | Insecure Deserialization | `BinaryFormatter\|JavaScriptSerializer\|TypeNameHandling\.All` | CRITICAL |
| ASP-PATH-TRAVERSAL | Path Traversal | `Path\.Combine.*Request\.\|File\.Open.*Request\.` | CRITICAL |
| ASP-CONN-STRING | Hardcoded Connection String | `Server=.*Password=` in source code | HIGH |
| ASP-LOGGING-SENSITIVE | Sensitive Data in Logs | `_logger\.Log.*password\|token\|secret` | HIGH |

---

## Execution Summary

| Phase | New Rules | New Files | Stack |
|---|---|---|---|
| Phase 1 | ~22 | `rules/python/security.yaml` + update `universal/entropy.yaml` | Universal + Python |
| Phase 2 | ~12 | `rules/django/security.yaml` | Django |
| Phase 3 | ~12 | `rules/express/security.yaml` | Express.js |
| Phase 4 | ~10 | Update `rules/golang/security.yaml` | Go |
| Phase 5 | ~8 | `rules/llm/security.yaml` | LLM/AI (OWASP LLM Top 10) |
| Phase 6 | ~8 | `rules/docker/security.yaml` | Docker/IaC |
| Phase 7 | ~5 | Update `rules/aspnetcore/security.yaml` | ASP.NET Core |
| **TOTAL** | **~77** | **4 new + 3 updated** | **+4 new stacks** |

Post-expansion: **69 + 77 = ~146 rules** across 10+ stacks.

## Files to Update

### Engine (default_rules.yaml)
All new rules MUST be added to `internal/engine/default_rules.yaml` — this is the source of truth.
The `rules/` directory will be updated as a mirror (documentation).

### Detector (detector.go)
No changes needed — Django and Express already detected.

### Verification
```bash
go build ./cmd/aitriage     # Binary compiles
go test ./...               # All tests pass
aitriage scan ./testdata    # Rules fire correctly
```
