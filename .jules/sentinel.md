## 2025-05-14 - Prevent Command Line Secret Leakage in Docker Escalate
**Vulnerability:** API keys (e.g. GEMINI_API_KEY) were passed to `docker run` using `-e KEY=VALUE` arguments.
**Learning:** Passing `-e KEY=VALUE` exposes the plaintext value in the host machine's process list (`ps aux`). Any user on the host can see these process arguments.
**Prevention:** Always use `-e KEY` without the value when the variable already exists in the environment. Docker will automatically pass the value securely into the container without exposing it in the process arguments.
