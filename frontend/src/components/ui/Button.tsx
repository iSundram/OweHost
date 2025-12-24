import type { ButtonHTMLAttributes, ReactNode } from 'react';

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'outline' | 'ghost' | 'danger';
  size?: 'sm' | 'md' | 'lg';
  isLoading?: boolean;
  loading?: boolean;
  leftIcon?: ReactNode;
  rightIcon?: ReactNode;
  children: ReactNode;
}

export function Button({
  variant = 'primary',
  size = 'md',
  isLoading = false,
  loading = false,
  leftIcon,
  rightIcon,
  children,
  disabled,
  className = '',
  ...props
}: ButtonProps) {
  const isButtonLoading = isLoading || loading;
  const baseStyles = `
    inline-flex items-center justify-center gap-2
    font-medium rounded-lg transition-all duration-200
    focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-offset-[var(--color-background)]
    disabled:opacity-50 disabled:cursor-not-allowed
  `;

  const variants = {
    primary: `
      bg-gradient-to-r from-[#7BA4D0] to-[#E7F0FA]
      hover:from-[#2E5E99] hover:to-[#7BA4D0]
      text-white
      focus:ring-[#E7F0FA]
      shadow-lg hover:shadow-xl
    `,
    secondary: `
      bg-[var(--color-surface)]
      hover:bg-[var(--color-surface-elevated)]
      text-[var(--color-text-primary)]
      border border-[var(--color-border)]
      focus:ring-[#7BA4D0]
    `,
    outline: `
      bg-transparent
      hover:bg-[var(--color-surface)]
      text-[var(--color-text-primary)]
      border border-[var(--color-border)]
      focus:ring-[#7BA4D0]
    `,
    ghost: `
      bg-transparent
      hover:bg-[var(--color-surface)]
      text-[var(--color-text-secondary)]
      hover:text-[var(--color-text-primary)]
      focus:ring-[#7BA4D0]
    `,
    danger: `
      bg-[var(--color-error)]
      hover:bg-red-600
      text-white
      focus:ring-red-500
    `,
  };

  const sizes = {
    sm: 'px-3 py-1.5 text-sm',
    md: 'px-4 py-2 text-sm',
    lg: 'px-6 py-3 text-base',
  };

  return (
    <button
      className={`${baseStyles} ${variants[variant]} ${sizes[size]} ${className}`}
      disabled={disabled || isButtonLoading}
      {...props}
    >
      {isButtonLoading ? (
        <svg
          className="animate-spin h-4 w-4"
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
        >
          <circle
            className="opacity-25"
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            strokeWidth="4"
          />
          <path
            className="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          />
        </svg>
      ) : (
        leftIcon
      )}
      {children}
      {rightIcon}
    </button>
  );
}
