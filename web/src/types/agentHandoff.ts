export interface AIAgentFinding {
  id: string;
  severity: string;
  file?: string;
  line?: number;
  title: string;
  disposition: string;
  recommendation?: string;
}

export interface AIAgentData {
  scan_date: string;
  score: number;
  grade: string;
  gate_status: string;
  policy: {
    profile: string;
    fail_on: string;
  };
  stats: {
    true_positives: number;
    needs_review: number;
    false_positives: number;
    total: number;
  };
  findings: AIAgentFinding[];
}

export interface AgentHandoffResponse {
  ok: true;
  schema_version: number;
  session_id: number;
  generated_at: string;
  session_status: string;
  gate_status: string;
  remediation_prompt_markdown: string;
  remediation_prompt_sha256: string;
  remediation_prompt_size_bytes: number;
  agent_data: AIAgentData;
  agent_data_sha256: string;
  agent_data_size_bytes: number;
}

export interface RunwayArtifactMetadata {
  kind: string;
  media_type: string;
  schema_version: number;
  sha256: string;
  size_bytes: number;
  created_at: string;
}
