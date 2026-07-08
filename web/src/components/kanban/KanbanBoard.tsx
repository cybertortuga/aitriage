import React, { useState } from 'react';
import {
  DndContext,
  DragOverlay,
  closestCorners,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
} from '@dnd-kit/core';
import type { DragStartEvent, DragOverEvent, DragEndEvent } from '@dnd-kit/core';
import { arrayMove, sortableKeyboardCoordinates } from '@dnd-kit/sortable';
import type { Finding } from '../../types';
import { Column } from './Column';
import { FindingCard } from './FindingCard';
import api from '../../services/api';
import FindingDetailModal from '../findings/FindingDetailModal';
import { useTranslation } from 'react-i18next';

const COLUMNS = ['backlog', 'todo', 'in_progress', 'review', 'done'] as const;

interface KanbanBoardProps {
  findings: Finding[];
  onFindingsChange: (findings: Finding[]) => void;
}

export const KanbanBoard: React.FC<KanbanBoardProps> = ({ findings, onFindingsChange }) => {
  const { t } = useTranslation('pages');
  const [activeId, setActiveId] = useState<number | null>(null);
  const [selectedFindingId, setSelectedFindingId] = useState<number | null>(null);

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 5,
      },
    }),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates }),
  );

  const handleDragStart = (event: DragStartEvent) => {
    setActiveId(event.active.id as number);
  };

  const handleDragOver = (event: DragOverEvent) => {
    const { active, over } = event;
    if (!over) return;

    const activeId = active.id;
    const overId = over.id;

    if (activeId === overId) return;

    const isActiveTask = active.data.current?.type === 'Task';
    const isOverTask = over.data.current?.type === 'Task';
    const isOverColumn = over.data.current?.type === 'Column';

    if (!isActiveTask) return;

    if (isActiveTask && isOverTask) {
      const activeIndex = findings.findIndex((t) => t.id === activeId);
      const overIndex = findings.findIndex((t) => t.id === overId);

      if (findings[activeIndex].kanban_column !== findings[overIndex].kanban_column) {
        const newFindings = [...findings];
        newFindings[activeIndex].kanban_column = findings[overIndex].kanban_column;
        onFindingsChange(arrayMove(newFindings, activeIndex, overIndex));
      } else {
        onFindingsChange(arrayMove(findings, activeIndex, overIndex));
      }
    }

    if (isActiveTask && isOverColumn) {
      const activeIndex = findings.findIndex((t) => t.id === activeId);
      const newFindings = [...findings];
      newFindings[activeIndex].kanban_column = overId as string;
      onFindingsChange(arrayMove(newFindings, activeIndex, activeIndex));
    }
  };

  const handleDragEnd = async (event: DragEndEvent) => {
    setActiveId(null);
    const { active, over } = event;
    if (!over) return;

    const activeFinding = findings.find((f) => f.id === active.id);
    if (!activeFinding) return;

    const newColumn =
      over.data.current?.type === 'Column' ? over.id : over.data.current?.task?.kanban_column;

    if (newColumn && newColumn !== activeFinding.kanban_column) {
      try {
        await api.patch(`/findings/${active.id}/move`, {
          kanban_column: newColumn,
        });
      } catch (err) {
        console.error('Failed to update kanban status', err);
      }
    }
  };

  const activeFinding = findings.find((f) => f.id === activeId);

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCorners}
      onDragStart={handleDragStart}
      onDragOver={handleDragOver}
      onDragEnd={handleDragEnd}
    >
      <div className="flex gap-6 h-full overflow-x-auto pb-4 items-start">
        {COLUMNS.map((col) => (
          <Column
            key={col}
            id={col}
            title={t(`kanban.columns.${col}`)}
            tasks={findings.filter((f) => f.kanban_column === col)}
            onFindingClick={(finding) => setSelectedFindingId(finding.id)}
          />
        ))}
      </div>
      <DragOverlay>
        {activeFinding ? <FindingCard finding={activeFinding} isOverlay /> : null}
      </DragOverlay>

      {selectedFindingId && (
        <FindingDetailModal
          finding={findings.find((f) => f.id === selectedFindingId) || null}
          isOpen={!!selectedFindingId}
          onClose={() => setSelectedFindingId(null)}
          onUpdate={() => {
            // Refresh findings logic could go here if needed,
            // or we rely on the prop updating from parent
          }}
        />
      )}
    </DndContext>
  );
};
