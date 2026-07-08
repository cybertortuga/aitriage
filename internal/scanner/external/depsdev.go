package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const depsDevAPIURL = "https://api.deps.dev/v3alpha/findingsbatch"

var registrySystemMap = map[string]string{
	"npm":        "NPM",
	"pypi":       "PYPI",
	"gomodproxy": "GO",
	"rubygems":   "RUBYGEMS",
	"crates.io":  "CARGO",
	"maven":      "MAVEN",
	"nuget":      "NUGET",
}

// Structs matching the deps.dev API schema
type depsDevRequest struct {
	Requests []depsDevRequestItem `json:"requests"`
}

type depsDevRequestItem struct {
	VersionKey *depsDevVersionKey `json:"versionKey,omitempty"`
	PackageKey *depsDevPackageKey `json:"packageKey,omitempty"`
}

type depsDevVersionKey struct {
	System  string `json:"system"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

type depsDevPackageKey struct {
	System string `json:"system"`
	Name   string `json:"name"`
}

type depsDevResponse struct {
	Responses []depsDevResponseItem `json:"responses"`
}

type depsDevResponseItem struct {
	Request  depsDevRequestItem `json:"request"`
	Findings depsDevFindings    `json:"findings"`
}

type depsDevFindings struct {
	RecommendedVersions []struct {
		VersionKey depsDevVersionKey `json:"versionKey"`
	} `json:"recommendedVersions"`
	PackageFindings []depsDevFindingItem `json:"packageFindings"`
	RequestedVersion struct {
		Findings []depsDevFindingItem `json:"findings"`
	} `json:"requestedVersion"`
}

type depsDevFindingItem struct {
	Type            string `json:"type"`
	Risk            string `json:"risk"`
	LowUsageContext *struct {
		AlternativePackages string `json:"alternativePackages"`
	} `json:"lowUsageContext"`
}

// riskWeight maps deps.dev risk levels to integer weight
func riskWeight(risk string) int {
	switch risk {
	case "RISK_CRITICAL":
		return 4
	case "RISK_HIGH":
		return 3
	case "RISK_MEDIUM":
		return 2
	case "RISK_LOW":
		return 1
	default:
		return 0
	}
}

// QueryDepsDev calls the open-source deps.dev findings API for the package list.
func QueryDepsDev(ctx context.Context, registry string, packages []DepPackageRequest) ([]DepFinding, error) {
	system := registrySystemMap[registry]
	if system == "" {
		return nil, fmt.Errorf("unsupported registry: %s", registry)
	}

	reqPayload := depsDevRequest{
		Requests: make([]depsDevRequestItem, 0, len(packages)),
	}

	for _, p := range packages {
		if p.Package == "" {
			continue
		}
		if p.Version != "" {
			reqPayload.Requests = append(reqPayload.Requests, depsDevRequestItem{
				VersionKey: &depsDevVersionKey{
					System:  system,
					Name:    p.Package,
					Version: p.Version,
				},
			})
		} else {
			reqPayload.Requests = append(reqPayload.Requests, depsDevRequestItem{
				PackageKey: &depsDevPackageKey{
					System: system,
					Name:   p.Package,
				},
			})
		}
	}

	if len(reqPayload.Requests) == 0 {
		return []DepFinding{}, nil
	}

	body, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, depsDevAPIURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "SecureCoder-AITriage-Fallback")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("deps.dev request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("deps.dev API returned status %d", resp.StatusCode)
	}

	var resPayload depsDevResponse
	if err := json.NewDecoder(resp.Body).Decode(&resPayload); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var findings []DepFinding

	for _, item := range resPayload.Responses {
		// Suggested upgrade version
		var suggestedVersion string
		if len(item.Findings.RecommendedVersions) > 0 {
			suggestedVersion = item.Findings.RecommendedVersions[0].VersionKey.Version
		}

		// Find highest package finding
		var highestPkgFinding *depsDevFindingItem
		for _, f := range item.Findings.PackageFindings {
			if riskWeight(f.Risk) > 0 {
				if highestPkgFinding == nil || riskWeight(f.Risk) > riskWeight(highestPkgFinding.Risk) {
					ref := f // local copy
					highestPkgFinding = &ref
				}
			}
		}

		// Find highest requested version finding
		var highestVerFinding *depsDevFindingItem
		for _, f := range item.Findings.RequestedVersion.Findings {
			if riskWeight(f.Risk) > 0 {
				if highestVerFinding == nil || riskWeight(f.Risk) > riskWeight(highestVerFinding.Risk) {
					ref := f // local copy
					highestVerFinding = &ref
				}
			}
		}

		pkgName := ""
		pkgVersion := ""
		if item.Request.PackageKey != nil {
			pkgName = item.Request.PackageKey.Name
		} else if item.Request.VersionKey != nil {
			pkgName = item.Request.VersionKey.Name
			pkgVersion = item.Request.VersionKey.Version
		}

		if highestPkgFinding != nil && (highestVerFinding == nil || riskWeight(highestPkgFinding.Risk) >= riskWeight(highestVerFinding.Risk)) {
			reason := ""
			action := ""
			switch highestPkgFinding.Type {
			case "NOT_FOUND":
				reason = fmt.Sprintf("Package \"%s\" does not exist.", pkgName)
				action = "Do not use this package. Choose a different package to use."
			case "MALICIOUS":
				reason = fmt.Sprintf("Package \"%s\" is malicious.", pkgName)
				action = "Do not use this package. Choose a different package to use."
			case "VULNERABLE":
				reason = fmt.Sprintf("All versions of package \"%s\" are affected by a critical vulnerability.", pkgName)
				action = "Do not use this package. Choose a different package to use."
			case "LOW_USAGE":
				reason = fmt.Sprintf("Package \"%s\" has extremely low usage, which may indicate a typosquatting attempt, or other quality issues.", pkgName)
				action = "Do not use this package. Choose a different package to use."
			case "COOLDOWN":
				reason = fmt.Sprintf("All versions of package \"%s\" were recently published and have not passed a cooldown period. This is a potential security risk.", pkgName)
				action = "Confirm with the user before using this package."
			case "DEPRECATED":
				reason = fmt.Sprintf("Package \"%s\" is deprecated which often has security implications.", pkgName)
				action = "Confirm with the user before using this package."
			default:
				reason = fmt.Sprintf("A \"%s\" finding with type \"%s\" was found.", highestPkgFinding.Risk, highestPkgFinding.Type)
			}

			var altPackages string
			if highestPkgFinding.LowUsageContext != nil {
				altPackages = highestPkgFinding.LowUsageContext.AlternativePackages
			}

			findings = append(findings, DepFinding{
				Registry:            registry,
				Package:             pkgName,
				Version:             pkgVersion,
				Reason:              reason,
				Action:              action,
				AlternativePackages: altPackages,
			})
			continue
		}

		if highestVerFinding != nil {
			reason := ""
			action := ""
			switch highestVerFinding.Type {
			case "NOT_FOUND":
				reason = fmt.Sprintf("Package \"%s\" at version \"%s\" does not exist.", pkgName, pkgVersion)
				action = "Do not use this version. Choose a different version to use."
			case "MALICIOUS":
				reason = fmt.Sprintf("Package \"%s\" at version \"%s\" is malicious.", pkgName, pkgVersion)
				action = "Do not use this version. Choose a different version to use."
			case "VULNERABLE":
				reason = fmt.Sprintf("Package \"%s\" at version \"%s\" is affected by a critical vulnerability.", pkgName, pkgVersion)
				action = "Do not use this version. Choose a different version to use."
			case "COOLDOWN":
				reason = fmt.Sprintf("Package \"%s\" at version \"%s\" was recently published and have not passed a cooldown period. This is a potential security risk.", pkgName, pkgVersion)
				action = "Do not use this version. Choose a different version to use."
			case "DEPRECATED":
				reason = fmt.Sprintf("Package \"%s\" at version \"%s\" is deprecated which often has security implications.", pkgName, pkgVersion)
				action = "Do not use this version. Choose a different version to use."
			default:
				reason = fmt.Sprintf("A \"%s\" finding with type \"%s\" was found.", highestVerFinding.Risk, highestVerFinding.Type)
			}

			findings = append(findings, DepFinding{
				Registry:         registry,
				Package:          pkgName,
				Version:          pkgVersion,
				Reason:           reason,
				Action:           action,
				SuggestedVersion: suggestedVersion,
			})
		}
	}

	return findings, nil
}
