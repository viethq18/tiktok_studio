'use client';

/** Slide rail (§28). */
import { cn } from '@/lib/utils';
import { useEditorStore } from '@/lib/editor-store';

export function SlideNavigator() {
  const design = useEditorStore((s) => s.design);
  const activeSlideId = useEditorStore((s) => s.activeSlideId);
  const setActiveSlide = useEditorStore((s) => s.setActiveSlide);

  if (!design) return null;

  return (
    <aside className="flex w-32 shrink-0 flex-col gap-3 overflow-y-auto border-l border-neutral-200 bg-white p-3 scrollbar-thin">
      {design.slides.map((slide) => {
        const headline = slide.elements.find((e) => e.type === 'text')?.content ?? '';
        return (
          <button
            key={slide.id}
            onClick={() => setActiveSlide(slide.id)}
            className={cn(
              'group overflow-hidden rounded-lg border-2 text-left transition',
              activeSlideId === slide.id ? 'border-neutral-900' : 'border-transparent hover:border-neutral-300',
            )}
          >
            <div
              className="relative flex aspect-4/5 items-center justify-center p-2"
              style={{ background: slide.background }}
            >
              <span className="line-clamp-4 text-[9px] leading-tight text-white mix-blend-difference">
                {headline}
              </span>
              <span className="absolute left-1 top-1 rounded bg-black/50 px-1 text-[9px] text-white">
                {String(slide.index).padStart(2, '0')}
              </span>
            </div>
          </button>
        );
      })}
    </aside>
  );
}
