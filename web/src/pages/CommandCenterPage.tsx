import React, { useState, useEffect, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { useTitle } from '../hooks/useTitle';
import { CCOverviewPanel } from './CCOverviewPanel';
import { FindingsPage } from './FindingsPage';
import { ScannersPage } from './ScannersPage';
import { TopologyPage } from './TopologyPage';
import { TerminalPage } from './TerminalPage';
import { ReportsPage } from './ReportsPage';
import { RulesPage } from './RulesPage';
import { AITriageFramework } from './AITriageFramework';
import { DashboardPage } from './DashboardPage';
import { securityService } from '../services/securityService';

/* ── Tab definitions ───────────────────────────────────────────────── */
const TABS = [
  { key: 0, label: 'Overview', icon: 'dashboard', hotkey: '1', group: 'core', i18nKey: 'commandCenter.tabs.overview' },
  { key: 1, label: 'Findings', icon: 'gavel', hotkey: '2', group: 'core', i18nKey: 'commandCenter.tabs.findings' },
  { key: 2, label: 'Scanners', icon: 'fingerprint', hotkey: '3', group: 'core', i18nKey: 'commandCenter.tabs.scanners' },
  { key: 3, label: 'Topology', icon: 'account_tree', hotkey: '4', group: 'core', i18nKey: 'commandCenter.tabs.topology' },
  { key: 4, label: 'Policies', icon: 'policy', hotkey: '5', group: 'core', i18nKey: 'commandCenter.tabs.policies' },
  { key: 5, label: 'Reports', icon: 'description', hotkey: '6', group: 'core', i18nKey: 'commandCenter.tabs.reports' },
  { key: 6, label: 'AI Triage', icon: 'smart_toy', hotkey: '7', group: 'securecoder', i18nKey: 'commandCenter.tabs.aiTriage' },
  { key: 7, label: 'Terminal', icon: 'terminal', hotkey: '8', group: 'ops', i18nKey: 'commandCenter.tabs.terminal' },
] as const;

/* ── Status types ──────────────────────────────────────────────────── */
interface SystemStatus {
  scanStatus: 'idle' | 'scanning' | 'done' | 'error';
  findingsCount: number;
  secureCoderConnected: boolean;
  tools: Record<string, boolean>;
  uptime: string;
  toolCount: number;
}

export const CommandCenterPage: React.FC = () => {
  const { t } = useTranslation('pages');
  useTitle(t('commandCenter.title'));
  const [activeTab, setActiveTab] = useState(0);
  const [status, setStatus] = useState<SystemStatus>({
    scanStatus: 'idle',
    findingsCount: 0,
    secureCoderConnected: false,
    tools: {},
    uptime: '00:00:00',
    toolCount: 0,
  });
  const [startTime] = useState(Date.now());

  /* ── Health polling ──────────────────────────────────────────────── */
  const refreshStatus = useCallback(async () => {
    try {
      const health = await securityService.getHealth();
      const tools = health.tools || {};
      const activeTools = Object.values(tools).filter(Boolean).length;
      setStatus((prev) => ({
        ...prev,
        tools,
        toolCount: activeTools,
        secureCoderConnected: !!tools['securecoder'],
        scanStatus: health.ok ? 'done' : 'error',
      }));
    } catch {
      setStatus((prev) => ({ ...prev, scanStatus: 'error' }));
    }
    try {
      const findings = await securityService.getFindings();
      setStatus((prev) => ({ ...prev, findingsCount: findings.length }));
    } catch {
      /* skip */
    }
  }, []);

  useEffect(() => {
    refreshStatus();
    const iv = setInterval(refreshStatus, 15000);
    return () => clearInterval(iv);
  }, [refreshStatus]);

  /* ── Uptime clock ───────────────────────────────────────────────── */
  useEffect(() => {
    const iv = setInterval(() => {
      const elapsed = Math.floor((Date.now() - startTime) / 1000);
      const h = String(Math.floor(elapsed / 3600)).padStart(2, '0');
      const m = String(Math.floor((elapsed % 3600) / 60)).padStart(2, '0');
      const s = String(elapsed % 60).padStart(2, '0');
      setStatus((prev) => ({ ...prev, uptime: `${h}:${m}:${s}` }));
    }, 1000);
    return () => clearInterval(iv);
  }, [startTime]);

  /* ── Keyboard shortcuts ─────────────────────────────────────────── */
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      const tag = (e.target as HTMLElement).tagName;
      if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT') return;
      if (e.key >= '1' && e.key <= '8') {
        e.preventDefault();
        setActiveTab(parseInt(e.key) - 1);
      }
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, []);

  /* ── Tab content renderer ───────────────────────────────────────── */
  const renderContent = () => {
    switch (activeTab) {
      case 0:
        return <CCOverviewPanel />;
      case 1:
        return <FindingsPage />;
      case 2:
        return <ScannersPage />;
      case 3:
        return <TopologyPage />;
      case 4:
        return <RulesPage />;
      case 5:
        return <ReportsPage />;
      case 6:
        return <AITriageFramework />;
      case 7:
        return <TerminalPage />;
      default:
        return <DashboardPage />;
    }
  };

  const scanColor = {
    idle: 'var(--v2-muted)',
    scanning: '#F59E0B',
    done: '#39ff14',
    error: 'var(--v2-red)',
  }[status.scanStatus];
  const scanLabel = {
    idle: t('commandCenter.status.standby'),
    scanning: t('commandCenter.status.scanning'),
    done: t('commandCenter.status.operational'),
    error: t('commandCenter.status.offline'),
  }[status.scanStatus];

  return (
    <div className="flex flex-col h-full overflow-hidden bg-v2-bg">
      {/* ═══ STATUS STRIP ═══ */}
      <div className="shrink-0 flex items-center justify-between px-6 h-10 border-b border-v2-border-soft bg-v2-surface">
        <div className="flex items-center gap-3">
          <div className="w-1.5 h-1.5 bg-v2-red rounded-full animate-pulse" />
          <span className="text-[11px] font-bold text-white tracking-widest uppercase">
            {t('commandCenter.title')}
          </span>
          <span className="text-[10px] text-v2-muted ml-2 font-mono">
            {t('commandCenter.subtitle')}
          </span>
        </div>
        <div className="flex items-center gap-5 text-[10px] font-bold uppercase tracking-wider">
          <div className="flex items-center gap-2">
            <div className="w-1.5 h-1.5 rounded-full" style={{ backgroundColor: scanColor }} />
            <span style={{ color: scanColor }}>{scanLabel}</span>
          </div>
          <div className="w-px h-3 bg-v2-border-soft" />
          <span className="text-v2-muted">
            <strong className="text-white">{status.toolCount}</strong> {t('commandCenter.scannersActive')}
          </span>
          <div className="w-px h-3 bg-v2-border-soft" />
          <span className={status.findingsCount > 0 ? 'text-v2-red' : 'text-success'}>
            <strong>{status.findingsCount}</strong> {t('commandCenter.findings')}
          </span>
          <div className="w-px h-3 bg-v2-border-soft" />
          <span className="text-v2-muted tabular-nums font-mono tracking-widest">
            {status.uptime}
          </span>
        </div>
      </div>

      {/* ═══ TAB BAR ═══ */}
      <div className="shrink-0 flex items-end px-4 border-b border-v2-border-soft overflow-x-auto bg-v2-surface-2 cyber-scrollbar">
        {TABS.map((tab, idx) => {
          const prevTab = idx > 0 ? TABS[idx - 1] : null;
          const showSep = prevTab && prevTab.group !== tab.group;
          const isActive = activeTab === tab.key;
          const isSC = tab.group === 'securecoder';

          return (
            <React.Fragment key={tab.key}>
              {showSep && (
                <div className="w-px mx-1 self-stretch my-2.5 bg-v2-border-soft opacity-50" />
              )}
              <button
                onClick={() => setActiveTab(tab.key)}
                className="relative flex items-center gap-2 px-5 py-2.5 transition-colors group text-[11px] font-bold tracking-widest uppercase"
                style={{
                  color: isActive ? (isSC ? 'var(--v2-red)' : '#FFFFFF') : 'var(--v2-muted)',
                  background: isActive
                    ? isSC
                      ? 'var(--v2-red-soft)'
                      : 'var(--v2-surface)'
                    : 'transparent',
                }}
              >
                {isActive && (
                  <div
                    className="absolute bottom-0 left-0 right-0 h-[2px]"
                    style={{ background: isSC ? 'var(--v2-red)' : 'var(--v2-fg)' }}
                  />
                )}
                <span
                  className="material-symbols-outlined"
                  style={{
                    fontSize: '16px',
                    fontVariationSettings: isActive ? "'FILL' 1" : "'FILL' 0",
                  }}
                >
                  {tab.icon}
                </span>
                <span className="whitespace-nowrap">{t(tab.i18nKey)}</span>
                {tab.hotkey && (
                  <span
                    className="ml-1 text-[9px] w-[16px] h-[16px] flex items-center justify-center border font-mono rounded"
                    style={{
                      borderColor: isActive
                        ? isSC
                          ? 'var(--v2-red-line)'
                          : 'var(--v2-border)'
                        : 'var(--v2-border-soft)',
                      color: isActive
                        ? isSC
                          ? 'var(--v2-red)'
                          : 'var(--v2-fg)'
                        : 'var(--v2-muted)',
                      opacity: isActive ? 1 : 0.6,
                    }}
                  >
                    {tab.hotkey}
                  </span>
                )}
              </button>
            </React.Fragment>
          );
        })}
      </div>

      {/* ═══ CONTENT ═══ */}
      <div className="flex-1 overflow-hidden bg-v2-bg">{renderContent()}</div>
    </div>
  );
};
