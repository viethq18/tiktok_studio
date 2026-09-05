/**
 * Client-side mirror of backend/internal/fontkit.Wrap.
 *
 * The blueprint calls out (§157) that a mismatch between client and server text
 * measurement is the subtlest bug in the whole design pipeline: overflow
 * warnings in the editor would disagree with what the export actually produces.
 * Keeping the SAME greedy algorithm on both sides is what prevents that. The
 * server remains authoritative — this is only for immediate editor feedback.
 */

let ctx: CanvasRenderingContext2D | null = null;

function measureContext(): CanvasRenderingContext2D | null {
  if (typeof document === 'undefined') return null;
  if (!ctx) ctx = document.createElement('canvas').getContext('2d');
  return ctx;
}

export function measureWidth(text: string, family: string, weight: number, size: number): number {
  const c = measureContext();
  if (!c) return text.length * size * 0.5;
  c.font = `${weight} ${size}px "${family}", system-ui, sans-serif`;
  return c.measureText(text).width;
}

/** Greedy word wrap; words wider than the box are split by character. */
export function wrap(
  text: string,
  family: string,
  weight: number,
  size: number,
  maxWidth: number,
): string[] {
  const lines: string[] = [];
  for (const paragraph of text.split('\n')) {
    const words = paragraph.split(/\s+/).filter(Boolean);
    if (words.length === 0) {
      lines.push('');
      continue;
    }
    let current = '';
    for (const word of words) {
      const candidate = current ? `${current} ${word}` : word;
      if (measureWidth(candidate, family, weight, size) <= maxWidth || current === '') {
        if (measureWidth(candidate, family, weight, size) > maxWidth && current === '') {
          let chunk = '';
          for (const ch of word) {
            const next = chunk + ch;
            if (measureWidth(next, family, weight, size) > maxWidth && chunk !== '') {
              lines.push(chunk);
              chunk = ch;
            } else {
              chunk = next;
            }
          }
          current = chunk;
          continue;
        }
        current = candidate;
        continue;
      }
      lines.push(current);
      current = word;
    }
    if (current) lines.push(current);
  }
  return lines;
}

export function measureBlockHeight(
  text: string,
  family: string,
  weight: number,
  size: number,
  lineHeight: number,
  maxWidth: number,
): number {
  const lines = wrap(text, family, weight, size, maxWidth);
  return lines.length * size * (lineHeight > 0 ? lineHeight : 1.3);
}

/** Character budgets that match backend/internal/design (§122). */
export const MAX_HEADLINE_CHARS = 90;
export const MAX_BODY_CHARS = 220;

export function overflowsSafeArea(
  el: { y: number; width: number; content?: string; style?: { fontFamily: string; fontSize: number; fontWeight: number; lineHeight: number } },
  safeArea: { y: number; height: number },
): boolean {
  if (!el.content || !el.style) return false;
  const h = measureBlockHeight(
    el.content, el.style.fontFamily, el.style.fontWeight, el.style.fontSize, el.style.lineHeight, el.width,
  );
  return el.y + h > safeArea.y + safeArea.height;
}
