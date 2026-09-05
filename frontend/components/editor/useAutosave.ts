'use client';

/**
 * Autosave (§32): debounce, never save on every mouse move, and surface
 * "Saving… / Saved ✓". A 409 means another tab won the race, so we reload the
 * server's version rather than clobbering it (§80).
 */
import * as React from 'react';

import { api, ApiError } from '@/lib/api/client';
import { useEditorStore } from '@/lib/editor-store';
import { useT } from '@/lib/i18n';
import { useLatest } from '@/lib/use-latest';

const DEBOUNCE_MS = 1500;

export function useAutosave(carouselId: string) {
  const design = useEditorStore((s) => s.design);
  const version = useEditorStore((s) => s.version);
  const saveState = useEditorStore((s) => s.saveState);
  const setSaveState = useEditorStore((s) => s.setSaveState);
  const setVersion = useEditorStore((s) => s.setVersion);
  const reconcile = useEditorStore((s) => s.reconcile);

  const inFlight = React.useRef(false);
  // The debounce timer fires later, so it must read the newest design, not the
  // one captured when the timer was scheduled.
  const latest = useLatest({ design, version });

  React.useEffect(() => {
    if (saveState !== 'dirty' || !design) return;

    const timer = setTimeout(async () => {
      if (inFlight.current) {
        // Another save is running; leave the state dirty so this effect re-fires.
        return;
      }
      inFlight.current = true;
      setSaveState('saving');
      try {
        const current = latest.current;
        if (!current.design) return;
        const res = await api.saveDesign(carouselId, current.version, current.design);
        setVersion(res.version);
        setSaveState('saved');
      } catch (err) {
        if (err instanceof ApiError && err.status === 409) {
          const body = err.body as { current?: { version: number; design: typeof design } } | undefined;
          if (body?.current?.design) {
            reconcile(body.current.design, body.current.version);
            setSaveState('conflict');
          } else {
            setSaveState('conflict');
          }
        } else {
          setSaveState('error');
        }
      } finally {
        inFlight.current = false;
      }
    }, DEBOUNCE_MS);

    return () => clearTimeout(timer);
  }, [saveState, design, carouselId, latest, setSaveState, setVersion, reconcile]);

  return saveState;
}

/** Localised label for the autosave indicator. */
export function useSaveStateLabel(state: string): string {
  const t = useT();
  switch (state) {
    case 'saving':
      return t('editor.save.saving');
    case 'saved':
      return t('editor.save.saved');
    case 'dirty':
      return t('editor.save.dirty');
    case 'conflict':
      return t('editor.save.conflict');
    case 'error':
      return t('editor.save.error');
    default:
      return '';
  }
}
