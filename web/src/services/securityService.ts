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
  duration: string;
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

  generateThreatModel: async (context: string): Promise<ChatResponse> => {
    return securityService.chat([
      {
        role: 'system',
        content: `You are a senior security architect performing STRIDE threat modeling.
Analyze the target and produce a structured threat model with these sections:
## STRIDE Analysis
For each category (Spoofing, Tampering, Repudiation, Information Disclosure, Denial of Service, Elevation of Privilege):
- List specific threats with severity (CRITICAL/HIGH/MEDIUM/LOW)
- Affected components
- Recommended mitigations
## Attack Surface Summary
## Risk Matrix
Use markdown tables and formatting.`,
      },
      { role: 'user', content: `Perform STRIDE threat analysis for: ${context}` },
    ]);
  },

  generateAuditReport: async (context: string): Promise<ChatResponse> => {
    return securityService.chat([
      {
        role: 'system',
        content: `You are a senior security auditor generating a comprehensive security audit report.
Structure the report with:
## Executive Summary
## Scope & Methodology 
## Threat Model (STRIDE)
## Vulnerability Findings
For each finding: ID, Title, Severity, CVSS, Description, Impact, Remediation
## Dependency Analysis
## Compliance Mapping (OWASP Top 10, CWE)
## Remediation Roadmap (prioritized)
## Risk Score & Grade
Use markdown formatting with tables, severity badges, and clear structure.`,
      },
      { role: 'user', content: `Generate a full security audit report for: ${context}` },
    ]);
  },

  generateSecurityPlan: async (context: string): Promise<ChatResponse> => {
    return securityService.chat([
      {
        role: 'system',
        content: `You are a security implementation planner. Create a detailed, actionable security implementation plan.
Include:
## Priority Actions (Critical/High severity)
## Implementation Steps (numbered, with code examples)
## Timeline Estimate
## Testing & Verification Plan
## Rollback Strategy`,
      },
      { role: 'user', content: `Create a security implementation plan for: ${context}` },
    ]);
  },

  generatePoC: async (context: string): Promise<ChatResponse> => {
    return securityService.chat([
      {
        role: 'system',
        content: `You are a penetration tester. Generate proof-of-concept attack simulations to verify vulnerabilities.
For each PoC:
## Vulnerability Target
## Attack Vector
## PoC Code (safe, non-destructive)
## Expected Result
## Verification Steps
## Remediation Verification
IMPORTANT: All PoCs must be safe and non-destructive. Include clear warnings.`,
      },
      { role: 'user', content: `Generate PoC attack simulations for: ${context}` },
    ]);
  },
};
