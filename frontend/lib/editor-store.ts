/**
 * Editor state (§94): local, not server state. History is an explicit stack of
 * Design JSON snapshots rather than a reliance on browser undo (§31).
 */
import { create } from 'zustand';

import type { Design } from '@/types';

const HISTORY_LIMIT = 60;

export type SaveState = 'idle' | 'dirty' | 'saving' | 'saved' | 'conflict' | 'error';

type EditorState = {
  design: Design | null;
  version: number;
  activeSlideId: string | null;
  selectedElementId: string | null;
  zoom: number;
  saveState: SaveState;
  past: Design[];
  future: Design[];

  init: (design: Design, version: number) => void;
  /** Replaces the design and pushes the previous one onto the undo stack. */
  commit: (next: Design) => void;
  /** Replaces the design without touching history (server reconciliation). */
  reconcile: (design: Design, version: number) => void;
  setActiveSlide: (slideId: string) => void;
  setSelected: (elementId: string | null) => void;
  setZoom: (zoom: number) => void;
  setSaveState: (s: SaveState) => void;
  setVersion: (v: number) => void;
  undo: () => void;
  redo: () => void;
  canUndo: () => boolean;
  canRedo: () => boolean;
};

const clone = (d: Design): Design => JSON.parse(JSON.stringify(d));

export const useEditorStore = create<EditorState>((set, get) => ({
  design: null,
  version: 0,
  activeSlideId: null,
  selectedElementId: null,
  zoom: 0.42,
  saveState: 'idle',
  past: [],
  future: [],

  init: (design, version) =>
    set({
      design,
      version,
      activeSlideId: design.slides[0]?.id ?? null,
      selectedElementId: null,
      past: [],
      future: [],
      saveState: 'idle',
    }),

  commit: (next) => {
    const { design, past } = get();
    if (!design) return;
    const trimmed = [...past, clone(design)].slice(-HISTORY_LIMIT);
    set({ design: next, past: trimmed, future: [], saveState: 'dirty' });
  },

  reconcile: (design, version) => set({ design, version, saveState: 'saved' }),

  setActiveSlide: (slideId) => set({ activeSlideId: slideId, selectedElementId: null }),
  setSelected: (elementId) => set({ selectedElementId: elementId }),
  setZoom: (zoom) => set({ zoom: Math.min(1.2, Math.max(0.15, zoom)) }),
  setSaveState: (saveState) => set({ saveState }),
  setVersion: (version) => set({ version }),

  undo: () => {
    const { past, design, future } = get();
    if (!design || past.length === 0) return;
    const previous = past[past.length - 1];
    set({
      design: previous,
      past: past.slice(0, -1),
      future: [clone(design), ...future].slice(0, HISTORY_LIMIT),
      saveState: 'dirty',
    });
  },

  redo: () => {
    const { future, design, past } = get();
    if (!design || future.length === 0) return;
    const [next, ...rest] = future;
    set({
      design: next,
      future: rest,
      past: [...past, clone(design)].slice(-HISTORY_LIMIT),
      saveState: 'dirty',
    });
  },

  canUndo: () => get().past.length > 0,
  canRedo: () => get().future.length > 0,
}));
