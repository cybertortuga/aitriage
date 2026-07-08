import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import Markdown from 'react-markdown';
import { securityService } from '../../services/securityService';

export const ThreatModelPanel: React.FC = () => {
  const { t } = useTranslation('components');

  const STRIDE = [
    {
      key: 'S',
      name: t('components.threatModelPanel.stride.spoofing.name'),
      icon: 'person_off',
      color: 'text-severity-critical',
      desc: t('components.threatModelPanel.stride.spoofing.desc'),
    },
    {
      key: 'T',
      name: t('components.threatModelPanel.stride.tampering.name'),
      icon: 'edit_off',
      color: 'text-severity-high',
      desc: t('components.threatModelPanel.stride.tampering.desc'),
    },
    {
      key: 'R',
      name: t('components.threatModelPanel.stride.repudiation.name'),
      icon: 'history_toggle_off',
      color: 'text-severity-medium',
      desc: t('components.threatModelPanel.stride.repudiation.desc'),
    },
    {
      key: 'I',
      name: t('components.threatModelPanel.stride.infoDisclosure.name'),
      icon: 'visibility',
      color: 'text-severity-high',
      desc: t('components.threatModelPanel.stride.infoDisclosure.desc'),
    },
    {
      key: 'D',
      name: t('components.threatModelPanel.stride.dos.name'),
      icon: 'block',
      color: 'text-severity-critical',
      desc: t('components.threatModelPanel.stride.dos.desc'),
    },
    {
      key: 'E',
      name: t('components.threatModelPanel.stride.eop.name'),
      icon: 'admin_panel_settings',
      color: 'text-severity-critical',
      desc: t('components.threatModelPanel.stride.eop.desc'),
    },
  ] as const;

  const [context, setContext] = useState('');
  const [result, setResult] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleGenerate = async () => {
    if (!context.trim()) return;
    setLoading(true);
    setError('');
    setResult('');
    try {
      const resp = await securityService.generateThreatModel(context);
      if (resp.ok) {
        setResult(resp.content);
      } else {
        setError(resp.error || t('components.threatModelPanel.generateFailed'));
      }
    } catch (e: any) {
      setError(e.message || t('components.threatModelPanel.connectionError'));
    }
    setLoading(false);
  };

  return (
    <div className="h-full overflow-y-auto p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <span className="material-symbols-outlined text-ai-accent" style={{ fontSize: '24px' }}>
            security
          </span>
          <div>
            <h2 className="text-label-caps tracking-widest text-primary">{t('components.threatModelPanel.title')}</h2>
            <p className="text-label-xs text-on-surface-variant/50 tracking-wider mt-0.5">
              {t('components.threatModelPanel.subtitle')}
            </p>
          </div>
        </div>
      </div>

      {/* STRIDE Grid */}
      <div className="grid grid-cols-3 gap-3">
        {STRIDE.map((cat) => (
          <div
            key={cat.key}
            className="cyber-panel p-4 group hover:border-ai-accent/30 transition-none"
          >
            <div className="flex items-center gap-3 mb-2">
              <div className="w-8 h-8 flex items-center justify-center border border-outline-variant/30 bg-surface-container-lowest">
                <span
                  className={`material-symbols-outlined ${cat.color}`}
                  style={{ fontSize: '18px' }}
                >
                  {cat.icon}
                </span>
              </div>
              <div>
                <span className={`text-2xl font-bold ${cat.color}`}>{cat.key}</span>
                <span className="text-label-xs text-on-surface-variant/60 ml-2 tracking-wider">
                  {cat.name.toUpperCase()}
                </span>
              </div>
            </div>
            <p className="text-label-xs text-on-surface-variant/40 leading-relaxed">{cat.desc}</p>
          </div>
        ))}
      </div>

      {/* Input */}
      <div className="cyber-panel p-4 space-y-3">
        <label className="text-label-xs text-on-surface-variant/60 tracking-widest">
          {t('components.threatModelPanel.targetContext')}
        </label>
        <textarea
          value={context}
          onChange={(e) => setContext(e.target.value)}
          placeholder={t('components.threatModelPanel.placeholder')}
          className="w-full h-24 bg-surface-container-lowest border border-outline-variant/30 p-3 text-mono-data text-primary placeholder:text-on-surface-variant/20 focus:outline-none focus:border-ai-accent/50 resize-none"
        />
        <div className="flex items-center gap-3">
          <button
            onClick={handleGenerate}
            disabled={loading || !context.trim()}
            className="px-5 py-2 bg-ai-accent text-white text-label-xs tracking-widest uppercase hover:bg-ai-accent/80 disabled:opacity-30 disabled:cursor-not-allowed flex items-center gap-2"
          >
            <span className="material-symbols-outlined" style={{ fontSize: '16px' }}>
              {loading ? 'progress_activity' : 'shield'}
            </span>
            {loading ? t('components.threatModelPanel.analyzing') : t('components.threatModelPanel.generateButton')}
          </button>
          {loading && (
            <span className="text-label-xs text-ai-accent/60 tracking-wider animate-pulse">
              {t('components.threatModelPanel.processing')}
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

      {/* Results */}
      {result && (
        <div className="cyber-panel p-6 border-ai-accent/20">
          <div className="flex items-center gap-2 mb-4 pb-3 border-b border-outline-variant/20">
            <span className="material-symbols-outlined text-ai-accent" style={{ fontSize: '16px' }}>
              verified_user
            </span>
            <span className="text-label-xs text-ai-accent tracking-widest">
              {t('components.threatModelPanel.resultsTitle')}
            </span>
          </div>
          <div className="prose prose-invert prose-sm max-w-none [&_h2]:text-primary [&_h2]:text-sm [&_h2]:tracking-widest [&_h2]:uppercase [&_h2]:mt-6 [&_h2]:mb-3 [&_h3]:text-on-surface-variant [&_h3]:text-xs [&_table]:text-xs [&_th]:text-left [&_th]:text-on-surface-variant/60 [&_th]:tracking-wider [&_th]:uppercase [&_td]:text-primary [&_code]:text-ai-accent [&_code]:bg-ai-accent/10 [&_code]:px-1 [&_strong]:text-severity-high [&_li]:text-on-surface-variant">
            <Markdown>{result}</Markdown>
          </div>
        </div>
      )}
    </div>
  );
};
