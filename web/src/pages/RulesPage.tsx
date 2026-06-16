import React, { useState, useEffect } from 'react';
import { useTitle } from '../hooks/useTitle';
import { useTranslation } from 'react-i18next';
import { useCopilotStore } from '../store/CopilotStore';

const SEV_COLORS: Record<string, string> = {
  CRITICAL: 'var(--color-severity-critical)',
  HIGH: 'var(--color-severity-high)',
  MEDIUM: 'var(--color-severity-medium)',
  LOW: 'var(--color-severity-low)',
};

export const RulesPage: React.FC = () => {
  const { t } = useTranslation('pages');
  useTitle(t('rules.title'));
  const [rules, setRules] = useState<any[]>([]);
  const [selectedRule, setSelectedRule] = useState<any | null>(null);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [selectedStack, setSelectedStack] = useState('ALL_STACKS');
  const [selectedSeverity, setSelectedSeverity] = useState('ALL_SEVERITIES');

  useEffect(() => {
    fetch('/api/rules')
      .then((res) => res.json())
      .then((data) => {
        if (data.ok) {
          setRules(data.rules || []);
          if (data.rules?.length > 0) setSelectedRule(data.rules[0]);
        }
      })
      .catch((err) => {
        console.error('Failed to fetch rules:', err);
      })
      .finally(() => setLoading(false));
  }, []);

  const filtered = rules.filter((r) => {
    const matchesSearch =
      !search ||
      r.name?.toLowerCase().includes(search.toLowerCase()) ||
      String(r.id).toLowerCase().includes(search.toLowerCase());
    const matchesStack = selectedStack === 'ALL_STACKS' || r.stack === selectedStack;
    const matchesSeverity =
      selectedSeverity === 'ALL_SEVERITIES' || r.severity?.toUpperCase() === selectedSeverity;
    return matchesSearch && matchesStack && matchesSeverity;
  });

  const stacks = Array.from(new Set(rules.map((r) => r.stack)))
    .filter(Boolean)
    .sort() as string[];
  const severities = ['CRITICAL', 'HIGH', 'MEDIUM', 'LOW'];

  return (
    <div className="flex flex-col h-full overflow-hidden">
      {/* Page Header */}
      <div className="px-4 py-2 flex flex-col md:flex-row justify-between items-start md:items-center gap-4 flex-shrink-0 cyber-header-premium border-b border-outline-variant/30">
        <div>
          <p className="text-[9px] font-bold tracking-widest text-on-surface-variant mb-0.5 opacity-70">
            {t('rules.breadcrumb', { count: rules.length })}
          </p>
          <h1 className="text-title-lg font-bold tracking-tight text-primary uppercase">
            {t('rules.header')}
            {filtered.length !== rules.length && (
              <span className="text-[9px] font-bold tracking-widest text-on-surface-variant ml-4 opacity-50 lowercase">
                {t('rules.visibleCount', { count: filtered.length })}
              </span>
            )}
          </h1>
          <p className="text-[9px] text-on-surface-variant mt-1 max-w-2xl">
            <strong>{t('rules.noteLabel')}</strong> {t('rules.noteContent')}
          </p>
        </div>
        <div className="flex flex-wrap items-center gap-3">
          {/* Stack Filter */}
          <select
            className="bg-surface-container-lowest border border-outline-variant text-label-caps text-on-surface-variant h-8 px-2 focus:border-primary outline-none cursor-pointer"
            value={selectedStack}
            onChange={(e) => setSelectedStack(e.target.value)}
          >
            <option value="ALL_STACKS">{t('rules.allStacks')}</option>
            {stacks.map((s) => (
              <option key={s} value={s}>
                {s.toUpperCase()}
              </option>
            ))}
          </select>

          {/* Severity Filter */}
          <select
            className="bg-surface-container-lowest border border-outline-variant text-label-caps text-on-surface-variant h-8 px-2 focus:border-primary outline-none cursor-pointer"
            value={selectedSeverity}
            onChange={(e) => setSelectedSeverity(e.target.value)}
          >
            <option value="ALL_SEVERITIES">{t('rules.allSeverities')}</option>
            {severities.map((s) => (
              <option key={s} value={s}>
                {t('rules.severity.' + s.toLowerCase())}
              </option>
            ))}
          </select>

          {/* Search Bar */}
          <div className="flex items-center border border-outline-variant bg-surface-container-lowest h-8 px-3 gap-2 w-56 focus-within:border-primary transition-none">
            <span
              className="material-symbols-outlined text-on-surface-variant"
              style={{ fontSize: '14px' }}
            >
              search
            </span>
            <input
              className="bg-transparent border-none focus:outline-none focus:ring-0 text-mono-data text-primary placeholder:text-on-surface-variant/40 w-full"
              placeholder={t('rules.searchPlaceholder')}
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
          </div>

          <button className="btn-primary h-8 px-4 flex items-center gap-2 text-label-caps">
            <span className="material-symbols-outlined" style={{ fontSize: '14px' }}>
              add
            </span>
            {t('rules.newRule')}
          </button>
        </div>
      </div>

      {/* Split Layout */}
      <div className="flex-1 flex overflow-hidden">
        {/* Master List */}
        <div className="w-1/2 flex flex-col border-r border-outline-variant overflow-hidden">
          <div className="cyber-grid-header flex items-center text-label-caps text-on-surface-variant shrink-0">
            <div className="w-10 py-3 px-3 text-center shrink-0">{t('rules.table.sts')}</div>
            <div className="w-24 py-3 px-3 shrink-0">{t('rules.table.id')}</div>
            <div className="flex-1 py-3 px-3 min-w-0">{t('rules.table.ruleName')}</div>
            <div className="w-24 py-3 px-3 shrink-0">{t('rules.table.severity')}</div>
            <div className="w-24 py-3 px-3 shrink-0">{t('rules.table.stack')}</div>
          </div>
          <div className="flex-1 overflow-y-auto cyber-scrollbar">
            {loading && (
              <div className="p-6 text-label-caps text-on-surface-variant animate-pulse">
                {t('rules.loading')}
              </div>
            )}
            {!loading && filtered.length === 0 && (
              <div className="p-12 flex flex-col items-center justify-center text-center opacity-50">
                <span className="material-symbols-outlined text-4xl mb-4">search_off</span>
                <p className="text-label-caps">{t('rules.noRules')}</p>
                <button
                  onClick={() => {
                    setSearch('');
                    setSelectedStack('ALL_STACKS');
                    setSelectedSeverity('ALL_SEVERITIES');
                  }}
                  className="mt-4 text-primary text-label-caps hover:underline"
                >
                  {t('rules.clearFilters')}
                </button>
              </div>
            )}
            {!loading &&
              filtered.map((rule) => {
                const isSelected = selectedRule?.id === rule.id;
                const sevColor =
                  SEV_COLORS[rule.severity?.toUpperCase()] || 'var(--color-on-surface-variant)';
                return (
                  <div
                    key={rule.id}
                    onClick={() => setSelectedRule(rule)}
                    className={`cyber-grid-row flex items-center cursor-pointer ${
                      isSelected
                        ? 'bg-surface-container-highest text-primary border-l-2 border-primary'
                        : 'border-l-2 border-transparent'
                    }`}
                  >
                    <div className="w-10 py-3 px-3 flex justify-center shrink-0">
                      <div className="status-dot bg-primary" />
                    </div>
                    <div className="w-24 py-3 px-3 text-mono-data text-on-surface-variant shrink-0 truncate">
                      {rule.id}
                    </div>
                    <div className="flex-1 py-3 px-3 text-mono-data text-on-surface truncate min-w-0">
                      {rule.name}
                    </div>
                    <div
                      className="w-24 py-3 px-3 text-label-caps shrink-0"
                      style={{ color: sevColor }}
                    >
                      {rule.severity}
                    </div>
                    <div className="w-24 py-3 px-3 text-label-caps text-on-surface-variant shrink-0 truncate">
                      {rule.stack}
                    </div>
                  </div>
                );
              })}
          </div>
        </div>

        {/* Detail Panel */}
        <div className="w-1/2 flex flex-col overflow-y-auto cyber-scrollbar bg-surface-container-low">
          {selectedRule ? (
            <div className="p-6 flex flex-col gap-6">
              <div className="flex justify-between items-start gap-4 border-b border-outline-variant pb-5">
                <div className="flex-1 min-w-0">
                  <div className="text-label-caps text-on-surface-variant mb-1">
                    {t('rules.detail.ruleId', { id: selectedRule.id })}
                  </div>
                  <h2 className="text-headline-sm text-primary">{selectedRule.name}</h2>
                </div>
                <div className="flex flex-col items-end gap-2 shrink-0">
                  <span
                    className="text-label-caps px-3 py-1 border border-outline-variant"
                    style={{ color: SEV_COLORS[selectedRule.severity?.toUpperCase()] }}
                  >
                    {selectedRule.severity}
                  </span>
                  <button
                    onClick={() => {
                      const { setContext, setIsOpen } = useCopilotStore.getState();
                      setContext(selectedRule);
                      setIsOpen(true);
                    }}
                    className="flex items-center gap-1.5 text-label-caps text-primary hover:underline"
                  >
                    <span className="material-symbols-outlined" style={{ fontSize: '14px' }}>
                      smart_toy
                    </span>
                    {t('rules.detail.askAiCopilot')}
                  </button>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div className="cyber-widget p-3">
                  <span className="text-label-caps text-on-surface-variant block mb-2">{t('rules.detail.target')}</span>
                  <span className="text-mono-data text-on-surface">
                    {selectedRule.target ?? 'code'}
                  </span>
                </div>
                <div className="cyber-widget p-3">
                  <span className="text-label-caps text-on-surface-variant block mb-2">{t('rules.detail.framework')}</span>
                  <span className="text-mono-data text-on-surface">{selectedRule.stack}</span>
                </div>
              </div>

              <div>
                <div className="text-label-caps text-on-surface-variant mb-3 flex items-center gap-2">
                  <span className="material-symbols-outlined text-[14px]">code</span>
                  {t('rules.detail.logicDefinition')}
                </div>
                <div className="cyber-widget p-4 text-body-sm text-on-surface leading-relaxed">
                  {selectedRule.condition ? t('rules.detail.conditionPrefix', { condition: selectedRule.condition }) : t('rules.detail.defaultLogic')}
                  {selectedRule.extensions && (
                    <div className="mt-2 text-on-surface-variant text-mono-data">
                      {t('rules.detail.extensionsPrefix', { extensions: selectedRule.extensions.join(', ') })}
                    </div>
                  )}
                </div>
              </div>

              {selectedRule.pattern && (
                <div>
                  <div className="text-label-caps text-on-surface-variant mb-3 flex items-center gap-2">
                    <span className="material-symbols-outlined text-[14px]">terminal</span>
                    {t('rules.detail.patternRegex')}
                  </div>
                  <pre className="cyber-widget p-4 text-mono-data text-secondary overflow-x-auto whitespace-pre">
                    {selectedRule.pattern}
                  </pre>
                </div>
              )}

              <div>
                <div className="text-label-caps text-on-surface-variant mb-3 flex items-center gap-2">
                  <span className="material-symbols-outlined text-[14px]">build</span>
                  {t('rules.detail.remediationGuidance')}
                </div>
                <div className="border-l-2 border-primary pl-4 text-body-sm text-on-surface leading-relaxed">
                  {selectedRule.suggestion ?? t('rules.detail.noRemediation')}
                </div>
              </div>

              <div className="border-t border-outline-variant pt-5 flex gap-3">
                <button className="btn-secondary flex-1 py-2 text-label-caps">{t('rules.detail.editRule')}</button>
                <button className="btn-mechanical-error flex-1 py-2 text-label-caps">
                  {t('rules.detail.disable')}
                </button>
              </div>
            </div>
          ) : (
            <div className="flex-1 flex items-center justify-center p-8">
              <span className="text-label-caps text-on-surface-variant opacity-40">
                {t('rules.detail.selectRule')}
              </span>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
