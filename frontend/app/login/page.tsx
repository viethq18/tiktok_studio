'use client';

/** Login (§105). Dev login only renders when the backend allows it. */
import { useQuery } from '@tanstack/react-query';
import { useRouter } from 'next/navigation';
import * as React from 'react';

import { Button, Card, ErrorNote, Input, Label, Spinner } from '@/components/ui/primitives';
import { api } from '@/lib/api/client';
import { useErrorMessage, useT } from '@/lib/i18n';

export default function LoginPage() {
  const router = useRouter();
  const t = useT();
  const errorMessage = useErrorMessage();
  const { data: config, isLoading } = useQuery({ queryKey: ['auth-config'], queryFn: api.authConfig });
  const [email, setEmail] = React.useState('');
  const [error, setError] = React.useState('');
  const [pending, setPending] = React.useState(false);

  async function devLogin(e: React.FormEvent) {
    e.preventDefault();
    setPending(true);
    setError('');
    try {
      await api.devLogin(email);
      router.replace('/projects');
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setPending(false);
    }
  }

  return (
    <main className="flex min-h-screen items-center justify-center px-4">
      <Card className="w-full max-w-md p-8">
        <h1 className="text-2xl font-semibold tracking-tight">{t('login.title')}</h1>
        <p className="mt-2 text-sm text-neutral-500">{t('login.subtitle')}</p>

        {isLoading ? (
          <div className="mt-8 flex justify-center">
            <Spinner />
          </div>
        ) : (
          <div className="mt-8 space-y-6">
            {config?.google_enabled ? (
              <Button size="lg" className="w-full" onClick={() => { window.location.href = api.googleLoginUrl(); }}>
                {t('login.google')}
              </Button>
            ) : null}

            {config?.dev_login_enabled ? (
              <form onSubmit={devLogin} className="space-y-3">
                {config.google_enabled ? (
                  <div className="flex items-center gap-3 text-xs text-neutral-400">
                    <span className="h-px flex-1 bg-neutral-200" />
                    {t('login.or')}
                    <span className="h-px flex-1 bg-neutral-200" />
                  </div>
                ) : null}
                <Label htmlFor="email">{t('login.devLogin')}</Label>
                <Input
                  id="email"
                  type="email"
                  required
                  placeholder={t('login.emailPlaceholder')}
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                />
                <Button type="submit" variant="secondary" className="w-full" disabled={pending}>
                  {pending ? t('login.submitting') : t('login.submit')}
                </Button>
              </form>
            ) : null}

            {!config?.google_enabled && !config?.dev_login_enabled ? (
              <ErrorNote>{t('login.noMethod')}</ErrorNote>
            ) : null}
            <ErrorNote>{error}</ErrorNote>
          </div>
        )}
      </Card>
    </main>
  );
}
