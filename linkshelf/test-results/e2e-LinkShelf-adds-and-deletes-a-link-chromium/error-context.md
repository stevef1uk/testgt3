# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: e2e.spec.js >> LinkShelf >> adds and deletes a link
- Location: tests/e2e.spec.js:17:3

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator: locator('#links a').filter({ hasText: 'Playwright link 1793698437362' }).locator('..').getByRole('button', { name: 'Delete' })
Expected: visible
Timeout: 5000ms
Error: element(s) not found

Call log:
  - Expect "toBeVisible" with timeout 5000ms
  - waiting for locator('#links a').filter({ hasText: 'Playwright link 1793698437362' }).locator('..').getByRole('button', { name: 'Delete' })

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
    - list:
      - listitem:
        - link "Playwright link 1793698437362":
          - /url: https://example.com/playwright
        - text: https://example.com/playwright
      - listitem:
        - link "Playwright link 1793698431107":
          - /url: https://example.com/playwright
        - text: https://example.com/playwright
      - listitem:
        - link "Playwright link 1793698379632":
          - /url: https://example.com/playwright
        - text: https://example.com/playwright
      - listitem:
        - link "Playwright link 1793698213335":
          - /url: https://example.com/playwright
        - text: https://example.com/playwright
      - listitem:
        - link "Playwright link 1793698207043":
          - /url: https://example.com/playwright
        - text: https://example.com/playwright
      - listitem:
        - link "Playwright link 1793698175126":
          - /url: https://example.com/playwright
        - text: https://example.com/playwright
      - listitem:
        - link "Playwright link 1793698168867":
          - /url: https://example.com/playwright
        - text: https://example.com/playwright
      - listitem:
        - link "Playwright link 1793697915057":
          - /url: https://example.com/playwright
        - text: https://example.com/playwright
      - listitem:
        - link "Playwright link 1793697908764":
          - /url: https://example.com/playwright
        - text: https://example.com/playwright
      - listitem:
        - link "Playwright link 1793697902457":
          - /url: https://example.com/playwright
        - text: https://example.com/playwright
      - listitem:
        - link "Playwright link 1793697880821":
          - /url: https://example.com/playwright
        - text: https://example.com/playwright
      - listitem:
        - link "Playwright link 1793697860671":
          - /url: https://example.com/playwright
        - text: https://example.com/playwright
      - listitem:
        - link "Playwright link 1793697854380":
          - /url: https://example.com/playwright
        - text: https://example.com/playwright
      - listitem:
        - link "Playwright link 1793697804398":
          - /url: https://example.com/playwright
        - text: https://example.com/playwright
      - listitem:
        - link "Playwright link 1793697797551":
          - /url: https://example.com/playwright
        - text: https://example.com/playwright
      - listitem:
        - link "Playwright link 1793697637329":
          - /url: https://example.com/playwright
        - text: https://example.com/playwright
      - listitem:
        - link "Playwright link 1793697630945":
          - /url: https://example.com/playwright
        - text: https://example.com/playwright
      - listitem:
        - link "E2E persist 1793697402370":
          - /url: https://example.com/e2e-persist
        - text: https://example.com/e2e-persist
      - listitem:
        - link "E2E delete 1793697372013":
          - /url: https://example.com/e2e-delete
        - text: https://example.com/e2e-delete
      - listitem:
        - link "E2E add 1793697371875":
          - /url: https://example.com/e2e-add
        - text: https://example.com/e2e-add
      - listitem:
        - link "E2E persist 1793697370531":
          - /url: https://example.com/e2e-persist
        - text: https://example.com/e2e-persist
      - listitem:
        - link "E2E delete 1793697340191":
          - /url: https://example.com/e2e-delete
        - text: https://example.com/e2e-delete
      - listitem:
        - link "E2E add 1793697340064":
          - /url: https://example.com/e2e-add
        - text: https://example.com/e2e-add
      - listitem:
        - link "E2E persist 1793697233098":
          - /url: https://example.com/e2e-persist
        - text: https://example.com/e2e-persist
      - listitem:
        - link "E2E delete 1793697202745":
          - /url: https://example.com/e2e-delete
        - text: https://example.com/e2e-delete
      - listitem:
        - link "E2E add 1793697202626":
          - /url: https://example.com/e2e-add
        - text: https://example.com/e2e-add
      - listitem:
        - link "E2E persist 1793697193127":
          - /url: https://example.com/e2e-persist
        - text: https://example.com/e2e-persist
      - listitem:
        - link "E2E delete 1793697162780":
          - /url: https://example.com/e2e-delete
        - text: https://example.com/e2e-delete
      - listitem:
        - link "E2E add 1793697162653":
          - /url: https://example.com/e2e-add
        - text: https://example.com/e2e-add
      - listitem:
        - link "E2E delete 1793697001289":
          - /url: https://example.com/e2e-delete
        - text: https://example.com/e2e-delete
      - listitem:
        - link "E2E add 1793697000800":
          - /url: https://example.com/e2e-add
        - text: https://example.com/e2e-add
```

# Test source

```ts
  1  | const { test, expect } = require('@playwright/test');
  2  | 
  3  | test.describe('LinkShelf', () => {
  4  |   test('loads the page and serves static assets', async ({ page }) => {
  5  |     await page.goto('/');
  6  |     await expect(page.locator('#app')).toBeVisible();
  7  |     await expect(page).toHaveTitle('LinkShelf');
  8  |     await expect(page.locator('#links-heading')).toHaveText('Saved links');
  9  |     await expect(page.locator('#links')).toBeVisible();
  10 | 
  11 |     const stylesheet = await page.request.get('/static/style.css');
  12 |     expect(stylesheet.ok()).toBeTruthy();
  13 |     const script = await page.request.get('/static/app.js');
  14 |     expect(script.ok()).toBeTruthy();
  15 |   });
  16 | 
  17 |   test('adds and deletes a link', async ({ page }) => {
  18 |     await page.goto('/');
  19 |     const title = `Playwright link ${Date.now()}`;
  20 |     await page.locator('#title').fill(title);
  21 |     await page.locator('#url').fill('https://example.com/playwright');
  22 |     await page.getByRole('button', { name: 'Add' }).click();
  23 | 
  24 |     const link = page.locator('#links a', { hasText: title });
  25 |     await expect(link).toBeVisible();
  26 | 
  27 |     const item = link.locator('..');
  28 |     const deleteButton = item.getByRole('button', { name: 'Delete' });
> 29 |     await expect(deleteButton).toBeVisible();
     |                                ^ Error: expect(locator).toBeVisible() failed
  30 |     await deleteButton.click();
  31 |     await expect(link).not.toBeVisible();
  32 |   });
  33 | });
```