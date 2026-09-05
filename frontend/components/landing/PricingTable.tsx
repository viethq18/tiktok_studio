'use client';

/**
 * Pricing cards + comparison table.
 *
 * Billing is not wired up, so no card offers to take money: paid plans show a
 * disabled "billing not live" control and the page states it plainly. Pretending
 * to sell something that cannot be bought would be worse than saying so.
 */
import { Check, Minus } from 'lucide-react';
import Link from 'next/link';
import * as React from 'react';

import { useAuthTarget } from '@/components/landing/Chrome';
import { Button, Card } from '@/components/ui/primitives';
import { useT } from '@/lib/i18n';
import { COMPARISON, PLANS, formatUSD, monthlyPrice, yearlyTotal, type Plan } from '@/lib/plans';
import { cn } from '@/lib/utils';

export function BillingToggle({
  yearly,
  onChange,
}: {
  yearly: boolean;
  onChange: (yearly: boolean) => void;
}) {
  const t = useT();
  return (
    <div className="inline-flex items-center gap-3">
      <div className="inline-flex rounded-lg border border-neutral-200 bg-white p-0.5">
        {[
          { value: false, label: t('pricing.monthly') },
          { value: true, label: t('pricing.yearly') },
        ].map((option) => (
          <button
            key={String(option.value)}
            onClick={() => onChange(option.value)}
            aria-pressed={yearly === option.value}
            className={cn(
              'rounded-md px-4 py-1.5 text-sm font-medium transition',
              yearly === option.value ? 'bg-neutral-900 text-white' : 'text-neutral-600 hover:bg-neutral-50',
            )}
          >
            {option.label}
          </button>
        ))}
      </div>
      <span className="rounded-full bg-emerald-50 px-2.5 py-1 text-xs font-medium text-emerald-700">
        {t('pricing.yearlyBadge')}
      </span>
    </div>
  );
}

export function PlanCards({ yearly }: { yearly: boolean }) {
  return (
    <div className="grid gap-5 lg:grid-cols-3">
      {PLANS.map((plan) => (
        <PlanCard key={plan.id} plan={plan} yearly={yearly} />
      ))}
    </div>
  );
}

function PlanCard({ plan, yearly }: { plan: Plan; yearly: boolean }) {
  const t = useT();
  const { target } = useAuthTarget();
  const price = monthlyPrice(plan, yearly);
  const isFree = plan.monthly === 0;

  return (
    <Card
      className={cn(
        'relative flex flex-col p-6',
        plan.popular && 'border-neutral-900 shadow-[0_1px_0_0_theme(colors.neutral.900)]',
      )}
    >
      {plan.popular ? (
        <span className="absolute -top-2.5 left-6 rounded-full bg-neutral-900 px-2.5 py-0.5 text-xs font-medium text-white">
          {t('pricing.popular')}
        </span>
      ) : null}

      <h3 className="font-semibold tracking-tight">{t(plan.nameKey)}</h3>
      <p className="mt-1 text-sm text-neutral-500">{t(plan.taglineKey)}</p>

      <div className="mt-5 flex items-baseline gap-1">
        <span className="text-4xl font-semibold tracking-tight tabular-nums">{formatUSD(price)}</span>
        <span className="text-sm text-neutral-500">{t('pricing.perMonth')}</span>
      </div>
      <p className="mt-1 h-5 text-xs text-neutral-500">
        {isFree
          ? null
          : yearly
            ? t('pricing.billedYearly', { amount: formatUSD(yearlyTotal(plan)) })
            : t('pricing.billedMonthly')}
      </p>

      <div className="mt-5">
        {isFree ? (
          <Link href={target} className="block">
            <Button className="w-full">{t('pricing.startFree')}</Button>
          </Link>
        ) : (
          <Button variant="secondary" className="w-full" disabled title={t('pricing.notLiveNote')}>
            {t('pricing.notLive')}
          </Button>
        )}
      </div>

      <ul className="mt-6 space-y-2.5">
        {plan.features.map((feature) => (
          <li key={feature.key} className="flex gap-2.5 text-sm text-neutral-700">
            <Check className="mt-0.5 h-4 w-4 shrink-0 text-neutral-900" strokeWidth={2.5} />
            {t(feature.key, feature.vars)}
          </li>
        ))}
      </ul>
    </Card>
  );
}

export function ComparisonTable() {
  const t = useT();
  return (
    <div className="mt-8 overflow-x-auto">
      <table className="w-full min-w-[540px] border-collapse text-sm">
        <thead>
          <tr className="border-b border-neutral-200">
            <th className="py-3 pr-4 text-left font-medium text-neutral-500">{t('pricing.feature')}</th>
            {PLANS.map((plan) => (
              <th key={plan.id} className="px-4 py-3 text-left font-medium">
                {t(plan.nameKey)}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {COMPARISON.map((row) => (
            <tr key={row.key} className="border-b border-neutral-100">
              <td className="py-3 pr-4 text-neutral-700">{t(row.key)}</td>
              {PLANS.map((plan) => {
                const value = row.values[plan.id];
                return (
                  <td key={plan.id} className="px-4 py-3">
                    {typeof value === 'boolean' ? (
                      value ? (
                        <Check className="h-4 w-4 text-neutral-900" strokeWidth={2.5} aria-label="Yes" />
                      ) : (
                        <Minus className="h-4 w-4 text-neutral-300" aria-label="No" />
                      )
                    ) : (
                      <span className="tabular-nums">{value}</span>
                    )}
                  </td>
                );
              })}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
