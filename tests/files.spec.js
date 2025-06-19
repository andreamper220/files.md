const {test, expect} = require('@playwright/test');

test.beforeEach(async ({page}) => {
    await page.goto('/app.html');

    await page.waitForSelector('.CodeMirror', {timeout: 10000});
    await page.waitForSelector('#sidebar-tree', {timeout: 5000});
});

test('should load files', async ({ page }) => {
    await page.evaluate(() => {
        window.getRootDirHandle = async function() {
            // Your mock code here
            const opfsRoot = await navigator.storage.getDirectory();
            const testDir = await opfsRoot.getDirectoryHandle('test-files', { create: true });

            const testFiles = [
                { name: 'README.md', content: 'Hello world' },
                { name: 'Notes.md', content: '**Bold text**' }
            ];

            for (const fileData of testFiles) {
                try {
                    await testDir.getFileHandle(fileData.name);
                } catch (error) {
                    const fileHandle = await testDir.getFileHandle(fileData.name, { create: true });
                    const writable = await fileHandle.createWritable();
                    await writable.write(fileData.content);
                    await writable.close();
                }
            }

            return testDir;
        };
    });

    await page.evaluate(() => {
        init(document.getElementById("editor"));
    });
});

test('create new in subfolder', async ({ page }) => {
    await page.evaluate(() => {
        window.getRootDirHandle = async function() {
            // Your mock code here
            const opfsRoot = await navigator.storage.getDirectory();
            const testDir = await opfsRoot.getDirectoryHandle('dir', { create: true });

            const testFiles = [
                { name: 'README.md', content: 'Hello world' },
                { name: 'Notes.md', content: '**Bold text**' }
            ];

            for (const fileData of testFiles) {
                try {
                    await testDir.getFileHandle(fileData.name);
                } catch (error) {
                    const fileHandle = await testDir.getFileHandle(fileData.name, { create: true });
                    const writable = await fileHandle.createWritable();
                    await writable.write(fileData.content);
                    await writable.close();
                }
            }

            return opfsRoot;
        };
    });

    await page.evaluate(() => {
        init(document.getElementById("editor"));
    });

    await page.click('#new-file');
    await page.waitForTimeout(100);
    await page.keyboard.type('New file');
    await page.waitForTimeout(100);
    await page.keyboard.press('Enter');
    await page.keyboard.type('content');
    await page.waitForTimeout(700);

    await page.click('#sidebar >> text=dir');
    await page.waitForTimeout(100);

    await page.click('#sidebar >> text=New file');
    await page.waitForTimeout(100);
    const codeMirrorContent = await page.evaluate(() => {
        const cm = document.querySelector('.CodeMirror').CodeMirror;
        return cm.getValue();
    });
    expect(codeMirrorContent).toBe("# New file\ncontent\n");
});

test('create new in root', async ({ page }) => {
    await page.evaluate(() => {
        window.getRootDirHandle = async function() {
            // Your mock code here
            const opfsRoot = await navigator.storage.getDirectory();
            const testDir = await opfsRoot.getDirectoryHandle('dir', { create: true });

            const testFiles = [
                { name: 'README.md', content: 'Hello world' },
                { name: 'Notes.md', content: '**Bold text**' }
            ];

            for (const fileData of testFiles) {
                try {
                    await testDir.getFileHandle(fileData.name);
                } catch (error) {
                    const fileHandle = await testDir.getFileHandle(fileData.name, { create: true });
                    const writable = await fileHandle.createWritable();
                    await writable.write(fileData.content);
                    await writable.close();
                }
            }

            return opfsRoot;
        };
    });

    await page.evaluate(() => {
        init(document.getElementById("editor"));
    });

    await page.click('#new-file');
    await page.waitForTimeout(100);
    await page.keyboard.type('New file');
    await page.waitForTimeout(100);
    await page.keyboard.press('Enter');
    await page.keyboard.type('content');
    await page.waitForTimeout(700);

    await page.click('#sidebar >> text=New file');
    await page.waitForTimeout(100);
    const codeMirrorContent = await page.evaluate(() => {
        const cm = document.querySelector('.CodeMirror').CodeMirror;
        return cm.getValue();
    });
    expect(codeMirrorContent).toBe("# New file\ncontent\n");
});

test('create new lower case', async ({ page }) => {
    await page.evaluate(() => {
        window.getRootDirHandle = async function() {
            // Your mock code here
            const opfsRoot = await navigator.storage.getDirectory();
            const testDir = await opfsRoot.getDirectoryHandle('test-files', { create: true });

            const testFiles = [
                { name: 'README.md', content: 'Hello world' },
                { name: 'Notes.md', content: '**Bold text**' }
            ];

            for (const fileData of testFiles) {
                try {
                    await testDir.getFileHandle(fileData.name);
                } catch (error) {
                    const fileHandle = await testDir.getFileHandle(fileData.name, { create: true });
                    const writable = await fileHandle.createWritable();
                    await writable.write(fileData.content);
                    await writable.close();
                }
            }

            return testDir;
        };
    });

    await page.evaluate(() => {
        init(document.getElementById("editor"));
    });

    await page.click('#new-file');
    await page.waitForTimeout(100);
    await page.keyboard.type('another file');
    await page.waitForTimeout(100);
    await page.keyboard.press('Enter');
    await page.keyboard.type('content');
    await page.waitForTimeout(700);

    await page.click('#sidebar >> text=another file');
    await page.waitForTimeout(100);
    const codeMirrorContent = await page.evaluate(() => {
        const cm = document.querySelector('.CodeMirror').CodeMirror;
        return cm.getValue();
    });
    expect(codeMirrorContent).toBe("# Another file\ncontent\n");
});

test('move file between directories', async ({ page }) => {
    await page.evaluate(() => {
        window.getRootDirHandle = async function() {
            const opfsRoot = await navigator.storage.getDirectory();
            const testDir = await opfsRoot.getDirectoryHandle('test-files', { create: true });

            const projectsDir = await testDir.getDirectoryHandle('projects', { create: true });
            const archiveDir = await testDir.getDirectoryHandle('archive', { create: true });

            const rootFiles = [
                { name: 'README.md', content: 'Hello world' },
                { name: 'Todo.md', content: '- Task 1\n- Task 2' }
            ];

            for (const fileData of rootFiles) {
                try {
                    await testDir.getFileHandle(fileData.name);
                } catch (error) {
                    const fileHandle = await testDir.getFileHandle(fileData.name, { create: true });
                    const writable = await fileHandle.createWritable();
                    await writable.write(fileData.content);
                    await writable.close();
                }
            }

            const projectFiles = [
                { name: 'Project A.md', content: 'Project A details' },
                { name: 'Project B.md', content: 'Project B details' }
            ];

            for (const fileData of projectFiles) {
                try {
                    await projectsDir.getFileHandle(fileData.name);
                } catch (error) {
                    const fileHandle = await projectsDir.getFileHandle(fileData.name, { create: true });
                    const writable = await fileHandle.createWritable();
                    await writable.write(fileData.content);
                    await writable.close();
                }
            }

            const archiveFiles = [
                { name: 'Old Project.md', content: 'Archived project' }
            ];

            for (const fileData of archiveFiles) {
                try {
                    await archiveDir.getFileHandle(fileData.name);
                } catch (error) {
                    const fileHandle = await archiveDir.getFileHandle(fileData.name, { create: true });
                    const writable = await fileHandle.createWritable();
                    await writable.write(fileData.content);
                    await writable.close();
                }
            }

            return testDir;
        };
    });

    await page.evaluate(() => {
        init(document.getElementById("editor"));
    });

    // Wait for initialization
    await page.waitForTimeout(500);

    // Open a file from the projects directory
    await page.click('#sidebar >> text=projects');
    await page.waitForTimeout(100);
    await page.click('#sidebar >> text=Project A');
    await page.waitForTimeout(200);

    // Verify we're in the right file
    const initialContent = await page.evaluate(() => {
        const cm = document.querySelector('.CodeMirror').CodeMirror;
        return cm.getValue();
    });
    expect(initialContent).toContain('Project A details');

    // Open move modal with Cmd+M
    await page.keyboard.press('Meta+m');
    await page.waitForTimeout(100);

    // Verify move modal is open
    const moveModalVisible = await page.isVisible('#move');
    expect(moveModalVisible).toBe(true);

    // Check that move destinations are shown
    const moveResults = await page.locator('#move-results li');
    const destinations = await moveResults.allTextContents();
    expect(destinations).toContain('/');
    expect(destinations).toContain('archive');
    expect(destinations).toContain('projects');

    // Move to archive directory by clicking
    await page.click('#move-results >> text=archive');
    await page.waitForTimeout(200);

    // Verify modal is closed
    const moveModalVisibleAfter = await page.isVisible('#move');
    expect(moveModalVisibleAfter).toBe(false);

    // Verify file is now in archive directory
    // Check if the sidebar reflects the change
    await page.click('#sidebar >> text=archive');
    await page.waitForTimeout(100);

    // Should see Project A in archive now
    const archiveFiles = await page.locator('#sidebar >> text=archive').locator('..').locator('text=Project A');
    expect(await archiveFiles.count()).toBe(1);

    // Verify content is preserved
    await page.click('#sidebar >> text=archive >> .. >> text=Project A');
    await page.waitForTimeout(200);

    const finalContent = await page.evaluate(() => {
        const cm = document.querySelector('.CodeMirror').CodeMirror;
        return cm.getValue();
    });
    expect(finalContent).toContain('Project A details');
});

test('move file using keyboard navigation', async ({ page }) => {
    await page.evaluate(() => {
        window.getRootDirHandle = async function() {
            const opfsRoot = await navigator.storage.getDirectory();
            const testDir = await opfsRoot.getDirectoryHandle('test-files', { create: true });

            // Create directories
            const workDir = await testDir.getDirectoryHandle('work', { create: true });
            const personalDir = await testDir.getDirectoryHandle('personal', { create: true });

            // Create a file in root
            const rootFiles = [
                { name: 'Meeting Notes.md', content: 'Important meeting notes' }
            ];

            for (const fileData of rootFiles) {
                try {
                    await testDir.getFileHandle(fileData.name);
                } catch (error) {
                    const fileHandle = await testDir.getFileHandle(fileData.name, { create: true });
                    const writable = await fileHandle.createWritable();
                    await writable.write(fileData.content);
                    await writable.close();
                }
            }

            return testDir;
        };
    });

    await page.evaluate(() => {
        init(document.getElementById("editor"));
    });

    await page.waitForTimeout(500);

    // Open the file from root
    await page.click('#sidebar >> text=Meeting Notes');
    await page.waitForTimeout(200);

    // Open move modal
    await page.keyboard.press('Meta+m');
    await page.waitForTimeout(100);

    // Use arrow keys to navigate
    await page.keyboard.press('ArrowDown');
    await page.waitForTimeout(100);
    await page.keyboard.press('ArrowDown');
    await page.waitForTimeout(100); // move to 'work'

    // Press Enter to select
    await page.keyboard.press('Enter');
    await page.waitForTimeout(200);

    // Verify file moved to work directory
    await page.click('#sidebar >> text=work');
    await page.waitForTimeout(100);

    const workFiles = await page.locator('#sidebar >> text=work').locator('..').locator('text=Meeting Notes');
    expect(await workFiles.count()).toBe(1);
});
