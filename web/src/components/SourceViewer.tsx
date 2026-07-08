import React from 'react';
import { PageLayout } from '../ui/PageLayout';
import { TelemetryCard, TelemetrySidebar } from '../ui/Telemetry';
import { MechanicalButton } from '../ui/MechanicalButton';
import { useTranslation } from 'react-i18next';

interface SourceViewerProps {
  filename: string;
  content: string;
  language?: string;
  highlightLines?: number[];
}

export const SourceViewer: React.FC<SourceViewerProps> = ({
  filename,
  content,
  language = 'javascript',
  highlightLines = [],
}) => {
  const { t } = useTranslation('components');
  const lines = content.split('\n');

  return (
    <PageLayout
      title={t('components.sourceViewer.title')}
      subtitle={t('components.sourceViewer.subtitle', { filename })}
      actions={<MechanicalButton variant="outline">{t('components.sourceViewer.downloadSource')}</MechanicalButton>}
    >
      <div className="flex-1 flex flex-col bg-background border border-outline-variant overflow-hidden">
        {/* Code Header */}
        <div className="bg-surface-container border-b border-outline-variant px-6 py-2 flex items-center justify-between">
          <div className="text-[10px] text-primary-fixed-dim font-bold tracking-widest uppercase">
            {filename.split('/').pop()} • {language.toUpperCase()}
          </div>
          <div className="flex gap-2">
            <div className="w-2 h-2 bg-outline-variant" />
            <div className="w-2 h-2 bg-outline-variant" />
            <div className="w-2 h-2 bg-outline-variant" />
          </div>
        </div>

        {/* Code Content */}
        <div className="flex-1 overflow-auto font-mono text-sm leading-relaxed p-6 selection:bg-primary-fixed-dim selection:text-background">
          {lines.map((line, i) => {
            const isHighlighted = highlightLines.includes(i + 1);
            return (
              <div
                key={i}
                className={`flex group ${isHighlighted ? 'bg-error/10 border-l-2 border-error -ml-[2px]' : ''}`}
              >
                <div
                  className={`w-12 pr-4 text-right select-none opacity-20 group-hover:opacity-60 transition-none ${isHighlighted ? 'text-error opacity-100' : ''}`}
                >
                  {(i + 1).toString().padStart(3, '0')}
                </div>
                <div
                  className={`flex-1 whitespace-pre pl-4 ${isHighlighted ? 'text-on-surface font-bold' : 'text-on-surface-variant'}`}
                >
                  {line || ' '}
                </div>
              </div>
            );
          })}
        </div>
      </div>

      <TelemetrySidebar title={t('components.sourceViewer.telemetry.title')}>
        <TelemetryCard label={t('components.sourceViewer.telemetry.loc')} value={lines.length.toString()} description={t('components.sourceViewer.telemetry.linesOfCode')} />
        <TelemetryCard
          label={t('components.sourceViewer.telemetry.encoding')}
          value="UTF-8"
          status="nominal"
          description={t('components.sourceViewer.telemetry.characterSet')}
        />
        <TelemetryCard
          label={t('components.sourceViewer.telemetry.language')}
          value={language.toUpperCase()}
          description={t('components.sourceViewer.telemetry.syntaxEngine')}
        />

        <div className="flex-1 border-t border-outline-variant pt-8">
          <h3 className="text-label-xs text-on-surface-variant mb-4 uppercase tracking-[0.2em]">
            {t('components.sourceViewer.telemetry.analysisHints')}
          </h3>
          <div className="space-y-4">
            <div className="bg-surface-container border border-outline-variant p-4">
              <div className="text-[10px] text-error font-bold mb-2 uppercase tracking-widest">
                {t('components.sourceViewer.telemetry.securityAlert')}
              </div>
              <p className="text-[11px] text-on-surface-variant leading-relaxed italic">
                {t('components.sourceViewer.telemetry.securityAlertDescription')}
              </p>
            </div>
          </div>
        </div>
      </TelemetrySidebar>
    </PageLayout>
  );
};
