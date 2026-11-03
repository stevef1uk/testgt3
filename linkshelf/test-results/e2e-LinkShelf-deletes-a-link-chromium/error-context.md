# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: e2e.spec.js >> LinkShelf >> deletes a link
- Location: tests/e2e.spec.js:23:3

# Error details

```
Test timeout of 30000ms exceeded.
```

```
Error: locator.click: Test timeout of 30000ms exceeded.
Call log:
  - waiting for locator('#links li').filter({ hasText: 'E2E delete 1793697001289' }).getByRole('button', { name: /delete/i })

```

# Page snapshot

```yaml
- main [ref=e2]:
  - heading "LinkShelf" [level=1] [ref=e3]
  - generic [ref=e4]:
    - generic [ref=e5]:
      - text: Title
      - textbox "Title" [ref=e6]
    - generic [ref=e7]:
      - text: URL
      - textbox "URL" [ref=e8]:
        - /placeholder: https://example.com
    - button "Add" [active] [ref=e9] [cursor=pointer]
    - alert [ref=e10]
  - region [ref=e11]:
    - heading "Saved links" [level=2] [ref=e12]
    - list [ref=e13]:
      - listitem [ref=e14]:
        - link "E2E delete 1793697001289" [ref=e15] [cursor=pointer]:
          - /url: https://example.com/e2e-delete
        - generic [ref=e16]: https://example.com/e2e-delete
      - listitem [ref=e17]:
        - link "E2E add 1793697000800" [ref=e18] [cursor=pointer]:
          - /url: https://example.com/e2e-add
        - generic [ref=e19]: https://example.com/e2e-add
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
  9  |     await expect(page.locator('#links')).toBeVisible();
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
> 33 |     await item.getByRole('button', { name: /delete/i }).click();
     |                                                         ^ Error: locator.click: Test timeout of 30000ms exceeded.
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
```