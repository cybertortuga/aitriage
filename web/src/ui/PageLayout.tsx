import React from 'react';
import { useTranslation } from 'react-i18next';

interface PageLayoutProps {
  title: string;
  subtitle?: string;
  actions?: React.ReactNode;
  children: React.ReactNode;
}

export const PageLayout: React.FC<PageLayoutProps> = ({ title, subtitle, actions, children }) => {
  const { t } = useTranslation('components');
  return (
    <div className="flex flex-col h-full overflow-hidden font-sans text-sm selection:bg-primary selection:text-white bg-transparent text-[#f4f4f5]">
      {/* Page Header */}
      <div className="px-10 py-8 flex justify-between items-end flex-shrink-0 border-b border-white/5 bg-surface/40 backdrop-blur-md relative z-10">
        <div className="flex flex-col gap-1.5">
          <span className="font-mono text-[10px] text-primary tracking-[0.2em] uppercase">
            {t('PageLayout.systemDirectory')}
          </span>
          <h1 className="font-display text-3xl font-semibold tracking-tight text-white">{title}</h1>
          {subtitle && <p className="font-sans text-sm text-white/40 mt-1">{subtitle}</p>}
        </div>
        {actions && <div className="flex gap-4">{actions}</div>}
      </div>

      {/* Main Content Area */}
      <div
        className="flex-1 flex overflow-hidden v2-fade-in"
        style={{ animationDuration: '400ms' }}
      >
        {children}
      </div>
    </div>
  );
};
