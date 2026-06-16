import React, { useState, useMemo } from 'react';
import { useTranslation } from 'react-i18next';

interface Column {
  key: string;
  label: string;
  width?: string;
  render?: (value: any, row: any) => React.ReactNode;
}

interface DataTableProps {
  columns: Column[];
  data: any[];
  onRowSelect?: (row: any, index: number) => void;
  selectedIndex?: number;
  searchable?: boolean;
  compact?: boolean;
  emptyMessage?: string;
  maxHeight?: string;
}

export const DataTable: React.FC<DataTableProps> = ({
  columns,
  data,
  onRowSelect,
  selectedIndex = -1,
  searchable = false,
  compact = false,
  emptyMessage,
  maxHeight = '100%',
}) => {
  const { t } = useTranslation('components');
  const [search, setSearch] = useState('');
  const [sortKey, setSortKey] = useState<string | null>(null);
  const [sortDir, setSortDir] = useState<'asc' | 'desc'>('asc');

  const filtered = useMemo(() => {
    let result = data;
    if (search) {
      const q = search.toLowerCase();
      result = data.filter((row) =>
        columns.some((col) => {
          const val = row[col.key];
          return val && String(val).toLowerCase().includes(q);
        }),
      );
    }
    if (sortKey) {
      result = [...result].sort((a, b) => {
        const av = a[sortKey] ?? '';
        const bv = b[sortKey] ?? '';
        const cmp = String(av).localeCompare(String(bv), undefined, { numeric: true });
        return sortDir === 'asc' ? cmp : -cmp;
      });
    }
    return result;
  }, [data, search, sortKey, sortDir, columns]);

  const handleSort = (key: string) => {
    if (sortKey === key) {
      setSortDir((d) => (d === 'asc' ? 'desc' : 'asc'));
    } else {
      setSortKey(key);
      setSortDir('asc');
    }
  };

  const py = compact ? 'py-1.5' : 'py-3';

  return (
    <div className="flex flex-col h-full">
      {searchable && (
        <div className="px-4 py-2 border-b border-outline-variant bg-surface-container-lowest/30">
          <div className="flex items-center gap-2">
            <span
              className="material-symbols-outlined text-on-surface-variant/40"
              style={{ fontSize: '16px' }}
            >
              search
            </span>
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder={t('components.dataTable.filter')}
              className="bg-transparent border-none outline-none text-mono-data text-on-surface flex-1 placeholder:text-on-surface-variant/30 placeholder:uppercase placeholder:tracking-widest placeholder:text-[10px]"
            />
            {search && (
              <button
                onClick={() => setSearch('')}
                className="text-on-surface-variant/40 hover:text-primary"
              >
                <span className="material-symbols-outlined" style={{ fontSize: '14px' }}>
                  close
                </span>
              </button>
            )}
            <span className="text-label-xs text-on-surface-variant/30">{filtered.length} {t('components.dataTable.rows')}</span>
          </div>
        </div>
      )}

      {/* Header */}
      <div className="flex px-4 py-2 border-b border-outline-variant bg-surface-container-lowest/50 shrink-0">
        {columns.map((col) => (
          <button
            key={col.key}
            onClick={() => handleSort(col.key)}
            className="text-left text-label-xs text-on-surface-variant/60 hover:text-primary uppercase tracking-widest transition-none flex items-center gap-1"
            style={{ width: col.width || `${100 / columns.length}%` }}
          >
            {col.label}
            {sortKey === col.key && (
              <span className="text-primary" style={{ fontSize: '10px' }}>
                {sortDir === 'asc' ? '▲' : '▼'}
              </span>
            )}
          </button>
        ))}
      </div>

      {/* Body */}
      <div className="flex-1 overflow-y-auto cyber-scrollbar" style={{ maxHeight }}>
        {filtered.length === 0 ? (
          <div className="flex items-center justify-center py-16 opacity-20">
            <span className="text-label-caps tracking-[0.3em]">{emptyMessage || t('components.dataTable.noData')}</span>
          </div>
        ) : (
          filtered.map((row, i) => (
            <div
              key={i}
              onClick={() => onRowSelect?.(row, i)}
              className={`flex px-4 ${py} border-b border-outline-variant/20 cursor-pointer transition-none ${
                selectedIndex === i
                  ? 'bg-primary/10 text-primary border-l-2 border-l-primary'
                  : 'hover:bg-surface-container-high'
              }`}
            >
              {columns.map((col) => (
                <div
                  key={col.key}
                  className="text-mono-data truncate pr-2"
                  style={{ width: col.width || `${100 / columns.length}%` }}
                >
                  {col.render ? col.render(row[col.key], row) : (row[col.key] ?? '—')}
                </div>
              ))}
            </div>
          ))
        )}
      </div>
    </div>
  );
};
