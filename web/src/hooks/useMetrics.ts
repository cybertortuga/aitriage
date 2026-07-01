import { useState, useEffect, useCallback } from 'react';
import api from '../services/api';
import i18n from '../i18n';

export interface TopFile {
  path: string;
  count: number;
}

export interface StatusBreakdown {
  status: string;
  count: number;
}

export interface StackBreakdown {
  stack: string;
  count: number;
}

export interface DashboardMetrics {
  total_products: number;
  active_engagements: number;
  open_findings: number;
  sla_breached: number;
  severity_counts: Record<string, number>;
  top_risky_products: Array<{ name: string; risk_score: number; trend: string }>;
  recent_engagements: Array<{ name: string; status: string; date: string }>;
  mttr: Record<string, string>;

  // Extended
  total_findings: number;
  resolved_findings: number;
  top_files: TopFile[];
  status_breakdown: StatusBreakdown[];
  stack_breakdown: StackBreakdown[];
  security_score: number;
  security_grade: string;
  total_engagements: number;
}

type RefreshOptions = { silent?: boolean };

export const useMetrics = () => {
  const [metrics, setMetrics] = useState<DashboardMetrics | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchMetrics = useCallback(async (options?: RefreshOptions) => {
    if (!options?.silent) setLoading(true);
    try {
      const { data } = await api.get('/metrics');
      if (data.ok) {
        setMetrics(data.metrics);
        setError(null);
      } else {
        setError(data.error || i18n.t('errors.fetchMetrics'));
      }
    } catch (err: any) {
      setError(err.response?.data?.error || err.message);
    } finally {
      if (!options?.silent) setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchMetrics();
  }, [fetchMetrics]);

  return { metrics, loading, error, refresh: fetchMetrics };
};
