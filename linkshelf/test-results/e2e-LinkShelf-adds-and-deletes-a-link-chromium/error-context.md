# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: e2e.spec.js >> LinkShelf >> adds and deletes a link
- Location: tests/e2e.spec.js:17:3

# Error details

```
Test timeout of 30000ms exceeded.
```

```
Error: page.waitForResponse: Test timeout of 30000ms exceeded.
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
        - link "Playwright link 1793698858417" [ref=e15] [cursor=pointer]:
          - /url: https://example.com/playwright
        - generic [ref=e16]: https://example.com/playwright
      - listitem [ref=e17]:
        - link "Playwright link 1793698827548" [ref=e18] [cursor=pointer]:
          - /url: https://example.com/playwright
        - generic [ref=e19]: https://example.com/playwright
      - listitem [ref=e20]:
        - link "Playwright link 1793698779734" [ref=e21] [cursor=pointer]:
          - /url: https://example.com/playwright
        - generic [ref=e22]: https://example.com/playwright
      - listitem [ref=e23]:
        - link "Playwright link 1793698716907" [ref=e24] [cursor=pointer]:
          - /url: https://example.com/playwright
        - generic [ref=e25]: https://example.com/playwright
      - listitem [ref=e26]:
        - link "Playwright link 1793698710599" [ref=e27] [cursor=pointer]:
          - /url: https://example.com/playwright
        - generic [ref=e28]: https://example.com/playwright
      - listitem [ref=e29]:
        - link "Playwright link 1793698686022" [ref=e30] [cursor=pointer]:
          - /url: https://example.com/playwright
        - generic [ref=e31]: https://example.com/playwright
      - listitem [ref=e32]:
        - link "Playwright link 1793698679719" [ref=e33] [cursor=pointer]:
          - /url: https://example.com/playwright
        - generic [ref=e34]: https://example.com/playwright
      - listitem [ref=e35]:
        - link "Playwright link 1793698467588" [ref=e36] [cursor=pointer]:
          - /url: https://example.com/playwright
        - generic [ref=e37]: https://example.com/playwright
      - listitem [ref=e38]:
        - link "Playwright link 1793698461295" [ref=e39] [cursor=pointer]:
          - /url: https://example.com/playwright
        - generic [ref=e40]: https://example.com/playwright
      - listitem [ref=e41]:
        - link "Playwright link 1793698455006" [ref=e42] [cursor=pointer]:
          - /url: https://example.com/playwright
        - generic [ref=e43]: https://example.com/playwright
      - listitem [ref=e44]:
        - link "Playwright link 1793698437362" [ref=e45] [cursor=pointer]:
          - /url: https://example.com/playwright
        - generic [ref=e46]: https://example.com/playwright
      - listitem [ref=e47]:
        - link "Playwright link 1793698431107" [ref=e48] [cursor=pointer]:
          - /url: https://example.com/playwright
        - generic [ref=e49]: https://example.com/playwright
      - listitem [ref=e50]:
        - link "Playwright link 1793698379632" [ref=e51] [cursor=pointer]:
          - /url: https://example.com/playwright
        - generic [ref=e52]: https://example.com/playwright
      - listitem [ref=e53]:
        - link "Playwright link 1793698213335" [ref=e54] [cursor=pointer]:
          - /url: https://example.com/playwright
        - generic [ref=e55]: https://example.com/playwright
      - listitem [ref=e56]:
        - link "Playwright link 1793698207043" [ref=e57] [cursor=pointer]:
          - /url: https://example.com/playwright
        - generic [ref=e58]: https://example.com/playwright
      - listitem [ref=e59]:
        - link "Playwright link 1793698175126" [ref=e60] [cursor=pointer]:
          - /url: https://example.com/playwright
        - generic [ref=e61]: https://example.com/playwright
      - listitem [ref=e62]:
        - link "Playwright link 1793698168867" [ref=e63] [cursor=pointer]:
          - /url: https://example.com/playwright
        - generic [ref=e64]: https://example.com/playwright
      - listitem [ref=e65]:
        - link "Playwright link 1793697915057" [ref=e66] [cursor=pointer]:
          - /url: https://example.com/playwright
        - generic [ref=e67]: https://example.com/playwright
      - listitem [ref=e68]:
        - link "Playwright link 1793697908764" [ref=e69] [cursor=pointer]:
          - /url: https://example.com/playwright
        - generic [ref=e70]: https://example.com/playwright
      - listitem [ref=e71]:
        - link "Playwright link 1793697902457" [ref=e72] [cursor=pointer]:
          - /url: https://example.com/playwright
        - generic [ref=e73]: https://example.com/playwright
      - listitem [ref=e74]:
        - link "Playwright link 1793697880821" [ref=e75] [cursor=pointer]:
          - /url: https://example.com/playwright
        - generic [ref=e76]: https://example.com/playwright
      - listitem [ref=e77]:
        - link "Playwright link 1793697860671" [ref=e78] [cursor=pointer]:
          - /url: https://example.com/playwright
        - generic [ref=e79]: https://example.com/playwright
      - listitem [ref=e80]:
        - link "Playwright link 1793697854380" [ref=e81] [cursor=pointer]:
          - /url: https://example.com/playwright
        - generic [ref=e82]: https://example.com/playwright
      - listitem [ref=e83]:
        - link "Playwright link 1793697804398" [ref=e84] [cursor=pointer]:
          - /url: https://example.com/playwright
        - generic [ref=e85]: https://example.com/playwright
      - listitem [ref=e86]:
        - link "Playwright link 1793697797551" [ref=e87] [cursor=pointer]:
          - /url: https://example.com/playwright
        - generic [ref=e88]: https://example.com/playwright
      - listitem [ref=e89]:
        - link "Playwright link 1793697637329" [ref=e90] [cursor=pointer]:
          - /url: https://example.com/playwright
        - generic [ref=e91]: https://example.com/playwright
      - listitem [ref=e92]:
        - link "Playwright link 1793697630945" [ref=e93] [cursor=pointer]:
          - /url: https://example.com/playwright
        - generic [ref=e94]: https://example.com/playwright
      - listitem [ref=e95]:
        - link "E2E persist 1793697402370" [ref=e96] [cursor=pointer]:
          - /url: https://example.com/e2e-persist
        - generic [ref=e97]: https://example.com/e2e-persist
      - listitem [ref=e98]:
        - link "E2E delete 1793697372013" [ref=e99] [cursor=pointer]:
          - /url: https://example.com/e2e-delete
        - generic [ref=e100]: https://example.com/e2e-delete
      - listitem [ref=e101]:
        - link "E2E add 1793697371875" [ref=e102] [cursor=pointer]:
          - /url: https://example.com/e2e-add
        - generic [ref=e103]: https://example.com/e2e-add
      - listitem [ref=e104]:
        - link "E2E persist 1793697370531" [ref=e105] [cursor=pointer]:
          - /url: https://example.com/e2e-persist
        - generic [ref=e106]: https://example.com/e2e-persist
      - listitem [ref=e107]:
        - link "E2E delete 1793697340191" [ref=e108] [cursor=pointer]:
          - /url: https://example.com/e2e-delete
        - generic [ref=e109]: https://example.com/e2e-delete
      - listitem [ref=e110]:
        - link "E2E add 1793697340064" [ref=e111] [cursor=pointer]:
          - /url: https://example.com/e2e-add
        - generic [ref=e112]: https://example.com/e2e-add
      - listitem [ref=e113]:
        - link "E2E persist 1793697233098" [ref=e114] [cursor=pointer]:
          - /url: https://example.com/e2e-persist
        - generic [ref=e115]: https://example.com/e2e-persist
      - listitem [ref=e116]:
        - link "E2E delete 1793697202745" [ref=e117] [cursor=pointer]:
          - /url: https://example.com/e2e-delete
        - generic [ref=e118]: https://example.com/e2e-delete
      - listitem [ref=e119]:
        - link "E2E add 1793697202626" [ref=e120] [cursor=pointer]:
          - /url: https://example.com/e2e-add
        - generic [ref=e121]: https://example.com/e2e-add
      - listitem [ref=e122]:
        - link "E2E persist 1793697193127" [ref=e123] [cursor=pointer]:
          - /url: https://example.com/e2e-persist
        - generic [ref=e124]: https://example.com/e2e-persist
      - listitem [ref=e125]:
        - link "E2E delete 1793697162780" [ref=e126] [cursor=pointer]:
          - /url: https://example.com/e2e-delete
        - generic [ref=e127]: https://example.com/e2e-delete
      - listitem [ref=e128]:
        - link "E2E add 1793697162653" [ref=e129] [cursor=pointer]:
          - /url: https://example.com/e2e-add
        - generic [ref=e130]: https://example.com/e2e-add
      - listitem [ref=e131]:
        - link "E2E delete 1793697001289" [ref=e132] [cursor=pointer]:
          - /url: https://example.com/e2e-delete
        - generic [ref=e133]: https://example.com/e2e-delete
      - listitem [ref=e134]:
        - link "E2E add 1793697000800" [ref=e135] [cursor=pointer]:
          - /url: https://example.com/e2e-add
        - generic [ref=e136]: https://example.com/e2e-add
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
  27 |     const item = page.locator('#links li').filter({ has: link });
> 28 |     const deleteResponse = page.waitForResponse((response) =>
     |                                 ^ Error: page.waitForResponse: Test timeout of 30000ms exceeded.
  29 |       response.url().includes('/api/links/') &&
  30 |       response.request().method() === 'DELETE'
  31 |     );
  32 |     await item.getByRole('button', { name: 'Delete' }).click();
  33 |     await deleteResponse;
  34 |     await expect(link).not.toBeVisible();
  35 |   });
  36 | });
```