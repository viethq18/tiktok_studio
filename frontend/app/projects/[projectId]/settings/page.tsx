'use client';

/** Project settings (§107): General, AI Strategy, Brand. */
import { useMutation, useQueryClient } from '@tanstack/react-query';
import Link from 'next/link';
import { useParams } from 'next/navigation';
import * as React from 'react';

import { AppShell } from '@/components/project/AppShell';
import { BrandPanel } from '@/components/project/BrandPanel';
import { Button, Card, ErrorNote, Input, Label, Spinner, Textarea } from '@/components/ui/primitives';
import { api } from '@/lib/api/client';
import { useContentLanguages, useProject } from '@/lib/api/hooks';
import { useErrorMessage, useT, type TranslationKey } from '@/lib/i18n';
import { cn } from '@/lib/utils';
import { DEFAULT_BRAND_DISPLAY, type Brand, type ProjectContext } from '@/types';

const TABS = ['general', 'strategy', 'brand'] as const;

export default function SettingsPage() {
  const { projectId } = useParams<{ projectId: string }>();
  const { data: project, isLoading } = useProject(projectId);
  const qc = useQueryClient();
  const t = useT();
  const errorMessage = useErrorMessage();
  const { data: languages } = useContentLanguages();
  const [tab, setTab] = React.useState<(typeof TABS)[number]>('general');

  const [name, setName] = React.useState('');
  const [niche, setNiche] = React.useState('');
  const [description, setDescription] = React.useState('');
  const [languageCode, setLanguageCode] = React.useState('');
  const [brand, setBrand] = React.useState<Brand>({ display: DEFAULT_BRAND_DISPLAY });
  const [context, setContext] = React.useState<ProjectContext | null>(null);
  const [saved, setSaved] = React.useState(false);
  const [loadedKey, setLoadedKey] = React.useState<string | null>(null);

  // Seed the editable copy once the project arrives, and again if its AI
  // context is regenerated — but never on top of the user's own edits.
  const projectKey = project ? `${project.id}:${project.context_version}` : null;
  if (project && projectKey !== loadedKey) {
    setLoadedKey(projectKey);
    setName(project.name);
    setNiche(project.niche);
    setDescription(project.description);
    setLanguageCode(project.language);
    setBrand({ ...project.brand, display: project.brand?.display ?? DEFAULT_BRAND_DISPLAY });
    setContext(project.context ?? null);
  }

  const save = useMutation({
    mutationFn: () =>
      api.updateProject(projectId, {
        name,
        niche,
        description,
        language: languageCode,
        brand,
        context: context ?? undefined,
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['project', projectId] });
      qc.invalidateQueries({ queryKey: ['projects'] });
      setSaved(true);
      setTimeout(() => setSaved(false), 2000);
    },
  });

  // Changing the brief changes which questions are worth asking, so the
  // generated form is rebuilt to match.
  const regenSchema = useMutation({
    mutationFn: () => api.regenerateSchema(projectId),
    onSuccess: (res) => qc.setQueryData(['schema', projectId], res),
  });

  if (isLoading || !project) {
    return (
      <AppShell>
        <div className="flex justify-center py-20">
          <Spinner className="h-6 w-6" />
        </div>
      </AppShell>
    );
  }

  const updateContext = (patch: Partial<ProjectContext>) =>
    setContext((c) => (c ? { ...c, ...patch } : c));

  return (
    <AppShell
      breadcrumb={
        <Link href={`/projects/${projectId}`} className="hover:text-neutral-900">
          {project.name}
        </Link>
      }
    >
      <h1 className="text-2xl font-semibold tracking-tight">{t('settings.title')}</h1>

      <div className="mt-6 flex gap-1 border-b border-neutral-200">
        {TABS.map((tabId) => (
          <button
            key={tabId}
            onClick={() => setTab(tabId)}
            className={cn(
              '-mb-px border-b-2 px-4 py-2 text-sm transition',
              tab === tabId
                ? 'border-neutral-900 font-medium text-neutral-900'
                : 'border-transparent text-neutral-500',
            )}
          >
            {t(`settings.tab.${tabId}` as TranslationKey)}
          </button>
        ))}
      </div>

      <Card className="mt-6 max-w-2xl space-y-5 p-6">
        {tab === 'general' ? (
          <>
            <div className="space-y-2">
              <Label htmlFor="name">{t('settings.name')}</Label>
              <Input id="name" value={name} onChange={(e) => setName(e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="niche">{t('settings.niche')}</Label>
              <Input id="niche" value={niche} onChange={(e) => setNiche(e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="description">{t('settings.description')}</Label>
              <Textarea
                id="description"
                rows={3}
                value={description}
                onChange={(e) => setDescription(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="language">{t('settings.language')}</Label>
              <select
                id="language"
                className="h-10 w-full rounded-lg border border-neutral-200 bg-white px-3 text-sm focus:border-neutral-900 focus:outline-none"
                value={languageCode}
                onChange={(e) => setLanguageCode(e.target.value)}
              >
                {(languages ?? []).map((lang) => (
                  <option key={lang.code} value={lang.code}>
                    {lang.native_name}
                  </option>
                ))}
              </select>
              <p className="text-xs text-neutral-500">{t('settings.languageHint')}</p>
            </div>
          </>
        ) : null}

        {tab === 'strategy' && context ? (
          <>
            <div className="space-y-2">
              <Label>{t('settings.audience')}</Label>
              <Textarea
                rows={2}
                value={context.audience?.description ?? ''}
                onChange={(e) =>
                  updateContext({ audience: { ...context.audience, description: e.target.value } })
                }
              />
            </div>
            <div className="space-y-2">
              <Label>{t('settings.demographics')}</Label>
              <Input
                value={context.audience?.demographics ?? ''}
                onChange={(e) =>
                  updateContext({ audience: { ...context.audience, demographics: e.target.value } })
                }
              />
            </div>
            <div className="space-y-2">
              <Label>{t('settings.tone')}</Label>
              <Input
                value={(context.tone ?? []).join(', ')}
                onChange={(e) => updateContext({ tone: splitList(e.target.value) })}
              />
            </div>
            <div className="space-y-2">
              <Label>{t('settings.writingStyle')}</Label>
              <Textarea
                rows={2}
                value={context.writing_style ?? ''}
                onChange={(e) => updateContext({ writing_style: e.target.value })}
              />
            </div>
            <div className="space-y-2">
              <Label>{t('settings.pillars')}</Label>
              <Input
                value={(context.content_pillars ?? []).map((p) => p.name).join(', ')}
                onChange={(e) =>
                  updateContext({
                    content_pillars: splitList(e.target.value).map((namePart) => ({
                      id: slugify(namePart),
                      name: namePart,
                    })),
                  })
                }
              />
            </div>
            <div className="space-y-2">
              <Label>{t('settings.ctaOptions')}</Label>
              <div className="space-y-2">
                {(context.cta_options ?? []).map((cta, i) => (
                  <div key={cta.id} className="flex items-center gap-3">
                    <span className="w-20 shrink-0 text-xs uppercase tracking-wide text-neutral-400">
                      {cta.id}
                    </span>
                    <Input
                      value={cta.label}
                      onChange={(e) => {
                        const next = [...context.cta_options];
                        next[i] = { ...cta, label: e.target.value };
                        updateContext({ cta_options: next });
                      }}
                    />
                  </div>
                ))}
              </div>
            </div>
            <div className="rounded-lg bg-neutral-50 p-4">
              <p className="text-sm text-neutral-600">{t('settings.regenerateNote')}</p>
              <Button
                variant="secondary"
                size="sm"
                className="mt-3"
                disabled={regenSchema.isPending}
                onClick={() => regenSchema.mutate()}
              >
                {regenSchema.isPending ? t('settings.regenerating') : t('settings.regenerate')}
              </Button>
            </div>
          </>
        ) : null}

        {tab === 'brand' ? <BrandPanel projectId={projectId} brand={brand} onChange={setBrand} /> : null}

        <div className="flex items-center gap-3 border-t border-neutral-100 pt-5">
          <Button disabled={save.isPending} onClick={() => save.mutate()}>
            {save.isPending ? t('common.saving') : t('common.save')}
          </Button>
          {saved ? <span className="text-sm text-emerald-600">{t('common.saved')}</span> : null}
        </div>
        <ErrorNote>{save.error ? errorMessage(save.error) : ''}</ErrorNote>
      </Card>
    </AppShell>
  );
}

function splitList(value: string): string[] {
  return value.split(',').map((s) => s.trim()).filter(Boolean);
}

function slugify(value: string): string {
  return value
    .toLowerCase()
    .normalize('NFD')
    .replace(/[̀-ͯ]/g, '')
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/(^-|-$)/g, '');
}
