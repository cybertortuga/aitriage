import React from 'react';

import { useTranslation } from 'react-i18next';

const events = [
  'SYSTEM_BOOT: KERNEL_INIT_COMPLETE',
  'MONITORING_SERVICE: ACTIVE',
  'SECURITY_PROTOCOL: ENFORCED',
  'VAULT_INTEGRITY: VERIFIED',
  'SCAN_ENGINE: STANDBY',
  'AUDIT_PIPELINE: OPERATIONAL',
  'NETWORK_SECURITY: ACTIVE',
  'HEARTBEAT: SYSTEM_OK',
];

export const SystemTicker: React.FC = () => {
  const { t } = useTranslation('components');
  return (
    <div className="h-6 bg-surface-container-highest border-t border-outline-variant overflow-hidden flex items-center shrink-0">
      <div className="flex-shrink-0 bg-primary px-3 h-full flex items-center">
        <span className="text-label-xs text-on-primary font-bold">{t('SystemTicker.logLive', 'SYSTEM_LOG_LIVE')}</span>
      </div>
      <div className="flex-1 overflow-hidden whitespace-nowrap flex items-center relative">
        <div className="animate-ticker flex gap-12 pl-6">
          {[...events, ...events].map((event, i) => (
            <div key={i} className="flex items-center gap-2">
              <div className="w-1 h-1 bg-primary/40 rotate-45" />
              <span className="text-label-xs text-on-surface-variant uppercase">{event}</span>
            </div>
          ))}
        </div>

        {/* Gradients to fade edges */}
        <div className="absolute inset-y-0 left-0 w-8 bg-gradient-to-r from-surface-container-highest to-transparent z-10" />
        <div className="absolute inset-y-0 right-0 w-8 bg-gradient-to-l from-surface-container-highest to-transparent z-10" />
      </div>
      <div className="flex-shrink-0 px-3 border-l border-outline-variant h-full flex items-center gap-4">
        <div className="flex items-center gap-1.5">
          <div className="w-1.5 h-1.5 bg-green-500 animate-pulse" />
          <span className="text-label-xs text-on-surface-variant">{t('SystemTicker.healthOk', 'HEALTH: OK')}</span>
        </div>
        <div className="text-label-xs text-on-surface-variant opacity-40">
          {new Date().toISOString()}
        </div>
      </div>
    </div>
  );
};
