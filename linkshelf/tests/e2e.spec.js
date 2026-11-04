const { test, expect } = require('@playwright/test');

test.describe('LinkShelf', () => {
  test('loads the UI and serves its JavaScript asset', async ({ page, request }) => {
    await page.goto('/');
    await expect(page.locator('#links')).toBeAttached();

    const asset = await request.get('/static/app.js');
    expect(asset.ok()).toBeTruthy();
    expect(await asset.text()).toContain('fetch');
  });

  test('adds a link and displays it in the links list', async ({ page }) => {
    await page.goto('/');

    const title = `Playwright link ${Date.now()}`;
    const url = 'https://example.com/e2e-add';

    await page.locator('input').nth(0).fill(title);
    await page.locator('input').nth(1).fill(url);
    await page.getByRole('button', { name: 'Add' }).click();

    const link = page.locator('#links').getByRole('link', { name: title });
    await expect(link).toBeVisible();
    await expect(link).toHaveAttribute('href', url);
  });

  test('deletes a link from the list', async ({ page }) => {
    await page.goto('/');

    const title = `Playwright delete ${Date.now()}`;
    await page.locator('input').nth(0).fill(title);
    await page.locator('input').nth(1).fill('https://example.com/e2e-delete');
    await page.getByRole('button', { name: /add/i }).click();

    const item = page.locator('#links').getByRole('listitem').filter({ hasText: title });
    await expect(item).toBeVisible();
    await item.getByRole('button', { name: /delete/i }).click();
    await expect(item).toHaveCount(0);
  });
});