import { defineConfig } from '@playwright/test';

/**
 * Editor tests drive a real browser because the bugs that matter here are
 * rendering bugs: whether a style change actually reaches the pixels. Type
 * checks and unit tests cannot see that.
 *
 * The stack (API, worker, web) must already be running — `make e2e` starts it.
 */
export default defineConfig({
  testDir: './e2e',
  timeout: 90_000,
  expect: { timeout: 15_000 },
  fullyParallel: false,
  workers: 1,
  reporter: [['list']],
  use: {
    baseURL: process.env.WEB_URL ?? 'http://localhost:3000',
    viewport: { width: 1440, height: 900 },
    trace: 'retain-on-failure',
  },
});
