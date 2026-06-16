import React from 'react';

interface MechanicalButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'outline' | 'error';
  size?: 'sm' | 'md' | 'lg';
  children: React.ReactNode;
}

export const MechanicalButton: React.FC<MechanicalButtonProps> = ({
  variant = 'secondary',
  size = 'md',
  children,
  className = '',
  ...props
}) => {
  const baseStyles =
    'transition-none active:translate-y-px disabled:opacity-40 disabled:pointer-events-none uppercase tracking-widest font-mono';

  const variantStyles = {
    primary: 'bg-primary text-on-primary hover:bg-primary-fixed-dim',
    secondary:
      'bg-transparent border border-outline-variant text-primary hover:bg-surface-container-high',
    outline:
      'bg-transparent border border-outline-variant text-on-surface-variant hover:text-primary hover:border-primary',
    error: 'bg-transparent border border-error text-error hover:bg-error/10',
  };

  const sizeStyles = {
    sm: 'px-4 py-1.5 text-label-xs',
    md: 'px-6 py-2.5 text-label-sm',
    lg: 'px-8 py-4 text-label-md',
  };

  return (
    <button
      className={`${baseStyles} ${variantStyles[variant]} ${sizeStyles[size]} ${className}`}
      {...props}
    >
      {children}
    </button>
  );
};
