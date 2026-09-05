'use client';

/** The Fabric canvas surface. All Fabric access goes through the adapter (§98). */
import * as React from 'react';

import { FabricCanvasAdapter } from '@/lib/fabric/adapter';
import { useEditorStore } from '@/lib/editor-store';
import { useLatest } from '@/lib/use-latest';

export function CanvasWorkspace({
  adapterRef,
}: {
  adapterRef: React.MutableRefObject<FabricCanvasAdapter | null>;
}) {
  const canvasEl = React.useRef<HTMLCanvasElement>(null);
  const design = useEditorStore((s) => s.design);
  const activeSlideId = useEditorStore((s) => s.activeSlideId);
  const zoom = useEditorStore((s) => s.zoom);
  const setSelected = useEditorStore((s) => s.setSelected);

  React.useEffect(() => {
    if (!canvasEl.current) return;
    const adapter = new FabricCanvasAdapter(canvasEl.current, {
      onSelect: (id) => setSelected(id),
      onChange: () => {
        // Read the store at call time, not through a captured snapshot. A ref
        // synced in an effect lags behind a state change made in the same tick,
        // and merging a stale slide back in silently reverted the edit that
        // triggered this callback.
        const { design: current, activeSlideId: slideId, commit } = useEditorStore.getState();
        if (!current || !slideId) return;
        const slide = current.slides.find((s) => s.id === slideId);
        if (!slide) return;
        const updated = adapter.serializeSlide(slide);
        commit({
          ...current,
          slides: current.slides.map((s) => (s.id === slideId ? updated : s)),
        });
      },
    });
    adapterRef.current = adapter;
    return () => {
      adapter.dispose();
      adapterRef.current = null;
    };
  }, [adapterRef, setSelected]);

  // A full canvas reload is expensive and steals focus, so it happens only when
  // the slide changes or its imagery changes — not on every keystroke, which
  // Fabric already reflects live.
  const slide = design?.slides.find((s) => s.id === activeSlideId);
  const imagerySignature = slide?.elements
    .filter((e) => e.type === 'image')
    .map((e) => `${e.asset_id ?? ''}:${e.url ?? ''}`)
    .join('|');

  const zoomRef = useLatest(zoom);
  React.useEffect(() => {
    const adapter = adapterRef.current;
    const { design: current, activeSlideId: slideId } = useEditorStore.getState();
    if (!adapter || !current || !slideId) return;
    void adapter.loadDesign(current, slideId, zoomRef.current);
  }, [activeSlideId, imagerySignature, adapterRef, zoomRef]);

  React.useEffect(() => {
    adapterRef.current?.setZoom(zoom);
  }, [zoom, adapterRef]);

  return (
    <div className="flex h-full items-center justify-center overflow-auto bg-neutral-100 p-8">
      <canvas ref={canvasEl} />
    </div>
  );
}
