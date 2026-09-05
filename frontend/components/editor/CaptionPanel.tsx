'use client';

/**
 * Post caption + hashtags.
 *
 * A carousel is only half the post: what ships to TikTok is images *plus* a
 * caption. Putting it in the editor — beside the slides it describes — means a
 * creator finishes everything in one place and copies it out at the end. The
 * server records whether a caption was hand-edited, so a later regeneration of
 * the carousel never quietly overwrites the creator's own words.
 */
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Check, Copy, RefreshCw, X } from 'lucide-react';
import * as React from 'react';

import { Button, ErrorNote, Input, Label, Spinner, Textarea } from '@/components/ui/primitives';
import { api } from '@/lib/api/client';
import { useErrorMessage, useT } from '@/lib/i18n';
import { useLatest } from '@/lib/use-latest';
import type { Caption } from '@/types';

const AUTOSAVE_MS = 1200;

export function CaptionPanel({ carouselId }: { carouselId: string }) {
  const t = useT();
  const errorMessage = useErrorMessage();
  const qc = useQueryClient();

  const { data, isLoading, error } = useQuery({
    queryKey: ['caption', carouselId],
    queryFn: () => api.getCaption(carouselId),
    retry: false,
  });

  const [caption, setCaption] = React.useState('');
  const [hashtags, setHashtags] = React.useState<string[]>([]);
  const [draftTag, setDraftTag] = React.useState('');
  const [copied, setCopied] = React.useState(false);
  const [loadedAt, setLoadedAt] = React.useState<string | null>(null);
  const [dirty, setDirty] = React.useState(false);

  // Adopt the server copy on first load and after a regeneration, but never on
  // top of edits in progress.
  if (data && data.updated_at !== loadedAt) {
    setLoadedAt(data.updated_at);
    setCaption(data.caption);
    setHashtags(data.hashtags ?? []);
    setDirty(false);
  }

  const save = useMutation({
    mutationFn: (next: { caption: string; hashtags: string[] }) =>
      api.updateCaption(carouselId, next.caption, next.hashtags),
    onSuccess: (fresh: Caption) => {
      setLoadedAt(fresh.updated_at);
      qc.setQueryData(['caption', carouselId], fresh);
      setDirty(false);
    },
  });

  const regenerate = useMutation({
    mutationFn: () => api.regenerateCaption(carouselId),
    onSuccess: (fresh) => qc.setQueryData(['caption', carouselId], fresh),
  });

  // Debounced autosave, matching how the canvas behaves (§32). The timer fires
  // later, so it reads the newest draft rather than the one it was scheduled with.
  const pending = useLatest({ caption, hashtags });
  React.useEffect(() => {
    if (!dirty) return;
    const timer = setTimeout(() => save.mutate(pending.current), AUTOSAVE_MS);
    return () => clearTimeout(timer);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [dirty, caption, hashtags]);

  const fullText = [caption, hashtags.join(' ')].filter(Boolean).join('\n\n');

  async function copyAll() {
    try {
      await navigator.clipboard.writeText(fullText);
      setCopied(true);
      setTimeout(() => setCopied(false), 1800);
    } catch {
      // Clipboard can be blocked; the textarea is still selectable by hand.
    }
  }

  function addTag() {
    const raw = draftTag.trim();
    if (!raw) return;
    const tag = '#' + raw.replace(/^#+/, '').toLowerCase().replace(/\s+/g, '');
    if (tag.length > 1 && !hashtags.includes(tag)) {
      setHashtags((prev) => [...prev, tag]);
      setDirty(true);
    }
    setDraftTag('');
  }

  if (isLoading || regenerate.isPending) {
    return (
      <div className="flex flex-col items-center gap-3 py-12 text-sm text-neutral-500">
        <Spinner />
        {regenerate.isPending ? t('caption.regenerating') : t('caption.generating')}
      </div>
    );
  }

  if (error && !data) {
    return (
      <div className="space-y-3">
        <ErrorNote>{errorMessage(error)}</ErrorNote>
        <Button variant="secondary" size="sm" className="w-full" onClick={() => regenerate.mutate()}>
          {t('caption.regenerate')}
        </Button>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div>
        <p className="text-sm font-medium">{t('caption.title')}</p>
        <p className="mt-1 text-xs text-neutral-500">{t('caption.hint')}</p>
      </div>

      <div className="space-y-2">
        <div className="flex items-center justify-between">
          <Label htmlFor="caption">{t('caption.label')}</Label>
          <span className="text-xs text-neutral-400">{t('caption.charCount', { n: caption.length })}</span>
        </div>
        <Textarea
          id="caption"
          rows={7}
          value={caption}
          onChange={(e) => {
            setCaption(e.target.value);
            setDirty(true);
          }}
        />
      </div>

      <div className="space-y-2">
        <Label>{t('caption.hashtags')}</Label>
        <div className="flex flex-wrap gap-1.5">
          {hashtags.map((tag) => (
            <span
              key={tag}
              className="inline-flex items-center gap-1 rounded-full bg-neutral-100 py-1 pl-2.5 pr-1 text-xs text-neutral-700"
            >
              {tag}
              <button
                onClick={() => {
                  setHashtags((prev) => prev.filter((h) => h !== tag));
                  setDirty(true);
                }}
                className="rounded-full p-0.5 hover:bg-neutral-300"
                aria-label={`${t('common.remove')} ${tag}`}
              >
                <X className="h-3 w-3" />
              </button>
            </span>
          ))}
        </div>
        <Input
          value={draftTag}
          placeholder={t('caption.addHashtag')}
          onChange={(e) => setDraftTag(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === 'Enter' || e.key === ',') {
              e.preventDefault();
              addTag();
            }
          }}
          onBlur={addTag}
        />
        <p className="text-xs text-neutral-400">{t('caption.hashtagsHint')}</p>
      </div>

      <div className="space-y-2">
        <Button className="w-full" onClick={copyAll}>
          {copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
          {copied ? t('common.copied') : t('caption.copyAll')}
        </Button>
        <Button
          variant="secondary"
          size="sm"
          className="w-full"
          disabled={regenerate.isPending}
          onClick={() => regenerate.mutate()}
        >
          <RefreshCw className="h-4 w-4" />
          {t('caption.regenerate')}
        </Button>
      </div>

      <p className="text-xs text-neutral-400">
        {save.isPending ? t('common.saving') : data?.edited ? t('caption.edited') : ''}
      </p>
      <ErrorNote>{save.error ? errorMessage(save.error) : ''}</ErrorNote>
    </div>
  );
}
