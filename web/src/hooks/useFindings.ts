import { useState, useEffect, useCallback } from 'react';
import type { Finding } from '../types';
import api from '../services/api';
import i18n from '../i18n';

export const useFindings = (productId?: number) => {
  const [findings, setFindings] = useState<Finding[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchFindings = useCallback(async () => {
    try {
      setLoading(true);
      const url = productId ? `/findings?product_id=${productId}` : '/findings';
      const { data } = await api.get(url);
      const raw: Finding[] = data?.findings || data || [];
      // Normalize: ensure every finding has a valid status & kanban_column
      // This prevents SIMPLE/ADVANCED desync where strict equality filters miss empty statuses
      const normalized = raw.map((f: Finding) => ({
        ...f,
        status: f.status || 'open',
        kanban_column: f.kanban_column || 'backlog',
      }));
      setFindings(normalized);
      setError(null);
    } catch (err: any) {
      setError(err.message || i18n.t('errors.fetchFindings'));
    } finally {
      setLoading(false);
    }
  }, [productId]);

  useEffect(() => {
    fetchFindings();
  }, [fetchFindings]);

  return { findings, loading, error, refresh: fetchFindings };
};
