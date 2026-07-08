import React from 'react';
import { useTranslation } from 'react-i18next';

interface StatusBarProps {
  scanStatus: 'idle' | 'scanning' | 'done' | 'error';
  findingsCount: number;
  secureCoderConnected: boolean;
  version?: string;
}

export const StatusBar: React.FC<StatusBarProps> = ({
  scanStatus,
  findingsCount,
  secureCoderConnected,
  version = '2.0.4',
}) => {
  const { t } = useTranslation('components');
  const statusColor = {
    idle: 'text-on-surface-variant/40',
    scanning: 'text-severity-high',
    done: 'text-success',
    error: 'text-error',
  }[scanStatus];

  const statusLabel = {
    idle: t('components.statusBar.idle'),
    scanning: t('components.statusBar.scanning'),
    done: t('components.statusBar.ready'),
    error: t('components.statusBar.error'),
  }[scanStatus];

  return (
    <div className="h-8 flex items-center px-4 gap-6 border-t border-outline-variant bg-surface-container-lowest/80 shrink-0">
      {/* Engine Status */}
      <div className="flex items-center gap-2">
        <div
          className={`w-1.5 h-1.5 ${scanStatus === 'scanning' ? 'bg-severity-high animate-pulse' : scanStatus === 'done' ? 'bg-success' : 'bg-on-surface-variant/30'}`}
        />
        <span className={`text-label-xs ${statusColor} tracking-widest`}>{statusLabel}</span>
      </div>

      {/* Divider */}
      <div className="w-px h-3 bg-outline-variant/30" />

      {/* Findings */}
      <div className="flex items-center gap-1.5">
        <span className="text-label-xs text-on-surface-variant/40 tracking-widest">{t('components.statusBar.findings')}</span>
        <span
          className={`text-label-xs font-bold tracking-wider ${findingsCount > 0 ? 'text-severity-high' : 'text-success'}`}
        >
          {String(findingsCount).padStart(3, '0')}
        </span>
      </div>

      {/* Divider */}
      <div className="w-px h-3 bg-outline-variant/30" />

      {/* SecureCoder */}
      <div className="flex items-center gap-1.5">
        <div
          className={`w-1.5 h-1.5 ${secureCoderConnected ? 'bg-ai-accent' : 'bg-on-surface-variant/20'}`}
        />
        <span
          className={`text-label-xs tracking-widest ${secureCoderConnected ? 'text-ai-accent' : 'text-on-surface-variant/30'}`}
        >
          {t('components.statusBar.secureCoder')}
        </span>
      </div>

      {/* Spacer */}
      <div className="flex-1" />

      {/* Version */}
      <span className="text-label-xs text-on-surface-variant/20 tracking-widest">
        AITRIAGE v{version}
      </span>
    </div>
  );
};
