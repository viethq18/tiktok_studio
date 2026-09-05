'use client';

/** Carousel dashboard for one project (§41). */
import { useMutation, useQueryClient } from '@tanstack/react-query';
import Link from 'next/link';
import { useParams } from 'next/navigation';
import * as React from 'react';

import { AppShell } from '@/components/project/AppShell';
import { Badge, Button, Card, EmptyState, Input, Spinner } from '@/components/ui/primitives';
import { api } from '@/lib/api/client';
import { useCarousels, useProject } from '@/lib/api/hooks';
import { useLocale, useT, type TranslationKey } from '@/lib/i18n';
import { cn, formatRelative } from '@/lib/utils';

const TABS = ['all', 'draft', 'ready', 'archived'] as const;

export default function ProjectPage() {
  const { projectId } = useParams<{ projectId: string }>();
  const t = useT();
  const { locale } = useLocale();
  const { data: project } = useProject(projectId);
  const [status, setStatus] = React.useState('all');
  const [query, setQuery] = React.useState('');
  const [debounced, setDebounced] = React.useState('');
  const qc = useQueryClient();

  React.useEffect(() => {
    const t = setTimeout(() => setDebounced(query), 250);
    return () => clearTimeout(t);
  }, [query]);

  const { data, isLoading } = useCarousels(projectId, { status, q: debounced });
  const carousels = data?.carousels ?? [];

  // While anything is generating, keep the grid fresh.
  const generating = carousels.some((c) => c.status === 'generating');
  React.useEffect(() => {
    if (!generating) return;
    const t = setInterval(() => qc.invalidateQueries({ queryKey: ['carousels', projectId] }), 3000);
    return () => clearInterval(t);
  }, [generating, projectId, qc]);

  const invalidate = () => qc.invalidateQueries({ queryKey: ['carousels', projectId] });
  const duplicate = useMutation({ mutationFn: api.duplicateCarousel, onSuccess: invalidate });
  const archive = useMutation({ mutationFn: api.archiveCarousel, onSuccess: invalidate });
  const remove = useMutation({ mutationFn: api.deleteCarousel, onSuccess: invalidate });

  return (
    <AppShell breadcrumb={project ? <Link href="/projects">Projects</Link> : null}>
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">{project?.name ?? '…'}</h1>
          <p className="mt-1 text-sm text-neutral-500">{project?.niche}</p>
        </div>
        <div className="flex gap-2">
          <Link href={`/projects/${projectId}/settings`}>
            <Button variant="secondary">{t('nav.settings')}</Button>
          </Link>
          <Link href={`/projects/${projectId}/new`}>
            <Button>{t('project.newCarousel')}</Button>
          </Link>
        </div>
      </div>

      <div className="mt-6 flex flex-wrap items-center gap-3">
        <Input
          className="max-w-xs"
          placeholder={t('project.search')}
          value={query}
          onChange={(e) => setQuery(e.target.value)}
        />
        <div className="flex gap-1">
          {TABS.map((tabId) => (
            <button
              key={tabId}
              onClick={() => setStatus(tabId)}
              className={cn(
                'rounded-lg px-3 py-1.5 text-sm transition',
                status === tabId ? 'bg-neutral-900 text-white' : 'text-neutral-600 hover:bg-neutral-100',
              )}
            >
              {t(`project.tab.${tabId}` as TranslationKey)}
            </button>
          ))}
        </div>
      </div>

      {isLoading ? (
        <div className="mt-16 flex justify-center">
          <Spinner className="h-6 w-6" />
        </div>
      ) : carousels.length === 0 ? (
        <div className="mt-8">
          <EmptyState
            title={t('project.emptyTitle')}
            description={t('project.emptyBody')}
            action={
              <Link href={`/projects/${projectId}/new`}>
                <Button size="lg">{t('project.emptyAction')}</Button>
              </Link>
            }
          />
        </div>
      ) : (
        <div className="mt-6 grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
          {carousels.map((c) => (
            <Card key={c.id} className="group overflow-hidden">
              <Link href={`/carousels/${c.id}`} className="block">
                <div className="relative aspect-4/5 bg-neutral-100">
                  {c.thumbnail_url ? (
                    // eslint-disable-next-line @next/next/no-img-element
                    <img src={c.thumbnail_url} alt="" className="h-full w-full object-cover" />
                  ) : (
                    <div className="flex h-full items-center justify-center text-sm text-neutral-400">
                      {c.status === 'generating' ? <Spinner /> : t('project.noThumbnail')}
                    </div>
                  )}
                </div>
              </Link>
              <div className="p-4">
                <div className="flex items-start justify-between gap-2">
                  <Link href={`/carousels/${c.id}`} className="line-clamp-2 text-sm font-medium">
                    {c.title}
                  </Link>
                  <Badge status={c.status}>{t(`status.${c.status}` as TranslationKey)}</Badge>
                </div>
                <p className="mt-2 text-xs text-neutral-400">
                  {t('projects.editedAt', { when: formatRelative(c.updated_at, locale) })}
                </p>
                <div className="mt-3 flex gap-1 opacity-0 transition group-hover:opacity-100">
                  <Button variant="ghost" size="sm" onClick={() => duplicate.mutate(c.id)}>
                    {t('project.duplicate')}
                  </Button>
                  {c.status !== 'archived' ? (
                    <Button variant="ghost" size="sm" onClick={() => archive.mutate(c.id)}>
                      {t('project.archive')}
                    </Button>
                  ) : null}
                  <Button
                    variant="ghost"
                    size="sm"
                    className="text-red-600 hover:bg-red-50"
                    onClick={() => {
                      if (confirm(t('project.confirmDelete'))) remove.mutate(c.id);
                    }}
                  >
                    {t('project.delete')}
                  </Button>
                </div>
              </div>
            </Card>
          ))}
        </div>
      )}
    </AppShell>
  );
}
