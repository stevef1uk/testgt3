# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: e2e.spec.js >> LinkShelf >> loads the UI and serves its JavaScript asset
- Location: tests/e2e.spec.js:4:3

# Error details

```
Error: expect(received).toContain(expected) // indexOf

Expected substring: "fetch"
Received string:    "document.addEventListener('DOMContentLoaded', function () {});"
```

# Page snapshot

```yaml
- main [ref=e2]:
  - heading "LinkShelf" [level=1] [ref=e3]
  - paragraph [ref=e4]: Save and manage your favorite links.
  - generic [ref=e5]:
    - generic [ref=e6]:
      - text: Title
      - textbox "Title" [ref=e7]
    - generic [ref=e8]:
      - text: URL
      - textbox "URL" [ref=e9]
    - button "Add" [ref=e10]
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
> 10 |     expect(await asset.text()).toContain('fetch');
     |                                ^ Error: expect(received).toContain(expected) // indexOf
  11 |   });
  12 | 
  13 |   test('adds a link and displays it in the links list', async ({ page }) => {
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
  24 |     await expect(link).toBeVisible();
  25 |     await expect(link).toHaveAttribute('href', url);
  26 |   });
  27 | 
  28 |   test('deletes a link from the list', async ({ page }) => {
  29 |     await page.goto('/');
  30 | 
  31 |     const title = `Playwright delete ${Date.now()}`;
  32 |     await page.locator('input').nth(0).fill(title);
  33 |     await page.locator('input').nth(1).fill('https://example.com/e2e-delete');
  34 |     await page.getByRole('button', { name: /add/i }).click();
  35 | 
  36 |     const item = page.locator('#links').getByRole('listitem').filter({ hasText: title });
  37 |     await expect(item).toBeVisible();
  38 |     await item.getByRole('button', { name: /delete/i }).click();
  39 |     await expect(item).toHaveCount(0);
  40 |   });
  41 | });
```