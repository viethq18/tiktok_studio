'use client';

/** Carousel editor (§28, §137). */
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Download, Redo2, Sparkles, Undo2 } from 'lucide-react';
import Link from 'next/link';
import { useParams } from 'next/navigation';
import * as React from 'react';

import { CanvasWorkspace } from '@/components/editor/CanvasWorkspace';
import { EditorSidebar, ZoomControls } from '@/components/editor/EditorSidebar';
import { SlideNavigator } from '@/components/editor/SlideNavigator';
import { useAutosave, useSaveStateLabel } from '@/components/editor/useAutosave';
import { GenerationProgress } from '@/components/generation/GenerationProgress';
import { Button, Card, ErrorNote, Spinner } from '@/components/ui/primitives';
import { api } from '@/lib/api/client';
import { useErrorMessage, useT } from '@/lib/i18n';
import { useDesign, useJob, useRequireAuth } from '@/lib/api/hooks';
import { useEditorStore } from '@/lib/editor-store';
import type { FabricCanvasAdapter } from '@/lib/fabric/adapter';

export default function EditorPage() {
  const { carouselId } = useParams<{ carouselId: string }>();
  useRequireAuth();
  const t = useT();
  const qc = useQueryClient();
  const adapterRef = React.useRef<FabricCanvasAdapter | null>(null);

  const { data: carouselData } = useQuery({
    queryKey: ['carousel', carouselId],
    queryFn: () => api.getCarousel(carouselId),
    refetchInterval: (query) => (query.state.data?.carousel.status === 'generating' ? 2000 : false),
  });
  const carousel = carouselData?.carousel;
  const isGenerating = carousel?.status === 'generating';

  const { data: design, isLoading: designLoading } = useDesign(isGenerating ? '' : carouselId);
  const { data: job } = useJob(isGenerating ? (carouselData?.job?.id ?? null) : null);

  const init = useEditorStore((s) => s.init);
  const undo = useEditorStore((s) => s.undo);
  const redo = useEditorStore((s) => s.redo);
  const canUndo = useEditorStore((s) => s.past.length > 0);
  const canRedo = useEditorStore((s) => s.future.length > 0);
  const saveState = useAutosave(carouselId);
  const saveLabel = useSaveStateLabel(saveState);

  // Seed the editor once per carousel. Re-running init on every design payload
  // would reset the active slide and wipe undo history each time something
  // refreshed the design — for example after picking an image on slide 3.
  const initializedFor = React.useRef<string | null>(null);
  React.useEffect(() => {
    if (!design || initializedFor.current === carouselId) return;
    initializedFor.current = carouselId;
    init(design.design, design.version);
  }, [design, carouselId, init]);

  // Once generation finishes, pull the design in without a page reload.
  React.useEffect(() => {
    if (job?.status === 'completed') {
      qc.invalidateQueries({ queryKey: ['carousel', carouselId] });
      qc.invalidateQueries({ queryKey: ['design', carouselId] });
    }
  }, [job?.status, carouselId, qc]);

  React.useEffect(() => {
    function onKey(e: KeyboardEvent) {
      const target = e.target as HTMLElement | null;
      if (target && ['INPUT', 'TEXTAREA'].includes(target.tagName)) return;
      const mod = e.metaKey || e.ctrlKey;
      if (!mod || e.key.toLowerCase() !== 'z') return;
      e.preventDefault();
      if (e.shiftKey) redo();
      else undo();
    }
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [undo, redo]);

  if (isGenerating) {
    return (
      <div className="flex min-h-screen items-center justify-center px-6">
        <Card className="w-full max-w-md p-8">
          <GenerationProgress job={job ?? carouselData?.job} />
        </Card>
      </div>
    );
  }

  if (designLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <Spinner className="h-6 w-6" />
      </div>
    );
  }

  if (!design) {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center gap-4 px-6">
        <p className="text-sm text-neutral-600">{t('editor.noDesign')}</p>
        <RegenerateButton carouselId={carouselId} />
      </div>
    );
  }

  return (
    <div className="flex h-screen flex-col overflow-hidden">
      <header className="flex h-14 shrink-0 items-center gap-4 border-b border-neutral-200 bg-white px-4">
        <Link
          href={carousel ? `/projects/${carousel.project_id}` : '/projects'}
          className="text-sm font-semibold tracking-tight"
        >
          ← Carousel Studio
        </Link>
        <span className="min-w-0 flex-1 truncate text-sm text-neutral-500">{carousel?.title}</span>

        <span className="text-xs text-neutral-400">{saveLabel}</span>
        <ZoomControls />
        <div className="flex items-center gap-1">
          <Button variant="ghost" size="sm" disabled={!canUndo} onClick={undo} aria-label={t('editor.undo')}>
            <Undo2 className="h-4 w-4" />
          </Button>
          <Button variant="ghost" size="sm" disabled={!canRedo} onClick={redo} aria-label={t('editor.redo')}>
            <Redo2 className="h-4 w-4" />
          </Button>
        </div>
        <ApplyBrandButton carouselId={carouselId} />
        <ExportButton carouselId={carouselId} />
      </header>

      <div className="flex min-h-0 flex-1">
        <EditorSidebar carouselId={carouselId} adapterRef={adapterRef} />
        <div className="min-w-0 flex-1">
          <CanvasWorkspace adapterRef={adapterRef} />
        </div>
        <SlideNavigator />
      </div>
    </div>
  );
}

function RegenerateButton({ carouselId }: { carouselId: string }) {
  const t = useT();
  const qc = useQueryClient();
  const regenerate = useMutation({
    mutationFn: () => api.regenerateCarousel(carouselId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['carousel', carouselId] }),
  });
  return (
    <Button disabled={regenerate.isPending} onClick={() => regenerate.mutate()}>
      {regenerate.isPending ? t('editor.regenerating') : t('editor.regenerate')}
    </Button>
  );
}

/**
 * Re-stamps the project's brand kit onto this carousel. Brand marks are applied
 * when a carousel is generated, so a creator who sets up their brand afterwards
 * needs a way to apply it without regenerating the whole thing.
 */
function ApplyBrandButton({ carouselId }: { carouselId: string }) {
  const t = useT();
  const qc = useQueryClient();
  const apply = useMutation({
    mutationFn: () => api.applyBrand(carouselId),
    onSuccess: (view) => {
      useEditorStore.getState().reconcile(view.design, view.version);
      qc.setQueryData(['design', carouselId], view);
    },
  });
  return (
    <Button variant="secondary" size="sm" disabled={apply.isPending} onClick={() => apply.mutate()}>
      <Sparkles className="h-4 w-4" />
      {apply.isPending ? t('editor.applyingBrand') : t('editor.applyBrand')}
    </Button>
  );
}

/** Export (§47): create the job, poll it, hand back a signed download link. */
function ExportButton({ carouselId }: { carouselId: string }) {
  const t = useT();
  const errorMessage = useErrorMessage();
  const [exportId, setExportId] = React.useState<string | null>(null);

  const create = useMutation({
    mutationFn: () => api.createExport(carouselId),
    onSuccess: (job) => setExportId(job.id),
  });

  const { data: job } = useQuery({
    queryKey: ['export', exportId],
    queryFn: () => api.getExport(exportId!),
    enabled: Boolean(exportId),
    refetchInterval: (query) => {
      const status = query.state.data?.status;
      return status === 'completed' || status === 'failed' ? false : 1200;
    },
  });

  // Start the download exactly once per finished export. Tracking that in a ref
  // rather than state keeps this effect from feeding another render.
  const downloaded = React.useRef<string | null>(null);
  React.useEffect(() => {
    if (job?.status !== 'completed' || !job.download_url) return;
    if (downloaded.current === job.id) return;
    downloaded.current = job.id;
    window.location.href = job.download_url;
  }, [job?.id, job?.status, job?.download_url]);

  const settled = job?.status === 'completed' || job?.status === 'failed';
  const busy = create.isPending || (Boolean(exportId) && !settled);

  return (
    <div className="flex items-center gap-2">
      <Button size="sm" disabled={busy} onClick={() => create.mutate()}>
        <Download className="h-4 w-4" />
        {busy ? t('editor.exporting', { progress: job?.progress ?? 0 }) : t('editor.export')}
      </Button>
      {create.error ? (
        <span className="text-xs text-red-600">
          {errorMessage(create.error)}
        </span>
      ) : null}
      {job?.status === 'failed' ? <ErrorNote>{t('editor.exportFailed')}</ErrorNote> : null}
    </div>
  );
}
