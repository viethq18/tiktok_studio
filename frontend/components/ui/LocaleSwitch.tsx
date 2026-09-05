'use client';

import { Languages } from 'lucide-react';

import { useLocale } from '@/lib/i18n';
import { LOCALES } from '@/lib/i18n/locale';
import { cn } from '@/lib/utils';

export function LocaleSwitch({ className }: { className?: string }) {
  const { locale, setLocale } = useLocale();
  return (
    <div className={cn('flex items-center gap-1 rounded-lg border border-neutral-200 p-0.5', className)}>
      <Languages className="ml-1.5 h-3.5 w-3.5 text-neutral-400" aria-hidden />
      {LOCALES.map((l) => (
        <button
          key={l.code}
          onClick={() => setLocale(l.code)}
          aria-current={locale === l.code}
          className={cn(
            'rounded px-2 py-1 text-xs font-medium transition',
            locale === l.code ? 'bg-neutral-900 text-white' : 'text-neutral-500 hover:bg-neutral-100',
          )}
        >
          {l.short}
        </button>
      ))}
    </div>
  );
}
