import { expect, test } from '@playwright/test';

import { canvasPixels, countColor, objectsForElement, seedCarousel, signIn } from './fixtures';

test.describe('editor text styling', () => {
  let carouselId: string;
  let sessionCookie: string;

  test.beforeAll(async ({ request }) => {
    ({ carouselId, sessionCookie } = await seedCarousel(request));
  });

  test.beforeEach(async ({ page }) => {
    await signIn(page, sessionCookie);
    await page.goto(`/carousels/${carouselId}`);
    await expect(page.locator('canvas.lower-canvas')).toBeVisible();
    // Wait for the background image and text to have been drawn.
    await page.waitForTimeout(2500);
  });

  test('exactly one canvas object per design element', async ({ page }) => {
    const counts = await objectsForElement(page);
    const duplicated = Object.entries(counts).filter(([, n]) => n > 1);
    expect(duplicated, `duplicate objects on canvas: ${JSON.stringify(duplicated)}`).toEqual([]);
  });

  test('changing colour repaints the canvas', async ({ page }) => {
    const before = await countColor(page, '#FF0000');

    // The colour input is native; set it and dispatch the change React listens for.
    const colorInput = page.locator('aside input[type="color"]').first();
    await colorInput.evaluate((el: HTMLInputElement) => {
      const setter = Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value')!.set!;
      setter.call(el, '#ff0000');
      el.dispatchEvent(new Event('input', { bubbles: true }));
      el.dispatchEvent(new Event('change', { bubbles: true }));
    });
    await page.waitForTimeout(600);

    const after = await countColor(page, '#FF0000');
    expect(after, `red pixels went ${before} → ${after}; the canvas did not repaint`).toBeGreaterThan(
      before + 500,
    );
  });

  test('changing alignment repaints the canvas', async ({ page }) => {
    const before = await canvasPixels(page);

    const alignSelect = page.locator('aside select').filter({ hasText: 'Trái' }).first();
    await alignSelect.selectOption('right');
    await page.waitForTimeout(600);

    const after = await canvasPixels(page);
    expect(after, 'alignment change did not alter the rendered canvas').not.toBe(before);
  });

  test('changing font size repaints the canvas', async ({ page }) => {
    const before = await canvasPixels(page);

    const sizeInput = page.locator('aside input[type="number"]').first();
    await sizeInput.fill('120');
    await page.waitForTimeout(600);

    const after = await canvasPixels(page);
    expect(after, 'font-size change did not alter the rendered canvas').not.toBe(before);
  });

  test('editing text in the sidebar repaints the canvas', async ({ page }) => {
    const before = await canvasPixels(page);

    const textarea = page.locator('aside textarea').first();
    await textarea.fill('XXXXX YYYYY ZZZZZ');
    await page.waitForTimeout(600);

    const after = await canvasPixels(page);
    expect(after, 'text edit did not alter the rendered canvas').not.toBe(before);
  });
});
