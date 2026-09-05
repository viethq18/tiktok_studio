'use client';

/** Onboarding (§6). One question, three examples, one button. */
import { useMutation } from '@tanstack/react-query';
import { useRouter } from 'next/navigation';
import * as React from 'react';

import { AppShell } from '@/components/project/AppShell';
import { Button, Card, ErrorNote, Label, Spinner, Textarea } from '@/components/ui/primitives';
import { api } from '@/lib/api/client';
import { useContentLanguages } from '@/lib/api/hooks';
import { useErrorMessage, useT } from '@/lib/i18n';
import { cn } from '@/lib/utils';

export default function OnboardingPage() {
  const router = useRouter();
  const t = useT();
  const errorMessage = useErrorMessage();
  const [niche, setNiche] = React.useState('');
  // No preselected language: a wrong default would silently produce every
  // carousel in the project in the wrong language.
  const [languageCode, setLanguageCode] = React.useState('');
  const [languageError, setLanguageError] = React.useState('');
  const examples = [t('onboarding.example1'), t('onboarding.example2'), t('onboarding.example3')];
  const { data: languages, isLoading: languagesLoading } = useContentLanguages();

  const create = useMutation({
    mutationFn: () => api.createProject(niche.trim(), languageCode),
    onSuccess: (project) => router.replace(`/projects/${project.id}`),
  });

  function submit() {
    if (!languageCode) {
      setLanguageError(t('onboarding.languageRequired'));
      return;
    }
    setLanguageError('');
    create.mutate();
  }

  return (
    <AppShell breadcrumb={t('onboarding.breadcrumb')}>
      <div className="mx-auto max-w-xl">
        <h1 className="text-2xl font-semibold tracking-tight">{t('onboarding.title')}</h1>
        <p className="mt-2 text-sm text-neutral-500">{t('onboarding.subtitle')}</p>

        <Card className="mt-6 p-6">
          <Textarea
            rows={4}
            autoFocus
            placeholder={t('onboarding.placeholder')}
            value={niche}
            onChange={(e) => setNiche(e.target.value)}
          />

          <div className="mt-4">
            <p className="text-xs font-medium uppercase tracking-wide text-neutral-400">{t('onboarding.examples')}</p>
            <div className="mt-2 flex flex-wrap gap-2">
              {examples.map((example) => (
                <button
                  key={example}
                  type="button"
                  onClick={() => setNiche(example)}
                  className="rounded-full border border-neutral-200 px-3 py-1.5 text-sm text-neutral-600 transition hover:bg-neutral-50"
                >
                  {example}
                </button>
              ))}
            </div>
          </div>

          <div className="mt-6 space-y-2">
            <Label>{t('onboarding.language')}</Label>
            <p className="text-xs text-neutral-500">{t('onboarding.languageHint')}</p>
            {languagesLoading ? (
              <Spinner />
            ) : (
              <div className="flex flex-wrap gap-2 pt-1">
                {(languages ?? []).map((lang) => (
                  <button
                    key={lang.code}
                    type="button"
                    onClick={() => {
                      setLanguageCode(lang.code);
                      setLanguageError('');
                    }}
                    className={cn(
                      'rounded-full border px-3.5 py-1.5 text-sm transition',
                      languageCode === lang.code
                        ? 'border-neutral-900 bg-neutral-900 text-white'
                        : 'border-neutral-200 text-neutral-700 hover:bg-neutral-50',
                    )}
                  >
                    {lang.native_name}
                  </button>
                ))}
              </div>
            )}
            {languageError ? <p className="text-xs text-red-600">{languageError}</p> : null}
          </div>

          <Button
            size="lg"
            className="mt-6 w-full"
            disabled={create.isPending || niche.trim().length < 8}
            onClick={submit}
          >
            {create.isPending ? t('onboarding.submitting') : t('onboarding.submit')}
          </Button>
          {create.isPending ? (
            <p className="mt-3 text-center text-xs text-neutral-500">
              {t('onboarding.working')}
            </p>
          ) : null}
          <div className="mt-3">
            <ErrorNote>
              {create.error ? errorMessage(create.error) : ''}
            </ErrorNote>
          </div>
        </Card>
      </div>
    </AppShell>
  );
}
