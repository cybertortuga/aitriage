import React, { useState } from 'react';
import { useProducts } from '../hooks/useProducts';
import { useNavigate } from 'react-router-dom';
import { useTitle } from '../hooks/useTitle';
import { useTranslation } from 'react-i18next';

export const ProductsPage: React.FC = () => {
  const { t } = useTranslation('pages');
  const { products, loading, error } = useProducts();
  const navigate = useNavigate();
  useTitle(t('assets.title'));
  const [search, setSearch] = useState('');

  const criticalTotal = products.reduce((acc, p) => acc + p.sla_critical, 0);
  const highTotal = products.reduce((acc, p) => acc + p.sla_high, 0);
  const filtered = products.filter(
    (p) => !search || p.name?.toLowerCase().includes(search.toLowerCase()),
  );

  const kpis = [
    { label: t('assets.totalAssets'), value: products.length },
    { label: t('assets.criticalVulns'), value: criticalTotal, highlight: true },
    { label: t('assets.highVulns'), value: highTotal },
  ];

  return (
    <div className="flex flex-col min-h-full">
      {/* Page Header */}
      <div className="px-4 py-2 border-b border-outline-variant flex justify-between items-center flex-shrink-0">
        <div>
          <p className="text-[9px] font-bold tracking-widest text-on-surface-variant mb-0.5">
            {t('assets.breadcrumb')}
          </p>
          <h1 className="text-title-lg font-bold tracking-tight text-primary uppercase">
            {t('assets.header')}
          </h1>
        </div>
        <button className="btn-primary h-8 px-4 flex items-center gap-2">
          <span className="material-symbols-outlined" style={{ fontSize: '16px' }}>
            add
          </span>
          <span>{t('assets.createAsset')}</span>
        </button>
      </div>

      {/* KPI Bar */}
      <div className="grid grid-cols-3 border-b border-outline-variant shrink-0">
        {kpis.map((kpi, i) => (
          <div key={i} className="px-4 py-2 border-r border-outline-variant last:border-r-0">
            <div className="text-label-caps font-label-caps text-on-surface-variant mb-2">
              {kpi.label}
            </div>
            <div
              className={`text-mono-metrics font-mono-metrics ${kpi.highlight ? 'text-severity-critical' : 'text-primary'}`}
            >
              {kpi.value}
            </div>
          </div>
        ))}
      </div>

      {/* Filters */}
      <div className="px-4 py-2 border-b border-outline-variant flex items-center gap-3 shrink-0">
        <div className="flex items-center border border-outline-variant bg-surface-container-lowest h-8 px-3 gap-2 w-72">
          <span
            className="material-symbols-outlined text-on-surface-variant"
            style={{ fontSize: '14px' }}
          >
            search
          </span>
          <input
            className="bg-transparent border-none focus:outline-none focus:ring-0 text-mono-data font-mono-data text-primary placeholder:text-on-surface-variant/50 w-full"
            placeholder={t('assets.searchPlaceholder')}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
        <button className="btn-secondary h-8 px-4 flex items-center gap-2">
          <span className="material-symbols-outlined" style={{ fontSize: '14px' }}>
            filter_list
          </span>
          <span>{t('assets.filters')}</span>
        </button>
      </div>

      {/* Table */}
      <div className="flex-1 overflow-hidden flex flex-col">
        {loading ? (
          <div className="flex-1 flex items-center justify-center">
            <span className="text-label-caps font-label-caps text-on-surface-variant animate-pulse">
              {t('assets.loading')}
            </span>
          </div>
        ) : error ? (
          <div className="m-8 p-4 border border-error bg-error-container/10 flex items-center gap-3">
            <div className="w-2 h-2 bg-error shrink-0" />
            <span className="text-label-caps font-label-caps text-error">{error}</span>
          </div>
        ) : (
          <>
            {/* Table Header */}
            <div className="cyber-grid-header flex items-center text-label-caps font-label-caps text-on-surface-variant shrink-0">
              <div className="w-20 py-3 px-4 shrink-0">{t('assets.table.id')}</div>
              <div className="flex-1 py-3 px-4 min-w-0">{t('assets.table.assetName')}</div>
              <div className="w-28 py-3 px-4 shrink-0">{t('assets.table.lifecycle')}</div>
              <div className="w-32 py-3 px-4 shrink-0">{t('assets.table.criticality')}</div>
              <div className="w-36 py-3 px-4 shrink-0">{t('assets.table.riskProfile')}</div>
              <div className="w-12 py-3 px-4 shrink-0" />
            </div>

            {/* Rows */}
            <div className="flex-1 overflow-y-auto cyber-scrollbar">
              {filtered.map((p) => (
                <div
                  key={p.id}
                  onClick={() => navigate(`/products/${p.id}`)}
                  className="cyber-grid-row flex items-center cursor-pointer group/row"
                >
                  <div className="w-20 py-3 px-4 text-mono-data font-mono-data text-on-surface-variant shrink-0">
                    {String(p.id).padStart(4, '0')}
                  </div>
                  <div className="flex-1 py-3 px-4 min-w-0">
                    <div className="text-mono-data font-mono-data text-primary truncate font-bold">
                      {p.name}
                    </div>
                    <div className="text-label-caps font-label-caps text-on-surface-variant opacity-50 truncate mt-0.5">
                      {p.description ?? t('assets.noDescription')}
                    </div>
                  </div>
                  <div className="w-28 py-3 px-4 shrink-0">
                    <span className="border border-outline-variant px-2 py-0.5 text-label-caps font-label-caps text-on-surface-variant">
                      {p.lifecycle}
                    </span>
                  </div>
                  <div className="w-32 py-3 px-4 text-label-caps font-label-caps text-on-surface-variant shrink-0">
                    {p.business_criticality}
                  </div>
                  <div className="w-36 py-3 px-4 shrink-0">
                    <div className="flex items-center gap-2">
                      <div className="flex flex-col text-label-caps font-label-caps leading-snug">
                        <span className="text-severity-critical">{t('assets.criticalInitial')} {p.sla_critical}</span>
                        <span className="text-severity-high">{t('assets.highInitial')} {p.sla_high}</span>
                      </div>
                      <div className="flex-grow progress-bar-container max-w-[50px] h-1.5 shrink-0">
                        <div
                          className="progress-bar-fill"
                          style={{
                            width: `${Math.min(100, p.sla_critical * 20 + p.sla_high * 10)}%`,
                            background:
                              p.sla_critical > 0
                                ? 'linear-gradient(90deg, #F87171, #EF4444)'
                                : 'linear-gradient(90deg, #FDBA74, #F97316)',
                          }}
                        />
                      </div>
                    </div>
                  </div>
                  <div className="w-12 py-3 px-4 flex justify-center shrink-0 opacity-0 group-hover/row:opacity-100">
                    <span
                      className="material-symbols-outlined text-primary"
                      style={{ fontSize: '18px' }}
                    >
                      chevron_right
                    </span>
                  </div>
                </div>
              ))}
              {filtered.length === 0 && (
                <div className="py-20 text-center">
                  <span className="text-label-caps font-label-caps text-on-surface-variant opacity-30">
                    {t('assets.noAssetsFound')}
                  </span>
                </div>
              )}
            </div>
          </>
        )}
      </div>
    </div>
  );
};
