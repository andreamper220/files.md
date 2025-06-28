// Global variables
let messages = [];
let chatContainer;
let messageInput;
const CHAT_FILENAME = 'Chat.txt';


function parseFileContent(content) {
    // Normalize line endings
    content = content.replace(/\r\n/g, '\n').replace(/\r/g, '\n');
    const lines = content.split('\n');

    const headerRegex = /^#### /;
    const timestampRegex = /^`\d{2}:\d{2}` /;

    const blocks = [];
    let currentBlock = '';

    for (const line of lines) {
        const isHeader = headerRegex.test(line);
        const isTimestamp = timestampRegex.test(line);

        if (isHeader || isTimestamp) {
            // Save previous block if exists
            if (currentBlock.length > 0) {
                blocks.push(currentBlock.trim());
                currentBlock = '';
            }

            // Start new block
            currentBlock = line;
        } else {
            // Continue current block
            if (currentBlock.length > 0) {
                currentBlock += '\n' + line;
            }
        }
    }

    // Add final block
    if (currentBlock.length > 0) {
        blocks.push(currentBlock.trim());
    }

    // Parse blocks into messages
    const messages = [];
    let currentDate = null;

    for (let i = 0; i < blocks.length; i++) {
        const block = blocks[i];

        // Check if block is a date header
        if (block.startsWith('####')) {
            currentDate = block.replace(/^#+\s*/, '').trim();
            continue;
        }

        // Check if block is a timestamped message
        const timeMatch = block.match(/^`(\d{2}:\d{2})`\s*([\s\S]*)$/);
        if (timeMatch) {
            const [, timestamp, text] = timeMatch;

            if (text.trim()) {
                messages.push({
                    index: i,
                    text: text.trim(),
                    timestamp: timestamp,
                    date: currentDate || new Date().toDateString()
                });
            }
        }
    }

    return messages;
}

function formatFileContent(messages) {
    if (messages.length === 0) return '';

    // Group messages by date
    const messagesByDate = {};
    messages.forEach(msg => {
        const date = msg.date || new Date().toDateString();
        if (!messagesByDate[date]) {
            messagesByDate[date] = [];
        }
        messagesByDate[date].push(msg);
    });

    let content = '';
    Object.entries(messagesByDate).forEach(([date, msgs]) => {
        if (content) content += '\n';
        content += `#### ${date}\n`;
        msgs.forEach(msg => {
            content += `\`${msg.timestamp}\` ${msg.text}\n`;
        });
    });

    return content;
}

async function loadData() {
    try {
        const file = await ((await getFileHandle(CHAT_FILENAME)).getFile());
        const content = await file.text();

        // Parse the content and load messages
        messages = parseFileContent(content);

        console.log(`Loaded ${messages.length} messages from ${CHAT_FILENAME}`);
    } catch (error) {
        console.error('Error loading data:', error);
        // Initialize with empty data if file doesn't exist or can't be read
        messages = [];
    }
}

async function saveData() {
    try {
        // For now, just save the current file's messages
        // You can extend this to save all files
        const content = formatFileContent(sidebarFiles[currentFile]);

        // You'll need to implement the file writing part
        // This is a placeholder for your file system API
        console.log('Would save to file:', content);

        // Example of what the save might look like:
        // const fileHandle = await getFileHandle(CHAT_FILENAME);
        // await fileHandle.write(content);

    } catch (error) {
        console.error('Error saving data:', error);
    }
}

function initChat() {
    chatContainer = document.getElementById('chat');
    messageInput = document.getElementById('chat-input');

    messageInput.addEventListener('keydown', function (e) {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            handleSend();
        }
    });

    loadData().then(() => {
        renderMessages();
    });
}

async function handleSend() {
    const text = messageInput.value.trim();
    if (!text) return;

    // const now = new Date();
    // const note = {
    //     id: Date.now(),
    //     text: text,
    //     timestamp: now.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
    //     date: now.toLocaleDateString('en-GB', {
    //         day: 'numeric',
    //         month: 'long',
    //         weekday: 'long'
    //     })
    // };
    //
    // sidebarFiles[currentFile].push(note);

    messageInput.value = '';
    // saveData();

    reply(text);
    await loadData();
    renderMessages();
    scrollToBottom();

    // Notify sidebar of changes (commented for future implementation)
    // updateSidebar();
}

function renderMessages() {
    if (messages.length === 0) {
        chatContainer.innerHTML = `
            <div class="empty-state">
                <div class="empty-title">What's on your mind?</div>
                <div class="empty-desc">You can dump here whatever thoughts you have</div>
            </div>
        `;
        return;
    }

    chatContainer.innerHTML = messages.map(message => `
        <div class="message" data-index="${message.index}">
            <div class="message-content" 
                 contenteditable="true" 
                 data-index="${message.index}"
                 spellcheck="false">${escapeHtml(message.text)}</div>
            <div class="message-footer">
                <span class="message-time">${message.timestamp}</span>
                <div class="message-actions">
                    <button class="action-btn submenu-btn to-file-btn" data-index="${message.index}">
                        📄
                        <span class="btn-label">To File</span>
                    </button>
                    <button class="action-btn submenu-btn to-dir-btn" data-index="${message.index}">
                        🗂
                        <span class="btn-label">To Dir</span>
                    </button>
                    <button class="action-btn to-todo-btn" data-index="${message.index}">
                        ✅   
                    <span class="btn-label">To Do</span>
                    </button>
                    <button class="action-btn to-read-btn" data-index="${message.index}">
                        📚
                        <span class="btn-label">To Read</span>
                    </button>
                    <button class="action-btn to-shop-btn" data-index="${message.index}">
                        🛒
                        <span class="btn-label">To Shop</span>
                    </button>
                    <button class="action-btn to-watch-btn" data-index="${message.index}">
                        📺
                        <span class="btn-label">To Watch</span>
                    </button>
                    <button class="action-btn journal-btn" data-index="${message.index}">
                        💚
                        <span class="btn-label">To Journal</span>
                    </button>
                    <button class="action-btn delete-btn" data-index="${message.index}">
                        🗑️
                        <span class="btn-label">Delete</span>
                    </button>
                </div>
            </div>
        </div>
    `).join('');

    attachEventListeners();
}


const fileOptions = [
    { id: 'journal-2024.txt', label: '2024' },
    { id: 'journal-2025.txt', label: '2025' },
    { id: 'journal-2026.txt', label: '2026' },
    { id: 'shop-2024.txt', label: 'Shop 24' },
    { id: 'shop-2025.txt', label: 'Shop 25' },
    { id: 'ideas.txt', label: 'Ideas' }
];

const dirOptions = [
    { id: 'folder1', label: 'Folder 1' },
    { id: 'folder2', label: 'Folder 2' },
    { id: 'folder3', label: 'Folder 3' },
    { id: 'archive', label: 'Archive' },
    { id: 'temp', label: 'Temp' },
    { id: 'backup', label: 'Backup' }
];

function attachEventListeners() {
    // Add event listeners for editing message content
    chatContainer.querySelectorAll('.message-content[contenteditable]').forEach(element => {
        element.addEventListener('blur', function (e) {
            saveEdit(e.target.dataset.noteId, e.target.textContent);
            e.target.classList.remove('editing');
        });

        element.addEventListener('focus', function (e) {
            e.target.classList.add('editing');
        });

        element.addEventListener('keydown', function (e) {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                e.target.blur();
            }
            if (e.key === 'Escape') {
                e.target.textContent = messages.find(n => n.id == e.target.dataset.noteId).text;
                e.target.blur();
            }
        });
    });

    // Add delete button listeners
    chatContainer.querySelectorAll('.journal-btn').forEach(btn => {
        btn.addEventListener('click', function (e) {
            e.stopPropagation();
            let cmd = {
                n: "mv_to_journal",
                t: "cmd",
                p: [btn.dataset.index.toString()]
            }
            replyCmd(JSON.stringify(cmd))
        });
    });

    chatContainer.querySelectorAll('.delete-btn').forEach(btn => {
        btn.addEventListener('click', function (e) {
            e.stopPropagation();
            deleteNote(btn.dataset.noteId);
        });
    });


    // Submenu
    document.querySelectorAll('.submenu-btn').forEach(btn => {
        // Add appropriate submenu based on button type
        if (btn.classList.contains('to-file-btn')) {
            addSubmenuToButton(btn, fileOptions, 'data-target-file');
        } else if (btn.classList.contains('to-dir-btn')) {
            addSubmenuToButton(btn, dirOptions, 'data-target-dir');
        }

        // Add click listener
        btn.addEventListener('click', (e) => {
            e.stopPropagation();
            const submenu = btn.querySelector('.btn-submenu');

            // Close other open submenus
            document.querySelectorAll('.btn-submenu.show').forEach(menu => {
                if (menu !== submenu) menu.classList.remove('show');
            });

            // Toggle current submenu
            submenu.classList.toggle('show');
        });
    });


    // Prevent submenu from closing when clicking inside it
    document.querySelectorAll('.btn-submenu').forEach(submenu => {
        submenu.addEventListener('click', (e) => {
            e.stopPropagation();
        });
    });

    // Handle submenu item clicks
    document.querySelectorAll('.submenu-item').forEach(item => {
        item.addEventListener('click', (e) => {
            e.stopPropagation();
            const messageIndex = item.closest('.submenu-btn').dataset.index;
            const targetFile = item.dataset.targetFile;
            // Your move logic here
            console.log(`Moving message ${messageIndex} to ${targetFile}`);
        });
    });
}


function saveEdit(noteId, newText) {
    const note = messages.find(n => n.id == noteId);
    if (note && newText.trim() !== '') {
        note.text = newText.trim();
        saveData();
    }
}

function deleteNote(noteId) {
    messages = messages.filter(n => n.id != noteId);
    saveData();
    renderMessages();
}

function scrollToBottom() {
    setTimeout(function () {
        chatContainer.scrollTop = chatContainer.scrollHeight;
    }, 100);
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Initialize when page loads
// document.addEventListener('DOMContentLoaded', init);

const chatInput = document.getElementById('chat-input');

function autoResize() {
    chatInput.style.height = 'auto';
    chatInput.style.height = Math.min(chatInput.scrollHeight, 250) + 'px';
}

// Add event listener for input changes
chatInput.addEventListener('input', autoResize);
// Initial resize to set proper height
autoResize();

// Add click listeners for submenu buttons

// Handle submenu item clicks
document.addEventListener('click', (e) => {
    if (e.target.classList.contains('submenu-item')) {
        e.stopPropagation();
        const messageIndex = e.target.closest('.submenu-btn').dataset.index;
        const targetFile = e.target.dataset.targetFile;
        const targetDir = e.target.dataset.targetDir;

        console.log(`Moving message ${messageIndex} to ${targetFile || targetDir}`);

        // Close submenu after selection
        e.target.closest('.btn-submenu').classList.remove('show');
    }
});

// Close submenus when clicking outside
document.addEventListener('click', () => {
    document.querySelectorAll('.btn-submenu.show').forEach(menu => {
        menu.classList.remove('show');
    });
});


function createSubmenu(options, dataAttribute) {
    return `
        <div class="btn-submenu">
            ${options.map(option =>
        `<div class="submenu-item" ${dataAttribute}="${option.id}">${option.label}</div>`
    ).join('')}
        </div>
    `;
}

// Function to add submenu to button
function addSubmenuToButton(button, options, dataAttribute) {
    // Remove existing submenu if any
    const existingSubmenu = button.querySelector('.btn-submenu');
    if (existingSubmenu) {
        existingSubmenu.remove();
    }

    // Add new submenu
    button.insertAdjacentHTML('beforeend', createSubmenu(options, dataAttribute));
}