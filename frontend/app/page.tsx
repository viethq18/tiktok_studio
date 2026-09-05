'use client';

/**
 * Landing page.
 *
 * It follows the shape a visitor expects from a SaaS site — nav with pricing,
 * hero with a primary and a secondary action, proof, problem, how it works,
 * feature deep-dives, who it fits, pricing preview, FAQ, closing CTA, footer —
 * while using the same palette and primitives as the signed-in product, so
 * arriving in the app feels like staying in one place.
 *
 * There is deliberately no testimonial or customer-logo strip: inventing either
 * would be fabricated social proof. The proof shown is what the product
 * verifiably does.
 */
import Link from 'next/link';
import * as React from 'react';

import { Faq, LandingFooter, LandingHeader, Section, useAuthTarget } from '@/components/landing/Chrome';
import { DeepDives, FactStrip } from '@/components/landing/DeepDives';
import { PlanCards } from '@/components/landing/PricingTable';
import { SlideStack } from '@/components/landing/SlideStack';
import { Button, Card } from '@/components/ui/primitives';
import { useT } from '@/lib/i18n';
import { cn } from '@/lib/utils';

/**
 * Advances a character counter on an interval. Kept as a hook so the state
 * update lives in the timer callback rather than in an effect body.
 */
function useTypewriter(length: number, stepMs: number): number {
  const [chars, setChars] = React.useState(0);

  React.useEffect(() => {
    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
      const done = setTimeout(() => setChars(length), 0);
      return () => clearTimeout(done);
    }
    const id = setInterval(() => {
      setChars((n) => {
        if (n >= length) {
          clearInterval(id);
          return n;
        }
        return n + 1;
      });
    }, stepMs);
    return () => clearInterval(id);
  }, [length, stepMs]);

  return chars;
}

export default function LandingPage() {
  return (
    <div className="min-h-screen bg-neutral-50">
      <LandingHeader />
      <Hero />
      <Problem />
      <Steps />
      <Deep />
      <Features />
      <UseCases />
      <PricingPreview />
      <FaqSection />
      <Close />
      <LandingFooter />
    </div>
  );
}

function Hero() {
  const t = useT();
  const { target, signedIn } = useAuthTarget();
  const sentence = t('landing.typed');
  const chars = useTypewriter(sentence.length, 34);
  const done = chars >= sentence.length;

  return (
    <>
      <section className="mx-auto grid max-w-6xl gap-12 px-6 py-16 lg:grid-cols-12 lg:gap-10 lg:py-20">
        <div className="lg:col-span-5">
          <p className="text-sm text-neutral-500">{t('landing.hero.eyebrow')}</p>
          {/* Each sentence is its own block so text-balance can break it on its
              own terms — `white-space: pre-line` suppresses balancing. */}
          <h1 className="mt-3 text-4xl font-semibold leading-[1.05] tracking-tight text-neutral-900 sm:text-5xl">
            {t('landing.hero.title')
              .split('\n')
              .map((line) => (
                <span key={line} className="block text-balance">
                  {line}
                </span>
              ))}
          </h1>
          <p className="mt-5 max-w-[54ch] leading-relaxed text-neutral-600">{t('landing.hero.body')}</p>

          {/* The input the whole product starts from, typing itself once. */}
          <Card className="mt-7 p-4">
            <p className="text-sm leading-relaxed text-neutral-900">
              {sentence.slice(0, chars)}
              <span
                className={cn(
                  'ml-0.5 inline-block h-[1.05em] w-[2px] translate-y-[0.18em] bg-neutral-900',
                  !done && 'caret',
                )}
              />
            </p>
          </Card>

          <div className="mt-7 flex flex-wrap items-center gap-3">
            <Link href={target}>
              <Button size="lg">{signedIn ? t('landing.hero.ctaSignedIn') : t('landing.hero.cta')}</Button>
            </Link>
            <Link href="/pricing">
              <Button size="lg" variant="secondary">
                {t('landing.hero.secondary')}
              </Button>
            </Link>
          </div>
          <p className="mt-3 text-sm text-neutral-500">{t('landing.hero.note')}</p>
        </div>

        <div className="lg:col-span-7 lg:pt-8">
          <SlideStack typed={done} />
        </div>
      </section>

      <div className="border-t border-neutral-200 bg-white">
        <div className="mx-auto max-w-6xl px-6 py-8">
          <FactStrip />
        </div>
      </div>
    </>
  );
}

function Problem() {
  const t = useT();
  return (
    <Section>
      <div className="max-w-2xl">
        <h2 className="text-2xl font-semibold tracking-tight sm:text-3xl">{t('landing.problem.title')}</h2>
        {/* The chores are struck through rather than listed: the design says what
            happens to them, so the copy does not have to. */}
        <p className="mt-5 leading-relaxed text-neutral-400 line-through decoration-neutral-300">
          {t('landing.chores')}
        </p>
        <p className="mt-6 leading-relaxed text-neutral-900">{t('landing.problem.after')}</p>
      </div>
    </Section>
  );
}

function Steps() {
  const t = useT();
  const steps = [
    { title: t('landing.step1.title'), body: t('landing.step1.body') },
    { title: t('landing.step2.title'), body: t('landing.step2.body') },
    { title: t('landing.step3.title'), body: t('landing.step3.body') },
    { title: t('landing.step4.title'), body: t('landing.step4.body') },
  ];
  return (
    <Section id="how" tone="raised">
      <h2 className="text-2xl font-semibold tracking-tight sm:text-3xl">{t('landing.steps.title')}</h2>
      {/* This content really is a sequence, so the numbering carries information. */}
      <ol className="mt-10 max-w-3xl border-l border-neutral-200">
        {steps.map((step, i) => (
          <li key={step.title} className="relative pb-10 pl-8 last:pb-0">
            <span className="absolute left-0 top-0 flex h-6 w-6 -translate-x-1/2 items-center justify-center rounded-full bg-neutral-900 text-xs font-medium text-white">
              {i + 1}
            </span>
            <h3 className="font-medium tracking-tight">{step.title}</h3>
            <p className="mt-1.5 max-w-[62ch] leading-relaxed text-neutral-600">{step.body}</p>
          </li>
        ))}
      </ol>
    </Section>
  );
}

function Deep() {
  const t = useT();
  return (
    <Section id="features">
      <h2 className="text-2xl font-semibold tracking-tight sm:text-3xl">{t('landing.deep.title')}</h2>
      <DeepDives />
    </Section>
  );
}

function Features() {
  const t = useT();
  const features = [
    { title: t('landing.feature1.title'), body: t('landing.feature1.body') },
    { title: t('landing.feature2.title'), body: t('landing.feature2.body') },
    { title: t('landing.feature3.title'), body: t('landing.feature3.body') },
    { title: t('landing.feature4.title'), body: t('landing.feature4.body') },
    { title: t('landing.feature5.title'), body: t('landing.feature5.body') },
    { title: t('landing.feature6.title'), body: t('landing.feature6.body') },
  ];
  return (
    <Section tone="raised">
      <h2 className="text-2xl font-semibold tracking-tight sm:text-3xl">{t('landing.features.title')}</h2>
      {/* Hairline rules rather than cards: nothing here needs a container of its own. */}
      <div className="mt-8 grid gap-x-14 sm:grid-cols-2">
        {features.map((f) => (
          <div key={f.title} className="border-t border-neutral-200 py-6">
            <h3 className="font-medium tracking-tight">{f.title}</h3>
            <p className="mt-1.5 max-w-[52ch] text-sm leading-relaxed text-neutral-600">{f.body}</p>
          </div>
        ))}
      </div>
    </Section>
  );
}

function UseCases() {
  const t = useT();
  const cases = [
    { title: t('landing.case1.title'), body: t('landing.case1.body'), seed: 'usecase-baby' },
    { title: t('landing.case2.title'), body: t('landing.case2.body'), seed: 'usecase-money' },
    { title: t('landing.case3.title'), body: t('landing.case3.body'), seed: 'usecase-health' },
  ];
  return (
    <Section>
      <h2 className="text-2xl font-semibold tracking-tight sm:text-3xl">{t('landing.cases.title')}</h2>
      <p className="mt-3 max-w-[60ch] text-neutral-600">{t('landing.cases.body')}</p>
      <div className="mt-8 grid gap-5 sm:grid-cols-3">
        {cases.map((useCase) => (
          <Card key={useCase.title} className="overflow-hidden">
            <div className="aspect-[16/9] bg-neutral-100">
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img
                src={`https://picsum.photos/seed/${useCase.seed}/480/270`}
                alt=""
                className="h-full w-full object-cover"
              />
            </div>
            <div className="p-5">
              <h3 className="font-medium tracking-tight">{useCase.title}</h3>
              <p className="mt-1.5 text-sm leading-relaxed text-neutral-600">{useCase.body}</p>
            </div>
          </Card>
        ))}
      </div>
    </Section>
  );
}

function PricingPreview() {
  const t = useT();
  return (
    <Section tone="raised">
      <div className="max-w-2xl">
        <h2 className="text-2xl font-semibold tracking-tight sm:text-3xl">{t('landing.pricing.title')}</h2>
        <p className="mt-3 text-neutral-600">{t('landing.pricing.body')}</p>
      </div>
      <div className="mt-8">
        {/* Yearly by default: it is the cheaper option, so it is the honest one to show first. */}
        <PlanCards yearly />
      </div>
      <Link
        href="/pricing"
        className="mt-6 inline-block text-sm font-medium text-neutral-900 underline underline-offset-4"
      >
        {t('landing.pricing.link')}
      </Link>
    </Section>
  );
}

function FaqSection() {
  const t = useT();
  return (
    <Section id="faq">
      <h2 className="text-2xl font-semibold tracking-tight sm:text-3xl">{t('landing.faq.title')}</h2>
      <Faq
        items={[
          { q: t('landing.faq.q1'), a: t('landing.faq.a1') },
          { q: t('landing.faq.q2'), a: t('landing.faq.a2') },
          { q: t('landing.faq.q3'), a: t('landing.faq.a3') },
          { q: t('landing.faq.q4'), a: t('landing.faq.a4') },
          { q: t('landing.faq.q5'), a: t('landing.faq.a5') },
          { q: t('landing.faq.q6'), a: t('landing.faq.a6') },
        ]}
      />
    </Section>
  );
}

function Close() {
  const t = useT();
  const { target, signedIn } = useAuthTarget();
  return (
    <Section tone="raised">
      <div className="mx-auto max-w-2xl text-center">
        <h2 className="text-2xl font-semibold leading-tight tracking-tight sm:text-4xl">
          {t('landing.cta.title')}
        </h2>
        <Link href={target} className="mt-8 inline-block">
          <Button size="lg">{signedIn ? t('landing.hero.ctaSignedIn') : t('landing.cta.button')}</Button>
        </Link>
        <p className="mt-3 text-sm text-neutral-500">{t('landing.hero.note')}</p>
      </div>
    </Section>
  );
}
