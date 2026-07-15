import { useCallback, useEffect, useState } from 'react';
import { fetchAgentHandoff } from '../services/runwayArtifacts';
import type { AgentHandoffResponse } from '../types/agentHandoff';

export const useAgentHandoff = (sessionId: number | null, enabled = true) => {
  const [result, setResult] = useState<{
    key: string;
    data: AgentHandoffResponse | null;
    error: string | null;
  }>({ key: '', data: null, error: null });
  const [reloadKey, setReloadKey] = useState(0);

  const reload = useCallback(() => setReloadKey((value) => value + 1), []);
  const requestKey = sessionId && enabled ? `${sessionId}:${reloadKey}` : '';

  useEffect(() => {
    if (!sessionId || !requestKey) return;

    const controller = new AbortController();
    fetchAgentHandoff(sessionId, controller.signal)
      .then((data) => setResult({ key: requestKey, data, error: null }))
      .catch((cause: unknown) => {
        if (controller.signal.aborted) return;
        setResult({
          key: requestKey,
          data: null,
          error: cause instanceof Error ? cause.message : 'Failed to load AI remediation handoff.',
        });
      });

    return () => controller.abort();
  }, [requestKey, sessionId]);

  const isCurrent = result.key === requestKey;
  return {
    data: isCurrent ? result.data : null,
    loading: Boolean(requestKey) && !isCurrent,
    error: isCurrent ? result.error : null,
    reload,
  };
};
