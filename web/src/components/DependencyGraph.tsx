import React from 'react';
import { PageLayout } from '../ui/PageLayout';
import { TelemetryCard, TelemetrySidebar } from '../ui/Telemetry';
import { MechanicalButton } from '../ui/MechanicalButton';
import { motion } from 'framer-motion';
import { useTranslation } from 'react-i18next';

interface DependencyNode {
  id: string;
  name: string;
  version: string;
  dependencies: string[];
}

interface DependencyGraphProps {
  nodes: DependencyNode[];
}

export const DependencyGraph: React.FC<DependencyGraphProps> = ({ nodes }) => {
  const { t } = useTranslation('components');
  return (
    <PageLayout
      title={t('components.dependencyGraph.title')}
      subtitle={t('components.dependencyGraph.subtitle')}
      actions={<MechanicalButton variant="outline">{t('components.dependencyGraph.refreshGraph')}</MechanicalButton>}
    >
      <div className="flex-1 relative bg-background border border-outline-variant overflow-hidden group">
        {/* Grid Background Effect */}
        <div
          className="absolute inset-0 opacity-[0.05] pointer-events-none"
          style={{
            backgroundImage: 'radial-gradient(circle at 1px 1px, var(--accent-color) 1px, transparent 0)',
            backgroundSize: '40px 40px',
          }}
        />

        {/* Graph Implementation with Framer Motion */}
        <div className="absolute inset-0 flex items-center justify-center">
          <div className="relative w-full h-full p-20">
            {nodes.slice(0, 8).map((node, i) => (
              <motion.div
                key={node.id}
                initial={{ opacity: 0, scale: 0.8 }}
                animate={{ opacity: 1, scale: 1 }}
                transition={{ delay: i * 0.1 }}
                className="absolute"
                style={{
                  left: `${20 + (i % 3) * 30}%`,
                  top: `${20 + Math.floor(i / 3) * 30}%`,
                }}
              >
                <div className="bg-surface-container border border-outline-variant p-4 w-48 relative">
                  <div className="absolute -top-1 -left-1 w-2 h-2 bg-primary" />
                  <div className="text-[10px] text-primary font-bold mb-1 truncate">
                    {node.name.toUpperCase()}
                  </div>
                  <div className="text-[9px] text-on-surface-variant opacity-60 font-mono italic">
                    v{node.version}
                  </div>
                  <div className="mt-3 flex gap-1">
                    {node.dependencies.slice(0, 3).map((_, idx) => (
                      <div key={idx} className="w-1 h-1 bg-outline-variant" />
                    ))}
                  </div>
                </div>
              </motion.div>
            ))}

            {/* Visual Connections */}
            <svg className="absolute inset-0 w-full h-full pointer-events-none opacity-15">
              <line
                x1="30%"
                y1="30%"
                x2="60%"
                y2="30%"
                stroke="var(--accent-color)"
                strokeWidth="1"
                strokeDasharray="4 4"
              />
              <line
                x1="30%"
                y1="30%"
                x2="30%"
                y2="60%"
                stroke="var(--accent-color)"
                strokeWidth="1"
                strokeDasharray="4 4"
              />
              <line
                x1="60%"
                y1="30%"
                x2="60%"
                y2="60%"
                stroke="var(--accent-color)"
                strokeWidth="1"
                strokeDasharray="4 4"
              />
            </svg>
          </div>
        </div>

        {/* Viewport Overlay */}
        <div className="absolute bottom-6 right-6 flex flex-col gap-2">
          <div className="bg-background/80 border border-outline-variant p-3 text-[10px] uppercase tracking-widest text-on-surface-variant">
            {t('components.dependencyGraph.renderEngine')}
          </div>
        </div>
      </div>

      <TelemetrySidebar title={t('components.dependencyGraph.telemetry.title')}>
        <TelemetryCard label={t('components.dependencyGraph.telemetry.graphDepth')} value={t('components.dependencyGraph.telemetry.sevenLayers')} description={t('components.dependencyGraph.telemetry.maxRecursionLimit')} />
        <TelemetryCard
          label={t('components.dependencyGraph.telemetry.orphanNodes')}
          value={t('components.dependencyGraph.telemetry.zeroDetected')}
          status="nominal"
          description={t('components.dependencyGraph.telemetry.isolatedComponents')}
        />
        <TelemetryCard
          label={t('components.dependencyGraph.telemetry.circularDeps')}
          value={t('components.dependencyGraph.telemetry.zeroDetected')}
          status="nominal"
          description={t('components.dependencyGraph.telemetry.recursiveIntegrity')}
        />

        <div className="flex-1 border-t border-outline-variant pt-8">
          <h3 className="text-label-xs text-on-surface-variant mb-4 uppercase tracking-[0.2em]">
            {t('components.dependencyGraph.telemetry.graphLegend')}
          </h3>
          <div className="space-y-4">
            <div className="flex items-center gap-3">
              <div className="w-2 h-2 bg-primary" />
              <span className="text-[10px] text-on-surface-variant uppercase tracking-wider">
                {t('components.dependencyGraph.telemetry.primaryModule')}
              </span>
            </div>
            <div className="flex items-center gap-3">
              <div className="w-2 h-2 bg-outline-variant" />
              <span className="text-[10px] text-on-surface-variant uppercase tracking-wider">
                {t('components.dependencyGraph.telemetry.upstreamDependency')}
              </span>
            </div>
            <div className="flex items-center gap-3">
              <div className="w-2 h-2 border border-primary border-dashed" />
              <span className="text-[10px] text-on-surface-variant uppercase tracking-wider">
                {t('components.dependencyGraph.telemetry.peerRelation')}
              </span>
            </div>
          </div>
        </div>
      </TelemetrySidebar>
    </PageLayout>
  );
};
