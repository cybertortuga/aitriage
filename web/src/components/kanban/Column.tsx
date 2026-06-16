import React, { useMemo } from 'react';
import { useDroppable } from '@dnd-kit/core';
import { SortableContext, verticalListSortingStrategy } from '@dnd-kit/sortable';
import type { Finding } from '../../types';
import { FindingCard } from './FindingCard';

interface ColumnProps {
  id: string;
  title: string;
  tasks: Finding[];
  onFindingClick?: (finding: Finding) => void;
}

export const Column: React.FC<ColumnProps> = ({ id, title, tasks, onFindingClick }) => {
  const { setNodeRef } = useDroppable({
    id,
    data: {
      type: 'Column',
    },
  });

  const taskIds = useMemo(() => tasks.map((t) => t.id), [tasks]);

  return (
    <div
      ref={setNodeRef}
      className="flex flex-col w-72 min-w-[18rem] flex-shrink-0 overflow-hidden skeuo-panel "
    >
      <div className="px-4 py-3 border-b border-outline-variant/30 flex items-center justify-between bg-surface-container-high/50">
        <span className="text-label-caps font-label-caps text-on-surface-variant tracking-widest">
          {title}
        </span>
        <span className="text-label-caps font-label-caps text-primary tracking-widest">
          {tasks.length}
        </span>
      </div>
      <div className="p-2 flex flex-col gap-2 flex-grow overflow-y-auto min-h-[150px] skeuo-inset border-x-0 border-b-0">
        <SortableContext items={taskIds} strategy={verticalListSortingStrategy}>
          {tasks.map((task) => (
            <FindingCard key={task.id} finding={task} onClick={() => onFindingClick?.(task)} />
          ))}
        </SortableContext>
      </div>
    </div>
  );
};
