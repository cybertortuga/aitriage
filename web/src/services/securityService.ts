import api from './api';

export interface ScannerStatus {
  ok: boolean;
  tools: Record<string, boolean>;
}

export interface Rule {
  id: string;
  name: string;
  severity: string;
  description: string;
  owasp?: string;
  stack?: string;
}

export interface Finding {
  id: string;
  name: string;
  severity: string;
  file: string;
  line: number;
  suggestion: string;
  owasp?: string;
  audit_status: string;
}

export interface ScanResponse {
  ok: boolean;
  scan_id: string;
  findings: Finding[];
  security_score: number;
  security_grade: string;
  health_check?: HealthCheckResult;
  duration: string;
}

export interface HealthCheckResult {
  score: number;
  grade: string;
  has_critical_failures: boolean;
  breakdown: {
    base_score: number;
    penalty: number;
    bonus: number;
    raw_weight: number;
    active_findings: number;
    ignored_findings: number;
    deduped_findings: number;
    penalty_by_source: Record<string, number>;
    count_by_severity: Record<string, number>;
    count_by_source?: Record<string, number>;
    count_by_class?: Record<string, number>;
  };
  policy?: {
    profile: string;
    fail_on: string;
    minimum_score: number;
    max_critical: number;
    max_high: number;
    max_medium: number;
    block_sources?: string[];
    block_classes?: string[];
  };
  verdict?: {
    passed: boolean;
    status: string;
    summary: string;
    blocking_reasons?: Array<{
      code: string;
      message: string;
      severity?: string;
      source?: string;
      class?: string;
      count?: number;
      threshold?: number;
    }>;
  };
}

export interface ChatResponse {
  ok: boolean;
  content: string;
  error?: string;
}

export interface AnalysisResponse {
  ok: boolean;
  analysis: string;
  error?: string;
}

export const securityService = {
  getHealth: async (): Promise<ScannerStatus> => {
    const { data } = await api.get('/health');
    return data;
  },

  getRules: async (): Promise<Rule[]> => {
    const { data } = await api.get('/rules');
    return data.rules || [];
  },

  getFindings: async (): Promise<Finding[]> => {
    const { data } = await api.get('/findings');
    return data.findings || [];
  },

  startScan: async (path: string, stack?: string): Promise<ScanResponse> => {
    const { data } = await api.post('/scan', { path, stack });
    return data;
  },

  chat: async (messages: { role: string; content: string }[]): Promise<ChatResponse> => {
    const { data } = await api.post('/chat', { messages });
    return data;
  },

  analyzeFinding: async (id: string): Promise<AnalysisResponse> => {
    const { data } = await api.post('/analyze', { id, type: 'finding' });
    return data;
  },

  getFileContent: async (path: string): Promise<{ ok: boolean; content: string }> => {
    const { data } = await api.get('/file', { params: { path } });
    return data;
  },

  triageFinding: async (
    id: string,
    file: string,
    action: 'FIX' | 'IGNORE',
    project: string,
  ): Promise<{ ok: boolean }> => {
    const { data } = await api.post('/triage', { id, file, action, project });
    return data;
  },

  getMetrics: async (): Promise<any> => {
    const { data } = await api.get('/metrics');
    return data.metrics;
  },

  getAISummary: async (productId?: number | null): Promise<string> => {
    const url = productId ? `/ai-summary?product_id=${productId}` : '/ai-summary';
    const { data } = await api.get(url);
    return data.summary || '';
  },

  /* ── SecureCoder Integration ─────────────────────────────────────── */
  // These methods use the unified SecureCoder framework from the server.
  // The system prompt is injected server-side via handleChat's SecureCoderFramework.
  // We just need to send the right user-level instruction — the server adds
  // the full MUST/MUST NOT ruleset and threat model methodology automatically.

  generateThreatModel: async (context: string): Promise<ChatResponse> => {
    return securityService.chat([
      {
        role: 'user',
        content: `Perform a comprehensive STRIDE threat model analysis for the following target.

## Target
${context}

## Required Sections
1. **STRIDE Analysis** — For each category (Spoofing, Tampering, Repudiation, Information Disclosure, Denial of Service, Elevation of Privilege): list specific threats with severity (CRITICAL/HIGH/MEDIUM/LOW), affected components, and recommended mitigations grounded in the SecureCoder Evaluation Ruleset.
2. **Entry Points & Trust Boundaries** — All untrusted input sources and trust boundary crossings.
3. **Attack Surface Summary** — Prioritized list of areas requiring immediate attention.
4. **Risk Matrix** — Use markdown tables.

Apply the MUST/MUST NOT rules from your Evaluation Ruleset to each mitigation recommendation.`,
      },
    ]);
  },

  generateAuditReport: async (context: string): Promise<ChatResponse> => {
    return securityService.chat([
      {
        role: 'user',
        content: `Generate a comprehensive security audit report for the following target.

## Target
${context}

## Required Sections
1. **Executive Summary** — Overall security posture, score, grade.
2. **Threat Model (STRIDE)** — Entry points, trust boundaries, sensitive data paths.
3. **Vulnerability Findings** — Table with columns: Vulnerability ID (CS-XXX-NNN), Severity, File, Line, Triage Status, Recommendation, Rationale.
4. **PoC Verification** — For each True Positive, step-by-step exploit reasoning.
5. **Dependency Analysis** — Known CVEs in dependencies.
6. **Compliance Mapping** — OWASP Top 10, CWE references.
7. **Remediation Roadmap** — Prioritized fix plan using the SecureCoder Evaluation Ruleset.

Use the CS-XXX-NNN vulnerability ID format. Ground all recommendations in the MUST/MUST NOT rules.`,
      },
    ]);
  },

  generateSecurityPlan: async (context: string): Promise<ChatResponse> => {
    return securityService.chat([
      {
        role: 'user',
        content: `Create a detailed, actionable security implementation plan for the following target.

## Target
${context}

## Required Sections
1. **Fix Plan Summary Table** — | # | Priority | File | Issue | Vuln IDs |
2. **Tasks** — For each task: Priority, Vuln IDs, File, Line, Function, Problem description (trace data flow from untrusted input to sink), Security rules violated (MUST/MUST NOT from the Evaluation Ruleset), Context.
3. **Execution Order** — Fix critical vulnerabilities first (RCE, SSTI, debug mode), then auth/authz gaps, then input validation, then hardening.
4. **Verification Plan** — Tests and commands to verify each fix.

CRITICAL: Describe problems, not solutions. The AI IDE will write the code. Every finding must have a concrete file path.`,
      },
    ]);
  },

  generatePoC: async (context: string): Promise<ChatResponse> => {
    return securityService.chat([
      {
        role: 'user',
        content: `Generate proof-of-concept verification for vulnerabilities in the following target.

## Target
${context}

## For Each Vulnerability
1. **Vulnerability Type & Severity**
2. **PoC Description** — What input/request would an attacker craft? Must be SAFE and NON-DESTRUCTIVE.
3. **Data Flow Trace** — Follow the exploit input through the code step by step.
4. **Interception Point** — Where would a fix intercept or neutralize the exploit?
5. **Conclusion** — Exploitable / Not Exploitable / Needs Manual Review.
6. **Verification Steps** — Commands to verify the vulnerability EXISTS (before fix) and is REMEDIATED (after fix).
7. **Regression Test** — Test that passes when secure, fails when vulnerable.

Use vulnerability-specific reasoning from the Evaluation Ruleset (Path Traversal, XSS, SQLi, SSRF, Command Injection, Hardcoded Secrets, JWT Bypass, SSTI).
ALL PoCs MUST BE SAFE AND NON-DESTRUCTIVE.`,
      },
    ]);
  },
};
