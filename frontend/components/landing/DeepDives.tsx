'use client';

/**
 * Alternating feature rows. Each visual is a small, honest mock of the real
 * surface it describes — the generated form, the canvas safe area, the caption
 * panel — rather than a decorative illustration.
 */
import { Check, Copy } from 'lucide-react';
import * as React from 'react';

import { useT } from '@/lib/i18n';
import { cn } from '@/lib/utils';

export function DeepDives() {
  const t = useT();

  const rows = [
    { eyebrow: t('landing.deep1.eyebrow'), title: t('landing.deep1.title'), body: t('landing.deep1.body'), visual: <FormVisual /> },
    { eyebrow: t('landing.deep2.eyebrow'), title: t('landing.deep2.title'), body: t('landing.deep2.body'), visual: <CanvasVisual /> },
    { eyebrow: t('landing.deep3.eyebrow'), title: t('landing.deep3.title'), body: t('landing.deep3.body'), visual: <CaptionVisual /> },
  ];

  return (
    <div className="mt-12 space-y-16 sm:space-y-20">
      {rows.map((row, i) => (
        <div key={row.title} className="grid items-center gap-8 lg:grid-cols-2 lg:gap-14">
          <div className={cn(i % 2 === 1 && 'lg:order-2')}>
            <p className="text-sm text-neutral-500">{row.eyebrow}</p>
            <h3 className="mt-2 text-xl font-semibold tracking-tight sm:text-2xl">{row.title}</h3>
            <p className="mt-3 max-w-[58ch] leading-relaxed text-neutral-600">{row.body}</p>
          </div>
          <div className={cn(i % 2 === 1 && 'lg:order-1')}>{row.visual}</div>
        </div>
      ))}
    </div>
  );
}

function Frame({ children }: { children: React.ReactNode }) {
  return (
    <div className="rounded-xl border border-neutral-200 bg-white p-5 shadow-[0_1px_2px_rgba(23,23,23,0.04)]">
      {children}
    </div>
  );
}

function FormVisual() {
  const t = useT();
  return (
    <Frame>
      <div className="space-y-4">
        <div>
          <p className="text-xs font-medium text-neutral-900">{t('landing.deep1.field1')}</p>
          <div className="mt-1.5 rounded-lg border border-neutral-200 px-3 py-2 text-xs text-neutral-500">
            {t('landing.slide1.headline')}
          </div>
        </div>
        <div>
          <p className="text-xs font-medium text-neutral-900">{t('landing.deep1.field2')}</p>
          <div className="mt-1.5 flex flex-wrap gap-1.5">
            {[t('landing.deep1.opt1'), t('landing.deep1.opt2'), t('landing.deep1.opt3')].map((opt, i) => (
              <span
                key={opt}
                className={cn(
                  'rounded-full border px-2.5 py-1 text-xs',
                  i === 0 ? 'border-neutral-900 bg-neutral-900 text-white' : 'border-neutral-200 text-neutral-600',
                )}
              >
                {opt}
              </span>
            ))}
          </div>
        </div>
        <div>
          <p className="text-xs font-medium text-neutral-900">{t('landing.deep1.field3')}</p>
          <div className="mt-2.5 flex items-center gap-2">
            <span className="text-xs text-neutral-400">3</span>
            <div className="relative h-1 flex-1 rounded-full bg-neutral-200">
              <div className="absolute inset-y-0 left-0 w-1/2 rounded-full bg-neutral-900" />
              <div className="absolute -top-1 left-1/2 h-3 w-3 -translate-x-1/2 rounded-full border-2 border-neutral-900 bg-white" />
            </div>
            <span className="text-xs text-neutral-400">7</span>
          </div>
        </div>
      </div>
    </Frame>
  );
}

function CanvasVisual() {
  const t = useT();
  return (
    <Frame>
      <div className="relative mx-auto aspect-4/5 w-full max-w-[260px] overflow-hidden rounded-lg bg-neutral-900">
        {/* eslint-disable-next-line @next/next/no-img-element */}
        <img
          src="https://picsum.photos/seed/safearea/540/675"
          alt=""
          className="absolute inset-0 h-full w-full object-cover opacity-60"
        />
        {/* The safe area, drawn the way the validator enforces it. */}
        <div className="absolute inset-x-[7.4%] inset-y-[5.9%] rounded border border-dashed border-white/50">
          <span className="absolute -top-2 left-2 bg-neutral-900/80 px-1 text-[8px] text-white/80">
            {t('landing.deep2.safeArea')}
          </span>
          <p className="absolute inset-x-2 top-1/3 text-center text-xs font-semibold leading-tight text-white">
            {t('landing.slide1.headline')}
          </p>
        </div>
        <div className="absolute inset-x-0 bottom-0 h-[6%] border-t border-dashed border-white/30 bg-white/5">
          <span className="absolute left-2 top-0.5 text-[8px] text-white/70">{t('landing.deep2.brandBand')}</span>
        </div>
      </div>
    </Frame>
  );
}

function CaptionVisual() {
  const t = useT();
  const tags = ['#fyp', '#xuhuong', '#learnontiktok', '#chamsocbe', '#mevabe'];
  return (
    <Frame>
      <p className="text-xs leading-relaxed text-neutral-700">
        Nhiều bố mẹ chỉ nhận ra bé thiếu ngủ khi bé đã quá mệt.
        <br />
        Lướt hết bộ ảnh để xem 5 dấu hiệu dễ nhận ra nhất.
      </p>
      <div className="mt-3 flex flex-wrap gap-1.5">
        {tags.map((tag) => (
          <span key={tag} className="rounded-full bg-neutral-100 px-2 py-0.5 text-[11px] text-neutral-600">
            {tag}
          </span>
        ))}
      </div>
      <div className="mt-4 flex items-center justify-center gap-2 rounded-lg bg-neutral-900 py-2 text-xs font-medium text-white">
        <Copy className="h-3.5 w-3.5" />
        {t('landing.deep3.copy')}
      </div>
    </Frame>
  );
}

export function FactStrip() {
  const t = useT();
  const facts = [
    { value: '9', label: t('landing.facts.languages') },
    { value: '3', label: t('landing.facts.ratios') },
    { value: '8', label: t('landing.facts.formulas') },
  ];
  return (
    <div className="flex flex-wrap items-center gap-x-10 gap-y-5">
      {facts.map((fact) => (
        <div key={fact.label} className="flex items-baseline gap-2">
          <span className="text-2xl font-semibold tabular-nums">{fact.value}</span>
          <span className="text-sm text-neutral-500">{fact.label}</span>
        </div>
      ))}
      <div className="flex items-center gap-2">
        <Check className="h-4 w-4 text-neutral-900" strokeWidth={2.5} />
        <span className="text-sm text-neutral-700">{t('landing.facts.export')}</span>
        <span className="rounded-full bg-neutral-100 px-2 py-0.5 text-xs text-neutral-500">
          {t('landing.facts.exportNote')}
        </span>
      </div>
    </div>
  );
}
