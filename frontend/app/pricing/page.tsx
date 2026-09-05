'use client';

import * as React from 'react';

import { Faq, LandingFooter, LandingHeader, Section } from '@/components/landing/Chrome';
import { BillingToggle, ComparisonTable, PlanCards } from '@/components/landing/PricingTable';
import { useT } from '@/lib/i18n';

export default function PricingPage() {
  const t = useT();
  const [yearly, setYearly] = React.useState(true);

  return (
    <div className="min-h-screen bg-neutral-50">
      <LandingHeader active="pricing" />

      <section className="mx-auto max-w-6xl px-6 py-16 sm:py-20">
        <div className="max-w-2xl">
          <h1 className="text-4xl font-semibold tracking-tight sm:text-5xl">{t('pricing.title')}</h1>
          <p className="mt-4 text-lg leading-relaxed text-neutral-600">{t('pricing.subtitle')}</p>
        </div>

        <div className="mt-8">
          <BillingToggle yearly={yearly} onChange={setYearly} />
        </div>

        <p className="mt-6 max-w-[70ch] rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-900">
          {t('pricing.notLiveNote')}
        </p>

        <div className="mt-8">
          <PlanCards yearly={yearly} />
        </div>
      </section>

      <Section tone="raised">
        <h2 className="text-2xl font-semibold tracking-tight sm:text-3xl">{t('pricing.compare')}</h2>
        <ComparisonTable />
      </Section>

      <Section>
        <h2 className="text-2xl font-semibold tracking-tight sm:text-3xl">{t('landing.faq.title')}</h2>
        <Faq
          items={[
            { q: t('pricing.faq.q1'), a: t('pricing.faq.a1') },
            { q: t('pricing.faq.q2'), a: t('pricing.faq.a2') },
            { q: t('pricing.faq.q3'), a: t('pricing.faq.a3') },
          ]}
        />
      </Section>

      <LandingFooter />
    </div>
  );
}
