import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import Markdown from 'react-markdown';
import { securityService } from '../../services/securityService';

export const AuditReportPanel: React.FC = () => {
  const { t } = useTranslation('components');
  const [context, setContext] = useState('');
  const [report, setReport] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleGenerate = async () => {
    if (!context.trim()) return;
    setLoading(true);
    setError('');
    setReport('');
    try {
      const resp = await securityService.generateAuditReport(context);
      if (resp.ok) {
        setReport(resp.content);
      } else {
        setError(resp.error || t('components.auditReportPanel.generateFailed'));
      }
    } catch (e: any) {
      setError(e.message || t('components.auditReportPanel.connectionError'));
    }
    setLoading(false);
  };

  const handlePrint = () => {
    const printWindow = window.open('', '_blank');
    if (printWindow) {
      printWindow.document.write(`
 <html><head><title>${t('components.auditReportPanel.printTitle')}</title>
 <style>body { font-family: monospace; padding: 2rem; max-width: 800px; margin: 0 auto; }
 h1,h2,h3 { color: #333; } table { border-collapse: collapse; width: 100%; }
 th,td { border: 1px solid #ddd; padding: 8px; text-align: left; }
 th { background: #f5f5f5; } code { background: #f0f0f0; padding: 2px 4px; }
 </style></head><body>${document.querySelector('.audit-report-content')?.innerHTML || ''}</body></html>
 `);
      printWindow.document.close();
      printWindow.print();
    }
  };

  return (
    <div className="h-full overflow-y-auto p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <span className="material-symbols-outlined text-ai-accent" style={{ fontSize: '24px' }}>
            assignment
          </span>
          <div>
            <h2 className="text-label-caps tracking-widest text-primary">{t('components.auditReportPanel.title')}</h2>
            <p className="text-label-xs text-on-surface-variant/50 tracking-wider mt-0.5">
              {t('components.auditReportPanel.subtitle')}
            </p>
          </div>
        </div>
        {report && (
          <button
            onClick={handlePrint}
            className="px-4 py-2 border border-outline-variant/30 text-label-xs tracking-widest text-on-surface-variant hover:text-primary hover:border-primary/30 flex items-center gap-2"
          >
            <span className="material-symbols-outlined" style={{ fontSize: '16px' }}>
              print
            </span>
            {t('components.auditReportPanel.export')}
          </button>
        )}
      </div>

      {/* Report Sections Preview */}
      <div className="grid grid-cols-4 gap-3">
        {[
          { icon: 'summarize', label: t('components.auditReportPanel.sections.executiveSummary'), desc: t('components.auditReportPanel.sections.executiveSummaryDesc') },
          { icon: 'security', label: t('components.auditReportPanel.sections.threatModel'), desc: t('components.auditReportPanel.sections.threatModelDesc') },
          { icon: 'bug_report', label: t('components.auditReportPanel.sections.findings'), desc: t('components.auditReportPanel.sections.findingsDesc') },
          { icon: 'checklist', label: t('components.auditReportPanel.sections.remediation'), desc: t('components.auditReportPanel.sections.remediationDesc') },
        ].map((section) => (
          <div key={section.label} className="cyber-panel p-3">
            <div className="flex items-center gap-2 mb-1">
              <span
                className="material-symbols-outlined text-ai-accent/60"
                style={{ fontSize: '16px' }}
              >
                {section.icon}
              </span>
              <span className="text-label-xs text-primary tracking-wider">
                {section.label.toUpperCase()}
              </span>
            </div>
            <p className="text-[10px] text-on-surface-variant/40 tracking-wider">{section.desc}</p>
          </div>
        ))}
      </div>

      {/* Input */}
      <div className="cyber-panel p-4 space-y-3">
        <label className="text-label-xs text-on-surface-variant/60 tracking-widest">
          {t('components.auditReportPanel.auditTarget')}
        </label>
        <textarea
          value={context}
          onChange={(e) => setContext(e.target.value)}
          placeholder={t('components.auditReportPanel.placeholder')}
          className="w-full h-24 bg-surface-container-lowest border border-outline-variant/30 p-3 text-mono-data text-primary placeholder:text-on-surface-variant/20 focus:outline-none focus:border-ai-accent/50 resize-none"
        />
        <div className="flex items-center gap-3">
          <button
            onClick={handleGenerate}
            disabled={loading || !context.trim()}
            className="px-5 py-2 bg-ai-accent text-white text-label-xs tracking-widest uppercase hover:bg-ai-accent/80 disabled:opacity-30 disabled:cursor-not-allowed flex items-center gap-2"
          >
            <span className="material-symbols-outlined" style={{ fontSize: '16px' }}>
              {loading ? 'progress_activity' : 'assignment'}
            </span>
            {loading ? t('components.auditReportPanel.generating') : t('components.auditReportPanel.generateButton')}
          </button>
          {loading && (
            <span className="text-label-xs text-ai-accent/60 tracking-wider animate-pulse">
              {t('components.auditReportPanel.loadingMessage')}
            </span>
          )}
        </div>
      </div>

      {/* Error */}
      {error && (
        <div className="cyber-panel p-4 border-error/30">
          <div className="flex items-center gap-2 text-error">
            <span className="material-symbols-outlined" style={{ fontSize: '16px' }}>
              error
            </span>
            <span className="text-label-xs tracking-wider">{error}</span>
          </div>
        </div>
      )}

      {/* Report */}
      {report && (
        <div className="cyber-panel p-6 border-ai-accent/20">
          <div className="flex items-center gap-2 mb-4 pb-3 border-b border-outline-variant/20">
            <span className="material-symbols-outlined text-success" style={{ fontSize: '16px' }}>
              task_alt
            </span>
            <span className="text-label-xs text-success tracking-widest">
              {t('components.auditReportPanel.reportGenerated')}
            </span>
            <span className="text-label-xs text-on-surface-variant/30 ml-auto tracking-wider">
              {new Date().toISOString().split('T')[0]}
            </span>
          </div>
          <div className="audit-report-content prose prose-invert prose-sm max-w-none [&_h2]:text-primary [&_h2]:text-sm [&_h2]:tracking-widest [&_h2]:uppercase [&_h2]:mt-6 [&_h2]:mb-3 [&_h3]:text-on-surface-variant [&_h3]:text-xs [&_table]:text-xs [&_th]:text-left [&_th]:text-on-surface-variant/60 [&_th]:tracking-wider [&_th]:uppercase [&_td]:text-primary [&_code]:text-ai-accent [&_code]:bg-ai-accent/10 [&_code]:px-1 [&_strong]:text-severity-high [&_li]:text-on-surface-variant">
            <Markdown>{report}</Markdown>
          </div>
        </div>
      )}
    </div>
  );
};
