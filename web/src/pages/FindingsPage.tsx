import React, { useEffect, useState } from 'react';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';
import confetti from 'canvas-confetti';
import { useTranslation } from 'react-i18next';
import { useFindings } from '../hooks/useFindings';
import { useTitle } from '../hooks/useTitle';
import { useCopilotStore } from '../store/CopilotStore';

interface Finding {
  id: number;
  rule_id?: string;
  title: string;
  severity: string;
  stack: string;
  status: string;
  file_path?: string;
  file?: string;
  line_number?: number;
  cwe_id?: string;
  cve_id?: string;
  description?: string;
  code_snippet?: string;
  fix_suggestion?: string;
  suggestion?: string;
  agent_prompt?: string;
  agent_prompt_generated_at?: string;
  verification_status?: string;
  verification_summary?: string;
  verification_last_run_at?: string;
  is_verified?: boolean;
  verified_at?: string;
}

const SEV_COLORS: Record<string, { dot: string; text: string; badge: string }> = {
  CRITICAL: { dot: 'bg-error', text: 'text-error', badge: 'border-error bg-error/10 text-error pulse-glow-critical' },
  HIGH: {
    dot: 'bg-severity-high',
    text: 'text-severity-high',
    badge: 'border-severity-high bg-severity-high/10 text-severity-high',
  },
  MEDIUM: {
    dot: 'bg-severity-medium',
    text: 'text-severity-medium',
    badge: 'border-severity-medium bg-severity-medium/10 text-severity-medium',
  },
  LOW: {
    dot: 'bg-on-surface-variant',
    text: 'text-on-surface-variant',
    badge: 'border-outline-variant text-on-surface-variant',
  },
};

const getSev = (sev: string) => SEV_COLORS[sev?.toUpperCase()] ?? SEV_COLORS.LOW;

export const FindingsPage: React.FC = () => {
  const { t } = useTranslation('pages');
  const { findings, loading, error, refresh } = useFindings() as {
    findings: Finding[];
    loading: boolean;
    error: string | null;
    refresh: (options?: { silent?: boolean }) => void;
  };

  const getSeverityLabel = (severity: string) => {
    const s = severity?.toUpperCase();
    if (s === 'CRITICAL') return t('critical');
    if (s === 'HIGH') return t('high');
    if (s === 'MEDIUM') return t('medium');
    if (s === 'LOW') return t('low');
    return severity;
  };

  const getStatusLabel = (status: string | undefined) => {
    const s = status?.toLowerCase() || 'open';
    if (s === 'sent_to_agent') return t('status_sent_to_agent');
    if (s === 'pending_verification') return t('status_pending_verification');
    if (s === 'verification_failed') return t('status_verification_failed');
    if (s === 'resolved' || s === 'fixed') return t('status_fixed');
    if (s === 'triage') return t('status_triage');
    if (s === 'false_positive') return t('status_false_positive');
    if (s === 'risk_accepted' || s === 'accepted_risk') return t('status_accepted_risk');
    return t('status_open');
  };

  const getStatusTone = (status: string | undefined) => {
    const s = status?.toLowerCase() || 'open';
    if (s === 'resolved' || s === 'fixed') return 'text-success';
    if (s === 'verification_failed') return 'text-error';
    if (s === 'pending_verification') return 'text-severity-medium';
    if (s === 'sent_to_agent' || s === 'triage') return 'text-[#38bdf8]';
    return 'text-primary';
  };

  useTitle(t('findings_title'));
  const [selectedId, setSelectedId] = useState<number | null>(null);
  const [search, setSearch] = useState('');
  const [selectedSeverity, setSelectedSeverity] = useState('ALL_SEVERITIES');
  const [triageStatus, setTriageStatus] = useState<Record<number, 'IDLE' | 'PROCESSING'>>({});
  const [agentPrompt, setAgentPrompt] = useState('');
  const [agentPromptStatus, setAgentPromptStatus] = useState<Record<number, 'IDLE' | 'PROCESSING'>>({});
  const [verificationStatus, setVerificationStatus] = useState<Record<number, 'IDLE' | 'PROCESSING'>>({});
  const [verificationResult, setVerificationResult] = useState<string | null>(null);

  const stacks = Array.from(new Set(findings.map((f: Finding) => f.stack)))
    .filter(Boolean)
    .sort() as string[];
  const [selectedStack, setSelectedStack] = useState('ALL_STACKS');

  const [sortField, setSortField] = useState<'id' | 'severity' | 'stack'>('severity');
  const [sortDir, setSortDir] = useState<'asc' | 'desc'>('desc');

  const filtered = findings
    .filter((f: Finding) => {
      const matchesSearch =
        !search ||
        f.title?.toLowerCase().includes(search.toLowerCase()) ||
        String(f.id).includes(search);
      const matchesSeverity =
        selectedSeverity === 'ALL_SEVERITIES' || f.severity?.toUpperCase() === selectedSeverity;
      const matchesStack = selectedStack === 'ALL_STACKS' || f.stack === selectedStack;
      return matchesSearch && matchesSeverity && matchesStack;
    })
    .sort((a, b) => {
      const dir = sortDir === 'asc' ? 1 : -1;
      if (sortField === 'id') return (a.id - b.id) * dir;
      if (sortField === 'stack') return (a.stack || '').localeCompare(b.stack || '') * dir;
      if (sortField === 'severity') {
        const sevMap: Record<string, number> = { CRITICAL: 4, HIGH: 3, MEDIUM: 2, LOW: 1 };
        const valA = sevMap[a.severity?.toUpperCase()] || 0;
        const valB = sevMap[b.severity?.toUpperCase()] || 0;
        return (valA - valB) * dir;
      }
      return 0;
    });

  const selectedFinding =
    filtered.find((f: Finding) => f.id === selectedId) ??
    (filtered.length > 0 ? filtered[0] : null);
  const showSelectedVerificationStatus =
    selectedFinding?.verification_status &&
    (selectedFinding.status?.toLowerCase() || 'open') === 'open';

  useEffect(() => {
    setAgentPrompt(selectedFinding?.agent_prompt ?? '');
    const status = selectedFinding?.status?.toLowerCase();
    const canShowVerificationResult = status === 'verification_failed' || status === 'resolved' || status === 'fixed';
    setVerificationResult(canShowVerificationResult ? selectedFinding?.verification_summary ?? null : null);
  }, [selectedFinding?.id, selectedFinding?.agent_prompt, selectedFinding?.verification_summary, selectedFinding?.status]);

  const severities = ['CRITICAL', 'HIGH', 'MEDIUM', 'LOW'];

  const triageFinding = async (id: number, action: 'FIX' | 'IGNORE' | 'TRIAGE') => {
    setTriageStatus((prev) => ({ ...prev, [id]: 'PROCESSING' }));
    try {
      const res = await fetch('/api/triage', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          id: String(id),
          project: selectedFinding?.file_path || '.',
          file: selectedFinding?.file_path || '',
          action: action,
        }),
      });
      const data = await res.json();
      if (data.ok) {
        if (action === 'TRIAGE') {
          const { setContext, setIsOpen } = useCopilotStore.getState();
          setContext(selectedFinding);
          setIsOpen(true);
        } else if (action === 'FIX') {
          confetti({
            particleCount: 150,
            spread: 70,
            origin: { y: 0.6 },
            colors: ['#ffffff', '#4caf50', '#38BDF8'],
          });
        }
        refresh();
      }
    } catch (err) {
      console.error('Triage failed', err);
    } finally {
      setTriageStatus((prev) => ({ ...prev, [id]: 'IDLE' }));
    }
  };

  const generateAgentPrompt = async () => {
    if (!selectedFinding) return;
    const id = selectedFinding.id;
    setAgentPromptStatus((prev) => ({ ...prev, [id]: 'PROCESSING' }));
    try {
      const res = await fetch(`/api/findings/${id}/agent-prompt`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
      });
      const data = await res.json();
      if (!data.ok) {
        throw new Error(data.error || 'Failed to generate prompt');
      }
      setAgentPrompt(data.prompt || '');
      setVerificationResult(null);
      try {
        await navigator.clipboard.writeText(data.prompt || '');
      } catch {}
      refresh({ silent: true });
    } catch (err) {
      console.error('Agent prompt generation failed', err);
      setVerificationResult(err instanceof Error ? err.message : 'Prompt generation failed');
    } finally {
      setAgentPromptStatus((prev) => ({ ...prev, [id]: 'IDLE' }));
    }
  };

  const verifyFinding = async () => {
    if (!selectedFinding) return;
    const id = selectedFinding.id;
    setVerificationStatus((prev) => ({ ...prev, [id]: 'PROCESSING' }));
    setVerificationResult(t('verification_running'));
    try {
      const res = await fetch(`/api/findings/${id}/verify`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({}),
      });
      const data = await res.json();
      if (!data.ok) {
        throw new Error(data.error || 'Verification failed');
      }
      setVerificationResult(data.summary || '');
      refresh({ silent: true });
    } catch (err) {
      console.error('Verification failed', err);
      setVerificationResult(err instanceof Error ? err.message : 'Verification failed');
    } finally {
      setVerificationStatus((prev) => ({ ...prev, [id]: 'IDLE' }));
    }
  };

  return (
    <div className="flex flex-col h-full overflow-hidden">
      {/* Page Header */}
      <div className="px-4 py-2 flex justify-between items-center flex-shrink-0 cyber-header-premium border-b border-outline-variant/30">
        <div>
          <p className="text-[9px] font-bold tracking-widest text-on-surface-variant mb-0.5">
            {t('sec_findings')}
          </p>
          <h1 className="text-title-lg font-bold tracking-tight text-primary uppercase">
            {t('vul_audit')}
          </h1>
        </div>
        <div className="flex items-center gap-3">
          {/* Stack Filter */}
          <select
            className="bg-surface-container-lowest border border-outline-variant rounded-lg text-label-xs text-on-surface-variant h-8 px-2 focus:border-primary focus:ring-1 focus:ring-primary/25 outline-none cursor-pointer uppercase tracking-widest transition-all duration-300"
            value={selectedStack}
            onChange={(e) => setSelectedStack(e.target.value)}
          >
            <option value="ALL_STACKS">{t('all_stacks')}</option>
            {stacks.map((s) => (
              <option key={s} value={s}>
                {s.toUpperCase()}
              </option>
            ))}
          </select>

          {/* Severity Filter */}
          <select
            className="bg-surface-container-lowest border border-outline-variant rounded-lg text-label-xs text-on-surface-variant h-8 px-2 focus:border-primary focus:ring-1 focus:ring-primary/25 outline-none cursor-pointer uppercase tracking-widest transition-all duration-300"
            value={selectedSeverity}
            onChange={(e) => setSelectedSeverity(e.target.value)}
          >
            <option value="ALL_SEVERITIES">{t('all_severities')}</option>
            {severities.map((s) => (
              <option key={s} value={s}>
                {getSeverityLabel(s).toUpperCase()}
              </option>
            ))}
          </select>

          <div className="flex items-center border border-outline-variant bg-surface-container-lowest h-8 px-3 gap-2 w-56 rounded-lg focus-within:border-primary focus-within:ring-1 focus-within:ring-primary/25 transition-all duration-300">
            <span
              className="material-symbols-outlined text-on-surface-variant"
              style={{ fontSize: '14px' }}
            >
              search
            </span>
            <input
              className="bg-transparent border-none focus:ring-0 focus:outline-none text-xs text-primary placeholder:text-on-surface-variant/50 w-full"
              placeholder={t('filter_findings')}
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
          </div>
        </div>
      </div>

      {/* Split Layout */}
      <div className="flex-1 flex overflow-hidden">
        {/* Master List */}
        <div className="w-1/2 flex flex-col border-r border-outline-variant overflow-hidden">
          {/* Table Header */}
          <div className="cyber-grid-header flex items-center text-label-xs text-on-surface-variant tracking-widest shrink-0">
            <div className="w-10 py-3 px-3 text-center shrink-0">{t('sts')}</div>
            <div
              className="w-24 py-3 px-3 shrink-0 cursor-pointer hover:text-primary transition-none flex items-center gap-1"
              onClick={() => {
                setSortField('id');
                setSortDir(sortDir === 'asc' ? 'desc' : 'asc');
              }}
            >
              {t('id')} {sortField === 'id' && (sortDir === 'asc' ? '↑' : '↓')}
            </div>
            <div className="flex-1 py-3 px-3 min-w-0">{t('finding')}</div>
            <div
              className="w-24 py-3 px-3 shrink-0 text-center cursor-pointer hover:text-primary transition-none flex items-center justify-center gap-1"
              onClick={() => {
                setSortField('stack');
                setSortDir(sortDir === 'asc' ? 'desc' : 'asc');
              }}
            >
              {t('stack')} {sortField === 'stack' && (sortDir === 'asc' ? '↑' : '↓')}
            </div>
            <div
              className="w-24 py-3 px-3 shrink-0 text-right cursor-pointer hover:text-primary transition-none flex items-center justify-end gap-1"
              onClick={() => {
                setSortField('severity');
                setSortDir(sortDir === 'asc' ? 'desc' : 'asc');
              }}
            >
              {t('severity')} {sortField === 'severity' && (sortDir === 'asc' ? '↑' : '↓')}
            </div>
          </div>

          {/* Rows */}
          <div className="flex-1 overflow-y-auto cyber-scrollbar">
            {loading && (
              <div className="flex flex-col items-center justify-center p-12 gap-4">
                <div className="flex gap-1.5">
                  {[0, 1, 2].map((i) => (
                    <div
                      key={i}
                      className="w-2 h-6 bg-primary animate-pulse"
                      style={{ animationDelay: `${i * 0.15}s` }}
                    />
                  ))}
                </div>
                <span className="text-label-caps text-on-surface-variant tracking-[0.2em] text-xs">
                  {t('loading_findings')}
                </span>
              </div>
            )}
            {error && (
              <div className="p-4 flex items-center gap-3 border border-error m-4 bg-error/5">
                <div className="w-2 h-2 bg-error shrink-0" />
                <span className="text-label-caps text-error">{error}</span>
              </div>
            )}
            {!loading &&
              !error &&
              filtered.map((f) => {
                const sev = getSev(f.severity);
                const isSelected = f.id === selectedId || (!selectedId && f.id === filtered[0]?.id);
                const isCrit = f.severity?.toUpperCase() === 'CRITICAL';
                return (
                  <div
                    key={f.id}
                    onClick={() => setSelectedId(f.id)}
                    className={`cyber-grid-row flex items-center cursor-pointer transition-all duration-250 ease-out ${
                      isSelected
                        ? 'bg-surface-bright text-primary border-l-2 border-l-primary shadow-[inset_4px_0_12px_rgba(139,92,246,0.1)]'
                        : 'border-l-2 border-l-transparent hover:bg-surface hover:border-l-primary'
                    }`}
                  >
                    <div className="w-10 py-3 px-3 flex justify-center shrink-0">
                      <div
                        className={`status-dot ${sev.dot} ${isCrit ? 'pulse-glow-critical' : ''}`}
                      />
                    </div>
                    <div className="w-24 py-3 px-3 text-mono-data text-on-surface-variant shrink-0">
                      #{f.id}
                    </div>
                    <div className="flex-1 py-3 px-3 text-mono-data text-on-surface truncate min-w-0">
                      {f.title}
                    </div>
                    <div className="w-24 py-3 px-3 shrink-0 text-center">
                      <span className="text-[9px] font-bold tracking-tighter px-1.5 py-0.5 border border-outline-variant text-on-surface-variant uppercase bg-surface-container-low/50">
                        {f.stack || t('default_core')}
                      </span>
                    </div>
                    <div
                      className={`w-24 py-3 px-3 text-label-xs tracking-widest shrink-0 text-right font-bold ${sev.text}`}
                    >
                      {getSeverityLabel(f.severity).toUpperCase()}
                    </div>
                  </div>
                );
              })}
            {!loading && filtered.length === 0 && (
              <div className="p-12 flex flex-col items-center justify-center text-center opacity-30">
                <span className="material-symbols-outlined text-4xl mb-4">search_off</span>
                <p className="text-label-caps">{t('no_findings')}</p>
              </div>
            )}
          </div>
        </div>

        {/* Detail Panel */}
        <div className="w-1/2 flex flex-col overflow-y-auto cyber-scrollbar bg-surface-container-lowest">
          {selectedFinding ? (
            <div className="p-6 flex flex-col gap-6 animate-in fade-in slide-in-from-right-4 ">
              {/* Finding header */}
              <div className="flex justify-between items-start gap-4 border-b border-outline-variant pb-5">
                <div className="flex-1 min-w-0">
                  <div className="text-label-xs text-on-surface-variant tracking-widest mb-1">
                    {t('finding_uppercase')} #{selectedFinding.id}
                  </div>
                  <h2 className="text-headline-sm text-primary uppercase tracking-tight">
                    {selectedFinding.title}
                  </h2>
                </div>
                <div className="flex flex-col items-end gap-2">
                  <span
                    className={`text-label-xs px-3 py-1 border tracking-widest shrink-0 font-bold ${getSev(selectedFinding.severity).badge}`}
                  >
                    {getSeverityLabel(selectedFinding.severity).toUpperCase()}
                  </span>
                  <button
                    onClick={() => {
                      const { setContext, setIsOpen } = useCopilotStore.getState();
                      setContext(selectedFinding);
                      setIsOpen(true);
                    }}
                    className="flex items-center gap-1.5 text-label-xs text-primary hover:underline group"
                  >
                    <span
                      className="material-symbols-outlined group-hover:rotate-12 transition-none"
                      style={{ fontSize: '14px' }}
                    >
                      smart_toy
                    </span>
                    {t('ask_copilot')}
                  </button>
                </div>
              </div>

              {/* Metadata chips */}
              <div className="grid grid-cols-2 gap-3">
                <div className="cyber-widget p-3 border-l-2 border-l-primary/30">
                  <span className="text-label-xs text-on-surface-variant block mb-2 tracking-widest">
                    {t('status_report')}
                  </span>
                  <span
                    className={`text-mono-data font-bold ${getStatusTone(selectedFinding.status)}`}
                  >
                    {getStatusLabel(selectedFinding.status).toUpperCase()}
                  </span>
                </div>
                <div className="cyber-widget p-3 border-l-2 border-l-primary/30">
                  <span className="text-label-xs text-on-surface-variant block mb-2 tracking-widest">
                    {t('affected_asset')}
                  </span>
                  <span className="text-mono-data text-on-surface truncate block text-xs underline decoration-outline-variant">
                    {selectedFinding.file_path ?? selectedFinding.file ?? '—'}
                  </span>
                </div>
              </div>

              {/* Description */}
              <div>
                <div className="text-label-xs text-on-surface-variant tracking-widest mb-3 flex items-center gap-2">
                  <span className="material-symbols-outlined text-[14px] text-primary">
                    description
                  </span>
                  {t('system_intelligence')}
                </div>
                <div className="cyber-widget p-4 text-body-sm text-on-surface leading-relaxed border-outline-variant/30">
                  {selectedFinding.description ?? t('no_description')}
                </div>
              </div>

              {/* Code snippet with Syntax Highlighting */}
              {selectedFinding.code_snippet && (
                <div>
                  <div className="text-label-xs text-on-surface-variant tracking-widest mb-3 flex items-center gap-2">
                    <span className="material-symbols-outlined text-[14px] text-primary">code</span>
                    {t('technical_evidence')}
                  </div>
                  <div className="cyber-widget border-outline-variant/30 overflow-hidden">
                    <SyntaxHighlighter
                      language={selectedFinding.stack?.toLowerCase() || 'javascript'}
                      style={vscDarkPlus}
                      customStyle={{
                        margin: 0,
                        padding: '1rem',
                        fontSize: '12px',
                        background: 'transparent',
                      }}
                    >
                      {selectedFinding.code_snippet}
                    </SyntaxHighlighter>
                  </div>
                </div>
              )}

              {/* Remediation */}
              <div>
                <div className="text-label-xs text-on-surface-variant tracking-widest mb-3 flex items-center gap-2">
                  <span className="material-symbols-outlined text-[14px] text-primary">
                    auto_fix_high
                  </span>
                  {t('remediation_protocol')}
                </div>
                <div className="border-l-2 border-primary pl-4 py-1 text-body-sm text-on-surface leading-relaxed italic opacity-80">
                  {selectedFinding.fix_suggestion ??
                    selectedFinding.suggestion ??
                    t('follow_standard')}
                </div>
              </div>

              {/* Agent handoff */}
              <div>
                <div className="text-label-xs text-on-surface-variant tracking-widest mb-3 flex items-center gap-2">
                  <span className="material-symbols-outlined text-[14px] text-primary">
                    smart_toy
                  </span>
                  {t('agent_handoff')}
                </div>
                <div className="cyber-widget p-4 border-outline-variant/30 flex flex-col gap-4">
                  <div className="flex items-center justify-between gap-3">
                    <div className="flex items-center gap-2 min-w-0">
                      <span
                        className={`text-label-xs px-2.5 py-1 border border-outline-variant tracking-widest font-bold ${getStatusTone(selectedFinding.status)}`}
                      >
                        {getStatusLabel(selectedFinding.status).toUpperCase()}
                      </span>
                      {showSelectedVerificationStatus && (
                        <span className="text-[10px] text-on-surface-variant uppercase tracking-widest truncate">
                          {selectedFinding.verification_status}
                        </span>
                      )}
                    </div>
                    <div className="flex items-center gap-2 shrink-0">
                      <button
                        onClick={generateAgentPrompt}
                        className="btn-primary px-3 py-2 rounded-lg text-label-xs flex items-center justify-center gap-2 hover:-translate-y-0.5 active:scale-[0.98] transition-all duration-300 ease-out cursor-pointer"
                        disabled={agentPromptStatus[selectedFinding.id] === 'PROCESSING'}
                      >
                        {agentPromptStatus[selectedFinding.id] === 'PROCESSING' ? (
                          <div className="flex gap-0.5 items-center">
                            {[0, 1, 2].map((i) => (
                              <div
                                key={i}
                                className="w-0.5 h-2.5 bg-current animate-pulse"
                                style={{ animationDelay: `${i * 0.15}s` }}
                              />
                            ))}
                          </div>
                        ) : (
                          <span className="material-symbols-outlined text-[14px]">content_paste</span>
                        )}
                        {t('agent_prompt')}
                      </button>
                      <button
                        onClick={verifyFinding}
                        className="btn-mechanical px-3 py-2 rounded-lg text-label-xs flex items-center justify-center gap-2 hover:-translate-y-0.5 active:scale-[0.98] transition-all duration-300 ease-out cursor-pointer"
                        disabled={verificationStatus[selectedFinding.id] === 'PROCESSING'}
                      >
                        {verificationStatus[selectedFinding.id] === 'PROCESSING' ? (
                          <div className="flex gap-0.5 items-center">
                            {[0, 1, 2].map((i) => (
                              <div
                                key={i}
                                className="w-0.5 h-2.5 bg-current animate-pulse"
                                style={{ animationDelay: `${i * 0.15}s` }}
                              />
                            ))}
                          </div>
                        ) : (
                          <span className="material-symbols-outlined text-[14px]">fact_check</span>
                        )}
                        {verificationStatus[selectedFinding.id] === 'PROCESSING'
                          ? t('verification_running_short')
                          : t('verify_fix')}
                      </button>
                    </div>
                  </div>
                  <div className="text-xs text-on-surface-variant leading-relaxed border-l border-success/20 pl-3">
                    {t('verification_rescan_hint')}
                  </div>
                  {agentPrompt && (
                    <textarea
                      readOnly
                      value={agentPrompt}
                      className="w-full min-h-40 resize-y bg-surface-container-lowest border border-outline-variant/70 rounded-lg p-3 text-xs leading-relaxed text-on-surface font-mono outline-none focus:border-primary cyber-scrollbar"
                    />
                  )}
                  {verificationResult && (verificationStatus[selectedFinding.id] === 'PROCESSING' || ['verification_failed', 'resolved', 'fixed'].includes(selectedFinding.status.toLowerCase())) && (
                    <div
                      className={`border-l-2 pl-3 py-2 text-body-sm leading-relaxed ${
                        selectedFinding.status === 'verification_failed'
                          ? 'border-error text-error'
                          : 'border-primary text-on-surface'
                      }`}
                    >
                      {verificationResult}
                    </div>
                  )}
                </div>
              </div>

              {/* Actions */}
              <div className="border-t border-outline-variant pt-5 flex gap-3">
                <button
                  onClick={() => triageFinding(selectedFinding.id, 'TRIAGE')}
                  className="btn-primary flex-1 py-3 rounded-lg text-label-xs flex items-center justify-center gap-2 hover:-translate-y-0.5 active:scale-[0.98] transition-all duration-300 ease-out cursor-pointer"
                  disabled={triageStatus[selectedFinding.id] === 'PROCESSING'}
                >
                  {triageStatus[selectedFinding.id] === 'PROCESSING' ? (
                     <div className="flex gap-0.5 items-center">
                      {[0, 1, 2].map((i) => (
                        <div
                          key={i}
                          className="w-0.5 h-2.5 bg-current animate-pulse"
                          style={{ animationDelay: `${i * 0.15}s` }}
                        />
                      ))}
                    </div>
                  ) : (
                    <span className="material-symbols-outlined text-[14px]">psychology</span>
                  )}
                  {t('triage_finding')}
                </button>
                <button
                  onClick={() => triageFinding(selectedFinding.id, 'IGNORE')}
                  className="btn-mechanical-error flex-1 py-3 rounded-lg text-label-xs flex items-center justify-center gap-2 hover:-translate-y-0.5 active:scale-[0.98] transition-all duration-300 ease-out cursor-pointer"
                  disabled={triageStatus[selectedFinding.id] === 'PROCESSING'}
                >
                  <span className="material-symbols-outlined text-[14px]">block</span>
                  {t('mark_false_positive')}
                </button>
              </div>
            </div>
          ) : (
            <div className="flex-1 flex items-center justify-center p-8">
              <div className="text-center opacity-20">
                <span className="material-symbols-outlined text-6xl mb-4">shield_with_heart</span>
                <p className="text-label-caps">{t('system_secure')}</p>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
