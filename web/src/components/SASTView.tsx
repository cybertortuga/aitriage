import React, { useState } from 'react';
import { PageLayout } from '../ui/PageLayout';
import { MechanicalButton } from '../ui/MechanicalButton';
import { DataGrid, DataGridRow } from '../ui/DataGrid';
import { TelemetryCard, TelemetrySidebar } from '../ui/Telemetry';
import { useTranslation } from 'react-i18next';

interface SASTFinding {
  tool: string;
  rule_id: string;
  file: string;
  line: number;
  message: string;
  severity: string;
}

interface SASTViewProps {
  findings: SASTFinding[];
}

export const SASTView: React.FC<SASTViewProps> = ({ findings }) => {
  const { t } = useTranslation('components');
  const [filter, setFilter] = useState('');
  const [selectedFinding, setSelectedFinding] = useState<SASTFinding | null>(null);

  const filtered = findings.filter(
    (f) =>
      f.message.toLowerCase().includes(filter.toLowerCase()) ||
      f.file.toLowerCase().includes(filter.toLowerCase()) ||
      f.tool.toLowerCase().includes(filter.toLowerCase()),
  );

  const getSeverityStyle = (sev: string) => {
    switch (sev.toUpperCase()) {
      case 'CRITICAL':
        return 'text-severity-critical font-bold';
      case 'HIGH':
        return 'text-severity-high font-bold';
      case 'MEDIUM':
        return 'text-severity-medium';
      default:
        return 'text-on-surface-variant opacity-60';
    }
  };

  return (
    <PageLayout
      title={t('components.sastView.title')}
      subtitle={t('components.sastView.subtitle')}
      actions={
        <>
          <div className="flex items-center gap-3 bg-surface-container border border-outline-variant px-4 py-2">
            <span className="text-[10px] text-primary-fixed-dim font-bold tracking-widest">
              {t('components.sastView.search')}
            </span>
            <input
              type="text"
              value={filter}
              onChange={(e) => setFilter(e.target.value)}
              placeholder={t('components.sastView.filterPlaceholder')}
              className="bg-transparent border-none text-on-surface focus:ring-0 text-code-sm w-64 uppercase font-code"
            />
          </div>
          <MechanicalButton variant="primary">{t('components.sastView.triggerScan')}</MechanicalButton>
        </>
      }
    >
      <DataGrid
        header={
          <>
            <div className="w-24 py-3 px-6">{t('components.sastView.headers.tool')}</div>
            <div className="w-16 py-3 px-6 text-center">{t('components.sastView.headers.severity')}</div>
            <div className="flex-1 py-3 px-6">{t('components.sastView.headers.messageBuffer')}</div>
            <div className="w-64 py-3 px-6 text-right">{t('components.sastView.headers.targetLocation')}</div>
          </>
        }
      >
        {filtered.map((f, idx) => (
          <DataGridRow
            key={idx}
            idx={idx}
            active={selectedFinding === f}
            onClick={() => setSelectedFinding(f)}
          >
            <div className="w-24 py-3 px-6 opacity-40 uppercase tracking-tighter italic">
              {f.tool}
            </div>
            <div
              className={`w-16 py-3 px-6 text-center text-[10px] ${getSeverityStyle(f.severity)}`}
            >
              {f.severity.substring(0, 3).toUpperCase()}
            </div>
            <div className="flex-1 py-3 px-6 font-bold text-on-surface truncate">
              {f.message.toUpperCase()}
            </div>
            <div className="w-64 py-3 px-6 text-right text-[10px] text-on-surface-variant font-mono truncate italic opacity-60">
              {f.file.split('/').pop()}:{f.line}
            </div>
          </DataGridRow>
        ))}
      </DataGrid>

      <TelemetrySidebar title={t('components.sastView.telemetry.title')}>
        {selectedFinding ? (
          <div className="space-y-8">
            <div className="space-y-2">
              <span className="text-[10px] text-primary-fixed-dim block uppercase font-bold tracking-widest">
                {t('components.sastView.telemetry.ruleIdentifier')}
              </span>
              <div className="text-sm font-bold text-on-surface break-all">
                {selectedFinding.rule_id}
              </div>
            </div>

            <TelemetryCard
              label={t('components.sastView.telemetry.severityLevel')}
              value={selectedFinding.severity.toUpperCase()}
              status={
                ['high', 'critical'].includes(selectedFinding.severity.toLowerCase())
                  ? 'error'
                  : selectedFinding.severity.toLowerCase() === 'medium'
                    ? 'warning'
                    : 'nominal'
              }
              description={t('components.sastView.telemetry.scannerDetermination')}
            />

            <div className="space-y-2">
              <span className="text-[10px] text-on-surface-variant block uppercase tracking-widest opacity-60">
                {t('components.sastView.telemetry.detailedReport')}
              </span>
              <div className="bg-background border border-outline-variant p-4 text-[11px] text-on-surface-variant leading-relaxed">
                {selectedFinding.message}
              </div>
            </div>

            <div className="space-y-2">
              <span className="text-[10px] text-on-surface-variant block uppercase tracking-widest opacity-60">
                {t('components.sastView.telemetry.resourcePath')}
              </span>
              <div className="text-[10px] font-mono p-3 bg-surface-container border border-outline-variant break-all text-on-surface italic">
                {selectedFinding.file}:{selectedFinding.line}
              </div>
            </div>

            <MechanicalButton variant="primary" className="w-full">
              {t('components.sastView.telemetry.initializeFix')}
            </MechanicalButton>
          </div>
        ) : (
          <div className="h-full flex flex-col items-center justify-center text-center opacity-30 italic p-12">
            <div className="text-4xl mb-6 font-thin">!</div>
            <div className="text-[10px] uppercase tracking-[0.3em] leading-relaxed">
              {t('components.sastView.telemetry.selectEvent')}
            </div>
          </div>
        )}
      </TelemetrySidebar>
    </PageLayout>
  );
};
