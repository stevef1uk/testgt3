# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: e2e.spec.js >> LinkShelf >> adds a link and persists it after a fresh page load
- Location: tests/e2e.spec.js:13:3

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator: locator('#links').getByRole('link', { name: 'Playwright link 1793794843245' })
Expected: visible
Timeout: 5000ms
Error: element(s) not found

Call log:
  - Expect "toBeVisible" with timeout 5000ms
  - waiting for locator('#links').getByRole('link', { name: 'Playwright link 1793794843245' })

```

```yaml
- main:
  - heading "LinkShelf" [level=1]
  - paragraph: Save and manage your favorite links.
  - text: Title
  - textbox "Title"
  - text: URL
  - textbox "URL"
  - button "Add"
  - list
```

# Test source

```ts
  1  | const { test, expect } = require('@playwright/test');
  2  | 
  3  | test.describe('LinkShelf', () => {
  4  |   test('loads the UI and serves its JavaScript asset', async ({ page, request }) => {
  5  |     await page.goto('/');
  6  |     await expect(page.locator('#links')).toBeAttached();
  7  | 
  8  |     const asset = await request.get('/static/app.js');
  9  |     expect(asset.ok()).toBeTruthy();
  10 |     expect(await asset.text()).toContain('fetch');
  11 |   });
  12 | 
  13 |   test('adds a link and persists it after a fresh page load', async ({ page }) => {
  14 |     await page.goto('/');
  15 | 
  16 |     const title = `Playwright link ${Date.now()}`;
  17 |     const url = 'https://example.com/e2e-add';
  18 | 
  19 |     await page.locator('input').nth(0).fill(title);
  20 |     await page.locator('input').nth(1).fill(url);
  21 |     await page.getByRole('button', { name: 'Add' }).click();
  22 | 
  23 |     const link = page.locator('#links').getByRole('link', { name: title });
> 24 |     await expect(link).toBeVisible();
     |                        ^ Error: expect(locator).toBeVisible() failed
  25 |     await expect(link).toHaveAttribute('href', url);
  26 | 
  27 |     await page.reload();
  28 |     const persistedLink = page.locator('#links').getByRole('link', { name: title });
  29 |     await expect(persistedLink).toBeVisible();
  30 |     await expect(persistedLink).toHaveAttribute('href', url);
  31 |   });
  32 | 
  33 |   test('deletes a link from the list', async ({ page }) => {
  34 |     await page.goto('/');
  35 | 
  36 |     const title = `Playwright delete ${Date.now()}`;
  37 |     await page.locator('input').nth(0).fill(title);
  38 |     await page.locator('input').nth(1).fill('https://example.com/e2e-delete');
  39 |     await page.getByRole('button', { name: /add/i }).click();
  40 | 
  41 |     const item = page.locator('#links').getByRole('listitem').filter({ hasText: title });
  42 |     await expect(item).toBeVisible();
  43 |     await item.getByRole('button', { name: /delete/i }).click();
  44 |     await expect(item).toHaveCount(0);
  45 |   });
  46 | });
```