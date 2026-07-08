import React from 'react';
import { PageLayout } from '../ui/PageLayout';
import { TelemetryCard, TelemetrySidebar } from '../ui/Telemetry';
import { MechanicalButton } from '../ui/MechanicalButton';
import { motion } from 'framer-motion';
import { useTranslation } from 'react-i18next';

export const Foundry: React.FC = () => {
  const { t } = useTranslation('components');
  return (
    <PageLayout
      title={t('components.foundry.title')}
      subtitle={t('components.foundry.subtitle')}
      actions={<MechanicalButton variant="primary">{t('components.foundry.deployAllPatches')}</MechanicalButton>}
    >
      <div className="flex-1 flex flex-col items-center justify-center space-y-12 bg-background border border-outline-variant relative overflow-hidden">
        {/* Decorative Scanned Lines */}
        <div className="absolute inset-0 pointer-events-none opacity-[0.02] bg-[linear-gradient(rgba(0,245,255,0.1)_1px,transparent_1px)] bg-[length:100%_4px]" />

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="text-center space-y-4 relative"
        >
          <div className="absolute -top-12 left-1/2 -translate-x-1/2 w-px h-12 bg-gradient-to-t from-primary-fixed-dim to-transparent opacity-20" />
          <h2 className="text-4xl font-thin text-on-surface tracking-[0.2em] uppercase">
            {t('components.foundry.systemReady')}
          </h2>
          <p className="text-on-surface-variant opacity-60 text-xs tracking-[0.3em] font-mono italic">
            {t('components.foundry.awaitingTriage')}
          </p>
        </motion.div>

        <div className="grid grid-cols-3 gap-8 w-full max-w-4xl px-12">
          {[
            { label: t('components.foundry.analyze'), icon: '◈' },
            { label: t('components.foundry.generate'), icon: '▣' },
            { label: t('components.foundry.validate'), icon: '⚙' },
          ].map((step, i) => (
            <motion.div
              key={step.label}
              initial={{ opacity: 0, scale: 0.9 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ delay: i * 0.2 }}
              className="bg-surface-container border border-outline-variant p-8 flex flex-col items-center gap-4 group hover:border-primary-fixed-dim transition-none"
            >
              <div className="text-3xl text-primary-fixed-dim group-hover:scale-110 transition-none">
                {step.icon}
              </div>
              <div className="text-[10px] font-bold tracking-[0.4em] uppercase text-on-surface opacity-40 group-hover:opacity-100">
                {step.label}
              </div>
            </motion.div>
          ))}
        </div>

        <div className="flex gap-4">
          <MechanicalButton variant="outline">{t('components.foundry.runDiagnostics')}</MechanicalButton>
          <MechanicalButton variant="outline">{t('components.foundry.clearPipeline')}</MechanicalButton>
        </div>
      </div>

      <TelemetrySidebar title={t('components.foundry.telemetry.title')}>
        <TelemetryCard label={t('components.foundry.telemetry.patchBuffer')} value={t('components.foundry.telemetry.ready')} description={t('components.foundry.telemetry.queuedRemediations')} />
        <TelemetryCard
          label={t('components.foundry.telemetry.llmAvailability')}
          value={t('components.foundry.telemetry.online')}
          status="nominal"
          description={t('components.foundry.telemetry.gemini')}
        />
        <TelemetryCard
          label={t('components.foundry.telemetry.compilerStatus')}
          value={t('components.foundry.telemetry.stable')}
          status="nominal"
          description={t('components.foundry.telemetry.buildVerification')}
        />

        <div className="flex-1 border-t border-outline-variant pt-8">
          <h3 className="text-label-xs text-on-surface-variant mb-4 uppercase tracking-[0.2em]">
            {t('components.foundry.telemetry.pipelineLogs')}
          </h3>
          <div className="bg-background border border-outline-variant p-4 font-mono text-[9px] text-primary-fixed-dim leading-relaxed h-48 overflow-y-auto">
            [08:42:11] INITIALIZING_CORE...
            <br />
            [08:42:12] CONNECTING_TO_ORCHESTRATOR...
            <br />
            [08:42:13] HANDSHAKE_SUCCESSFUL.
            <br />
            [08:42:15] AWAITING_TRIAGE_DATA...
            <br />
            <span className="animate-pulse">_</span>
          </div>
        </div>
      </TelemetrySidebar>
    </PageLayout>
  );
};
