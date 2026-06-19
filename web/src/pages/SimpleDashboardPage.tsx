import React, { useState, useMemo, useEffect, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import Markdown from 'react-markdown';
import { useTranslation } from 'react-i18next';
import { useFindings } from '../hooks/useFindings';
import { useMetrics } from '../hooks/useMetrics';
import { useProducts } from '../hooks/useProducts';
import type { Finding, Product } from '../types';

interface SimpleDashboardPageProps {
  onNavigateToChat?: (findingOrPrompt?: Finding | string) => void;
}

/* ── Path Input with Browse ── */
type BrowserEntry = { name: string; is_dir: boolean; path: string };




const displayPath = (p: string) => p.replace(/^\/host/, '~');

const PathInput: React.FC<{ value: string; onChange: (p: string) => void }> = ({ value, onChange }) => {
  const { t } = useTranslation('pages');
  const [browsing, setBrowsing] = useState(false);
  const [entries, setEntries] = useState<BrowserEntry[]>([]);
  const [browsePath, setBrowsePath] = useState('/host');
  const [loading, setLoading] = useState(false);

  const browse = async (path: string) => {
    setLoading(true);
    try {
      const res = await fetch(`/api/browser?path=${encodeURIComponent(path)}`);
      const data = await res.json();
      if (data.ok) {
        setEntries(data.entries?.filter((e: BrowserEntry) => e.is_dir) || []);
        setBrowsePath(data.path || path);
      }
    } catch { /* ignore */ }
    setLoading(false);
  };

  const openBrowser = () => {
    setBrowsing(true);
    browse(value && value !== '/project' ? value : '/host');
  };

  const goUp = () => {
    if (browsePath === '/host' || browsePath === '/') return;
    const parts = browsePath.split('/');
    parts.pop();
    const parent = parts.join('/') || '/';
    browse(parent);
  };

  const selectEntry = (entry: BrowserEntry) => browse(entry.path);

  const confirmBrowse = () => {
    onChange(browsePath);
    setBrowsing(false);
  };

  // Breadcrumb segments
  const breadcrumbs = useMemo(() => {
    const parts = browsePath.split('/').filter(Boolean);
    const result: { label: string; path: string }[] = [];
    let acc = '';
    parts.forEach((p, i) => {
      acc += '/' + p;
      result.push({ label: i === 0 && p === 'host' ? '~' : p, path: acc });
    });
    return result;
  }, [browsePath]);

  return (
    <div className="space-y-2">
      {/* Text input */}
      <div className="flex gap-1.5">
        <div className="relative flex-1">
          <span className="material-symbols-outlined text-[14px] text-[#3f3f46] absolute left-2.5 top-1/2 -translate-y-1/2">folder</span>
          <input
            value={value}
            onChange={e => onChange(e.target.value)}
            placeholder="/host/Desktop/my-project"
            className="w-full bg-surface-bright border border-[rgba(255,255,255,0.06)] rounded-lg pl-8 pr-3 py-2 text-[12px] text-[#f4f4f5] font-mono placeholder:text-[#3f3f46] outline-none focus:border-[rgba(255,255,255,0.12)] transition-colors"
          />
        </div>
        <button
          onClick={openBrowser}
          className="shrink-0 w-8 h-8 flex items-center justify-center rounded-lg border border-[rgba(255,255,255,0.06)] text-[#52525b] hover:text-[#a1a1aa] hover:bg-[rgba(255,255,255,0.03)] transition-colors"
          title="Browse folders"
        >
          <span className="material-symbols-outlined text-[16px]">folder_open</span>
        </button>
      </div>

      {/* Browse panel */}
      {browsing && (
        <div className="border border-[rgba(255,255,255,0.08)] rounded-lg overflow-hidden">
          {/* Header */}
          <div className="flex items-center gap-2 px-3 py-2 border-b border-[rgba(255,255,255,0.06)] bg-[rgba(255,255,255,0.02)]">
            <span className="material-symbols-outlined text-[13px] text-[#3f3f46]">folder_special</span>
            <span className="text-[11px] text-[#52525b]">{t('SimpleDashboardPage.scanRoot', 'SCAN_ROOT')}</span>
            <div className="flex-1" />
            <button onClick={() => browse('/host')} className="text-[10px] text-[#52525b] hover:text-[#a1a1aa] transition-colors">{t('SimpleDashboardPage.root')}</button>
          </div>

          {/* Breadcrumb path */}
          <div className="flex items-center gap-0.5 px-3 py-1.5 border-b border-[rgba(255,255,255,0.04)] overflow-x-auto" style={{ scrollbarWidth: 'none' }}>
            <button onClick={() => browse('/host')} className="text-[11px] text-[#52525b] hover:text-[#a1a1aa] shrink-0">~</button>
            {breadcrumbs.slice(1).map((b, i) => (
              <React.Fragment key={b.path}>
                <span className="text-[10px] text-[#3f3f46] mx-0.5">/</span>
                <button onClick={() => browse(b.path)}
                  className={`text-[11px] shrink-0 transition-colors ${i === breadcrumbs.length - 2 ? 'text-[#a1a1aa]' : 'text-[#52525b] hover:text-[#a1a1aa]'}`}>
                  {b.label}
                </button>
              </React.Fragment>
            ))}
          </div>

          {/* Entries */}
          <div className="max-h-[200px] overflow-y-auto" style={{ scrollbarWidth: 'thin', scrollbarColor: 'rgba(255,255,255,0.06) transparent' }}>
            {/* Go up */}
            {browsePath !== '/host' && browsePath !== '/' && (
              <button onClick={goUp}
                className="w-full flex items-center gap-2 px-3 py-1.5 text-left hover:bg-[rgba(255,255,255,0.03)] transition-colors border-b border-[rgba(255,255,255,0.03)]">
                <span className="material-symbols-outlined text-[13px] text-[#3f3f46]">arrow_upward</span>
                <span className="text-[11px] text-[#52525b]">..</span>
              </button>
            )}
            {loading ? (
              <div className="flex items-center justify-center py-4">
                <div className="w-3 h-3 border border-[#27272a] border-t-[#52525b] rounded-full animate-spin" />
              </div>
            ) : entries.length === 0 ? (
              <div className="py-4 text-center text-[11px] text-[#3f3f46]">{t('SimpleDashboardPage.emptyDirectory')}</div>
            ) : (
              entries.map(e => (
                <button
                  key={e.path}
                  onClick={() => selectEntry(e)}
                  className="w-full flex items-center gap-2 px-3 py-1.5 text-left hover:bg-[rgba(255,255,255,0.03)] transition-colors"
                >
                  <span className="material-symbols-outlined text-[13px] text-[#52525b]">folder</span>
                  <span className="text-[12px] text-[#a1a1aa] truncate">{e.name}</span>
                </button>
              ))
            )}
          </div>

          {/* Actions */}
          <div className="flex items-center gap-2 px-3 py-2 border-t border-[rgba(255,255,255,0.06)] bg-[rgba(255,255,255,0.01)]">
            <span className="text-[10px] text-[#3f3f46] font-mono truncate flex-1">{displayPath(browsePath)}</span>
            <button onClick={() => setBrowsing(false)} className="text-[11px] text-[#52525b] hover:text-[#a1a1aa] px-2 py-1">{t('SimpleDashboardPage.cancel')}</button>
            <button onClick={confirmBrowse} className="text-[11px] text-[#f4f4f5] bg-[#27272a] hover:bg-[#3f3f46] px-3 py-1 rounded transition-colors">
              Select
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

/* ── Scan Panel (right column) ── */
type ScanStatus = { state: 'idle' | 'scanning' | 'done' | 'error'; findings?: number; duration?: string; error?: string };

interface ScanPanelProps {
  onScanComplete?: () => void;
}

const ScanPanel: React.FC<ScanPanelProps> = ({ onScanComplete }) => {
  const { t } = useTranslation('pages');
  const [external,  ] = useState(true);
  const [scanPath, setScanPath] = useState('/host');
  const [projects, setProjects] = useState<BrowserEntry[]>([]);
  const [loadingProjects, setLoadingProjects] = useState(true);
  const [showCustomPath, setShowCustomPath] = useState(false);
  const [showScanners, setShowScanners] = useState(false);
  const [scanStatuses, setScanStatuses] = useState<Record<string, ScanStatus>>({});
  const [activeScans, setActiveScans] = useState(0); // for Scan All progress
  const [totalScans, setTotalScans] = useState(0);
  const [elapsed, setElapsed] = useState(0);
  const [scanningProject, setScanningProject] = useState<string | null>(null);
  const [tools, setTools] = useState({ semgrep: true, gitleaks: true, trivy: true, bandit: true });
  const [toolStatus, setToolStatus] = useState<Record<string, boolean>>({});

  const [currentPath, setCurrentPath] = useState('/host');

  const loadPath = useCallback((path: string) => {
    setLoadingProjects(true);
    fetch(`/api/browser?path=${encodeURIComponent(path)}`)
      .then(r => r.json())
      .then(d => { 
        if (d.ok) {
          setProjects(d.entries?.filter((e: BrowserEntry) => e.is_dir) || []);
        } 
      })
      .catch(() => {})
      .finally(() => setLoadingProjects(false));
  }, []);

  useEffect(() => {
    loadPath(currentPath);
  }, [currentPath, loadPath]);

  useEffect(() => {
    fetch('/api/health').then(r => r.json()).then(d => { if (d.ok && d.tools) setToolStatus(d.tools); }).catch(() => {});
  }, []);

  // Scan phase simulation
  const [scanPhase, setScanPhase] = useState(0);
  const [scanLogs, setScanLogs] = useState<string[]>([]);
  const phases = [
    { name: 'Core', desc: 'AST parsing & pattern matching', icon: 'memory' },
    { name: 'Semgrep', desc: 'SAST rules & taint analysis', icon: 'shield' },
    { name: 'Gitleaks', desc: 'Secrets & credential detection', icon: 'key' },
    { name: 'Trivy', desc: 'CVE & dependency vulnerabilities', icon: 'inventory_2' },
    { name: 'Bandit', desc: 'Python-specific security checks', icon: 'bug_report' },
  ];
  const logMessages = [
    'Indexing source files...', 'Building AST...', 'Running pattern rules...',
    'Checking injection patterns...', 'Scanning for SQL injection...', 'Analyzing auth flows...',
    'Detecting hardcoded secrets...', 'Checking API keys...', 'Scanning .env files...',
    'Resolving dependencies...', 'Checking CVE database...', 'Analyzing lock files...',
    'Scanning Python imports...', 'Checking subprocess calls...', 'Detecting unsafe deserialization...',
    'Analyzing template injection...', 'Checking XSS vectors...', 'Scanning CSRF protections...',
  ];

  // Elapsed timer + phase cycling during scan
  useEffect(() => {
    if (!scanningProject) { setElapsed(0); setScanPhase(0); setScanLogs([]); return; }
    const t = setInterval(() => setElapsed(e => e + 1), 1000);
    const p = setInterval(() => setScanPhase(ph => (ph + 1) % phases.length), 4000);
    const l = setInterval(() => {
      const msg = logMessages[Math.floor(Math.random() * logMessages.length)];
      setScanLogs(prev => [...prev.slice(-4), msg]);
    }, 1200);
    return () => { clearInterval(t); clearInterval(p); clearInterval(l); };
  }, [scanningProject]);

  const runScan = async (path: string): Promise<boolean> => {
    setScanningProject(path);
    setScanStatuses(prev => ({ ...prev, [path]: { state: 'scanning' } }));
    try {
      const res = await fetch('/api/scan', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ path, external }),
      });
      const data = await res.json();
      if (data.ok) {
        setScanStatuses(prev => ({ ...prev, [path]: { state: 'done', findings: data.findings?.length ?? 0, duration: data.duration } }));
        return true;
      } else {
        setScanStatuses(prev => ({ ...prev, [path]: { state: 'error', error: data.error || 'Failed' } }));
        return false;
      }
    } catch {
      setScanStatuses(prev => ({ ...prev, [path]: { state: 'error', error: 'Connection error' } }));
      return false;
    } finally {
      setScanningProject(null);
    }
  };

  const scanAll = async () => {
    setTotalScans(projects.length);
    setActiveScans(0);
    for (let i = 0; i < projects.length; i++) {
      setActiveScans(i + 1);
      await runScan(projects[i].path);
    }
    if (onScanComplete) {
      setTimeout(() => onScanComplete(), 2000);
    } else {
      setTimeout(() => window.location.reload(), 2000);
    }
  };

  const scanOne = async (path: string) => {
    setTotalScans(1);
    setActiveScans(1);
    await runScan(path);
    if (onScanComplete) {
      setTimeout(() => onScanComplete(), 1500);
    } else {
      setTimeout(() => window.location.reload(), 1500);
    }
  };

  const isAnyScanRunning = !!scanningProject;
  const completedCount = Object.values(scanStatuses).filter(s => s.state === 'done').length;
  const totalFindings = Object.values(scanStatuses).filter(s => s.state === 'done').reduce((sum, s) => sum + (s.findings ?? 0), 0);

  const toolList = [
    { key: 'semgrep', label: 'Semgrep', desc: 'SAST analysis' },
    { key: 'gitleaks', label: 'Gitleaks', desc: 'Secret detection' },
    { key: 'trivy', label: 'Trivy', desc: 'Dependency scan' },
    { key: 'bandit', label: 'Bandit', desc: 'Python security' },
  ];

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="flex items-center gap-2 px-4 py-3 border-b border-[rgba(255,255,255,0.06)] bg-[rgba(255,255,255,0.005)]">
        {currentPath !== '/host' && (
          <button 
            onClick={() => {
              const parts = currentPath.split('/');
              parts.pop();
              const parent = parts.join('/') || '/host';
              setCurrentPath(parent === '/' ? '/host' : parent);
            }}
            className="w-5 h-5 flex items-center justify-center rounded border border-[rgba(255,255,255,0.06)] text-[#71717a] hover:text-[#a1a1aa] hover:bg-[rgba(255,255,255,0.02)] transition-colors cursor-pointer mr-0.5 shrink-0"
            title="Go back"
          >
            <span className="material-symbols-outlined text-[13px]">arrow_back</span>
          </button>
        )}
        <div className="flex-1 min-w-0">
          <span className="text-[12px] font-bold text-[#f4f4f5] tracking-wide block truncate uppercase">
            {currentPath === '/host' ? t('SimpleDashboardPage.projects') : currentPath.replace(/^\/host\/?/, '') || 'Projects'}
          </span>
        </div>
        <span className="text-[10px] text-[#3f3f46] tabular-nums shrink-0 font-mono">
          {projects.length} DIRS
        </span>
      </div>

      {/* Scan progress panel */}
      {isAnyScanRunning && (
        <div className="border-b border-[rgba(255,255,255,0.06)] bg-[rgba(255,255,255,0.02)]">
          {/* Current scanner phase */}
          <div className="px-4 py-2.5">
            <div className="flex items-center justify-between mb-1">
              <div className="flex items-center gap-1.5">
                <div className="w-4 h-4 border-2 border-[#3f3f46] border-t-[#22c55e] rounded-full animate-spin" />
                <span className="text-[11px] text-[#f4f4f5] font-medium">{scanningProject?.split('/').pop()}</span>
              </div>
              <span className="text-[10px] text-[#52525b] tabular-nums">{t('SimpleDashboardPage.elapsed', { seconds: elapsed })}</span>
            </div>
            {/* Phase indicator */}
            <div className="flex items-center gap-1.5 mt-1.5">
              <span className="material-symbols-outlined text-[12px] text-[#22c55e]">{phases[scanPhase].icon}</span>
              <span className="text-[10px] text-[#a1a1aa] font-medium">{phases[scanPhase].name}</span>
              <span className="text-[10px] text-[#3f3f46]">— {t(`SimpleDashboardPage.phases.${scanPhase}.desc`, { defaultValue: phases[scanPhase].desc })}</span>
            </div>
            {/* Phase progress dots */}
            <div className="flex gap-1 mt-2">
              {phases.map((ph, i) => (
                <div key={ph.name} className={`flex-1 h-1 rounded-full transition-all duration-500 ${
                  i < scanPhase ? 'bg-[#22c55e]' : i === scanPhase ? 'bg-[#22c55e] animate-pulse' : 'bg-[#18181b]'
                }`} />
              ))}
            </div>
            {/* Batch progress */}
            {totalScans > 1 && (
              <div className="flex items-center justify-between mt-2">
                <span className="text-[9px] text-[#3f3f46]">{t('SimpleDashboardPage.scanProgress', { active: activeScans, total: totalScans })}</span>
                <div className="w-20 h-0.5 bg-[#18181b] rounded-full overflow-hidden">
                  <div className="h-full bg-[#52525b] rounded-full transition-all" style={{ width: `${(activeScans / totalScans) * 100}%` }} />
                </div>
              </div>
            )}
          </div>
          {/* Mini log */}
          <div className="px-4 py-1.5 border-t border-[rgba(255,255,255,0.04)] bg-[rgba(0,0,0,0.15)] font-mono">
            {scanLogs.slice(-3).map((log, i) => (
              <div key={i} className={`text-[9px] leading-relaxed transition-opacity duration-300 ${i === scanLogs.slice(-3).length - 1 ? 'text-[#52525b]' : 'text-[#27272a]'}`}>
                <span className="text-[#3f3f46] mr-1">$</span>{log}
              </div>
            ))}
          </div>
        </div>
      )}

      <div className="flex-1 overflow-y-auto" style={{ scrollbarWidth: 'thin', scrollbarColor: 'rgba(255,255,255,0.06) transparent' }}>
        {/* Project list */}
        <div className="p-1.5 space-y-0.5">
          {loadingProjects ? (
            <div className="flex items-center justify-center py-8">
              <div className="w-3 h-3 border border-[#27272a] border-t-[#52525b] rounded-full animate-spin" />
            </div>
          ) : projects.length === 0 ? (
            <div className="py-6 text-center text-[11px] text-[#3f3f46]">{t('SimpleDashboardPage.noProjects')}</div>
          ) : (
            projects.map(p => {
              const status = scanStatuses[p.path];
              const isActive = scanningProject === p.path;
              const isDimmed = isAnyScanRunning && !isActive;
              return (
                <div key={p.path}
                  className={`flex items-center gap-2.5 px-3 py-2 rounded-lg transition-all duration-300 ${
                    isActive ? 'bg-[rgba(34,197,94,0.06)] border-l-2 border-l-[#22c55e] border-y border-r border-[rgba(34,197,94,0.1)]'
                    : status?.state === 'done' ? 'border-l-2 border-l-[#22c55e]/40 border-y border-r border-transparent'
                    : status?.state === 'error' ? 'border-l-2 border-l-[#ef4444]/40 border-y border-r border-transparent'
                    : 'border border-transparent hover:bg-[rgba(255,255,255,0.02)]'
                  } ${isDimmed ? 'opacity-30' : ''}`}
                >
                  <div 
                    onClick={() => !isAnyScanRunning && setCurrentPath(p.path)}
                    className="flex-1 flex items-center gap-2.5 min-w-0 cursor-pointer group"
                  >
                    {/* Icon */}
                    {isActive ? (
                      <div className="w-4 h-4 border-2 border-[#27272a] border-t-[#22c55e] rounded-full animate-spin shrink-0" />
                    ) : status?.state === 'done' ? (
                      <span className="material-symbols-outlined text-[16px] text-[#22c55e] shrink-0">check_circle</span>
                    ) : status?.state === 'error' ? (
                      <span className="material-symbols-outlined text-[16px] text-[#ef4444] shrink-0">error</span>
                    ) : (
                      <span className="material-symbols-outlined text-[15px] text-[#52525b] group-hover:text-[var(--accent-color)] transition-colors shrink-0">folder</span>
                    )}
                    <div className="flex-1 min-w-0">
                      <div className={`text-[12px] truncate group-hover:text-[#f4f4f5] transition-colors ${isActive ? 'text-[#f4f4f5] font-medium' : 'text-[#a1a1aa]'}`}>{p.name}</div>
                      {status?.state === 'done' && (
                        <div className="text-[10px] text-[#52525b] mt-0.5">
                          <span className="text-[#22c55e]">{status.findings}</span> {t('issues')} · {status.duration}
                        </div>
                      )}
                      {status?.state === 'error' && (
                        <div className="text-[10px] text-[#ef4444] mt-0.5 truncate">{status.error}</div>
                      )}
                    </div>
                  </div>
                  
                  {!isActive && (
                    <button
                      onClick={() => scanOne(p.path)}
                      disabled={isAnyScanRunning}
                      className="shrink-0 text-[11px] font-medium text-[#f4f4f5] bg-[#27272a] hover:bg-[#3f3f46] border border-[#3f3f46] hover:border-[#52525b] rounded-md px-3 py-1 transition-all disabled:opacity-20 cursor-pointer"
                    >
                      {status?.state === 'done' ? 'Rescan' : 'Scan'}
                    </button>
                  )}
                </div>
              );
            })
          )}
        </div>

        {/* Custom path */}
        <div className="px-4 pt-2 pb-1">
          <button onClick={() => setShowCustomPath(!showCustomPath)}
            className="text-[10px] text-[#3f3f46] hover:text-[#52525b] transition-colors flex items-center gap-1">
            <span className="material-symbols-outlined text-[12px]">{showCustomPath ? 'expand_less' : 'expand_more'}</span>
            Custom path
          </button>
          <AnimatePresence initial={false}>
            {showCustomPath && (
              <motion.div
                key="custom-path"
                initial={{ height: 0, opacity: 0 }}
                animate={{ height: "auto", opacity: 1 }}
                exit={{ height: 0, opacity: 0 }}
                transition={{ duration: 0.2 }}
                className="overflow-hidden"
              >
                <div className="mt-2"><PathInput value={scanPath} onChange={setScanPath} /></div>
              </motion.div>
            )}
          </AnimatePresence>
        </div>

        {/* Scanners */}
        <div className="px-4 pt-2 pb-2">
          <button onClick={() => setShowScanners(!showScanners)}
            className="w-full flex items-center justify-between text-[10px] text-[#3f3f46] hover:text-[#52525b] transition-colors">
            <span className="flex items-center gap-1">
              <span className="material-symbols-outlined text-[12px]">{showScanners ? 'expand_less' : 'expand_more'}</span>
              Scanners
            </span>
            <span>{Object.values(tools).filter(Boolean).length}/{Object.keys(tools).length} {t('SimpleDashboardPage.active')}</span>
          </button>
          <AnimatePresence initial={false}>
            {showScanners && (
              <motion.div
                key="scanners"
                initial={{ height: 0, opacity: 0 }}
                animate={{ height: "auto", opacity: 1 }}
                exit={{ height: 0, opacity: 0 }}
                transition={{ duration: 0.2 }}
                className="overflow-hidden"
              >
                <div className="mt-2 space-y-0.5">
                  {toolList.map(t => {
                    const installed = toolStatus[t.key];
                    const enabled = (tools as any)[t.key];
                    return (
                      <div key={t.key} className="flex items-center gap-2 px-2 py-1.5 rounded hover:bg-[rgba(255,255,255,0.02)]">
                        <button onClick={() => setTools(prev => ({ ...prev, [t.key]: !enabled }))}
                          className={`w-3.5 h-3.5 rounded border flex items-center justify-center transition-colors ${enabled ? 'bg-[#f4f4f5] border-[#f4f4f5]' : 'border-[#3f3f46]'}`}>
                          {enabled && <span className="material-symbols-outlined text-[10px] text-[#0a0a0b]">check</span>}
                        </button>
                        <span className="text-[11px] text-[#a1a1aa] flex-1">{t.label}</span>
                        {installed !== undefined && <span className={`w-1.5 h-1.5 rounded-full ${installed ? 'bg-[#22c55e]' : 'bg-[#ef4444]'}`} />}
                      </div>
                    );
                  })}
                </div>
              </motion.div>
            )}
          </AnimatePresence>
        </div>

        {/* Scan summary */}
        {completedCount > 0 && !isAnyScanRunning && (
          <div className="mx-3 mb-2 px-3 py-2 rounded-lg bg-[rgba(34,197,94,0.06)] border border-[rgba(34,197,94,0.12)]">
            <div className="text-[11px] text-[#22c55e] font-medium">{completedCount} project{completedCount > 1 ? 's' : ''} scanned</div>
            <div className="text-[10px] text-[#52525b] mt-0.5">{totalFindings} total {t('issues')} found</div>
          </div>
        )}
      </div>

      {/* Action button */}
      <div className="p-3 border-t border-[rgba(255,255,255,0.06)]">
        <button
          onClick={() => showCustomPath ? scanOne(scanPath) : scanAll()}
          disabled={isAnyScanRunning}
          className="w-full flex items-center justify-center gap-2 px-4 py-2.5 rounded-lg bg-[#f4f4f5] text-[#0a0a0b] text-[13px] font-medium hover:bg-[#e4e4e7] disabled:opacity-40 transition-all"
        >
          {isAnyScanRunning ? (
            <>
              <div className="w-3.5 h-3.5 border-2 border-[#0a0a0b]/30 border-t-[#0a0a0b] rounded-full animate-spin" />
              Scanning {scanningProject?.split('/').pop()}...
            </>
          ) : (
            <>
              <span className="material-symbols-outlined text-[16px]">play_arrow</span>
              {showCustomPath ? 'Run Scan' : `Scan All (${projects.length})`}
            </>
          )}
        </button>
      </div>
    </div>
  );
};

/* ── Main Dashboard ── */
type GroupBy = 'none' | 'severity' | 'title' | 'file' | 'scanner' | 'product';
type SortBy = 'severity' | 'title' | 'file';
const PAGE_SIZE = 25;

/* ── Inline AI IDE Prompts ── */
const INLINE_PROMPT_CATEGORIES = [
  {
    id: 'security-scan',
    labelKey: 'aiPrompts.categories.securityScan',
    icon: 'shield',
    prompts: [
      { id: 'scan-project', titleKey: 'aiPrompts.prompts.scanProject.title', descKey: 'aiPrompts.prompts.scanProject.desc', promptKey: 'aiPrompts.prompts.scanProject.prompt', tags: ['Cursor', 'Copilot', 'Windsurf'], features: ['CVSS 3.1', 'CWE Mapping', '7 Analysis Categories', 'Structured Output'] },
      { id: 'scan-file', titleKey: 'aiPrompts.prompts.scanFile.title', descKey: 'aiPrompts.prompts.scanFile.desc', promptKey: 'aiPrompts.prompts.scanFile.prompt', tags: ['Cursor', 'Copilot'], features: ['Data Flow Tracing', 'Trust Boundary Analysis', 'Confidence Level', 'Exploit Scenarios'] },
      { id: 'scan-deps', titleKey: 'aiPrompts.prompts.scanDeps.title', descKey: 'aiPrompts.prompts.scanDeps.desc', promptKey: 'aiPrompts.prompts.scanDeps.prompt', tags: ['Cursor', 'Copilot', 'Windsurf'], features: ['NVD / OSV / GHSA', 'Typosquatting Check', 'License Audit', 'Lockfile Integrity'] },
    ],
  },
  {
    id: 'remediation',
    labelKey: 'aiPrompts.categories.remediation',
    icon: 'auto_fix_high',
    prompts: [
      { id: 'fix-vuln', titleKey: 'aiPrompts.prompts.fixVuln.title', descKey: 'aiPrompts.prompts.fixVuln.desc', promptKey: 'aiPrompts.prompts.fixVuln.prompt', tags: ['Cursor', 'Copilot'], features: ['SecureCoder Workflow', 'Root Cause Analysis', 'PoC Verification', 'Regression Test'] },
      { id: 'fix-batch', titleKey: 'aiPrompts.prompts.fixBatch.title', descKey: 'aiPrompts.prompts.fixBatch.desc', promptKey: 'aiPrompts.prompts.fixBatch.prompt', tags: ['Cursor', 'Windsurf'], features: ['4-Phase Pipeline', 'Threat Model First', 'Implementation Plan', 'Verification Pass'] },
      { id: 'secure-refactor', titleKey: 'aiPrompts.prompts.secureRefactor.title', descKey: 'aiPrompts.prompts.secureRefactor.desc', promptKey: 'aiPrompts.prompts.secureRefactor.prompt', tags: ['Cursor', 'Copilot', 'Windsurf'], features: ['Defense in Depth', 'CSP / CSRF / HSTS', 'Rate Limiting', 'Constant-Time Ops'] },
    ],
  },
  {
    id: 'analysis',
    labelKey: 'aiPrompts.categories.analysis',
    icon: 'analytics',
    prompts: [
      { id: 'threat-model', titleKey: 'aiPrompts.prompts.threatModel.title', descKey: 'aiPrompts.prompts.threatModel.desc', promptKey: 'aiPrompts.prompts.threatModel.prompt', tags: ['Cursor', 'Copilot'], features: ['STRIDE Methodology', 'DFD Diagram', 'Attack Trees', 'Risk Scoring'] },
      { id: 'code-review', titleKey: 'aiPrompts.prompts.codeReview.title', descKey: 'aiPrompts.prompts.codeReview.desc', promptKey: 'aiPrompts.prompts.codeReview.prompt', tags: ['Cursor', 'Copilot', 'Windsurf'], features: ['Pass/Fail Checklist', '5 Security Domains', 'GDPR / CCPA', 'Code Quality'] },
      { id: 'attack-surface', titleKey: 'aiPrompts.prompts.attackSurface.title', descKey: 'aiPrompts.prompts.attackSurface.desc', promptKey: 'aiPrompts.prompts.attackSurface.prompt', tags: ['Cursor'], features: ['4 Surface Categories', 'Risk Assessment Table', 'ASCII Surface Map', 'Top 5 Targets'] },
    ],
  },
  {
    id: 'compliance',
    labelKey: 'aiPrompts.categories.compliance',
    icon: 'verified_user',
    prompts: [
      { id: 'owasp-check', titleKey: 'aiPrompts.prompts.owaspCheck.title', descKey: 'aiPrompts.prompts.owaspCheck.desc', promptKey: 'aiPrompts.prompts.owaspCheck.prompt', tags: ['Cursor', 'Copilot'], features: ['OWASP 2025', 'Per-Category Assessment', 'Compliance Scorecard', 'A01-A10 Coverage'] },
      { id: 'write-tests', titleKey: 'aiPrompts.prompts.writeTests.title', descKey: 'aiPrompts.prompts.writeTests.desc', promptKey: 'aiPrompts.prompts.writeTests.prompt', tags: ['Cursor', 'Copilot', 'Windsurf'], features: ['5 Test Categories', 'Injection Payloads', 'Auth & IDOR Tests', 'Race Conditions'] },
    ],
  },
];

const SecureCoderPanel: React.FC<{ activeProducts: any[]; findings: any[] }> = ({ activeProducts, findings }) => {
  const { t } = useTranslation('pages');
  const [expandedCat, setExpandedCat] = useState<string | null>(null);
  
  // Agent Runway state
  const [runwayOpen, setRunwayOpen] = useState(false);
  const [runwayStep, setRunwayStep] = useState(0); // 0: Select project, 1: Threat Model, 2: Security Plan, 3: Remediation, 4: Scanner & PoC, 5: Report, 6: Complete
  const [runwayProject, setRunwayProject] = useState<any | null>(null);
  const [runwayLoading, setRunwayLoading] = useState(false);
  const [runwayError, setRunwayError] = useState('');
  const [runwayAutoMode, setRunwayAutoMode] = useState(false);
  const [runwayAutoPhase, setRunwayAutoPhase] = useState(''); // human-readable current auto phase
  
  const [runwayThreatModel, setRunwayThreatModel] = useState('');
  const [runwaySecurityPlan, setRunwaySecurityPlan] = useState('');
  const [runwayRemediation, setRunwayRemediation] = useState('');
  const [runwayPoC, setRunwayPoC] = useState('');
  const [runwayAuditReport, setRunwayAuditReport] = useState('');
  const [runwayScanCountBefore, setRunwayScanCountBefore] = useState(0);
  const [runwayScanCountAfter, setRunwayScanCountAfter] = useState(0);
  const [runwaySessionId, setRunwaySessionId] = useState<number | null>(null);
  const [runwayExporting, setRunwayExporting] = useState(false);

  // Ignore state
  const [ignoredFindings, setIgnoredFindings] = useState<any[]>([]);
  const [loadingIgnored, setLoadingIgnored] = useState(false);
  
  // Scan state
  const [scanPath, setScanPath] = useState('');
  const [scanResult, setScanResult] = useState<any>(null);
  const [scanning, setScanning] = useState(false);
  const [quickScanType, setQuickScanType] = useState<'file' | 'dir'>('file');
  
  // Dep state
  const [depRegistry, setDepRegistry] = useState('npm');
  const [depPackage, setDepPackage] = useState('');
  const [depResult, setDepResult] = useState<any>(null);
  const [depScanning, setDepScanning] = useState(false);

  // SecureCoder Configuration states
  const [configEnabled, setConfigEnabled] = useState(true);
  const [configScannerBackend, setConfigScannerBackend] = useState('semgrep');
  const [configRuleSet, setConfigRuleSet] = useState('fast');
  const [configAutostartFixes, setConfigAutostartFixes] = useState(true);
  const [configIgnoreMode, setConfigIgnoreMode] = useState('workspace');
  const [configDebug, setConfigDebug] = useState(false);
  const [configLoading, setConfigLoading] = useState(false);
  const [configSaving, setConfigSaving] = useState(false);
  const [configError, setConfigError] = useState('');
  const [configSuccess, setConfigSuccess] = useState(false);

  // Onboarding Wizard states
  const [onboardingOpen, setOnboardingOpen] = useState(false);
  const [onboardingStep, setOnboardingStep] = useState(0);
  const [wizAgreementChecked, setWizAgreementChecked] = useState(false);
  const [onboardingIgnoreContent, setOnboardingIgnoreContent] = useState(
    '# Default glob patterns\n*test.*\n*_test.*\n**/*_test.*\n**/test/**\nnode_modules/\nvendor/\n.git/'
  );

  // Wiz CLI Authentication states
  const [wizStatus, setWizStatus] = useState<any>({ authenticated: false });
  const [wizAuthLoading, setWizAuthLoading] = useState(false);
  const [wizLoginSession, setWizLoginSession] = useState<any>(null);
  const [pollingInterval, setPollingInterval] = useState<any>(null);

  // Ignore File Editor states
  const [ignoreEditorOpen, setIgnoreEditorOpen] = useState(false);
  const [ignoreEditorContent, setIgnoreEditorContent] = useState('');
  const [ignoreEditorSaving, setIgnoreEditorSaving] = useState(false);

  const fetchIgnored = useCallback(async () => {
    setLoadingIgnored(true);
    try {
      const res = await fetch('/api/securecoder/ignored');
      const data = await res.json();
      if (data.entries) setIgnoredFindings(data.entries);
    } catch (e) {
      console.error(e);
    }
    setLoadingIgnored(false);
  }, []);

  const fetchConfig = useCallback(async () => {
    setConfigLoading(true);
    setConfigError('');
    try {
      const res = await fetch('/api/securecoder/config');
      const data = await res.json();
      setConfigEnabled(data.enabled ?? true);
      setConfigScannerBackend(data.scannerBackend ?? 'semgrep');
      setConfigRuleSet(data.ruleSet ?? 'fast');
      setConfigAutostartFixes(data.autostartFixes ?? true);
      setConfigIgnoreMode(data.ignoreMode ?? 'workspace');
      setConfigDebug(data.debug ?? false);
    } catch (e) {
      console.error(e);
      setConfigError('Failed to load configuration.');
    } finally {
      setConfigLoading(false);
    }
  }, []);

  const handleSaveConfig = async (overrideSettings?: any) => {
    setConfigSaving(true);
    setConfigError('');
    setConfigSuccess(false);
    try {
      const res = await fetch('/api/securecoder/config', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          enabled: overrideSettings?.enabled ?? configEnabled,
          scannerBackend: overrideSettings?.scannerBackend ?? configScannerBackend,
          ruleSet: overrideSettings?.ruleSet ?? configRuleSet,
          autostartFixes: overrideSettings?.autostartFixes ?? configAutostartFixes,
          ignoreMode: overrideSettings?.ignoreMode ?? configIgnoreMode,
          debug: overrideSettings?.debug ?? configDebug
        })
      });
      const data = await res.json();
      if (data.ok) {
        setConfigSuccess(true);
        setTimeout(() => setConfigSuccess(false), 3000);
      } else {
        setConfigError(data.error || 'Failed to save configuration.');
      }
    } catch (e) {
      console.error(e);
      setConfigError('Network error occurred.');
    } finally {
      setConfigSaving(false);
    }
  };

  const fetchWizStatus = useCallback(async () => {
    setWizAuthLoading(true);
    try {
      const res = await fetch('/api/securecoder/wiz/status');
      const data = await res.json();
      setWizStatus(data);
    } catch (e) {
      console.error(e);
    } finally {
      setWizAuthLoading(false);
    }
  }, []);

  const handleWizStartLogin = async () => {
    try {
      const res = await fetch('/api/securecoder/wiz/login', { method: 'POST' });
      const data = await res.json();
      setWizLoginSession(data);

      if (pollingInterval) clearInterval(pollingInterval);
      const interval = setInterval(async () => {
        try {
          const pollRes = await fetch('/api/securecoder/wiz/login/poll');
          const pollData = await pollRes.json();
          setWizLoginSession(pollData);
          if (pollData.completed || pollData.status === 'success' || pollData.status === 'failed') {
            clearInterval(interval);
            fetchWizStatus();
          }
        } catch (e) {
          console.error(e);
          clearInterval(interval);
        }
      }, 2000);
      setPollingInterval(interval);
    } catch (e) {
      console.error(e);
    }
  };

  const handleWizLogout = async () => {
    try {
      await fetch('/api/securecoder/wiz/logout', { method: 'POST' });
      setWizLoginSession(null);
      fetchWizStatus();
    } catch (e) {
      console.error(e);
    }
  };

  const fetchIgnoreFile = async () => {
    try {
      const res = await fetch('/api/securecoder/ignore-file');
      const data = await res.json();
      setIgnoreEditorContent(data.content || '');
    } catch (e) {
      console.error(e);
    }
  };

  const handleSaveIgnoreFile = async (contentToSave: string) => {
    setIgnoreEditorSaving(true);
    try {
      await fetch('/api/securecoder/ignore-file', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ content: contentToSave })
      });
    } catch (e) {
      console.error(e);
    } finally {
      setIgnoreEditorSaving(false);
    }
  };

  const handleClearAllIgnored = async () => {
    try {
      const res = await fetch('/api/securecoder/ignored', { method: 'DELETE' });
      const data = await res.json();
      if (data.ok) {
        setIgnoredFindings([]);
      }
    } catch (e) {
      console.error(e);
    }
  };

  const handleScan = async () => {
    if (!scanPath) return;
    setScanning(true);
    try {
      const endpoint = quickScanType === 'file' ? '/api/securecoder/scan' : '/api/securecoder/scan-directory';
      const body = quickScanType === 'file' 
        ? { filePath: scanPath }
        : { path: scanPath, external: true };

      const res = await fetch(endpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body)
      });
      const data = await res.json();
      setScanResult(data.findings || []);
    } catch (e) {
      console.error(e);
    }
    setScanning(false);
  };

  const handleDepScan = async () => {
    if (!depPackage) return;
    setDepScanning(true);
    try {
      let pkgName = depPackage.trim();
      let pkgVersion = '';

      if (pkgName.includes('@')) {
        const parts = pkgName.split('@');
        if (pkgName.startsWith('@')) {
          pkgName = '@' + parts[1];
          pkgVersion = parts[2] || '';
        } else {
          pkgName = parts[0];
          pkgVersion = parts[1] || '';
        }
      }

      const res = await fetch('/api/securecoder/dependency/scan', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ registry: depRegistry, packages: [{ package: pkgName, version: pkgVersion }] })
      });
      const data = await res.json();
      setDepResult(data.unsafeDependencies || []);
    } catch (e) {
      console.error(e);
    }
    setDepScanning(false);
  };

  useEffect(() => {
    if (expandedCat === 'ignore') fetchIgnored();
    if (expandedCat === 'config') fetchConfig();
  }, [expandedCat, fetchIgnored, fetchConfig]);

  useEffect(() => {
    if (configScannerBackend === 'wiz') {
      fetchWizStatus();
    }
  }, [configScannerBackend, fetchWizStatus]);

  useEffect(() => {
    return () => {
      if (pollingInterval) clearInterval(pollingInterval);
    };
  }, [pollingInterval]);

  // --- Runway DB persistence helpers ---
  const saveRunwayToDB = useCallback(async (
    sessionId: number,
    step: number,
    data: {
      status?: string;
      auto_mode?: boolean;
      threat_model?: string;
      security_plan?: string;
      remediation?: string;
      poc?: string;
      audit_report?: string;
      scan_count_before?: number;
      scan_count_after?: number;
      error_message?: string;
      product_id?: number;
    }
  ) => {
    try {
      await fetch(`/api/runway/${sessionId}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ current_step: step, ...data })
      });
      if (data.status === 'completed') {
        fetch(`/api/runway/export/${sessionId}`, { method: 'POST' })
          .then(res => res.json())
          .then(resData => {
            if (resData.ok) {
              console.log('Runway report auto-saved to project directory:', resData.saved_to);
            } else {
              console.error('Failed to auto-save runway report:', resData.error);
            }
          })
          .catch(e => console.error('Error auto-exporting report:', e));
      }
    } catch (e) {
      console.error('Failed to save runway session:', e);
    }
  }, []);

  const restoreRunwayFromSession = useCallback((session: any) => {
    if (!session) return;
    setRunwaySessionId(session.id);
    setRunwayStep(session.current_step || 0);
    setRunwayAutoMode(session.auto_mode || false);
    setRunwayThreatModel(session.threat_model || '');
    setRunwaySecurityPlan(session.security_plan || '');
    setRunwayRemediation(session.remediation || '');
    setRunwayPoC(session.poc || '');
    setRunwayAuditReport(session.audit_report || '');
    setRunwayScanCountBefore(session.scan_count_before || 0);
    setRunwayScanCountAfter(session.scan_count_after || 0);
    if (session.error_message) setRunwayError(session.error_message);
    // Restore project from activeProducts
    const proj = activeProducts.find(p => p.id === session.product_id);
    if (proj) {
      setRunwayProject(proj);
      setRunwayOpen(true);
    }
  }, [activeProducts]);

  // Restore active runway session from DB on mount
  useEffect(() => {
    if (activeProducts.length === 0) return;
    let cancelled = false;
    (async () => {
      // Check each product for an active session
      for (const prod of activeProducts) {
        try {
          const res = await fetch(`/api/runway?product_id=${prod.id}`);
          const data = await res.json();
          if (!cancelled && data.ok && data.session && data.session.status === 'in_progress' && data.session.current_step > 0) {
            restoreRunwayFromSession(data.session);
            break;
          }
        } catch (e) { /* ignore */ }
      }
    })();
    return () => { cancelled = true; };
  }, [activeProducts, restoreRunwayFromSession]);

  // Runway handlers
  const handleInitRunway = async () => {
    if (!runwayProject) return;
    const project = runwayProject; // capture locally to avoid stale closure
    const projectFindings = findings.filter(f => f.product_id === project.id && f.status !== 'triage');
    const scanCountBefore = projectFindings.length;
    setRunwayScanCountBefore(scanCountBefore);
    setRunwayStep(1);

    // Create DB session
    let sessionId = runwaySessionId;
    if (!sessionId) {
      try {
        const createRes = await fetch('/api/runway', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ product_id: project.id, auto_mode: false })
        });
        const createData = await createRes.json();
        if (createData.ok && createData.session) {
          sessionId = createData.session.id;
          setRunwaySessionId(sessionId);
          await saveRunwayToDB(createData.session.id, 1, { scan_count_before: scanCountBefore, product_id: project.id, auto_mode: false });
        }
      } catch (e) {
        console.error('Failed to create runway session in DB:', e);
      }
    }

    // Auto-run full pipeline inline using local variables
    setRunwayLoading(true);
    setRunwayError('');
    try {
      // Step 1: Threat Model
      const threatModel = await runThreatModelStep(project, findings);
      setRunwayThreatModel(threatModel);
      setRunwayStep(2);
      if (sessionId) await saveRunwayToDB(sessionId, 2, { threat_model: threatModel, product_id: project.id });

      // Step 2: Security Plan
      const plan = await runSecurityPlanStep(project, threatModel);
      setRunwaySecurityPlan(plan);
      setRunwayStep(3);
      if (sessionId) await saveRunwayToDB(sessionId, 3, { security_plan: plan, product_id: project.id });

      // Step 3: Remediation
      const remediation = await runRemediationStep(project, plan);
      setRunwayRemediation(remediation);
      setRunwayStep(4);
      if (sessionId) await saveRunwayToDB(sessionId, 4, { remediation, product_id: project.id });

      // Step 4: Scan & PoC
      const { poc, scanCountAfter } = await runScanAndPoCStep(project, remediation);
      setRunwayPoC(poc);
      setRunwayScanCountAfter(scanCountAfter);
      setRunwayStep(5);
      if (sessionId) await saveRunwayToDB(sessionId, 5, { poc, scan_count_after: scanCountAfter, product_id: project.id });

      // Step 5: Audit Report
      const report = await runReportStep(project, threatModel, remediation, poc);
      setRunwayAuditReport(report);
      setRunwayStep(6);
      if (sessionId) await saveRunwayToDB(sessionId, 6, { audit_report: report, product_id: project.id });

      // Step 6: Complete
      await runCompleteStep(scanCountBefore, scanCountAfter);
      setRunwayStep(7);
      if (sessionId) await saveRunwayToDB(sessionId, 7, { status: 'completed', product_id: project.id });
    } catch (e: any) {
      setRunwayError(e.message || 'Pipeline error.');
    } finally {
      setRunwayLoading(false);
    }
  };


  // --- Individual step handlers (used by both manual and auto mode) ---

  const runThreatModelStep = async (project: any, findingsArr: any[]): Promise<string> => {
    const projectFindings = findingsArr.filter((f: any) => f.product_id === project.id && f.status !== 'triage');
    const findingsText = projectFindings.map((f: any) => `- [${f.severity?.toUpperCase()}] ${f.title} in ${f.file_path || 'N/A'}:${f.line_number || 'N/A'}`).join('\n');
    const prompt = `You are a threat modeling expert following the determine_threat_model methodology.
Respond in English regardless of the programming language or comments in the source code.
Analyze the project: "${project.name}".
Findings Context:\n${findingsText}\n\nBuild a comprehensive threat model: Identify components, trust boundaries, and sensitive data paths. Provide a STRIDE analysis table and prioritized mitigations in markdown.`;
    const res = await fetch('/api/chat', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ messages: [{ role: 'user', content: prompt }] })
    });
    const data = await res.json();
    if (!data.ok) throw new Error(data.error || 'Failed to generate threat model.');
    return data.content || 'No threat model generated.';
  };

  const runSecurityPlanStep = async (project: any, threatModel: string): Promise<string> => {
    const prompt = `You are a security architect. Respond in English regardless of the programming language or comments in the source code.
Create a security implementation plan for project "${project.name}" based on the following STRIDE Threat Model:\n\n${threatModel}\n\nOutline high-level fix priorities and specific security verification tests/checkpoints in markdown.`;
    const res = await fetch('/api/chat', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ messages: [{ role: 'user', content: prompt }] })
    });
    const data = await res.json();
    if (!data.ok) throw new Error(data.error || 'Failed to generate plan.');
    return data.content || 'No security plan generated.';
  };

  const runRemediationStep = async (project: any, securityPlan: string): Promise<string> => {
    const prompt = `Respond in English regardless of the programming language or comments in the source code.
Generate a targeted security patch (before/after code diffs) to fix the security vulnerabilities for project "${project.name}" based on this security plan:\n\n${securityPlan}`;
    const res = await fetch('/api/chat', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ messages: [{ role: 'user', content: prompt }] })
    });
    const data = await res.json();
    if (!data.ok) throw new Error(data.error || 'Failed to generate fixes.');
    return data.content || 'No fixes generated.';
  };

  const runScanAndPoCStep = async (project: any, remediation: string): Promise<{ poc: string; scanCountAfter: number }> => {
    const path = project.repo_url || '/host';
    const scanRes = await fetch('/api/securecoder/scan-directory', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ path, external: true })
    });
    const scanData = await scanRes.json();
    const newCount = scanData.findings ? scanData.findings.length : 0;

    const prompt = `Respond in English regardless of the programming language or comments in the source code.
Generate a Proof-of-Concept (PoC) exploit scenario and verification analysis.
Remediation Diffs:\n${remediation}\n\nDescribe how the exploit works on the unpatched code, and demonstrate why it is now blocked in the patched state in markdown.`;
    const res = await fetch('/api/chat', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ messages: [{ role: 'user', content: prompt }] })
    });
    const data = await res.json();
    if (!data.ok) throw new Error(data.error || 'Failed to generate PoC.');
    return { poc: data.content || 'No PoC generated.', scanCountAfter: newCount };
  };

  const runReportStep = async (project: any, threatModel: string, remediation: string, poc: string): Promise<string> => {
    const prompt = `Respond in English regardless of the programming language or comments in the source code.
Compile a professional Security Audit Report in CS-XXX-NNN markdown format for project "${project.name}".
Threat Model:\n${threatModel}\n\nRemediation:\n${remediation}\n\nPoC Verification:\n${poc}`;
    const res = await fetch('/api/chat', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ messages: [{ role: 'user', content: prompt }] })
    });
    const data = await res.json();
    if (!data.ok) throw new Error(data.error || 'Failed to compile report.');
    return data.content || 'No report generated.';
  };

  const runCompleteStep = async (scanCountBefore: number, scanCountAfter: number): Promise<void> => {
    const res = await fetch('/api/securecoder/fix_completed', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        findingsCountBefore: scanCountBefore,
        findingsCountAfter: scanCountAfter,
        findingsByFiletypeAfter: 'ts:0, go:0, py:0'
      })
    });
    const data = await res.json();
    if (!data.ok) throw new Error(data.error || 'Failed to complete runway.');
  };

  // --- Manual step wrappers (keep existing behavior) ---

  const handleGenerateThreatModel = async () => {
    setRunwayLoading(true);
    setRunwayError('');
    try {
      const result = await runThreatModelStep(runwayProject, findings);
      setRunwayThreatModel(result);
      setRunwayStep(2);
      if (runwaySessionId) await saveRunwayToDB(runwaySessionId, 2, { threat_model: result, product_id: runwayProject.id });
      // Auto-chain: immediately start next step
      setRunwayLoading(true);
      try {
        const plan = await runSecurityPlanStep(runwayProject, result);
        setRunwaySecurityPlan(plan);
        setRunwayStep(3);
        if (runwaySessionId) await saveRunwayToDB(runwaySessionId, 3, { security_plan: plan, product_id: runwayProject.id });
        // Auto-chain: remediation
        setRunwayLoading(true);
        try {
          const remediation = await runRemediationStep(runwayProject, plan);
          setRunwayRemediation(remediation);
          setRunwayStep(4);
          if (runwaySessionId) await saveRunwayToDB(runwaySessionId, 4, { remediation, product_id: runwayProject.id });
          // Auto-chain: scan & PoC
          setRunwayLoading(true);
          try {
            const { poc, scanCountAfter } = await runScanAndPoCStep(runwayProject, remediation);
            setRunwayPoC(poc);
            setRunwayScanCountAfter(scanCountAfter);
            setRunwayStep(5);
            if (runwaySessionId) await saveRunwayToDB(runwaySessionId, 5, { poc, scan_count_after: scanCountAfter, product_id: runwayProject.id });
            // Auto-chain: report
            setRunwayLoading(true);
            try {
              const report = await runReportStep(runwayProject, result, remediation, poc);
              setRunwayAuditReport(report);
              setRunwayStep(6);
              if (runwaySessionId) await saveRunwayToDB(runwaySessionId, 6, { audit_report: report, product_id: runwayProject.id });
              // Auto-chain: complete
              setRunwayLoading(true);
              try {
                await runCompleteStep(runwayScanCountBefore, scanCountAfter);
                setRunwayStep(7);
                if (runwaySessionId) await saveRunwayToDB(runwaySessionId, 7, { status: 'completed', product_id: runwayProject?.id });
              } catch (e: any) {
                setRunwayError(e.message || 'Network error logging results.');
              } finally {
                setRunwayLoading(false);
              }
            } catch (e: any) {
              setRunwayError(e.message || 'Report compilation failed.');
              setRunwayLoading(false);
            }
          } catch (e: any) {
            setRunwayError(e.message || 'Verification failed.');
            setRunwayLoading(false);
          }
        } catch (e: any) {
          setRunwayError(e.message || 'Network error.');
          setRunwayLoading(false);
        }
      } catch (e: any) {
        setRunwayError(e.message || 'Network error.');
        setRunwayLoading(false);
      }
    } catch (e: any) {
      setRunwayError(e.message || 'Network error occurred.');
      setRunwayLoading(false);
    }
  };

  const handleGenerateSecurityPlan = async () => {
    setRunwayLoading(true);
    setRunwayError('');
    try {
      const result = await runSecurityPlanStep(runwayProject, runwayThreatModel);
      setRunwaySecurityPlan(result);
      setRunwayStep(3);
      if (runwaySessionId) await saveRunwayToDB(runwaySessionId, 3, { security_plan: result, product_id: runwayProject.id });
      // Auto-chain: remediation → scan → report → complete
      setRunwayLoading(true);
      const remediation = await runRemediationStep(runwayProject, result);
      setRunwayRemediation(remediation);
      setRunwayStep(4);
      if (runwaySessionId) await saveRunwayToDB(runwaySessionId, 4, { remediation, product_id: runwayProject.id });

      setRunwayLoading(true);
      const { poc, scanCountAfter } = await runScanAndPoCStep(runwayProject, remediation);
      setRunwayPoC(poc);
      setRunwayScanCountAfter(scanCountAfter);
      setRunwayStep(5);
      if (runwaySessionId) await saveRunwayToDB(runwaySessionId, 5, { poc, scan_count_after: scanCountAfter, product_id: runwayProject.id });

      setRunwayLoading(true);
      const report = await runReportStep(runwayProject, runwayThreatModel, remediation, poc);
      setRunwayAuditReport(report);
      setRunwayStep(6);
      if (runwaySessionId) await saveRunwayToDB(runwaySessionId, 6, { audit_report: report, product_id: runwayProject.id });

      setRunwayLoading(true);
      await runCompleteStep(runwayScanCountBefore, scanCountAfter);
      setRunwayStep(7);
      if (runwaySessionId) await saveRunwayToDB(runwaySessionId, 7, { status: 'completed', product_id: runwayProject?.id });
    } catch (e: any) {
      setRunwayError(e.message || 'Pipeline error.');
    } finally {
      setRunwayLoading(false);
    }
  };

  const handleGenerateRemediation = async () => {
    setRunwayLoading(true);
    setRunwayError('');
    try {
      const result = await runRemediationStep(runwayProject, runwaySecurityPlan);
      setRunwayRemediation(result);
      setRunwayStep(4);
      if (runwaySessionId) await saveRunwayToDB(runwaySessionId, 4, { remediation: result, product_id: runwayProject.id });
      // Auto-chain: scan → report → complete
      setRunwayLoading(true);
      const { poc, scanCountAfter } = await runScanAndPoCStep(runwayProject, result);
      setRunwayPoC(poc);
      setRunwayScanCountAfter(scanCountAfter);
      setRunwayStep(5);
      if (runwaySessionId) await saveRunwayToDB(runwaySessionId, 5, { poc, scan_count_after: scanCountAfter, product_id: runwayProject.id });

      setRunwayLoading(true);
      const report = await runReportStep(runwayProject, runwayThreatModel, result, poc);
      setRunwayAuditReport(report);
      setRunwayStep(6);
      if (runwaySessionId) await saveRunwayToDB(runwaySessionId, 6, { audit_report: report, product_id: runwayProject.id });

      setRunwayLoading(true);
      await runCompleteStep(runwayScanCountBefore, scanCountAfter);
      setRunwayStep(7);
      if (runwaySessionId) await saveRunwayToDB(runwaySessionId, 7, { status: 'completed', product_id: runwayProject?.id });
    } catch (e: any) {
      setRunwayError(e.message || 'Pipeline error.');
    } finally {
      setRunwayLoading(false);
    }
  };

  const handleRunScanAndPoC = async () => {
    setRunwayLoading(true);
    setRunwayError('');
    try {
      const { poc, scanCountAfter } = await runScanAndPoCStep(runwayProject, runwayRemediation);
      setRunwayPoC(poc);
      setRunwayScanCountAfter(scanCountAfter);
      setRunwayStep(5);
      if (runwaySessionId) await saveRunwayToDB(runwaySessionId, 5, { poc, scan_count_after: scanCountAfter, product_id: runwayProject.id });
      // Auto-chain: report → complete
      setRunwayLoading(true);
      const report = await runReportStep(runwayProject, runwayThreatModel, runwayRemediation, poc);
      setRunwayAuditReport(report);
      setRunwayStep(6);
      if (runwaySessionId) await saveRunwayToDB(runwaySessionId, 6, { audit_report: report, product_id: runwayProject.id });

      setRunwayLoading(true);
      await runCompleteStep(runwayScanCountBefore, scanCountAfter);
      setRunwayStep(7);
      if (runwaySessionId) await saveRunwayToDB(runwaySessionId, 7, { status: 'completed', product_id: runwayProject?.id });
    } catch (e: any) {
      setRunwayError(e.message || 'Pipeline error.');
    } finally {
      setRunwayLoading(false);
    }
  };

  const handleGenerateReport = async () => {
    setRunwayLoading(true);
    setRunwayError('');
    try {
      const result = await runReportStep(runwayProject, runwayThreatModel, runwayRemediation, runwayPoC);
      setRunwayAuditReport(result);
      setRunwayStep(6);
      if (runwaySessionId) await saveRunwayToDB(runwaySessionId, 6, { audit_report: result, product_id: runwayProject.id });
      // Auto-chain: complete
      setRunwayLoading(true);
      await runCompleteStep(runwayScanCountBefore, runwayScanCountAfter);
      setRunwayStep(7);
      if (runwaySessionId) await saveRunwayToDB(runwaySessionId, 7, { status: 'completed', product_id: runwayProject?.id });
    } catch (e: any) {
      setRunwayError(e.message || 'Pipeline error.');
    } finally {
      setRunwayLoading(false);
    }
  };

  const handleCompleteRunway = async () => {
    setRunwayLoading(true);
    setRunwayError('');
    try {
      await runCompleteStep(runwayScanCountBefore, runwayScanCountAfter);
      setRunwayStep(7);
      if (runwaySessionId) await saveRunwayToDB(runwaySessionId, 7, { status: 'completed', product_id: runwayProject?.id });
    } catch (e: any) {
      setRunwayError(e.message || 'Network error logging results.');
    } finally {
      setRunwayLoading(false);
    }
  };

  // --- AUTO-RUN: One-click full pipeline ---

  const RUNWAY_PHASES = [
    { label: t('SimpleDashboardPage.runway.phases.init'), shortLabel: 'INIT' },
    { label: t('SimpleDashboardPage.runway.phases.threatModel'), shortLabel: 'THREAT MODEL' },
    { label: t('SimpleDashboardPage.runway.phases.plan'), shortLabel: 'SECURITY PLAN' },
    { label: t('SimpleDashboardPage.runway.phases.remediation'), shortLabel: 'REMEDIATION' },
    { label: t('SimpleDashboardPage.runway.phases.poc'), shortLabel: 'SCAN & POC' },
    { label: t('SimpleDashboardPage.runway.phases.report'), shortLabel: 'REPORT' },
    { label: t('SimpleDashboardPage.runway.phases.sync'), shortLabel: 'SYNC' },
    { label: t('SimpleDashboardPage.runway.phases.complete'), shortLabel: 'DONE' },
  ];

  const handleRunwayAutoRun = async () => {
    if (!runwayProject) return;
    setRunwayAutoMode(true);
    setRunwayLoading(true);
    setRunwayError('');

    // Reset all results
    setRunwayThreatModel('');
    setRunwaySecurityPlan('');
    setRunwayRemediation('');
    setRunwayPoC('');
    setRunwayAuditReport('');
    setRunwayScanCountAfter(0);

    const projectFindings = findings.filter(f => f.product_id === runwayProject.id && f.status !== 'triage');
    const scanCountBefore = projectFindings.length;
    setRunwayScanCountBefore(scanCountBefore);

    // Create DB session
    let sessionId = runwaySessionId;
    if (!sessionId) {
      try {
        const createRes = await fetch('/api/runway', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ product_id: runwayProject.id, auto_mode: true })
        });
        const createData = await createRes.json();
        if (createData.ok && createData.session) {
          sessionId = createData.session.id;
          setRunwaySessionId(sessionId);
        }
      } catch (e) {
        console.error('Failed to create runway session in DB:', e);
      }
    }

    try {
      // Step 1: Threat Model
      setRunwayStep(1);
      setRunwayAutoPhase(RUNWAY_PHASES[1].label);
      const threatModel = await runThreatModelStep(runwayProject, findings);
      setRunwayThreatModel(threatModel);
      if (sessionId) await saveRunwayToDB(sessionId, 1, { threat_model: threatModel, scan_count_before: scanCountBefore, product_id: runwayProject.id, auto_mode: true });

      // Step 2: Security Plan
      setRunwayStep(2);
      setRunwayAutoPhase(RUNWAY_PHASES[2].label);
      const securityPlan = await runSecurityPlanStep(runwayProject, threatModel);
      setRunwaySecurityPlan(securityPlan);
      if (sessionId) await saveRunwayToDB(sessionId, 2, { security_plan: securityPlan, product_id: runwayProject.id });

      // Step 3: Remediation
      setRunwayStep(3);
      setRunwayAutoPhase(RUNWAY_PHASES[3].label);
      const remediation = await runRemediationStep(runwayProject, securityPlan);
      setRunwayRemediation(remediation);
      if (sessionId) await saveRunwayToDB(sessionId, 3, { remediation, product_id: runwayProject.id });

      // Step 4: Scan & PoC
      setRunwayStep(4);
      setRunwayAutoPhase(RUNWAY_PHASES[4].label);
      const { poc, scanCountAfter } = await runScanAndPoCStep(runwayProject, remediation);
      setRunwayPoC(poc);
      setRunwayScanCountAfter(scanCountAfter);
      if (sessionId) await saveRunwayToDB(sessionId, 4, { poc, scan_count_after: scanCountAfter, product_id: runwayProject.id });

      // Step 5: Audit Report
      setRunwayStep(5);
      setRunwayAutoPhase(RUNWAY_PHASES[5].label);
      const report = await runReportStep(runwayProject, threatModel, remediation, poc);
      setRunwayAuditReport(report);
      if (sessionId) await saveRunwayToDB(sessionId, 5, { audit_report: report, product_id: runwayProject.id });

      // Step 6: Sync & Complete
      setRunwayStep(6);
      setRunwayAutoPhase(RUNWAY_PHASES[6].label);
      await runCompleteStep(scanCountBefore, scanCountAfter);

      // Done
      setRunwayStep(7);
      setRunwayAutoPhase(RUNWAY_PHASES[7].label);
      if (sessionId) await saveRunwayToDB(sessionId, 7, { status: 'completed', product_id: runwayProject.id });
    } catch (e: any) {
      const errMsg = e.message || 'Auto-run pipeline failed.';
      setRunwayError(errMsg);
      if (sessionId) await saveRunwayToDB(sessionId, runwayStep, { status: 'failed', error_message: errMsg, product_id: runwayProject.id });
    } finally {
      setRunwayLoading(false);
      setRunwayAutoMode(false);
    }
  };

  const handleResetRunway = async () => {
    // Delete session from DB
    if (runwaySessionId) {
      try {
        await fetch(`/api/runway/${runwaySessionId}`, { method: 'DELETE' });
      } catch (e) {
        console.error('Failed to delete runway session:', e);
      }
    }
    setRunwaySessionId(null);
    setRunwayStep(0);
    setRunwayProject(null);
    setRunwayThreatModel('');
    setRunwaySecurityPlan('');
    setRunwayRemediation('');
    setRunwayPoC('');
    setRunwayAuditReport('');
    setRunwayScanCountBefore(0);
    setRunwayScanCountAfter(0);
    setRunwayError('');
    setRunwayAutoMode(false);
    setRunwayAutoPhase('');
  };

  const handleDownloadMarkdown = () => {
    if (!runwayProject) return;
    
    let md = `# 🛡️ AITriage Security Audit Report\n\n`;
    md += `**Project**: ${runwayProject.name}\n`;
    md += `**Date**: ${new Date().toLocaleString()}\n`;
    md += `**Session ID**: ${runwaySessionId || 'N/A'}\n`;
    md += `**Findings**: ${runwayScanCountBefore} before → ${runwayScanCountAfter} after\n\n`;
    md += `---\n\n`;

    if (runwayThreatModel) {
      md += `## 1. STRIDE Threat Model\n\n${runwayThreatModel}\n\n---\n\n`;
    }
    if (runwaySecurityPlan) {
      md += `## 2. Security Implementation Plan\n\n${runwaySecurityPlan}\n\n---\n\n`;
    }
    if (runwayRemediation) {
      md += `## 3. Remediation Patches\n\n${runwayRemediation}\n\n---\n\n`;
    }
    if (runwayPoC) {
      md += `## 4. Proof of Concept Verification\n\n${runwayPoC}\n\n---\n\n`;
    }
    if (runwayAuditReport) {
      md += `## 5. Audit Report\n\n${runwayAuditReport}\n\n---\n\n`;
    }
    md += `\n*Generated by AITriage SecureCoder Agent*\n`;

    const blob = new Blob([md], { type: 'text/markdown;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const downloadAnchor = document.createElement('a');
    downloadAnchor.setAttribute("href", url);
    const dateStr = new Date().toISOString().split('T')[0];
    downloadAnchor.setAttribute("download", `runway-report-${runwaySessionId || 'session'}-${dateStr}.md`);
    document.body.appendChild(downloadAnchor);
    downloadAnchor.click();
    downloadAnchor.remove();
    URL.revokeObjectURL(url);
  };

  const handleExportToProject = async () => {
    if (!runwaySessionId) return;
    setRunwayExporting(true);
    try {
      const res = await fetch(`/api/runway/export/${runwaySessionId}`, { method: 'POST' });
      const data = await res.json();
      if (data.ok) {
        alert(t('SimpleDashboardPage.runway.exportSuccess', { path: data.saved_to || 'aitriage/' }));
      } else {
        alert(data.error || 'Failed to export report.');
      }
    } catch (e) {
      alert('Error exporting report: network failure.');
    } finally {
      setRunwayExporting(false);
    }
  };

  return (
    <motion.div variants={itemVariants} className="border border-[rgba(255,255,255,0.06)] rounded-lg bg-[rgba(255,255,255,0.01)] overflow-hidden mb-6">
      <div className="flex items-center justify-between px-6 py-4 border-b border-[rgba(255,255,255,0.06)]">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded-lg bg-[var(--accent-color-soft)] border border-[var(--accent-color-line)] flex items-center justify-center">
            <span className="material-symbols-outlined text-[16px]" style={{ color: 'var(--accent-color)', fontVariationSettings: "'FILL' 1" }}>security</span>
          </div>
          <div className="flex-1">
            <h3 className="text-[12px] font-bold text-[#f4f4f5] tracking-wider uppercase">{t('SimpleDashboardPage.runway.securecoderIntegration')}</h3>
            <p className="text-[10px] text-[#52525b] tracking-widest uppercase">{t('SimpleDashboardPage.runway.aiAgentCompatibilityLayer')}</p>
          </div>
        </div>
        <button
          onClick={() => setRunwayOpen(!runwayOpen)}
          className="px-3 py-1 bg-[var(--accent-color)] hover:bg-[var(--accent-color-hover)] text-[var(--accent-color-on-text)] rounded text-[10px] font-bold uppercase tracking-wider flex items-center gap-1.5 shadow-[0_0_15px_var(--accent-color-line)] hover:shadow-[0_0_25px_var(--accent-color-soft)] transition-all duration-300 cursor-pointer"
        >
          <span className="material-symbols-outlined text-[12px]">{runwayOpen ? 'close' : 'bolt'}</span>
          {runwayOpen ? t('SimpleDashboardPage.runway.closeRunway') : t('SimpleDashboardPage.runway.runwayWizard')}
        </button>
      </div>

      {runwayOpen ? (
        <div className="p-6 bg-[rgba(0,0,0,0.15)] space-y-4">
          <div className="flex items-center justify-between border-b border-[rgba(255,255,255,0.06)] pb-3">
            <span className="text-[11px] font-bold text-[#a1a1aa] uppercase tracking-wider">{t('SimpleDashboardPage.runway.agentRunway')}</span>
            <div className="flex items-center gap-3">
              {runwayAutoMode && (
                <span className="text-[9px] text-[var(--accent-color)] font-mono uppercase tracking-widest animate-pulse flex items-center gap-1">
                  <span className="material-symbols-outlined text-[10px]">auto_mode</span>
                  AUTO
                </span>
              )}
              <span className="text-[10px] text-[var(--accent-color)] font-mono">{t('SimpleDashboardPage.runway.stepIndicator', { current: runwayStep })}</span>
            </div>
          </div>

          {/* Stepper progress bar */}
          <div className="flex gap-1">
            {Array.from({ length: 7 }).map((_, i) => {
              const stepNum = i + 1;
              const isCompleted = stepNum < runwayStep || (stepNum === runwayStep && runwayStep === 7);
              const isActive = stepNum === runwayStep && runwayStep < 7;
              return (
                <div
                  key={i}
                  className={`h-1.5 flex-1 rounded-full transition-all duration-500 relative overflow-hidden ${
                    isCompleted 
                      ? 'bg-[var(--accent-color)]' 
                      : isActive && runwayLoading 
                      ? 'bg-[rgba(255,255,255,0.06)]' 
                      : isActive 
                      ? 'bg-[var(--accent-color)] opacity-40' 
                      : 'bg-[rgba(255,255,255,0.06)]'
                  }`}
                >
                  {isActive && runwayLoading && (
                    <div className="absolute inset-0 bg-gradient-to-r from-[var(--accent-color)] via-[var(--accent-color-hover)] to-transparent animate-[shimmer_1.5s_ease-in-out_infinite]" style={{ backgroundSize: '200% 100%' }} />
                  )}
                </div>
              );
            })}
          </div>

          {/* Auto-mode phase labels */}
          {runwayAutoMode && runwayStep > 0 && runwayStep < 7 && (
            <div className="flex gap-1 -mt-2">
              {RUNWAY_PHASES.slice(0, -1).map((_phase, i) => (
                <div key={i} className={`flex-1 text-center text-[7px] font-mono uppercase tracking-wider pt-1 transition-colors duration-300 ${
                  i < runwayStep ? 'text-[var(--accent-color)]' : i === runwayStep ? 'text-white' : 'text-[#27272a]'
                }`}>
                  {i < runwayStep ? '✓' : i === runwayStep ? '◉' : '·'}
                </div>
              ))}
            </div>
          )}

          {/* Auto-mode current status */}
          {runwayAutoMode && runwayLoading && (
            <div 
              className="flex items-center gap-3 py-3 px-4 rounded-lg border"
              style={{
                backgroundColor: 'var(--accent-color-soft)',
                borderColor: 'var(--accent-color-line)'
              }}
            >
              <div className="w-5 h-5 border-2 border-[rgba(255,255,255,0.08)] border-t-[var(--accent-color)] rounded-full animate-spin shrink-0" />
              <div className="flex-1">
                <span className="text-[11px] text-white font-semibold uppercase tracking-wider">{runwayAutoPhase}</span>
                <span className="text-[10px] text-[#52525b] ml-2 font-mono">{t('SimpleDashboardPage.runway.stepIndicator', { current: runwayStep }).toLowerCase()}</span>
              </div>
              <button
                onClick={() => { setRunwayAutoMode(false); }}
                className="text-[9px] text-[#ef4444] hover:text-[#f87171] uppercase font-mono font-bold tracking-wider transition-colors"
              >
                {t('SimpleDashboardPage.runway.abort')}
              </button>
            </div>
          )}

          {runwayError && (
            <div className="p-3 bg-[rgba(239,68,68,0.08)] border border-[rgba(239,68,68,0.15)] rounded text-[11px] text-[#ef4444] font-medium flex items-center gap-2">
              <span className="material-symbols-outlined text-[14px]">error</span>
              {runwayError}
              {runwayStep > 0 && runwayStep < 7 && (
                <button onClick={handleRunwayAutoRun} className="ml-auto text-[10px] text-[#ef4444] hover:text-[#f87171] uppercase font-mono font-bold underline">{t('SimpleDashboardPage.runway.retry')}</button>
              )}
            </div>
          )}

          {/* STEP 0: Project selection */}
          {runwayStep === 0 && (
            <div className="space-y-4 pt-2">
              <p className="text-[11px] text-[#71717a] leading-relaxed">{t('SimpleDashboardPage.runway.selectProjectStartDesc')}</p>
              <div className="flex gap-2">
                <select
                  value={runwayProject ? runwayProject.id : ''}
                  onChange={e => {
                    const id = Number(e.target.value);
                    setRunwayProject(activeProducts.find(p => p.id === id) || null);
                  }}
                  className="flex-1 bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.06)] rounded px-3 py-2 text-[12px] text-white outline-none focus:border-[var(--accent-color)] cursor-pointer"
                >
                  <option value="">{t('SimpleDashboardPage.runway.chooseProject')}</option>
                  {activeProducts.map(p => <option key={p.id} value={p.id}>{p.name}</option>)}
                </select>
              </div>
              <div className="flex gap-2">
                <button
                  onClick={handleRunwayAutoRun}
                  disabled={!runwayProject}
                  className="flex-1 px-4 py-2.5 bg-[var(--accent-color)] hover:bg-[var(--accent-color-hover)] text-[var(--accent-color-on-text)] rounded text-[12px] font-bold uppercase tracking-wider disabled:opacity-30 transition-all duration-300 flex items-center justify-center gap-2 shadow-[0_0_15px_var(--accent-color-line)] hover:shadow-[0_0_25px_var(--accent-color-soft)]"
                >
                  <span className="material-symbols-outlined text-[16px]">play_arrow</span>
                  {t('SimpleDashboardPage.runway.startFullPipeline')}
                </button>
                <button
                  onClick={handleInitRunway}
                  disabled={!runwayProject}
                  className="px-3 py-2.5 bg-[rgba(255,255,255,0.04)] border border-[rgba(255,255,255,0.08)] hover:bg-[rgba(255,255,255,0.08)] text-[#a1a1aa] hover:text-[#f4f4f5] rounded text-[10px] font-bold uppercase tracking-wider disabled:opacity-30 transition-colors flex items-center gap-1.5"
                  title={t('SimpleDashboardPage.runway.runEachStepManually')}
                >
                  <span className="material-symbols-outlined text-[12px]">tune</span>
                  {t('SimpleDashboardPage.runway.stepByStep')}
                </button>
              </div>
            </div>
          )}

          {/* STEP 1: Threat Model */}
          {runwayStep === 1 && !runwayAutoMode && (
            <div className="space-y-3 pt-2">
              <div className="flex items-center justify-between">
                <span className="text-[11px] text-white font-semibold uppercase font-mono">{t('SimpleDashboardPage.runway.step1Title', { name: runwayProject.name })}</span>
                <span className="text-[9px] text-[#71717a] font-mono">{t('SimpleDashboardPage.runway.foundIssuesCount', { count: runwayScanCountBefore })}</span>
              </div>
              {runwayLoading ? (
                <div className="flex flex-col items-center justify-center py-6 gap-2">
                  <div className="w-5 h-5 border-2 border-[rgba(255,255,255,0.08)] border-t-[var(--accent-color)] rounded-full animate-spin" />
                  <span className="text-[10px] text-[#71717a] font-mono uppercase tracking-widest animate-pulse">{t('SimpleDashboardPage.runway.runningStrideThreatModel')}</span>
                </div>
              ) : runwayThreatModel ? (
                <div className="bg-[rgba(0,0,0,0.3)] border border-[rgba(255,255,255,0.04)] rounded p-3 max-h-56 overflow-y-auto text-[11px] text-[#a1a1aa] leading-relaxed prose prose-invert select-text" style={{ scrollbarWidth: 'thin' }}>
                  <Markdown>{runwayThreatModel}</Markdown>
                </div>
              ) : (
                <p className="text-[11px] text-[#71717a]">{t('SimpleDashboardPage.runway.readyThreatModel')}</p>
              )}
              <div className="flex justify-between pt-1">
                <button onClick={handleResetRunway} className="text-[10px] text-[#71717a] hover:text-[#a1a1aa] uppercase font-mono font-bold">{t('SimpleDashboardPage.runway.reset')}</button>
                {!runwayThreatModel ? (
                  <button onClick={handleGenerateThreatModel} disabled={runwayLoading} className="px-4 py-1.5 bg-[var(--accent-color)] hover:bg-[var(--accent-color-hover)] text-[var(--accent-color-on-text)] rounded text-[11px] font-bold uppercase tracking-wider transition-colors">
                    {t('SimpleDashboardPage.runway.analyzeThreats')}
                  </button>
                ) : (
                  <button onClick={() => setRunwayStep(2)} className="px-4 py-1.5 bg-[var(--accent-color)] hover:bg-[var(--accent-color-hover)] text-[var(--accent-color-on-text)] rounded text-[11px] font-bold uppercase tracking-wider transition-colors">
                    {t('SimpleDashboardPage.runway.nextPlan')}
                  </button>
                )}
              </div>
            </div>
          )}

          {/* STEP 2: Security Plan */}
          {runwayStep === 2 && !runwayAutoMode && (
            <div className="space-y-3 pt-2">
              <span className="text-[11px] text-white font-semibold uppercase font-mono">{t('SimpleDashboardPage.runway.step2Title')}</span>
              {runwayLoading ? (
                <div className="flex flex-col items-center justify-center py-6 gap-2">
                  <div className="w-5 h-5 border-2 border-[rgba(255,255,255,0.08)] border-t-[var(--accent-color)] rounded-full animate-spin" />
                  <span className="text-[10px] text-[#71717a] font-mono uppercase tracking-widest animate-pulse">{t('SimpleDashboardPage.runway.runningSecurityPlan')}</span>
                </div>
              ) : runwaySecurityPlan ? (
                <div className="bg-[rgba(0,0,0,0.3)] border border-[rgba(255,255,255,0.04)] rounded p-3 max-h-56 overflow-y-auto text-[11px] text-[#a1a1aa] leading-relaxed prose prose-invert select-text" style={{ scrollbarWidth: 'thin' }}>
                  <Markdown>{runwaySecurityPlan}</Markdown>
                </div>
              ) : (
                <p className="text-[11px] text-[#71717a]">{t('SimpleDashboardPage.runway.readySecurityPlan')}</p>
              )}
              <div className="flex justify-between pt-1">
                <button onClick={() => setRunwayStep(1)} className="text-[10px] text-[#71717a] hover:text-[#a1a1aa] uppercase font-mono font-bold">{t('SimpleDashboardPage.runway.reset')}</button>
                {!runwaySecurityPlan ? (
                  <button onClick={handleGenerateSecurityPlan} disabled={runwayLoading} className="px-4 py-1.5 bg-[var(--accent-color)] hover:bg-[var(--accent-color-hover)] text-[var(--accent-color-on-text)] rounded text-[11px] font-bold uppercase tracking-wider transition-colors">
                    {t('SimpleDashboardPage.runway.createPlan')}
                  </button>
                ) : (
                  <button onClick={() => setRunwayStep(3)} className="px-4 py-1.5 bg-[var(--accent-color)] hover:bg-[var(--accent-color-hover)] text-[var(--accent-color-on-text)] rounded text-[11px] font-bold uppercase tracking-wider transition-colors">
                    {t('SimpleDashboardPage.runway.nextPatch')}
                  </button>
                )}
              </div>
            </div>
          )}

          {/* STEP 3: Remediation Diffs */}
          {runwayStep === 3 && !runwayAutoMode && (
            <div className="space-y-3 pt-2">
              <span className="text-[11px] text-white font-semibold uppercase font-mono">{t('SimpleDashboardPage.runway.step3Title')}</span>
              {runwayLoading ? (
                <div className="flex flex-col items-center justify-center py-6 gap-2">
                  <div className="w-5 h-5 border-2 border-[rgba(255,255,255,0.08)] border-t-[var(--accent-color)] rounded-full animate-spin" />
                  <span className="text-[10px] text-[#71717a] font-mono uppercase tracking-widest animate-pulse">{t('SimpleDashboardPage.runway.runningRemediation')}</span>
                </div>
              ) : runwayRemediation ? (
                <div className="bg-[rgba(0,0,0,0.3)] border border-[rgba(255,255,255,0.04)] rounded p-3 max-h-56 overflow-y-auto text-[11px] text-[#a1a1aa] leading-relaxed prose prose-invert select-text" style={{ scrollbarWidth: 'thin' }}>
                  <Markdown>{runwayRemediation}</Markdown>
                </div>
              ) : (
                <p className="text-[11px] text-[#71717a]">{t('SimpleDashboardPage.runway.readyRemediation')}</p>
              )}
              <div className="flex justify-between pt-1">
                <button onClick={() => setRunwayStep(2)} className="text-[10px] text-[#71717a] hover:text-[#a1a1aa] uppercase font-mono font-bold">{t('SimpleDashboardPage.runway.reset')}</button>
                {!runwayRemediation ? (
                  <button onClick={handleGenerateRemediation} disabled={runwayLoading} className="px-4 py-1.5 bg-[var(--accent-color)] hover:bg-[var(--accent-color-hover)] text-[var(--accent-color-on-text)] rounded text-[11px] font-bold uppercase tracking-wider transition-colors">
                    {t('SimpleDashboardPage.runway.generatePatches')}
                  </button>
                ) : (
                  <button onClick={() => setRunwayStep(4)} className="px-4 py-1.5 bg-[var(--accent-color)] hover:bg-[var(--accent-color-hover)] text-[var(--accent-color-on-text)] rounded text-[11px] font-bold uppercase tracking-wider transition-colors">
                    {t('SimpleDashboardPage.runway.nextVerify')}
                  </button>
                )}
              </div>
            </div>
          )}

          {/* STEP 4: Scan & PoC */}
          {runwayStep === 4 && !runwayAutoMode && (
            <div className="space-y-3 pt-2">
              <span className="text-[11px] text-white font-semibold uppercase font-mono">{t('SimpleDashboardPage.runway.step4Title')}</span>
              {runwayLoading ? (
                <div className="flex flex-col items-center justify-center py-6 gap-2">
                  <div className="w-5 h-5 border-2 border-[rgba(255,255,255,0.08)] border-t-[var(--accent-color)] rounded-full animate-spin" />
                  <span className="text-[10px] text-[#71717a] font-mono uppercase tracking-widest animate-pulse">{t('SimpleDashboardPage.runway.runningScanPoC')}</span>
                </div>
              ) : runwayPoC ? (
                <div className="space-y-2">
                  <div className="text-[10px] text-[#22c55e] font-mono bg-[rgba(34,197,94,0.08)] border border-[rgba(34,197,94,0.15)] px-2.5 py-1.5 rounded">
                    {t('SimpleDashboardPage.runway.scanPoCResult', { count: runwayScanCountAfter, fixed: runwayScanCountBefore - runwayScanCountAfter })}
                  </div>
                  <div className="bg-[rgba(0,0,0,0.3)] border border-[rgba(255,255,255,0.04)] rounded p-3 max-h-48 overflow-y-auto text-[11px] text-[#a1a1aa] leading-relaxed prose prose-invert select-text" style={{ scrollbarWidth: 'thin' }}>
                    <Markdown>{runwayPoC}</Markdown>
                  </div>
                </div>
              ) : (
                <p className="text-[11px] text-[#71717a]">{t('SimpleDashboardPage.runway.readyScanPoC')}</p>
              )}
              <div className="flex justify-between pt-1">
                <button onClick={() => setRunwayStep(3)} className="text-[10px] text-[#71717a] hover:text-[#a1a1aa] uppercase font-mono font-bold">{t('SimpleDashboardPage.runway.reset')}</button>
                {!runwayPoC ? (
                  <button onClick={handleRunScanAndPoC} disabled={runwayLoading} className="px-4 py-1.5 bg-[var(--accent-color)] hover:bg-[var(--accent-color-hover)] text-[var(--accent-color-on-text)] rounded text-[11px] font-bold uppercase tracking-wider transition-colors">
                    {t('SimpleDashboardPage.runway.verifyFixes')}
                  </button>
                ) : (
                  <button onClick={() => setRunwayStep(5)} className="px-4 py-1.5 bg-[var(--accent-color)] hover:bg-[var(--accent-color-hover)] text-[var(--accent-color-on-text)] rounded text-[11px] font-bold uppercase tracking-wider transition-colors">
                    {t('SimpleDashboardPage.runway.nextReport')}
                  </button>
                )}
              </div>
            </div>
          )}

          {/* STEP 5: Audit Report */}
          {runwayStep === 5 && !runwayAutoMode && (
            <div className="space-y-3 pt-2">
              <span className="text-[11px] text-white font-semibold uppercase font-mono">{t('SimpleDashboardPage.runway.step5Title')}</span>
              {runwayLoading ? (
                <div className="flex flex-col items-center justify-center py-6 gap-2">
                  <div className="w-5 h-5 border-2 border-[rgba(255,255,255,0.08)] border-t-[var(--accent-color)] rounded-full animate-spin" />
                  <span className="text-[10px] text-[#71717a] font-mono uppercase tracking-widest animate-pulse">{t('SimpleDashboardPage.runway.runningReport')}</span>
                </div>
              ) : runwayAuditReport ? (
                <div className="bg-[rgba(0,0,0,0.3)] border border-[rgba(255,255,255,0.04)] rounded p-3 max-h-56 overflow-y-auto text-[11px] text-[#a1a1aa] leading-relaxed prose prose-invert select-text" style={{ scrollbarWidth: 'thin' }}>
                  <Markdown>{runwayAuditReport}</Markdown>
                </div>
              ) : (
                <p className="text-[11px] text-[#71717a]">{t('SimpleDashboardPage.runway.readyReport')}</p>
              )}
              <div className="flex justify-between pt-1">
                <button onClick={() => setRunwayStep(4)} className="text-[10px] text-[#71717a] hover:text-[#a1a1aa] uppercase font-mono font-bold">{t('SimpleDashboardPage.runway.reset')}</button>
                {!runwayAuditReport ? (
                  <button onClick={handleGenerateReport} disabled={runwayLoading} className="px-4 py-1.5 bg-[var(--accent-color)] hover:bg-[var(--accent-color-hover)] text-[var(--accent-color-on-text)] rounded text-[11px] font-bold uppercase tracking-wider transition-colors">
                    {t('SimpleDashboardPage.runway.generateReport')}
                  </button>
                ) : (
                  <button onClick={() => setRunwayStep(6)} className="px-4 py-1.5 bg-[var(--accent-color)] hover:bg-[var(--accent-color-hover)] text-[var(--accent-color-on-text)] rounded text-[11px] font-bold uppercase tracking-wider transition-colors">
                    {t('SimpleDashboardPage.runway.nextFinish')}
                  </button>
                )}
              </div>
            </div>
          )}

          {/* STEP 6: Complete Runway */}
          {runwayStep === 6 && !runwayAutoMode && (
            <div className="space-y-3 pt-2">
              <span className="text-[11px] text-white font-semibold uppercase font-mono">{t('SimpleDashboardPage.runway.step6Title')}</span>
              <p className="text-[11px] text-[#71717a] leading-relaxed">{t('SimpleDashboardPage.runway.readyComplete')}</p>
              <div className="flex justify-between pt-1">
                <button onClick={() => setRunwayStep(5)} className="text-[10px] text-[#71717a] hover:text-[#a1a1aa] uppercase font-mono font-bold">{t('SimpleDashboardPage.runway.reset')}</button>
                <button onClick={handleCompleteRunway} disabled={runwayLoading} className="px-4 py-1.5 bg-[var(--accent-color)] hover:bg-[var(--accent-color-hover)] text-[var(--accent-color-on-text)] rounded text-[11px] font-bold uppercase tracking-wider transition-colors">
                  {runwayLoading ? t('SimpleDashboardPage.runway.syncing') : t('SimpleDashboardPage.runway.syncResultsAndFinish')}
                </button>
              </div>
            </div>
          )}

          {/* STEP 7: Completed Success */}
          {runwayStep === 7 && (
            <div className="space-y-3 pt-2 text-center py-4">
              <span className="material-symbols-outlined text-[36px] text-[#22c55e]">check_circle</span>
              <h4 className="text-[12px] font-bold text-white uppercase tracking-wider">{t('SimpleDashboardPage.runway.remediationComplete')}</h4>
              <p className="text-[11px] text-[#71717a] leading-relaxed max-w-xs mx-auto">
                {t('SimpleDashboardPage.runway.remediationCompleteDesc')}
              </p>
              <div className="flex flex-col gap-2 max-w-xs mx-auto mt-2">
                <button
                  onClick={handleExportToProject}
                  disabled={runwayExporting}
                  className="px-4 py-1.5 bg-[var(--accent-color)] hover:bg-[var(--accent-color-hover)] text-[var(--accent-color-on-text)] rounded text-[11px] font-bold uppercase tracking-wider transition-colors flex items-center justify-center gap-1.5 cursor-pointer disabled:opacity-50"
                >
                  <span className="material-symbols-outlined text-[13px]">ios_share</span>
                  {runwayExporting ? t('SimpleDashboardPage.runway.syncing') : t('SimpleDashboardPage.runway.exportToProject')}
                </button>
                <button
                  onClick={handleDownloadMarkdown}
                  className="px-4 py-1.5 bg-[rgba(255,255,255,0.04)] border border-[rgba(255,255,255,0.08)] hover:bg-[rgba(255,255,255,0.08)] text-[#f4f4f5] rounded text-[11px] font-bold uppercase tracking-wider transition-colors flex items-center justify-center gap-1.5 cursor-pointer"
                >
                  <span className="material-symbols-outlined text-[13px]">download</span>
                  {t('SimpleDashboardPage.runway.downloadMdReport')}
                </button>
                <button
                  onClick={handleResetRunway}
                  className="px-4 py-1.5 bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.06)] hover:bg-[rgba(255,255,255,0.05)] text-[#a1a1aa] rounded text-[11px] font-bold uppercase tracking-wider transition-colors"
                >
                  {t('SimpleDashboardPage.runway.runAgain')}
                </button>
              </div>
            </div>
          )}
        </div>
      ) : (
        <div className="divide-y divide-[rgba(255,255,255,0.04)]">
          {/* Quick Target Scan */}
          <div>
            <button onClick={() => setExpandedCat(expandedCat === 'scan' ? null : 'scan')} className="w-full flex items-center gap-3 px-6 py-3 hover:bg-[rgba(255,255,255,0.02)] transition-colors group">
              <span className={`material-symbols-outlined text-[16px] transition-colors ${expandedCat === 'scan' ? 'text-[var(--accent-color)]' : 'text-[#3f3f46] group-hover:text-[var(--accent-color)]'}`}>document_scanner</span>
              <span className="text-[11px] font-semibold tracking-wider text-[#a1a1aa] uppercase flex-1 text-left">{t('SimpleDashboardPage.runway.quickTargetScan')}</span>
              <span className={`material-symbols-outlined text-[14px] text-[#3f3f46] transition-transform duration-200 ${expandedCat === 'scan' ? 'rotate-180 text-[var(--accent-color)]' : ''}`}>expand_more</span>
            </button>
            <AnimatePresence>
              {expandedCat === 'scan' && (
                <motion.div initial={{ height: 0 }} animate={{ height: 'auto' }} exit={{ height: 0 }} className="overflow-hidden">
                  <div className="px-6 pb-4 pt-2">
                    <div className="flex gap-4 mb-2.5">
                      <label className="flex items-center gap-1.5 text-[10px] text-[#a1a1aa] cursor-pointer">
                        <input
                          type="radio"
                          name="quickScanTarget"
                          checked={quickScanType === 'file'}
                          onChange={() => { setQuickScanType('file'); setScanResult(null); }}
                          className="accent-[var(--accent-color)]"
                        />
                        {t('SimpleDashboardPage.runway.file')}
                      </label>
                      <label className="flex items-center gap-1.5 text-[10px] text-[#a1a1aa] cursor-pointer">
                        <input
                          type="radio"
                          name="quickScanTarget"
                          checked={quickScanType === 'dir'}
                          onChange={() => { setQuickScanType('dir'); setScanResult(null); }}
                          className="accent-[var(--accent-color)]"
                        />
                        {t('SimpleDashboardPage.runway.directory')}
                      </label>
                    </div>
                    <div className="flex gap-2">
                      <input
                        type="text"
                        value={scanPath}
                        onChange={e => setScanPath(e.target.value)}
                        placeholder={quickScanType === 'file' ? '/path/to/file.ts' : '/path/to/directory'}
                        className="flex-1 bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.06)] rounded px-3 py-1.5 text-[11px] text-white outline-none focus:border-[var(--accent-color)]"
                      />
                      <button onClick={handleScan} disabled={scanning} className="px-4 py-1.5 bg-[var(--accent-color)] hover:bg-[var(--accent-color-hover)] text-[var(--accent-color-on-text)] rounded text-[11px] font-bold uppercase tracking-wider disabled:opacity-50 transition-colors">
                        {scanning ? t('SimpleDashboardPage.runway.scanning') : t('SimpleDashboardPage.runway.scan')}
                      </button>
                    </div>
                    {scanResult && (
                      <div className="mt-3 bg-[rgba(0,0,0,0.2)] border border-[rgba(255,255,255,0.04)] rounded p-3 max-h-48 overflow-y-auto" style={{ scrollbarWidth: 'thin' }}>
                        {scanResult.length === 0 ? (
                          <div className="text-[10px] text-[#a1a1aa]">{t('SimpleDashboardPage.runway.noVulnsFound')}</div>
                        ) : (
                          <div className="space-y-2">
                            {scanResult.map((f: any, i: number) => (
                              <div key={i} className="text-[10px] border-b border-[rgba(255,255,255,0.03)] pb-1.5 last:border-0 last:pb-0">
                                <div className="flex items-center justify-between">
                                  <span className="text-[#a1a1aa] font-mono font-bold truncate pr-2">{f.subcategory || f.ruleId}</span>
                                  <span className="text-red-400 font-bold uppercase shrink-0 text-[8px] border border-red-500/20 px-1 rounded">{f.labels?.severity}</span>
                                </div>
                                <div className="text-[#71717a] mt-0.5 select-text leading-snug">{f.message}</div>
                                {f.location?.path && (
                                  <div className="text-[8px] text-[#52525b] font-mono mt-0.5 truncate" title={f.location.path}>
                                    {f.location.path.split('/').pop()}:{f.location.range?.textRange?.startLine || f.location.range?.startLine}
                                  </div>
                                )}
                              </div>
                            ))}
                          </div>
                        )}
                      </div>
                    )}
                  </div>
                </motion.div>
              )}
            </AnimatePresence>
          </div>

          {/* Dependency Scan */}
          <div>
            <button onClick={() => setExpandedCat(expandedCat === 'deps' ? null : 'deps')} className="w-full flex items-center gap-3 px-6 py-3 hover:bg-[rgba(255,255,255,0.02)] transition-colors group">
              <span className={`material-symbols-outlined text-[16px] transition-colors ${expandedCat === 'deps' ? 'text-[var(--accent-color)]' : 'text-[#3f3f46] group-hover:text-[var(--accent-color)]'}`}>package_2</span>
              <span className="text-[11px] font-semibold tracking-wider text-[#a1a1aa] uppercase flex-1 text-left">{t('SimpleDashboardPage.runway.dependencyScanner')}</span>
              <span className={`material-symbols-outlined text-[14px] text-[#3f3f46] transition-transform duration-200 ${expandedCat === 'deps' ? 'rotate-180 text-[var(--accent-color)]' : ''}`}>expand_more</span>
            </button>
            <AnimatePresence>
              {expandedCat === 'deps' && (
                <motion.div initial={{ height: 0 }} animate={{ height: 'auto' }} exit={{ height: 0 }} className="overflow-hidden">
                  <div className="px-6 pb-4 pt-2">
                    <div className="flex gap-2">
                      <select value={depRegistry} onChange={e => setDepRegistry(e.target.value)} className="bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.06)] rounded px-2 text-[11px] text-white outline-none focus:border-[var(--accent-color)]">
                        <option value="npm">npm</option>
                        <option value="pypi">PyPI</option>
                        <option value="gomodproxy">Go</option>
                      </select>
                      <input type="text" value={depPackage} onChange={e => setDepPackage(e.target.value)} placeholder={t('SimpleDashboardPage.runway.packageName')} className="flex-1 bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.06)] rounded px-3 py-1.5 text-[11px] text-white outline-none focus:border-[var(--accent-color)]" />
                      <button onClick={handleDepScan} disabled={depScanning} className="px-4 py-1.5 bg-[var(--accent-color)] hover:bg-[var(--accent-color-hover)] text-[var(--accent-color-on-text)] rounded text-[11px] font-bold uppercase tracking-wider disabled:opacity-50 transition-colors">
                        {depScanning ? t('SimpleDashboardPage.runway.scanning') : t('SimpleDashboardPage.runway.scan')}
                      </button>
                    </div>
                    {depResult && (
                      <div className="mt-3 bg-[rgba(0,0,0,0.3)] border border-[rgba(255,255,255,0.04)] rounded p-3 max-h-40 overflow-y-auto">
                        {depResult.length === 0 ? (
                          <div className="text-[10px] text-[#a1a1aa]">{t('SimpleDashboardPage.runway.packageAppearsSafe')}</div>
                        ) : (
                          <div className="space-y-2">
                            {depResult.map((d: any, i: number) => (
                              <div key={i} className="text-[10px]">
                                <span className="text-orange-400 font-bold uppercase">{d.package}</span>
                                <div className="text-[#71717a] mt-0.5">{d.reason}</div>
                              </div>
                            ))}
                          </div>
                        )}
                      </div>
                    )}
                  </div>
                </motion.div>
              )}
            </AnimatePresence>
          </div>

          {/* Ignored Findings */}
          <div>
            <button onClick={() => setExpandedCat(expandedCat === 'ignore' ? null : 'ignore')} className="w-full flex items-center gap-3 px-6 py-3 hover:bg-[rgba(255,255,255,0.02)] transition-colors group">
              <span className={`material-symbols-outlined text-[16px] transition-colors ${expandedCat === 'ignore' ? 'text-[var(--accent-color)]' : 'text-[#3f3f46] group-hover:text-[var(--accent-color)]'}`}>visibility_off</span>
              <span className="text-[11px] font-semibold tracking-wider text-[#a1a1aa] uppercase flex-1 text-left">{t('SimpleDashboardPage.runway.ignoredFindings')}</span>
              <span className="text-[10px] text-[#3f3f46] mr-1">{ignoredFindings.length}</span>
              <span className={`material-symbols-outlined text-[14px] text-[#3f3f46] transition-transform duration-200 ${expandedCat === 'ignore' ? 'rotate-180 text-[var(--accent-color)]' : ''}`}>expand_more</span>
            </button>
            <AnimatePresence>
              {expandedCat === 'ignore' && (
                <motion.div initial={{ height: 0 }} animate={{ height: 'auto' }} exit={{ height: 0 }} className="overflow-hidden">
                  <div className="px-6 pb-4 pt-2 space-y-3">
                    <div className="flex gap-2">
                      <button
                        onClick={async () => {
                          await fetchIgnoreFile();
                          setIgnoreEditorOpen(true);
                        }}
                        className="flex-1 py-1.5 bg-[rgba(255,255,255,0.03)] hover:bg-[rgba(255,255,255,0.06)] border border-[rgba(255,255,255,0.06)] hover:border-[rgba(255,255,255,0.12)] text-[#f4f4f5] rounded text-[10px] font-bold uppercase tracking-wider transition-colors flex items-center justify-center gap-1.5 cursor-pointer"
                      >
                        <span className="material-symbols-outlined text-[12px]">edit</span>
                        {t('SimpleDashboardPage.runway.ignoreFile')}
                      </button>
                      <button
                        onClick={handleClearAllIgnored}
                        disabled={ignoredFindings.length === 0}
                        className="flex-1 py-1.5 bg-[rgba(239,68,68,0.06)] border border-[rgba(239,68,68,0.12)] hover:bg-[rgba(239,68,68,0.12)] text-[#ef4444] rounded text-[10px] font-bold uppercase tracking-wider transition-colors flex items-center justify-center gap-1.5 disabled:opacity-30 disabled:cursor-not-allowed cursor-pointer"
                      >
                        <span className="material-symbols-outlined text-[12px]">delete_sweep</span>
                        {t('SimpleDashboardPage.runway.clearSuppressions')}
                      </button>
                    </div>
                    {loadingIgnored ? (
                      <div className="text-[10px] text-[#71717a]">{t('SimpleDashboardPage.runway.loading')}</div>
                    ) : ignoredFindings.length === 0 ? (
                      <div className="text-[10px] text-[#71717a]">{t('SimpleDashboardPage.runway.noSuppressedFindings')}</div>
                    ) : (
                      <div className="space-y-2 max-h-60 overflow-y-auto pr-2" style={{ scrollbarWidth: 'thin', scrollbarColor: 'rgba(255,255,255,0.06) transparent' }}>
                        {ignoredFindings.map((f: any, i: number) => (
                          <div key={i} className="bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.04)] p-2.5 rounded group/ignore">
                            <div className="flex justify-between items-start mb-1">
                              <div className="text-[10px] font-mono text-[#e4e4e7] truncate flex-1">{f.ruleId}</div>
                              <button
                                onClick={async (e) => {
                                  e.stopPropagation();
                                  try {
                                    await fetch(`/api/securecoder/ignored?vulnId=${encodeURIComponent(f.vulnId)}`, { method: 'DELETE' });
                                    fetchIgnored();
                                  } catch (err) {
                                    console.error(err);
                                  }
                                }}
                                className="text-[9px] text-[#71717a] hover:text-red-400 transition-colors ml-1 uppercase border border-[rgba(255,255,255,0.04)] px-1.5 py-0.5 rounded cursor-pointer"
                              >
                                {t('SimpleDashboardPage.runway.restore')}
                              </button>
                            </div>
                            <div className="text-[9px] text-[#71717a] font-mono truncate">{f.filePath}:{f.lineNumber}</div>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                </motion.div>
              )}
            </AnimatePresence>
          </div>

          {/* Config / Settings */}
          <div>
            <button onClick={() => setExpandedCat(expandedCat === 'config' ? null : 'config')} className="w-full flex items-center gap-3 px-6 py-3 hover:bg-[rgba(255,255,255,0.02)] transition-colors group">
              <span className={`material-symbols-outlined text-[16px] transition-colors ${expandedCat === 'config' ? 'text-[var(--accent-color)]' : 'text-[#3f3f46] group-hover:text-[var(--accent-color)]'}`}>settings</span>
              <span className="text-[11px] font-semibold tracking-wider text-[#a1a1aa] uppercase flex-1 text-left">{t('SimpleDashboardPage.runway.configurationSettings')}</span>
              <span className={`material-symbols-outlined text-[14px] text-[#3f3f46] transition-transform duration-200 ${expandedCat === 'config' ? 'rotate-180 text-[var(--accent-color)]' : ''}`}>expand_more</span>
            </button>
            <AnimatePresence>
              {expandedCat === 'config' && (
                <motion.div initial={{ height: 0 }} animate={{ height: 'auto' }} exit={{ height: 0 }} className="overflow-hidden">
                  <div className="px-6 pb-4 pt-2 space-y-4">
                    {configLoading ? (
                      <div className="text-[10px] text-[#71717a]">{t('SimpleDashboardPage.runway.loadingConfig')}</div>
                    ) : (
                      <>
                        {configError && (
                          <div className="p-2 bg-[rgba(239,68,68,0.08)] border border-[rgba(239,68,68,0.15)] rounded text-[10px] text-[#ef4444]">{configError}</div>
                        )}
                        {configSuccess && (
                          <div className="p-2 bg-[rgba(34,197,94,0.08)] border border-[rgba(34,197,94,0.15)] rounded text-[10px] text-[#22c55e]">{t('SimpleDashboardPage.runway.configSavedSuccess')}</div>
                        )}
                        
                        {/* Enabled Switch */}
                        <div className="flex items-center justify-between">
                          <label className="text-[11px] text-[#a1a1aa] font-medium">{t('SimpleDashboardPage.runway.enableIntegration')}</label>
                          <input type="checkbox" checked={configEnabled} onChange={e => setConfigEnabled(e.target.checked)} className="accent-[var(--accent-color)] cursor-pointer" />
                        </div>

                        {/* Scanner Backend */}
                        <div className="space-y-1">
                          <label className="text-[10px] text-[#71717a] font-bold uppercase tracking-wider">{t('SimpleDashboardPage.runway.scannerBackend')}</label>
                          <select value={configScannerBackend} onChange={e => setConfigScannerBackend(e.target.value)} className="w-full bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.06)] rounded px-3 py-1.5 text-[11px] text-white outline-none focus:border-[var(--accent-color)] cursor-pointer">
                            <option value="semgrep">{t('SimpleDashboardPage.runway.scannerSemgrep')}</option>
                            <option value="wiz">{t('SimpleDashboardPage.runway.scannerWiz')}</option>
                            <option value="aitriage">{t('SimpleDashboardPage.runway.scannerAitriage')}</option>
                          </select>
                        </div>

                        {/* Wiz Authentication Details (Only if Wiz is selected) */}
                        {configScannerBackend === 'wiz' && (
                          <div className="border border-[rgba(255,255,255,0.06)] bg-[rgba(255,255,255,0.015)] rounded-lg p-3 space-y-3">
                            <div className="flex items-center justify-between">
                              <span className="text-[10px] font-bold text-[#71717a] uppercase tracking-wider">{t('SimpleDashboardPage.runway.wizCliAuth')}</span>
                              {wizAuthLoading ? (
                                <span className="text-[9px] text-[#71717a]">{t('SimpleDashboardPage.runway.checking')}</span>
                              ) : wizStatus?.authenticated ? (
                                <span className="px-2 py-0.5 bg-[rgba(34,197,94,0.1)] border border-[rgba(34,197,94,0.2)] text-[#22c55e] text-[9px] font-bold rounded uppercase">{t('SimpleDashboardPage.runway.authorized')}</span>
                              ) : (
                                <span className="px-2 py-0.5 bg-[rgba(239,68,68,0.1)] border border-[rgba(239,68,68,0.2)] text-[#ef4444] text-[9px] font-bold rounded uppercase">{t('SimpleDashboardPage.runway.noAuth')}</span>
                              )}
                            </div>

                            {wizStatus?.authenticated ? (
                              <div className="space-y-2">
                                <div className="text-[10px] text-[#a1a1aa] leading-relaxed">
                                  {t('SimpleDashboardPage.runway.expiresIn', { hours: wizStatus.hoursRemaining })}
                                </div>
                                <button
                                  onClick={handleWizLogout}
                                  className="w-full py-1.5 bg-[rgba(239,68,68,0.08)] hover:bg-[rgba(239,68,68,0.15)] border border-[rgba(239,68,68,0.15)] text-[#ef4444] rounded text-[10px] font-bold uppercase tracking-wider transition-colors cursor-pointer"
                                >
                                  {t('SimpleDashboardPage.runway.disconnectWiz')}
                                </button>
                              </div>
                            ) : (
                              <div className="space-y-2">
                                {wizLoginSession ? (
                                  <div className="space-y-2.5 p-2.5 bg-[rgba(0,0,0,0.3)] border border-[rgba(255,255,255,0.04)] rounded text-[10px]">
                                    {wizLoginSession.status === 'starting' && (
                                      <div className="text-[#a1a1aa] flex items-center gap-2">
                                        <span className="w-2.5 h-2.5 border-2 border-white/20 border-t-white rounded-full animate-spin"></span>
                                        {t('SimpleDashboardPage.runway.initializingCli')}
                                      </div>
                                    )}
                                    {wizLoginSession.status === 'prompt' && (
                                      <div className="space-y-2">
                                        <div className="text-[#e4e4e7] font-semibold text-[10px] uppercase">{t('SimpleDashboardPage.runway.deviceVerificationCode')}</div>
                                        <div className="bg-black/60 border border-white/10 rounded px-3 py-1.5 text-center font-mono text-[14px] text-sky-400 font-bold select-all tracking-wider">
                                          {wizLoginSession.userCode}
                                        </div>
                                        <div className="text-[#71717a] leading-normal text-[10px]">
                                          {t('SimpleDashboardPage.runway.goToAuthPage')}
                                        </div>
                                        <a
                                          href={wizLoginSession.verificationUrl}
                                          target="_blank"
                                          rel="noopener noreferrer"
                                          className="block text-center py-1.5 bg-[var(--accent-color)] hover:bg-[var(--accent-color-hover)] text-[var(--accent-color-on-text)] rounded text-[10px] font-bold uppercase tracking-wider transition-colors"
                                        >
                                          {t('SimpleDashboardPage.runway.openVerificationLink')}
                                        </a>
                                      </div>
                                    )}
                                    {wizLoginSession.status === 'failed' && (
                                      <div className="text-red-400 text-[10px]">
                                        {t('SimpleDashboardPage.runway.errorPrefix')}{wizLoginSession.error || t('SimpleDashboardPage.runway.authAborted')}
                                      </div>
                                    )}
                                    {wizLoginSession.status === 'success' && (
                                      <div className="text-[#22c55e] text-[10px] font-bold">
                                        {t('SimpleDashboardPage.runway.wizCliAuthenticated')}
                                      </div>
                                    )}
                                  </div>
                                ) : (
                                  <button
                                    onClick={handleWizStartLogin}
                                    className="w-full py-1.5 bg-[var(--accent-color)] hover:bg-[var(--accent-color-hover)] text-[var(--accent-color-on-text)] rounded text-[10px] font-bold uppercase tracking-wider transition-colors cursor-pointer"
                                  >
                                    {t('SimpleDashboardPage.runway.authenticateCli')}
                                  </button>
                                )}
                              </div>
                            )}
                          </div>
                        )}

                        {/* Scan Mode / Rule Set */}
                        <div className="space-y-1">
                          <label className="text-[10px] text-[#71717a] font-bold uppercase tracking-wider">{t('SimpleDashboardPage.runway.scanMode')}</label>
                          <select value={configRuleSet} onChange={e => setConfigRuleSet(e.target.value)} className="w-full bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.06)] rounded px-3 py-1.5 text-[11px] text-white outline-none focus:border-[var(--accent-color)] cursor-pointer">
                            <option value="fast">{t('SimpleDashboardPage.runway.scanModeFast')}</option>
                            <option value="all">{t('SimpleDashboardPage.runway.scanModeAll')}</option>
                          </select>
                        </div>

                        {/* Ignore Mode */}
                        <div className="space-y-1">
                          <label className="text-[10px] text-[#71717a] font-bold uppercase tracking-wider">{t('SimpleDashboardPage.runway.ignoreMode')}</label>
                          <select value={configIgnoreMode} onChange={e => setConfigIgnoreMode(e.target.value)} className="w-full bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.06)] rounded px-3 py-1.5 text-[11px] text-white outline-none focus:border-[var(--accent-color)] cursor-pointer">
                            <option value="workspace">{t('SimpleDashboardPage.runway.ignoreModeWorkspace')}</option>
                            <option value="comment">{t('SimpleDashboardPage.runway.ignoreModeComment')}</option>
                          </select>
                        </div>

                        {/* Autostart Fixes */}
                        <div className="flex items-center justify-between">
                          <label className="text-[11px] text-[#a1a1aa] font-medium">{t('SimpleDashboardPage.runway.autostartFixes')}</label>
                          <input type="checkbox" checked={configAutostartFixes} onChange={e => setConfigAutostartFixes(e.target.checked)} className="accent-[var(--accent-color)] cursor-pointer" />
                        </div>

                        {/* Debug Mode */}
                        <div className="flex items-center justify-between">
                          <label className="text-[11px] text-[#a1a1aa] font-medium">{t('SimpleDashboardPage.runway.debugMode')}</label>
                          <input type="checkbox" checked={configDebug} onChange={e => setConfigDebug(e.target.checked)} className="accent-[var(--accent-color)] cursor-pointer" />
                        </div>

                        <div className="flex gap-2 pt-1">
                          <button
                            onClick={() => {
                              setOnboardingStep(0);
                              setOnboardingOpen(true);
                            }}
                            className="flex-1 px-3 py-2 bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.06)] hover:bg-[rgba(255,255,255,0.06)] text-[#f4f4f5] rounded text-[10px] font-bold uppercase tracking-wider transition-colors cursor-pointer flex items-center justify-center gap-1.5"
                          >
                            <span className="material-symbols-outlined text-[12px]">explore</span>
                            {t('SimpleDashboardPage.runway.onboarding')}
                          </button>
                          <button onClick={() => handleSaveConfig()} disabled={configSaving} className="flex-[2] px-4 py-2 bg-[var(--accent-color)] hover:bg-[var(--accent-color-hover)] text-[var(--accent-color-on-text)] rounded text-[10px] font-bold uppercase tracking-wider transition-colors disabled:opacity-50 cursor-pointer">
                            {configSaving ? t('SimpleDashboardPage.runway.saving') : t('SimpleDashboardPage.runway.saveConfiguration')}
                          </button>
                        </div>
                      </>
                    )}
                  </div>
                </motion.div>
              )}
            </AnimatePresence>
          </div>
        </div>
      )}

      {/* ── MODALS ── */}

      {/* Onboarding slideshow modal */}
      <AnimatePresence>
        {onboardingOpen && (
          <div className="fixed inset-0 bg-black/70 backdrop-blur-sm z-[100] flex items-center justify-center p-4">
            <motion.div
              initial={{ opacity: 0, scale: 0.95 }}
              animate={{ opacity: 1, scale: 1 }}
              exit={{ opacity: 0, scale: 0.95 }}
              className="bg-[#0e0e11] border border-[rgba(255,255,255,0.08)] rounded-xl w-[480px] p-6 shadow-[0_24px_50px_rgba(0,0,0,0.85)] flex flex-col space-y-4 text-left relative overflow-hidden"
            >
              {/* Step indicator */}
              <div className="flex justify-between items-center text-[10px] font-bold text-[#71717a] tracking-wider uppercase">
                <span>{t('SimpleDashboardPage.runway.setupTitle')}</span>
                <span>{t('SimpleDashboardPage.runway.setupStepIndicator', { current: onboardingStep + 1, total: 4 })}</span>
              </div>

              {/* Step 0: Welcome */}
              {onboardingStep === 0 && (
                <div className="space-y-3">
                  <div className="w-10 h-10 rounded-lg bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.06)] flex items-center justify-center text-sky-400">
                    <span className="material-symbols-outlined text-[24px]">shield</span>
                  </div>
                  <h3 className="text-[14px] font-bold text-white uppercase tracking-wider">{t('SimpleDashboardPage.runway.setupWelcomeTitle')}</h3>
                  <p className="text-[11px] text-[#a1a1aa] leading-relaxed">
                    {t('SimpleDashboardPage.runway.setupWelcomeDesc')}
                  </p>
                </div>
              )}

              {/* Step 1: How it Works */}
              {onboardingStep === 1 && (
                <div className="space-y-3">
                  <h3 className="text-[14px] font-bold text-white uppercase tracking-wider">{t('SimpleDashboardPage.runway.setupFeaturesTitle')}</h3>
                  <div className="space-y-2 text-[11px] text-[#a1a1aa]">
                    <div className="flex items-start gap-2.5">
                      <span className="material-symbols-outlined text-[14px] text-sky-400 mt-0.5">bolt</span>
                      <div>
                        <strong className="text-white">{t('SimpleDashboardPage.runway.setupFeatureRunwayTitle')}</strong> {t('SimpleDashboardPage.runway.setupFeatureRunwayDesc')}
                      </div>
                    </div>
                    <div className="flex items-start gap-2.5">
                      <span className="material-symbols-outlined text-[14px] text-sky-400 mt-0.5">package_2</span>
                      <div>
                        <strong className="text-white">{t('SimpleDashboardPage.runway.setupFeatureDepTitle')}</strong> {t('SimpleDashboardPage.runway.setupFeatureDepDesc')}
                      </div>
                    </div>
                    <div className="flex items-start gap-2.5">
                      <span className="material-symbols-outlined text-[14px] text-sky-400 mt-0.5">sync</span>
                      <div>
                        <strong className="text-white">{t('SimpleDashboardPage.runway.setupFeatureIdeTitle')}</strong> {t('SimpleDashboardPage.runway.setupFeatureIdeDesc')}
                      </div>
                    </div>
                  </div>
                </div>
              )}

              {/* Step 2: Scanner Configuration */}
              {onboardingStep === 2 && (
                <div className="space-y-3">
                  <h3 className="text-[14px] font-bold text-white uppercase tracking-wider">{t('SimpleDashboardPage.runway.setupScannerBackendTitle')}</h3>
                  <p className="text-[11px] text-[#71717a]">
                    {t('SimpleDashboardPage.runway.setupScannerBackendDesc')}
                  </p>
                  <div className="space-y-2">
                    <div
                      onClick={() => setConfigScannerBackend('semgrep')}
                      className={`p-3 rounded-lg border cursor-pointer transition-colors flex items-center justify-between ${
                        configScannerBackend === 'semgrep'
                          ? 'bg-[rgba(255,255,255,0.03)] border-[var(--accent-color)] text-white'
                          : 'bg-transparent border-[rgba(255,255,255,0.06)] text-[#a1a1aa] hover:border-[rgba(255,255,255,0.12)]'
                      }`}
                    >
                      <div className="text-left">
                        <div className="text-[11px] font-bold uppercase tracking-wider">{t('SimpleDashboardPage.runway.scannerSemgrep')}</div>
                        <div className="text-[10px] text-[#71717a] mt-0.5">{t('SimpleDashboardPage.runway.setupSemgrepDesc')}</div>
                      </div>
                      {configScannerBackend === 'semgrep' && <span className="material-symbols-outlined text-[16px] text-[var(--accent-color)]">check_circle</span>}
                    </div>

                    <div
                      onClick={() => setConfigScannerBackend('wiz')}
                      className={`p-3 rounded-lg border cursor-pointer transition-colors flex items-center justify-between ${
                        configScannerBackend === 'wiz'
                          ? 'bg-[rgba(255,255,255,0.03)] border-[var(--accent-color)] text-white'
                          : 'bg-transparent border-[rgba(255,255,255,0.06)] text-[#a1a1aa] hover:border-[rgba(255,255,255,0.12)]'
                      }`}
                    >
                      <div className="text-left">
                        <div className="text-[11px] font-bold uppercase tracking-wider flex items-center gap-1.5">
                          {t('SimpleDashboardPage.runway.setupWizTitle')}
                          <span className="px-1.5 py-0.5 bg-sky-950 border border-sky-800 text-sky-400 text-[8px] font-bold rounded uppercase">{t('SimpleDashboardPage.runway.setupWizTag')}</span>
                        </div>
                        <div className="text-[10px] text-[#71717a] mt-0.5">{t('SimpleDashboardPage.runway.setupWizDesc')}</div>
                      </div>
                      {configScannerBackend === 'wiz' && <span className="material-symbols-outlined text-[16px] text-[var(--accent-color)]">check_circle</span>}
                    </div>
                  </div>

                  {configScannerBackend === 'wiz' && (
                    <div className="flex items-start gap-2 pt-1">
                      <input
                        type="checkbox"
                        id="wiz-agreement"
                        checked={wizAgreementChecked}
                        onChange={e => setWizAgreementChecked(e.target.checked)}
                        className="mt-0.5 accent-[var(--accent-color)] cursor-pointer"
                      />
                      <label htmlFor="wiz-agreement" className="text-[9px] text-[#71717a] leading-normal cursor-pointer select-none">
                        {t('SimpleDashboardPage.runway.setupWizAgreement')}
                      </label>
                    </div>
                  )}
                </div>
              )}

              {/* Step 3: Ignore patterns setup */}
              {onboardingStep === 3 && (
                <div className="space-y-3">
                  <h3 className="text-[14px] font-bold text-white uppercase tracking-wider">{t('SimpleDashboardPage.runway.setupIgnoreTitle')}</h3>
                  <p className="text-[11px] text-[#71717a]">
                    {t('SimpleDashboardPage.runway.setupIgnoreDesc')}
                  </p>
                  <textarea
                    value={onboardingIgnoreContent}
                    onChange={e => setOnboardingIgnoreContent(e.target.value)}
                    className="bg-[#08080a] border border-[rgba(255,255,255,0.06)] rounded p-3 text-[11px] text-[#e4e4e7] font-mono outline-none focus:border-[var(--accent-color)] w-full h-32 resize-none"
                    style={{ scrollbarWidth: 'thin', scrollbarColor: 'rgba(255,255,255,0.06) transparent' }}
                  />
                </div>
              )}

              {/* Navigation controls */}
              <div className="flex justify-between items-center pt-2 border-t border-[rgba(255,255,255,0.06)]">
                <button
                  onClick={() => setOnboardingOpen(false)}
                  className="px-3.5 py-1.5 bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.06)] hover:bg-[rgba(255,255,255,0.05)] text-[#a1a1aa] hover:text-white rounded-lg text-[10px] font-bold uppercase tracking-wider transition-colors cursor-pointer"
                >
                  {t('SimpleDashboardPage.runway.cancel')}
                </button>
                <div className="flex gap-2">
                  {onboardingStep > 0 && (
                    <button
                      onClick={() => setOnboardingStep(prev => prev - 1)}
                      className="px-3.5 py-1.5 bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.06)] hover:bg-[rgba(255,255,255,0.05)] text-[#f4f4f5] rounded-lg text-[10px] font-bold uppercase tracking-wider transition-colors cursor-pointer"
                    >
                      {t('SimpleDashboardPage.runway.back')}
                    </button>
                  )}
                  {onboardingStep < 3 ? (
                    <button
                      onClick={() => {
                        if (onboardingStep === 2 && configScannerBackend === 'wiz' && !wizAgreementChecked) {
                          alert(t('SimpleDashboardPage.runway.setupWizAgreementAlert'));
                          return;
                        }
                        setOnboardingStep(prev => prev + 1);
                      }}
                      className="px-4 py-1.5 bg-[var(--accent-color)] hover:bg-[var(--accent-color-hover)] text-[var(--accent-color-on-text)] rounded-lg text-[10px] font-bold uppercase tracking-wider transition-colors cursor-pointer"
                    >
                      {t('SimpleDashboardPage.runway.next')}
                    </button>
                  ) : (
                    <button
                      onClick={async () => {
                        await handleSaveIgnoreFile(onboardingIgnoreContent);
                        await handleSaveConfig({
                          enabled: true,
                          scannerBackend: configScannerBackend,
                          ruleSet: configRuleSet,
                          autostartFixes: configAutostartFixes,
                          ignoreMode: configIgnoreMode,
                          debug: configDebug
                        });
                        setOnboardingOpen(false);
                      }}
                      className="px-4 py-1.5 bg-[#22c55e] hover:bg-[#16a34a] text-white rounded-lg text-[10px] font-bold uppercase tracking-wider transition-colors cursor-pointer"
                    >
                      {t('SimpleDashboardPage.runway.finishAndSave')}
                    </button>
                  )}
                </div>
              </div>
            </motion.div>
          </div>
        )}
      </AnimatePresence>

      {/* Ignore File Editor Modal */}
      <AnimatePresence>
        {ignoreEditorOpen && (
          <div className="fixed inset-0 bg-black/70 backdrop-blur-sm z-[100] flex items-center justify-center p-4">
            <motion.div
              initial={{ opacity: 0, scale: 0.95 }}
              animate={{ opacity: 1, scale: 1 }}
              exit={{ opacity: 0, scale: 0.95 }}
              className="bg-[#0e0e11] border border-[rgba(255,255,255,0.08)] rounded-xl w-[500px] p-6 shadow-[0_24px_50px_rgba(0,0,0,0.85)] flex flex-col space-y-4 text-left relative overflow-hidden"
            >
              <div className="flex justify-between items-center">
                <h3 className="text-[12px] font-bold text-white uppercase tracking-wider">{t('SimpleDashboardPage.runway.editIgnoreTitle')}</h3>
                <span className="px-2 py-0.5 bg-zinc-900 border border-zinc-800 text-zinc-500 text-[8px] font-bold rounded uppercase">{t('SimpleDashboardPage.runway.workspaceIgnoreTag')}</span>
              </div>
              <p className="text-[11px] text-[#71717a] leading-normal">
                {t('SimpleDashboardPage.runway.editIgnoreDesc')}
              </p>
              <textarea
                value={ignoreEditorContent}
                onChange={e => setIgnoreEditorContent(e.target.value)}
                className="bg-[#08080a] border border-[rgba(255,255,255,0.06)] rounded p-3.5 text-[11px] text-[#e4e4e7] font-mono outline-none focus:border-[var(--accent-color)] w-full h-48 resize-none"
                style={{ scrollbarWidth: 'thin', scrollbarColor: 'rgba(255,255,255,0.06) transparent' }}
              />
              <div className="flex justify-end gap-2 pt-2 border-t border-[rgba(255,255,255,0.06)]">
                <button
                  onClick={() => setIgnoreEditorOpen(false)}
                  className="px-4 py-1.5 bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.06)] hover:bg-[rgba(255,255,255,0.05)] text-[#a1a1aa] hover:text-white rounded-lg text-[10px] font-bold uppercase tracking-wider transition-colors cursor-pointer"
                >
                  {t('SimpleDashboardPage.runway.cancel')}
                </button>
                <button
                  onClick={async () => {
                    await handleSaveIgnoreFile(ignoreEditorContent);
                    setIgnoreEditorOpen(false);
                  }}
                  disabled={ignoreEditorSaving}
                  className="px-4 py-1.5 bg-[var(--accent-color)] hover:bg-[var(--accent-color-hover)] text-[var(--accent-color-on-text)] rounded-lg text-[10px] font-bold uppercase tracking-wider transition-colors disabled:opacity-50 cursor-pointer"
                >
                  {ignoreEditorSaving ? t('SimpleDashboardPage.runway.saving') : t('SimpleDashboardPage.runway.saveIgnoreFile')}
                </button>
              </div>
            </motion.div>
          </div>
        )}
      </AnimatePresence>
    </motion.div>
  );
};

interface InlineAIPromptsProps {
  activeProducts: any[];
  findings: any[];
}

const InlineAIPrompts: React.FC<InlineAIPromptsProps> = ({ activeProducts, findings }) => {
  const { t } = useTranslation('components');
  const [expandedCat, setExpandedCat] = useState<string | null>(null);
  const [copiedId, setCopiedId] = useState<string | null>(null);
  const [previewId, setPreviewId] = useState<string | null>(null);

  // Run in UI state
  const [runProjectId, setRunProjectId] = useState<Record<string, number>>({});
  const [activeSelectPromptId, setActiveSelectPromptId] = useState<string | null>(null);
  const [runOutput, setRunOutput] = useState<Record<string, string>>({});
  const [runLoading, setRunLoading] = useState<Record<string, boolean>>({});
  const [runError, setRunError] = useState<Record<string, string>>({});

  const handleCopy = useCallback((promptKey: string, id: string) => {
    const text = t(promptKey);
    navigator.clipboard.writeText(text).then(() => {
      setCopiedId(id);
      setTimeout(() => setCopiedId(null), 2000);
    });
  }, [t]);

  const handleOpenIDE = useCallback((promptKey: string, ide: string) => {
    const text = t(promptKey);
    navigator.clipboard.writeText(text).then(() => {
      const scheme = ide === 'Cursor' ? 'cursor://' : 'windsurf://';
      window.location.href = scheme;
    });
  }, [t]);

  const handleRunInUI = async (promptId: string, promptKey: string) => {
    const prodId = runProjectId[promptId];
    if (!prodId) return;
    const proj = activeProducts.find(p => p.id === prodId);
    if (!proj) return;

    setRunLoading(prev => ({ ...prev, [promptId]: true }));
    setRunError(prev => ({ ...prev, [promptId]: '' }));
    setRunOutput(prev => ({ ...prev, [promptId]: '' }));

    try {
      const projectFindings = findings.filter(f => f.product_id === prodId && f.status !== 'triage');
      const findingsText = projectFindings.map(f => `- [${f.severity?.toUpperCase()}] ${f.title} | File: ${f.file_path || 'N/A'}${f.line_number ? ':' + f.line_number : ''}`).join('\n');
      
      const basePrompt = t(promptKey);
      const fullPrompt = `Analyze the project "${proj.name}" using the following guidelines:\n\n${basePrompt}\n\nProject findings context:\n${findingsText}\n\nPerform this security analysis and output in structured markdown format.`;

      const res = await fetch('/api/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ messages: [{ role: 'user', content: fullPrompt }] })
      });
      const data = await res.json();
      if (data.ok) {
        setRunOutput(prev => ({ ...prev, [promptId]: data.content }));
      } else {
        setRunError(prev => ({ ...prev, [promptId]: data.error || 'Failed to run analysis.' }));
      }
    } catch (e) {
      setRunError(prev => ({ ...prev, [promptId]: 'Network error occurred.' }));
    } finally {
      setRunLoading(prev => ({ ...prev, [promptId]: false }));
    }
  };

  return (
    <motion.div variants={itemVariants} className="border border-[rgba(255,255,255,0.06)] rounded-lg bg-[rgba(255,255,255,0.01)] overflow-hidden">
      {/* Header */}
      <div className="flex items-center gap-3 px-6 py-4 border-b border-[rgba(255,255,255,0.06)]">
        <div className="w-8 h-8 rounded-lg bg-[rgba(255,255,255,0.04)] border border-[rgba(255,255,255,0.08)] flex items-center justify-center">
          <span className="material-symbols-outlined text-[16px] text-[#a1a1aa]" style={{ fontVariationSettings: "'FILL' 1" }}>terminal</span>
        </div>
        <div className="flex-1">
          <h3 className="text-[12px] font-bold text-[#f4f4f5] tracking-wider uppercase">{t('aiPrompts.title')}</h3>
          <p className="text-[10px] text-[#52525b] tracking-widest uppercase">EXECUTE AGENT SKILLS INTERACTIVELY</p>
        </div>
        <div className="flex items-center gap-1.5">
          <span className="text-[9px] text-[#3f3f46] tracking-wider font-medium">RUNS DIRECTLY IN WEB UI</span>
        </div>
      </div>

      {/* Categories */}
      <div className="divide-y divide-[rgba(255,255,255,0.04)]">
        {INLINE_PROMPT_CATEGORIES.map((cat) => {
          const isOpen = expandedCat === cat.id;
          return (
            <div key={cat.id}>
              <button
                onClick={() => setExpandedCat(isOpen ? null : cat.id)}
                className="w-full flex items-center gap-3 px-6 py-3 hover:bg-[rgba(255,255,255,0.02)] transition-colors group"
              >
                <span className={`material-symbols-outlined text-[16px] transition-colors ${isOpen ? 'text-[var(--accent-color)]' : 'text-[#3f3f46] group-hover:text-[var(--accent-color)]'}`}>{cat.icon}</span>
                <span className={`text-[11px] font-semibold tracking-wider uppercase flex-1 text-left transition-colors ${isOpen ? 'text-[#f4f4f5]' : 'text-[#a1a1aa] group-hover:text-[#f4f4f5]'}`}>{t(cat.labelKey)}</span>
                <span className={`text-[10px] mr-1 transition-colors ${isOpen ? 'text-[var(--accent-color)]' : 'text-[#3f3f46] group-hover:text-[var(--accent-color)]'}`}>{cat.prompts.length}</span>
                <span className={`material-symbols-outlined text-[14px] text-[#3f3f46] transition-transform duration-200 ${isOpen ? 'rotate-180 text-[var(--accent-color)]' : 'group-hover:text-[var(--accent-color)]'}`}>expand_more</span>
              </button>

              <AnimatePresence>
                {isOpen && (
                  <motion.div
                    initial={{ height: 0, opacity: 0 }}
                    animate={{ height: 'auto', opacity: 1 }}
                    exit={{ height: 0, opacity: 0 }}
                    transition={{ duration: 0.2, ease: 'easeInOut' }}
                    className="overflow-hidden"
                  >
                    <div className="px-6 pb-4 grid gap-2.5" style={{ gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))' }}>
                      {cat.prompts.map((prompt) => {
                        const isCopied = copiedId === prompt.id;
                        const isPreviewing = previewId === prompt.id;
                        const isLoading = runLoading[prompt.id];
                        const error = runError[prompt.id];
                        const output = runOutput[prompt.id];
                        const selectedProj = runProjectId[prompt.id];

                        return (
                          <div key={prompt.id} className="rounded-lg border border-[rgba(255,255,255,0.06)] bg-[rgba(255,255,255,0.015)] hover:border-[rgba(255,255,255,0.1)] transition-all p-3.5 group/card">
                            <div className="text-[12px] font-medium text-[#e4e4e7] mb-0.5">{t(prompt.titleKey)}</div>
                            <div className="text-[10px] text-[#52525b] mb-2 leading-relaxed">{t(prompt.descKey)}</div>
                            
                            {/* Methodology Features */}
                            {prompt.features && prompt.features.length > 0 && (
                              <div className="flex items-center gap-1 flex-wrap mb-2.5">
                                {prompt.features.map((feature: string) => (
                                  <span 
                                    key={feature} 
                                    className="text-[7.5px] px-1.5 py-[3px] rounded-sm font-mono font-bold uppercase tracking-wider leading-none"
                                    style={{
                                      color: 'var(--accent-color, #8b5cf6)',
                                      backgroundColor: 'color-mix(in srgb, var(--accent-color, #8b5cf6) 8%, transparent)',
                                      border: '1px solid color-mix(in srgb, var(--accent-color, #8b5cf6) 15%, transparent)',
                                    }}
                                  >
                                    {feature}
                                  </span>
                                ))}
                              </div>
                            )}

                            {/* Select Project Inline for Execution */}
                            <div className="flex gap-2 mb-3 relative">
                              <div className="flex-1 relative">
                                <button
                                  type="button"
                                  onClick={(e) => {
                                    e.stopPropagation();
                                    setActiveSelectPromptId(prev => prev === prompt.id ? null : prompt.id);
                                  }}
                                  className="w-full h-9 bg-[rgba(255,255,255,0.015)] hover:bg-[rgba(255,255,255,0.04)] border border-[rgba(255,255,255,0.06)] hover:border-[rgba(255,255,255,0.12)] rounded-lg px-3 text-[12px] text-[#a1a1aa] hover:text-white flex items-center justify-between transition-all cursor-pointer outline-none focus:border-[var(--accent-color)]"
                                >
                                  <span className="truncate">
                                    {selectedProj && activeProducts.find(p => p.id === selectedProj)
                                      ? activeProducts.find(p => p.id === selectedProj)!.name
                                      : 'Select Project...'}
                                  </span>
                                  <span className={`material-symbols-outlined text-[16px] text-[#71717a] transition-transform ${activeSelectPromptId === prompt.id ? 'rotate-180' : ''}`}>
                                    keyboard_arrow_down
                                  </span>
                                </button>

                                <AnimatePresence>
                                  {activeSelectPromptId === prompt.id && (
                                    <>
                                      <div 
                                        className="fixed inset-0 z-40 cursor-default" 
                                        onClick={() => setActiveSelectPromptId(null)} 
                                      />
                                      <motion.div
                                        initial={{ opacity: 0, y: 4 }}
                                        animate={{ opacity: 1, y: 0 }}
                                        exit={{ opacity: 0, y: 4 }}
                                        transition={{ duration: 0.15 }}
                                        className="absolute left-0 right-0 mt-1.5 z-50 bg-[#161618] border border-[rgba(255,255,255,0.08)] rounded-lg shadow-[0_8px_30px_rgb(0,0,0,0.65)] backdrop-blur-xl py-1 max-h-48 overflow-y-auto"
                                        style={{ scrollbarWidth: 'thin', scrollbarColor: 'rgba(255,255,255,0.06) transparent' }}
                                      >
                                        <div
                                          onClick={() => {
                                            setRunProjectId(prev => ({ ...prev, [prompt.id]: 0 }));
                                            setActiveSelectPromptId(null);
                                          }}
                                          className={`px-3 py-2 text-[12px] cursor-pointer transition-colors flex items-center justify-between ${
                                            !selectedProj 
                                              ? 'bg-[var(--accent-color-soft)] text-white font-semibold' 
                                              : 'text-[#71717a] hover:text-[#e4e4e7] hover:bg-[rgba(255,255,255,0.03)]'
                                          }`}
                                        >
                                          <span>Select Project...</span>
                                          {!selectedProj && <span className="material-symbols-outlined text-xs text-[var(--accent-color)]">check</span>}
                                        </div>
                                        {activeProducts.map(p => {
                                          const isSelected = selectedProj === p.id;
                                          return (
                                            <div
                                              key={p.id}
                                              onClick={() => {
                                                setRunProjectId(prev => ({ ...prev, [prompt.id]: p.id }));
                                                setActiveSelectPromptId(null);
                                              }}
                                              className={`px-3 py-2 text-[12px] cursor-pointer transition-colors flex items-center justify-between ${
                                                isSelected 
                                                  ? 'bg-[var(--accent-color-soft)] text-white font-semibold' 
                                                  : 'text-[#a1a1aa] hover:text-white hover:bg-[rgba(255,255,255,0.03)]'
                                              }`}
                                            >
                                              <span className="truncate">{p.name}</span>
                                              {isSelected && <span className="material-symbols-outlined text-xs text-[var(--accent-color)]">check</span>}
                                            </div>
                                          );
                                        })}
                                      </motion.div>
                                    </>
                                  )}
                                </AnimatePresence>
                              </div>

                              <button
                                onClick={() => handleRunInUI(prompt.id, prompt.promptKey)}
                                disabled={!selectedProj || isLoading}
                                className="h-9 px-4 bg-[var(--accent-color)] hover:bg-[var(--accent-color-hover)] text-[var(--accent-color-on-text)] text-[12px] rounded-lg font-bold uppercase tracking-wider disabled:opacity-30 disabled:pointer-events-none transition-all flex items-center gap-1.5 shrink-0 shadow-md cursor-pointer"
                              >
                                {isLoading ? (
                                  <div className="w-3.5 h-3.5 border-2 border-[var(--accent-color-on-text)]/30 border-t-[var(--accent-color-on-text)] rounded-full animate-spin" />
                                ) : (
                                  <span className="material-symbols-outlined text-[14px]">play_arrow</span>
                                )}
                                Run
                              </button>
                            </div>

                            <div className="flex items-center gap-1.5 flex-wrap">
                              <button
                                onClick={() => handleCopy(prompt.promptKey, prompt.id)}
                                className={`flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg text-[11px] font-medium tracking-wide transition-all duration-200 cursor-pointer ${
                                  isCopied
                                    ? 'bg-[rgba(34,197,94,0.08)] border border-[rgba(34,197,94,0.2)] text-[#22c55e]'
                                    : 'bg-[rgba(255,255,255,0.04)] border border-[rgba(255,255,255,0.08)] text-[#a1a1aa] hover:bg-[rgba(255,255,255,0.06)] hover:text-[#f4f4f5]'
                                }`}
                                title={t('aiPrompts.copyPrompt')}
                              >
                                <span className="material-symbols-outlined text-[13px]">{isCopied ? 'check_circle' : 'content_copy'}</span>
                              </button>
                              
                              <button
                                onClick={() => handleOpenIDE(prompt.promptKey, 'Cursor')}
                                className="flex items-center gap-1 px-2.5 py-1.5 rounded-lg text-[11px] font-medium tracking-wide bg-[rgba(56,189,248,0.04)] border border-[rgba(56,189,248,0.15)] text-[#38bdf8] hover:bg-[rgba(56,189,248,0.1)] transition-all duration-200 cursor-pointer"
                                title="Copy & Open in Cursor"
                              >
                                <span className="material-symbols-outlined text-[13px]">open_in_new</span>
                                Cursor
                              </button>

                              <button
                                onClick={() => handleOpenIDE(prompt.promptKey, 'Windsurf')}
                                className="flex items-center gap-1 px-2.5 py-1.5 rounded-lg text-[11px] font-medium tracking-wide bg-[rgba(16,185,129,0.04)] border border-[rgba(16,185,129,0.15)] text-[#10b981] hover:bg-[rgba(16,185,129,0.1)] transition-all duration-200 cursor-pointer"
                                title="Copy & Open in Windsurf"
                              >
                                <span className="material-symbols-outlined text-[13px]">open_in_new</span>
                                Windsurf
                              </button>
                              
                              <button
                                onClick={() => setPreviewId(isPreviewing ? null : prompt.id)}
                                className="flex items-center gap-1 px-2.5 py-1.5 rounded-lg text-[11px] font-medium tracking-wide bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.06)] text-[#52525b] hover:text-[#a1a1aa] hover:bg-[rgba(255,255,255,0.04)] transition-all cursor-pointer"
                              >
                                <span className="material-symbols-outlined text-[13px]">{isPreviewing ? 'visibility_off' : 'visibility'}</span>
                                {isPreviewing ? t('aiPrompts.hide') : t('aiPrompts.preview')}
                              </button>
                            </div>

                            {/* Run Result Area */}
                            {isLoading && (
                              <div className="mt-3 flex flex-col items-center justify-center p-4 border border-[rgba(255,255,255,0.04)] bg-[rgba(0,0,0,0.15)] rounded-lg gap-2 text-[#71717a]">
                                <div className="w-4 h-4 border-2 border-[rgba(255,255,255,0.08)] border-t-[var(--accent-color)] rounded-full animate-spin" />
                                <span className="text-[9px] font-mono uppercase tracking-widest animate-pulse">Analyzing Project...</span>
                              </div>
                            )}

                            {error && (
                              <div className="mt-3 p-2.5 bg-[rgba(239,68,68,0.08)] border border-[rgba(239,68,68,0.15)] text-[#ef4444] text-[10px] rounded font-medium font-mono">
                                {error}
                              </div>
                            )}

                            {output && (
                              <div className="mt-3 bg-[rgba(0,0,0,0.3)] border border-[rgba(255,255,255,0.04)] rounded-lg p-3 max-h-56 overflow-y-auto text-[11px] text-[#a1a1aa] leading-relaxed prose prose-invert select-text" style={{ scrollbarWidth: 'thin' }}>
                                <div className="flex items-center justify-between border-b border-[rgba(255,255,255,0.05)] pb-1.5 mb-2">
                                  <span className="text-[9px] text-[var(--accent-color)] font-mono font-bold uppercase">Analysis Result</span>
                                  <button onClick={() => setRunOutput(prev => ({ ...prev, [prompt.id]: '' }))} className="text-[10px] text-[#52525b] hover:text-[#a1a1aa] material-symbols-outlined font-bold">close</button>
                                </div>
                                <Markdown>{output}</Markdown>
                              </div>
                            )}

                            <AnimatePresence>
                              {isPreviewing && (
                                <motion.div
                                  initial={{ height: 0, opacity: 0 }}
                                  animate={{ height: 'auto', opacity: 1 }}
                                  exit={{ height: 0, opacity: 0 }}
                                  transition={{ duration: 0.15 }}
                                  className="overflow-hidden"
                                >
                                  <pre className="mt-2.5 p-3 rounded-lg bg-[rgba(0,0,0,0.3)] border border-[rgba(255,255,255,0.04)] text-[10px] text-[#71717a] font-mono leading-relaxed whitespace-pre-wrap max-h-36 overflow-y-auto" style={{ scrollbarWidth: 'thin', scrollbarColor: 'rgba(255,255,255,0.06) transparent' }}>
                                    {t(prompt.promptKey)}
                                  </pre>
                                </motion.div>
                              )}
                            </AnimatePresence>
                          </div>
                        );
                      })}
                    </div>
                  </motion.div>
                )}
              </AnimatePresence>
            </div>
          );
        })}
      </div>

      {/* Footer */}
      <div className="px-6 py-2.5 border-t border-[rgba(255,255,255,0.04)] bg-[rgba(255,255,255,0.01)]">
        <div className="flex items-center gap-2 text-[9px] text-[#3f3f46]">
          <span className="material-symbols-outlined text-[12px]">info</span>
          <span>Analyze projects using pre-configured security agent prompts or sync metrics.</span>
        </div>
      </div>
    </motion.div>
  );
};

const containerVariants = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: {
      staggerChildren: 0.05
    }
  }
} as const;

const itemVariants = {
  hidden: { opacity: 0, y: 15 },
  visible: {
    opacity: 1,
    y: 0,
    transition: {
      type: "spring" as const,
      stiffness: 260,
      damping: 25
    }
  }
};

const sevDot = (sev: string) => {
  switch (sev?.toLowerCase()) {
    case 'critical': return '#ef4444';
    case 'high': return '#f97316';
    case 'medium': return '#eab308';
    case 'low': return '#52525b';
    default: return '#3f3f46';
  }
};

const FindingRow: React.FC<{
  f: Finding;
  isExpanded: boolean;
  onToggle: () => void;
  productMap: Map<number, Product>;
  setProductFilter: (id: number) => void;
  setPage: (p: number) => void;
  handleTriage: (f: Finding, action: string) => void;
  onNavigateToChat?: (f: Finding) => void;
  isSelected?: boolean;
  onToggleSelect?: (e: React.MouseEvent | React.ChangeEvent) => void;
  isTriaging?: boolean;
}> = ({ f, isExpanded, onToggle, productMap, setProductFilter, setPage, handleTriage, onNavigateToChat, isSelected, onToggleSelect, isTriaging }) => {
  const [hovered, setHovered] = useState(false);
  
  return (
    <div className="border-b border-[rgba(255,255,255,0.03)] last:border-b-0">
      <div 
        onMouseEnter={() => setHovered(true)}
        onMouseLeave={() => setHovered(false)}
        className="w-full px-4 py-2.5 flex items-center gap-3 hover:bg-[rgba(255,255,255,0.015)] border-l-2 transition-all duration-300 group"
        style={{ 
          borderLeftColor: isExpanded || hovered ? sevDot(f.severity) : 'transparent',
          boxShadow: isExpanded ? `inset 4px 0 12px ${sevDot(f.severity)}08` : 'none'
        }}
      >
        {onToggleSelect && (
          <input
            type="checkbox"
            checked={isSelected || false}
            onChange={onToggleSelect}
            onClick={e => e.stopPropagation()}
            className="accent-[var(--accent-color)] cursor-pointer select-checkbox file-checkbox w-3.5 h-3.5 rounded bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.08)]"
          />
        )}
        <button 
          onClick={onToggle}
          className="flex-1 text-left flex items-start gap-3 min-w-0"
        >
          <span 
            className="w-1.5 h-1.5 rounded-full shrink-0 mt-1.5 transition-all duration-300 group-hover:scale-125" 
            style={{ 
              backgroundColor: sevDot(f.severity),
              boxShadow: `0 0 6px ${sevDot(f.severity)}b0`
            }} 
          />
        <div className="flex-1 min-w-0">
          <div className={`text-[12px] font-medium leading-snug transition-colors duration-200 ${isExpanded ? 'text-white font-semibold' : 'text-[#d4d4d8] group-hover:text-white'}`}>
            {f.title}
          </div>
          <div className="flex items-center gap-2 mt-1 min-w-0">
            {f.file_path && (
              <span className="min-w-0 truncate text-[10px] text-[#52525b] font-mono group-hover:text-[#71717a] transition-colors">
                {f.file_path}{f.line_number ? `:${f.line_number}` : ''}
              </span>
            )}
            {f.product_id && productMap.has(f.product_id) && (
              <button 
                onClick={e => { e.stopPropagation(); setProductFilter(f.product_id!); setPage(0); }}
                className="shrink-0 whitespace-nowrap text-[9px] text-[#71717a] border border-[rgba(255,255,255,0.04)] rounded px-1.5 py-0.5 hover:text-[#a1a1aa] hover:border-[rgba(255,255,255,0.08)] bg-[rgba(255,255,255,0.01)] transition-colors font-mono uppercase tracking-wider"
                title={`Filter by ${productMap.get(f.product_id!)?.name}`}
              >
                {productMap.get(f.product_id!)?.name}
              </button>
            )}
            {f.stack && (
              <span className="shrink-0 whitespace-nowrap text-[9px] text-[#52525b] border border-[rgba(255,255,255,0.04)] rounded px-1.5 py-0.5 font-mono uppercase tracking-wider bg-[rgba(255,255,255,0.01)]">
                {f.stack}
              </span>
            )}
            {f.status && f.status !== 'open' && (
              <span className={`shrink-0 whitespace-nowrap text-[9px] px-1.5 py-0.5 rounded font-mono uppercase tracking-wider ${
                f.status === 'triage' 
                  ? 'text-[#38bdf8] bg-[rgba(56,189,248,0.06)] border border-[rgba(56,189,248,0.12)]' 
                  : f.status === 'false_positive' 
                    ? 'text-[#71717a] bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.04)]' 
                    : 'text-[#f59e0b] bg-[rgba(245,158,11,0.06)] border border-[rgba(245,158,11,0.12)]'
              }`}>
                {f.status.replace('_', ' ')}
              </span>
            )}
            {f.ai_triage_status && (
              <span className={`shrink-0 whitespace-nowrap text-[9px] px-1.5 py-0.5 rounded font-mono uppercase tracking-wider border transition-all ${
                f.ai_triage_status === 'true_positive' 
                  ? 'text-[#ef4444] bg-[rgba(239,68,68,0.08)] border-[rgba(239,68,68,0.2)] shadow-[0_0_8px_rgba(239,68,68,0.15)] font-bold' 
                  : f.ai_triage_status === 'false_positive' 
                    ? 'text-[#22c55e] bg-[rgba(34,197,94,0.08)] border-[rgba(34,197,94,0.2)] shadow-[0_0_8px_rgba(34,197,94,0.15)] font-bold' 
                    : 'text-[#eab308] bg-[rgba(234,179,8,0.08)] border-[rgba(234,179,8,0.2)] shadow-[0_0_8px_rgba(234,179,8,0.15)] font-bold'
              }`}>
                AI: {f.ai_triage_status.replace('_', ' ')}
              </span>
            )}
          </div>
        </div>
        <span className={`material-symbols-outlined text-[14px] text-[#3f3f46] group-hover:text-[#71717a] transition-transform duration-300 mt-0.5 ${isExpanded ? 'rotate-180 text-[var(--accent-color)]' : ''}`}>
          expand_more
        </span>
        </button>
      </div>
      
      <AnimatePresence initial={false}>
        {isExpanded && (
          <motion.div
            key="details"
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: "auto", opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={{ type: "spring", stiffness: 350, damping: 32 }}
            className="overflow-hidden border-l border-[rgba(255,255,255,0.04)] ml-5"
          >
            <div className="px-4 pb-4 pt-1.5 space-y-3 bg-[rgba(0,0,0,0.15)] border-y border-[rgba(255,255,255,0.02)]">
              <div className="flex items-center gap-2 flex-wrap">
                <span className="text-[9px] px-2 py-0.5 rounded font-mono font-bold uppercase border" 
                  style={{ 
                    color: sevDot(f.severity), 
                    backgroundColor: `${sevDot(f.severity)}08`,
                    borderColor: `${sevDot(f.severity)}1a`
                  }}
                >
                  {f.severity}
                </span>
                {f.file_path && <span className="text-[10px] text-[#52525b] font-mono truncate">{f.file_path}{f.line_number ? `:${f.line_number}` : ''}</span>}
              </div>
              
              {f.description && (
                <p className="text-[12px] text-[#a1a1aa] leading-relaxed select-text">
                  {f.description}
                </p>
              )}
              
              {(f.fix_suggestion || f.suggestion) && (
                <div className="text-[12px] text-[#a1a1aa] leading-relaxed border-l border-[rgba(255,255,255,0.08)] pl-3 my-2 select-text font-mono bg-[rgba(255,255,255,0.005)] p-2 rounded-r-md">
                  <span className="text-[9px] text-[#52525b] uppercase tracking-wider block mb-1 font-bold">Remediation Guidance</span>
                  {f.fix_suggestion || f.suggestion}
                </div>
              )}

              {f.ai_triage_summary && (
                <div className="text-[12px] text-[#a1a1aa] leading-relaxed border-l-2 pl-3 my-2.5 select-text p-2.5 rounded-r-md"
                  style={{
                    borderColor: f.ai_triage_status === 'true_positive' ? '#ef4444' : f.ai_triage_status === 'false_positive' ? '#22c55e' : '#eab308',
                    backgroundColor: f.ai_triage_status === 'true_positive' ? 'rgba(239,68,68,0.02)' : f.ai_triage_status === 'false_positive' ? 'rgba(34,197,94,0.02)' : 'rgba(234,179,8,0.02)'
                  }}
                >
                  <div className="flex items-center gap-1.5 mb-1.5">
                    <span className="material-symbols-outlined text-[14px]" style={{ color: f.ai_triage_status === 'true_positive' ? '#ef4444' : f.ai_triage_status === 'false_positive' ? '#22c55e' : '#eab308' }}>psychology</span>
                    <span className="text-[9px] uppercase tracking-wider font-bold" style={{ color: f.ai_triage_status === 'true_positive' ? '#ef4444' : f.ai_triage_status === 'false_positive' ? '#22c55e' : '#eab308' }}>
                      AI Triage Conclusion: {f.ai_triage_status?.replace('_', ' ').toUpperCase()}
                    </span>
                  </div>
                  <p className="text-[#e4e4e7] text-[14px] font-sans leading-relaxed">
                    {f.ai_triage_summary}
                  </p>
                </div>
              )}
              
              <div className="flex items-center gap-1.5 pt-1.5 flex-wrap">
                <button 
                  onClick={e => { 
                    e.stopPropagation(); 
                    if (isTriaging) return;
                    handleTriage(f, 'triage'); 
                  }} 
                  disabled={isTriaging || f.status === 'false_positive'}
                  className="text-[10px] text-[#38bdf8] border border-[rgba(56,189,248,0.15)] hover:bg-[rgba(56,189,248,0.06)] px-2.5 py-1 rounded-md flex items-center gap-1 transition-colors font-medium disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {isTriaging ? (
                    <>
                      <div className="w-3 h-3 border-2 border-t-[#38bdf8] border-[rgba(255,255,255,0.1)] rounded-full animate-spin shrink-0" />
                      <span>Analyzing...</span>
                    </>
                  ) : (
                    <>
                      <span className="material-symbols-outlined text-[12px]">psychology</span>Triage
                    </>
                  )}
                </button>
                <button 
                  onClick={e => { e.stopPropagation(); handleTriage(f, 'false_positive'); }} 
                  className="text-[10px] text-[#71717a] border border-[rgba(255,255,255,0.06)] hover:bg-[rgba(255,255,255,0.03)] px-2.5 py-1 rounded-md flex items-center gap-1 transition-colors font-medium"
                >
                  <span className="material-symbols-outlined text-[12px]">block</span>False Positive
                </button>
                <button 
                  onClick={e => { e.stopPropagation(); handleTriage(f, 'risk_accepted'); }} 
                  className="text-[10px] text-[#f59e0b] border border-[rgba(245,158,11,0.15)] hover:bg-[rgba(245,158,11,0.06)] px-2.5 py-1 rounded-md flex items-center gap-1 transition-colors font-medium"
                >
                  <span className="material-symbols-outlined text-[12px]">verified_user</span>Accept Risk
                </button>
                
                <div className="w-px h-3.5 bg-[rgba(255,255,255,0.06)] mx-1" />
                
                {onNavigateToChat && (
                  <button 
                    onClick={e => { e.stopPropagation(); onNavigateToChat(f); }} 
                    className="text-[10px] text-[#f4f4f5] bg-[#27272a] hover:bg-[#3f3f46] border border-[rgba(255,255,255,0.04)] px-2.5 py-1 rounded-md flex items-center gap-1 transition-all"
                  >
                    <span className="material-symbols-outlined text-[12px]">smart_toy</span>Ask AI
                  </button>
                )}
                
                <button 
                  onClick={e => { 
                    e.stopPropagation(); 
                    const prompt = `Fix this security vulnerability in my code:\n\n**Issue:** ${f.title}\n**Severity:** ${f.severity?.toUpperCase()}\n**File:** ${f.file_path || 'unknown'}${f.line_number ? `:${f.line_number}` : ''}\n**Scanner:** ${f.stack || 'core'}\n**Project:** ${productMap.get(f.product_id!)?.name || 'unknown'}\n\n**Description:**\n${f.description || 'No description'}\n\n**Recommendation:**\n${f.fix_suggestion || f.suggestion || 'No recommendation available'}\n\nPlease provide:\n1. Root cause analysis\n2. Fixed code with explanation\n3. How to verify the fix\n4. Prevention best practices`; 
                    navigator.clipboard.writeText(prompt); 
                    const btn = e.currentTarget;
                    btn.textContent = '✓ Copied!'; 
                    setTimeout(() => { 
                      try { 
                        btn.innerHTML = '<span class="material-symbols-outlined text-[12px]">content_paste</span>Fix Prompt'; 
                      } catch {} 
                    }, 1500); 
                  }} 
                  className="text-[10px] text-[var(--accent-color)] border border-[rgba(139,92,246,0.15)] hover:bg-[rgba(139,92,246,0.06)] px-2.5 py-1 rounded-md flex items-center gap-1 transition-colors font-medium"
                >
                  <span className="material-symbols-outlined text-[12px]">content_paste</span>Fix Prompt
                </button>
                
                <button 
                  onClick={e => { 
                    e.stopPropagation(); 
                    navigator.clipboard.writeText(`${f.title}\n${f.severity}\n${f.description || ''}\n${f.fix_suggestion || ''}`); 
                    const btn = e.currentTarget;
                    const oldHtml = btn.innerHTML;
                    btn.innerHTML = '<span class="material-symbols-outlined text-[12px] text-[#22c55e]">check</span>';
                    setTimeout(() => { try { btn.innerHTML = oldHtml; } catch {} }, 1500);
                  }} 
                  className="text-[10px] text-[#52525b] hover:text-[#a1a1aa] hover:bg-[rgba(255,255,255,0.02)] p-1 rounded-md transition-all shrink-0"
                >
                  <span className="material-symbols-outlined text-[12px]">content_copy</span>
                </button>
              </div>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
};

export const SimpleDashboardPage: React.FC<SimpleDashboardPageProps> = ({ onNavigateToChat }) => {
  const { t, i18n } = useTranslation('pages');
  const { findings, loading: findingsLoading, refresh: refreshFindings } = useFindings() as any;
  const { metrics, loading: metricsLoading, refresh: refreshMetrics } = useMetrics();
  const { products } = useProducts();

  const [globalScanning, setGlobalScanning] = useState(false);
  const [globalScanPath, setGlobalScanPath] = useState('/host');
  const [globalScanLogs, setGlobalScanLogs] = useState<string[]>([]);
  const [globalScanPhase, setGlobalScanPhase] = useState(0);
  const [globalScanElapsed, setGlobalScanElapsed] = useState(0);
  
  // AI query state
  const [aiQuery, setAiQuery] = useState('');

  const scanPhases = [
    { name: 'Core', desc: 'AST parsing & pattern matching', icon: 'memory' },
    { name: 'Semgrep', desc: 'SAST rules & taint analysis', icon: 'shield' },
    { name: 'Gitleaks', desc: 'Secrets & credential detection', icon: 'key' },
    { name: 'Trivy', desc: 'CVE & dependency vulnerabilities', icon: 'inventory_2' },
    { name: 'Bandit', desc: 'Python-specific security checks', icon: 'bug_report' },
  ];
  
  const scanLogMessages = [
    'Indexing source files...', 'Building AST...', 'Running pattern rules...',
    'Checking injection patterns...', 'Scanning for SQL injection...', 'Analyzing auth flows...',
    'Detecting hardcoded secrets...', 'Checking API keys...', 'Scanning .env files...',
    'Resolving dependencies...', 'Checking CVE database...', 'Analyzing lock files...',
    'Scanning Python imports...', 'Checking subprocess calls...', 'Detecting unsafe deserialization...',
    'Analyzing template injection...', 'Checking XSS vectors...', 'Scanning CSRF protections...',
  ];

  const handleGlobalScan = async (path: string) => {
    setGlobalScanning(true);
    setGlobalScanLogs(['Initializing in-place scan...']);
    setGlobalScanPhase(0);
    setGlobalScanElapsed(0);

    const tTimer = setInterval(() => setGlobalScanElapsed(e => e + 1), 1000);
    const pTimer = setInterval(() => setGlobalScanPhase(ph => (ph + 1) % scanPhases.length), 3000);
    const lTimer = setInterval(() => {
      const msg = scanLogMessages[Math.floor(Math.random() * scanLogMessages.length)];
      setGlobalScanLogs(prev => [...prev.slice(-10), msg]);
    }, 700);

    try {
      const res = await fetch('/api/scan', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ path, external: true }),
      });
      const data = await res.json();
      if (data.ok) {
        refreshFindings?.();
        refreshMetrics?.();
      } else {
        alert(data.error || 'Scan failed');
      }
    } catch (err: any) {
      alert(err.message || 'Connection error during scan');
    } finally {
      clearInterval(tTimer);
      clearInterval(pTimer);
      clearInterval(lTimer);
      setGlobalScanning(false);
    }
  };

  const [expandedIds, setExpandedIds] = useState<Set<number>>(new Set());
  const [activeFilter, setActiveFilter] = useState<string | null>(null);
  const [productFilter, setProductFilter] = useState<number | null>(null);
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [searchQuery, setSearchQuery] = useState('');
  const [groupBy, setGroupBy] = useState<GroupBy>('none');
  const [sortBy, setSortBy] = useState<SortBy>('severity');
  const [page, setPage] = useState(0);
  const [collapsedGroups, setCollapsedGroups] = useState<Set<string>>(new Set());
  const [aiSummary, setAiSummary] = useState<string>('');
  const [aiSummaryLoading, setAiSummaryLoading] = useState(false);
  const [aiSummaryProjectId, setAiSummaryProjectId] = useState<number | null>(null);
  const [aiSummaryLang, setAiSummaryLang] = useState<'en' | 'ru'>('ru');
  const [toolStatus, setToolStatus] = useState<Record<string, boolean>>({});

  useEffect(() => {
    fetch('/api/health').then(r => r.json()).then(d => { if (d.ok && d.tools) setToolStatus(d.tools); }).catch(() => {});
  }, []);

  // SecureCoder Bulk Selection & Active File Scope State
  const [selectedFindings, setSelectedFindings] = useState<Set<number>>(new Set());
  const [configAutostartFixes, setConfigAutostartFixes] = useState(true);
  const [bulkIgnoreModalOpen, setBulkIgnoreModalOpen] = useState(false);
  const [bulkIgnoreReason, setBulkIgnoreReason] = useState('False Positive');
  const [bulkIgnoring, setBulkIgnoring] = useState(false);
  const [bulkFixCopied, setBulkFixCopied] = useState(false);
  const [scopeType, setScopeType] = useState<'all' | 'activeFile'>('all');
  const [activeFilePath, setActiveFilePath] = useState<string>('');

  useEffect(() => {
    fetch('/api/securecoder/config')
      .then(r => r.json())
      .then(data => {
        if (data.autostartFixes !== undefined) setConfigAutostartFixes(data.autostartFixes);
      })
      .catch(() => {});
  }, []);

  const handleSaveConfig = async (overrideSettings?: any) => {
    try {
      await fetch('/api/securecoder/config', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          autostartFixes: overrideSettings?.autostartFixes ?? configAutostartFixes
        })
      });
    } catch (e) {
      console.error(e);
    }
  };

  const handleBulkIgnore = async () => {
    setBulkIgnoring(true);
    try {
      const selectedObjects = findings.filter((f: any) => selectedFindings.has(f.id));
      for (const f of selectedObjects) {
        await fetch(`/api/findings/${f.id}`, {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ action: 'status', status: bulkIgnoreReason === 'False Positive' ? 'false_positive' : 'risk_accepted' })
        });
        
        await fetch('/api/securecoder/ignore', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            filePath: f.file_path || '',
            ruleId: f.rule_id || '',
            codeSnippet: f.code_snippet || '',
            lineNumber: f.line_number || 0,
            vulnerabilityClass: f.title || '',
            reason: bulkIgnoreReason
          })
        });
      }
      setSelectedFindings(new Set());
      setBulkIgnoreModalOpen(false);
      window.location.reload();
    } catch (e) {
      console.error(e);
    } finally {
      setBulkIgnoring(false);
    }
  };

  const handleBulkFix = () => {
    const selectedObjects = findings.filter((f: any) => selectedFindings.has(f.id));
    if (selectedObjects.length === 0) return;

    let prompt = `Fix these security vulnerabilities in my code:\n\n`;
    selectedObjects.forEach((f: any, idx: number) => {
      prompt += `### Finding #${idx + 1}: ${f.title}\n`;
      prompt += `- **Severity:** ${f.severity?.toUpperCase()}\n`;
      prompt += `- **File:** ${f.file_path || 'unknown'}${f.line_number ? `:${f.line_number}` : ''}\n`;
      prompt += `- **Scanner:** ${f.stack || 'core'}\n`;
      prompt += `- **Description:** ${f.description || 'No description'}\n`;
      if (f.code_snippet) {
        prompt += `- **Code Snippet:**\n\`\`\`\n${f.code_snippet}\n\`\`\`\n`;
      }
      if (f.fix_suggestion || f.suggestion) {
        prompt += `- **Recommendation:** ${f.fix_suggestion || f.suggestion}\n`;
      }
      prompt += `\n`;
    });

    prompt += `Please perform a root-cause analysis for each finding and generate targeted before/after code patches and PoC verification guides according to the SecureCoder guidelines.`;

    navigator.clipboard.writeText(prompt);
    setBulkFixCopied(true);
    setTimeout(() => setBulkFixCopied(false), 2000);
  };

  const [isProjectsPanelOpen, setIsProjectsPanelOpen] = useState(() => {
    try {
      return localStorage.getItem('projects_panel_open') !== 'false';
    } catch {
      return true;
    }
  });

  const handleToggleProjectsPanel = useCallback(() => {
    setIsProjectsPanelOpen((prev) => {
      const next = !prev;
      try {
        localStorage.setItem('projects_panel_open', String(next));
      } catch {}
      return next;
    });
  }, []);

  // Build product map for quick lookup
  const productMap = useMemo(() => {
    const m = new Map<number, Product>();
    products?.forEach(p => m.set(p.id, p));
    return m;
  }, [products]);

  // Get products that have findings
  const activeProducts = useMemo(() => {
    if (!findings || !products) return [];
    const ids = new Set<number>();
    findings.forEach((f: Finding) => { if (f.product_id) ids.add(f.product_id); });
    return products.filter(p => ids.has(p.id));
  }, [findings, products]);

  const loading = findingsLoading || metricsLoading;

  const closedStatuses = ['resolved', 'closed', 'false_positive', 'risk_accepted'];

  const sevCounts = useMemo(() => {
    const c = { critical: 0, high: 0, medium: 0, low: 0 };
    findings?.forEach((f: Finding) => {
      if (productFilter !== null && f.product_id !== productFilter) return;
      // Exclude resolved/closed findings — match backend metrics query
      const st = (f.status || 'open').toLowerCase();
      if (closedStatuses.includes(st)) return;
      const s = f.severity?.toLowerCase();
      if (s === 'critical') c.critical++;
      else if (s === 'high') c.high++;
      else if (s === 'medium') c.medium++;
      else c.low++;
    });
    return c;
  }, [findings, productFilter]);

  const score = useMemo(() => {
    if (productFilter === null && metrics?.security_score !== undefined) {
      // Global score: use backend value for consistency with ADVANCED mode
      return metrics.security_score;
    }
    // Per-product score: calculate client-side (no backend per-product score)
    const penalty = sevCounts.critical * 10 + sevCounts.high * 4 + sevCounts.medium * 1;
    const s = 100 - penalty;
    return s < 0 ? 0 : s;
  }, [sevCounts, productFilter, metrics]);


  const projectStats = useMemo(() => {
    let total = 0;
    let resolved = 0;
    findings?.forEach((f: Finding) => {
      if (productFilter !== null && f.product_id !== productFilter) return;
      total++;
      // Count findings that are in terminal/resolved states
      const st = (f.status || 'open').toLowerCase();
      if (closedStatuses.includes(st)) {
        resolved++;
      }
    });
    return { total, resolved };
  }, [findings, productFilter]);

  const uniqueFilePaths = useMemo(() => {
    if (!findings) return [];
    const paths = new Set<string>();
    findings.forEach((f: Finding) => {
      if (productFilter !== null && f.product_id !== productFilter) return;
      if (f.file_path) paths.add(f.file_path);
    });
    return Array.from(paths).sort();
  }, [findings, productFilter]);

  const sevOrder: Record<string, number> = { critical: 0, high: 1, medium: 2, low: 3 };

  const filteredFindings = useMemo(() => {
    if (!findings) return [];
    let filtered = [...findings];
    if (productFilter !== null) {
      filtered = filtered.filter((f: Finding) => f.product_id === productFilter);
    }
    if (activeFilter) {
      filtered = filtered.filter((f: Finding) => f.severity?.toLowerCase() === activeFilter);
    }
    if (statusFilter !== 'all') {
      filtered = filtered.filter((f: Finding) => (f.status || 'open') === statusFilter);
    }
    if (scopeType === 'activeFile' && activeFilePath) {
      filtered = filtered.filter((f: Finding) => f.file_path === activeFilePath);
    }
    if (searchQuery.trim()) {
      const q = searchQuery.toLowerCase();
      filtered = filtered.filter((f: Finding) =>
        f.title?.toLowerCase().includes(q) ||
        f.description?.toLowerCase().includes(q) ||
        f.file_path?.toLowerCase().includes(q)
      );
    }
    filtered.sort((a: Finding, b: Finding) => {
      if (sortBy === 'severity') return (sevOrder[a.severity?.toLowerCase()] ?? 4) - (sevOrder[b.severity?.toLowerCase()] ?? 4);
      if (sortBy === 'title') return (a.title || '').localeCompare(b.title || '');
      if (sortBy === 'file') return (a.file_path || '').localeCompare(b.file_path || '');
      return 0;
    });
    return filtered;
  }, [findings, productFilter, activeFilter, statusFilter, scopeType, activeFilePath, searchQuery, sortBy]);

  const groups = useMemo(() => {
    if (groupBy === 'none') return null;
    const map = new Map<string, Finding[]>();
    filteredFindings.forEach((f: Finding) => {
      let key: string;
      switch (groupBy) {
        case 'severity': key = (f.severity || 'unknown').toUpperCase(); break;
        case 'title': key = f.title || 'Untitled'; break;
        case 'file': key = f.file_path || 'No file'; break;
        case 'scanner': key = f.stack || 'core'; break;
        case 'product': key = f.product_id ? (productMap.get(f.product_id)?.name || `Project #${f.product_id}`) : 'Unassigned'; break;
        default: key = 'Other';
      }
      if (!map.has(key)) map.set(key, []);
      map.get(key)!.push(f);
    });
    const entries = Array.from(map.entries());
    if (groupBy === 'severity') entries.sort((a, b) => (sevOrder[a[0].toLowerCase()] ?? 4) - (sevOrder[b[0].toLowerCase()] ?? 4));
    else entries.sort((a, b) => b[1].length - a[1].length);
    return entries;
  }, [filteredFindings, groupBy]);

  const totalPages = Math.ceil(filteredFindings.length / PAGE_SIZE);
  const pagedFindings = groupBy === 'none' ? filteredFindings.slice(page * PAGE_SIZE, (page + 1) * PAGE_SIZE) : [];

  const toggleGroup = (key: string) => {
    setCollapsedGroups(prev => { const n = new Set(prev); n.has(key) ? n.delete(key) : n.add(key); return n; });
  };

  const [triagingIds, setTriagingIds] = useState<Set<number>>(new Set());

  const handleTriage = async (f: Finding, action: string) => {
    if (action === 'triage') {
      setTriagingIds(prev => {
        const next = new Set(prev);
        next.add(f.id);
        return next;
      });
      try {
        const res = await fetch(`/api/findings/${f.id}/ai-triage`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' }
        });
        const data = await res.json();
        if (!data.ok) {
          alert(data.error || 'AI Triage failed');
        }
      } catch (e) {
        console.error(e);
        alert('Failed to connect to AI Triage service');
      } finally {
        setTriagingIds(prev => {
          const next = new Set(prev);
          next.delete(f.id);
          return next;
        });
        window.location.reload();
      }
      return;
    }

    try {
      await fetch(`/api/findings/${f.id}`, { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ action: 'status', status: action }) });
      
      if (action === 'false_positive' || action === 'risk_accepted') {
        const reason = action === 'false_positive' ? 'False Positive' : 'Accepted Risk';
        await fetch('/api/securecoder/ignore', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            filePath: f.file_path || '',
            ruleId: f.rule_id || '',
            codeSnippet: f.code_snippet || '',
            lineNumber: f.line_number || 0,
            vulnerabilityClass: f.title || '',
            reason: reason
          })
        });
      }
      window.location.reload();
    } catch (e) {
      console.error(e);
    }
  };

  const sevDot = (sev: string) => {
    switch (sev?.toLowerCase()) {
      case 'critical': return '#ef4444'; case 'high': return '#f97316'; case 'medium': return '#eab308'; case 'low': return '#3f3f46'; default: return '#3f3f46';
    }
  };

  if (loading) {
    return (<div className="flex h-full items-center justify-center bg-v2-bg"><div className="flex items-center gap-3"><div className="w-4 h-4 border-2 border-[#27272a] border-t-[var(--accent-color)] rounded-full animate-spin" /><span className="text-[13px] text-[#71717a] font-mono tracking-wider uppercase">Loading...</span></div></div>);
  }

  return (
    <div className="flex h-full overflow-hidden">
      <div className="flex-1 overflow-y-auto" style={{ scrollbarWidth: 'thin', scrollbarColor: 'rgba(255,255,255,0.06) transparent' }}>
        <motion.div 
          variants={containerVariants}
          initial="hidden"
          animate="visible"
          className="px-8 py-6"
        >
          <div className="flex flex-col gap-6">
            {/* ── AI COMMAND & SCAN BAR ── */}
            <motion.div 
              variants={itemVariants} 
              className="border border-[rgba(255,255,255,0.06)] rounded-2xl p-6 bg-background/80 backdrop-blur-xl flex flex-col gap-5 shadow-2xl relative overflow-hidden group"
            >
              {/* Subtle glass glow background */}
              <div className="absolute inset-0 bg-gradient-to-br from-[rgba(255,255,255,0.02)] to-[rgba(255,255,255,0.002)] pointer-events-none" />
              
              {/* Row 1: AI Prompt Input */}
              <div className="flex flex-col gap-2 relative z-10">
                <label className="text-[10px] font-bold text-[#71717a] tracking-[0.2em] uppercase font-mono flex items-center gap-1.5">
                  <span className="material-symbols-outlined text-[13px] text-[var(--accent-color)]">smart_toy</span>
                  AI Security Copilot Command Bar
                </label>
                <div className="flex gap-2">
                  <div className="relative flex-1">
                    <span className="material-symbols-outlined text-[16px] text-[#52525b] absolute left-3.5 top-1/2 -translate-y-1/2">
                      search
                    </span>
                    <input
                      value={aiQuery}
                      onChange={(e) => setAiQuery(e.target.value)}
                      onKeyDown={(e) => {
                        if (e.key === 'Enter' && aiQuery.trim()) {
                          onNavigateToChat?.(aiQuery);
                        }
                      }}
                      placeholder="Ask the security agent to check CORS rules, analyze auth flows, or review code..."
                      className="w-full bg-surface-bright border border-[rgba(255,255,255,0.06)] rounded-xl pl-10 pr-4 py-3 text-[13px] text-[#f4f4f5] placeholder:text-[#52525b] outline-none focus:border-[rgba(255,255,255,0.15)] transition-all font-sans"
                    />
                  </div>
                  <button
                    onClick={() => {
                      if (aiQuery.trim() && onNavigateToChat) {
                        onNavigateToChat(aiQuery);
                      }
                    }}
                    disabled={!aiQuery.trim()}
                    className="v2-btn px-6 bg-white hover:bg-[#e4e4e7] text-black font-semibold text-[13px] rounded-xl flex items-center gap-2 h-[46px] transition-all cursor-pointer disabled:opacity-30 disabled:pointer-events-none"
                  >
                    <span className="material-symbols-outlined text-[16px]">bolt</span>
                    Ask AI
                  </button>
                </div>
                
                {/* Suggestions */}
                <div className="flex flex-wrap items-center gap-1.5 mt-1">
                  <span className="text-[10px] text-[#52525b] font-medium mr-1 uppercase tracking-wider font-mono">Suggestions:</span>
                  {[
                    "Find SQL Injection in /host",
                    "Check if CORS is secure",
                    "Generate STRIDE model for Next.js"
                  ].map((s) => (
                    <button
                      key={s}
                      onClick={() => onNavigateToChat?.(s)}
                      className="text-[11px] text-[#a1a1aa] bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.05)] hover:bg-[rgba(255,255,255,0.08)] hover:text-white px-3 py-1 rounded-full transition-all cursor-pointer font-sans"
                    >
                      {s}
                    </button>
                  ))}
                </div>
              </div>

              {/* Separator line */}
              <div className="h-[1px] bg-[rgba(255,255,255,0.05)] relative z-10" />

              {/* Row 2: In-place Directory Scanner */}
              <div className="flex flex-col gap-2 relative z-10">
                <label className="text-[10px] font-bold text-[#71717a] tracking-[0.2em] uppercase font-mono flex items-center gap-1.5">
                  <span className="material-symbols-outlined text-[13px] text-[var(--accent-color)]">folder_open</span>
                  Quick Workspace Scanner
                </label>
                <div className="flex gap-2">
                  <div className="relative flex-1">
                    <span className="material-symbols-outlined text-[15px] text-[#52525b] absolute left-3.5 top-1/2 -translate-y-1/2">
                      terminal
                    </span>
                    <input
                      value={globalScanPath}
                      onChange={(e) => setGlobalScanPath(e.target.value)}
                      placeholder="/host/my-project"
                      className="w-full bg-surface-bright border border-[rgba(255,255,255,0.06)] rounded-xl pl-10 pr-4 py-3 text-[13px] text-[#a1a1aa] placeholder:text-[#3f3f46] outline-none focus:border-[rgba(255,255,255,0.15)] transition-all font-mono"
                    />
                  </div>
                  <button
                    onClick={() => handleGlobalScan(globalScanPath)}
                    className="v2-btn px-6 border border-[#27272a] hover:bg-[rgba(255,255,255,0.03)] text-white font-semibold text-[13px] rounded-xl flex items-center gap-2 h-[46px] transition-all cursor-pointer"
                  >
                    <span className="material-symbols-outlined text-[16px]">play_arrow</span>
                    Scan Now
                  </button>
                </div>
              </div>
            </motion.div>

            {/* ── TOP BENTO ROW ── */}
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 items-stretch">
              
              {/* Score Card */}
              <motion.div 
                variants={itemVariants} 
                className="lg:col-span-1 border border-[rgba(255,255,255,0.06)] rounded-2xl p-6 bg-background/80 backdrop-blur-xl flex flex-col justify-between shadow-2xl relative overflow-hidden min-h-[260px] h-full group"
              >
                {/* Advanced Animated Background Glow */}
                <div className="absolute inset-0 overflow-hidden pointer-events-none rounded-2xl">
                  <div 
                    className="absolute -top-[100px] -right-[50px] w-[300px] h-[300px] rounded-full mix-blend-screen opacity-20 filter blur-[90px] group-hover:opacity-30 transition-all duration-700 ease-in-out group-hover:scale-110"
                    style={{
                      background: 'radial-gradient(circle, var(--accent-color) 0%, transparent 70%)'
                    }}
                  />
                  <div 
                    className="absolute -bottom-[100px] -left-[50px] w-[200px] h-[200px] rounded-full mix-blend-screen opacity-10 filter blur-[70px] group-hover:opacity-20 transition-all duration-1000 ease-in-out group-hover:scale-125 delay-150"
                    style={{
                      background: 'radial-gradient(circle, var(--accent-color-hover) 0%, transparent 70%)'
                    }}
                  />
                </div>

                {/* Optional glassmorphism overlay */}
                <div className="absolute inset-0 bg-gradient-to-br from-[rgba(255,255,255,0.03)] to-[rgba(255,255,255,0.005)] rounded-2xl pointer-events-none" />

                {/* Header Section */}
                <div className="flex items-start justify-between z-10">
                  <div>
                    <div className="flex items-center gap-2 mb-1">
                      <div className="w-6 h-6 rounded-md bg-[rgba(255,255,255,0.06)] border border-[rgba(255,255,255,0.05)] flex items-center justify-center shadow-inner">
                        <span className="material-symbols-outlined text-[14px] text-[#e4e4e7]">security</span>
                      </div>
                      <h3 className="text-sm font-semibold text-[#f4f4f5] tracking-wide font-sans uppercase">
                        {t('securityScore')}
                      </h3>
                    </div>
                    <div className="text-[11px] text-[#71717a] font-mono tracking-wider ml-8 uppercase">
                      {productFilter !== null && productMap.has(productFilter) 
                        ? `${productMap.get(productFilter)!.name}` 
                        : t('allProjects')}
                    </div>
                  </div>
                  
                  {/* Premium Status Badge */}
                  <div 
                    className="relative flex items-center gap-1.5 px-2.5 py-1 rounded-full border shadow-sm backdrop-blur-md"
                    style={{
                      color: 'var(--accent-color)',
                      backgroundColor: 'var(--accent-color-soft)',
                      borderColor: 'var(--accent-color-line)'
                    }}
                  >
                    <span className="relative flex h-2 w-2">
                      <span className="animate-ping absolute inline-flex h-full w-full rounded-full opacity-75" style={{ backgroundColor: 'var(--accent-color)' }}></span>
                      <span className="relative inline-flex rounded-full h-2 w-2" style={{ backgroundColor: 'var(--accent-color)' }}></span>
                    </span>
                    <span className="text-[10px] font-bold tracking-widest uppercase">
                      {score < 30 ? t('criticalRisk') : score < 60 ? t('highRisk') : score < 80 ? t('mediumRisk') : t('secureStatus')}
                    </span>
                  </div>
                </div>

                {/* Middle: Gauge (Left) and Severity bars (Right) side-by-side */}
                <div className="flex gap-6 items-center my-5 relative z-10 flex-1">
                  {/* Left Column: Radial Progress Gauge (Redesigned) */}
                  <div className="relative shrink-0 group-hover:scale-105 transition-transform duration-500 ease-out">
                    <svg className="w-[110px] h-[110px] transform -rotate-90 filter drop-shadow-xl overflow-visible" viewBox="0 0 100 100" style={{ overflow: 'visible' }}>
                      <circle
                        cx="50"
                        cy="50"
                        r="42"
                        className="stroke-[rgba(255,255,255,0.04)] fill-none"
                        strokeWidth="8"
                      />
                      <circle
                        cx="50"
                        cy="50"
                        r="42"
                        className="fill-none stroke-current transition-all duration-1500 ease-out"
                        strokeWidth="8"
                        strokeDasharray={2 * Math.PI * 42}
                        strokeDashoffset={2 * Math.PI * 42 * (1 - score / 100)}
                        strokeLinecap="round"
                        style={{
                          color: 'var(--accent-color)',
                          filter: 'drop-shadow(0 0 8px var(--accent-color-line))'
                        }}
                      />
                    </svg>
                    <div className="absolute inset-0 flex flex-col items-center justify-center">
                      <span 
                        className="text-[34px] font-black tracking-tighter leading-none"
                        style={{
                          background: 'linear-gradient(135deg, #ffffff 0%, var(--accent-color) 100%)',
                          WebkitBackgroundClip: 'text',
                          WebkitTextFillColor: 'transparent',
                          filter: 'drop-shadow(0px 2px 4px rgba(0,0,0,0.4))'
                        }}
                      >
                        {score}
                      </span>
                      <span className="text-[9px] text-[#71717a] font-bold tracking-[0.2em] mt-0.5 uppercase">
                        {i18n.language?.startsWith('ru') ? 'ИЗ 100' : 'SCORE'}
                      </span>
                    </div>
                  </div>

                  {/* Right Column: High-density Severity list */}
                  <div className="flex-1 flex flex-col justify-center space-y-2.5">
                    {(['critical', 'high', 'medium', 'low'] as const).map(s => {
                      const count = sevCounts[s];
                      const color = sevDot(s);
                      const totalCount = sevCounts.critical + sevCounts.high + sevCounts.medium + sevCounts.low;
                      const pct = totalCount > 0 ? (count / totalCount) * 100 : 0;
                      return (
                        <div key={s} className="flex flex-col gap-1.5 group/bar cursor-default">
                          <div className="flex items-center justify-between text-xs font-mono leading-none">
                            <span className="flex items-center gap-1.5 text-[#a1a1aa] uppercase text-[10px] tracking-wider font-semibold transition-colors group-hover/bar:text-[#e4e4e7]">
                              <span className="w-1.5 h-1.5 rounded-full shrink-0 shadow-[0_0_4px_currentColor]" style={{ backgroundColor: color, color: color }} />
                              {s === 'critical' ? (i18n.language?.startsWith('ru') ? 'Крит' : 'Crit') : s === 'high' ? (i18n.language?.startsWith('ru') ? 'Высок' : 'High') : s === 'medium' ? (i18n.language?.startsWith('ru') ? 'Сред' : 'Med') : (i18n.language?.startsWith('ru') ? 'Низк' : 'Low')}
                            </span>
                            <span className="font-bold text-[11px] transition-colors" style={{ color: count > 0 ? color : '#52525b' }}>{count}</span>
                          </div>
                          <div className="w-full h-1.5 bg-[rgba(255,255,255,0.03)] rounded-full overflow-hidden shadow-inner border border-[rgba(255,255,255,0.02)]">
                            <div 
                              className="h-full rounded-full transition-all duration-1000 ease-out relative" 
                              style={{ 
                                width: `${pct}%`, 
                                backgroundColor: color,
                                opacity: count > 0 ? 1 : 0.1,
                                boxShadow: count > 0 ? `0 0 8px ${color}80` : 'none'
                              }} 
                            />
                          </div>
                        </div>
                      );
                    })}
                  </div>
                </div>

                {/* Remediation/Resolution progress bar */}
                {metrics && (
                  <div className="border-t border-[rgba(255,255,255,0.04)] pt-4 mt-1 z-10 font-mono">
                    <div className="flex justify-between items-center text-[10px] text-[#a1a1aa] font-bold uppercase tracking-widest mb-2">
                      <span className="flex items-center gap-1.5">
                        <span className="material-symbols-outlined text-[13px]" style={{ color: 'var(--accent-color)' }}>fact_check</span>
                        {t('remediationProgress')}
                      </span>
                      <span className="text-[#f4f4f5] tabular-nums">
                        {projectStats.total > 0 
                          ? `${Math.round((projectStats.resolved / projectStats.total) * 100)}%` 
                          : '0%'}
                      </span>
                    </div>
                    <div className="w-full h-2 bg-[rgba(255,255,255,0.03)] rounded-full overflow-hidden shadow-inner border border-[rgba(255,255,255,0.02)] relative p-[1px]">
                      <div 
                        className="h-full rounded-full transition-all duration-1000 ease-out relative overflow-hidden"
                        style={{
                          width: `${projectStats.total > 0 ? (projectStats.resolved / projectStats.total) * 100 : 0}%`,
                          background: 'linear-gradient(90deg, var(--accent-color-hover) 0%, var(--accent-color) 100%)',
                          boxShadow: '0 0 10px var(--accent-color-line)'
                        }}
                      />
                    </div>
                    <div className="flex justify-between text-[10px] text-[#52525b] mt-2 font-medium tracking-wide">
                      <span>{t('resolvedCount', { count: projectStats.resolved })}</span>
                      <span>{t('totalFindingsCount', { count: projectStats.total })}</span>
                    </div>
                  </div>
                )}
              </motion.div>

              {/* ── AI Security Summary Section ── */}
              <div className="lg:col-span-2">
                {(() => {
                  const summaryProjectId = productFilter ?? activeProducts.sort((a, b) => {
                    const aCount = findings?.filter((f: Finding) => f.product_id === a.id).length || 0;
                    const bCount = findings?.filter((f: Finding) => f.product_id === b.id).length || 0;
                    return bCount - aCount;
                  })[0]?.id ?? null;

                  if (!summaryProjectId || !productMap.has(summaryProjectId)) {
                    return (
                      <motion.div variants={itemVariants} className="h-full border border-dashed border-[rgba(139,92,246,0.2)] rounded-xl p-6 bg-[rgba(139,92,246,0.02)] flex flex-col justify-center items-center gap-3 text-center min-h-[250px]">
                        <span className="material-symbols-outlined text-[32px] text-[#3f3f46]">analytics</span>
                        <div>
                          <div className="text-[14px] font-medium text-[#71717a] mb-1">No projects scanned yet</div>
                          <div className="text-[12px] text-[#52525b]">Scan a repository from the sidebar to generate an AI security summary</div>
                        </div>
                      </motion.div>
                    );
                  }

                  const proj = productMap.get(summaryProjectId)!;
                  const pf = findings?.filter((f: Finding) => f.product_id === summaryProjectId) || [];

                  const pCrit = pf.filter((f: Finding) => f.severity?.toLowerCase() === 'critical').length;
                  const pHigh = pf.filter((f: Finding) => f.severity?.toLowerCase() === 'high').length;
                  const pMed = pf.filter((f: Finding) => f.severity?.toLowerCase() === 'medium').length;
                  const pLow = pf.filter((f: Finding) => f.severity?.toLowerCase() === 'low').length;
                  
                  return (
                    <motion.div variants={itemVariants} className="h-full rounded-xl overflow-hidden border border-[rgba(255,255,255,0.08)] bg-gradient-to-br from-[rgba(255,255,255,0.02)] to-[rgba(255,255,255,0.005)] shadow-lg p-5 relative flex flex-col min-h-[250px]">
                      {/* Header */}
                      <div className="flex items-center justify-between mb-4 border-b border-[rgba(255,255,255,0.06)] pb-3 shrink-0">
                        <div className="flex items-center gap-2">
                          <span className="material-symbols-outlined text-lg text-[#71717a]">psychology</span>
                          <span className="text-xs text-[#e4e4e7] uppercase tracking-[0.15em] font-bold">{t('aiSecuritySummary')}</span>
                        </div>
                        <div className="flex items-center gap-2.5">
                          <select value={aiSummaryLang} onChange={e => setAiSummaryLang(e.target.value as 'en' | 'ru')}
                            className="text-xs text-[#a1a1aa] bg-[rgba(255,255,255,0.04)] border border-[rgba(255,255,255,0.08)] rounded px-2.5 py-1 outline-none hover:border-[rgba(255,255,255,0.12)] transition-colors font-bold tracking-wider uppercase font-sans cursor-pointer">
                            <option value="ru">RU</option>
                            <option value="en">EN</option>
                          </select>
                          {activeProducts.length > 1 && (
                            <select value={summaryProjectId} onChange={e => setProductFilter(Number(e.target.value))}
                              className="text-xs text-[#a1a1aa] bg-[rgba(255,255,255,0.04)] border border-[rgba(255,255,255,0.08)] rounded px-2.5 py-1 outline-none hover:border-[rgba(255,255,255,0.12)] transition-colors font-bold tracking-wider uppercase font-sans max-w-[150px] cursor-pointer">
                              {activeProducts.map(p => <option key={p.id} value={p.id}>{p.name}</option>)}
                            </select>
                          )}
                          {aiSummaryProjectId === summaryProjectId && aiSummary && !aiSummaryLoading && (
                            <button onClick={() => { setAiSummary(''); setAiSummaryProjectId(null); }}
                              className="text-xs text-[#71717a] hover:text-[#e4e4e7] transition-colors flex items-center gap-1.5 font-bold uppercase tracking-wider font-mono">
                              <span className="material-symbols-outlined text-[13px]">refresh</span>Regenerate
                            </button>
                          )}
                        </div>
                      </div>

                      {/* AI Summary Content */}
                      <div className="flex-1 overflow-y-auto pr-2" style={{ scrollbarWidth: 'thin', scrollbarColor: 'rgba(255,255,255,0.06) transparent' }}>
                        {aiSummaryLoading && aiSummaryProjectId === summaryProjectId ? (
                          <div className="flex flex-col items-center justify-center h-full gap-3 text-[#71717a] min-h-[140px]">
                            <div className="flex gap-1.5">
                              {[0, 1, 2].map(i => (
                                <div key={i} className="w-2 h-2 rounded-full bg-[#52525b] animate-bounce" style={{ animationDelay: `${i * 150}ms` }} />
                              ))}
                            </div>
                            <span className="text-xs font-mono tracking-widest uppercase">Analyzing repository security...</span>
                          </div>
                        ) : aiSummaryProjectId === summaryProjectId && aiSummary ? (
                          <div className="text-[13px] text-[#a1a1aa] leading-relaxed prose prose-invert max-w-none [&_strong]:text-[#e4e4e7] [&_strong]:font-semibold [&_code]:text-[#e4e4e7] [&_code]:bg-[#27272a] [&_code]:border [&_code]:border-[#3f3f46] [&_code]:px-1.5 [&_code]:py-0.5 [&_code]:rounded-md [&_code]:text-[12px] [&_p]:mb-2.5 last:[&_p]:mb-0 select-text">
                            <Markdown>{aiSummary}</Markdown>
                          </div>
                        ) : (
                          <div className="grid grid-cols-1 md:grid-cols-12 gap-6 h-full items-center min-h-[140px]">
                            {/* Left: System Specifications */}
                            <div className="md:col-span-7 md:border-r border-[rgba(255,255,255,0.06)] pr-5 space-y-3 flex flex-col justify-center h-full">
                              <div className="flex items-center gap-1.5">
                                <span className="material-symbols-outlined text-sm text-[var(--accent-color)]">folder_open</span>
                                <span className="text-xs text-[#71717a] font-mono tracking-widest uppercase font-bold">ACTIVE REPOSITORY</span>
                              </div>
                              <div className="text-base font-extrabold text-white tracking-tight">{proj.name}</div>
                              <div className="grid grid-cols-2 gap-2.5 pt-1">
                                {[
                                  { name: 'Trivy', key: 'trivy', label: 'Deps & Vulns' },
                                  { name: 'Semgrep', key: 'semgrep', label: 'SAST Engine' },
                                  { name: 'Gitleaks', key: 'gitleaks', label: 'Secrets Scan' },
                                  { name: 'Bandit', key: 'bandit', label: 'Python SAST' }
                                ].map(tool => {
                                  const installed = toolStatus[tool.key];
                                  return (
                                    <div key={tool.key} className="flex items-center gap-2.5 p-2 rounded-lg bg-[rgba(255,255,255,0.015)] border border-[rgba(255,255,255,0.04)]">
                                      <div className={`w-2 h-2 rounded-full shrink-0 ${installed ? 'bg-[#22c55e]' : 'bg-[#ef4444]'}`} />
                                      <div className="flex-1 min-w-0 font-mono">
                                        <div className="text-xs text-[#e4e4e7] font-bold leading-none">{tool.name}</div>
                                        <div className="text-[10px] text-[#71717a] mt-1 leading-none">{tool.label}</div>
                                      </div>
                                    </div>
                                  );
                                })}
                              </div>
                            </div>

                            {/* Right: Neural Prompt activation */}
                            <div className="md:col-span-5 flex flex-col items-center justify-center text-center p-4 rounded-xl bg-[rgba(255,255,255,0.015)] border border-[rgba(255,255,255,0.05)] relative overflow-hidden group h-full">
                              <div className="absolute inset-0 bg-radial-gradient from-[rgba(139,92,246,0.02)] to-transparent pointer-events-none group-hover:opacity-100 transition-opacity" />
                              <span className="material-symbols-outlined text-[30px] text-[var(--accent-color)] animate-pulse mb-2">psychology</span>
                              <p className="text-xs text-[#a1a1aa] leading-normal max-w-[220px] mb-4">
                                Generate summary utilizing SecureCoder LLM agent pipeline
                              </p>
                              <button
                                onClick={async () => {
                                  setAiSummary('');
                                  setAiSummaryLoading(true);
                                  setAiSummaryProjectId(summaryProjectId);
                                  try {
                                    const projFindings = pf.map((f: Finding) =>
                                      `- [${f.severity?.toUpperCase()}] ${f.title} | File: ${f.file_path || 'N/A'}${f.line_number ? ':' + f.line_number : ''}`
                                    ).join('\n');
                                    const summaryPrompt = `You are AITriage, an expert security engineer powered by SecureCoder. Analyze this project and provide a comprehensive security summary.
        
        Project: ${proj.name}
        Total findings: ${pf.length} (${pCrit} critical, ${pHigh} high, ${pMed} medium, ${pLow} low)
        
        Findings:
        ${projFindings}
        
        Provide exactly 4 sections in your response (do not use Markdown headers like # or ##, just bold text for labels):
        1. A brief 1-sentence overview of what this project likely is based on its name, files, and vulnerabilities.
        2. **Security Status:** [Emoji 🔴/🟡/🟢] [Brief status explanation].
        3. **Main Priority:** [Top issue to fix immediately and why].
        4. **Quick Win:** [Easiest thing to improve right now].
        
        Be specific: cite file names, vulnerabilities. Be concise but thorough. ${aiSummaryLang === 'ru' ? 'Respond completely in Russian (Русский язык), translating the labels to Russian (e.g. **Статус безопасности:**, **Главный приоритет:**, **Быстрое улучшение:**).' : 'Respond in English.'}`;
                                    const res = await fetch('/api/chat', {
                                      method: 'POST',
                                      headers: { 'Content-Type': 'application/json' },
                                      body: JSON.stringify({ messages: [{ role: 'user', content: summaryPrompt }] }),
                                    });
                                    const data = await res.json();
                                    setAiSummary(data.ok ? (data.content || 'No response') : `Error: ${data.error}`);
                                  } catch (err) {
                                    setAiSummary('Failed to generate summary. Check API key configuration.');
                                  } finally {
                                    setAiSummaryLoading(false);
                                  }
                                }}
                                className="btn-ai-generate w-full py-2.5 rounded-lg text-xs font-bold tracking-widest uppercase transition-all shadow-md flex items-center justify-center gap-1.5 cursor-pointer"
                              >
                                <span className="material-symbols-outlined text-sm">bolt</span>
                                {t('generate')}
                              </button>
                            </div>
                          </div>
                        )}
                      </div>
                    </motion.div>
                  );
                })()}
              </div>
            </div>

            {/* Urgency Banner */}
            {sevCounts.critical > 0 && (
              <motion.div variants={itemVariants} className="flex items-center gap-4 px-5 py-3 rounded-xl border border-[rgba(239,68,68,0.15)] bg-gradient-to-r from-[rgba(239,68,68,0.06)] to-[rgba(239,68,68,0.01)] shadow-sm">
                <span className="material-symbols-outlined text-[#ef4444] text-[20px] animate-pulse">warning</span>
                <div className="flex-1">
                  <span className="text-[13px] text-[#f4f4f5] font-semibold tracking-wide">{sevCounts.critical} {sevCounts.critical === 1 ? t('criticalIssueRequires') : t('criticalIssuesRequire')} {t('immediateAttention')}</span>
                  <span className="text-[12px] text-[#a1a1aa] ml-2 font-mono">— {sevCounts.high} {t('highSeverityAlsoPending')}</span>
                </div>
                <button onClick={() => { setActiveFilter('critical'); setPage(0); }}
                  className="text-[11px] text-[#ef4444] border border-[rgba(239,68,68,0.25)] hover:bg-[rgba(239,68,68,0.1)] font-bold px-3.5 py-1.2 rounded-lg transition-colors shrink-0 uppercase tracking-wider font-mono">
                  {t('showCritical')}
                </button>
              </motion.div>
            )}

            {/* ── MAIN CONTENT SPLIT ── */}
            <div className="grid grid-cols-1 xl:grid-cols-3 gap-6 items-start">
              
              {/* LEFT: Findings & Toolbar */}
              <div className="xl:col-span-2 space-y-4">
                {/* ── Floating Bulk Action Bar ── */}
                {selectedFindings.size > 0 && (
                  <motion.div
                    initial={{ opacity: 0, y: -12, scale: 0.98 }}
                    animate={{ opacity: 1, y: 0, scale: 1 }}
                    exit={{ opacity: 0, y: -12, scale: 0.98 }}
                    transition={{ duration: 0.3, ease: [0.16, 1, 0.3, 1] }}
                    className="flex items-center justify-between gap-4 px-5 py-3 rounded-xl border border-[rgba(255,255,255,0.06)] bg-[#111113] shadow-xl relative overflow-hidden"
                  >

                    {/* Left: selection info */}
                    <div className="flex items-center gap-4 shrink-0 relative z-10">
                      <div
                        className="flex items-center gap-2 px-3 py-1 rounded-full font-mono font-black text-[10px] uppercase tracking-widest shadow-[0_0_15px_var(--accent-color-soft)]"
                        style={{
                          background: 'var(--accent-color-soft)',
                          border: '1px solid var(--accent-color-line)',
                          color: 'var(--accent-color)'
                        }}
                      >
                        <span
                          className="inline-flex items-center justify-center w-4 h-4 rounded-full text-[9px] font-black"
                          style={{ background: 'var(--accent-color)', color: 'var(--accent-color-on-text)' }}
                        >
                          {selectedFindings.size}
                        </span>
                        SELECTED
                      </div>
                      <button
                        onClick={() => setSelectedFindings(new Set())}
                        className="flex items-center gap-1.5 text-[10px] font-mono font-bold uppercase tracking-wider transition-colors duration-200 cursor-pointer text-[#52525b] hover:text-[var(--accent-color)]"
                      >
                        <span className="material-symbols-outlined text-[13px]">close</span>
                        CLEAR SELECTION
                      </button>
                    </div>

                    {/* Divider */}
                    <div className="w-px self-stretch" style={{ background: 'rgba(255,255,255,0.06)' }} />

                    {/* Right: actions */}
                    <div className="flex items-center gap-2.5 relative z-10">
                      {/* Fix Mode select */}
                      <div className="relative">
                        <select
                          value={configAutostartFixes ? 'auto' : 'review'}
                          onChange={async (e) => {
                            const auto = e.target.value === 'auto';
                            setConfigAutostartFixes(auto);
                            await handleSaveConfig({ autostartFixes: auto });
                          }}
                          className="appearance-none pl-3.5 pr-8 py-1.5 rounded-lg text-[11px] font-mono font-semibold uppercase tracking-wider outline-none cursor-pointer transition-all duration-300 bg-[rgba(255,255,255,0.015)] border border-[rgba(255,255,255,0.04)] text-[#a1a1aa] hover:text-white hover:border-[var(--accent-color-line)] focus:border-[var(--accent-color)] focus:bg-[rgba(0,0,0,0.2)] focus:shadow-[0_0_10px_var(--accent-color-soft)]"
                          style={{
                            backgroundImage: `url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='10' height='10' viewBox='0 0 24 24' fill='none' stroke='%2371717a' stroke-width='2'%3E%3Cpath d='m6 9 6 6 6-6'/%3E%3C/svg%3E")`,
                            backgroundRepeat: 'no-repeat',
                            backgroundPosition: 'right 8px center',
                            backgroundSize: '8px'
                          }}
                        >
                          <option value="auto" className="bg-[#111113] text-[#f4f4f5]">Fix Mode: Auto</option>
                          <option value="review" className="bg-[#111113] text-[#f4f4f5]">Fix Mode: Review First</option>
                        </select>
                      </div>

                      {/* Fix Selected – primary accent filled with translate hover */}
                      <button
                        onClick={handleBulkFix}
                        className="flex items-center gap-1.5 pl-3.5 pr-4 py-1.5 rounded-lg text-[11px] font-bold uppercase tracking-widest transition-all duration-300 cursor-pointer bg-[var(--accent-color)] hover:bg-[var(--accent-color-hover)] text-[var(--accent-color-on-text)] shadow-[0_0_12px_var(--accent-color-line)] hover:shadow-[0_0_22px_var(--accent-color-soft)] hover:-translate-y-[1px] active:translate-y-0"
                      >
                        <span className="material-symbols-outlined text-[14px]">{bulkFixCopied ? 'check' : 'auto_fix_high'}</span>
                        {bulkFixCopied ? 'Copied!' : 'Fix Selected'}
                      </button>

                      {/* Ignore Selected – secondary ghost outline with translate hover */}
                      <button
                        onClick={() => setBulkIgnoreModalOpen(true)}
                        className="flex items-center gap-1.5 pl-3.5 pr-4 py-1.5 rounded-lg text-[11px] font-bold uppercase tracking-widest transition-all duration-300 cursor-pointer bg-[rgba(255,255,255,0.015)] border border-[rgba(255,255,255,0.04)] text-[#a1a1aa] hover:text-white hover:bg-[var(--accent-color-soft)] hover:border-[var(--accent-color-line)] hover:-translate-y-[1px] active:translate-y-0 hover:shadow-[0_0_12px_var(--accent-color-soft)]"
                      >
                        <span className="material-symbols-outlined text-[14px]">do_not_disturb_on</span>
                        Ignore Selected
                      </button>
                    </div>
                  </motion.div>
                )}

                {/* ── Toolbar ── */}
                <motion.div variants={itemVariants} className="space-y-2 bg-[rgba(255,255,255,0.01)] border border-[rgba(255,255,255,0.06)] p-3 rounded-xl shadow-sm">
                  {/* Row 1: Filters & Search */}
                  <div className="flex items-center justify-between gap-4 flex-wrap">
                    <div className="flex items-center gap-3 flex-wrap">
                      {/* Select All Checkbox */}
                      <div className="flex items-center justify-center border-r border-[rgba(255,255,255,0.06)] pr-3">
                        <input
                          type="checkbox"
                          checked={filteredFindings.length > 0 && filteredFindings.every(f => selectedFindings.has(f.id))}
                          ref={el => {
                            if (el) {
                              const selCount = filteredFindings.filter(f => selectedFindings.has(f.id)).length;
                              el.indeterminate = selCount > 0 && selCount < filteredFindings.length;
                            }
                          }}
                          onChange={(e) => {
                            const checked = e.target.checked;
                            setSelectedFindings(prev => {
                              const next = new Set(prev);
                              filteredFindings.forEach(f => {
                                if (checked) {
                                  next.add(f.id);
                                } else {
                                  next.delete(f.id);
                                }
                              });
                              return next;
                            });
                          }}
                          className="accent-[var(--accent-color)] cursor-pointer select-checkbox w-3.5 h-3.5 rounded bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.08)]"
                          title="Select All"
                        />
                      </div>

                      {/* Scope Toggles */}
                      <div className="flex items-center gap-1.5 border-r border-[rgba(255,255,255,0.06)] pr-3">
                        <button
                          onClick={() => setScopeType(scopeType === 'all' ? 'activeFile' : 'all')}
                          className={`flex items-center gap-1 px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider transition-all border ${
                            scopeType === 'activeFile'
                              ? 'border-[var(--accent-color-line)] bg-[var(--accent-color-soft)] text-[var(--accent-color)] font-semibold'
                              : 'border-[rgba(255,255,255,0.03)] bg-[rgba(255,255,255,0.015)] text-[#a1a1aa] hover:text-white'
                          }`}
                          title="Toggle Active File scope"
                        >
                          <span className="material-symbols-outlined text-[12px]">{scopeType === 'activeFile' ? 'description' : 'folder_copy'}</span>
                          {scopeType === 'activeFile' ? 'Active' : 'All'}
                        </button>

                        {scopeType === 'activeFile' && (
                          <select
                            value={activeFilePath}
                            onChange={e => { setActiveFilePath(e.target.value); setPage(0); }}
                            className="bg-[rgba(255,255,255,0.015)] border border-[rgba(255,255,255,0.04)] rounded px-1.5 py-0.5 text-[10px] text-white outline-none hover:border-[rgba(255,255,255,0.08)] transition-all cursor-pointer font-mono max-w-[120px]"
                          >
                            <option value="">-- Active File --</option>
                            {uniqueFilePaths.map(path => (
                              <option key={path} value={path}>{path.split('/').pop()}</option>
                            ))}
                          </select>
                        )}
                      </div>

                      {/* Severity pills */}
                      <div className="flex items-center gap-1.5 flex-wrap">
                        {[
                          { key: null, label: t('filterAll'), count: findings?.length ?? 0 },
                          { key: 'critical', label: t('severityCritical'), count: sevCounts.critical, color: '#ef4444' },
                          { key: 'high', label: t('severityHigh'), count: sevCounts.high, color: '#f97316' },
                          { key: 'medium', label: t('severityMedium'), count: sevCounts.medium, color: '#eab308' },
                          { key: 'low', label: t('severityLow'), count: sevCounts.low, color: '#3f3f46' },
                        ].map(f => (
                          <button key={f.key ?? 'all'} onClick={() => { setActiveFilter(f.key); setPage(0); }}
                            className={`flex items-center gap-1.5 px-2.5 py-1 rounded-md text-[11px] font-medium transition-all border ${activeFilter === f.key ? 'border-[rgba(255,255,255,0.12)] bg-[rgba(255,255,255,0.06)] text-white shadow-sm font-semibold' : 'border-[rgba(255,255,255,0.03)] bg-[rgba(255,255,255,0.015)] text-[#a1a1aa] hover:text-white hover:bg-[rgba(255,255,255,0.03)]'}`}>
                            {f.color && <span className="w-1.5 h-1.5 rounded-full shrink-0" style={{ backgroundColor: f.color }} />}{f.label} 
                            <span className="text-[10px] opacity-50 font-mono ml-0.5">({f.count})</span>
                          </button>
                        ))}
                      </div>
                    </div>

                    {/* Search */}
                    <div className="flex items-center gap-2 flex-1 md:flex-none justify-end min-w-[240px]">
                      <div className="relative w-full">
                        <span className="material-symbols-outlined text-[14px] text-[#52525b] absolute left-2.5 top-1/2 -translate-y-1/2">search</span>
                        <input 
                          value={searchQuery} 
                          onChange={e => { setSearchQuery(e.target.value); setPage(0); }} 
                          placeholder={t('searchPlaceholder')}
                          className="w-full bg-[rgba(0,0,0,0.15)] border border-[rgba(255,255,255,0.05)] rounded-md pl-8 pr-3 py-1 text-[12px] text-[#f4f4f5] placeholder:text-[#52525b] outline-none focus:border-[var(--accent-color)] focus:bg-[rgba(0,0,0,0.25)] transition-all shadow-inner font-sans" 
                        />
                      </div>
                      <div className="text-[12px] text-[#52525b] font-sans shrink-0 bg-[rgba(255,255,255,0.02)] px-2 py-1 rounded-md border border-[rgba(255,255,255,0.03)]">
                        {filteredFindings.length} ISSUES
                      </div>
                    </div>
                  </div>

                  <div className="h-px w-full bg-[rgba(255,255,255,0.04)]" />

                  {/* Row 2: Selects & Active Filters */}
                  <div className="flex items-center justify-between gap-4 flex-wrap">
                    <div className="flex items-center gap-2 flex-wrap">
                      {/* Product filter */}
                      {activeProducts.length > 0 && (
                        <select 
                          value={productFilter ?? ''} 
                          onChange={e => { setProductFilter(e.target.value ? Number(e.target.value) : null); setPage(0); }}
                          className="bg-[rgba(255,255,255,0.015)] border border-[rgba(255,255,255,0.04)] rounded-md px-2.5 py-1 text-[12px] uppercase font-sans tracking-wider text-[#a1a1aa] hover:text-white hover:border-[rgba(255,255,255,0.08)] transition-all cursor-pointer appearance-none pr-6 outline-none shadow-sm"
                          style={{ 
                            backgroundImage: `url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='10' height='10' viewBox='0 0 24 24' fill='none' stroke='%2371717a' stroke-width='2'%3E%3Cpath d='m6 9 6 6 6-6'/%3E%3C/svg%3E")`, 
                            backgroundRepeat: 'no-repeat', 
                            backgroundPosition: 'right 6px center',
                            backgroundSize: '8px'
                          }}
                        >
                          <option value="">{t('allProjects')}</option>
                          {activeProducts.map(p => <option key={p.id} value={p.id}>{p.name}</option>)}
                        </select>
                      )}
                      
                      <select 
                        value={statusFilter} 
                        onChange={e => { setStatusFilter(e.target.value); setPage(0); }}
                        className="bg-[rgba(255,255,255,0.015)] border border-[rgba(255,255,255,0.04)] rounded-md px-2.5 py-1 text-[12px] uppercase font-sans tracking-wider text-[#a1a1aa] hover:text-white hover:border-[rgba(255,255,255,0.08)] transition-all cursor-pointer appearance-none pr-6 outline-none shadow-sm"
                        style={{ 
                          backgroundImage: `url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='10' height='10' viewBox='0 0 24 24' fill='none' stroke='%2371717a' stroke-width='2'%3E%3Cpath d='m6 9 6 6 6-6'/%3E%3C/svg%3E")`, 
                          backgroundRepeat: 'no-repeat', 
                          backgroundPosition: 'right 6px center',
                          backgroundSize: '8px'
                        }}
                      >
                        <option value="all">{t('statusAll')}</option>
                        <option value="open">{t('statusOpen')}</option>
                        <option value="triage">{t('statusTriage')}</option>
                        <option value="false_positive">{t('statusFalsePositive')}</option>
                        <option value="risk_accepted">{t('statusAccepted')}</option>
                      </select>
                      
                      <select 
                        value={groupBy} 
                        onChange={e => { setGroupBy(e.target.value as GroupBy); setPage(0); }}
                        className="bg-[rgba(255,255,255,0.015)] border border-[rgba(255,255,255,0.04)] rounded-md px-2.5 py-1 text-[12px] uppercase font-sans tracking-wider text-[#a1a1aa] hover:text-white hover:border-[rgba(255,255,255,0.08)] transition-all cursor-pointer appearance-none pr-6 outline-none shadow-sm"
                        style={{ 
                          backgroundImage: `url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='10' height='10' viewBox='0 0 24 24' fill='none' stroke='%2371717a' stroke-width='2'%3E%3Cpath d='m6 9 6 6 6-6'/%3E%3C/svg%3E")`, 
                          backgroundRepeat: 'no-repeat', 
                          backgroundPosition: 'right 6px center',
                          backgroundSize: '8px'
                        }}
                      >
                        <option value="none">{t('flatList')}</option>
                        <option value="severity">{t('groupSeverity')}</option>
                        <option value="title">{t('groupTitle')}</option>
                        <option value="file">{t('groupFile')}</option>
                        <option value="scanner">{t('groupScanner')}</option>
                        <option value="product">{t('groupProject')}</option>
                      </select>
                      
                      <select 
                        value={sortBy} 
                        onChange={e => setSortBy(e.target.value as SortBy)}
                        className="bg-[rgba(255,255,255,0.015)] border border-[rgba(255,255,255,0.04)] rounded-md px-2.5 py-1 text-[12px] uppercase font-sans tracking-wider text-[#a1a1aa] hover:text-white hover:border-[rgba(255,255,255,0.08)] transition-all cursor-pointer appearance-none pr-6 outline-none shadow-sm"
                        style={{ 
                          backgroundImage: `url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='10' height='10' viewBox='0 0 24 24' fill='none' stroke='%2371717a' stroke-width='2'%3E%3Cpath d='m6 9 6 6 6-6'/%3E%3C/svg%3E")`, 
                          backgroundRepeat: 'no-repeat', 
                          backgroundPosition: 'right 6px center',
                          backgroundSize: '8px'
                        }}
                      >
                        <option value="severity">{t('sortSeverity')}</option>
                        <option value="title">{t('sortTitle')}</option>
                        <option value="file">{t('sortFile')}</option>
                      </select>
                      {expandedIds.size > 0 && (
                        <button
                          onClick={() => setExpandedIds(new Set())}
                          className="flex items-center gap-1 px-2.5 py-1 rounded-md text-[11px] font-bold uppercase tracking-wider transition-all border border-[var(--accent-color-line)] bg-[var(--accent-color-soft)] text-[var(--accent-color)] hover:bg-[var(--accent-color-hover)] hover:text-[var(--accent-color-on-text)]"
                          title="Collapse all findings"
                        >
                          <span className="material-symbols-outlined text-[14px]">unfold_less</span>
                          {i18n.language?.startsWith('ru') ? 'Свернуть все' : 'Collapse All'}
                        </button>
                      )}
                    </div>

                    {/* Active filters summary */}
                    {(activeFilter || productFilter !== null || statusFilter !== 'all' || searchQuery) && (
                      <div className="flex items-center gap-1.5 flex-wrap">
                        {productFilter !== null && (
                          <span className="inline-flex items-center gap-1 text-[9px] text-[#e4e4e7] bg-[rgba(255,255,255,0.04)] px-2 py-0.5 rounded border border-[rgba(255,255,255,0.03)] font-mono uppercase">
                            <span className="material-symbols-outlined text-[10px] text-[#a1a1aa]">folder</span>{productMap.get(productFilter)?.name || `Project #${productFilter}`}
                            <button onClick={() => setProductFilter(null)} className="ml-1 text-[#a1a1aa] hover:text-white material-symbols-outlined text-[11px] leading-none">close</button>
                          </span>
                        )}
                        {activeFilter && (
                          <span className="inline-flex items-center gap-1 text-[9px] text-[#e4e4e7] bg-[rgba(255,255,255,0.04)] px-2 py-0.5 rounded border border-[rgba(255,255,255,0.03)] font-mono uppercase">
                            <span className="w-1 h-1 rounded-full" style={{ backgroundColor: sevDot(activeFilter) }} />{activeFilter}
                            <button onClick={() => setActiveFilter(null)} className="ml-1 text-[#a1a1aa] hover:text-white material-symbols-outlined text-[11px] leading-none">close</button>
                          </span>
                        )}
                        {statusFilter !== 'all' && (
                          <span className="inline-flex items-center gap-1 text-[9px] text-[#e4e4e7] bg-[rgba(255,255,255,0.04)] px-2 py-0.5 rounded border border-[rgba(255,255,255,0.03)] font-mono uppercase">
                            {statusFilter.replace('_', ' ')}
                            <button onClick={() => setStatusFilter('all')} className="ml-1 text-[#a1a1aa] hover:text-white material-symbols-outlined text-[11px] leading-none">close</button>
                          </span>
                        )}
                        {searchQuery && (
                          <span className="inline-flex items-center gap-1 text-[9px] text-[#e4e4e7] bg-[rgba(255,255,255,0.04)] px-2 py-0.5 rounded border border-[rgba(255,255,255,0.03)] font-mono uppercase">
                            "{searchQuery}"
                            <button onClick={() => setSearchQuery('')} className="ml-1 text-[#a1a1aa] hover:text-white material-symbols-outlined text-[11px] leading-none">close</button>
                          </span>
                        )}
                        <button onClick={() => { setActiveFilter(null); setProductFilter(null); setStatusFilter('all'); setSearchQuery(''); setPage(0); }}
                          className="text-[9px] transition-colors ml-1 font-bold bg-[rgba(239,68,68,0.08)] hover:bg-[rgba(239,68,68,0.15)] text-[#ef4444] px-2 py-0.5 rounded font-mono uppercase tracking-wider">{t('SimpleDashboardPage.clearAll')}</button>
                      </div>
                    )}
                  </div>
                </motion.div>

                {/* ── Findings ── */}
                {filteredFindings.length === 0 ? (
                  <motion.div variants={itemVariants} className="py-12 text-center text-[12px] text-[#71717a] bg-[rgba(255,255,255,0.01)] rounded-xl border border-[rgba(255,255,255,0.04)] font-mono uppercase tracking-wider">
                    <span className="material-symbols-outlined text-[36px] text-[#3f3f46] mb-2 block">search_off</span>
                    {searchQuery ? t('SimpleDashboardPage.noIssuesForQuery', { query: searchQuery }) : t('SimpleDashboardPage.noIssues')}
                  </motion.div>
                ) : groupBy !== 'none' && groups ? (
                  <motion.div variants={itemVariants} className="space-y-3">
                    {groups.map(([key, items]) => {
                      const isCollapsed = collapsedGroups.has(key);
                      return (
                        <div key={key} className="border border-[rgba(255,255,255,0.06)] rounded-xl overflow-hidden bg-[rgba(255,255,255,0.01)] shadow-sm">
                          <div className="flex items-center gap-3 px-4 py-2.5 hover:bg-[rgba(255,255,255,0.02)] border-b border-[rgba(255,255,255,0.02)]">
                            <input
                              type="checkbox"
                              checked={items.every(item => selectedFindings.has(item.id))}
                              ref={el => {
                                if (el) {
                                  const isAllSel = items.every(item => selectedFindings.has(item.id));
                                  el.indeterminate = items.some(item => selectedFindings.has(item.id)) && !isAllSel;
                                }
                              }}
                              onChange={(e) => {
                                const checked = e.target.checked;
                                setSelectedFindings(prev => {
                                  const next = new Set(prev);
                                  items.forEach(item => {
                                    if (checked) {
                                      next.add(item.id);
                                    } else {
                                      next.delete(item.id);
                                    }
                                  });
                                  return next;
                                });
                              }}
                              onClick={e => e.stopPropagation()}
                              className="accent-[var(--accent-color)] cursor-pointer select-checkbox w-3.5 h-3.5 rounded bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.08)]"
                            />
                            <button onClick={() => toggleGroup(key)} className="flex-1 text-left flex items-center gap-3">
                              <span className={`material-symbols-outlined text-[14px] text-[#71717a] transition-transform ${isCollapsed ? '' : 'rotate-90'}`}>chevron_right</span>
                              {groupBy === 'severity' && <span className="w-2 h-2 rounded-full shadow-sm" style={{ backgroundColor: sevDot(key) }} />}
                              <span className="text-[12px] text-[#f4f4f5] font-semibold truncate flex-1 tracking-wide">{key}</span>
                              <span className="text-[10px] text-[#a1a1aa] bg-[rgba(255,255,255,0.06)] px-2 py-0.5 rounded-md font-mono">{items.length}</span>
                            </button>
                          </div>
                          {!isCollapsed && (
                            <div className="divide-y divide-[rgba(255,255,255,0.03)]">
                              {items.map(f => (
                                <FindingRow
                                  key={f.id}
                                  f={f}
                                  isExpanded={expandedIds.has(f.id)}
                                  onToggle={() => {
                                    setExpandedIds(prev => {
                                      const next = new Set(prev);
                                      if (next.has(f.id)) next.delete(f.id);
                                      else next.add(f.id);
                                      return next;
                                    });
                                  }}
                                  productMap={productMap}
                                  setProductFilter={setProductFilter}
                                  setPage={setPage}
                                  handleTriage={handleTriage}
                                  onNavigateToChat={onNavigateToChat}
                                  isSelected={selectedFindings.has(f.id)}
                                  isTriaging={triagingIds.has(f.id)}
                                  onToggleSelect={() => {
                                    setSelectedFindings(prev => {
                                      const next = new Set(prev);
                                      if (next.has(f.id)) {
                                        next.delete(f.id);
                                      } else {
                                        next.add(f.id);
                                      }
                                      return next;
                                    });
                                  }}
                                />
                              ))}
                            </div>
                          )}
                        </div>
                      );
                    })}
                  </motion.div>
                ) : (
                  <motion.div variants={itemVariants} className="space-y-3">
                    <div className="border border-[rgba(255,255,255,0.06)] rounded-xl overflow-hidden divide-y divide-[rgba(255,255,255,0.04)] bg-[rgba(255,255,255,0.01)] shadow-sm">
                      {pagedFindings.map(f => (
                        <FindingRow
                          key={f.id}
                          f={f}
                          isExpanded={expandedIds.has(f.id)}
                          onToggle={() => {
                            setExpandedIds(prev => {
                              const next = new Set(prev);
                              if (next.has(f.id)) next.delete(f.id);
                              else next.add(f.id);
                              return next;
                            });
                          }}
                          productMap={productMap}
                          setProductFilter={setProductFilter}
                          setPage={setPage}
                          handleTriage={handleTriage}
                          onNavigateToChat={onNavigateToChat}
                          isSelected={selectedFindings.has(f.id)}
                          isTriaging={triagingIds.has(f.id)}
                          onToggleSelect={() => {
                            setSelectedFindings(prev => {
                              const next = new Set(prev);
                              if (next.has(f.id)) {
                                next.delete(f.id);
                              } else {
                                next.add(f.id);
                              }
                              return next;
                            });
                          }}
                        />
                      ))}
                    </div>
                    {totalPages > 1 && (
                      <div className="flex items-center justify-between pt-1 px-1">
                        <button onClick={() => setPage(p => Math.max(0, p - 1))} disabled={page === 0} className="text-[11px] text-[#a1a1aa] hover:text-[#f4f4f5] disabled:opacity-30 flex items-center gap-1 transition-colors font-bold uppercase tracking-wider font-mono bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.04)] px-2.5 py-1 rounded-md hover:bg-[rgba(255,255,255,0.04)]">
                          <span className="material-symbols-outlined text-[14px]">chevron_left</span>{t('SimpleDashboardPage.previous')}</button>
                        <div className="flex items-center gap-1">
                          {Array.from({ length: Math.min(totalPages, 7) }, (_, i) => {
                            const p = totalPages <= 7 ? i : page <= 3 ? i : page >= totalPages - 4 ? totalPages - 7 + i : page - 3 + i;
                            return (
                              <button key={p} onClick={() => setPage(p)}
                                className={`w-6 h-6 rounded-md text-[10px] font-bold font-mono transition-all ${p === page ? 'bg-[rgba(255,255,255,0.08)] text-white border border-[rgba(255,255,255,0.12)]' : 'text-[#71717a] hover:text-[#e4e4e7] hover:bg-[rgba(255,255,255,0.02)]'}`}>
                                {p + 1}
                              </button>
                            );
                          })}
                        </div>
                        <button onClick={() => setPage(p => Math.min(totalPages - 1, p + 1))} disabled={page >= totalPages - 1} className="text-[11px] text-[#a1a1aa] hover:text-[#f4f4f5] disabled:opacity-30 flex items-center gap-1 transition-colors font-bold uppercase tracking-wider font-mono bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.04)] px-2.5 py-1 rounded-md hover:bg-[rgba(255,255,255,0.04)]">
                          Next<span className="material-symbols-outlined text-[14px]">chevron_right</span></button>
                      </div>
                    )}
                  </motion.div>
                )}
              </div>

              {/* RIGHT: AI IDE Prompts (sticky) */}
              <div className="xl:col-span-1 sticky top-6 space-y-6">
                <SecureCoderPanel activeProducts={activeProducts} findings={findings} />
                <InlineAIPrompts activeProducts={activeProducts} findings={findings} />
              </div>
            </div>
          </div>
        </motion.div>
      </div>
      {/* Projects Panel */}
      <div className="relative flex shrink-0 h-full">
        {/* Toggle Button for Projects Panel */}
        <button
          onClick={handleToggleProjectsPanel}
          className={`absolute -left-6 top-[calc(50%-120px)] -translate-y-1/2 z-40 flex items-center justify-center w-6 h-20 rounded-l-lg transition-all duration-300 ${
            isProjectsPanelOpen
              ? 'bg-surface-bright border border-r-0 border-[rgba(255,255,255,0.06)] text-[var(--accent-color)] hover:bg-surface'
              : 'bg-surface border border-r-0 border-[rgba(255,255,255,0.06)] text-[#71717a] hover:text-[var(--accent-color)] hover:bg-surface-bright'
          }`}
          title={isProjectsPanelOpen ? t('SimpleDashboardPage.hideProjects') : t('SimpleDashboardPage.showProjects')}
        >
          <span
            className="material-symbols-outlined text-[14px] transition-transform duration-300"
            style={{ transform: isProjectsPanelOpen ? 'rotate(0deg)' : 'rotate(180deg)' }}
          >
            chevron_right
          </span>
        </button>

        <AnimatePresence initial={false}>
          {isProjectsPanelOpen && (
            <motion.div
              initial={{ width: 0, opacity: 0 }}
              animate={{ width: 340, opacity: 1 }}
              exit={{ width: 0, opacity: 0 }}
              transition={{ duration: 0.3, ease: [0.4, 0, 0.2, 1] }}
              className="w-[340px] shrink-0 border-l border-[rgba(255,255,255,0.06)] flex flex-col overflow-hidden bg-surface"
            >
              <ScanPanel onScanComplete={() => { refreshFindings?.(); refreshMetrics?.(); }} />
            </motion.div>
          )}
        </AnimatePresence>
      </div>

      {/* Bulk Ignore Triage Justification Modal */}
      <AnimatePresence>
        {bulkIgnoreModalOpen && (
          <div className="fixed inset-0 bg-black/70 backdrop-blur-sm z-[100] flex items-center justify-center p-4">
            <motion.div
              initial={{ opacity: 0, scale: 0.95 }}
              animate={{ opacity: 1, scale: 1 }}
              exit={{ opacity: 0, scale: 0.95 }}
              className="bg-[#0e0e11] border border-[rgba(255,255,255,0.08)] rounded-xl w-[400px] p-6 shadow-[0_24px_50px_rgba(0,0,0,0.85)] flex flex-col space-y-4 text-left relative overflow-hidden"
            >
              <div>
                <h3 className="text-[12px] font-bold text-white uppercase tracking-wider">Ignore {selectedFindings.size} Selected Findings</h3>
                <p className="text-[11px] text-[#71717a] mt-1 leading-normal">
                  Choose the triage status and justification reason to suppress these findings in bulk.
                </p>
              </div>

              <div className="space-y-1">
                <label className="text-[10px] text-[#71717a] font-bold uppercase tracking-wider">Triage Justification</label>
                <select
                  value={bulkIgnoreReason}
                  onChange={e => setBulkIgnoreReason(e.target.value)}
                  className="w-full bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.06)] rounded-md px-3 py-1.5 text-[11px] text-white outline-none focus:border-[var(--accent-color)] cursor-pointer"
                >
                  <option value="False Positive">False Positive (Inaccurate finding)</option>
                  <option value="Accepted Risk">Accepted Risk (Accept risk, do not fix)</option>
                  <option value="Won't Fix">Won't Fix (Acknowledge, but keep as is)</option>
                </select>
              </div>

              <div className="flex justify-end gap-2 pt-2 border-t border-[rgba(255,255,255,0.06)]">
                <button
                  onClick={() => setBulkIgnoreModalOpen(false)}
                  className="px-3.5 py-1.5 bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.06)] hover:bg-[rgba(255,255,255,0.05)] text-[#a1a1aa] hover:text-white rounded-lg text-[10px] font-bold uppercase tracking-wider transition-colors cursor-pointer"
                >
                  Cancel
                </button>
                <button
                  onClick={handleBulkIgnore}
                  disabled={bulkIgnoring}
                  className="px-4 py-1.5 bg-red-600 hover:bg-red-500 text-white rounded-lg text-[10px] font-bold uppercase tracking-wider transition-colors cursor-pointer disabled:opacity-50"
                >
                  {bulkIgnoring ? 'Ignoring...' : 'Ignore Findings'}
                </button>
              </div>
            </motion.div>
          </div>
        )}
      </AnimatePresence>

      {/* ── PREMIUM SCANNING OVERLAY ── */}
      <AnimatePresence>
        {globalScanning && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 z-[200] flex items-center justify-center bg-background/90 backdrop-blur-md"
          >
            <div className="w-[500px] border border-[rgba(255,255,255,0.08)] bg-surface rounded-2xl shadow-2xl overflow-hidden flex flex-col p-6 font-sans">
              <div className="flex flex-col items-center gap-4 text-center pb-6 border-b border-[rgba(255,255,255,0.06)]">
                {/* pulsing visual radar/circle */}
                <div className="relative w-16 h-16 flex items-center justify-center">
                  <div className="absolute inset-0 border-2 border-[var(--accent-color-line)] rounded-full animate-ping opacity-25" />
                  <div className="w-12 h-12 border-2 border-dashed border-[var(--accent-color)] rounded-full animate-spin flex items-center justify-center">
                    <span className="material-symbols-outlined text-[20px] text-[var(--accent-color)]">
                      shield
                    </span>
                  </div>
                </div>
                <div>
                  <h3 className="text-sm font-bold text-white uppercase tracking-widest font-mono">
                    AI Security Triage Audit Active
                  </h3>
                  <p className="text-xs text-[#71717a] mt-1 font-mono truncate max-w-[400px]">
                    Directory: {globalScanPath}
                  </p>
                </div>
              </div>

              {/* Progress bar */}
              <div className="py-4 text-left">
                <div className="flex justify-between items-center text-[10px] text-[#71717a] font-mono mb-1.5 uppercase">
                  <span>Phase: {scanPhases[globalScanPhase].name}</span>
                  <span>{globalScanElapsed}s elapsed</span>
                </div>
                <div className="w-full h-1.5 bg-surface-bright rounded-full overflow-hidden">
                  <div 
                    className="h-full bg-[var(--accent-color)] rounded-full transition-all duration-500 shadow-[0_0_8px_var(--accent-color-line)]"
                    style={{ width: `${((globalScanPhase + 1) / scanPhases.length) * 100}%` }}
                  />
                </div>
                <p className="text-[10px] text-[#52525b] mt-2 font-mono italic">
                  — {scanPhases[globalScanPhase].desc}
                </p>
              </div>

              {/* Console log box */}
              <div className="bg-black/40 border border-[rgba(255,255,255,0.04)] rounded-lg p-4 font-mono text-[10px] text-[#52525b] h-32 overflow-y-auto flex flex-col justify-end gap-1.5 text-left">
                {globalScanLogs.map((log, i) => (
                  <div key={i} className={i === globalScanLogs.length - 1 ? 'text-[#a1a1aa]' : ''}>
                    <span className="text-[#3f3f46] mr-1.5">$</span>{log}
                  </div>
                ))}
              </div>

              <div className="pt-4 mt-2 text-center text-[9px] text-[#3f3f46] font-mono uppercase tracking-[0.2em]">
                Do not close or reload the browser window
              </div>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
};
