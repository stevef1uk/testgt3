const { defineConfig, devices } = require('@playwright/test');

module.exports = defineConfig({
  testDir: './tests',
  retries: 0,
  reporter: 'list',
  use: {
    baseURL: 'http://localhost:8080',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
  webServer: {
    command: 'go run ./cmd/server',
    cwd: '.',
    port: 8080,
  },
});