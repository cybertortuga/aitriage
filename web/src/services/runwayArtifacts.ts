import type { AgentHandoffResponse } from '../types/agentHandoff';

const parseAPIError = async (response: Response, fallback: string) => {
  try {
    const payload = await response.json();
    return typeof payload?.error === 'string' ? payload.error : fallback;
  } catch {
    return fallback;
  }
};

export const runwayArtifactURL = (sessionId: number, kind: string) =>
  `/api/runway/artifacts/${sessionId}/${encodeURIComponent(kind)}`;

export const downloadRunwayArtifact = async (
  sessionId: number,
  kind: string,
  filename: string,
): Promise<boolean> => {
  const response = await fetch(runwayArtifactURL(sessionId, kind));
  if (!response.ok) return false;

  const blob = await response.blob();
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement('a');
  anchor.href = url;
  anchor.download = filename;
  document.body.appendChild(anchor);
  anchor.click();
  anchor.remove();
  URL.revokeObjectURL(url);
  return true;
};

export const fetchAgentHandoff = async (
  sessionId: number,
  signal?: AbortSignal,
): Promise<AgentHandoffResponse> => {
  const response = await fetch(`/api/runway/handoff/${sessionId}`, {
    signal,
    headers: { Accept: 'application/json' },
  });
  if (!response.ok) {
    throw new Error(await parseAPIError(response, 'Failed to load AI remediation handoff.'));
  }
  return response.json() as Promise<AgentHandoffResponse>;
};
