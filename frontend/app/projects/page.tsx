'use client';

/** Project dashboard (§106). */
import { useMutation, useQueryClient } from '@tanstack/react-query';
import Link from 'next/link';
import * as React from 'react';

import { AppShell } from '@/components/project/AppShell';
import { Button, Card, EmptyState, Spinner } from '@/components/ui/primitives';
import { api } from '@/lib/api/client';
import { useProjects } from '@/lib/api/hooks';
import { useLocale, useT } from '@/lib/i18n';
import { formatRelative } from '@/lib/utils';

export default function ProjectsPage() {
  const t = useT();
  const { locale } = useLocale();
  const { data, isLoading } = useProjects();
  const qc = useQueryClient();

  const remove = useMutation({
    mutationFn: (id: string) => api.deleteProject(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['projects'] }),
  });

  const projects = data?.projects ?? [];

  return (
    <AppShell>
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold tracking-tight">{t('projects.title')}</h1>
        <Link href="/onboarding">
          <Button>{t('projects.new')}</Button>
        </Link>
      </div>

      {isLoading ? (
        <div className="mt-16 flex justify-center">
          <Spinner className="h-6 w-6" />
        </div>
      ) : projects.length === 0 ? (
        <div className="mt-8">
          <EmptyState
            title={t('projects.emptyTitle')}
            description={t('projects.emptyBody')}
            action={
              <Link href="/onboarding">
                <Button size="lg">{t('projects.emptyAction')}</Button>
              </Link>
            }
          />
        </div>
      ) : (
        <div className="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {projects.map((p) => (
            <Card key={p.id} className="group flex flex-col p-5 transition hover:border-neutral-300">
              <Link href={`/projects/${p.id}`} className="flex-1">
                <h2 className="font-medium tracking-tight">{p.name}</h2>
                <p className="mt-1 line-clamp-2 text-sm text-neutral-500">{p.niche}</p>
                <p className="mt-4 text-xs text-neutral-400">
                  {t('projects.carouselCount', { count: p.carousel_count })} ·{' '}
                  {t('projects.editedAt', { when: formatRelative(p.updated_at, locale) })}
                </p>
              </Link>
              <div className="mt-4 flex gap-2 opacity-0 transition group-hover:opacity-100">
                <Link href={`/projects/${p.id}/settings`}>
                  <Button variant="ghost" size="sm">
                    {t('nav.settings')}
                  </Button>
                </Link>
                <Button
                  variant="ghost"
                  size="sm"
                  className="text-red-600 hover:bg-red-50"
                  onClick={() => {
                    if (confirm(t('projects.confirmDelete', { name: p.name }))) remove.mutate(p.id);
                  }}
                >
                  {t('project.delete')}
                </Button>
              </div>
            </Card>
          ))}
        </div>
      )}
    </AppShell>
  );
}
