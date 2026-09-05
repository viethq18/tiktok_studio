import type { APIRequestContext, Page } from '@playwright/test';

export const API = process.env.API_URL ?? 'http://localhost:8080';

/**
 * Builds a ready carousel through the API — far faster and less brittle than
 * clicking through onboarding, and the editor is what these tests are about.
 */
export type Seeded = { carouselId: string; sessionCookie: string };

export async function seedCarousel(request: APIRequestContext): Promise<Seeded> {
  const email = `e2e-${Date.now()}@example.com`;
  await request.post(`${API}/api/v1/auth/dev-login`, { data: { email } });

  const project = await (
    await request.post(`${API}/api/v1/projects`, {
      data: { niche: 'Chăm sóc trẻ sơ sinh 0-3 tuổi cho bố mẹ mới', language: 'vi' },
    })
  ).json();

  const gen = await (
    await request.post(`${API}/api/v1/projects/${project.id}/carousels/generate`, {
      data: {
        inputs: { topic: 'Ba dấu hiệu bé đang thiếu ngủ', slide_count: 4, cta: 'save' },
        ratio: '4:5',
      },
    })
  ).json();

  const state = await request.storageState();
  const session = state.cookies.find((c) => c.name === 'tks_session');
  if (!session) throw new Error('dev login did not set a session cookie');

  for (let i = 0; i < 60; i++) {
    const job = await (await request.get(`${API}/api/v1/jobs/${gen.job_id}`)).json();
    if (job.status === 'completed') {
      return { carouselId: gen.carousel_id as string, sessionCookie: session.value };
    }
    if (job.status === 'failed') throw new Error(`generation failed: ${job.error_message}`);
    await new Promise((r) => setTimeout(r, 1000));
  }
  throw new Error('generation timed out');
}

/**
 * Injects the API session into the browser. Playwright gives each test a fresh
 * request context, so the cookie value has to be carried explicitly rather than
 * copied from the current context's storage state.
 */
export async function signIn(page: Page, sessionCookie: string) {
  await page.context().addCookies([
    { name: 'tks_session', value: sessionCookie, domain: 'localhost', path: '/' },
  ]);
}

/** Fabric renders into canvas.lower-canvas; this is the visible output. */
export async function canvasPixels(page: Page): Promise<string> {
  return page.evaluate(() => {
    const el = document.querySelector('canvas.lower-canvas') as HTMLCanvasElement | null;
    if (!el) throw new Error('no fabric canvas on the page');
    return el.toDataURL('image/png');
  });
}

/** Counts pixels of a given colour, so a colour change is measurable. */
export async function countColor(page: Page, hex: string): Promise<number> {
  return page.evaluate((target) => {
    const el = document.querySelector('canvas.lower-canvas') as HTMLCanvasElement | null;
    if (!el) throw new Error('no fabric canvas on the page');
    const ctx = el.getContext('2d')!;
    const { data } = ctx.getImageData(0, 0, el.width, el.height);
    const r = parseInt(target.slice(1, 3), 16);
    const g = parseInt(target.slice(3, 5), 16);
    const b = parseInt(target.slice(5, 7), 16);
    let n = 0;
    for (let i = 0; i < data.length; i += 4) {
      if (Math.abs(data[i] - r) < 24 && Math.abs(data[i + 1] - g) < 24 && Math.abs(data[i + 2] - b) < 24) n++;
    }
    return n;
  }, hex);
}

/** How many Fabric objects claim a given design element id. */
export async function objectsForElement(page: Page): Promise<Record<string, number>> {
  return page.evaluate(() => {
    const counts: Record<string, number> = {};
    const canvas = (window as unknown as { __tksCanvas?: { getObjects(): unknown[] } }).__tksCanvas;
    if (!canvas) return counts;
    for (const obj of canvas.getObjects()) {
      const id = (obj as { tksMeta?: { elementId?: string } }).tksMeta?.elementId;
      if (id) counts[id] = (counts[id] ?? 0) + 1;
    }
    return counts;
  });
}
