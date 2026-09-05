'use client';

/**
 * Generation progress (§15, §102).
 *
 * The step list comes from the backend job, so the UI always shows what the
 * pipeline is actually doing rather than a generic "Generating…".
 */
import { Check } from 'lucide-react';

import { Spinner } from '@/components/ui/primitives';
import { useT, type TranslationKey } from '@/lib/i18n';
import { cn } from '@/lib/utils';
import type { Job } from '@/types';

export function GenerationProgress({ job }: { job?: Job }) {
  const t = useT();
  const steps = job?.steps ?? [];

  // The backend owns the pipeline shape and sends a stable status per step; the
  // wording is chosen here so it follows the UI language, not the server's.
  const label = (status: string, fallback: string) => {
    const key = `generation.step.${status}` as TranslationKey;
    const translated = t(key);
    return translated === key ? fallback : translated;
  };

  return (
    <div>
      <h2 className="text-lg font-medium tracking-tight">
        {job?.status === 'failed' ? t('generation.failedTitle') : t('generation.title')}
      </h2>

      <div className="mt-4 h-1.5 w-full overflow-hidden rounded-full bg-neutral-200">
        <div
          className={cn('h-full rounded-full transition-all duration-500',
            job?.status === 'failed' ? 'bg-red-500' : 'bg-neutral-900')}
          style={{ width: `${job?.progress ?? 0}%` }}
        />
      </div>

      <ul className="mt-6 space-y-3">
        {steps.map((step) => (
          <li key={step.status} className="flex items-center gap-3 text-sm">
            <span className="flex h-5 w-5 shrink-0 items-center justify-center">
              {step.state === 'done' ? (
                <Check className="h-4 w-4 text-emerald-600" strokeWidth={3} />
              ) : step.state === 'active' ? (
                <Spinner className="h-4 w-4" />
              ) : (
                <span className="h-2 w-2 rounded-full bg-neutral-300" />
              )}
            </span>
            <span
              className={cn(
                step.state === 'done' && 'text-neutral-500',
                step.state === 'active' && 'font-medium text-neutral-900',
                step.state === 'pending' && 'text-neutral-400',
              )}
            >
              {label(step.status, step.label)}
            </span>
          </li>
        ))}
      </ul>

      {job?.status === 'failed' ? (
        <p className="mt-6 rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">
          {job.error_message || t('generation.failedBody')}
        </p>
      ) : null}
    </div>
  );
}
