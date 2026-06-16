import React from 'react';
import { useTranslation } from 'react-i18next';

interface TelemetryCardProps {
  label: string;
  value: string | number;
  status?: 'nominal' | 'warning' | 'error' | 'active';
  description?: string;
}

export const TelemetryCard: React.FC<TelemetryCardProps> = ({
  label,
  value,
  status = 'nominal',
  description,
}) => {
  const statusColors = {
    nominal: 'text-primary-fixed-dim',
    warning: 'text-[#ff9900]',
    error: 'text-error',
    active: 'text-primary-container',
  };

  return (
    <div className="metrics-chip group hover:border-primary-container/30 transition-none">
      <div className="flex justify-between items-start mb-1">
        <span className="metrics-chip-label">{label}</span>
        {status === 'active' && (
          <div className="w-1.5 h-1.5 bg-primary-container animate-pulse mt-1" />
        )}
      </div>
      <div className={`metrics-chip-value ${statusColors[status]}`}>{value}</div>
      {description && (
        <div className="mt-2 text-label-xs text-on-surface-variant opacity-40 italic">
          // {description}
        </div>
      )}
    </div>
  );
};

export const TelemetrySidebar: React.FC<{ title: string; children: React.ReactNode }> = ({
  title,
  children,
}) => {
  const { t } = useTranslation('components');
  return (
    <div className="w-96 bg-surface-container-lowest flex flex-col p-8 border-l border-outline-variant flex-shrink-0">
      <h3 className="text-label-xs text-on-surface-variant mb-6 uppercase tracking-[0.2em] font-black italic">
        :: {title}
      </h3>
      <div className="flex-1 overflow-y-auto scrollbar-hide space-y-6">{children}</div>
      <div className="mt-auto pt-8 border-t border-outline-variant flex items-center justify-between">
        <div className="flex items-center gap-2">
          <div className="w-2 h-2 bg-primary-fixed-dim animate-pulse" />
          <span className="text-label-xs text-on-surface-variant">{t('Telemetry.uplinkStable')}</span>
        </div>
        <span className="text-mono-data text-on-surface-variant opacity-30">v4.2.0_SEC_CORE</span>
      </div>
    </div>
  );
};
