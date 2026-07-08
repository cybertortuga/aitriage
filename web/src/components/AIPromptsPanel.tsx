import React, { useState, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { motion, AnimatePresence } from 'framer-motion';

interface PromptItem {
  id: string;
  titleKey: string;
  descKey: string;
  icon: string;
  promptKey: string;
  tags: string[];
}

const PROMPT_CATEGORIES = [
  {
    id: 'security-scan',
    labelKey: 'aiPrompts.categories.securityScan',
    icon: 'shield',
    prompts: [
      {
        id: 'scan-project',
        titleKey: 'aiPrompts.prompts.scanProject.title',
        descKey: 'aiPrompts.prompts.scanProject.desc',
        icon: 'radar',
        promptKey: 'aiPrompts.prompts.scanProject.prompt',
        tags: ['Cursor', 'Copilot', 'Windsurf'],
      },
      {
        id: 'scan-file',
        titleKey: 'aiPrompts.prompts.scanFile.title',
        descKey: 'aiPrompts.prompts.scanFile.desc',
        icon: 'description',
        promptKey: 'aiPrompts.prompts.scanFile.prompt',
        tags: ['Cursor', 'Copilot'],
      },
      {
        id: 'scan-dependencies',
        titleKey: 'aiPrompts.prompts.scanDeps.title',
        descKey: 'aiPrompts.prompts.scanDeps.desc',
        icon: 'account_tree',
        promptKey: 'aiPrompts.prompts.scanDeps.prompt',
        tags: ['Cursor', 'Copilot', 'Windsurf'],
      },
    ],
  },
  {
    id: 'remediation',
    labelKey: 'aiPrompts.categories.remediation',
    icon: 'auto_fix_high',
    prompts: [
      {
        id: 'fix-vuln',
        titleKey: 'aiPrompts.prompts.fixVuln.title',
        descKey: 'aiPrompts.prompts.fixVuln.desc',
        icon: 'healing',
        promptKey: 'aiPrompts.prompts.fixVuln.prompt',
        tags: ['Cursor', 'Copilot'],
      },
      {
        id: 'fix-batch',
        titleKey: 'aiPrompts.prompts.fixBatch.title',
        descKey: 'aiPrompts.prompts.fixBatch.desc',
        icon: 'dynamic_feed',
        promptKey: 'aiPrompts.prompts.fixBatch.prompt',
        tags: ['Cursor', 'Windsurf'],
      },
      {
        id: 'secure-refactor',
        titleKey: 'aiPrompts.prompts.secureRefactor.title',
        descKey: 'aiPrompts.prompts.secureRefactor.desc',
        icon: 'build',
        promptKey: 'aiPrompts.prompts.secureRefactor.prompt',
        tags: ['Cursor', 'Copilot', 'Windsurf'],
      },
    ],
  },
  {
    id: 'analysis',
    labelKey: 'aiPrompts.categories.analysis',
    icon: 'analytics',
    prompts: [
      {
        id: 'threat-model',
        titleKey: 'aiPrompts.prompts.threatModel.title',
        descKey: 'aiPrompts.prompts.threatModel.desc',
        icon: 'security',
        promptKey: 'aiPrompts.prompts.threatModel.prompt',
        tags: ['Cursor', 'Copilot'],
      },
      {
        id: 'code-review',
        titleKey: 'aiPrompts.prompts.codeReview.title',
        descKey: 'aiPrompts.prompts.codeReview.desc',
        icon: 'rate_review',
        promptKey: 'aiPrompts.prompts.codeReview.prompt',
        tags: ['Cursor', 'Copilot', 'Windsurf'],
      },
      {
        id: 'attack-surface',
        titleKey: 'aiPrompts.prompts.attackSurface.title',
        descKey: 'aiPrompts.prompts.attackSurface.desc',
        icon: 'bug_report',
        promptKey: 'aiPrompts.prompts.attackSurface.prompt',
        tags: ['Cursor'],
      },
    ],
  },
  {
    id: 'compliance',
    labelKey: 'aiPrompts.categories.compliance',
    icon: 'verified_user',
    prompts: [
      {
        id: 'owasp-check',
        titleKey: 'aiPrompts.prompts.owaspCheck.title',
        descKey: 'aiPrompts.prompts.owaspCheck.desc',
        icon: 'checklist',
        promptKey: 'aiPrompts.prompts.owaspCheck.prompt',
        tags: ['Cursor', 'Copilot'],
      },
      {
        id: 'write-tests',
        titleKey: 'aiPrompts.prompts.writeTests.title',
        descKey: 'aiPrompts.prompts.writeTests.desc',
        icon: 'science',
        promptKey: 'aiPrompts.prompts.writeTests.prompt',
        tags: ['Cursor', 'Copilot', 'Windsurf'],
      },
    ],
  },
];

interface AIPromptsPanelProps {
  isOpen: boolean;
  onToggle: () => void;
}

export const AIPromptsPanel: React.FC<AIPromptsPanelProps> = ({ isOpen, onToggle }) => {
  const { t } = useTranslation('components');
  const [expandedCategory, setExpandedCategory] = useState<string | null>('security-scan');
  const [copiedId, setCopiedId] = useState<string | null>(null);
  const [expandedPrompt, setExpandedPrompt] = useState<string | null>(null);

  const handleCopy = useCallback((promptKey: string, id: string) => {
    const text = t(promptKey);
    navigator.clipboard.writeText(text).then(() => {
      setCopiedId(id);
      setTimeout(() => setCopiedId(null), 2000);
    });
  }, [t]);

  const toggleCategory = useCallback((id: string) => {
    setExpandedCategory((prev) => (prev === id ? null : id));
  }, []);

  const togglePrompt = useCallback((id: string) => {
    setExpandedPrompt((prev) => (prev === id ? null : id));
  }, []);

  return (
    <div className="relative flex shrink-0 h-full">
      {/* Toggle Tab (always visible, anchored to left edge of panel area) */}
      <button
        onClick={onToggle}
        className={`absolute -left-6 top-1/2 -translate-y-1/2 z-50 flex items-center justify-center w-6 h-20 rounded-l-lg transition-all duration-300 ${
          isOpen
            ? 'bg-v2-surface-2 border border-r-0 border-v2-border-soft text-v2-red hover:bg-v2-surface'
            : 'bg-v2-surface border border-r-0 border-v2-border-soft text-v2-muted hover:text-v2-red hover:bg-v2-surface-2'
        }`}
        title={t('aiPrompts.togglePanel')}
      >
        <span
          className="material-symbols-outlined text-[14px] transition-transform duration-300"
          style={{ transform: isOpen ? 'rotate(0deg)' : 'rotate(180deg)' }}
        >
          chevron_right
        </span>
      </button>

      {/* Panel */}
      <AnimatePresence>
        {isOpen && (
          <motion.aside
            initial={{ width: 0, opacity: 0 }}
            animate={{ width: 340, opacity: 1 }}
            exit={{ width: 0, opacity: 0 }}
            transition={{ duration: 0.3, ease: [0.4, 0, 0.2, 1] }}
            className="shrink-0 border-l border-v2-border-soft bg-v2-surface flex flex-col h-full overflow-hidden"
          >
            {/* Header */}
            <div className="px-4 py-4 border-b border-v2-border-soft bg-v2-surface-2 shrink-0">
              <div className="flex items-center gap-2.5 mb-1">
                <div className="w-7 h-7 rounded-lg bg-v2-red-soft border border-v2-red-line flex items-center justify-center">
                  <span className="material-symbols-outlined text-v2-red text-[15px]" style={{ fontVariationSettings: "'FILL' 1" }}>
                    terminal
                  </span>
                </div>
                <div>
                  <h3 className="text-[11px] font-bold text-white tracking-wider uppercase">
                    {t('aiPrompts.title')}
                  </h3>
                  <p className="text-[9px] text-v2-muted tracking-widest">
                    {t('aiPrompts.subtitle')}
                  </p>
                </div>
              </div>
            </div>

            {/* IDE Badges */}
            <div className="px-4 py-2.5 border-b border-v2-border-soft bg-v2-bg shrink-0 flex items-center gap-1.5 overflow-x-auto">
              <span className="text-[9px] text-v2-muted tracking-widest font-bold shrink-0 mr-1">
                {t('aiPrompts.compatibleWith')}
              </span>
              {['Cursor', 'GitHub Copilot', 'Windsurf', 'Cline'].map((ide) => (
                <span
                  key={ide}
                  className="shrink-0 text-[9px] px-2 py-0.5 rounded border border-v2-border-soft bg-v2-surface-2 text-v2-fg-2 font-mono"
                >
                  {ide}
                </span>
              ))}
            </div>

            {/* Prompt Categories */}
            <div className="flex-1 overflow-y-auto cyber-scrollbar">
              {PROMPT_CATEGORIES.map((category) => {
                const isExpanded = expandedCategory === category.id;
                return (
                  <div key={category.id} className="border-b border-v2-border-soft">
                    {/* Category Header */}
                    <button
                      onClick={() => toggleCategory(category.id)}
                      className="w-full flex items-center gap-3 px-4 py-3 hover:bg-v2-surface-2 transition-colors group"
                    >
                      <span className="material-symbols-outlined text-[16px] text-v2-muted group-hover:text-v2-red transition-colors">
                        {category.icon}
                      </span>
                      <span className="text-[10px] font-bold tracking-wider text-v2-fg-2 uppercase flex-1 text-left">
                        {t(category.labelKey)}
                      </span>
                      <span className="text-[9px] text-v2-muted mr-1">{category.prompts.length}</span>
                      <span
                        className={`material-symbols-outlined text-[14px] text-v2-muted transition-transform duration-200 ${
                          isExpanded ? 'rotate-180' : ''
                        }`}
                      >
                        expand_more
                      </span>
                    </button>

                    {/* Prompts inside category */}
                    <AnimatePresence>
                      {isExpanded && (
                        <motion.div
                          initial={{ height: 0, opacity: 0 }}
                          animate={{ height: 'auto', opacity: 1 }}
                          exit={{ height: 0, opacity: 0 }}
                          transition={{ duration: 0.2, ease: 'easeInOut' }}
                          className="overflow-hidden"
                        >
                          {category.prompts.map((prompt: PromptItem) => {
                            const isCopied = copiedId === prompt.id;
                            const isPromptExpanded = expandedPrompt === prompt.id;
                            return (
                              <div
                                key={prompt.id}
                                className="mx-3 mb-2 rounded-lg border border-v2-border-soft bg-v2-bg hover:border-v2-border transition-all"
                              >
                                {/* Prompt card header */}
                                <div className="flex items-start gap-2.5 px-3 py-2.5">
                                  <span
                                    className="material-symbols-outlined text-[14px] text-v2-muted mt-0.5 shrink-0"
                                    style={{ fontVariationSettings: "'FILL' 0" }}
                                  >
                                    {prompt.icon}
                                  </span>
                                  <div className="flex-1 min-w-0">
                                    <div className="text-[11px] font-semibold text-white leading-tight mb-0.5">
                                      {t(prompt.titleKey)}
                                    </div>
                                    <div className="text-[9px] text-v2-muted leading-snug">
                                      {t(prompt.descKey)}
                                    </div>
                                    {/* Tags */}
                                    <div className="flex gap-1 mt-1.5 flex-wrap">
                                      {prompt.tags.map((tag) => (
                                        <span
                                          key={tag}
                                          className="text-[8px] px-1.5 py-0.5 rounded bg-v2-surface-2 border border-v2-border-soft text-v2-muted font-mono"
                                        >
                                          {tag}
                                        </span>
                                      ))}
                                    </div>
                                  </div>
                                </div>

                                {/* Action buttons */}
                                <div className="flex items-center gap-1.5 px-3 pb-2.5">
                                  <button
                                    onClick={() => handleCopy(prompt.promptKey, prompt.id)}
                                    className={`flex items-center gap-1.5 px-2.5 py-1 rounded text-[9px] font-bold tracking-wider transition-all duration-200 ${
                                      isCopied
                                        ? 'bg-emerald-500/10 border border-emerald-500/30 text-emerald-400'
                                        : 'bg-v2-red-soft border border-v2-red-line text-v2-red hover:bg-v2-red/20'
                                    }`}
                                  >
                                    <span className="material-symbols-outlined text-[12px]">
                                      {isCopied ? 'check_circle' : 'content_copy'}
                                    </span>
                                    {isCopied ? t('aiPrompts.copied') : t('aiPrompts.copyPrompt')}
                                  </button>
                                  <button
                                    onClick={() => togglePrompt(prompt.id)}
                                    className="flex items-center gap-1 px-2 py-1 rounded text-[9px] font-bold tracking-wider bg-v2-surface-2 border border-v2-border-soft text-v2-muted hover:text-white hover:bg-v2-surface transition-all"
                                  >
                                    <span className="material-symbols-outlined text-[12px]">
                                      {isPromptExpanded ? 'visibility_off' : 'visibility'}
                                    </span>
                                    {isPromptExpanded ? t('aiPrompts.hide') : t('aiPrompts.preview')}
                                  </button>
                                </div>

                                {/* Expandable Preview */}
                                <AnimatePresence>
                                  {isPromptExpanded && (
                                    <motion.div
                                      initial={{ height: 0, opacity: 0 }}
                                      animate={{ height: 'auto', opacity: 1 }}
                                      exit={{ height: 0, opacity: 0 }}
                                      transition={{ duration: 0.15 }}
                                      className="overflow-hidden"
                                    >
                                      <div className="mx-3 mb-3 p-3 rounded-lg bg-v2-surface-2 border border-v2-border-soft">
                                        <pre className="text-[10px] text-v2-fg-2 font-mono leading-relaxed whitespace-pre-wrap max-h-40 overflow-y-auto cyber-scrollbar">
                                          {t(prompt.promptKey)}
                                        </pre>
                                      </div>
                                    </motion.div>
                                  )}
                                </AnimatePresence>
                              </div>
                            );
                          })}
                          <div className="h-2" />
                        </motion.div>
                      )}
                    </AnimatePresence>
                  </div>
                );
              })}
            </div>

            {/* Footer */}
            <div className="px-4 py-3 border-t border-v2-border-soft bg-v2-surface-2 shrink-0">
              <div className="flex items-center gap-2 text-[9px] text-v2-muted">
                <span className="material-symbols-outlined text-[12px]">info</span>
                <span className="leading-tight">
                  {t('aiPrompts.footerHint')}
                </span>
              </div>
            </div>
          </motion.aside>
        )}
      </AnimatePresence>
    </div>
  );
};

