const {test, expect} = require('@playwright/test');

test.beforeEach(async ({page}) => {
    await page.addInitScript(() => {
        window.API_HOST = 'http://localhost:8080'; // Your test server
        localStorage.setItem('token', 'token');
    });

    await page.goto('/app.html');

    await page.waitForSelector('.CodeMirror', {timeout: 10000});
    await page.waitForSelector('#sidebar-tree', {timeout: 5000});
});

test('sync', async ({ page }) => {
    await page.pause();
});

