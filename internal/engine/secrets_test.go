package engine

import (
	"strings"
	"testing"
)

// hitIDs returns the set of detected signature IDs for src.
func hitIDs(src string) map[string]bool {
	ids := map[string]bool{}
	for _, h := range detectSecretHits(src, true) {
		ids[h.id] = true
	}
	return ids
}

func TestDetectKnownSecretFormats(t *testing.T) {
	cases := []struct {
		name   string
		src    string
		wantID string
	}{
		{"aws-akid", `const k = "AKIAIOSFODNN7EXAMPLE"`, "SECRET-AWS-AKID"},
		{"github-token", `token := "ghp_abcdefghijklmnopqrstuvwxyz0123456789"`, "SECRET-GH-TOKEN"},
		{"github-pat", `x = "github_pat_11ABCDEFG0aBcDeFgHiJ_kLmNoPqRsTuVwXyZ0123456789aBcDeFgHiJkLmNoPqRsTuVwXyZ012345"`, "SECRET-GH-PAT"},
		{"slack-token", `t = "xoxb-` + `123456789012-abcdefghijklmnop"`, "SECRET-SLACK-TOKEN"},
		{"slack-webhook", `url = "https://hooks.slack.com/services/` + `T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"`, "SECRET-SLACK-WEBHOOK"},
		{"stripe", `k = "sk_live_` + `abcdefghijklmnopqrstuvwx"`, "SECRET-STRIPE"},
		{"openai", `k = "sk-abcdefghijklmnopqrstuvwxyz1234"`, "SECRET-OPENAI"},
		{"anthropic", `k = "sk-ant-abcdefghijklmnopqrstuvwxyz1234"`, "SECRET-ANTHROPIC"},
		{"google", `k = "AIzaSyA1234567890abcdefghijklmnopqrstuvw"`, "SECRET-GOOGLE-API"},
		{"sendgrid", `k = "SG.abcdefghijklmnopqrstuv.abcdefghijklmnopqrstuvwxyz0123456789ABCDEF"`, "SECRET-SENDGRID"},
		{"npm", `k = "npm_abcdefghijklmnopqrstuvwxyz0123456789"`, "SECRET-NPM"},
		{"jwt", `t = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N"`, "SECRET-JWT"},
		{"private-key", "key := `-----BEGIN RSA PRIVATE KEY-----\nMIIabc\n-----END RSA PRIVATE KEY-----`", "SECRET-PRIVATE-KEY"},
		{"conn-postgres", `dsn = "postgres://admin:S3cretDbPass@db.internal:5432/app"`, "SECRET-CONN-STRING"},
		{"conn-mongo-srv", `uri = "mongodb+srv://user:p4ssw0rdXYZ@cluster0.mongodb.net/test"`, "SECRET-CONN-STRING"},
		{"conn-mysql", `dsn = "mysql://root:hunter2pass@127.0.0.1:3306/db"`, "SECRET-CONN-STRING"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ids := hitIDs(c.src)
			if !ids[c.wantID] {
				t.Errorf("src %q: expected %s, got IDs %v", c.src, c.wantID, ids)
			}
		})
	}
}

func TestDetectGenericAssignments(t *testing.T) {
	cases := map[string]string{
		"api key blob": `apiKey = "a1B2c3D4e5F6g7H8i9J0k1L2m3N4o5P6"`,
		"strong pw":    `password = "Kx9$mQ2vLp8@nR4wZt"`,
		"secret token": `client_secret: "z9Y8x7W6v5U4t3S2r1Q0p9O8n7M6l5K4"`,
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			hits := detectSecretHits(src, true)
			found := false
			for _, h := range hits {
				if h.id == "SECRET-ASSIGN" {
					found = true
				}
			}
			if !found {
				t.Errorf("expected SECRET-ASSIGN for %q, got %+v", src, hits)
			}
		})
	}
}

func TestSecretLeakEvidenceIsRedacted(t *testing.T) {
	src := `k = "AKIAIOSFODNN7EXAMPLE"`
	hits := detectSecretHits(src, true)
	if len(hits) == 0 {
		t.Fatal("expected a hit")
	}
	if strings.Contains(hits[0].evidence, "AKIAIOSFODNN7EXAMPLE") {
		t.Errorf("evidence must not contain the full secret: %q", hits[0].evidence)
	}
	if !strings.Contains(hits[0].evidence, "redacted") {
		t.Errorf("evidence should be marked redacted: %q", hits[0].evidence)
	}
}

func TestMultipleSecretsInOneFile(t *testing.T) {
	src := strings.Join([]string{
		`const aws = "AKIAIOSFODNN7EXAMPLE"`,
		`const gh  = "ghp_abcdefghijklmnopqrstuvwxyz0123456789"`,
		`const db  = "postgres://admin:S3cretDbPass@db:5432/app"`,
		`const oai = "sk-abcdefghijklmnopqrstuvwxyz1234"`,
		`apiKey    = "a1B2c3D4e5F6g7H8i9J0k1L2m3N4o5P6"`,
	}, "\n")

	hits := detectSecretHits(src, true)
	if len(hits) < 5 {
		t.Fatalf("expected all 5 secrets, got %d: %+v", len(hits), hits)
	}
	// Each must be on its own line (1..5) — proves we collect all, not just the first.
	lines := map[int]bool{}
	for _, h := range hits {
		lines[h.line] = true
	}
	for ln := 1; ln <= 5; ln++ {
		if !lines[ln] {
			t.Errorf("no secret detected on line %d; lines hit: %v", ln, lines)
		}
	}
}

func TestNoFalsePositives(t *testing.T) {
	negatives := []string{
		`apiKey = process.env.API_KEY`,
		`token = "${GITHUB_TOKEN}"`,
		`password = "example"`,
		`password = "changeme"`,
		`secret = "your_secret_here"`,
		`url = "https://api.example.com/v1/users"`,
		`name = "getUserProfileByIdAndTenant"`,
		`message = "the quick brown fox jumps over"`,
		`dbHost = "redis://cache.internal:6379/0"`, // host:port, no credentials
		`const greeting = "Hello, world! Welcome."`,
	}
	for _, src := range negatives {
		t.Run(src, func(t *testing.T) {
			if hits := detectSecretHits(src, true); len(hits) != 0 {
				t.Errorf("false positive on %q: %+v", src, hits)
			}
		})
	}
}
