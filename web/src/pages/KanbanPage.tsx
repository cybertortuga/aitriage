import React, { useState, useEffect, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { KanbanBoard } from '../components/kanban/KanbanBoard';
import { useTitle } from '../hooks/useTitle';
import type { Finding } from '../types';
import api from '../services/api';

export const KanbanPage: React.FC = () => {
  const { t } = useTranslation('pages');
  useTitle(t('kanban.title'));

  const [findings, setFindings] = useState<Finding[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchFindings = useCallback(async () => {
    try {
      setLoading(true);
      const response = await api.get('/findings');
      setFindings(response.data?.findings || response.data || []);
      setError(null);
    } catch {
      setError(t('kanban.errorLoad'));
    } finally {
      setLoading(false);
    }
  }, [t]);

  useEffect(() => {
    fetchFindings();
  }, [fetchFindings]);

  return (
    <div className="flex flex-col h-full overflow-hidden">
      {/* Page Header */}
      <div className="px-4 py-2 border-b border-outline-variant flex justify-between items-center flex-shrink-0">
        <div>
          <p className="text-[9px] font-bold tracking-widest text-on-surface-variant mb-0.5">
            {t('kanban.breadcrumb')}
          </p>
          <h1 className="text-title-lg font-bold tracking-tight text-primary uppercase">
            {t('kanban.header')}
          </h1>
        </div>
        <button onClick={fetchFindings} className="btn-secondary h-8 px-4 flex items-center gap-2">
          <span className="material-symbols-outlined" style={{ fontSize: '14px' }}>
            refresh
          </span>
          {t('kanban.refresh')}
        </button>
      </div>

      {/* Board */}
      <div className="flex-1 overflow-hidden">
        {loading ? (
          <div className="flex items-center justify-center h-full py-24">
            <span className="text-label-caps text-on-surface-variant animate-pulse">
              {t('kanban.loading')}
            </span>
          </div>
        ) : error ? (
          <div className="m-8 p-4 border border-error bg-error-container/10 flex items-center gap-3">
            <div className="w-2 h-2 bg-error shrink-0" />
            <span className="text-label-caps text-error">{error}</span>
          </div>
        ) : (
          <KanbanBoard findings={findings} onFindingsChange={setFindings} />
        )}
      </div>
    </div>
  );
};
