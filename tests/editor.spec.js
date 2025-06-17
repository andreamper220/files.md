const {test, expect} = require('@playwright/test');

test.describe('Files.md Text Editor Sync Tests', () => {
    test.beforeEach(async ({page}) => {
        await page.goto('http://app.localhost:8080/');

        await page.waitForSelector('.CodeMirror', {timeout: 10000});
        await page.waitForSelector('#sidebar-tree', {timeout: 5000});
    });

    test('should load the Files.md editor', async ({page}) => {
        await expect(page).toHaveTitle('Files.md (Alpha version)');

        await expect(page.locator('#sidebar')).toBeVisible();
        await expect(page.locator('.CodeMirror')).toBeVisible();
        await expect(page.locator('#open-folder')).toBeVisible();
    });

    test('should open markdown file via quick panel and see bold text formatting', async ({page}) => {
        const isMac = process.platform === 'darwin';
        const modifier = isMac ? 'Meta' : 'Control';
        await page.keyboard.press(`${modifier}+k`);

        await page.waitForSelector('#search', {timeout: 3000});
        await page.locator('#search-input').fill('Markdown');
        await page.keyboard.press('Enter');

        await page.waitForTimeout(1000);
        await page.waitForSelector('.CodeMirror', {timeout: 5000});

        const codeMirrorContent = await page.locator('.CodeMirror').textContent();

        expect(codeMirrorContent).toContain('**Bold text**');
        expect(codeMirrorContent).toContain('**bold**');
        expect(codeMirrorContent).toContain('__bold__');

        await expect(page.locator('.CodeMirror')).toContainText('Bold text');
        await expect(page.locator('.CodeMirror')).toContainText('**bold**');

        await expect(page.locator('.CodeMirror')).toContainText('using');
    });

    test('insert link', async ({page}) => {
        const isMac = process.platform === 'darwin';
        const modifier = isMac ? 'Meta' : 'Control';

        await page.waitForSelector('.CodeMirror', {timeout: 5000});

        await page.click('.CodeMirror');
        await page.keyboard.press('Control+a');
        await page.keyboard.press('Delete');
        await page.keyboard.type('[markdown');
        await page.keyboard.press('Enter');

        await page.waitForTimeout(500);
        const content = await page.locator('.CodeMirror-code').textContent();

        console.log('Content:', content);
        expect(content).toContain('[Markdown Guide](/Markdown Guide.md)');
        // await page.pause()
    });

});

