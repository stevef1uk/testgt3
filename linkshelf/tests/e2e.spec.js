const { test, expect } = require('@playwright/test');

test.describe('LinkShelf', () => {
  test('loads the UI', async ({ page }) => {
    await page.goto('/');
    await expect(page.locator('#title')).toBeVisible();
    await expect(page.locator('#url')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Add' })).toBeVisible();
    await expect(page.locator('#links')).toBeVisible();
  });

  test('adds a link', async ({ page }) => {
    await page.goto('/');

    const title = `E2E add ${Date.now()}`;
    await page.locator('#title').fill(title);
    await page.locator('#url').fill('https://example.com/e2e-add');
    await page.getByRole('button', { name: 'Add' }).click();

    await expect(page.locator('#links li', { hasText: title })).toBeVisible();
  });

  test('deletes a link', async ({ page }) => {
    await page.goto('/');

    const title = `E2E delete ${Date.now()}`;
    await page.locator('#title').fill(title);
    await page.locator('#url').fill('https://example.com/e2e-delete');
    await page.getByRole('button', { name: 'Add' }).click();

    const item = page.locator('#links li', { hasText: title });
    await expect(item).toBeVisible();
    await item.getByRole('button', { name: /delete/i }).click();
    await expect(item).toHaveCount(0);
  });

  test('serves static assets', async ({ page }) => {
    const css = await page.request.get('/static/style.css');
    expect(css.status()).toBe(200);

    const js = await page.request.get('/static/app.js');
    expect(js.status()).toBe(200);
  });
});