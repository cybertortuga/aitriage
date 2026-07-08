import React from 'react';
import { motion } from 'framer-motion';
import { useTranslation } from 'react-i18next';

export const Brand: React.FC<{ className?: string }> = ({ className = '' }) => {
  const { t } = useTranslation('components');
  return (
  <div className={`flex items-center gap-3 ${className}`}>
    <div className="w-8 h-8 border-2 border-primary-fixed-dim relative flex items-center justify-center">
      <div className="absolute inset-0 bg-primary-fixed-dim/10 animate-pulse" />
      <span className="text-primary-fixed-dim font-black text-xl italic select-none">A</span>
      <div className="absolute -top-1 -right-1 w-2 h-2 bg-background border border-outline-variant" />
    </div>
      <div className="flex flex-col">
        <span className="text-display-md text-on-surface leading-none tracking-tighter uppercase font-black italic">
          AITriage
        </span>
        <span className="text-[8px] text-on-surface-variant tracking-[0.4em] uppercase font-bold opacity-60">
          {t('Brand.subtitle', 'Forensic_Engine')}
        </span>
      </div>
    </div>
  );
};

export const NavItem: React.FC<{
  active?: boolean;
  icon: string;
  label: string;
  onClick: () => void;
  disabled?: boolean;
}> = ({ active, icon, label, onClick, disabled }) => (
  <button
    onClick={onClick}
    disabled={disabled}
    className={`
 w-full flex items-center gap-4 px-6 py-4 transition-none group relative overflow-hidden
 ${
   active
     ? 'bg-surface-container-high text-primary-fixed-dim'
     : 'text-on-surface-variant hover:bg-surface-container hover:text-on-surface'
 }
 ${disabled ? 'opacity-30 cursor-not-allowed' : 'cursor-pointer'}
 `}
  >
    {active && (
      <motion.div
        layoutId="nav-active"
        className="absolute left-0 top-0 bottom-0 w-1 bg-primary-fixed-dim"
      />
    )}
    <span className="material-symbols-outlined text-[20px] group-hover:scale-110 transition-none">
      {icon}
    </span>
    <span className="text-[10px] font-bold uppercase tracking-[0.2em]">{label}</span>
    {active && <div className="ml-auto w-1.5 h-1.5 bg-primary-fixed-dim animate-pulse" />}
  </button>
);
