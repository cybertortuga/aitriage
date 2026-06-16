import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import Markdown from 'react-markdown';
import { securityService } from '../../services/securityService';

type ActiveView = 'plan' | 'poc';

export const SecurityPlanPanel: React.FC = () => {
  const { t } = useTranslation('components');
  const [context, setContext] = useState('');
  const [plan, setPlan] = useState('');
  const [poc, setPoc] = useState('');
  const [loading, setLoading] = useState(false);
  const [activeView, setActiveView] = useState<ActiveView>('plan');
  const [error, setError] = useState('');

  const handleGeneratePlan = async () => {
    if (!context.trim()) return;
    setLoading(true);
    setError('');
    try {
      const resp = await securityService.generateSecurityPlan(context);
      if (resp.ok) {
        setPlan(resp.content);
        setActiveView('plan');
      } else {
        setError(resp.error || t('SecurityPlanPanel.failedToGeneratePlan'));
      }
    } catch (e: any) {
      setError(e.message || t('SecurityPlanPanel.connectionError'));
    }
    setLoading(false);
  };

  const handleGeneratePoC = async () => {
    if (!context.trim()) return;
    setLoading(true);
    setError('');
    try {
      const resp = await securityService.generatePoC(context);
      if (resp.ok) {
        setPoc(resp.content);
        setActiveView('poc');
      } else {
        setError(resp.error || t('SecurityPlanPanel.failedToGeneratePoC'));
      }
    } catch (e: any) {
      setError(e.message || t('SecurityPlanPanel.connectionError'));
    }
    setLoading(false);
  };

  const content = activeView === 'plan' ? plan : poc;

  return (
    <div className="h-full overflow-y-auto p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center gap-3">
        <span className="material-symbols-outlined text-ai-accent" style={{ fontSize: '24px' }}>
          build
        </span>
        <div>
          <h2 className="text-label-caps tracking-widest text-primary">
            {t('SecurityPlanPanel.remediationAndVerification')}
          </h2>
          <p className="text-label-xs text-on-surface-variant/50 tracking-wider mt-0.5">
            {t('SecurityPlanPanel.securityPlanAndPoc')}
          </p>
        </div>
      </div>

      {/* Toggle */}
      <div className="flex gap-0 border border-outline-variant/30 w-fit">
        <button
          onClick={() => setActiveView('plan')}
          className={`px-4 py-2 text-label-xs tracking-widest flex items-center gap-2 ${
            activeView === 'plan'
              ? 'bg-ai-accent/20 text-ai-accent border-r border-outline-variant/30'
              : 'text-on-surface-variant/50 hover:text-primary border-r border-outline-variant/30'
          }`}
        >
          <span className="material-symbols-outlined" style={{ fontSize: '16px' }}>
            checklist
          </span>
          {t('SecurityPlanPanel.implPlan')}
        </button>
        <button
          onClick={() => setActiveView('poc')}
          className={`px-4 py-2 text-label-xs tracking-widest flex items-center gap-2 ${
            activeView === 'poc'
              ? 'bg-severity-critical/20 text-severity-critical'
              : 'text-on-surface-variant/50 hover:text-primary'
          }`}
        >
          <span className="material-symbols-outlined" style={{ fontSize: '16px' }}>
            bug_report
          </span>
          {t('SecurityPlanPanel.pocAttacks')}
        </button>
      </div>

      {/* Input */}
      <div className="cyber-panel p-4 space-y-3">
        <label className="text-label-xs text-on-surface-variant/60 tracking-widest">
          {activeView === 'plan'
            ? t('SecurityPlanPanel.vulnContextRemediation')
            : t('SecurityPlanPanel.targetForPocSimulation')}
        </label>
        <textarea
          value={context}
          onChange={(e) => setContext(e.target.value)}
          placeholder={
            activeView === 'plan'
              ? t('SecurityPlanPanel.planPlaceholder')
              : t('SecurityPlanPanel.pocPlaceholder')
          }
          className="w-full h-24 bg-surface-container-lowest border border-outline-variant/30 p-3 text-mono-data text-primary placeholder:text-on-surface-variant/20 focus:outline-none focus:border-ai-accent/50 resize-none"
        />
        <div className="flex items-center gap-3">
          {activeView === 'plan' ? (
            <button
              onClick={handleGeneratePlan}
              disabled={loading || !context.trim()}
              className="px-5 py-2 bg-ai-accent text-white text-label-xs tracking-widest uppercase hover:bg-ai-accent/80 disabled:opacity-30 disabled:cursor-not-allowed flex items-center gap-2"
            >
              <span className="material-symbols-outlined" style={{ fontSize: '16px' }}>
                {loading ? 'progress_activity' : 'build'}
              </span>
              {loading ? t('SecurityPlanPanel.planning') : t('SecurityPlanPanel.generateSecurityPlan')}
            </button>
          ) : (
            <button
              onClick={handleGeneratePoC}
              disabled={loading || !context.trim()}
              className="px-5 py-2 bg-severity-critical text-white text-label-xs tracking-widest uppercase hover:bg-severity-critical/80 disabled:opacity-30 disabled:cursor-not-allowed flex items-center gap-2"
            >
              <span className="material-symbols-outlined" style={{ fontSize: '16px' }}>
                {loading ? 'progress_activity' : 'bug_report'}
              </span>
              {loading ? t('SecurityPlanPanel.simulating') : t('SecurityPlanPanel.generatePocAttacks')}
            </button>
          )}
          {loading && (
            <span className="text-label-xs text-ai-accent/60 tracking-wider animate-pulse">
              {t('SecurityPlanPanel.secureCoderProcessing')}
            </span>
          )}
        </div>
      </div>

      {/* Warning for PoC */}
      {activeView === 'poc' && (
        <div className="cyber-panel p-3 border-severity-medium/30 flex items-center gap-3">
          <span
            className="material-symbols-outlined text-severity-medium"
            style={{ fontSize: '18px' }}
          >
            warning
          </span>
          <span className="text-label-xs text-severity-medium/80 tracking-wider">
            {t('SecurityPlanPanel.pocWarning')}
          </span>
        </div>
      )}

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

      {/* Content */}
      {content && (
        <div
          className={`cyber-panel p-6 ${activeView === 'poc' ? 'border-severity-critical/20' : 'border-ai-accent/20'}`}
        >
          <div className="flex items-center gap-2 mb-4 pb-3 border-b border-outline-variant/20">
            <span
              className="material-symbols-outlined"
              style={{
                fontSize: '16px',
                color:
                  activeView === 'poc'
                    ? 'var(--color-severity-critical)'
                    : 'var(--color-ai-accent)',
              }}
            >
              {activeView === 'poc' ? 'bug_report' : 'task_alt'}
            </span>
            <span
              className={`text-label-xs tracking-widest ${activeView === 'poc' ? 'text-severity-critical' : 'text-ai-accent'}`}
            >
              {activeView === 'poc' ? t('SecurityPlanPanel.pocSimulationResults') : t('SecurityPlanPanel.implementationPlan')}
            </span>
          </div>
          <div className="prose prose-invert prose-sm max-w-none [&_h2]:text-primary [&_h2]:text-sm [&_h2]:tracking-widest [&_h2]:uppercase [&_h2]:mt-6 [&_h2]:mb-3 [&_h3]:text-on-surface-variant [&_h3]:text-xs [&_table]:text-xs [&_th]:text-left [&_th]:text-on-surface-variant/60 [&_th]:tracking-wider [&_th]:uppercase [&_td]:text-primary [&_code]:text-ai-accent [&_code]:bg-ai-accent/10 [&_code]:px-1 [&_strong]:text-severity-high [&_li]:text-on-surface-variant">
            <Markdown>{content}</Markdown>
          </div>
        </div>
      )}
    </div>
  );
};
