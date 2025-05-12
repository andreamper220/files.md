const saverInterval = 50; // ms, how often to save currently open file
const loaderInterval = 3000; // ms, how often to load files from local system and sync with server

let unsavedChanges = false;
let isSaving = false;

// Files structure:
// {
//   "dir": [
//     {
//       "filename": [
//         {
//           content: "File content here...",
//           lastModified: <timestamp>,
//           handle: <file handle>,
//           imageUrl: <image url if any>
//         },
//         ...
//       ]
//     },
//     ...
//   ]
// }
let files = [];
const supportedFileTypes = ['md', 'txt', 'png', 'jpg', 'jpeg', 'webp', 'gif',];
const systemDirs = ["img", "archive", "_read_", "_watch_", "_shop_", "today", "later", "journal", "habits", "triggers", "places"];

let filesMetadata = {files: {}, timestamps: {}};
const SYNC_STORAGE_KEY = 'files';

// Returns files in flattened structure:
// {
//   "dir": {
//      ...
//   },
//   "dir/dir2": {
//      ...
//   },
// }
// The code is quite messy. We have to make lots of optimizations,
// otherwise it's going to be slow even with 5K files.
async function loadLocalFiles(rootDirHandle) {
    let newFiles = {};

    // Loads files recursively
    async function loadDir(dirHandle, path = "", depth = 1) {
        const entries = [];
        for await (const entry of dirHandle.values()) {
            entries.push(entry);
        }
        entries.sort((a, b) => a.name.localeCompare(b.name));

        const dirPromises = [];
        for (const entry of entries) {
            const filename = entry.name.normalize("NFC");

            if (entry.kind === 'directory') {
                if (filename.startsWith('.') || depth >= 5) continue;

                const dir = `${path}${filename}/`;
                newFiles[filename] = {};
                dirPromises.push({handle: entry, dir, depth: depth + 1});
            } else if (entry.kind === 'file' && supportedFileTypes.includes(filename.split('.').pop())) {
                const dir = path.replace(/\/+$/, '');
                if (!newFiles[dir]) newFiles[dir] = {};

                // Reuse existing file handle if it exists
                if (files?.[dir]?.[filename] !== undefined) {
                    newFiles[dir][filename] = files[dir][filename];
                    continue;
                }
                newFiles[dir][filename] = {handle: entry};

                entry.getFile().then(file => {
                    newFiles[dir][filename].lastModified = file.lastModified;
                });

                if (dir === 'img') {
                    getImageUrl(entry).then(imageUrl => {
                        newFiles[dir][filename].imageUrl = imageUrl;
                    });
                }
            }
        }

        await Promise.all(dirPromises.map(({handle, dir, depth}) =>
            loadDir(handle, dir, depth)
        ));
    }

    await loadDir(rootDirHandle);

    // Remove empty dirs
    for (const dir in newFiles) {
        if (Object.keys(newFiles[dir]).length === 0) {
            delete newFiles[dir];
        }
    }

    // Load metadata
    const savedStates = localStorage.getItem(SYNC_STORAGE_KEY);
    if (savedStates) {
        filesMetadata = JSON.parse(savedStates);
    }

    return newFiles;
}

async function syncWithServer() {
    const startTime = performance.now();
    console.log("Starting sync with server...");

    // Send locally modified files and timestamps of last seen dirs from the server
    let server = {};
    let filesToSend = await collectLocallyModifiedFiles();
    try {
        let response = await fetch('https://habits.files.md/sync', {
            method: 'POST',
            headers: {'Content-Type': 'application/json', 'Authorization': localStorage.getItem('token')},
            body: JSON.stringify({
                // TODO rem
                files: [],
                files: filesToSend,
                timestamps: filesMetadata['timestamps'] || [],
            })
        });
        if (!response.ok) {
            console.log(`Server responded with ${response.status}`);
            return;
        }

        server = await response.json();
    } catch (error) {
        console.error("Network error occurred:", error.message);
        return;
    }

    // Write files received from the server
    for (const fileInfo of server.files) {
        const {path, content, lastModified} = fileInfo;

        console.log("Syncing " + path);
        let fileHandle = await getFileHandle(path);
        let file = await fileHandle.getFile()
        let clientHash = hash(await file.text());
        let serverHash = hash(content);
        if (clientHash !== serverHash) {
            console.log("Hashes do not match, writing file...");
            // TODO rem
            // const writable = await fileHandle.createWritable();
            // await writable.write(content);
            // await writable.close();
        } else {
            console.log("Hashes match, no need to write file.");
        }
        if (!filesMetadata['files'][dir]) filesMetadata['files'][dir] = {};
        filesMetadata['files'][dir][filename] = {
            hash: hash(content),
            lastModified: lastModified,
            path: path
        };
    }
    filesMetadata['timestamps'] = server.timestamps;
    localStorage.setItem(SYNC_STORAGE_KEY, JSON.stringify(filesMetadata));
    console.log("Sync completed in " + (performance.now() - startTime) + "ms");
}

async function syncFileWithServer(dir, filename) {
    let file = getFileHandle(`${dir}/${filename}`);
    let lastModified = filesMetadata?.files?.[dir]?.[filename]?.lastModified ?? 0;
    console.log(lastModified, file);

    return;
    let serverFile = {};
    try {
        let response = await fetch('https://habits.files.md/syncFile', {
            method: 'POST',
            headers: {'Content-Type': 'application/json', 'Authorization': localStorage.getItem('token')},
            body: JSON.stringify({
                Path: `${dir}/${filename}`,
                LastModified: lastModified,
                Content: '',
            })
        });
        if (!response.ok) {
            console.log(`Server responded with ${response.status}`);
            return;
        }

        serverFile = await response.json();
    } catch (error) {
        console.error("Network error occurred:", error.message);
        return;
    }
}

async function collectLocallyModifiedFiles() {
    const filesToSend = [];
    const promises = [];
    for (const dir in files) {
        if (dir === 'img') continue; // Skip image directory

        for (const filename in files[dir]) {
            const promise = getFileIfChanged(dir, filename)
                .then(result => {
                    if (result) filesToSend.push(result);
                });
            promises.push(promise);
        }
    }

    await Promise.all(promises);
    return filesToSend;
}

async function getFileIfChanged(dir, filename) {
    try {
        const fileData = files[dir][filename];
        if (!fileData?.handle) return null;

        const file = await fileData.handle.getFile();
        const content = await file.text();

        const path = filesMetadata?.files?.[dir]?.[filename]?.path;
        if (!path) {
            console.log(`File ${dir}/${filename} not found on server, skipping...`);
            return null;
        }

        const serverHash = filesMetadata?.files?.[dir]?.[filename]?.hash;
        const serverTime = filesMetadata?.files?.[dir]?.[filename]?.lastModified;

        if (serverHash !== hash(content)) {
            return {
                content,
                path,
                lastModified: serverTime,
            };
        }

        return null;
    } catch (error) {
        console.error(`Error processing ${dir}/${filename}:`, error);
        return null;
    }
}

async function getFileHandle(path) {
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
                currentDirHandle = await currentDirHandle.getDirectoryHandle(dirName, {create: true});
            } catch (error) {
                console.error(`Error getting directory handle for '${dirName}':`, error);
                return null;
            }
        }
    }

    let fileHandle;
    try {
        fileHandle = await currentDirHandle.getFileHandle(filename, {create: true});
    } catch (error) {
        console.error(`Error getting file handle for '${dir}/${filename}':`, error);
        return null;
    }

    return fileHandle;
}

function getMetadata(path) {
    const parts = path.split('/');
    const filename = parts.pop();
    const dir = parts.join('/');

    if (filesMetadata['files']?.[dir]?.[filename]) {
        return filesMetadata['files'][dir][filename];
    } else {
        return null;
    }
}

function addMetadata(path, content, lastModified) {
    filesMetadata['files'] = filesMetadata['files'] ?? {};
    filesMetadata['files'][dir] = filesMetadata['files'][dir] ?? {};

    const parts = path.split('/');
    const filename = parts.pop();

    filesMetadata['files'][dir][filename] = {
        hash: hash(content),
        lastModified: lastModified,
        path: path
    };
}

function saveMetadata() {
    localStorage.setItem(SYNC_STORAGE_KEY, JSON.stringify(filesMetadata));
}

async function saveCurrentFile() {
    if (!unsavedChanges) return;

    // Wait until not saving
    while (isSaving) {
        await new Promise(r => setTimeout(r, 50));
    }

    isSaving = true;
    try {
        const dir = editor.currentDir;
        const filename = editor.currentFile;
        const fileData = files[dir][filename];
        if (fileData && fileData.handle) {
            let content = getCurrentContent();
            const writable = await fileData.handle.createWritable();
            await writable.write(content);
            // Buffer is flushed on disk at this moment. It could be interrupted
            // by the event loop, so we use isSaving guard.
            await writable.close();
        } else {
            if (fileData.handle) {
                alert(`Cannot save ${filename}. No file handle found.`);
            }
        }
    } catch (error) {
        console.error("Error during save:", error);
    }

    isSaving = false;
    unsavedChanges = false;
}

function hash(str) {
    let hash = 0;
    for (let i = 0, len = str.length; i < len; i++) {
        let chr = str.charCodeAt(i);
        hash = (hash << 5) - hash + chr;
        hash |= 0;
    }

    return hash;
}

async function initFiles() {
    const rootDirHandle = await getRootDirHandle();

    const startTime = performance.now();
    files = await loadLocalFiles(rootDirHandle);
    console.log(`Files loaded in ${performance.now() - startTime}ms`);
    await syncWithServer();

    window.loader = setInterval(async function () {
        // Check if current file has been modified
        let dir = editor.currentDir;
        let file = editor.currentFile;
        // TODO handle removed file cases etc
        const updatedFile = await files[dir]?.[file].handle.getFile();
        let newContent = await updatedFile.text();
        // TODO dirty hack, we replace links on the fly
        let currentContent = getCurrentContent();
        if (!unsavedChanges) {
            newContent = newContent.replace(/\[\[(.+?)\|.*?\]\]/g, '[[$1]]');
            if (norm(currentContent) !== norm(newContent)) {
                await showFile(dir, file, false);
            }
        }
    }, loaderInterval)
}

window.addEventListener('beforeunload', function () {
    clearInterval(window.loader);
    clearInterval(window.saver);
});


// Worker to process the saving queue
window.saver = setInterval(saveCurrentFile, saverInterval);