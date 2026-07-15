import React, { useEffect, useMemo, useRef, useState } from 'react';
import Markdown from 'react-markdown';
import { useTranslation } from 'react-i18next';
import { useAgentHandoff } from '../../hooks/useAgentHandoff';
import { runwayArtifactURL } from '../../services/runwayArtifacts';
import './AgentHandoffPanel.css';

type HandoffView = 'prompt' | 'data';

interface AgentHandoffPanelProps {
  sessionId: number | null;
  className?: string;
  compact?: boolean;
  enabled?: boolean;
}

const extractImplementationBrief = (markdown: string) => {
  const start = markdown.indexOf('```markdown');
  if (start < 0) return markdown;
  const bodyStart = start + '```markdown'.length;
  const end = markdown.indexOf('\n```', bodyStart);
  return markdown.slice(bodyStart, end < 0 ? undefined : end).trim();
};

const formatBytes = (bytes: number) => {
  if (bytes < 1024) return `${bytes} B`;
  return `${(bytes / 1024).toFixed(bytes < 10 * 1024 ? 1 : 0)} KB`;
};

const copyText = async (value: string) => {
  if (navigator.clipboard?.writeText) {
    await navigator.clipboard.writeText(value);
    return;
  }
  const input = document.createElement('textarea');
  input.value = value;
  input.style.position = 'fixed';
  input.style.opacity = '0';
  document.body.appendChild(input);
  input.select();
  document.execCommand('copy');
  input.remove();
};

export const AgentHandoffPanel: React.FC<AgentHandoffPanelProps> = ({
  sessionId,
  className = '',
  compact = false,
  enabled = true,
}) => {
  const { t } = useTranslation('pages');
  const { data, loading, error, reload } = useAgentHandoff(sessionId, enabled);
  const [activeView, setActiveView] = useState<HandoffView | null>(null);
  const [copied, setCopied] = useState<HandoffView | null>(null);
  const [copyError, setCopyError] = useState('');
  const [wrapJSON, setWrapJSON] = useState(true);
  const [promptMode, setPromptMode] = useState<'preview' | 'raw'>('preview');
  const backButtonRef = useRef<HTMLButtonElement>(null);
  const triggerRefs = useRef<Record<HandoffView, HTMLButtonElement | null>>({ prompt: null, data: null });
  const lastViewRef = useRef<HandoffView>('prompt');

  const prompt = useMemo(
    () => extractImplementationBrief(data?.remediation_prompt_markdown || ''),
    [data?.remediation_prompt_markdown],
  );
  const agentJSON = useMemo(
    () => (data ? JSON.stringify(data.agent_data, null, 2) : ''),
    [data],
  );

  const handleCopy = async (view: HandoffView) => {
    try {
      await copyText(view === 'prompt' ? prompt : agentJSON);
      setCopyError('');
      setCopied(view);
      window.setTimeout(() => setCopied((current) => (current === view ? null : current)), 1600);
    } catch {
      setCopyError(t('AgentHandoffPanel.copyFailed'));
    }
  };

  const openView = (view: HandoffView) => {
    lastViewRef.current = view;
    setCopyError('');
    setActiveView(view);
  };

  const closeView = () => {
    setActiveView(null);
    window.requestAnimationFrame(() => triggerRefs.current[lastViewRef.current]?.focus());
  };

  useEffect(() => {
    if (activeView) backButtonRef.current?.focus();
  }, [activeView]);

  if (!sessionId) return null;

  return (
    <section className={`agent-handoff ${compact ? 'agent-handoff--compact' : ''} ${className}`} aria-label={t('AgentHandoffPanel.regionLabel')}>
      <header className="agent-handoff__header">
        <div className="agent-handoff__identity">
          <span className="material-symbols-outlined" aria-hidden="true">smart_toy</span>
          <div>
            <strong>{t('AgentHandoffPanel.title')}</strong>
            <span>{t('AgentHandoffPanel.subtitle')}</span>
          </div>
        </div>
        {data && (
          <div className="agent-handoff__status" data-gate={data.gate_status.toLowerCase()}>
            <i aria-hidden="true" />
            {data.gate_status}
          </div>
        )}
      </header>

      {loading && (
        <div className="agent-handoff__state" role="status">
          <span className="agent-handoff__spinner" aria-hidden="true" />
          {t('AgentHandoffPanel.loading')}
        </div>
      )}

      {error && (
        <div className="agent-handoff__error" role="alert">
          <span className="material-symbols-outlined" aria-hidden="true">error</span>
          <div><strong>{t('AgentHandoffPanel.unavailable')}</strong><span>{error}</span></div>
          <button type="button" onClick={reload}>{t('AgentHandoffPanel.retry')}</button>
        </div>
      )}

      {data && !activeView && (
        <div className="agent-handoff__menu">
          <button
            ref={(node) => { triggerRefs.current.prompt = node; }}
            type="button"
            onClick={() => openView('prompt')}
            disabled={!prompt}
            aria-expanded="false"
            aria-controls={`agent-handoff-${sessionId}-detail`}
            aria-label={t('AgentHandoffPanel.openPrompt')}
          >
            <span className="agent-handoff__menu-icon material-symbols-outlined" aria-hidden="true">assignment</span>
            <span className="agent-handoff__menu-copy">
              <strong>{t('AgentHandoffPanel.promptTitle')}</strong>
              <small>{prompt ? t('AgentHandoffPanel.promptDescription') : t('AgentHandoffPanel.noActionable')}</small>
            </span>
            <span className="agent-handoff__meta">{t('AgentHandoffPanel.actionableWithSize', { count: data.agent_data.findings.length, size: formatBytes(data.remediation_prompt_size_bytes) })}</span>
            <span className="material-symbols-outlined agent-handoff__chevron" aria-hidden="true">chevron_right</span>
          </button>
          <button
            ref={(node) => { triggerRefs.current.data = node; }}
            type="button"
            onClick={() => openView('data')}
            aria-expanded="false"
            aria-controls={`agent-handoff-${sessionId}-detail`}
            aria-label={t('AgentHandoffPanel.openData')}
          >
            <span className="agent-handoff__menu-icon material-symbols-outlined" aria-hidden="true">data_object</span>
            <span className="agent-handoff__menu-copy">
              <strong>{t('AgentHandoffPanel.dataTitle')}</strong>
              <small>{t('AgentHandoffPanel.dataDescription')}</small>
            </span>
            <span className="agent-handoff__meta">JSON · v{data.schema_version} · {formatBytes(data.agent_data_size_bytes)}</span>
            <span className="material-symbols-outlined agent-handoff__chevron" aria-hidden="true">chevron_right</span>
          </button>
        </div>
      )}

      {data && activeView && (
        <div className="agent-handoff__detail" id={`agent-handoff-${sessionId}-detail`}>
          <div className="agent-handoff__toolbar">
            <button ref={backButtonRef} type="button" className="agent-handoff__back" onClick={closeView}>
              <span className="material-symbols-outlined" aria-hidden="true">arrow_back</span>
              {t('AgentHandoffPanel.back')}
            </button>
            <div className="agent-handoff__toolbar-title">
              <h3>{activeView === 'prompt' ? t('AgentHandoffPanel.promptTitle') : t('AgentHandoffPanel.dataTitle')}</h3>
              <span>{activeView === 'prompt' ? t('AgentHandoffPanel.copyIntoIDE') : t('AgentHandoffPanel.canonicalData')}</span>
            </div>
            <div className="agent-handoff__actions">
              {activeView === 'prompt' && (
                <div className="agent-handoff__segmented" aria-label="Prompt display mode">
                  <button type="button" onClick={() => setPromptMode('preview')} aria-pressed={promptMode === 'preview'}>{t('AgentHandoffPanel.preview')}</button>
                  <button type="button" onClick={() => setPromptMode('raw')} aria-pressed={promptMode === 'raw'}>{t('AgentHandoffPanel.raw')}</button>
                </div>
              )}
              {activeView === 'data' && (
                <button type="button" onClick={() => setWrapJSON((value) => !value)} aria-pressed={wrapJSON}>
                  <span className="material-symbols-outlined" aria-hidden="true">wrap_text</span>
                  {wrapJSON ? t('AgentHandoffPanel.noWrap') : t('AgentHandoffPanel.wrap')}
                </button>
              )}
              <button type="button" onClick={() => handleCopy(activeView)}>
                <span className="material-symbols-outlined" aria-hidden="true">{copied === activeView ? 'check' : 'content_copy'}</span>
                {copied === activeView ? t('AgentHandoffPanel.copied') : t('AgentHandoffPanel.copy')}
              </button>
              <a
                href={runwayArtifactURL(sessionId, activeView === 'prompt' ? 'remediation_prompt_markdown' : 'agent_data_json')}
                download
              >
                <span className="material-symbols-outlined" aria-hidden="true">download</span>
                {t('AgentHandoffPanel.download')}
              </a>
            </div>
          </div>

          <div className="agent-handoff__content" data-view={activeView}>
            {activeView === 'prompt' ? (
              promptMode === 'preview' ? (
                <div className="agent-handoff__markdown"><Markdown>{prompt}</Markdown></div>
              ) : (
                <pre className="is-wrapped"><code>{prompt}</code></pre>
              )
            ) : (
              <pre className={wrapJSON ? 'is-wrapped' : ''}><code>{agentJSON}</code></pre>
            )}
          </div>
          <footer className="agent-handoff__footer">
            <span>SHA-256</span>
            <code>{activeView === 'prompt' ? data.remediation_prompt_sha256 : data.agent_data_sha256}</code>
          </footer>
          <div className="agent-handoff__announce" aria-live="polite">
            {copyError || (copied === activeView ? t('AgentHandoffPanel.copiedAnnouncement', { item: activeView === 'prompt' ? 'Prompt' : 'JSON' }) : '')}
          </div>
        </div>
      )}
    </section>
  );
};
