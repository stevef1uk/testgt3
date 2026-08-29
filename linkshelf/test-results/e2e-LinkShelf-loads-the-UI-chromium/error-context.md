# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: e2e.spec.js >> LinkShelf >> loads the UI
- Location: tests/e2e.spec.js:4:3

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator:  locator('#links')
Expected: visible
Received: hidden
Timeout:  5000ms

Call log:
  - Expect "toBeVisible" with timeout 5000ms
  - waiting for locator('#links')
    14 × locator resolved to <ul id="links"></ul>
       - unexpected value "hidden"

```

```yaml
- main:
  - heading "LinkShelf" [level=1]
  - text: Title
  - textbox "Title"
  - text: URL
  - textbox "URL":
    - /placeholder: https://example.com
  - button "Add"
  - alert
  - region "Saved links":
    - heading "Saved links" [level=2]
    - list
```

# Test source

```ts
  1  | const { test, expect } = require('@playwright/test');
  2  | 
  3  | test.describe('LinkShelf', () => {
  4  |   test('loads the UI', async ({ page }) => {
  5  |     await page.goto('/');
  6  |     await expect(page.locator('#title')).toBeVisible();
  7  |     await expect(page.locator('#url')).toBeVisible();
  8  |     await expect(page.getByRole('button', { name: 'Add' })).toBeVisible();
> 9  |     await expect(page.locator('#links')).toBeVisible();
     |                                          ^ Error: expect(locator).toBeVisible() failed
  10 |   });
  11 | 
  12 |   test('adds a link', async ({ page }) => {
  13 |     await page.goto('/');
  14 | 
  15 |     const title = `E2E add ${Date.now()}`;
  16 |     await page.locator('#title').fill(title);
  17 |     await page.locator('#url').fill('https://example.com/e2e-add');
  18 |     await page.getByRole('button', { name: 'Add' }).click();
  19 | 
  20 |     await expect(page.locator('#links li', { hasText: title })).toBeVisible();
  21 |   });
  22 | 
  23 |   test('deletes a link', async ({ page }) => {
  24 |     await page.goto('/');
  25 | 
  26 |     const title = `E2E delete ${Date.now()}`;
  27 |     await page.locator('#title').fill(title);
  28 |     await page.locator('#url').fill('https://example.com/e2e-delete');
  29 |     await page.getByRole('button', { name: 'Add' }).click();
  30 | 
  31 |     const item = page.locator('#links li', { hasText: title });
  32 |     await expect(item).toBeVisible();
  33 |     await item.getByRole('button', { name: /delete/i }).click();
  34 |     await expect(item).toHaveCount(0);
  35 |   });
  36 | 
  37 |   test('serves static assets', async ({ page }) => {
  38 |     const css = await page.request.get('/static/style.css');
  39 |     expect(css.status()).toBe(200);
  40 | 
  41 |     const js = await page.request.get('/static/app.js');
  42 |     expect(js.status()).toBe(200);
  43 |   });
  44 | });
  45 | 
```