import React, { useState, useEffect, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { useMetrics } from '../hooks/useMetrics';
import { FileBrowser } from '../components/FileBrowser';
import { securityService } from '../services/securityService';
import { useFindings } from '../hooks/useFindings';
import { useCopilotStore } from '../store/CopilotStore';
import type { Finding } from '../types';
import { useCountUp } from '../hooks/useCountUp';
import { ProgressRing } from '../ui/ProgressRing';
import { useProducts } from '../hooks/useProducts';



export const CCOverviewPanel: React.FC = () => {
  const { t } = useTranslation('pages');
  const { metrics, loading, refresh } = useMetrics();
  const { findings, loading: findingsLoading, refresh: refreshFindings } = useFindings() as any;
  const { products, loading: productsLoading } = useProducts();
  const { setIsOpen, setContext } = useCopilotStore();

  const [selectedProductId, setSelectedProductId] = useState<number | null>(null);
  const [showBrowser, setShowBrowser] = useState(false);
  const [scanning, setScanning] = useState(false);
  const [selectedFindingId, setSelectedFindingId] = useState<number | null>(null);
  const [aiSummary, setAiSummary] = useState<string>('');
  const [aiSummaryLoading, setAiSummaryLoading] = useState<boolean>(true);
  const [remediationPlan, setRemediationPlan] = useState<string>('');
  const [remediationLoading, setRemediationLoading] = useState<boolean>(false);
  const [copied, setCopied] = useState(false);

  // CC Tabs state
  const [leftTab, setLeftTab] = useState<'summary' | 'scans'>('summary');
  const [workbenchTab, setWorkbenchTab] = useState<'active' | 'resolved' | 'all'>('active');
  const [workbenchSeverityFilter, setWorkbenchSeverityFilter] = useState<string | null>(null);
  const [rightTab, setRightTab] = useState<'action_plan' | 'details' | 'code'>('action_plan');

  const fetchAISummary = useCallback(async () => {
    setAiSummaryLoading(true);
    try {
      const summary = await securityService.getAISummary(selectedProductId);
      setAiSummary(summary);
    } catch (err) {
      console.error('Failed to fetch AI summary', err);
      setAiSummary(t('pages.ccoverview.failedAiSummary', 'Failed to load AI summary.'));
    } finally {
      setAiSummaryLoading(false);
    }
  }, [selectedProductId, t]);

  useEffect(() => {
    fetchAISummary();
  }, [fetchAISummary]);

  // Product specific filtering
  const productFindings = selectedProductId
    ? (findings || []).filter((f: Finding) => f.product_id === selectedProductId)
    : (findings || []);



  const workbenchFindings = productFindings.filter((f: Finding) => {
    const status = (f.status || 'open').toLowerCase();
    const isResolved = ['resolved', 'closed', 'false_positive', 'risk_accepted'].includes(status);
    
    if (workbenchTab === 'active' && isResolved) return false;
    if (workbenchTab === 'resolved' && !isResolved) return false;
    
    if (workbenchSeverityFilter) {
      if (f.severity?.toUpperCase() !== workbenchSeverityFilter.toUpperCase()) return false;
    }
    
    return true;
  });

  const selectedFinding =
    workbenchFindings.find((f: Finding) => f.id === selectedFindingId) || workbenchFindings[0] || null;

  useEffect(() => {
    if (selectedFindingId === null && workbenchFindings.length > 0) {
      setSelectedFindingId(workbenchFindings[0].id);
    }
  }, [workbenchFindings, selectedFindingId]);

  const handleStartScan = async (path: string) => {
    setShowBrowser(false);
    setScanning(true);
    try {
      await securityService.startScan(path);
      refresh?.();
      refreshFindings?.();
      fetchAISummary();
    } catch (err) {
      console.error('Scan failed', err);
    } finally {
      setScanning(false);
    }
  };

  const handleGlobalRefresh = () => {
    refresh?.();
    refreshFindings?.();
    fetchAISummary();
  };

  // Severity counts for product specific findings
  const sc = { CRITICAL: 0, HIGH: 0, MEDIUM: 0, LOW: 0 };
  productFindings.forEach((f: Finding) => {
    const status = (f.status || 'open').toLowerCase();
    if (!['resolved', 'closed', 'false_positive', 'risk_accepted'].includes(status)) {
      const sev = f.severity?.toUpperCase();
      if (sev === 'CRITICAL') sc.CRITICAL++;
      else if (sev === 'HIGH') sc.HIGH++;
      else if (sev === 'MEDIUM') sc.MEDIUM++;
      else if (sev === 'LOW') sc.LOW++;
    }
  });

  const calculateSecurityScore = (findingsList: Finding[]) => {
    let score = 100;
    findingsList.forEach((f: Finding) => {
      const status = (f.status || 'open').toLowerCase();
      if (!['resolved', 'closed', 'false_positive', 'risk_accepted'].includes(status)) {
        const sev = (f.severity || '').toLowerCase();
        if (sev === 'critical') score -= 10;
        else if (sev === 'high') score -= 4;
        else if (sev === 'medium') score -= 1;
      }
    });
    return Math.max(0, score);
  };

  const getSecurityGrade = (score: number) => {
    if (score >= 90) return 'A';
    if (score >= 75) return 'B';
    if (score >= 60) return 'C';
    if (score >= 45) return 'D';
    return 'F';
  };

  const computedScore = selectedProductId ? calculateSecurityScore(productFindings) : (metrics?.security_score ?? 100);
  const computedGrade = selectedProductId ? getSecurityGrade(computedScore) : (metrics?.security_grade ?? 'A');

  const totalFindingsCount = productFindings.length;
  const resolvedFindingsCount = productFindings.filter((f: Finding) => ['resolved', 'closed', 'false_positive', 'risk_accepted'].includes((f.status || 'open').toLowerCase())).length;
  const openFindingsCount = totalFindingsCount - resolvedFindingsCount;
  const resolvedPct = totalFindingsCount > 0 ? Math.round((resolvedFindingsCount / totalFindingsCount) * 100) : 0;

  const animatedScore = useCountUp(computedScore);
  const animatedCrit = useCountUp(sc.CRITICAL);
  const animatedHigh = useCountUp(sc.HIGH);
  const animatedMed = useCountUp(sc.MEDIUM);
  const animatedLow = useCountUp(sc.LOW);
  const animatedOpen = useCountUp(openFindingsCount);
  const animatedResolvedPct = useCountUp(resolvedPct);

  const crit = sc.CRITICAL;
  const high = sc.HIGH;
  const med = sc.MEDIUM;

  // Filter scans
  const filteredScans = selectedProductId && products.find(p => p.id === selectedProductId)
    ? (metrics?.recent_engagements || []).filter(e => {
        const prod = products.find(p => p.id === selectedProductId);
        return prod && (e.name.toLowerCase().includes(prod.name.toLowerCase()) || prod.name.toLowerCase().includes(e.name.toLowerCase()));
      })
    : (metrics?.recent_engagements || []);

  if (loading || !metrics || productsLoading) {
    return (
      <div className="flex h-full items-center justify-center">
        <span className="text-label-caps text-v2-muted animate-pulse tracking-widest">
          {t('pages.ccoverview.loading', 'LOADING...')}
        </span>
      </div>
    );
  }

  const scoreColor =
    computedScore < 50
      ? 'text-error'
      : computedScore < 70
        ? 'text-warning'
        : 'text-success';
  const gradeBorder =
    computedScore < 50
      ? 'border-error'
      : computedScore < 70
        ? 'border-warning'
        : 'border-success';

  const activePrompt = selectedFinding
    ? `You are a security engineer assigned to remediate a vulnerability.
Finding: ${selectedFinding.title} (${selectedFinding.severity})
File: ${selectedFinding.file_path || selectedFinding.file || 'unknown'}:${selectedFinding.line_number || '0'}
Description: ${selectedFinding.description || selectedFinding.fix_suggestion || selectedFinding.suggestion || 'No description.'}

Provide a secure code patch and step-by-step fix plan.`
    : '';

  return (
    <div className="flex flex-col h-full overflow-hidden bg-v2-bg">
      {showBrowser && (
        <FileBrowser onSelect={handleStartScan} onCancel={() => setShowBrowser(false)} />
      )}

      {scanning && (
        <div className="fixed inset-0 z-[60] flex items-center justify-center bg-black/90 backdrop-blur-sm animate-modal-enter">
          <div className="flex flex-col items-center gap-8">
            <div className="relative flex items-center justify-center">
              <ProgressRing size={96} strokeWidth={3} indeterminate className="animated-ring" />
              <span className="material-symbols-outlined absolute text-[32px] text-primary animate-pulse">
                sync
              </span>
            </div>
            <span className="text-[12px] font-semibold text-primary tracking-[0.3em] uppercase font-mono">
              {t('pages.ccoverview.analyzingSystem', 'Analyzing System...')}
            </span>
          </div>
        </div>
      )}

      {/* Header bar */}
      <div className="px-6 py-4 flex justify-between items-center shrink-0 border-b border-v2-border-soft bg-v2-surface">
        <div className="flex items-center gap-6">
          <div>
            <p className="text-[10px] font-bold text-v2-muted mb-1 tracking-widest uppercase">
              {t('pages.ccoverview.breadcrumb', 'Root // Security')}
            </p>
            <h1 className="text-2xl font-bold text-white tracking-tight">{t('pages.ccoverview.title', 'Security Dashboard')}</h1>
          </div>
          <div className="h-8 w-[1px] bg-v2-border-soft shrink-0 self-end mb-1" />
          <div className="self-end mb-0.5">
            <select
              value={selectedProductId ?? ''}
              onChange={(e) => {
                const val = e.target.value;
                setSelectedProductId(val ? Number(val) : null);
              }}
              className="bg-[#18181b] border border-v2-border-soft rounded-lg px-3 py-1.5 text-xs text-[#a1a1aa] outline-none focus:border-primary/50 transition-colors h-10 cursor-pointer min-w-[200px]"
            >
              <option value="">{t('pages.ccoverview.allProjects', 'All Projects')}</option>
              {products.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.name}
                </option>
              ))}
            </select>
          </div>
        </div>
        <div className="flex items-center gap-3">
          <button onClick={handleGlobalRefresh} className="v2-btn v2-btn-ghost group h-10 px-4">
            <span className="material-symbols-outlined text-[16px] group-hover:rotate-180 transition-transform duration-500">
              refresh
            </span>
            <span className="tracking-widest text-[11px] uppercase font-bold">{t('pages.ccoverview.refresh', 'Refresh')}</span>
          </button>
          <button onClick={() => setShowBrowser(true)} className="v2-btn v2-btn-ghost h-10 px-4">
            <span className="material-symbols-outlined text-[16px]">folder_open</span>
            <span className="tracking-widest text-[11px] uppercase font-bold">{t('pages.ccoverview.dir', 'Dir')}</span>
          </button>
          <button onClick={() => handleStartScan('.')} className="v2-btn v2-btn-red h-10 px-5">
            <span className="material-symbols-outlined text-[16px]">bolt</span>
            <span className="tracking-widest text-[11px] uppercase font-bold">{t('pages.ccoverview.quickScan', 'Quick Scan')}</span>
          </button>
        </div>
      </div>

      {/* KPI strip */}
      <div className="grid grid-cols-5 border-b border-v2-border-soft shrink-0 bg-v2-surface font-sans">
        <div className="col-span-2 px-6 py-4 border-r border-v2-border-soft flex items-center justify-between hover:bg-v2-surface-2 transition-colors">
          <div className="flex items-center gap-4">
            <div
              className={`w-12 h-12 border-2 ${gradeBorder} rounded-lg flex items-center justify-center`}
            >
              <span className={`text-xl font-black ${scoreColor}`}>
                {computedGrade || '–'}
              </span>
            </div>
            <div>
              <div className="text-[10px] font-bold text-v2-muted tracking-widest uppercase mb-1 font-mono">
                {t('pages.ccoverview.health', 'Health')}
              </div>
              <div className={`text-lg font-bold leading-tight ${scoreColor}`}>
                {computedScore >= 80
                  ? t('pages.ccoverview.statusOptimal', 'Optimal')
                  : computedScore >= 60
                    ? t('pages.ccoverview.statusDegraded', 'Degraded')
                    : t('pages.ccoverview.statusCritical', 'Critical')}
              </div>
            </div>
          </div>
          <div className="text-right">
            <div className="text-[11px] text-v2-muted font-mono">{t('pages.ccoverview.score', 'Score')}</div>
            <div className={`text-2xl font-bold font-mono ${scoreColor}`}>{animatedScore}</div>
          </div>
        </div>

        <div className="col-span-2 px-6 py-4 border-r border-v2-border-soft flex flex-col justify-center hover:bg-v2-surface-2 transition-colors">
          <div className="flex items-center justify-between mb-2">
            <span className="text-[10px] font-bold text-v2-muted tracking-widest uppercase font-mono">
              {t('pages.ccoverview.openFindings', 'Open Findings')}
            </span>
            <span
              className={`text-xl font-bold font-mono ${openFindingsCount > 0 ? 'text-error' : 'text-primary'}`}
            >
              {animatedOpen}
            </span>
          </div>
          <div className="flex items-center gap-3 text-[11px] font-bold uppercase font-mono">
            <span className={crit > 0 ? 'text-error' : 'text-v2-muted'}>{t('pages.ccoverview.crit', 'Crit')}: {animatedCrit}</span>
            <span className={high > 0 ? 'text-primary' : 'text-v2-muted'}>
              {t('pages.ccoverview.high', 'High')}: {animatedHigh}
            </span>
            <span className={med > 0 ? 'text-v2-fg-2' : 'text-v2-muted'}>{t('pages.ccoverview.med', 'Med')}: {animatedMed}</span>
            <span className="text-v2-muted">{t('pages.ccoverview.low', 'Low')}: {animatedLow}</span>
          </div>
        </div>

        <div className="col-span-1 px-6 py-4 hover:bg-v2-surface-2 transition-colors flex flex-col justify-between">
          <div className="flex items-center justify-between mb-2">
            <span className="text-[10px] font-bold text-v2-muted tracking-widest uppercase font-mono">
              {t('pages.ccoverview.resolved', 'Resolved')}
            </span>
            <span className="material-symbols-outlined text-[16px] text-v2-muted">
              check_circle
            </span>
          </div>
          <div className="text-2xl font-bold text-white font-mono">{animatedResolvedPct}%</div>
        </div>
      </div>

      {/* Main grid */}
      <div className="flex-1 overflow-y-auto">
        <div className="p-6 grid grid-cols-12 gap-6 max-w-[1600px] mx-auto h-full">
          {/* LEFT COLUMN: Summary + Recent Scans (Tabbed) */}
          <div className="col-span-12 lg:col-span-4 flex flex-col gap-6 h-full">
            <div
              className="v2-card flex flex-col p-0 overflow-hidden flex-1 animate-page-transition luxury-card-hover h-full"
              style={{ animationDelay: '0ms', animationFillMode: 'both' }}
            >
              {/* Custom Header with Tabs */}
              <div className="px-5 py-3 border-b border-v2-border-soft flex justify-between items-center bg-v2-surface shrink-0">
                <div className="flex items-center gap-4">
                  <button
                    onClick={() => setLeftTab('summary')}
                    className={`flex items-center gap-2 pb-0.5 border-b-2 transition-all ${
                      leftTab === 'summary'
                        ? 'border-primary text-white font-bold'
                        : 'border-transparent text-v2-muted hover:text-white'
                    }`}
                  >
                    <span className="material-symbols-outlined text-[14px]">smart_toy</span>
                    <span className="text-[11px] tracking-widest uppercase font-mono">{t('pages.ccoverview.aiSummaryTitle', 'AI_SUMMARY')}</span>
                  </button>
                  <button
                    onClick={() => setLeftTab('scans')}
                    className={`flex items-center gap-2 pb-0.5 border-b-2 transition-all ${
                      leftTab === 'scans'
                        ? 'border-primary text-white font-bold'
                        : 'border-transparent text-v2-muted hover:text-white'
                    }`}
                  >
                    <span className="material-symbols-outlined text-[14px]">history</span>
                    <span className="text-[11px] tracking-widest uppercase font-mono">{t('pages.ccoverview.recentScansTitle', 'RECENT_SCANS')}</span>
                  </button>
                </div>
                {leftTab === 'scans' && (
                  <span className="text-[10px] font-bold text-v2-muted uppercase font-mono">
                    {filteredScans.length} {t('pages.ccoverview.total', 'Total')}
                  </span>
                )}
              </div>

              <div className="p-6 flex-1 overflow-y-auto">
                {leftTab === 'summary' ? (
                  <div className="flex gap-4 items-start">
                    <div>
                      <h3 className="text-base font-bold text-white mb-2 font-display">
                        {t('pages.ccoverview.statusAnalysis', 'Status Analysis')}
                      </h3>
                      <p className="text-sm text-v2-fg-2 leading-relaxed">
                        {aiSummaryLoading ? t('pages.ccoverview.analyzingProject', 'Analyzing project state...') : aiSummary}
                      </p>
                    </div>
                  </div>
                ) : (
                  <div className="divide-y divide-v2-border-soft -mx-6 -my-6">
                    {filteredScans.length === 0 ? (
                      <div className="p-8 text-center text-v2-muted text-[11px] tracking-widest uppercase font-mono">
                        {t('pages.ccoverview.noScans', 'No Scans')}
                      </div>
                    ) : (
                      filteredScans.map((e, i) => (
                        <div
                          key={i}
                          className="flex items-center justify-between p-4 hover:bg-v2-surface-2 transition-colors"
                        >
                          <div>
                            <div className="text-[13px] font-bold text-white mb-1">{e.name}</div>
                            <div className="text-[11px] text-v2-muted font-mono">{e.date}</div>
                          </div>
                          <span
                            className={`text-[10px] font-bold tracking-widest px-3 py-1 rounded border uppercase font-mono ${
                              e.status === 'completed'
                                ? 'border-v2-border-soft text-v2-muted'
                                : 'border-v2-red-line bg-v2-red-soft text-primary'
                            }`}
                          >
                            {e.status === 'completed' ? t('pages.ccoverview.scanDone', 'Done') : t('pages.ccoverview.scanActive', 'Active')}
                          </span>
                        </div>
                      ))
                    )}
                  </div>
                )}
              </div>
            </div>
          </div>

          {/* MIDDLE COLUMN: Workbench (Tabbed with Severity Pills) */}
          <div
            className="col-span-12 lg:col-span-4 v2-card flex flex-col p-0 overflow-hidden h-full animate-page-transition luxury-card-hover"
            style={{ animationDelay: '120ms', animationFillMode: 'both' }}
          >
            <div className="px-5 py-3 border-b border-v2-border-soft flex flex-col gap-2 bg-v2-surface shrink-0">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-4">
                  <button
                    onClick={() => { setWorkbenchTab('active'); setWorkbenchSeverityFilter(null); }}
                    className={`flex items-center gap-1.5 pb-0.5 border-b-2 transition-all ${
                      workbenchTab === 'active'
                        ? 'border-primary text-white font-bold'
                        : 'border-transparent text-v2-muted hover:text-white'
                    }`}
                  >
                    <span className="text-[11px] tracking-widest uppercase font-mono">{t('pages.ccoverview.active', 'ACTIVE')}</span>
                  </button>
                  <button
                    onClick={() => { setWorkbenchTab('resolved'); setWorkbenchSeverityFilter(null); }}
                    className={`flex items-center gap-1.5 pb-0.5 border-b-2 transition-all ${
                      workbenchTab === 'resolved'
                        ? 'border-primary text-white font-bold'
                        : 'border-transparent text-v2-muted hover:text-white'
                    }`}
                  >
                    <span className="text-[11px] tracking-widest uppercase font-mono">{t('pages.ccoverview.resolved', 'RESOLVED')}</span>
                  </button>
                  <button
                    onClick={() => { setWorkbenchTab('all'); setWorkbenchSeverityFilter(null); }}
                    className={`flex items-center gap-1.5 pb-0.5 border-b-2 transition-all ${
                      workbenchTab === 'all'
                        ? 'border-primary text-white font-bold'
                        : 'border-transparent text-v2-muted hover:text-white'
                    }`}
                  >
                    <span className="text-[11px] tracking-widest uppercase font-mono">{t('pages.ccoverview.all', 'ALL')}</span>
                  </button>
                </div>
                <span className="text-[10px] font-mono text-v2-muted">
                  {workbenchFindings.length}
                </span>
              </div>
              
              {/* Severity Pills */}
              <div className="flex items-center gap-1.5 pt-1">
                {(['CRITICAL', 'HIGH', 'MEDIUM', 'LOW'] as const).map(sev => {
                  const isActive = workbenchSeverityFilter === sev;
                  const color = sev === 'CRITICAL' ? 'bg-[#ef4444]/10 border-[#ef4444]/20 text-[#ef4444]' :
                                sev === 'HIGH' ? 'bg-[#f97316]/10 border-[#f97316]/20 text-[#f97316]' :
                                sev === 'MEDIUM' ? 'bg-[#eab308]/10 border-[#eab308]/20 text-[#eab308]' :
                                'bg-[#3f3f46]/10 border-[#3f3f46]/20 text-[#a1a1aa]';
                  const activeColor = sev === 'CRITICAL' ? 'bg-[#ef4444] text-white border-[#ef4444]' :
                                      sev === 'HIGH' ? 'bg-[#f97316] text-white border-[#f97316]' :
                                      sev === 'MEDIUM' ? 'bg-[#eab308] text-black border-[#eab308]' :
                                      'bg-[#a1a1aa] text-black border-[#a1a1aa]';
                  return (
                    <button
                      key={sev}
                      onClick={() => setWorkbenchSeverityFilter(isActive ? null : sev)}
                      className={`text-[9px] font-bold px-2 py-0.5 rounded border transition-all uppercase font-mono cursor-pointer ${
                        isActive ? activeColor : `${color} hover:bg-opacity-20`
                      }`}
                    >
                      {sev}
                    </button>
                  );
                })}
              </div>
            </div>

            <div className="flex-1 p-2 flex flex-col gap-2 overflow-y-auto bg-v2-surface">
              {findingsLoading ? (
                <div className="text-center py-12 text-xs text-v2-muted uppercase tracking-widest animate-pulse font-mono">
                  {t('pages.ccoverview.loadingFindings', 'Loading findings...')}
                </div>
              ) : workbenchFindings.length === 0 ? (
                <div className="text-center py-12 text-sm text-v2-muted font-mono">
                  {t('pages.ccoverview.noActiveFindings', 'No findings matching filters.')}
                </div>
              ) : (
                workbenchFindings.map((f: Finding, idx: number) => {
                  const isSelected = selectedFinding?.id === f.id;
                  const sevLabel = f.severity?.toUpperCase() || 'INFO';

                  return (
                    <div
                      key={f.id}
                      onClick={() => {
                        setSelectedFindingId(f.id);
                        setRemediationPlan('');
                      }}
                      style={{ animationDelay: `${180 + idx * 40}ms`, animationFillMode: 'both' }}
                      className={`p-4 border rounded-lg flex items-center justify-between cursor-pointer transition-all animate-page-transition ${
                        isSelected
                          ? 'bg-v2-surface-2 border-primary shadow-[0_0_0_1px_var(--accent-color-line)]'
                          : 'bg-v2-surface border-v2-border-soft hover:bg-v2-surface-2 hover:border-primary/30'
                      }`}
                    >
                      <div className="flex flex-col gap-1.5 flex-1 min-w-0 mr-3">
                        <div className="flex items-center gap-2">
                          <span
                            className={`text-[10px] font-bold px-2 py-0.5 rounded-full uppercase border font-mono ${
                              sevLabel === 'CRITICAL' || sevLabel === 'HIGH'
                                ? 'bg-v2-red-soft border-v2-red-line text-primary'
                                : 'bg-v2-bg border-v2-border-soft text-v2-muted'
                            }`}
                          >
                            {sevLabel}
                          </span>
                          <span className="font-mono text-[11px] truncate text-v2-fg-2">
                            {f.file_path
                              ? f.file_path.split('/').pop()
                              : f.file
                                ? f.file.split('/').pop()
                                : 'core'}
                          </span>
                        </div>
                        <span className="text-[13px] text-white font-medium truncate font-sans">
                          {f.title}
                        </span>
                      </div>
                    </div>
                  );
                })
              )}
            </div>
          </div>

          {/* RIGHT COLUMN: Vulnerability Details (Tabbed Remediation Lab) */}
          <div
            className="col-span-12 lg:col-span-4 v2-card flex flex-col p-0 overflow-hidden h-full animate-page-transition luxury-card-hover"
            style={{ animationDelay: '180ms', animationFillMode: 'both' }}
          >
            <div className="px-5 py-3 border-b border-v2-border-soft flex justify-between items-center bg-v2-surface shrink-0">
              <div className="flex items-center gap-4">
                <button
                  onClick={() => setRightTab('action_plan')}
                  className={`flex items-center gap-1.5 pb-0.5 border-b-2 transition-all ${
                    rightTab === 'action_plan'
                      ? 'border-primary text-white font-bold'
                      : 'border-transparent text-v2-muted hover:text-white'
                  }`}
                >
                  <span className="text-[11px] tracking-widest uppercase font-mono">{t('pages.ccoverview.aiActionPlan', 'AI ACTION PLAN')}</span>
                </button>
                <button
                  onClick={() => setRightTab('details')}
                  className={`flex items-center gap-1.5 pb-0.5 border-b-2 transition-all ${
                    rightTab === 'details'
                      ? 'border-primary text-white font-bold'
                      : 'border-transparent text-v2-muted hover:text-white'
                  }`}
                >
                  <span className="text-[11px] tracking-widest uppercase font-mono">{t('pages.ccoverview.details', 'DETAILS')}</span>
                </button>
                <button
                  onClick={() => setRightTab('code')}
                  className={`flex items-center gap-1.5 pb-0.5 border-b-2 transition-all ${
                    rightTab === 'code'
                      ? 'border-primary text-white font-bold'
                      : 'border-transparent text-v2-muted hover:text-white'
                  }`}
                >
                  <span className="text-[11px] tracking-widest uppercase font-mono">{t('pages.ccoverview.codeContext', 'CODE CONTEXT')}</span>
                </button>
              </div>
            </div>

            <div className="flex-1 flex flex-col bg-v2-surface overflow-hidden">
              {selectedFinding ? (
                <>
                  {rightTab === 'action_plan' && (
                    <div className="flex-1 flex flex-col p-6 overflow-y-auto h-full">
                      <div className="flex items-center justify-between mb-4 shrink-0">
                        <span className="v2-tag">
                          <span className="material-symbols-outlined text-[14px]">terminal</span>
                          {t('pages.ccoverview.actionRequired', 'Action Required')}
                        </span>
                        <button
                          onClick={async () => {
                            setRemediationLoading(true);
                            try {
                              const res = await securityService.analyzeFinding(
                                selectedFinding.id.toString(),
                              );
                              if (res.ok) setRemediationPlan(res.analysis);
                              else setRemediationPlan(res.error || t('pages.ccoverview.failedToGeneratePlan', 'Failed to generate plan.'));
                            } catch (err: any) {
                              setRemediationPlan(err.message || t('pages.ccoverview.errorOccurred', 'Error occurred.'));
                            } finally {
                              setRemediationLoading(false);
                            }
                          }}
                          disabled={remediationLoading}
                          className="v2-btn v2-btn-red px-4 py-2 font-mono"
                        >
                          {remediationLoading ? (
                            <>
                              <span className="material-symbols-outlined text-[16px] animate-spin">
                                sync
                              </span>
                              <span>{t('pages.ccoverview.analyzing', 'ANALYZING...')}</span>
                            </>
                          ) : (
                            <span>{t('pages.ccoverview.aiAnalyze', 'AI ANALYZE')}</span>
                          )}
                        </button>
                      </div>

                      <div className="flex-1 bg-v2-bg border border-v2-border-soft rounded-lg p-5 text-[13px] text-v2-fg-2 font-mono overflow-y-auto relative mb-4">
                        <div className="mb-12 whitespace-pre-wrap">
                          {remediationPlan ? (
                            <div>
                              <div className="text-[11px] text-primary font-bold uppercase tracking-widest mb-3 pb-2 border-b border-v2-border-soft">
                                {t('pages.ccoverview.aiRemediationPlan', 'AI Remediation Plan')}
                              </div>
                              {remediationPlan}
                            </div>
                          ) : (
                            activePrompt
                          )}
                        </div>

                        <div className="absolute bottom-4 right-4 flex gap-2">
                          <button
                            onClick={() => {
                              setContext(selectedFinding);
                              useCopilotStore
                                .getState()
                                .setPromptToSubmit(remediationPlan || activePrompt);
                              setIsOpen(true);
                            }}
                            className="v2-btn v2-btn-ghost px-3 py-1.5"
                          >
                            <span className="material-symbols-outlined text-[14px]">smart_toy</span>
                            {t('pages.ccoverview.copilot', 'Copilot')}
                          </button>
                          <button
                            onClick={() => {
                              navigator.clipboard.writeText(remediationPlan || activePrompt);
                              setCopied(true);
                              setTimeout(() => setCopied(false), 2000);
                            }}
                            className="v2-btn v2-btn-ghost px-3 py-1.5"
                          >
                            <span className="material-symbols-outlined text-[14px]">
                              {copied ? 'check' : 'content_copy'}
                            </span>
                            {copied ? t('pages.ccoverview.copied', 'Copied') : t('pages.ccoverview.copy', 'Copy')}
                          </button>
                        </div>
                      </div>
                    </div>
                  )}

                  {rightTab === 'details' && (
                    <div className="flex-1 p-6 flex flex-col space-y-6 overflow-y-auto h-full">
                      <div className="bg-v2-surface-2 p-5 rounded-lg border border-v2-border-soft">
                        <h4 className="text-xs font-bold text-white uppercase tracking-wider mb-2">{t('pages.ccoverview.description', 'Description')}</h4>
                        <p className="text-sm text-v2-fg-2 leading-relaxed whitespace-pre-wrap">{selectedFinding.description || t('pages.ccoverview.noDescription', 'No description available.')}</p>
                      </div>

                      {selectedFinding.impact && (
                        <div className="bg-v2-surface-2 p-5 rounded-lg border border-v2-border-soft">
                          <h4 className="text-xs font-bold text-white uppercase tracking-wider mb-2">{t('pages.ccoverview.impact', 'Impact')}</h4>
                          <p className="text-sm text-v2-fg-2 leading-relaxed whitespace-pre-wrap">{selectedFinding.impact}</p>
                        </div>
                      )}

                      <div className="grid grid-cols-2 gap-4">
                        <div className="bg-v2-surface-2 p-4 rounded-lg border border-v2-border-soft">
                          <div className="text-[10px] font-bold text-v2-muted uppercase tracking-wider mb-1">Rule ID</div>
                          <div className="font-mono text-[12px] text-white truncate" title={selectedFinding.rule_id}>{selectedFinding.rule_id}</div>
                        </div>
                        <div className="bg-v2-surface-2 p-4 rounded-lg border border-v2-border-soft">
                          <div className="text-[10px] font-bold text-v2-muted uppercase tracking-wider mb-1">CWE ID</div>
                          <div className="font-mono text-[12px] text-white">{selectedFinding.cwe_id || 'N/A'}</div>
                        </div>
                      </div>
                    </div>
                  )}

                  {rightTab === 'code' && (
                    <div className="flex-1 p-6 flex flex-col space-y-4 overflow-y-auto h-full">
                      <div className="bg-v2-surface-2 p-4 rounded-lg border border-v2-border-soft shrink-0">
                        <div className="text-[10px] font-bold text-v2-muted uppercase tracking-wider mb-1">{t('pages.ccoverview.file', 'File')}</div>
                        <div className="font-mono text-xs text-white break-all">{selectedFinding.file_path || selectedFinding.file || 'N/A'}</div>
                        <div className="text-[10px] font-bold text-v2-muted uppercase tracking-wider mt-3 mb-1">{t('pages.ccoverview.line', 'Line')}</div>
                        <div className="font-mono text-xs text-white">{selectedFinding.line_number || 0}</div>
                      </div>

                      {selectedFinding.code_snippet ? (
                        <div className="flex-1 flex flex-col overflow-hidden min-h-[200px]">
                          <div className="text-[10px] font-bold text-v2-muted uppercase tracking-wider mb-2 shrink-0">{t('pages.ccoverview.codeSnippet', 'Code Snippet')}</div>
                          <pre className="flex-1 bg-v2-bg border border-v2-border-soft rounded-lg p-4 font-mono text-xs text-[#a1a1aa] overflow-auto whitespace-pre">
                            <code>{selectedFinding.code_snippet}</code>
                          </pre>
                        </div>
                      ) : (
                        <div className="flex-1 border border-dashed border-v2-border-soft rounded-lg flex items-center justify-center p-8 text-center text-xs text-v2-muted font-mono">
                          {t('pages.ccoverview.noCodeSnippet', 'No code snippet available for this finding.')}
                        </div>
                      )}
                    </div>
                  )}
                </>
              ) : (
                <div className="flex-1 flex items-center justify-center text-sm text-v2-muted text-center p-8">
                  {t('pages.ccoverview.selectFinding', 'Select a finding in the Workbench to see remediation details.')}
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
