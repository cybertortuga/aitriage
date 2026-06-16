import { useState, useEffect, useCallback } from 'react';
import api from '../services/api';
import i18n from '../i18n';

export interface ReportHistoryItem {
  id: number;
  timestamp: string;
  target_scope: string;
  format: string;
  status: string;
  download_url: string;
}

export interface ExecutiveSummary {
  total_findings: number;
  by_severity: Record<string, number>;
  by_status: Record<string, number>;
}

export const useReports = () => {
  const [executiveSummary, setExecutiveSummary] = useState<ExecutiveSummary | null>(null);
  const [reportHistory, setReportHistory] = useState<ReportHistoryItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [generating, setGenerating] = useState(false);
  const [generateError, setGenerateError] = useState<string | null>(null);
  const [generateSuccess, setGenerateSuccess] = useState(false);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const [summaryRes, historyRes] = await Promise.all([
        api.get('/reports/executive'),
        api.get('/reports/history'),
      ]);
      setExecutiveSummary(summaryRes.data);
      setReportHistory(historyRes.data.reports || []);
    } catch (err) {
      console.error(i18n.t('errors.fetchReports'), err);
    } finally {
      setLoading(false);
    }
  }, []);

  const downloadCSV = useCallback(() => {
    const baseURL = api.defaults.baseURL;
    window.open(`${baseURL}/reports/executive?format=csv`, '_blank');
  }, []);

  const generateReport = useCallback(
    async (format: string, options: { includeDeps: boolean; sign: boolean }) => {
      setGenerating(true);
      setGenerateError(null);
      setGenerateSuccess(false);
      try {
        const { data } = await api.post('/reports/generate', {
          format,
          scope: 'all-findings',
          options: {
            include_deps: options.includeDeps,
            sign: options.sign,
          },
        });
        if (data.ok) {
          setGenerateSuccess(true);
          // Refresh history to show new report
          await fetchData();
          // Auto-clear success after 4s
          setTimeout(() => setGenerateSuccess(false), 4000);
        } else {
          setGenerateError(data.error || 'Generation failed');
        }
      } catch (err) {
        setGenerateError(err instanceof Error ? err.message : 'Generation failed');
      } finally {
        setGenerating(false);
      }
    },
    [fetchData],
  );

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  return {
    executiveSummary,
    reportHistory,
    loading,
    generating,
    generateError,
    generateSuccess,
    downloadCSV,
    generateReport,
    refresh: fetchData,
  };
};
