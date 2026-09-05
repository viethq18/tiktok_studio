'use client';

/**
 * Left panel (§28, §97): the slide's copy, its background image, and the post
 * caption. The MVP deliberately exposes little — the AI is meant to get it
 * 80–95% right, and this is the "fix the last few percent" surface (§29).
 */
import { Hash, Image as ImageIcon, Type } from 'lucide-react';
import * as React from 'react';

import { CaptionPanel } from '@/components/editor/CaptionPanel';
import { ImagePicker } from '@/components/editor/ImagePicker';
import { Button, Label, Textarea } from '@/components/ui/primitives';
import { MAX_BODY_CHARS, MAX_HEADLINE_CHARS, measureBlockHeight } from '@/lib/design/measure';
import { useEditorStore } from '@/lib/editor-store';
import { useT } from '@/lib/i18n';
import { cn } from '@/lib/utils';
import type { DesignElement } from '@/types';
import type { FabricCanvasAdapter } from '@/lib/fabric/adapter';

const FONTS = ['Inter', 'Roboto', 'Montserrat', 'Nunito', 'Playfair Display', 'Be Vietnam Pro'];
const TABS = [
  { id: 'content', icon: Type },
  { id: 'image', icon: ImageIcon },
  { id: 'caption', icon: Hash },
] as const;

export function EditorSidebar({
  carouselId,
  adapterRef,
}: {
  carouselId: string;
  adapterRef: React.MutableRefObject<FabricCanvasAdapter | null>;
}) {
  const t = useT();
  const [tab, setTab] = React.useState<(typeof TABS)[number]['id']>('content');
  const design = useEditorStore((s) => s.design);
  const activeSlideId = useEditorStore((s) => s.activeSlideId);
  const selectedId = useEditorStore((s) => s.selectedElementId);
  const setSelected = useEditorStore((s) => s.setSelected);
  const commit = useEditorStore((s) => s.commit);

  const slide = design?.slides.find((s) => s.id === activeSlideId);

  function patchElement(elementId: string, patch: Partial<DesignElement>) {
    const current = useEditorStore.getState().design;
    if (!current || !slide) return;
    commit({
      ...current,
      slides: current.slides.map((s) =>
        s.id !== slide.id
          ? s
          : { ...s, elements: s.elements.map((e) => (e.id === elementId ? { ...e, ...patch } : e)) },
      ),
    });
  }

  if (!design || !slide) return null;

  // Brand marks are managed in project settings, not per slide.
  const texts = slide.elements.filter((e) => e.type === 'text' && !e.role?.startsWith('brand_'));

  return (
    <aside className="flex w-80 shrink-0 flex-col border-r border-neutral-200 bg-white">
      <div className="flex border-b border-neutral-200">
        {TABS.map(({ id, icon: Icon }) => (
          <button
            key={id}
            onClick={() => setTab(id)}
            className={cn(
              'flex flex-1 items-center justify-center gap-1.5 py-3 text-xs transition',
              tab === id ? 'border-b-2 border-neutral-900 font-medium' : 'text-neutral-500 hover:bg-neutral-50',
            )}
          >
            <Icon className="h-4 w-4" />
            {t(`editor.tab.${id}`)}
          </button>
        ))}
      </div>

      <div className="flex-1 overflow-y-auto p-4 scrollbar-thin">
        {tab === 'content' ? (
          <div className="space-y-5">
            {texts.map((el) => {
              const limit = ['hook', 'headline', 'cta'].includes(el.role ?? '')
                ? MAX_HEADLINE_CHARS
                : MAX_BODY_CHARS;
              const length = (el.content ?? '').length;
              const height = el.style
                ? measureBlockHeight(
                    el.content ?? '', el.style.fontFamily, el.style.fontWeight,
                    el.style.fontSize, el.style.lineHeight, el.width,
                  )
                : 0;
              const overflows = el.y + height > design.canvas.safe_area.y + design.canvas.safe_area.height;

              return (
                <div
                  key={el.id}
                  className={cn(
                    'rounded-lg border p-3 transition',
                    selectedId === el.id ? 'border-neutral-900' : 'border-neutral-200',
                  )}
                  onFocus={() => {
                    setSelected(el.id);
                    adapterRef.current?.selectElement(el.id);
                  }}
                >
                  <div className="mb-2 flex items-center justify-between">
                    <Label className="text-xs uppercase tracking-wide text-neutral-400">{el.role}</Label>
                    <span className={cn('text-xs', length > limit ? 'text-red-600' : 'text-neutral-400')}>
                      {length}/{limit}
                    </span>
                  </div>

                  <Textarea
                    rows={el.role === 'body' ? 4 : 2}
                    value={el.content ?? ''}
                    onChange={(e) => {
                      patchElement(el.id, { content: e.target.value });
                      // The canvas only reloads on slide or imagery changes, so
                      // text has to be pushed through explicitly.
                      adapterRef.current?.updateTextContent(el.id, e.target.value);
                    }}
                  />

                  {overflows ? <p className="mt-2 text-xs text-amber-700">{t('editor.overflow')}</p> : null}

                  <div className="mt-3 grid grid-cols-2 gap-2">
                    <Field label={t('editor.font')}>
                      <select
                        className="mt-1 h-8 w-full rounded border border-neutral-200 px-2 text-xs"
                        value={el.style?.fontFamily}
                        onChange={(e) => {
                          patchElement(el.id, { style: { ...el.style!, fontFamily: e.target.value } });
                          adapterRef.current?.updateTextStyle(el.id, { fontFamily: e.target.value });
                        }}
                      >
                        {FONTS.map((f) => (
                          <option key={f}>{f}</option>
                        ))}
                      </select>
                    </Field>

                    <Field label={t('editor.fontSize')}>
                      <input
                        type="number"
                        min={18}
                        max={140}
                        className="mt-1 h-8 w-full rounded border border-neutral-200 px-2 text-xs"
                        value={el.style?.fontSize ?? 36}
                        onChange={(e) => {
                          const fontSize = Number(e.target.value);
                          if (!Number.isFinite(fontSize)) return;
                          patchElement(el.id, { style: { ...el.style!, fontSize } });
                          adapterRef.current?.updateTextStyle(el.id, { fontSize });
                        }}
                      />
                    </Field>

                    <Field label={t('editor.color')}>
                      <input
                        type="color"
                        className="mt-1 h-8 w-full cursor-pointer rounded border border-neutral-200"
                        value={el.style?.color ?? '#FFFFFF'}
                        onChange={(e) => {
                          const color = e.target.value.toUpperCase();
                          patchElement(el.id, { style: { ...el.style!, color } });
                          adapterRef.current?.updateTextStyle(el.id, { fill: color });
                        }}
                      />
                    </Field>

                    <Field label={t('editor.weight')}>
                      <select
                        className="mt-1 h-8 w-full rounded border border-neutral-200 px-2 text-xs"
                        value={el.style?.fontWeight ?? 400}
                        onChange={(e) => {
                          const fontWeight = Number(e.target.value);
                          patchElement(el.id, { style: { ...el.style!, fontWeight } });
                          adapterRef.current?.updateTextStyle(el.id, { fontWeight });
                        }}
                      >
                        <option value={400}>{t('editor.weight.regular')}</option>
                        <option value={700}>{t('editor.weight.bold')}</option>
                      </select>
                    </Field>

                    <Field label={t('editor.align')} className="col-span-2">
                      <select
                        className="mt-1 h-8 w-full rounded border border-neutral-200 px-2 text-xs"
                        value={el.style?.textAlign}
                        onChange={(e) => {
                          const textAlign = e.target.value as 'left' | 'center' | 'right';
                          patchElement(el.id, { style: { ...el.style!, textAlign } });
                          adapterRef.current?.updateTextStyle(el.id, { textAlign });
                        }}
                      >
                        <option value="left">{t('editor.align.left')}</option>
                        <option value="center">{t('editor.align.center')}</option>
                        <option value="right">{t('editor.align.right')}</option>
                      </select>
                    </Field>
                  </div>
                </div>
              );
            })}
            {texts.length === 0 ? <p className="text-sm text-neutral-500">{t('editor.noText')}</p> : null}
          </div>
        ) : tab === 'image' ? (
          <ImagePicker carouselId={carouselId} slide={slide} />
        ) : (
          <CaptionPanel carouselId={carouselId} />
        )}
      </div>
    </aside>
  );
}

function Field({
  label,
  className,
  children,
}: {
  label: string;
  className?: string;
  children: React.ReactNode;
}) {
  return (
    <div className={className}>
      <Label className="text-xs text-neutral-500">{label}</Label>
      {children}
    </div>
  );
}

export function ZoomControls() {
  const t = useT();
  const zoom = useEditorStore((s) => s.zoom);
  const setZoom = useEditorStore((s) => s.setZoom);
  return (
    <div className="flex items-center gap-1 rounded-lg border border-neutral-200 bg-white px-1">
      <Button variant="ghost" size="sm" onClick={() => setZoom(zoom - 0.06)} aria-label={t('editor.zoomOut')}>
        −
      </Button>
      <span className="w-12 text-center text-xs tabular-nums text-neutral-600">{Math.round(zoom * 100)}%</span>
      <Button variant="ghost" size="sm" onClick={() => setZoom(zoom + 0.06)} aria-label={t('editor.zoomIn')}>
        +
      </Button>
    </div>
  );
}
