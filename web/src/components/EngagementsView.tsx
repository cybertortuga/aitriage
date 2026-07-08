import React, { useState, useEffect } from 'react';
import { PageLayout } from '../ui/PageLayout';
import { DataGrid, DataGridRow } from '../ui/DataGrid';
import { TelemetrySidebar, TelemetryCard } from '../ui/Telemetry';
import { MechanicalButton } from '../ui/MechanicalButton';
import type { Engagement } from '../types';
import api from '../services/api';
import { useTranslation } from 'react-i18next';

export const EngagementsView: React.FC = () => {
  const { t } = useTranslation('components');
  const [engagements, setEngagements] = useState<Engagement[]>([]);
  const [selectedEngagement, setSelectedEngagement] = useState<Engagement | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api
      .get('/engagements')
      .then((res) => setEngagements(res.data.engagements || []))
      .catch((err) => console.error('Failed to load engagements', err))
      .finally(() => setLoading(false));
  }, []);

  const downloadReport = (id: string, format: 'json' | 'csv') => {
    window.open(`/api/reports/engagement/${id}?format=${format}`, '_blank');
  };

  const downloadExecutiveReport = (format: 'json' | 'csv') => {
    window.open(`/api/reports/executive?format=${format}`, '_blank');
  };

  return (
    <PageLayout
      title={t('components.engagementsView.title')}
      subtitle={t('components.engagementsView.subtitle')}
      actions={
        <div className="flex gap-2">
          <MechanicalButton variant="outline" onClick={() => downloadExecutiveReport('csv')}>
            {t('components.engagementsView.executiveReportCsv')}
          </MechanicalButton>
          <MechanicalButton variant="primary" onClick={() => {}}>
            {t('components.engagementsView.newEngagement')}
          </MechanicalButton>
        </div>
      }
    >
      <div className="flex-1 flex flex-col overflow-hidden">
        <DataGrid
          header={
            <>
              <div className="w-16 py-3 px-6 text-center">{t('components.engagementsView.headers.id')}</div>
              <div className="flex-1 py-3 px-6">{t('components.engagementsView.headers.engagementName')}</div>
              <div className="w-48 py-3 px-6 text-center">{t('components.engagementsView.headers.status')}</div>
              <div className="w-32 py-3 px-6 text-center">{t('components.engagementsView.headers.startDate')}</div>
            </>
          }
        >
          {loading ? (
            <div className="flex-1 flex items-center justify-center animate-pulse text-primary-fixed-dim uppercase tracking-[0.5em] text-[10px] italic">
              {t('components.engagementsView.synchronizing')}
            </div>
          ) : (
            engagements.map((e, idx) => (
              <DataGridRow
                key={e.id}
                idx={idx}
                active={selectedEngagement?.id === e.id}
                onClick={() => setSelectedEngagement(e)}
              >
                <div className="w-16 py-3 px-6 text-center opacity-40 font-mono text-[10px]">
                  {e.id.slice(0, 4)}
                </div>
                <div className="flex-1 py-3 px-6 font-bold text-on-surface uppercase italic tracking-tight">
                  {e.name}
                </div>
                <div className="w-48 py-3 px-6 flex justify-center">
                  <span
                    className={`px-3 py-1 border text-[9px] font-black uppercase tracking-widest ${
                      e.status === 'completed'
                        ? 'border-primary-fixed-dim text-primary-fixed-dim bg-primary-fixed-dim/5'
                        : e.status === 'in_progress'
                          ? 'border-warning text-warning bg-warning/5'
                          : 'border-outline-variant text-on-surface-variant'
                    }`}
                  >
                    {e.status}
                  </span>
                </div>
                <div className="w-32 py-3 px-6 text-center font-mono text-[10px] opacity-60">
                  {e.start_date.split('T')[0]}
                </div>
              </DataGridRow>
            ))
          )}
        </DataGrid>
      </div>

      <TelemetrySidebar title={t('components.engagementsView.telemetry.title')}>
        {selectedEngagement ? (
          <div className="space-y-8">
            <div className="space-y-2">
              <span className="text-[10px] text-primary-fixed-dim block uppercase font-bold tracking-widest opacity-60">
                {t('components.engagementsView.telemetry.auditIdentifier')}
              </span>
              <div className="text-sm font-bold text-on-surface italic uppercase">
                {selectedEngagement.name}
              </div>
            </div>

            <TelemetryCard
              label={t('components.engagementsView.telemetry.currentStatus')}
              value={selectedEngagement.status.toUpperCase()}
              status={selectedEngagement.status === 'completed' ? 'nominal' : 'warning'}
              description={t('components.engagementsView.telemetry.lifecycleStage')}
            />

            <div className="border-t border-outline-variant pt-8 space-y-6">
              <h3 className="text-label-xs text-on-surface-variant mb-4 uppercase tracking-[0.2em] opacity-40">
                {t('components.engagementsView.telemetry.forensicReporting')}
              </h3>

              <div className="grid grid-cols-2 gap-2">
                <MechanicalButton
                  variant="outline"
                  className="w-full text-[10px]"
                  onClick={() => downloadReport(selectedEngagement.id, 'json')}
                >
                  {t('components.engagementsView.telemetry.exportJson')}
                </MechanicalButton>
                <MechanicalButton
                  variant="outline"
                  className="w-full text-[10px]"
                  onClick={() => downloadReport(selectedEngagement.id, 'csv')}
                >
                  {t('components.engagementsView.telemetry.exportCsv')}
                </MechanicalButton>
              </div>

              <MechanicalButton
                variant="primary"
                className="w-full py-4 tracking-[0.3em] italic"
                onClick={() => {}}
              >
                {t('components.engagementsView.telemetry.viewFullAudit')}
              </MechanicalButton>
            </div>
          </div>
        ) : (
          <div className="h-full flex flex-col items-center justify-center text-center opacity-30 italic p-12">
            <div className="text-4xl mb-6 font-thin">?</div>
            <div className="text-[10px] uppercase tracking-[0.3em] leading-relaxed">
              {t('components.engagementsView.telemetry.selectAudit')}
            </div>
          </div>
        )}
      </TelemetrySidebar>
    </PageLayout>
  );
};
