'use client';

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import * as React from 'react';

import { LocaleProvider } from '@/lib/i18n';
import type { Locale } from '@/lib/i18n/locale';

export function Providers({ children, locale }: { children: React.ReactNode; locale: Locale }) {
  // Server state lives here; editor state lives in Zustand (§94).
  const [client] = React.useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: { staleTime: 15_000, retry: 1, refetchOnWindowFocus: false },
        },
      }),
  );
  return (
    <QueryClientProvider client={client}>
      <LocaleProvider initialLocale={locale}>{children}</LocaleProvider>
    </QueryClientProvider>
  );
}
