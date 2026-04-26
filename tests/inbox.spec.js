const {test, expect} = require('@playwright/test');

test.beforeEach(async ({page}) => {
    await page.goto('/index.html');

    // await page.waitForSelector('.CodeMirror', {timeout: 10000});
    await page.waitForSelector('#tree', {timeout: 5000});
});

test('send message to chat', async ({ page }) => {
    await page.click(`#tree .tree-description:has-text('today')`);
    await page.waitForSelector('#inbox');
    await page.keyboard.type('My message');
    await page.waitForTimeout(300);
    // TODO I believe chat is reloaded 2 times for some reason, it blinks, and thus removes previous message
    // Or wait for timeout before typing message doesn't help hmm
    await page.keyboard.press('Enter');

    await page.pause();
    await page.waitForSelector('.message');
    let content = await page.textContent('.message-content')
    expect(content).toBe('My message');

});

test('select all in chat input selects input text, not bubbles', async ({page}) => {
    await page.click(`#tree .tree-description:has-text('today')`);
    await page.waitForSelector('#inbox');
    await page.keyboard.type('First message');
    await page.waitForTimeout(300);
    await page.keyboard.press('Enter');
    await page.waitForSelector('.message');

    await page.locator('#inbox-input').click();
    await page.keyboard.type('to be cleared');
    await expect(page.locator('#inbox-input')).toHaveValue('to be cleared');

    const modifier = process.platform === 'darwin' ? 'Meta' : 'Control';
    await page.keyboard.press(`${modifier}+a`);
    await page.keyboard.press('Delete');

    await expect(page.locator('#inbox-input')).toHaveValue('');
    await expect(page.locator('.message-content')).toHaveText('First message');
});

test('move to dir creates a new file inside that dir', async ({page}) => {
    await page.evaluate(() => {
        window.getTemporaryStorageDirHandle = async function () {
            const root = await navigator.storage.getDirectory();
            await root.getDirectoryHandle('projects', {create: true});
            return root;
        };
    });
    await page.evaluate(() => init(document.getElementById('editor')));

    await page.click(`#tree .tree-description:has-text('today')`);
    await page.waitForSelector('#inbox');
    await page.keyboard.type('MyTask');
    await page.waitForTimeout(300);
    await page.keyboard.press('Enter');
    await page.waitForSelector('.message');

    await page.hover('.message');
    await page.locator('.to-file-btn').first().click({force: true});
    await page.waitForSelector('#search', {state: 'visible'});

    await page.locator('#search-results li[data-dir="projects"]').click();
    await page.waitForSelector('.message', {state: 'detached'});

    const exists = await page.evaluate(async () => {
        const root = await navigator.storage.getDirectory();
        const projects = await root.getDirectoryHandle('projects');
        try { await projects.getFileHandle('MyTask.md'); return true; }
        catch { return false; }
    });
    expect(exists).toBe(true);
});

test('move to root creates a new file at root', async ({page}) => {
    await page.evaluate(() => {
        window.getTemporaryStorageDirHandle = async function () {
            return await navigator.storage.getDirectory();
        };
    });
    await page.evaluate(() => init(document.getElementById('editor')));

    await page.click(`#tree .tree-description:has-text('today')`);
    await page.waitForSelector('#inbox');
    await page.keyboard.type('RootMsg');
    await page.waitForTimeout(300);
    await page.keyboard.press('Enter');
    await page.waitForSelector('.message');

    await page.hover('.message');
    await page.locator('.to-file-btn').first().click({force: true});
    await page.waitForSelector('#search', {state: 'visible'});

    await page.locator('#search-results li[data-dir=""]').click();
    await page.waitForSelector('.message', {state: 'detached'});

    const exists = await page.evaluate(async () => {
        const root = await navigator.storage.getDirectory();
        try { await root.getFileHandle('RootMsg.md'); return true; }
        catch { return false; }
    });
    expect(exists).toBe(true);
});

test('move to existing file appends content', async ({page}) => {
    await page.evaluate(() => {
        window.getTemporaryStorageDirHandle = async function () {
            const root = await navigator.storage.getDirectory();
            const fh = await root.getFileHandle('Notes.md', {create: true});
            const w = await fh.createWritable();
            await w.write('# Notes');
            await w.close();
            return root;
        };
    });
    await page.evaluate(() => init(document.getElementById('editor')));

    await page.click(`#tree .tree-description:has-text('today')`);
    await page.waitForSelector('#inbox');
    await page.keyboard.type('Append me');
    await page.waitForTimeout(300);
    await page.keyboard.press('Enter');
    await page.waitForSelector('.message');

    await page.hover('.message');
    await page.locator('.to-file-btn').first().click({force: true});
    await page.waitForSelector('#search', {state: 'visible'});

    await page.locator('#search-results li[data-path="/Notes.md"]').click();
    await page.waitForSelector('.message', {state: 'detached'});

    await page.click(`#tree .tree-description:has-text('Notes')`);
    await page.waitForTimeout(200);
    const content = await page.evaluate(() => document.querySelector('.CodeMirror').CodeMirror.getValue());
    expect(content).toContain('# Notes');
    expect(content).toContain('Append me');
});

test('system dirs (archive, today) are hidden in move-to-file modal', async ({page}) => {
    await page.evaluate(() => {
        window.getTemporaryStorageDirHandle = async function () {
            const root = await navigator.storage.getDirectory();
            await root.getDirectoryHandle('archive', {create: true});
            await root.getDirectoryHandle('projects', {create: true});
            return root;
        };
    });
    await page.evaluate(() => init(document.getElementById('editor')));

    await page.click(`#tree .tree-description:has-text('today')`);
    await page.waitForSelector('#inbox');
    await page.keyboard.type('Hello');
    await page.waitForTimeout(300);
    await page.keyboard.press('Enter');
    await page.waitForSelector('.message');

    await page.hover('.message');
    await page.locator('.to-file-btn').first().click({force: true});
    await page.waitForSelector('#search', {state: 'visible'});

    await expect(page.locator('#search-results li[data-dir="projects"]')).toBeVisible();
    await expect(page.locator('#search-results li[data-dir="archive"]')).toHaveCount(0);
    await expect(page.locator('#search-results li[data-dir="today"]')).toHaveCount(0);
});

test('send to chat and move to recent file', async ({ page }) => {
    await page.evaluate(() => {
        window.getTemporaryStorageDirHandle = async function() {
            const root = await navigator.storage.getDirectory();
            const fileHandle = await root.getFileHandle('File.md', { create: true });

            return root;
        };
    });

    await page.evaluate(() => {
        init(document.getElementById("editor"));
    });

    await page.click(`#tree .tree-description:has-text('today')`);
    await page.waitForSelector('#inbox');
    await page.keyboard.type('My message');
    await page.waitForTimeout(300);
    await page.keyboard.press('Enter');

    await page.waitForSelector('.message', );
    let content = await page.textContent('.message-content')
    expect(content).toBe('My message');

    await page.hover('.message-hover-zone', {force: true});
    await page.waitForTimeout(500);
    await page.waitForSelector('.message-actions', { state: 'visible' });

    await page.locator('.action-btn').filter({ hasText: 'file' }).click();
    await page.waitForTimeout(500);
    await page.waitForSelector('.message', { state: 'detached' });

    let dateString = await page.evaluate(() => {
        const date = new Date();
        const months = ['January', 'February', 'March', 'April', 'May', 'June',
            'July', 'August', 'September', 'October', 'November', 'December'];
        const days = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];

        const day = date.getDate();
        const month = months[date.getMonth()];
        const year = date.getFullYear();
        const weekday = days[date.getDay()];

        return `${day} ${month} ${year}, ${weekday}`;
    });
    dateString = dateString.replace(/,/g, ',')

    await clickAndExpectContent(page, 'file', `# File\n#### ${dateString}\nMy message`);
});


async function clickAndExpectContent(page, filePath, expectedContent) {
    const parts = filePath.split('/');
    const dirs = parts.slice(0, -1);
    const file = parts[parts.length - 1];

    for (const dir of dirs) {
        const isSelected = await page.locator(`#tree .tree-description:has-text('${dir}')`).evaluate(el => el.classList.contains('expanded'));
        if (!isSelected) {
            await page.click(`#tree .tree-description:has-text('${dir}')`);
            await page.waitForTimeout(100);
        }
    }

    await page.click(`#tree .tree-description:has-text('${file}')`);
    await page.waitForTimeout(200);

    const codeMirrorContent = await page.evaluate(() => {
        const cm = document.querySelector('.CodeMirror').CodeMirror;
        return cm.getValue();
    });
    expect(codeMirrorContent).toBe(expectedContent);
}
