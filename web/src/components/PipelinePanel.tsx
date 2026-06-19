import React, { useState } from 'react';
import Markdown from 'react-markdown';
import { usePipeline } from '../hooks/usePipeline';

const PIPELINE_STEPS = [
  { step: 1, label: 'Threat Model', icon: 'security', desc: 'STRIDE analysis, TP/FP/NR classification' },
  { step: 2, label: 'PoC Verification', icon: 'bug_report', desc: 'Proves exploitability of True Positives' },
  { step: 3, label: 'Security Report', icon: 'description', desc: 'Full CS-XXX-NNN report with dispositions' },
  { step: 4, label: 'Fix Specification', icon: 'build', desc: 'AI IDE prompt for Cursor/Claude/Antigravity' },
];

interface PipelinePanelProps {
  productId?: number;
}

export const PipelinePanel: React.FC<PipelinePanelProps> = ({ productId = 1 }) => {
  const { status, progress, currentStep, steps, result, error, runPipeline, cancel } = usePipeline();
  const [activeTab, setActiveTab] = useState<'report' | 'fixspec'>('report');
  const [copied, setCopied] = useState(false);

  const handleRun = () => {
    runPipeline(productId);
  };

  const handleCopy = (text: string) => {
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const getStepStatus = (stepNum: number): 'pending' | 'active' | 'done' | 'warning' => {
    if (status === 'idle') return 'pending';
    const completed = steps.filter(s => s.step === stepNum && s.progress > (stepNum * 20 + 10));
    const hasWarning = steps.find(s => s.step === stepNum && s.warning);
    if (hasWarning) return 'warning';
    if (completed.length > 0 || (result && stepNum <= 4)) return 'done';
    if (currentStep?.step === stepNum) return 'active';
    return 'pending';
  };

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="px-6 py-4 border-b border-v2-border-soft bg-v2-surface shrink-0">
        <div className="flex items-center justify-between">
          <div>
            <div className="flex items-center gap-2 mb-1">
              <span className="w-1.5 h-3.5 bg-primary rounded-sm" />
              <span className="text-[10px] font-bold tracking-widest text-v2-fg uppercase font-mono">
                SecureCoder Pipeline
              </span>
            </div>
            <p className="text-[11px] text-v2-muted">
              Full pipeline: ThreatModel → PoC → Report → FixSpec (same as CI/CD)
            </p>
          </div>
          <div className="flex items-center gap-3">
            {status === 'running' ? (
              <button
                onClick={cancel}
                className="v2-tag cursor-pointer bg-v2-red-soft border-v2-red-line text-v2-red hover:bg-v2-red/20"
              >
                <span className="material-symbols-outlined text-[13px]">stop</span>
                Cancel
              </button>
            ) : (
              <button
                onClick={handleRun}
                className="v2-btn v2-btn-red px-4 py-2 text-[11px] font-bold tracking-widest uppercase disabled:opacity-50"
              >
                <span className="material-symbols-outlined text-[15px]">rocket_launch</span>
                {status === 'done' ? 'Re-Run Pipeline' : 'Run Pipeline'}
              </button>
            )}
          </div>
        </div>
      </div>

      {/* Pipeline Steps */}
      <div className="px-6 py-4 border-b border-v2-border-soft bg-v2-surface-2 shrink-0">
        <div className="flex items-center gap-2">
          {PIPELINE_STEPS.map((ps, i) => {
            const stepStatus = getStepStatus(ps.step);
            return (
              <React.Fragment key={ps.step}>
                <div className={`flex items-center gap-2 px-3 py-2 rounded-lg border transition-all ${
                  stepStatus === 'active' ? 'border-primary bg-primary/10 shadow-[0_0_12px_rgba(139,92,246,0.15)]' :
                  stepStatus === 'done' ? 'border-emerald-500/40 bg-emerald-500/10' :
                  stepStatus === 'warning' ? 'border-amber-500/40 bg-amber-500/10' :
                  'border-v2-border-soft bg-v2-surface'
                }`}>
                  <span className={`material-symbols-outlined text-[16px] ${
                    stepStatus === 'active' ? 'text-primary animate-pulse' :
                    stepStatus === 'done' ? 'text-emerald-400' :
                    stepStatus === 'warning' ? 'text-amber-400' :
                    'text-v2-muted'
                  }`}>
                    {stepStatus === 'done' ? 'check_circle' :
                     stepStatus === 'warning' ? 'warning' :
                     stepStatus === 'active' ? 'pending' :
                     ps.icon}
                  </span>
                  <div>
                    <div className={`text-[10px] font-bold tracking-wider ${
                      stepStatus === 'active' ? 'text-primary' :
                      stepStatus === 'done' ? 'text-emerald-400' :
                      stepStatus === 'warning' ? 'text-amber-400' :
                      'text-v2-muted'
                    }`}>{ps.label}</div>
                    <div className="text-[9px] text-v2-muted truncate max-w-[120px]">{ps.desc}</div>
                  </div>
                </div>
                {i < PIPELINE_STEPS.length - 1 && (
                  <span className={`text-[10px] ${stepStatus === 'done' ? 'text-emerald-400' : 'text-v2-muted'}`}>→</span>
                )}
              </React.Fragment>
            );
          })}
        </div>

        {/* Progress bar */}
        {status === 'running' && (
          <div className="mt-3">
            <div className="flex items-center justify-between mb-1">
              <span className="text-[10px] text-v2-muted font-mono">
                {currentStep?.label || 'Initializing...'}
              </span>
              <span className="text-[10px] text-primary font-mono font-bold">{progress}%</span>
            </div>
            <div className="h-1 bg-v2-surface rounded-full overflow-hidden">
              <div
                className="h-full bg-gradient-to-r from-primary to-primary/60 rounded-full transition-all duration-700 ease-out"
                style={{ width: `${progress}%` }}
              />
            </div>
          </div>
        )}

        {/* Stats after completion */}
        {result && (
          <div className="mt-3 flex items-center gap-4">
            <div className="flex items-center gap-1.5">
              <span className="w-2 h-2 rounded-full bg-error" />
              <span className="text-[10px] font-mono text-v2-fg">{result.stats.tp} True Positive</span>
            </div>
            <div className="flex items-center gap-1.5">
              <span className="w-2 h-2 rounded-full bg-emerald-400" />
              <span className="text-[10px] font-mono text-v2-fg">{result.stats.fp} False Positive</span>
            </div>
            <div className="flex items-center gap-1.5">
              <span className="w-2 h-2 rounded-full bg-amber-400" />
              <span className="text-[10px] font-mono text-v2-fg">{result.stats.nr} Needs Review</span>
            </div>
            <div className="text-[9px] text-v2-muted font-mono ml-auto">
              {result.usage.total_tokens.toLocaleString()} tokens
              (≈${((result.usage.prompt_tokens * 0.00000015) + (result.usage.completion_tokens * 0.0000006)).toFixed(4)})
            </div>
          </div>
        )}
      </div>

      {/* Error */}
      {error && (
        <div className="px-6 py-3 bg-error/10 border-b border-error/30">
          <div className="flex items-center gap-2 text-error text-[12px]">
            <span className="material-symbols-outlined text-[16px]">error</span>
            {error}
          </div>
        </div>
      )}

      {/* Results */}
      {result && (
        <div className="flex-1 flex flex-col overflow-hidden">
          {/* Tabs */}
          <div className="flex border-b border-v2-border-soft shrink-0">
            <button
              onClick={() => setActiveTab('report')}
              className={`px-5 py-3 text-[11px] font-bold tracking-widest uppercase transition-colors ${
                activeTab === 'report'
                  ? 'text-primary border-b-2 border-primary bg-v2-surface'
                  : 'text-v2-muted hover:text-v2-fg'
              }`}
            >
              <span className="material-symbols-outlined text-[14px] mr-1.5 align-middle">description</span>
              Security Report
            </button>
            <button
              onClick={() => setActiveTab('fixspec')}
              className={`px-5 py-3 text-[11px] font-bold tracking-widest uppercase transition-colors ${
                activeTab === 'fixspec'
                  ? 'text-primary border-b-2 border-primary bg-v2-surface'
                  : 'text-v2-muted hover:text-v2-fg'
              }`}
            >
              <span className="material-symbols-outlined text-[14px] mr-1.5 align-middle">build</span>
              AI Fix Spec
            </button>
            <div className="ml-auto flex items-center px-3 gap-2">
              <button
                onClick={() => handleCopy(activeTab === 'report' ? result.report : result.fix_spec)}
                className="v2-tag cursor-pointer hover:bg-v2-surface-2"
              >
                <span className="material-symbols-outlined text-[13px]">
                  {copied ? 'check' : 'content_copy'}
                </span>
                {copied ? 'Copied!' : 'Copy'}
              </button>
            </div>
          </div>

          {/* Content */}
          <div className="flex-1 overflow-y-auto cyber-scrollbar p-6">
            <div className="prose prose-invert prose-sm max-w-none
              [&_p]:text-[13px] [&_p]:leading-relaxed [&_p]:text-v2-fg-2
              [&_li]:text-[13px] [&_li]:text-v2-fg-2
              [&_code]:text-primary [&_code]:bg-primary/10 [&_code]:px-1
              [&_strong]:text-white
              [&_h1]:text-lg [&_h1]:text-white [&_h1]:border-b [&_h1]:border-v2-border-soft [&_h1]:pb-2
              [&_h2]:text-base [&_h2]:text-white
              [&_h3]:text-sm [&_h3]:text-v2-fg
              [&_table]:text-[12px] [&_th]:bg-v2-surface-2 [&_th]:px-3 [&_th]:py-1.5
              [&_td]:px-3 [&_td]:py-1.5 [&_td]:border-v2-border-soft
              [&_blockquote]:border-l-primary [&_blockquote]:bg-primary/5">
              <Markdown>
                {activeTab === 'report' ? result.report : result.fix_spec}
              </Markdown>
            </div>
          </div>
        </div>
      )}

      {/* Empty state */}
      {status === 'idle' && !result && (
        <div className="flex-1 flex items-center justify-center">
          <div className="text-center max-w-md">
            <span className="material-symbols-outlined text-6xl text-v2-muted/30 block mb-4">rocket_launch</span>
            <h3 className="text-sm font-bold text-v2-fg tracking-wider uppercase mb-2">SecureCoder Pipeline</h3>
            <p className="text-[12px] text-v2-muted leading-relaxed mb-6">
              Run the full 4-step security pipeline, identical to CI/CD.
              Analyzes all findings with ThreatModel → PoC → Report → FixSpec.
            </p>
            <button
              onClick={handleRun}
              className="v2-btn v2-btn-red px-6 py-2.5 text-[11px] font-bold tracking-widest uppercase"
            >
              <span className="material-symbols-outlined text-[15px]">rocket_launch</span>
              Launch Pipeline
            </button>
          </div>
        </div>
      )}
    </div>
  );
};
