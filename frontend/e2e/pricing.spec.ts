import { expect, test } from '@playwright/test';

/**
 * Billing is not implemented, so these tests pin two things: the arithmetic the
 * page shows, and the fact that it never offers to take money.
 */
test.describe('pricing', () => {
  test('monthly prices are $0 / $20 / $100', async ({ page }) => {
    await page.goto('/pricing');
    await page.getByRole('button', { name: /Theo tháng|Monthly/ }).click();

    for (const amount of ['$0', '$20', '$100']) {
      await expect(page.getByText(amount, { exact: true }).first()).toBeVisible();
    }
  });

  test('yearly billing takes 20% off and shows the annual total', async ({ page }) => {
    await page.goto('/pricing');
    await page.getByRole('button', { name: /Theo năm|Yearly/ }).click();

    // 20 → 16 and 100 → 80 per month.
    await expect(page.getByText('$16', { exact: true }).first()).toBeVisible();
    await expect(page.getByText('$80', { exact: true }).first()).toBeVisible();
    // 16 × 12 = 192, 80 × 12 = 960.
    await expect(page.getByText(/\$192/).first()).toBeVisible();
    await expect(page.getByText(/\$960/).first()).toBeVisible();
  });

  test('paid plans cannot be purchased and the page says why', async ({ page }) => {
    await page.goto('/pricing');
    await expect(page.getByText(/Thanh toán chưa được kích hoạt|Billing is not switched on/)).toBeVisible();

    const paidButtons = page.getByRole('button', { name: /Thanh toán chưa mở|Billing not live/ });
    await expect(paidButtons).toHaveCount(2);
    for (const button of await paidButtons.all()) {
      await expect(button).toBeDisabled();
    }
  });

  test('the landing page links to pricing from the nav and the hero', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByRole('link', { name: /Bảng giá|^Pricing$/ }).first()).toBeVisible();

    await page.getByRole('link', { name: /Xem bảng giá|See pricing/ }).click();
    await expect(page).toHaveURL(/\/pricing$/);
  });

  test('pricing renders in English too', async ({ page, context }) => {
    await context.addCookies([{ name: 'tks_locale', value: 'en', domain: 'localhost', path: '/' }]);
    await page.goto('/pricing');
    await expect(page.getByRole('heading', { name: 'Pricing', level: 1 })).toBeVisible();
    await expect(page.getByText('Save 20%')).toBeVisible();
  });
});
