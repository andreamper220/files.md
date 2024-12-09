// Files structure:
// {
//   folder: [
//     {
//       filename: [
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
const allowedFileTypes = [
    'md',
    'txt',
    'png',
    'jpg',
    'jpeg',
    'webp',
    'gif',
];

// HyperMD/Codemirror editor
let editor = null;

let focusedItemIndex = -1;

function initEditor(el) {
    editor = HyperMD.fromTextArea(el, {
        mode: "hypermd",
        lineNumbers: false,
        extraKeys: {
            // "Shift-Space": "autocomplete",
            'Cmd-[': false,
            'Cmd-]': false,
        },
        hintOptions: {
            hint: CompleteEmoji.createHintFunc(),
            closeCharacters: /$^/,
            closeOnUnfocus: false,
            completeSingle: false,
            alignWithWord: false
        },
        hmdFoldEmoji: {
            myEmoji: createAutocompleteDict
        }
    });
    editor.setSize(null, "100%");

    editor.hmdResolveURL = function (path) {
        if (typeof path === 'undefined') {
            return path
        }

        path = path.replace(/%20/g, " ");

        if (/^(?!http|https|\[).+\.md$/.test(path)) {
            let parts = path.split('/');
            if (parts.length === 1) {
                showFile("", path);
                return;
            }
            showFile(parts[0], parts[1]);
            return path;
        }

        const match = path.match(/^img\/(.+\.(png|jpg|jpeg|gif|webp))$/i);

        if (match && files['img'] && files['img'][match[1]]) {
            return files['img'][match[1]].imageUrl;
        }

        return path;
    };

    editor.hmdReadLink = async function (path) {
        path = path.replace('[', '').replace(']', '');
        let parts = path.split('/');
        if (parts.length === 1) {
            await showFile("", path + '.md');
            return;
        }

        await showFile(parts[0], parts[1] + '.md');
    };

    editor.on("inputRead", async function (cm, change) {
        if (change.text.length === 1 && change.text[0] === '[') {
            editor.showHint({
                completeSingle: false,
                updateOnCursorActivity: true,
            })
        }
    })

    editor.on("change", async function (cm, changeObj) {
        // Save on user input only
        if (changeObj.origin && changeObj.origin !== "setValue") {
            await saveFile();
        }
    });

    editor.setOption("viewportMargin", Infinity);

    // TODO Image uploading
    editor.on("paste", async (_, event) => {
        const items = (event.clipboardData || event.originalEvent.clipboardData).items;
        for (const item of items) {
            if (item.kind === "file" && item.type.startsWith("image/")) {
                const file = item.getAsFile();
                const fileName = `${new Date().toISOString().replace(/[:.]/g, '-')}.png`;
                await saveImageToDirectory(file, fileName);

                const markdownImageSyntax = `![](img/${fileName})`;
                editor.replaceSelection(markdownImageSyntax);
                // if (fileHandle) {
                //     // Insert the Markdown image syntax into the editor
                //     const markdownImageSyntax = `![[${fileName}]]`;
                //     editor.replaceSelection(markdownImageSyntax);
                //     console.log(`Image saved as: ${fileName}`);
                // } else {
                //     console.error("Failed to save the image.");
                // }
            }
        }
    });

    editor.addKeyMap({
        'Cmd-Y': function (cm) {
            cm.replaceSelection('✅ ');
            cm.focus();
        },
        'Cmd-B': function (cm) {
            const selectedText = cm.getSelection();
            const isBold = selectedText.startsWith("**") && selectedText.endsWith("**");

            let start = cm.getCursor("start");
            let end = cm.getCursor("end");
            if (isBold) {
                cm.replaceSelection(selectedText.slice(2, -2));
                cm.setSelection({line: start.line, ch: start.ch}, {line: end.line, ch: end.ch - 4});
            } else {
                cm.replaceSelection(`**${selectedText}**`);
                cm.setSelection({line: start.line, ch: start.ch}, {line: end.line, ch: end.ch + 4});
            }
            cm.focus();
        },
        'Cmd-I': function (cm) {
            const selectedText = cm.getSelection();
            const isItalic = selectedText.startsWith("*") && selectedText.endsWith("*");

            let start = cm.getCursor("start");
            let end = cm.getCursor("end");
            if (isItalic) {
                cm.replaceSelection(selectedText.slice(1, -1));
                cm.setSelection({line: start.line, ch: start.ch}, {line: end.line, ch: end.ch - 2});
            } else {
                cm.replaceSelection(`*${selectedText}*`);
                cm.setSelection({line: start.line, ch: start.ch}, {line: end.line, ch: end.ch + 2});
            }
            cm.focus();
        },
    });
}

function createAutocompleteDict() {
    const dict = {};
    const ignoredFolders = ["img", "archive", "_read_", "_watch_", "_shop_", "today", "later", "journal", "journal/past", "habits", "triggers", "places", ""];

    Object.keys(files).forEach(folder => {
        if (ignoredFolders.includes(folder)) return;

        Object.keys(files[folder]).forEach(filename => {
            const key = `${filename.replace(/\.md$/, "")}`;
            const filePath = `${filename.replace(/\.md$/, "")}](${folder}/${filename})`;
            dict[key] = filePath;
        });
    });

    return dict;
}

async function saveImageToDirectory(file, fileName) {
    try {
        let imgDirHandle = await getSavedDirectoryHandle();
        if (!imgDirHandle) return null;

        // Create a file handle for the image file
        const fileHandle = await imgDirHandle.getFileHandle(fileName, {create: true});
        const writable = await fileHandle.createWritable();
        await writable.write(file);
        await writable.close();

        console.log("Image saved");
        return fileHandle;
    } catch (error) {
        console.error("Error saving image:", error);
        return null;
    }
}

async function getImageUrl(fileHandle) {
    const file = await fileHandle.getFile();
    return URL.createObjectURL(file);
}

function buildSidebar() {
    let root = new TreeNode("files");
    for (dir in files) {
        if (dir === '' || dir === 'img') {
            continue;
        }

        let dirNode = new TreeNode(dir, {expanded: false});
        for (let file in files[dir]) {
            let fileNode = new TreeNode(file.replace(/\.md$/, ''), {expanded: false});
            fileNode.on('click', async function (n, node) {
                await showFile(node.parent.toString(), node.toString() + ".md");
            });
            dirNode.addChild(fileNode);
        }
        root.addChild(dirNode);
    }

    if (files['']) {
        // Adding root files after folders
        for (let file in files[""]) {
            let fileNode = new TreeNode(file.replace(/\.md$/, ''), {expanded: false});
            fileNode.on('click', async function (n, node) {
                await showFile("", node.toString() + ".md");
            });
            root.addChild(fileNode)
        }
    }

    new TreeView(root, "#editor-sidebar", {
        show_root: false,
    });
}

async function openDirectory() {
    let dirHandle = await window.showDirectoryPicker();
    document.getElementById('welcome').style.display = 'none';
    await saveDirectoryHandle(dirHandle);
    await loadDirectory(dirHandle)
    await showRandomFile();
}

async function loadDirectory(dirHandle, path = "", depth = 1) {
    const entries = [];
    for await (const entry of dirHandle.values()) {
        entries.push(entry);
    }
    entries.sort((a, b) => a.name.localeCompare(b.name));

    for (const entry of entries) {
        const filename = entry.name.normalize("NFC");
        if (entry.kind === 'directory') {
            if (filename.startsWith('.')) continue;

            if (depth < 5) {
                const folder = `${path}${filename}/`;
                files[filename] = {};
                await loadDirectory(entry, folder, depth + 1);
            }
        } else if (entry.kind === 'file' &&  allowedFileTypes.includes(filename.split('.').pop())
        ) {
            const folder = path.split('/').filter(Boolean).join('/');
            if (!files[folder]) files[folder] = {};
            let file = await entry.getFile();

            files[folder][filename] = {handle: entry, lastModified: file.lastModified};
            if (folder === 'img') {
                files[folder][filename].imageUrl = await getImageUrl(entry)
            }
        }
    }

    // Remove empty folders
    for (const folder in files) {
        if (Object.keys(files[folder]).length === 0) {
            delete files[folder];
        }
    }
    buildSidebar();
}

async function showRandomFile() {
    const ignoredFolders = ["img", "archive", "_read_", "_watch_", "_shop_", "today", "later", "journal", "habits", "triggers", "places", ""];

    const allFiles = [];
    for (let folder in files) {
        if (ignoredFolders.includes(folder)) continue;

        for (let file in files[folder]) {
            allFiles.push({folder, file});
        }
    }

    if (allFiles.length === 0) {
        console.error("No files found to open.");
        return;
    }

    const randomFile = allFiles[Math.floor(Math.random() * allFiles.length)];

    try {
        await showFile(randomFile.folder, randomFile.file);
    } catch (error) {
        console.error("Failed to open random file:", error);
    }
}

async function showFile(folder, filename, saveToHistory = true) {
    filename = filename.normalize("NFC");
    const fileData = files[folder][filename];
    const file = await fileData.handle.getFile();
    const header = filename.replace(/\.md$/, "").replace(/^\w/, (c) => c.toUpperCase());
    let content = await file.text();
    content = `# ${header}\n${content}`;
    content = content.replace(/\[\[(.+?)\|.*?\]\]/g, '[[$1]]');

    editor.currentFolder = folder;
    editor.currentFile = filename;
    if (saveToHistory) {
        const state = {folder: folder, file: filename};
        history.pushState(state, '');
    }

    editor.getDoc().setValue(content);
    editor.clearHistory();

    // Set cursor at the end of the page.
    // We need to execute this code after some rendering loop. If we don't do that,
    // Images and other heavy stuff won't be loaded
    setTimeout(() => {
        const lastLine = editor.lastLine();
        let targetLine = lastLine;
        for (let i = lastLine; i >= 0; i--) {
            const lineContent = editor.getLine(i).trim();
            if (!lineContent.startsWith("[") && (!lineContent.endsWith("]") || !lineContent.endsWith(")"))) {
                targetLine = i;
                break;
            }
        }
        const targetChar = editor.getLine(targetLine).length;
        editor.setCursor({line: targetLine, ch: targetChar});
        editor.scrollTo(null, 0);
        // TODO only focus if there's no quick dialogue
        editor.focus();
    }, 100);
}

async function saveFile() {
    const folder = editor.currentFolder;
    const filename = editor.currentFile;
    const fileData = files[folder][filename];
    if (fileData && fileData.handle) {
        let content = editor.getValue();
        const header = filename.replace('.md', '').replace(/^\w/, (c) => c.toUpperCase());
        content = content.trimStart();
        if (content.startsWith(`# ${header}`)) {
            content = content.slice(`# ${header}`.length).trimStart();
        }

        const writable = await fileData.handle.createWritable();
        await writable.write(content);
        await writable.close();
    } else {
        alert(`Cannot save ${filename}. No file handle found.`);
    }
}

function initDB() {
    return new Promise((resolve, reject) => {
        const request = indexedDB.open('files', 1);
        request.onerror = () => reject(request.error);
        request.onsuccess = () => resolve(request.result);
        request.onupgradeneeded = () => {
            const db = request.result;
            if (!db.objectStoreNames.contains('handles')) {
                db.createObjectStore('handles');
            }
        };
    });
}

async function saveDirectoryHandle(directoryHandle) {
    const db = await initDB();
    const transaction = db.transaction('handles', 'readwrite');
    const store = transaction.objectStore('handles');
    await store.put(directoryHandle, 'savedDirectoryHandle');
}

async function getSavedDirectoryHandle() {
    const db = await initDB();
    return new Promise((resolve, reject) => {
        const transaction = db.transaction('handles', 'readonly');
        const store = transaction.objectStore('handles');
        const request = store.get('savedDirectoryHandle');
        request.onsuccess = () => resolve(request.result);
        request.onerror = () => reject(request.error);
    });
}

async function init(el) {
    initEditor(el);
    buildSidebar();
    const savedDirectoryHandle = await getSavedDirectoryHandle();
    if (savedDirectoryHandle instanceof FileSystemDirectoryHandle) {
        const permission = await savedDirectoryHandle.queryPermission({mode: 'read'});
        if (permission !== 'granted') {
            document.getElementById('welcome').style.display = 'flex';
        }
        await loadDirectory(savedDirectoryHandle);
        await showRandomFile();
    } else {
        document.getElementById('welcome').style.display = 'flex';
    }
}

function updateFocusedItem(resultsList) {
    document.querySelectorAll('#goToFileResults li').forEach(li => li.classList.remove('focused'));
    resultsList.forEach((item, index) => {
        if (index === focusedItemIndex) {
            item.classList.add('focused');
            item.scrollIntoView({block: "nearest"});
        } else {
            item.classList.remove('focused');
        }
    });
}

function openSearchModal() {
    document.getElementById('goToFile').style.display = 'block';
    const inputField = document.getElementById('goToFileInput');
    inputField.focus();

    focusedItemIndex = -1;
    const goToFileResults = document.getElementById('goToFileResults');
    goToFileResults.innerHTML = '';
    loadRecentFiles();
}

document.addEventListener('keydown', (event) => {
    if (event.metaKey && event.key === 'p') {
        event.preventDefault();
        document.getElementById('goToFileInput').value = ''
        openSearchModal();
    }

    if (event.metaKey && event.key === 'k') {
        event.preventDefault();
        document.getElementById('goToFileInput').value = ''
        openSearchModal();
    }
});

function closeSearchModal() {
    document.getElementById('goToFile').style.display = 'none';
}

function loadRecentFiles() {
    const ignoredDirs = ["archive", "_read_", "_watch_", "_shop_", "habits", "triggers", "journal", "today", "later", "insights"];
    let results = [];
    for (const folder of Object.keys(files)) {
        if (folder !== 'img' && !ignoredDirs.includes(folder)) {
            for (const filename of Object.keys(files[folder])) {
                results.push({
                    folder,
                    filename,
                    lastModified: files[folder][filename].lastModified,
                });
            }
        }
    }

    results = results
        .sort((a, b) => b.lastModified - a.lastModified)
        .slice(0, 8);

    showResults(results);
}

function filterFiles() {
    const search = document.getElementById('goToFileInput').value.toLowerCase();
    if (search.trim() === '') {
        loadRecentFiles();
        return;
    }

    const list = document.getElementById('goToFileResults');
    list.innerHTML = '';

    const lowPriorityFolders = ["archive", "_read_", "_watch_", "_shop_", "habits", "triggers", "today", "later"];

    let results = [];

    let searchableFiles = Object.keys(files)
        .filter(folder => folder !== 'img')
        .flatMap(folder =>
            Object.keys(files[folder]).map(filename => ({
                folder,
                filename,
                lastModified: files[folder][filename].lastModified,
            }))
        );

    // Levenshtein distance
    searchableFiles.forEach(({filename, folder}) => {
        const potentialMatch = filename.replace(/\.md$/, "");
        let similarityScore = similarity(search, potentialMatch);
        if (similarityScore >= 70) {
            if (lowPriorityFolders.includes(folder)) {
                similarityScore -= 30;
            }
            results.push({filename: filename, folder: folder, score: similarityScore});
        }
    });

    // Substring
    searchableFiles.forEach(({filename, folder}) => {
        const potentialMatch = filename.replace(/\.md$/, "");
        const isSubstringMatch = potentialMatch.toLowerCase().includes(search.toLowerCase());
        if (!isSubstringMatch) {
            return;
        }

        let matchedPercent = search.length / potentialMatch.length * 100;
        results.push({
            filename: filename,
            folder: folder,
            score: Math.round(matchedPercent)
        });
    });

    const uniqueResultsMap = new Map();
    for (let i = 0; i < results.length; i++) {
        const item = results[i];
        const key = `${item.filename}-${item.folder}`;

        if (!uniqueResultsMap.has(key) || uniqueResultsMap.get(key).score < item.score) {
            uniqueResultsMap.set(key, item);
        }
    }
    results = Array.from(uniqueResultsMap.values()).sort((a, b) => b.score - a.score);
    showResults(results);
}

function showResults(results) {
    const list = document.getElementById('goToFileResults');
    results.forEach(({folder, filename}, index) => {
        const listItem = document.createElement('li');
        let title = filename.replace(/\.md$/, "")
        if (folder !== '') {
            listItem.textContent = `${folder}/${title}`;
        } else {
            listItem.textContent = title;
        }
        listItem.setAttribute('data-path', `${folder}/${filename}`);
        listItem.setAttribute('data-index', index);
        listItem.onclick = () => {
            showFile(folder, filename);
            closeSearchModal();
        };
        listItem.onmouseenter = () => {
            document.querySelectorAll('#goToFileResults li').forEach(li => li.classList.remove('focused'));
            listItem.classList.add('focused');
            focusedItemIndex = index;
        };
        list.appendChild(listItem);
    });

    focusedItemIndex = 0;
    updateFocusedItem(list.querySelectorAll('li'));
}

function closeGoToFile() {
    document.getElementById('goToFile').style.display = 'none';
}

document.addEventListener('keydown', (event) => {
    if (event.key === 'Escape') {
        closeGoToFile();
    }
});

// Toggle focus mode
document.addEventListener('keydown', function (event) {
    if ((event.altKey || event.metaKey) && event.key === 'Enter') {
        event.preventDefault();
        const sidebar = document.getElementById('editor-sidebar');
        if (sidebar.style.display === 'none') {
            sidebar.style.display = 'block';
        } else {
            sidebar.style.display = 'none'; // Hide the sidebar
        }
    }
});

window.addEventListener('popstate', (event) => {
    // event.preventDefault();
    const state = event.state; // Get the state object
    if (state) {
        showFile(state['folder'], state['file'], false);
    }
});

document.getElementById('goToFile').addEventListener('keydown', (event) => {
    const resultsList = document.getElementById('goToFileResults').querySelectorAll('li');

    if (event.key === 'Enter') {
        event.preventDefault();
        if (resultsList[focusedItemIndex]) {
            const [folder, filename] = resultsList[focusedItemIndex].getAttribute('data-path').split('/');
            showFile(folder, filename);
            closeSearchModal();
        }
    }

    if (event.key === 'ArrowDown') {
        event.preventDefault();
        focusedItemIndex = (focusedItemIndex + 1) % resultsList.length;
        updateFocusedItem(resultsList);
    } else if (event.key === 'ArrowUp') {
        event.preventDefault();
        focusedItemIndex = (focusedItemIndex - 1 + resultsList.length) % resultsList.length;
        updateFocusedItem(resultsList);
    }
});