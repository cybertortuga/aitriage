import React, { useState } from 'react';
import { PageLayout } from '../ui/PageLayout';
import { MechanicalButton } from '../ui/MechanicalButton';
import { DataGrid, DataGridRow } from '../ui/DataGrid';
import { TelemetryCard, TelemetrySidebar } from '../ui/Telemetry';
import type { Dependency } from '../types';
import { useTranslation } from 'react-i18next';

interface DependenciesProps {
  dependencies: Dependency[];
}

export const Dependencies: React.FC<DependenciesProps> = ({ dependencies }) => {
  const { t } = useTranslation('components');
  const [searchTerm, setSearchTerm] = useState('');

  const filteredDeps = dependencies.filter(
    (d) =>
      d.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      d.ecosystem.toLowerCase().includes(searchTerm.toLowerCase()),
  );

  return (
    <PageLayout
      title={t('components.dependencies.title')}
      subtitle={t('components.dependencies.subtitle', { count: dependencies.length })}
      actions={
        <>
          <div className="flex items-center gap-3 bg-surface-container border border-outline-variant px-4 py-2">
            <span className="text-[10px] text-primary-fixed-dim font-bold tracking-widest">
              {t('components.dependencies.filter')}
            </span>
            <input
              type="text"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              placeholder={t('components.dependencies.searchPlaceholder')}
              className="bg-transparent border-none text-on-surface focus:ring-0 text-code-sm w-64 uppercase font-code"
            />
          </div>
          <MechanicalButton>{t('components.dependencies.exportSbom')}</MechanicalButton>
        </>
      }
    >
      <DataGrid
        header={
          <>
            <div className="w-16 py-3 px-6 text-center">{t('components.dependencies.headers.id')}</div>
            <div className="flex-1 py-3 px-6">{t('components.dependencies.headers.componentName')}</div>
            <div className="w-32 py-3 px-6">{t('components.dependencies.headers.version')}</div>
            <div className="w-32 py-3 px-6">{t('components.dependencies.headers.ecosystem')}</div>
            <div className="w-48 py-3 px-6 text-right">{t('components.dependencies.headers.scope')}</div>
          </>
        }
      >
        {filteredDeps.map((dep, idx) => (
          <DataGridRow key={`${dep.name}-${idx}`} idx={idx}>
            <div className="w-16 py-3 px-6 text-center opacity-40">
              {idx.toString().padStart(3, '0')}
            </div>
            <div className="flex-1 py-3 px-6 font-bold text-on-surface">
              {dep.name.toUpperCase()}
            </div>
            <div className="w-32 py-3 px-6">
              <span className="text-[10px] bg-background border border-outline-variant px-2 py-1 text-primary-fixed-dim">
                {dep.version}
              </span>
            </div>
            <div className="w-32 py-3 px-6">
              <span className="text-[10px] opacity-60 uppercase tracking-tighter italic">
                [{dep.ecosystem}]
              </span>
            </div>
            <div className="w-48 py-3 px-6 text-right text-[10px] uppercase opacity-40 tracking-widest italic">
              {dep.type}
            </div>
          </DataGridRow>
        ))}
      </DataGrid>

      <TelemetrySidebar title={t('components.dependencies.telemetry.title')}>
        <TelemetryCard
          label={t('components.dependencies.telemetry.directDeps')}
          value={dependencies.filter((d) => d.type === 'direct').length}
          description={t('components.dependencies.telemetry.immediateUpstream')}
        />
        <TelemetryCard
          label={t('components.dependencies.telemetry.transitiveDeps')}
          value={dependencies.filter((d) => d.type !== 'direct').length}
          description={t('components.dependencies.telemetry.deepGraphNodes')}
        />
        <TelemetryCard
          label={t('components.dependencies.telemetry.riskVectors')}
          value={t('components.dependencies.telemetry.zeroDetected')}
          status="nominal"
          description={t('components.dependencies.telemetry.supplyChainIntegrity')}
        />

        <div className="flex-1 border-t border-outline-variant pt-8">
          <h3 className="text-label-xs text-on-surface-variant mb-4 uppercase tracking-[0.2em]">
            {t('components.dependencies.telemetry.integrityCheck')}
          </h3>
          <div className="bg-background border border-outline-variant p-4 text-[11px] text-on-surface-variant leading-relaxed">
            {t('components.dependencies.telemetry.integrityCheckDescription')}
          </div>
        </div>
      </TelemetrySidebar>
    </PageLayout>
  );
};
