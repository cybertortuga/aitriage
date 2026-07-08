import React from 'react';
import type { AuditLog } from '../../hooks/useAdmin';
import { format } from 'date-fns';
import { useTranslation } from 'react-i18next';

interface AuditLogTabProps {
  logs: AuditLog[];
}

export const AuditLogTab: React.FC<AuditLogTabProps> = ({ logs }) => {
  const { t } = useTranslation('components');
  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center bg-surface-container-low p-6 border border-outline-variant relative overflow-hidden">
        <div className="absolute top-0 left-0 w-1 h-full bg-on-surface-variant/20" />
        <div>
          <h3 className="text-label-caps text-on-surface-variant tracking-[0.3em] italic">
            {t('components.auditLogTab.complianceAuditStream')}
          </h3>
          <p className="text-label-xs text-on-surface-variant opacity-40 mt-1 uppercase">
            {t('components.auditLogTab.description')}
          </p>
        </div>
        <button className="btn-mechanical-secondary px-6 py-2 text-[10px] font-black uppercase tracking-widest flex items-center gap-2">
          <span className="material-symbols-outlined text-[16px]">download</span>[
          {t('components.auditLogTab.downloadBundle')} ]
        </button>
      </div>

      <div className="cyber-panel p-4 font-mono text-[10px] leading-relaxed max-h-[600px] overflow-auto cyber-scrollbar bg-surface-container/40">
        {logs.length === 0 ? (
          <div className="text-center py-20 opacity-40 italic flex flex-col items-center gap-4">
            <span className="material-symbols-outlined text-[32px]">folder_off</span>
            <span className="text-label-caps tracking-widest">
              {t('components.auditLogTab.noRecords')}
            </span>
          </div>
        ) : (
          <div className="space-y-1">
            {logs.map((log) => (
              <div
                key={log.id}
                className="flex gap-4 py-1.5 border-b border-outline-variant/10 last:border-0 hover:bg-white/5 transition-none px-3 group"
              >
                <span className="opacity-30 shrink-0 tabular-nums">
                  {format(new Date(log.created_at), 'yyyy-MM-dd HH:mm:ss')}
                </span>
                <span className="text-primary/60 font-bold shrink-0 min-w-[120px]">
                  [{log.username || 'SYSTEM'}]
                </span>
                <span className="text-on-surface opacity-80 uppercase tracking-tighter">
                  {log.action}
                </span>
                <span className="opacity-40 truncate italic">
                  {log.entity_type && `::${log.entity_type}`}
                  {log.entity_id && `(${log.entity_id})`}
                </span>
                <span className="ml-auto font-bold text-success/60 group-hover:text-success transition-none">
                  ::{t('components.auditLogTab.success')}
                </span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};
