import React from 'react';

import { useTranslation } from 'react-i18next';

export const LoadingScreen: React.FC = () => {
  const { t } = useTranslation('components');
  return (
    <div className="flex flex-col items-center justify-center h-64 gap-6">
      <div className="flex gap-1.5">
        {[0, 1, 2].map((i) => (
          <div
            key={i}
            className="w-2 h-6 bg-primary-fixed-dim animate-pulse"
            style={{ animationDelay: `${i * 0.15}s` }}
          />
        ))}
      </div>
      <span className="text-label-caps font-label-caps text-primary-fixed-dim tracking-[0.3em]">
        {t('LoadingScreen.initializing', 'INITIALIZING...')}
      </span>
    </div>
  );
};
