// Forbidden chars on Windows / PWA / Unix.
const FORBIDDEN_FILENAME_CHARS = ['<', '>', ':', '"', '|', '\\', '?', '*', '\x00', '/'];

function sanitizeFilename(filename) {
    return FORBIDDEN_FILENAME_CHARS.reduce((result, ch) => result.replaceAll(ch, ''), filename);
}

function findForbiddenChar(name) {
    for (const ch of FORBIDDEN_FILENAME_CHARS) {
        if (name.includes(ch)) return ch;
    }
    return null;
}

async function getFileHandle(path, create = false) {
    let dir, filename;
    if (path.includes('/')) {
        const parts = path.split('/');
        filename = parts.pop();
        dir = parts.join('/');
    } else {
        dir = '';
        filename = path;
    }

    const dirs = dir.split('/');
    let currentDirHandle = await getRootDirHandle();
    for (const dirName of dirs) {
        if (dirName) {
            try {
                currentDirHandle = await currentDirHandle.getDirectoryHandle(dirName, {create: create});
            } catch (error) {
                throw error;
            }
        }
    }

    let fileHandle;
    try {
        fileHandle = await currentDirHandle.getFileHandle(filename, {create: create});
    } catch (error) {
        throw error;
    }

    return fileHandle;
}

async function read(path) {
    let fileHandle = await getFileHandle(path)
    let file = await fileHandle.getFile();

    return await file.text();
}

async function write(path, content) {
    let fileHandle = await getFileHandle(path, true);
    const writable = await fileHandle.createWritable();
    await writable.write(content);
    await writable.close();
}

async function writeAtEnd(path, content) {
    let fileHandle = await getFileHandle(path, true);
    if (fileHandle === null) {
        // TODO fix once Chromium fixes the bug
        throw new Error('Invalid file name');
    }

    const writable = await fileHandle.createWritable({ keepExistingData: true });
    await writable.seek(await fileHandle.getFile().then(file => file.size));
    await writable.write(content);
    await writable.close();

    const file = await fileHandle.getFile();
    return file.lastModified;
}

// TODO save metadata & files
// Write only if content is different.
async function writeIfContentIsDifferent(path, content) {
    let fileHandle = await getFileHandle(path, true);
    if (fileHandle === null) {
        // TODO fix once Chromium fixes the bug
        throw new Error('Invalid file name');
    }

    const fileExists = !await exists(path);
    if (fileExists || !await isContentEqual(path, content)) {
        // TODO what if we're syncing first time and already have changes?
        log('Hashes do not match, writing file...', path);
        const writable = await fileHandle.createWritable();
        await writable.write(content);
        await writable.close();
    } else {
        log('Hashes match, no need to write file.');
    }

    const file = await fileHandle.getFile();
    return file.lastModified;
}

// Works only for files.
async function exists(path) {
    try {
        await getFileHandle(path);
        return true;
    } catch (error) {
        if (error.name === 'NotFoundError') {
            return false
        }
        throw error
    }
}

async function remove(path) {
    let fileHandle = await getFileHandle(path);
    if (fileHandle === null) {
        // TODO fix once Chromium fixes the bug
        logError('Malformed name, skipping file...');
        return;
    }
    await fileHandle.remove()
    log(`File ${path} removed successfully.`);

    removeMemFile(path);
}

async function rename(oldpath, newpath) {
    let content = await read(oldpath)
    await write(newpath, content)
    await remove(oldpath)
}

// removeDir deletes a directory and everything under it. Files are deleted
// one-by-one so the in-memory file tree and server-sync bookkeeping stay in
// sync; the empty parent entry is then pruned from OPFS.
async function removeDir(dirPath) {
    const filePaths = collectFilePathsInDir(dirPath);
    for (const p of filePaths) {
        try {
            await remove(p);
        } catch (err) {
            logError('removeDir: failed to remove file', p, err);
        }
    }

    const parts = trimPrefix(dirPath, '/').split('/').filter(Boolean);
    const dirName = parts.pop();
    const parentPath = '/' + parts.join('/');

    const rootHandle = await getRootDirHandle();
    let parentHandle = rootHandle;
    for (const seg of parts) {
        parentHandle = await parentHandle.getDirectoryHandle(seg);
    }
    try {
        await parentHandle.removeEntry(dirName, { recursive: true });
    } catch (err) {
        logError('removeDir: removeEntry failed', dirPath, err);
    }

    removeMemDir(dirPath);
    log(`Dir ${dirPath} removed.`);
}

// renameDir moves every file under oldDirPath into a sibling directory called
// newName. Per-file moves keep server-sync bookkeeping intact; afterwards the
// empty old directory entry is removed from OPFS.
async function renameDir(oldDirPath, newName) {
    const parts = trimPrefix(oldDirPath, '/').split('/').filter(Boolean);
    parts.pop();
    const parentPath = '/' + parts.join('/');
    const newDirPath = joinPath(parentPath, newName);

    if (newDirPath === oldDirPath) return;

    const filePaths = collectFilePathsInDir(oldDirPath);
    if (filePaths.length === 0) {
        // Empty dir: make sure the new dir exists before we drop the old one,
        // otherwise rename silently deletes it.
        await createDir(newDirPath);
    }
    for (const oldFilePath of filePaths) {
        const rel = oldFilePath.slice(oldDirPath.length);
        const newFilePath = newDirPath + rel;
        try {
            await moveFile(oldFilePath, newFilePath);
        } catch (err) {
            logError('renameDir: failed to move file', oldFilePath, err);
        }
    }

    const rootHandle = await getRootDirHandle();
    let parentHandle = rootHandle;
    for (const seg of parts) {
        parentHandle = await parentHandle.getDirectoryHandle(seg);
    }
    const oldDirName = oldDirPath.split('/').filter(Boolean).pop();
    try {
        await parentHandle.removeEntry(oldDirName, { recursive: true });
    } catch (err) {
        logError('renameDir: removeEntry old dir failed', oldDirPath, err);
    }

    removeMemDir(oldDirPath);
    log(`Dir ${oldDirPath} renamed to ${newDirPath}.`);
}

// collectFilePathsInDir returns absolute paths of every file under dirPath,
// using the in-memory file tree so we don't hit OPFS for the listing.
function collectFilePathsInDir(dirPath) {
    const collected = [];
    walk(files, (p, isFile) => {
        if (!isFile) return;
        if (p === dirPath || p.startsWith(dirPath + '/')) {
            collected.push(p);
        }
    });
    return collected;
}

// removeMemDir drops a directory subtree from the in-memory file map.
function removeMemDir(dirPath) {
    const parts = trimPrefix(dirPath, '/').split('/').filter(Boolean);
    const dirName = parts.pop();
    let cur = files;
    for (const seg of parts) {
        cur = cur[seg + '/'];
        if (!cur) return;
    }
    delete cur[dirName + '/'];
}

async function mkdir(path) {
    try {
        let currentDirHandle = await getRootDirHandle();
        await currentDirHandle.getDirectoryHandle(path, {create: true});
    } catch (e) {
        logError(e);
        throw e;
    }
}

async function mkdirAll(path) {
    const dirs = path.split('/');
    let currentDirHandle = await getRootDirHandle();
    for (const dirName of dirs) {
        if (dirName) {
            await mkdir(path)
        }
    }
}

// createDir creates an empty directory on OPFS at the given path and registers
// it in the in-memory file tree so the sidebar picks it up.
async function createDir(dirPath) {
    const parts = trimPrefix(dirPath, '/').split('/').filter(Boolean);
    if (parts.length === 0) return;

    let dirHandle = await getRootDirHandle();
    for (const seg of parts) {
        dirHandle = await dirHandle.getDirectoryHandle(seg, { create: true });
    }

    let cur = files;
    for (const seg of parts) {
        const key = seg + '/';
        if (!cur[key]) cur[key] = {};
        cur = cur[key];
    }
    log(`Dir ${dirPath} created.`);
}

async function writeMediaFile(fileName, file) {
    try {
        const rootHandle = await getRootDirHandle();

        let mediaHandle;
        try {
            mediaHandle = await rootHandle.getDirectoryHandle('media');
        } catch {
            mediaHandle = await rootHandle.getDirectoryHandle('media', {create: true});
        }

        const fileHandle = await mediaHandle.getFileHandle(fileName, {create: true});
        const writable = await fileHandle.createWritable();
        await writable.write(file);
        await writable.close();

        const path = '/media/' + fileName;
        addMemFile(path, {
            isFile: true,
            path: path,
            imageUrl: await getImageUrl(fileHandle),
        });

        return fileHandle;
    } catch (error) {
        console.error('Error saving file:', error);
        return null;
    }
}

function generateSafeFilename(originalName) {
    const now = new Date();
    const timestamp = `${String(now.getDate()).padStart(2, '0')}.${String(now.getMonth() + 1).padStart(2, '0')}.${now.getFullYear()} ${String(now.getHours()).padStart(2, '0')}:${String(now.getMinutes()).padStart(2, '0')}`;
    return `${timestamp}-${originalName}`.replace(/[<>:"/\\|?*\s]/g, '-');
}