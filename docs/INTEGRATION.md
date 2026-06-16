# AITriage: CI/CD "Entropy-Shield" Integration 🛡️

Integrate AITriage into your development workflow to automatically block PRs that contain AI-generated hallucinations or insecure defaults.

## 🐙 GitHub Actions

Add this file to `.github/workflows/aitriage.yml`:

```yaml
name: AITriage Security Scan
on: [push, pull_request]

jobs:
  audit:
    name: AI Risk Audit
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'
          
      - name: Install AITriage
        run: go install github.com/cybertortuga/aitriage/cmd/aitriage@latest
        
      - name: Run Scan & Fail on Grade F
        run: |
          # Run scan and capture score
          SCORE=$(aitriage scan . --format json | jq '.Score')
          echo "AITriage Security Score: $SCORE"
          if [ "$SCORE" -lt 50 ]; then
            echo "FATAL: Codebase grade is F. High suspicion of AI-induced vulnerabilities."
            exit 1
          fi
```

## 🦊 GitLab CI/CD

Add this to `.gitlab-ci.yml`:

```yaml
aitriage_scan:
  image: golang:1.22
  script:
    - go install github.com/cybertortuga/aitriage/cmd/aitriage@latest
    - aitriage scan . --format json > report.json
    - score=$(cat report.json | jq '.Score')
    - if [ "$score" -lt 50 ]; then exit 1; fi
  artifacts:
    reports:
      metrics: report.json
```

## 🔐 Best Practices
1. **Set Thresholds**: Start with a low threshold (e.g. 50) and gradually increase it as you clean up your AI-induced debt.
2. **Force Stack**: Use `--stack nextjs` if the auto-detection is not sufficient for your project.
3. **Scan early**: Run AITriage on every commit to catch "Entropy-Coding" as it happens.
