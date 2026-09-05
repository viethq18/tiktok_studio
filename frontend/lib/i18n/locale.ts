/**
 * Locale primitives with no React and no 'use client', so both the server
 * layout and client components can use them.
 */
export type Locale = 'vi' | 'en';

export const LOCALES: { code: Locale; label: string; short: string }[] = [
  { code: 'vi', label: 'Tiếng Việt', short: 'VI' },
  { code: 'en', label: 'English', short: 'EN' },
];

export const LOCALE_COOKIE = 'tks_locale';
export const DEFAULT_LOCALE: Locale = 'vi';

export function parseLocale(value: string | undefined | null): Locale {
  return value === 'en' || value === 'vi' ? value : DEFAULT_LOCALE;
}
