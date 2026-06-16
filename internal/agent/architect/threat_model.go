package architect

// ThreatModel represents a security analysis of an architecture.
type ThreatModel struct {
	RiskScore       int            `json:"risk_score"` // 0-100 (100 = most risky)
	AttackSurface   []AttackVector `json:"attack_surface"`
	Recommendations []string       `json:"recommendations"`
}

// AttackVector represents a specific threat to a component.
type AttackVector struct {
	Target     string `json:"target"`
	ThreatType string `json:"threat_type"`
	Severity   string `json:"severity"` // HIGH, MEDIUM, LOW
	Mitigation string `json:"mitigation"`
}

// GenerateThreatModel analyzes detected components and creates a threat model.
func GenerateThreatModel(components []Component, projectPath string) ThreatModel {
	var tm ThreatModel
	var score int
	hasDB := false
	hasProxy := false
	hasDocker := fileExists(projectPath + "/Dockerfile")
	hasCompose := fileExists(projectPath+"/docker-compose.yml") || fileExists(projectPath+"/docker-compose.yaml")
	hasEnv := fileExists(projectPath + "/.env")

	for _, c := range components {
		switch c.Type {
		case "db":
			hasDB = true
			if c.Name == "PostgreSQL" || c.Name == "MySQL" || c.Name == "MongoDB" {
				tm.AttackSurface = append(tm.AttackSurface, AttackVector{
					Target:     c.Name,
					ThreatType: "SQL/NoSQL Injection",
					Severity:   "HIGH",
					Mitigation: "Use parameterized queries/ORMs exclusively.",
				})
				tm.AttackSurface = append(tm.AttackSurface, AttackVector{
					Target:     c.Name,
					ThreatType: "Unauthenticated Access",
					Severity:   "HIGH",
					Mitigation: "Ensure DB binds to localhost/internal network and requires strong passwords.",
				})
				score += 20
			}
		case "cache":
			if c.Name == "Redis" {
				tm.AttackSurface = append(tm.AttackSurface, AttackVector{
					Target:     c.Name,
					ThreatType: "Unauthenticated Cache Access",
					Severity:   "MEDIUM",
					Mitigation: "Enable Redis AUTH and bind to internal IP only.",
				})
				score += 10
			}
		case "proxy":
			hasProxy = true
			if c.Name == "Nginx" {
				tm.AttackSurface = append(tm.AttackSurface, AttackVector{
					Target:     c.Name,
					ThreatType: "Misconfigured Security Headers",
					Severity:   "MEDIUM",
					Mitigation: "Add HSTS, CSP, X-Frame-Options, and X-Content-Type-Options headers.",
				})
			}
		case "storage":
			if c.Name == "MinIO" || c.Name == "S3" {
				tm.AttackSurface = append(tm.AttackSurface, AttackVector{
					Target:     c.Name,
					ThreatType: "Public Bucket Exposure",
					Severity:   "HIGH",
					Mitigation: "Disable public access policies, use presigned URLs for client access.",
				})
				score += 15
			}
		case "message_broker":
			tm.AttackSurface = append(tm.AttackSurface, AttackVector{
				Target:     c.Name,
				ThreatType: "Message Interception",
				Severity:   "MEDIUM",
				Mitigation: "Enable TLS for broker connections and authenticate publishers/subscribers.",
			})
			score += 10
		}
	}

	// Architectural anti-patterns
	if !hasProxy && hasDB {
		tm.AttackSurface = append(tm.AttackSurface, AttackVector{
			Target:     "Application",
			ThreatType: "Direct Database Exposure",
			Severity:   "MEDIUM",
			Mitigation: "Implement an API Gateway or Reverse Proxy to filter incoming traffic.",
		})
		score += 10
	}

	if hasEnv && !hasDocker && !hasCompose {
		tm.Recommendations = append(tm.Recommendations, "Environment variables loaded from .env without containerization. Ensure .env is explicitly excluded from version control.")
	}

	if hasCompose && !hasProxy {
		tm.Recommendations = append(tm.Recommendations, "Docker Compose deployment lacks a reverse proxy. Consider adding Nginx/Traefik for TLS termination and routing.")
	}

	tm.RiskScore = score
	if tm.RiskScore > 100 {
		tm.RiskScore = 100
	}

	return tm
}
