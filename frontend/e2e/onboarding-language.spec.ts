import { expect, test } from '@playwright/test';

import { API } from './fixtures';

/**
 * A project's content language decides what every carousel in it is written in,
 * so it is an explicit choice with no default: a wrong default would only
 * surface after the first generation.
 */
test.describe('project content language', () => {
  test.beforeEach(async ({ page, request }) => {
    const email = `e2e-lang-${Date.now()}@example.com`;
    await request.post(`${API}/api/v1/auth/dev-login`, { data: { email } });
    const state = await request.storageState();
    const session = state.cookies.find((c) => c.name === 'tks_session')!;
    await page.context().addCookies([
      { name: 'tks_session', value: session.value, domain: 'localhost', path: '/' },
    ]);
  });

  test('creating a project requires picking a language', async ({ page }) => {
    await page.goto('/onboarding');
    await page.locator('textarea').fill('Personal finance for people in their twenties');

    await page.getByRole('button', { name: /Tạo project|Create project/ }).click();

    // No language chosen: the form must refuse rather than silently defaulting.
    await expect(page.getByText(/Hãy chọn ngôn ngữ|Pick a content language/)).toBeVisible();
    await expect(page).toHaveURL(/\/onboarding/);
  });

  test('the chosen language reaches the created project', async ({ page }) => {
    await page.goto('/onboarding');
    await page.locator('textarea').fill('Personal finance for people in their twenties');
    await page.getByRole('button', { name: 'English', exact: true }).click();
    await page.getByRole('button', { name: /Tạo project|Create project/ }).click();

    await page.waitForURL(/\/projects\/[0-9a-f-]{36}$/, { timeout: 60_000 });
    const projectId = page.url().split('/').pop()!;

    const project = await (await page.request.get(`${API}/api/v1/projects/${projectId}`)).json();
    expect(project.language).toBe('en');
  });

  test('the login screen has no language switch', async ({ page }) => {
    await page.context().clearCookies();
    await page.goto('/login');
    await expect(page.getByRole('button', { name: 'EN', exact: true })).toHaveCount(0);
    await expect(page.getByRole('button', { name: 'VI', exact: true })).toHaveCount(0);
  });
});
