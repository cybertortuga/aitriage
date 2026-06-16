## 2025-02-20 - [Pre-compile regexp]
**Learning:** Compiling regex dynamically within a function call is slow, caching it globally saves processing time and is safe since regexp is thread-safe in Go.
**Action:** Always verify if regexp.MustCompile is called inside a hot loop or frequently executed function, and hoist it to the package level.
