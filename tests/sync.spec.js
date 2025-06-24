const {test, expect} = require('@playwright/test');
const fs = require('fs').promises;
const path = require('path');
const crypto = require('crypto');

const getServerDir = (workerIndex) => `../storage/${currentWorkerIndex}`;
const getTokensDir = () => `../storage/-1`;
let currentWorkerIndex = '-1';

test.beforeEach(async ({page}, testInfo) => {
    currentWorkerIndex = testInfo.workerIndex.toString();

    await fs.rm(getServerDir(), { recursive: true, force: true });
    await fs.mkdir(getServerDir(), { recursive: true });

    await fs.writeFile(getTokensDir() + `/` + saltToken(currentWorkerIndex), currentWorkerIndex, 'utf8');
});

async function setup(page) {
    await page.addInitScript((workerIndex) => {
        window.API_HOST = 'http://localhost:8080';
        localStorage.setItem('token', workerIndex);
    }, currentWorkerIndex);

    await page.goto('/app.html');

    await page.evaluate(()=> {
        window.getRootDirHandle = async function() {
            const root = await navigator.storage.getDirectory();
            const subdir = await root.getDirectoryHandle('subdir', { create: true });

            const files = [
                { name: 'README.md', content: 'Hello world' },
                { name: 'Notes.md', content: 'Some Text' }
            ];

            for (const file of files) {
                try {
                    await subdir.getFileHandle(file.name);
                } catch (error) {
                    const fileHandle = await subdir.getFileHandle(file.name, { create: true });
                    const writable = await fileHandle.createWritable();
                    await writable.write(file.content);
                    await writable.close();
                }
            }

            return root;
        };
    })
    await page.evaluate(() => {
        init(document.getElementById('editor'));
    });

    await page.waitForSelector('.CodeMirror', {timeout: 10000});
    await page.waitForSelector('#sidebar-tree', {timeout: 5000});
}

test('sync new files from server', async ({ page }) => {
    await createFileOnServer('file.md', 'test content');
    await createFileOnServer('another.md', '*italic*');

    await setup(page);

    // Check that existing files are not removed
    await expectFileContent(page, 'subdir/Notes', "# Notes\nSome Text");
    await expectFileContent(page, 'subdir/README', "# README\nHello world");

    // Check that new files are added
    await expectFileContent(page, 'file', "# File\ntest content");
    await expectFileContent(page, 'another', "# Another\n*italic*");
});

test('get changes for current file from server', async ({ page }) => {
    await createFileOnServer('file.md', 'test content');
    await createFileOnServer('another.md', '*italic*');

    await setup(page);

    // Check that existing files are not removed
    await expectFileContent(page, 'file', "# File\ntest content");
    await expectCurrentContent(page, "# File\ntest content");

    await createFileOnServer('file.md', 'test content\nadded');
    await page.waitForTimeout(2000);
    await expectCurrentContent(page, "# File\ntest content\nadded");
});

async function createFileOnServer(filepath, content) {
    const p = path.join(getServerDir(), filepath);
    try {
        await fs.writeFile(p, content, 'utf8');
    } catch (error) {
        console.error('Error creating file:', error);
    }
}

function saltToken(token, salt = "") {
    return crypto.createHash('sha256')
        .update(token + salt)
        .digest('hex');
}

async function expectFileContent(page, filePath, expectedContent) {
    const parts = filePath.split('/');
    const dirs = parts.slice(0, -1);
    const file = parts[parts.length - 1];

    for (const dir of dirs) {
        const isSelected = await page.locator(`#sidebar-tree .tj_description:has-text("${dir}")`).evaluate(el => el.classList.contains('expanded'));
        if (!isSelected) {
            await page.click(`#sidebar-tree .tj_description:has-text("${dir}")`);
            await page.waitForTimeout(100);
        }
    }

    await page.click(`#sidebar-tree .tj_description:has-text("${file}")`);
    await page.waitForTimeout(200);

    const codeMirrorContent = await page.evaluate(() => {
        const cm = document.querySelector('.CodeMirror').CodeMirror;
        return cm.getValue();
    });
    expect(codeMirrorContent).toBe(expectedContent);
}

async function expectCurrentContent(page, content) {
    const codeMirrorContent = await page.evaluate(() => {
        const cm = document.querySelector('.CodeMirror').CodeMirror;
        return cm.getValue();
    });
    expect(codeMirrorContent).toBe(content);
}