// tests/e2e/playwright.config.js
const { defineConfig, devices } = require('@playwright/test');
const path = require('path');

module.exports = defineConfig({
    testDir: '.',
    fullyParallel: true,
    forbidOnly: !!process.env.CI,
    retries: process.env.CI ? 2 : 0,
    workers: process.env.CI ? 1 : undefined,
    reporter: 'html',

    use: {
        // baseURL: `file://${path.resolve('../web/app.html')}`,
        baseUrl: 'http://app.localhost:3000',
        trace: 'on-first-retry',
        screenshot: 'only-on-failure',
        video: 'retain-on-failure',
        actionTimeout: 5000,
        navigationTimeout: 5000,
    },

    timeout: 5000,
    expect: {
        timeout: 3000
    },

    webServer: {
        command: 'npx http-server ../web -p 3000',
        port: 3000,
        reuseExistingServer: !process.env.CI,
    },

    projects: [
        {
            name: 'chromium',
            use: { ...devices['Desktop Chrome'] },
        },
        // Firefox and Webkit aren't currently supporting showDirectoryPicker
        // {
        //     name: 'firefox',
        //     use: { ...devices['Desktop Firefox'] },
        // },
        // {
        //     name: 'webkit',
        //     use: { ...devices['Desktop Safari'] },
        // },
    ],
});