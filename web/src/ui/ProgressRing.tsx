import React from 'react';

interface ProgressRingProps {
  radius?: number;
  strokeWidth?: number;
  progress?: number; // 0 to 100
  className?: string;
  size?: number; // total size
  indeterminate?: boolean;
}

export const ProgressRing: React.FC<ProgressRingProps> = ({
  radius = 36,
  strokeWidth = 3,
  progress = 0,
  className = '',
  size = 80,
  indeterminate = false,
}) => {
  const normalizedRadius = radius - strokeWidth;
  const circumference = normalizedRadius * 2 * Math.PI;
  const strokeDashoffset =
    circumference - (Math.min(Math.max(progress, 0), 100) / 100) * circumference;

  return (
    <svg
      height={size}
      width={size}
      className={`relative ${className}`}
      viewBox={`0 0 ${size} ${size}`}
    >
      {/* Background circle */}
      <circle
        stroke="rgba(255, 255, 255, 0.04)"
        fill="transparent"
        strokeWidth={strokeWidth}
        r={normalizedRadius}
        cx={size / 2}
        cy={size / 2}
      />
      {/* Foreground circle with animation */}
      <circle
        stroke="var(--accent-color)" /* dynamic primary accent */
        fill="transparent"
        strokeWidth={strokeWidth}
        strokeDasharray={`${circumference} ${circumference}`}
        r={normalizedRadius}
        cx={size / 2}
        cy={size / 2}
        style={
          indeterminate
            ? undefined
            : {
                strokeDashoffset,
                transform: 'rotate(-90deg)',
                transformOrigin: '50% 50%',
                transition: 'stroke-dashoffset 0.5s cubic-bezier(0.16, 1, 0.3, 1)',
              }
        }
        className={indeterminate ? 'animated-ring origin-center -rotate-90' : 'origin-center'}
        strokeDashoffset={indeterminate ? undefined : strokeDashoffset}
        pathLength={indeterminate ? undefined : undefined}
      />
    </svg>
  );
};
