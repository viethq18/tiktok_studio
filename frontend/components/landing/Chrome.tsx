'use client';

/**
 * Nav and footer shared by the landing and pricing pages, so the public side of
 * the product has one chrome and matches the signed-in shell.
 */
import Link from 'next/link';
import * as React from 'react';

import { LocaleSwitch } from '@/components/ui/LocaleSwitch';
import { Button } from '@/components/ui/primitives';
import { useMe } from '@/lib/api/hooks';
import { useT } from '@/lib/i18n';
import { cn } from '@/lib/utils';

export function useAuthTarget() {
  const { data: user } = useMe();
  return { target: user ? '/projects' : '/login', signedIn: Boolean(user) };
}

export function LandingHeader({ active }: { active?: 'pricing' }) {
  const t = useT();
  const { target, signedIn } = useAuthTarget();

  const links = [
    { href: '/#how', label: t('landing.nav.how') },
    { href: '/#features', label: t('landing.nav.featuresNav') },
    { href: '/pricing', label: t('landing.nav.pricing'), key: 'pricing' as const },
  ];

  return (
    <header className="sticky top-0 z-30 border-b border-neutral-200 bg-white/90 backdrop-blur">
      <div className="mx-auto flex h-14 max-w-6xl items-center gap-6 px-6">
        <Link href="/" className="text-sm font-semibold tracking-tight">
          Carousel Studio
        </Link>
        <nav className="hidden items-center gap-5 md:flex">
          {links.map((link) => (
            <Link
              key={link.href}
              href={link.href}
              className={cn(
                'text-sm transition hover:text-neutral-900',
                active && link.key === active ? 'text-neutral-900' : 'text-neutral-500',
              )}
            >
              {link.label}
            </Link>
          ))}
        </nav>
        <div className="ml-auto flex items-center gap-3">
          <LocaleSwitch />
          {signedIn ? null : (
            <Link href="/login" className="hidden text-sm text-neutral-500 transition hover:text-neutral-900 sm:block">
              {t('landing.nav.signIn')}
            </Link>
          )}
          <Link href={target}>
            <Button size="sm">{signedIn ? t('landing.hero.ctaSignedIn') : t('landing.nav.start')}</Button>
          </Link>
        </div>
      </div>
    </header>
  );
}

export function LandingFooter() {
  const t = useT();
  return (
    <footer className="border-t border-neutral-200 bg-white">
      <div className="mx-auto max-w-6xl px-6 py-12">
        <div className="grid gap-8 sm:grid-cols-2 lg:grid-cols-4">
          <div className="lg:col-span-2">
            <p className="text-sm font-semibold tracking-tight">{t('landing.footer')}</p>
            <p className="mt-2 max-w-[38ch] text-sm text-neutral-500">{t('footer.tagline')}</p>
            <div className="mt-4">
              <LocaleSwitch className="w-fit" />
            </div>
          </div>

          <FooterColumn title={t('footer.product')}>
            <FooterLink href="/#how">{t('footer.howItWorks')}</FooterLink>
            <FooterLink href="/#features">{t('footer.features')}</FooterLink>
            <FooterLink href="/pricing">{t('footer.pricing')}</FooterLink>
          </FooterColumn>

          <FooterColumn title={t('footer.resources')}>
            <FooterLink href="/#faq">{t('footer.faq')}</FooterLink>
            {/* Honest about what does not exist yet rather than linking nowhere. */}
            <FooterPending>{t('footer.docs')}</FooterPending>
            <FooterPending>{t('footer.status')}</FooterPending>
          </FooterColumn>
        </div>

        <div className="mt-10 border-t border-neutral-200 pt-6 text-sm text-neutral-500">
          © {new Date().getFullYear()} {t('landing.footer')}
        </div>
      </div>
    </footer>
  );
}

function FooterColumn({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div>
      <p className="text-sm font-medium text-neutral-900">{title}</p>
      <ul className="mt-3 space-y-2">{children}</ul>
    </div>
  );
}

function FooterLink({ href, children }: { href: string; children: React.ReactNode }) {
  return (
    <li>
      <Link href={href} className="text-sm text-neutral-500 transition hover:text-neutral-900">
        {children}
      </Link>
    </li>
  );
}

function FooterPending({ children }: { children: React.ReactNode }) {
  const t = useT();
  return (
    <li className="flex items-center gap-2 text-sm text-neutral-400">
      {children}
      <span className="rounded-full bg-neutral-100 px-1.5 py-0.5 text-[10px] text-neutral-500">
        {t('footer.comingSoon')}
      </span>
    </li>
  );
}

/** A section wrapper with the alternating background the page uses. */
export function Section({
  id,
  tone = 'plain',
  className,
  children,
}: {
  id?: string;
  tone?: 'plain' | 'raised';
  className?: string;
  children: React.ReactNode;
}) {
  return (
    <section id={id} className={cn('border-t border-neutral-200', tone === 'raised' && 'bg-white')}>
      <div className={cn('mx-auto max-w-6xl px-6 py-16 sm:py-20', className)}>{children}</div>
    </section>
  );
}

/** Native disclosure: accessible and keyboard-operable without any JS. */
export function Faq({ items }: { items: { q: string; a: string }[] }) {
  return (
    <div className="mt-8 max-w-3xl divide-y divide-neutral-200 border-y border-neutral-200">
      {items.map((item) => (
        <details key={item.q} className="group py-4">
          <summary className="flex cursor-pointer list-none items-center justify-between gap-4 font-medium tracking-tight">
            {item.q}
            <span className="shrink-0 text-neutral-400 transition group-open:rotate-45" aria-hidden>
              +
            </span>
          </summary>
          <p className="mt-3 max-w-[70ch] leading-relaxed text-neutral-600">{item.a}</p>
        </details>
      ))}
    </div>
  );
}
