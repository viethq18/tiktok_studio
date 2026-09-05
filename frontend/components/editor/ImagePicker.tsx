'use client';

/**
 * Image picker (§33, §38): candidates per slide, an editable search keyword,
 * "research more" for the next page, plus local upload. Selection is by
 * candidate id — the server holds the URL, so the browser never hands the
 * backend a URL to fetch (§84).
 */
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Search, Upload } from 'lucide-react';
import * as React from 'react';

import { Button, ErrorNote, Input, Label, Spinner } from '@/components/ui/primitives';
import { api } from '@/lib/api/client';
import { useErrorMessage, useT } from '@/lib/i18n';
import { useEditorStore } from '@/lib/editor-store';
import type { DesignSlide, ImageCandidate } from '@/types';

export function ImagePicker({ carouselId, slide }: { carouselId: string; slide: DesignSlide }) {
  const t = useT();
  const errorMessage = useErrorMessage();
  const qc = useQueryClient();
  const design = useEditorStore((s) => s.design);
  const commit = useEditorStore((s) => s.commit);
  const fileInput = React.useRef<HTMLInputElement>(null);

  const { data, isLoading, error } = useQuery({
    queryKey: ['slide-images', carouselId, slide.id],
    queryFn: () => api.slideImages(carouselId, slide.id),
  });

  // The keyword is editable and seeded from whatever the AI chose for this
  // slide; re-seeded when the user switches slides.
  const [keyword, setKeyword] = React.useState('');
  const [seededFor, setSeededFor] = React.useState<string | null>(null);
  const [results, setResults] = React.useState<ImageCandidate[] | null>(null);
  const [page, setPage] = React.useState(1);

  const suggested = slide.image_query || data?.query || '';
  if (seededFor !== slide.id && (suggested || data)) {
    setSeededFor(slide.id);
    setKeyword(suggested);
    setResults(null);
    setPage(1);
  }

  /** Persist the edited keyword so a later regeneration or revisit reuses it. */
  function rememberKeyword(next: string) {
    if (!design || next === slide.image_query) return;
    commit({
      ...design,
      slides: design.slides.map((s) => (s.id === slide.id ? { ...s, image_query: next } : s)),
    });
  }

  const search = useMutation({
    mutationFn: (opts: { query: string; page: number; append: boolean }) =>
      api.searchMoreImages(carouselId, slide.id, opts.query, opts.page).then((res) => ({ res, opts })),
    onSuccess: ({ res, opts }) => {
      setPage(res.page);
      setResults((prev) => (opts.append && prev ? [...prev, ...res.candidates] : res.candidates));
      rememberKeyword(opts.query);
    },
  });

  /** Applying an image produces a new design version on the server. */
  async function pullFreshDesign() {
    const fresh = await api.getDesign(carouselId);
    // Merge into the store without re-initialising: the active slide and the
    // undo history must survive picking an image.
    useEditorStore.getState().reconcile(fresh.design, fresh.version);
    qc.setQueryData(['design', carouselId], fresh);
  }

  const select = useMutation({
    mutationFn: (candidateId: string) =>
      api.selectImage(carouselId, slide.id, candidateId, useEditorStore.getState().version),
    onSuccess: pullFreshDesign,
  });

  const upload = useMutation({
    mutationFn: async (file: File) => {
      const asset = await api.uploadAsset(file, undefined, carouselId);
      const current = useEditorStore.getState();
      if (!current.design) return;
      const next = {
        ...current.design,
        slides: current.design.slides.map((s) =>
          s.id !== slide.id
            ? s
            : {
                ...s,
                elements: s.elements.map((e) =>
                  e.type === 'image' ? { ...e, asset_id: asset.id, url: asset.url } : e,
                ),
              },
        ),
      };
      await api.saveDesign(carouselId, current.version, next);
    },
    onSuccess: pullFreshDesign,
  });

  const candidates = results ?? data?.candidates ?? [];
  const busy = select.isPending || upload.isPending;
  const failure = select.error ?? upload.error ?? search.error ?? error;
  const currentAssetId = slide.elements.find((e) => e.type === 'image')?.asset_id;

  return (
    <div className="space-y-4">
      <div className="space-y-2">
        <Label htmlFor="image-keyword">{t('image.keyword', { index: slide.index })}</Label>
        <div className="flex gap-2">
          <Input
            id="image-keyword"
            value={keyword}
            placeholder={t('image.keywordPlaceholder')}
            onChange={(e) => setKeyword(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter' && keyword.trim()) {
                search.mutate({ query: keyword.trim(), page: 1, append: false });
              }
            }}
          />
          <Button
            size="sm"
            className="shrink-0"
            disabled={search.isPending || !keyword.trim()}
            onClick={() => search.mutate({ query: keyword.trim(), page: 1, append: false })}
            aria-label={t('image.search')}
          >
            <Search className="h-4 w-4" />
          </Button>
        </div>
        <p className="text-xs text-neutral-400">{t('image.keywordHint')}</p>
      </div>

      {isLoading ? (
        <div className="flex justify-center py-10">
          <Spinner />
        </div>
      ) : (
        <div className="grid grid-cols-2 gap-2">
          {candidates.map((c) => (
            <button
              key={c.id}
              disabled={busy}
              onClick={() => select.mutate(c.id)}
              className="group relative overflow-hidden rounded-lg border-2 border-transparent transition hover:border-neutral-400 disabled:opacity-50"
            >
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img src={c.thumb_url} alt={c.description ?? ''} className="aspect-4/5 w-full object-cover" />
              {c.photographer_name ? (
                <span className="absolute inset-x-0 bottom-0 truncate bg-black/50 px-1 py-0.5 text-[9px] text-white">
                  {c.photographer_name}
                </span>
              ) : null}
            </button>
          ))}
          {candidates.length === 0 ? (
            <p className="col-span-2 py-6 text-center text-sm text-neutral-500">{t('image.empty')}</p>
          ) : null}
        </div>
      )}

      <Button
        variant="secondary"
        size="sm"
        className="w-full"
        disabled={search.isPending || !keyword.trim()}
        onClick={() => search.mutate({ query: keyword.trim(), page: page + 1, append: true })}
      >
        {search.isPending ? t('image.searching') : t('image.searchMore')}
      </Button>

      <div className="border-t border-neutral-100 pt-4">
        <input
          ref={fileInput}
          type="file"
          accept="image/jpeg,image/png,image/webp"
          className="hidden"
          onChange={(e) => {
            const file = e.target.files?.[0];
            if (file) upload.mutate(file);
            e.target.value = '';
          }}
        />
        <Button
          variant="secondary"
          size="sm"
          className="w-full"
          disabled={busy}
          onClick={() => fileInput.current?.click()}
        >
          <Upload className="h-4 w-4" />
          {upload.isPending ? t('image.uploading') : t('image.upload')}
        </Button>
      </div>

      {currentAssetId ? (
        <p className="text-xs text-neutral-400">{t('image.current', { id: currentAssetId.slice(0, 8) })}</p>
      ) : null}

      <ErrorNote>
        {failure ? errorMessage(failure) : ''}
      </ErrorNote>
    </div>
  );
}
