'use client';

import * as React from 'react';

import { useT } from '@/lib/i18n';
import { cn } from '@/lib/utils';

type ButtonProps = React.ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: 'primary' | 'secondary' | 'ghost' | 'danger';
  size?: 'sm' | 'md' | 'lg';
};

export function Button({ className, variant = 'primary', size = 'md', ...props }: ButtonProps) {
  return (
    <button
      className={cn(
        'inline-flex items-center justify-center gap-2 rounded-lg font-medium transition',
        'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-neutral-900 focus-visible:ring-offset-2',
        'disabled:pointer-events-none disabled:opacity-50',
        size === 'sm' && 'h-8 px-3 text-sm',
        size === 'md' && 'h-10 px-4 text-sm',
        size === 'lg' && 'h-12 px-6 text-base',
        variant === 'primary' && 'bg-neutral-900 text-white hover:bg-neutral-700',
        variant === 'secondary' && 'border border-neutral-200 bg-white text-neutral-900 hover:bg-neutral-50',
        variant === 'ghost' && 'text-neutral-600 hover:bg-neutral-100 hover:text-neutral-900',
        variant === 'danger' && 'bg-red-600 text-white hover:bg-red-500',
        className,
      )}
      {...props}
    />
  );
}

export const Input = React.forwardRef<HTMLInputElement, React.InputHTMLAttributes<HTMLInputElement>>(
  function Input({ className, ...props }, ref) {
    return (
      <input
        ref={ref}
        className={cn(
          'h-10 w-full rounded-lg border border-neutral-200 bg-white px-3 text-sm text-neutral-900',
          'placeholder:text-neutral-400 focus:border-neutral-900 focus:outline-none',
          className,
        )}
        {...props}
      />
    );
  },
);

export const Textarea = React.forwardRef<HTMLTextAreaElement, React.TextareaHTMLAttributes<HTMLTextAreaElement>>(
  function Textarea({ className, ...props }, ref) {
    return (
      <textarea
        ref={ref}
        className={cn(
          'w-full rounded-lg border border-neutral-200 bg-white p-3 text-sm text-neutral-900',
          'placeholder:text-neutral-400 focus:border-neutral-900 focus:outline-none',
          className,
        )}
        {...props}
      />
    );
  },
);

export function Label({ className, ...props }: React.LabelHTMLAttributes<HTMLLabelElement>) {
  return <label className={cn('block text-sm font-medium text-neutral-900', className)} {...props} />;
}

export function Card({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn('rounded-xl border border-neutral-200 bg-white', className)} {...props} />;
}

const badgeTones: Record<string, string> = {
  draft: 'bg-neutral-100 text-neutral-600',
  generating: 'bg-blue-50 text-blue-700',
  ready: 'bg-emerald-50 text-emerald-700',
  archived: 'bg-neutral-100 text-neutral-500',
  failed: 'bg-red-50 text-red-700',
};

export function Badge({ status, children }: { status?: string; children: React.ReactNode }) {
  return (
    <span
      className={cn(
        'inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium',
        badgeTones[status ?? ''] ?? 'bg-neutral-100 text-neutral-600',
      )}
    >
      {children}
    </span>
  );
}

export function Spinner({ className }: { className?: string }) {
  const t = useT();
  return (
    <span
      className={cn(
        'inline-block h-4 w-4 animate-spin rounded-full border-2 border-neutral-300 border-t-neutral-900',
        className,
      )}
      role="status"
      aria-label={t('common.loading')}
    />
  );
}

export function ErrorNote({ children }: { children: React.ReactNode }) {
  if (!children) return null;
  return (
    <p className="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700" role="alert">
      {children}
    </p>
  );
}

export function EmptyState({ title, description, action }: { title: string; description?: string; action?: React.ReactNode }) {
  return (
    <div className="rounded-xl border border-dashed border-neutral-300 px-6 py-16 text-center">
      <p className="text-base font-medium text-neutral-900">{title}</p>
      {description ? <p className="mt-1 text-sm text-neutral-500">{description}</p> : null}
      {action ? <div className="mt-5">{action}</div> : null}
    </div>
  );
}
