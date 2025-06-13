function sendMessage() {
    const msg = input.value.trim();
    if (msg === '') return;

    addMessage(msg, 'user');

    if (msg.startsWith('/')) {
        replyCmd(msg);
    } else {
        (async () => {
            reply(msg);
            // let update = await window.newUpdate(msg, null)
            // processResponse(await window.send(update));
        })();
    }

    input.value = '';
    input.rows = 1;
    sendButton.disabled = true;
    document.querySelector('#messages').scrollTop = messagesContainer.scrollHeight;
}

async function sendCommand(command) {
    // let update = await window.newUpdate('', await window.newCmd(command.substring(1), null));
    // processResponse(await window.send(update));
}

function processResponse(response) {
    response = JSON.parse(response);
    document.querySelectorAll('.bot').forEach(element => element.remove());
    document.querySelectorAll('.button-container').forEach(element => element.remove());
    response.Messages?.forEach(message => {
        const bubble = addMessage(message.Text, 'bot');
        document.querySelector('#messages').appendChild(bubble);
        if (message.Buttons && message.Buttons.length > 0) {
            attachKeyboard(message.Buttons)
        }
    });
    input.focus();
}

function attachKeyboard(buttons) {
    console.log(JSON.stringify(buttons));

    const buttonContainer = document.createElement('div');
    buttonContainer.classList.add('button-container');

    buttons.forEach((row) => {
        if (Array.isArray(row)) {
            const rowContainer = document.createElement('div');
            rowContainer.classList.add('button-row');

            row.forEach((btn) => {
                const button = document.createElement('button');
                button.innerText = btn.Name;
                button.classList.add('telegram-button'); // Add a class for styling
                button.onclick = async () => {
                    // let update = await window.newUpdate('', btn.Cmd)
                    // processResponse(await window.send(update));
                    replyCmd(JSON.stringify(btn.Cmd))
                };
                rowContainer.appendChild(button);
            });
            buttonContainer.appendChild(rowContainer);
        } else {
            const button = document.createElement('button');
            button.innerText = row.Name;
            button.classList.add('telegram-button'); // Add a class for styling
            button.onclick = async () => {
                replyCmd(JSON.stringify(row.Cmd))
                // let update = await window.newUpdate('', row.Cmd)
                // processResponse(await window.send(update));
            };
            buttonContainer.appendChild(button);
        }
    });

    document.querySelector('#messages').appendChild(buttonContainer);
    const messagesContainer = document.querySelector('#messages');
    messagesContainer.scrollTop = messagesContainer.scrollHeight;
}

function addMessage(text, author) {
    const urlPattern = /(\bhttps?:\/\/[^\s]+)/g;
    text = text.replace(urlPattern, '<a href="$1">$1</a>');

    const messageBubble = document.createElement('div');
    messageBubble.classList.add('message', author);
    messageBubble.innerHTML = text.replace(/\n/g, '<br>');
    messagesContainer.appendChild(messageBubble);

    return messageBubble
}

const input = document.getElementById('input-field');
const inputContainer = document.getElementById('input-container');
const sendButton = document.getElementById('send-button');
const messagesContainer = document.getElementById('messages');
const commandPopup = document.getElementById('command-popup');

const commands = [
    {command: '/today', display: '🏠 Today'},
    {command: '/files', display: '📄 Files'},
    {command: '/dirs', display: '🗂 Dirs'},
    {command: '/checklists', display: '☑️ Checklists'},
    {command: '/schedule', display: '📆 Schedule'},
    {command: '/stats', display: '📊 Stats'},
    {command: '/postpone', display: '🦥 Postpone'},
    {command: '/rename', display: '✏️ Rename'},
    {command: '/move', display: '➡️ Move'},
    {command: '/settings', display: '⚙️ Settings'},
    {command: '/help', display: '📕 Help'}
];

// function showCommandPopup() {
//     const inputText = input.value.trim();
//
//     // Filter the commands based on the input text
//     const filteredCommands = commands.filter(cmd => cmd.command.startsWith(inputText));
//
//     // Populate the popup with filtered commands
//     commandPopup.innerHTML = '';
//     filteredCommands.forEach((cmd) => {
//         const cmdItem = document.createElement('div');
//         cmdItem.classList.add('command-item');
//         cmdItem.textContent = cmd.display;
//
//         cmdItem.onclick = () => {
//             input.value = cmd.command; // Set the command input to the actual command (e.g., /today)
//             input.focus();
//             sendMessage();
//             hideCommandPopup();
//         };
//
//         commandPopup.appendChild(cmdItem);
//     });
//
//     if (filteredCommands.length > 0) {
//         commandPopup.style.display = 'block';
//     } else {
//         hideCommandPopup();
//     }
// }

function showCommandPopup() {
    const inputText = input.value.trim();

    const filteredCommands = commands.filter(cmd => cmd.command.startsWith(inputText));

    commandPopup.innerHTML = '';
    filteredCommands.forEach((cmd, index) => {
        const cmdItem = document.createElement('div');
        cmdItem.classList.add('command-item');
        cmdItem.textContent = cmd.display;
        cmdItem.setAttribute('data-index', index);

        cmdItem.onclick = () => {
            input.value = cmd.command;
            input.focus();
            sendMessage();
            hideCommandPopup();
        };

        commandPopup.appendChild(cmdItem);
    });

    if (filteredCommands.length === 0) {
        hideCommandPopup();
        return;
    }

    const firstCommandItem = commandPopup.querySelector('.command-item');
    if (firstCommandItem) {
        firstCommandItem.classList.add('focused');
    }

    commandPopup.classList.remove('hidden');

}

function insertFocusedCommand() {
    const focusedCommand = commandPopup.querySelector('.command-item.focused');
    if (focusedCommand) {
        const commandText = focusedCommand.textContent.trim();
        const selectedCommand = commands.find(cmd => cmd.display === commandText);
        if (selectedCommand) {
            input.value = selectedCommand.command; // Insert the actual command into the input
            input.focus();
            hideCommandPopup();
        }
    }
}

function hideCommandPopup() {
    commandPopup.classList.add('hidden');
    commandPopup.innerHTML = ''; // Clear the popup content
}

window.onload = () => {
    sendCommand('/today')
    input.focus();
};

document.addEventListener('scroll', () => {
    input.focus();
});

input.addEventListener('input', () => {
    // When paste is happening, there could be more that one line (even not separated by \n)
    while (input.scrollHeight > input.clientHeight && input.rows < 30) {
        input.rows += 1;
    }

    input.style.height = 'auto';
    sendButton.disabled = input.value.trim() === '';
    if (input.value.trim() === '') {
        input.rows = 1
    }

    if (input.value.startsWith('/')) {
        showCommandPopup();
    } else {
        hideCommandPopup();
    }
});

input.addEventListener('keydown', (event) => {
    if (event.key === 'Enter' && !event.shiftKey) {
        event.preventDefault();
        if (input.value.startsWith('/')) {
            insertFocusedCommand();
        } else {
            sendMessage();
        }
    }

    if (event.key === 'Enter') {
        if (event.shiftKey) {
            event.preventDefault();
            input.value += '\n';
            input.rows += 1;
        } else {
            event.preventDefault();
            sendMessage();
        }
    }
});

sendButton.addEventListener('click', sendMessage);

document.body.addEventListener('click', function (e) {
    if (e.target && e.target.nodeName == 'A' && e.target.href) {
        const url = e.target.href;
        if (
            !url.startsWith('http://#') &&
            !url.startsWith('https://#') &&
            !url.startsWith('file://') &&
            !url.startsWith('http://wails.localhost:')
        ) {
            e.preventDefault();
            window.runtime.BrowserOpenURL(url);
        }
    }
});

window.addEventListener("focus", async () => {
    input.focus();
});

function isModifierKey(event) {
    return event.metaKey || event.ctrlKey || event.altKey;
}

document.addEventListener('keydown', function (event) {
    if (isModifierKey(event) && event.key === 'Enter') {
        event.preventDefault();

        window.location.href = '/';
        // setTimeout(() => {
        window.resizeTo(screen.availWidth, screen.availHeight);
        window.moveTo(0, 0);
        // }, 100);
    }
});

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

async function openDir() {
    let dirHandle = await window.showDirectoryPicker();
    await saveDirectoryHandle(dirHandle);
    await dirHandle.getFileHandle('test.txt', {
        create: true
    });
}

async function getRootDirHandle() {
    const db = await initDB();
    return new Promise((resolve, reject) => {
        const transaction = db.transaction('handles', 'readonly');
        const store = transaction.objectStore('handles');
        const request = store.get('savedDirectoryHandle');
        request.onsuccess = () => resolve(request.result);
        request.onerror = () => reject(request.error);
    });
}

async function saveDirectoryHandle(directoryHandle) {
    const db = await initDB();
    const transaction = db.transaction('handles', 'readwrite');
    const store = transaction.objectStore('handles');
    await store.put(directoryHandle, 'savedDirectoryHandle');
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
    if (currentDirHandle === undefined) {
        console.error('PLS open dir')
    }

    for (const dirName of dirs) {
        if (dirName) {
            currentDirHandle = await currentDirHandle.getDirectoryHandle(dirName, {create: create});
        }
    }

    let fileHandle;
    fileHandle = await currentDirHandle.getFileHandle(filename, {create: create});

    return fileHandle;
}

async function read(args) {
    let path = args[0];
    let fileHandle = await getFileHandle(path)
    let file = await fileHandle.getFile();

    return await file.text();
}

async function write(args) {
    let path = args[0];
    let content = args[1];

    let fileHandle = await getFileHandle(path, true);
    const writable = await fileHandle.createWritable();
    await writable.write(content);
    await writable.close();
}

async function exists(args) {
    let path = args[0]
    if (path === "") {
        return true
    }

    try {
        await getFileHandle(path);
        return true;
    } catch (error) {
        if (error.name === 'NotFoundError') {
            return false
        }
// TODO better way to handle dirs
        if (error.name === 'TypeMismatchError') {
            return true;
        }
        console.log("EXISTS:", error, "PATH: ", path);
        throw error
    }
}

async function readDir(args) {
    let path = args[0];
    const dirs = path.split('/');
    let currentDirHandle = await getRootDirHandle();
    if (currentDirHandle === undefined) {
        console.error('PLS open dir');
        return [];
    }

// Navigate to the target directory
    for (const dirName of dirs) {
        if (dirName) {
            currentDirHandle = await currentDirHandle.getDirectoryHandle(dirName, {create: false});
        }
    }

    const entries = [];

    for await (const [name, handle] of currentDirHandle.entries()) {
        if (handle.kind === 'directory') {
// For directories
            entries.push({
                name: name,
                isDir: true,
                modTime: null // File System Access API doesn't provide modTime for directories
            });
        } else if (handle.kind === 'file') {
// For files
            const file = await handle.getFile();
            entries.push({
                name: name,
                isDir: false,
                modTime: file.lastModified // Returns timestamp in milliseconds
            });
        }
    }

    return entries;
}

async function mkdir(args) {
    let path = args[0];
    console.log(path);
    try {
        let currentDirHandle = await getRootDirHandle();
        await currentDirHandle.getDirectoryHandle(path, {create: true});
    } catch (error) {
        console.log(error);
        throw error;
    }
}

async function mkdirAll(args) {
    let path = args[0]
    const dirs = path.split('/');
    let currentDirHandle = await getRootDirHandle();
    for (const dirName of dirs) {
        if (dirName) {
            await mkdir([path])
        }
    }
}

function receive(val) {
    processResponse(val)
    console.log(val);
}