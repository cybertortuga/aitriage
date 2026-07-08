import React, { forwardRef } from 'react';
import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

interface TuiPanelProps {
  children: React.ReactNode;
  title?: string;
  className?: string;
  headerRight?: React.ReactNode;
}

export const TuiPanel = forwardRef<HTMLDivElement, TuiPanelProps>(
  ({ children, title, className, headerRight }, ref) => {
    return (
      <div
        ref={ref}
        className={cn(
          'bg-surface-container border border-outline-variant flex flex-col relative',
          className,
        )}
      >
        {/* Accent corners */}
        <div className="absolute top-0 left-0 w-1 h-1 bg-primary-fixed-dim" />
        <div className="absolute bottom-0 right-0 w-1 h-1 bg-primary-fixed-dim" />

        {title && (
          <div className="flex items-center justify-between border-b border-outline-variant px-4 py-2 bg-surface-container-high/50">
            <div className="flex items-center gap-3">
              <div className="w-1.5 h-1.5 bg-primary-fixed-dim" />
              <h3 className="text-[10px] font-black uppercase tracking-[0.2em] text-on-surface italic">
                {title}
              </h3>
            </div>
            {headerRight && <div className="flex items-center">{headerRight}</div>}
          </div>
        )}
        <div className="flex-1 overflow-auto">{children}</div>
      </div>
    );
  },
);

TuiPanel.displayName = 'TuiPanel';
