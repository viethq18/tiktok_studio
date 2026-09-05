'use client';

import Link from 'next/link';
import * as React from 'react';

import { LocaleSwitch } from '@/components/ui/LocaleSwitch';
import { Button, Spinner } from '@/components/ui/primitives';
import { useLogout, useRequireAuth } from '@/lib/api/hooks';
import { useT } from '@/lib/i18n';

/** Chrome shared by every signed-in screen (§101). */
export function AppShell({ children, breadcrumb }: { children: React.ReactNode; breadcrumb?: React.ReactNode }) {
  const { data: user, isLoading } = useRequireAuth();
  const logout = useLogout();
  const t = useT();

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <Spinner className="h-6 w-6" />
      </div>
    );
  }
  if (!user) return null;

  return (
    <div className="min-h-screen">
      <header className="sticky top-0 z-20 border-b border-neutral-200 bg-white/90 backdrop-blur">
        <div className="mx-auto flex h-14 max-w-6xl items-center gap-4 px-6">
          <Link href="/projects" className="text-sm font-semibold tracking-tight">
            Carousel Studio
          </Link>
          <div className="min-w-0 flex-1 truncate text-sm text-neutral-500">{breadcrumb}</div>
          <LocaleSwitch />
          <span className="hidden text-sm text-neutral-500 sm:inline">{user.email}</span>
          <Button variant="ghost" size="sm" onClick={() => logout.mutate()}>
            {t('nav.signOut')}
          </Button>
        </div>
      </header>
      <main className="mx-auto max-w-6xl px-6 py-8">{children}</main>
    </div>
  );
}
