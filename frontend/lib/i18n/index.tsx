'use client';

/**
 * UI localisation (§164).
 *
 * UI language and content language are independent: this switches the product
 * chrome, while what the AI writes follows the project's own `language` field.
 * The locale is persisted in a cookie so a reload keeps it.
 */
import * as React from 'react';

import { en } from './en';
import { DEFAULT_LOCALE, LOCALE_COOKIE, type Locale } from './locale';
import { vi, type TranslationKey } from './vi';

export { LOCALES, LOCALE_COOKIE, DEFAULT_LOCALE, parseLocale } from './locale';
export type { Locale };

const dictionaries: Record<Locale, Record<TranslationKey, string>> = { vi, en };

type Ctx = {
  locale: Locale;
  setLocale: (locale: Locale) => void;
  t: (key: TranslationKey, vars?: Record<string, string | number>) => string;
};

const LocaleContext = React.createContext<Ctx | null>(null);

function readCookie(): Locale | null {
  if (typeof document === 'undefined') return null;
  const match = document.cookie.match(new RegExp(`(?:^|; )${LOCALE_COOKIE}=(vi|en)`));
  return (match?.[1] as Locale) ?? null;
}

// The cookie is external state, so the locale is read through
// useSyncExternalStore: the server snapshot is the default locale (keeping the
// first paint identical on both sides) and the client snapshot is the cookie.
const listeners = new Set<() => void>();

function subscribe(listener: () => void) {
  listeners.add(listener);
  return () => listeners.delete(listener);
}

function getClientSnapshot(): Locale {
  return readCookie() ?? DEFAULT_LOCALE;
}

export function LocaleProvider({
  children,
  initialLocale = DEFAULT_LOCALE,
}: {
  children: React.ReactNode;
  /** Read from the cookie on the server so the first paint is already correct. */
  initialLocale?: Locale;
}) {
  const getServerSnapshotForRequest = React.useCallback(() => initialLocale, [initialLocale]);
  const locale = React.useSyncExternalStore(subscribe, getClientSnapshot, getServerSnapshotForRequest);

  React.useEffect(() => {
    // The server already set <html lang>; this follows a client-side switch.
    document.documentElement.lang = locale;
  }, [locale]);

  const setLocale = React.useCallback((next: Locale) => {
    document.cookie = `${LOCALE_COOKIE}=${next}; path=/; max-age=31536000; samesite=lax`;
    listeners.forEach((listener) => listener());
  }, []);

  const t = React.useCallback(
    (key: TranslationKey, vars?: Record<string, string | number>) => {
      const template = dictionaries[locale][key] ?? dictionaries[DEFAULT_LOCALE][key] ?? key;
      if (!vars) return template;
      return template.replace(/\{(\w+)\}/g, (match, name: string) =>
        name in vars ? String(vars[name]) : match,
      );
    },
    [locale],
  );

  const value = React.useMemo(() => ({ locale, setLocale, t }), [locale, setLocale, t]);
  return <LocaleContext.Provider value={value}>{children}</LocaleContext.Provider>;
}

function useLocaleContext(): Ctx {
  const ctx = React.useContext(LocaleContext);
  if (!ctx) throw new Error('useT must be used inside <LocaleProvider>');
  return ctx;
}

export function useT() {
  return useLocaleContext().t;
}

export function useLocale() {
  const { locale, setLocale } = useLocaleContext();
  return { locale, setLocale };
}

export type { TranslationKey };

/**
 * Maps a backend error to a localised message. The server sends stable codes
 * and a Vietnamese fallback; the UI language is decided here, on the client.
 */
export function useErrorMessage() {
  const t = useT();
  return React.useCallback(
    (err: unknown): string => {
      if (!err) return '';
      const code = (err as { code?: string }).code;
      const key = `errors.${code}` as TranslationKey;
      if (code && key in vi) return t(key);
      const message = (err as { message?: string }).message;
      return message || t('common.error');
    },
    [t],
  );
}
