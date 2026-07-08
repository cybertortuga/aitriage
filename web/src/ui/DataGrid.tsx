import React from 'react';
import { motion, AnimatePresence } from 'framer-motion';

interface DataGridProps {
  header: React.ReactNode;
  children: React.ReactNode;
}

export const DataGrid: React.FC<DataGridProps> = ({ header, children }) => {
  return (
    <div className="flex-1 flex flex-col min-h-0 border-r border-outline-variant overflow-hidden bg-background">
      <div className="brutalist-grid-header flex-shrink-0">{header}</div>
      <div className="flex-1 overflow-y-auto bg-surface-container scrollbar-hide">
        <AnimatePresence>{children}</AnimatePresence>
      </div>
    </div>
  );
};

interface DataGridRowProps {
  children: React.ReactNode;
  onClick?: () => void;
  active?: boolean;
  idx?: number;
}

export const DataGridRow: React.FC<DataGridRowProps> = ({ children, onClick, active, idx = 0 }) => {
  return (
    <motion.div
      initial={{ opacity: 0, y: 5 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: idx * 0.01 }}
      onClick={onClick}
      className={`brutalist-grid-row ${active ? 'bg-primary-container/10 border-l-2 border-l-primary-container' : ''} ${onClick ? 'cursor-pointer' : ''}`}
    >
      {children}
    </motion.div>
  );
};
