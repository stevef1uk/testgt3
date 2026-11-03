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

    const response = await page.request.get('/api/links');
    expect(response.ok()).toBeTruthy();
    const payload = await response.json();
    const links = Array.isArray(payload) ? payload : (payload.links || payload.items || []);
    const created = links.find((item) => item.title === title);
    expect(created).toBeTruthy();
    expect(created.id).toBeTruthy();

    const deleteResponse = await page.request.delete(`/api/links/${created.id}`);
    expect(deleteResponse.ok()).toBeTruthy();

    const remainingResponse = await page.request.get('/api/links');
    const remainingPayload = await remainingResponse.json();
    const remaining = Array.isArray(remainingPayload)
      ? remainingPayload
      : (remainingPayload.links || remainingPayload.items || []);
    expect(remaining.some((item) => item.id === created.id)).toBeFalsy();
  });
});