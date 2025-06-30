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

    // TODO write clearer way
    let numblocks = 0
    for (let i = 0; i < blocks.length; i++) {
        const block = blocks[i];

        // Check if block is a date header
        if (block.startsWith('####')) {
            currentDate = block.replace(/^#+\s*/, '').trim();
            numblocks++;
            continue;
        }

        // Check if block is a timestamped message
        const timeMatch = block.match(/^`(\d{2}:\d{2})`\s*([\s\S]*)$/);
        if (timeMatch) {
            const [, timestamp, text] = timeMatch;

            if (text.trim()) {
                messages.push({
                    index: i - numblocks,
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
        const file = await ((await getFileHandle(CHAT_FILENAME, true)).getFile());
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
        const content = formatFileContent(files[currentFile]);

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
            autoResize();
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
                    <button class="action-btn to-file-btn" data-index="${message.index}">
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
                    <button class="action-btn to-checklist-btn" data-index="${message.index}" data-dir="_read_">
                        📚
                        <span class="btn-label">To Read</span>
                    </button>
                    <button class="action-btn to-checklist-btn" data-index="${message.index}" data-dir="_shop_">
                        🛒
                        <span class="btn-label">To Shop</span>
                    </button>
                    <button class="action-btn to-checklist-btn" data-index="${message.index}" data-dir="_watch_">
                        📺
                        <span class="btn-label">To Watch</span>
                    </button>
                    <button class="action-btn to-journal-btn" data-index="${message.index}">
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

function attachEventListeners() {
    chatContainer.addEventListener('mousedown', function(e) {
        // If clicking outside messages, prepare for multi-select
        if (!e.target.closest('.message')) {
            let allMessages = Array.from(chatContainer.querySelectorAll('.message'));
            let startMessage = null;

            function handleMouseMove(e) {
                const currentMessage = e.target.closest('.message');
                if (currentMessage) {
                    document.getSelection().removeAllRanges(); // Prevent text selection

                    if (!startMessage) {
                        startMessage = currentMessage;
                        document.querySelectorAll('.message.selected').forEach(m => m.classList.remove('selected'));
                        currentMessage.classList.add('selected');
                    } else if (currentMessage !== startMessage) {
                        // Select range like normal message selection
                        const startIndex = allMessages.indexOf(startMessage);
                        const endIndex = allMessages.indexOf(currentMessage);
                        const minIndex = Math.min(startIndex, endIndex);
                        const maxIndex = Math.max(startIndex, endIndex);

                        document.querySelectorAll('.message.selected').forEach(m => m.classList.remove('selected'));

                        for (let i = minIndex; i <= maxIndex; i++) {
                            allMessages[i].classList.add('selected');
                        }
                    }
                }
            }

            function handleMouseUp() {
                document.removeEventListener('mousemove', handleMouseMove);
                document.removeEventListener('mouseup', handleMouseUp);
            }

            document.addEventListener('mousemove', handleMouseMove);
            document.addEventListener('mouseup', handleMouseUp);
            return;
        }

        const message = e.target.closest('.message');
        if (!message || e.target.closest('.message-actions')) {
            return;
        }

        if (isMetaKey(e)) {
            message.classList.toggle('selected');
            return;
        }

        document.querySelectorAll('.message.selected').forEach(m => m.classList.remove('selected'));
        message.classList.add('selected');

        let startMessage = message;
        let allMessages = Array.from(chatContainer.querySelectorAll('.message'));

        function handleMouseMove(e) {
            const currentMessage = e.target.closest('.message');
            if (currentMessage && currentMessage !== startMessage) {
                document.getSelection().removeAllRanges();

                const startIndex = allMessages.indexOf(startMessage);
                const endIndex = allMessages.indexOf(currentMessage);
                const minIndex = Math.min(startIndex, endIndex);
                const maxIndex = Math.max(startIndex, endIndex);

                document.querySelectorAll('.message.selected').forEach(m => m.classList.remove('selected'));

                for (let i = minIndex; i <= maxIndex; i++) {
                    allMessages[i].classList.add('selected');
                }
            }
        }

        function handleMouseUp() {
            document.removeEventListener('mousemove', handleMouseMove);
            document.removeEventListener('mouseup', handleMouseUp);
        }

        document.addEventListener('mousemove', handleMouseMove);
        document.addEventListener('mouseup', handleMouseUp);
    });

    chatContainer.addEventListener('click', function(e) {
        // Only clear selection if clicking outside messages AND not dragging
        if (!e.target.closest('.message') && !e.detail > 1) {
            document.querySelectorAll('.message.selected').forEach(m => m.classList.remove('selected'));
        }
    });

    // Add event listeners for editing message content
    // chatContainer.querySelectorAll('.message-content[contenteditable]').forEach(element => {
    //     element.addEventListener('blur', function (e) {
    //         saveEdit(e.target.dataset.noteId, e.target.textContent);
    //         e.target.classList.remove('editing');
    //     });
    //
    //     element.addEventListener('focus', function (e) {
    //         e.target.classList.add('editing');
    //     });
    //
    //     element.addEventListener('keydown', function (e) {
    //         if (e.key === 'Enter' && !e.shiftKey) {
    //             e.preventDefault();
    //             e.target.blur();
    //         }
    //         if (e.key === 'Escape') {
    //             e.target.textContent = messages.find(n => n.id == e.target.dataset.noteId).text;
    //             e.target.blur();
    //         }
    //     });
    // });

    chatContainer.querySelectorAll('.to-file-btn').forEach(btn => {
        btn.addEventListener('click', function (e) {
            e.stopPropagation();
            searchModal.open('', btn.dataset.index, e.target)
            chatInput.focus();
        });
    });

    chatContainer.querySelectorAll('.to-journal-btn').forEach(btn => {
        btn.addEventListener('click', function (e) {
            e.stopPropagation();
            const selectedMessages = document.querySelectorAll('.message.selected');
            let indices = [];
            let messagesToRemove = [];
            if (selectedMessages.length > 0) {
                indices = Array.from(selectedMessages).map(msg => msg.dataset.index);
                messagesToRemove = selectedMessages;
            } else {
                indices = [btn.dataset.index];
                messagesToRemove = [btn.closest('.message')];
            }

            sendCmd('mv_to_journal', indices);
            messagesToRemove.forEach(message => {
                message.classList.add('removing');
                setTimeout(() => {
                    message.remove();
                }, 300);
            });
            chatInput.focus();
        });
    });

    chatContainer.querySelectorAll('.to-checklist-btn').forEach(btn => {
        btn.addEventListener('click', function (e) {
            e.stopPropagation();
            const selectedMessages = document.querySelectorAll('.message.selected');
            let indices = [];
            let messagesToRemove = [];
            if (selectedMessages.length > 0) {
                indices = Array.from(selectedMessages).map(msg => msg.dataset.index);
                messagesToRemove = selectedMessages;
            } else {
                indices = [btn.dataset.index];
                messagesToRemove = [btn.closest('.message')];
            }

            sendCmd('mv_to_chk', [indices.join(','), btn.dataset.dir]);
            messagesToRemove.forEach(message => {
                message.classList.add('removing');
                setTimeout(() => {
                    message.remove();
                }, 300);
            });
            chatInput.focus();
        });
    });

    chatContainer.querySelectorAll('.delete-btn').forEach(btn => {
        btn.addEventListener('click', function (e) {
            e.stopPropagation();
            deleteNote(btn.dataset.noteId);
        });
    });

    // Enable editing on double-click
    chatContainer.querySelectorAll('.message-content').forEach(content => {
        content.addEventListener('dblclick', function(e) {
            e.stopPropagation();
            this.style.pointerEvents = 'auto';
            this.classList.add('editing');
            this.focus();
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

function autoResize() {
    if (chatInput.value === '') {
        chatInput.style.height = '';
        return;
    }

    if (chatInput.value.split('\n').length <= 1) {
        return;
    }

    chatInput.style.height = 'auto';
    chatInput.style.height = Math.min(chatInput.scrollHeight, 250) + 'px';
}

// Add event listener for input changes
chatInput.addEventListener('input', autoResize);
// Initial resize to set proper height
autoResize();


function sendCmd(cmd, params) {
    let cmdObj = {
        n: cmd,
        t: "cmd",
        p: params.map(p => p.toString()),
    }
    replyCmd(JSON.stringify(cmdObj));
}