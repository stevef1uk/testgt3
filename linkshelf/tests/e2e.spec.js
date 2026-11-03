const { test, expect } = require('@playwright/test');

test.describe('LinkShelf', () => {
  test('loads the page and serves static assets', async ({ page }) => {
    await page.goto('/');
    await expect(page.locator('#app')).toBeVisible();
    await expect(page).toHaveTitle('LinkShelf');
    await expect(page.locator('#links-heading')).toHaveText('Saved links');
    await expect(page.locator('#links')).toBeVisible();

    const stylesheet = await page.request.get('/static/style.css');
    expect(stylesheet.ok()).toBeTruthy();
    const script = await page.request.get('/static/app.js');
    expect(script.ok()).toBeTruthy();
  });

  test('adds and deletes a link', async ({ page }) => {
    await page.goto('/');
    const title = `Playwright link ${Date.now()}`;
    await page.locator('#title').fill(title);
    await page.locator('#url').fill('https://example.com/playwright');
    await page.getByRole('button', { name: 'Add' }).click();

    const link = page.locator('#links a', { hasText: title });
    await expect(link).toBeVisible();

    const item = link.locator('..');
    const deleteButton = item.getByRole('button', { name: 'Delete' });
    await expect(deleteButton).toBeVisible();
    await deleteButton.click();
    await expect(link).not.toBeVisible();
  });
});