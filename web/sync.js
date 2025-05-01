let filesMetadata = {files: {}, timestamps: {}};
const SYNC_STORAGE_KEY = 'files';

async function initFilesMetadata() {
    const savedStates = localStorage.getItem(SYNC_STORAGE_KEY);
    if (savedStates) {
        filesMetadata = JSON.parse(savedStates);
    }
}

function saveFilesMetadata() {
    localStorage.setItem(SYNC_STORAGE_KEY, JSON.stringify(filesMetadata));
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

async function syncWithServer() {
    console.log("Starting sync with server...");

    const filesToSync = {};
    for (const dir in files) {
        if (!filesToSync[dir]) filesToSync[dir] = {};

        for (const filename in files[dir]) {
            try {
                if (dir === 'img') continue;

                let content = "";
                if (files[dir][filename].handle) {
                    const file = await files[dir][filename].handle.getFile();
                    content = await file.text();
                } else {
                    content = files[dir][filename]?.content || "";
                }

                let serverHash = filesMetadata?.files?.[dir]?.[filename]?.hash;
                let fileWasModifiedLocally = serverHash !== hash(content)
                if (fileWasModifiedLocally) {
                    filesToSync[dir][filename] = {
                        content: content,
                        lastModified: files[dir][filename].lastModified
                    };
                }
            } catch (error) {
                console.error(`Error processing ${dir}/${filename}:`, error);
            }
        }
    }

    try {
        const response = await fetch('https://habits.files.md/sync', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json', 'Authorization': localStorage.getItem('token')},
            body: JSON.stringify({
                files: filesToSync,
                timestamps: filesMetadata['timestamps'] || [],
            })
        });

        if (!response.ok) {
            throw new Error(`Server responded with ${response.status}`);
        }

        const server = await response.json();
        for (const fileInfo of server.files) {
            console.log(`Syncing file: ${fileInfo.path}`);
            const { path, content, lastModified} = fileInfo;

            let dir, filename;
            if (path.includes('/')) {
                const parts = path.split('/');
                filename = parts.pop();
                dir = parts.join('/');
            } else {
                dir = '';
                filename = path;
            }

            if (!files[dir]) files[dir] = {};

            if (!files[dir][filename] || !files[dir][filename].handle) {
                files[dir][filename] = {
                    content,
                    lastModified: lastModified
                };
            } else {
                // For files with handles, we would write to the file
                // But this is commented out in your code
                // const writable = await files[dir][filename].handle.createWritable();
                // await writable.write(content);
                // await writable.close();
            }

            if (!filesMetadata['files'][dir]) filesMetadata['files'][dir] = {};
            filesMetadata['files'][dir][filename] = {
                hash: hash(content),
                lastModified: lastModified
            };
        }
        filesMetadata['timestamps'] = server.timestamps;

        // Process files to upload
        // for (const fileInfo of syncResult.filesToUpload) {
        //     const { dir, filename } = fileInfo;
        //
        //     try {
        //         let content = "";
        //         if (files[dir][filename].handle) {
        //             const file = await files[dir][filename].handle.getFile();
        //             content = await file.text();
        //         } else {
        //             content = files[dir][filename].content;
        //         }
        //
        //         // Upload file to server
        //         const uploadResponse = await fetch(`/sync/upload`, {
        //             method: 'POST',
        //             headers: { 'Content-Type': 'application/json' },
        //             body: JSON.stringify({
        //                 dir,
        //                 filename,
        //                 content
        //             })
        //         });
        //
        //         if (uploadResponse.ok) {
        //             const result = await uploadResponse.json();
        //
        //             // Update server file state
        //             if (!filesMetadata[dir]) filesMetadata[dir] = {};
        //             filesMetadata[dir][filename] = {
        //                 hash: result.hash,
        //                 lastModified: Date.now()
        //             };
        //         }
        //     } catch (error) {
        //         console.error(`Error uploading ${dir}/${filename}:`, error);
        //     }
        // }
        saveFilesMetadata();
        console.log("Sync completed successfully");

    } catch (error) {
        console.error("Sync failed:", error);
    }
}

// Modify your init function to call sync after loading files
async function init(el) {
    initEditor(el);

    const savedDirectoryHandle = await getSavedDirectoryHandle();
    const userHasOpenedDirectory = savedDirectoryHandle instanceof FileSystemDirectoryHandle;
    if (!userHasOpenedDirectory) {
        document.getElementById('welcome').style.display = 'block';
        files = defaultFiles;
        buildSidebar();
        await showFile("", "Welcome.md");
        return;
    }

    const permission = await savedDirectoryHandle.queryPermission({mode: 'read'});
    if (permission !== 'granted') {
        document.getElementById('welcome').style.display = 'block';
    }

    files = await loadFiles(savedDirectoryHandle);

    // Initialize server file states and sync
    await initFilesMetadata();
    await syncWithServer();

    changesPollingInterval = setInterval(async function() {
        // Existing code...
    }, 3000);

    buildSidebar();
    await showRandomFile();
}

// // Add sync to the file save function
// async function saveFile() {
//     const dir = editor.currentDir;
//     const filename = editor.currentFile;
//     const fileData = files[dir][filename];
//     if (fileData && fileData.handle) {
//         let content = getCurrentContent();
//         const writable = await fileData.handle.createWritable();
//         await writable.write(content);
//         await writable.close();
//
//         // Calculate hash after saving
//         const hash = await calculateFileHash(content);
//
//         // Check if file needs to be synced with server
//         if (!filesMetadata[dir] ||
//             !filesMetadata[dir][filename] ||
//             filesMetadata[dir][filename].hash !== hash) {
//
//             // Queue sync for this file
//             await syncFileWithServer(dir, filename, content, hash);
//         }
//     } else {
//         if (fileData.handle) {
//             alert(`Cannot save ${filename}. No file handle found.`);
//         }
//     }
// }

// // Sync a single file with the server
// async function syncFileWithServer(dir, filename, content, hash) {
//     try {
//         const response = await fetch(`/sync/upload`, {
//             method: 'POST',
//             headers: { 'Content-Type': 'application/json' },
//             body: JSON.stringify({ dir, filename, content })
//         });
//
//         if (response.ok) {
//             // Update server file state
//             if (!filesMetadata[dir]) filesMetadata[dir] = {};
//             filesMetadata[dir][filename] = {
//                 hash,
//                 lastModified: Date.now()
//             };
//             saveFilesMetadata();
//         }
//     } catch (error) {
//         console.error(`Error syncing ${dir}/${filename}:`, error);
//     }
// }
//
