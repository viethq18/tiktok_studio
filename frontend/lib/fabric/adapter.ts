/**
 * FabricCanvasAdapter (§98).
 *
 * Design JSON is the source of truth (§2.5). Nothing outside this file talks to
 * Fabric: the editor calls loadDesign / serializeDesign and works in logical
 * canvas coordinates (1080×1350), never viewport pixels (§99, §100).
 */
import * as fabric from 'fabric';

import type { Design, DesignElement, DesignSlide } from '@/types';

type ObjectMeta = { elementId: string; elementType: DesignElement['type'] };

export type AdapterCallbacks = {
  onChange?: () => void;
  onSelect?: (elementId: string | null) => void;
};

export class FabricCanvasAdapter {
  private canvas: fabric.Canvas;
  private slide: DesignSlide | null = null;
  private design: Design | null = null;
  private suppress = false;
  private disposed = false;
  /** Increments per loadDesign call so a superseded load can bail out. */
  private loadToken = 0;

  constructor(el: HTMLCanvasElement, private callbacks: AdapterCallbacks = {}) {
    this.canvas = new fabric.Canvas(el, {
      preserveObjectStacking: true,
      selection: true,
      backgroundColor: '#ffffff',
      controlsAboveOverlay: true,
    });

    const emitChange = () => {
      if (!this.suppress) this.callbacks.onChange?.();
    };
    if (process.env.NODE_ENV !== 'production') {
      (window as unknown as { __tksCanvas?: fabric.Canvas }).__tksCanvas = this.canvas;
    }

    this.canvas.on('object:modified', emitChange);
    this.canvas.on('text:changed', emitChange);
    this.canvas.on('selection:created', () => this.emitSelection());
    this.canvas.on('selection:updated', () => this.emitSelection());
    this.canvas.on('selection:cleared', () => this.callbacks.onSelect?.(null));
  }

  private emitSelection() {
    const active = this.canvas.getActiveObject();
    const meta = active ? ((active as unknown as { tksMeta?: ObjectMeta }).tksMeta ?? null) : null;
    this.callbacks.onSelect?.(meta?.elementId ?? null);
  }

  /**
   * Renders one slide. Zoom maps the logical canvas onto the viewport.
   *
   * Two properties matter here and both come from the same structure:
   *
   *  1. Loading is ASYNC (images are fetched), so two overlapping calls used to
   *     interleave clear/add and leave two objects per element on the canvas.
   *     The second copy painted over the first, and every later mutation —
   *     which resolves an element by id and so found the first copy — became
   *     invisible. A generation token makes a superseded load bail out.
   *  2. Objects are built BEFORE the canvas is cleared, so a reload never shows
   *     an empty canvas while images are in flight.
   */
  async loadDesign(design: Design, slideId: string, zoom: number) {
    const token = ++this.loadToken;
    const slide = design.slides.find((s) => s.id === slideId) ?? design.slides[0];
    if (!slide) return;

    const ordered = [...slide.elements].sort((a, b) => a.z - b.z);
    const built: fabric.FabricObject[] = [];
    for (const el of ordered) {
      const obj = await this.build(el, design);
      // A newer load (or a disposal) happened while we awaited: drop this one.
      if (this.disposed || token !== this.loadToken) return;
      if (obj) built.push(obj);
    }
    if (this.disposed || token !== this.loadToken) return;

    this.design = design;
    this.slide = slide;
    this.suppress = true;
    this.canvas.clear();
    this.canvas.backgroundColor = slide.background || design.palette.background || '#ffffff';
    this.setZoom(zoom);
    for (const obj of built) this.canvas.add(obj);
    this.canvas.requestRenderAll();
    this.suppress = false;
  }

  setZoom(zoom: number) {
    if (!this.design) return;
    const { width, height } = this.design.canvas;
    this.canvas.setDimensions({ width: width * zoom, height: height * zoom });
    this.canvas.setZoom(zoom);
    this.canvas.requestRenderAll();
  }

  private async build(el: DesignElement, design: Design): Promise<fabric.FabricObject | null> {
    const meta: ObjectMeta = { elementId: el.id, elementType: el.type };
    const common = {
      left: el.x,
      top: el.y,
      opacity: el.opacity ?? 1,
      selectable: !el.locked,
      evented: !el.locked,
      lockMovementX: el.locked,
      lockMovementY: el.locked,
    };

    if (el.type === 'text') {
      const style = el.style!;
      const textbox = new fabric.Textbox(el.content ?? '', {
        ...common,
        width: el.width,
        fontSize: style.fontSize,
        fontFamily: style.fontFamily,
        fontWeight: style.fontWeight,
        fill: style.color,
        textAlign: style.textAlign,
        lineHeight: style.lineHeight,
        // Corner handles would let the creator scale the glyphs; the design
        // contract wants a resizable *box* with a stable font size.
        lockScalingY: true,
        splitByGrapheme: false,
        objectCaching: false,
      });
      (textbox as unknown as { tksMeta: ObjectMeta }).tksMeta = meta;
      textbox.setControlsVisibility({ mt: false, mb: false, tl: false, tr: false, bl: false, br: false });
      return textbox;
    }

    if (el.type === 'image') {
      const group: fabric.FabricObject[] = [];
      if (el.url) {
        try {
          const img = await fabric.FabricImage.fromURL(el.url, { crossOrigin: 'anonymous' });
          // "cover": fill the box, crop the overflow.
          const scale = Math.max(el.width / (img.width ?? 1), el.height / (img.height ?? 1));
          img.set({
            left: el.x + el.width / 2,
            top: el.y + el.height / 2,
            originX: 'center',
            originY: 'center',
            scaleX: scale,
            scaleY: scale,
          });
          group.push(img);
        } catch {
          // A dead presigned URL must not break the whole canvas.
        }
      }
      const backdrop = new fabric.Rect({
        left: el.x, top: el.y, width: el.width, height: el.height,
        fill: group.length ? 'transparent' : design.palette.primary || '#111111',
      });
      const overlay = el.overlay
        ? new fabric.Rect({
            left: el.x, top: el.y, width: el.width, height: el.height,
            fill: el.overlay.color, opacity: el.overlay.opacity,
          })
        : null;

      const parts = [backdrop, ...group, ...(overlay ? [overlay] : [])];
      const composed = new fabric.Group(parts, {
        ...common,
        // The background is positioned by the design, not dragged by hand.
        selectable: false,
        evented: false,
      });
      (composed as unknown as { tksMeta: ObjectMeta }).tksMeta = meta;
      return composed;
    }

    const rect = new fabric.Rect({
      ...common,
      width: el.width,
      height: el.height,
      fill: el.fill ?? design.palette.accent,
      rx: el.radius ?? 0,
      ry: el.radius ?? 0,
      selectable: false,
      evented: false,
    });
    (rect as unknown as { tksMeta: ObjectMeta }).tksMeta = meta;
    return rect;
  }

  /**
   * Reads Fabric state back into the slide, in logical coordinates. Only the
   * properties a creator can actually change are read back; everything else
   * stays as the design generator produced it.
   */
  serializeSlide(slide: DesignSlide): DesignSlide {
    const byId = new Map<string, fabric.FabricObject>();
    for (const obj of this.canvas.getObjects()) {
      const meta = (obj as unknown as { tksMeta?: ObjectMeta }).tksMeta;
      if (meta) byId.set(meta.elementId, obj);
    }

    return {
      ...slide,
      elements: slide.elements.map((el) => {
        const obj = byId.get(el.id);
        if (!obj) return el;
        if (el.type === 'text') {
          const tb = obj as fabric.Textbox;
          return {
            ...el,
            x: round(tb.left ?? el.x),
            y: round(tb.top ?? el.y),
            width: round((tb.width ?? el.width) * (tb.scaleX ?? 1)),
            height: round((tb.height ?? el.height) * (tb.scaleY ?? 1)),
            content: tb.text ?? el.content,
            opacity: tb.opacity ?? el.opacity,
          };
        }
        return el;
      }),
    };
  }

  selectElement(elementId: string | null) {
    if (!elementId) {
      this.canvas.discardActiveObject();
      this.canvas.requestRenderAll();
      return;
    }
    const target = this.find(elementId);
    if (target) {
      this.canvas.setActiveObject(target);
      this.canvas.requestRenderAll();
    }
  }

  /**
   * Applies a change from the properties panel without a full reload.
   *
   * These updates originate from the store, so they must NOT emit onChange:
   * doing so would round-trip the canvas back into the store and overwrite the
   * very change that is being applied.
   */
  updateTextStyle(
    elementId: string,
    patch: Partial<{ fontSize: number; fontFamily: string; fontWeight: number; fill: string; textAlign: string; lineHeight: number }>,
  ) {
    const target = this.find(elementId);
    if (!target) return;
    target.set(patch as Record<string, unknown>);
    this.canvas.requestRenderAll();
  }

  /** Mirrors sidebar text edits onto the canvas (§30). */
  updateTextContent(elementId: string, content: string) {
    const target = this.find(elementId);
    if (!(target instanceof fabric.Textbox)) return;
    target.set({ text: content });
    this.canvas.requestRenderAll();
  }

  private find(elementId: string): fabric.FabricObject | undefined {
    return this.canvas
      .getObjects()
      .find((o) => (o as unknown as { tksMeta?: ObjectMeta }).tksMeta?.elementId === elementId);
  }

  enterEditing(elementId: string) {
    const target = this.find(elementId);
    if (target instanceof fabric.Textbox) {
      this.canvas.setActiveObject(target);
      target.enterEditing();
      target.selectAll();
    }
  }

  currentSlideId(): string | null {
    return this.slide?.id ?? null;
  }

  dispose() {
    this.disposed = true;
    this.loadToken++;
    void this.canvas.dispose();
  }
}

function round(v: number): number {
  return Math.round(v * 100) / 100;
}
