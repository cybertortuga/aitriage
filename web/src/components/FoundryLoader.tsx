import React from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Shield, Zap, Cpu, Search, Activity } from 'lucide-react';
export interface LoadingStep {
  name: string;
  status: 'pending' | 'active' | 'done';
}
import { Brand } from '../ui/Brand';
import { useTranslation } from 'react-i18next';

interface FoundryLoaderProps {
  steps: LoadingStep[];
  progress: number;
  currentStep: string;
}

export const FoundryLoader: React.FC<FoundryLoaderProps> = ({ steps, progress, currentStep }) => {
  const { t } = useTranslation('components');
  const getIcon = (name: string) => {
    if (name.includes('AST')) return <Cpu size={16} />;
    if (name.includes('Entropy')) return <Activity size={16} />;
    if (name.includes('Git')) return <Search size={16} />;
    if (name.includes('Dependency')) return <Zap size={16} />;
    return <Shield size={16} />;
  };

  return (
    <div className="fixed inset-0 z-[100] bg-background flex flex-col items-center justify-center overflow-hidden">
      {/* Background Cyber Grid */}
      <div
        className="absolute inset-0 opacity-[0.05] pointer-events-none"
        style={{
          backgroundImage:
            'linear-gradient(rgba(0,245,255,0.1) 1px, transparent 1px), linear-gradient(90deg, rgba(0,245,255,0.1) 1px, transparent 1px)',
          backgroundSize: '40px 40px',
        }}
      />

      <div className="relative w-full max-w-2xl px-12 space-y-16">
        {/* Branding & Status */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="flex flex-col items-center"
        >
          <Brand className="mb-12 scale-125" />

          <div className="flex flex-col items-center gap-2">
            <h2 className="text-display-md text-on-surface tracking-[0.3em] uppercase italic font-black">
              {t('components.foundryLoader.title')}
            </h2>
            <div className="flex items-center gap-3">
              <span className="w-2 h-2 bg-primary-fixed-dim animate-ping" />
              <span className="text-[10px] text-primary-fixed-dim font-bold tracking-[0.4em] uppercase">
                {t('components.foundryLoader.subtitle')}
              </span>
            </div>
          </div>
        </motion.div>

        {/* Progress System */}
        <div className="space-y-4">
          <div className="flex justify-between items-end">
            <div className="flex flex-col gap-1">
              <span className="text-[10px] text-on-surface-variant uppercase tracking-widest opacity-60">
                {t('components.foundryLoader.activeProcess')}
              </span>
              <span className="text-sm text-on-surface font-bold font-mono tracking-tighter uppercase truncate max-w-md">
                &gt; {currentStep || t('components.foundryLoader.initializingCore')}
              </span>
            </div>
            <div className="text-right flex flex-col gap-1">
              <span className="text-[10px] text-on-surface-variant uppercase tracking-widest opacity-60">
                {t('components.foundryLoader.loadRatio')}
              </span>
              <span className="text-xl text-primary-fixed-dim font-black font-mono">
                {Math.round(progress)}%
              </span>
            </div>
          </div>

          <div className="h-2 bg-surface-container border border-outline-variant p-[2px]">
            <motion.div
              className="h-full bg-primary-fixed-dim relative overflow-hidden"
              initial={{ width: 0 }}
              animate={{ width: `${progress}%` }}
              transition={{ type: 'spring', stiffness: 50, damping: 20 }}
            >
              <div className="absolute inset-0 bg-[linear-gradient(90deg,transparent,rgba(255,255,255,0.3),transparent)] animate-shimmer" />
            </motion.div>
          </div>
        </div>

        {/* Subsystem Manifest */}
        <div className="grid grid-cols-1 gap-2">
          <AnimatePresence>
            {steps.map((step, idx) => (
              <motion.div
                key={step.name}
                initial={{ opacity: 0, x: -10 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: idx * 0.05 }}
                className={`
 flex items-center gap-6 p-4 border transition-none 
 ${
   step.status === 'active'
     ? 'bg-surface-container-high border-primary-fixed-dim'
     : step.status === 'done'
       ? 'bg-surface-container-low border-outline-variant opacity-60'
       : 'bg-transparent border-transparent opacity-20'
 }
 `}
              >
                <div
                  className={`${step.status === 'active' ? 'text-primary-fixed-dim' : 'text-on-surface-variant'}`}
                >
                  {getIcon(step.name)}
                </div>

                <div className="flex-1">
                  <div
                    className={`text-[10px] uppercase tracking-[0.2em] font-bold ${
                      step.status === 'active' ? 'text-on-surface' : 'text-on-surface-variant'
                    }`}
                  >
                    {step.name}
                  </div>
                </div>

                <div className="flex items-center gap-4">
                  {step.status === 'done' && (
                    <span className="text-[10px] text-primary-fixed-dim font-black uppercase tracking-widest italic">
                      {t('components.foundryLoader.nominal')}
                    </span>
                  )}

                  {step.status === 'active' && (
                    <div className="flex gap-1">
                      {[0, 1, 2].map((i) => (
                        <motion.div
                          key={i}
                          animate={{ scale: [1, 1.5, 1], opacity: [0.3, 1, 0.3] }}
                          transition={{ duration: 0.8, repeat: Infinity, delay: i * 0.2 }}
                          className="w-1.5 h-1.5 bg-primary-fixed-dim"
                        />
                      ))}
                    </div>
                  )}
                </div>
              </motion.div>
            ))}
          </AnimatePresence>
        </div>

        {/* Global Footer Warning */}
        <div className="text-center">
          <p className="text-[9px] text-on-surface-variant opacity-30 uppercase tracking-[0.4em] font-mono">
            {t('components.foundryLoader.warning')}
          </p>
        </div>
      </div>
    </div>
  );
};
