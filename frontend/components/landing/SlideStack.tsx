'use client';

/**
 * The hero shows the product's own output: three 4:5 slides in the same
 * proportion, overlay and typographic hierarchy the generator produces. The
 * artefact is the argument, so it is drawn honestly rather than illustrated.
 */
import * as React from 'react';

import { useT } from '@/lib/i18n';
import { cn } from '@/lib/utils';

type SlideSpec = {
  seed: string;
  role: 'hook' | 'point' | 'cta';
  headline: string;
  body?: string;
  tilt: number;
  delay: number;
};

export function SlideStack({ typed }: { typed: boolean }) {
  const t = useT();

  const slides: SlideSpec[] = [
    { seed: 'nightlamp', role: 'hook', headline: t('landing.slide1.headline'), tilt: -4, delay: 0 },
    {
      seed: 'quietroom',
      role: 'point',
      headline: t('landing.slide2.headline'),
      body: t('landing.slide2.body'),
      tilt: 1,
      delay: 110,
    },
    { seed: 'warmhands', role: 'cta', headline: t('landing.slide3.headline'), tilt: 5, delay: 220 },
  ];

  return (
    <div className="flex items-center justify-center gap-3 sm:gap-5">
      {slides.map((slide, i) => (
        <article
          key={slide.seed}
          style={
            {
              '--tilt': `${slide.tilt}deg`,
              animationDelay: typed ? `${slide.delay}ms` : undefined,
            } as React.CSSProperties
          }
          className={cn(
            'relative aspect-4/5 w-[30%] max-w-[210px] shrink-0 overflow-hidden rounded-xl',
            'border border-neutral-200 shadow-[0_16px_40px_-24px_rgba(23,23,23,0.5)]',
            i === 1 && 'z-10 w-[34%] max-w-[236px]',
            typed ? 'deal-in' : 'opacity-0',
          )}
        >
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img
            src={`https://picsum.photos/seed/${slide.seed}/540/675`}
            alt=""
            className="absolute inset-0 h-full w-full object-cover"
          />
          <div className="absolute inset-0 bg-neutral-900/60" />
          <div
            className={cn(
              'absolute inset-0 flex flex-col p-4',
              slide.role === 'point' ? 'justify-end' : 'justify-center',
            )}
          >
            {slide.role === 'hook' ? <span className="mb-3 h-[3px] w-8 bg-white/80" /> : null}
            <h3
              className={cn(
                'font-semibold leading-[1.15] text-white',
                slide.role === 'hook' ? 'text-sm sm:text-base' : 'text-xs sm:text-sm',
              )}
            >
              {slide.headline}
            </h3>
            {slide.body ? (
              <p className="mt-2 text-[10px] leading-snug text-white/75">{slide.body}</p>
            ) : null}
          </div>
          <span className="absolute left-3 top-3 rounded bg-black/45 px-1.5 py-0.5 text-[10px] tabular-nums text-white/80">
            {String(i + 1).padStart(2, '0')}
          </span>
        </article>
      ))}
    </div>
  );
}
