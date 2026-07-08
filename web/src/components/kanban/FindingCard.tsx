import React from 'react';
import { useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import type { Finding } from '../../types';
import { useTranslation } from 'react-i18next';

interface Props {
  finding: Finding;
  isOverlay?: boolean;
  onClick?: () => void;
}

export const FindingCard: React.FC<Props> = ({ finding, isOverlay, onClick }) => {
  const { t } = useTranslation('pages');
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: finding.id,
  });

  const style = {
    transform: CSS.Translate.toString(transform),
    transition,
    opacity: isDragging ? 0.3 : 1,
    background: 'var(--color-surface-container-low)',
  };

  const getSeverityLabel = (severity: string) => {
    const s = severity?.toUpperCase();
    if (s === 'CRITICAL') return t('critical');
    if (s === 'HIGH') return t('high');
    if (s === 'MEDIUM') return t('medium');
    if (s === 'LOW') return t('low');
    return severity;
  };

  const severityColor =
    finding.severity === 'CRITICAL'
      ? 'text-error'
      : finding.severity === 'HIGH'
        ? 'text-amber-500'
        : 'text-on-surface-variant';

  const severityBorder =
    finding.severity === 'CRITICAL'
      ? 'border-error/50'
      : finding.severity === 'HIGH'
        ? 'border-amber-500/40'
        : 'border-outline-variant';

  return (
    <div
      ref={setNodeRef}
      style={style}
      {...attributes}
      {...listeners}
      onClick={() => {
        // Prevent opening if it was a drag (isDragging is true during drag)
        if (isDragging) return;
        onClick?.();
      }}
      className={`group border ${severityBorder} p-3 cursor-grab active:cursor-grabbing hover:border-primary transition-none relative ${
        isOverlay ? 'opacity-90' : ''
      }`}
    >
      {/* Top Bar */}
      <div className="flex justify-between items-start mb-3">
        <div className="flex flex-col flex-1">
          <div className="flex items-center gap-2 mb-1">
            <div className="text-[8px] font-mono font-black opacity-40 uppercase tracking-tighter">
              {finding.id} // {finding.file_path || finding.file}
            </div>
            {finding.stack && (
              <div className="px-1.5 py-0.5 border border-primary/30 bg-primary/5 text-[7px] font-black uppercase tracking-widest text-primary flex items-center shrink-0">
                {finding.stack}
              </div>
            )}
          </div>
          <div
            className={`text-[10px] font-black uppercase tracking-tight italic line-clamp-2 leading-tight ${severityColor}`}
          >
            {finding.title}
          </div>
        </div>
      </div>

      {/* Middle Content */}
      <div className="flex items-center gap-3 mt-4 pt-4 border-t border-outline-variant/30">
        <div className="flex flex-col">
          <div className="text-[7px] opacity-40 uppercase font-black">{t('kanban.card.severity')}</div>
          <div className={`text-[9px] font-black uppercase tracking-widest ${severityColor}`}>
            {getSeverityLabel(finding.severity).toUpperCase()}
          </div>
        </div>
        <div className="w-px h-6 bg-outline-variant/30" />
        <div className="flex flex-col">
          <div className="text-[7px] opacity-40 uppercase font-black">{t('kanban.card.slaClock')}</div>
          <div className="text-[9px] font-black font-mono uppercase tracking-tighter text-on-surface">
            {finding.severity === 'CRITICAL'
              ? t('kanban.card.hoursLeft', { hours: 12 })
              : t('kanban.card.hoursLeft', { hours: 48 })}
          </div>
        </div>
      </div>

      {/* Bottom Actions Reveal */}
      <div className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-none">
        <span className="material-symbols-outlined text-sm opacity-40 hover:text-primary-fixed-dim">
          more_vert
        </span>
      </div>
    </div>
  );
};
