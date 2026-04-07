<img src="https://github.com/zakirullin/files.md/raw/main/web/icon.png" alt="Files.md logo" title="Files.md" align="right" height="76" />

# Files.md  
A simple application for your `.md` files.

<img src="https://github.com/zakirullin/files.md/raw/main/web/app.png" alt="Files.md screenshot" title="Files.md"/>

You can store whole your life:
- 📝 Notes, Projects
- ✅ Tasks, Checklists
- 💚 Journal, Habits

All in plain `.md` files, locally.  

You can try it out:
https://app.files.md

## Why?

1) I used `files.md` to grow my knowledge about Software Development and Brain.
2) I added new notes to the base. One idea per note.
3) I made connections between notes. Everything should be connected, just as in our brain.
4) I spend time travelling through notes and thinking it through.
5) Software development and brain-related notes suddenly appeared very related.
6) **I got the insight**.
7) I wrote an article about [Cognitive Load in Software Development](https://github.com/zakirullin/cognitive-load).

Many considered it a great write-up.  

All this activity helped me to:
- **Think deeply**  
- **Write insightful texts**
- **Connect knowledge across domains**  

To achieve all that, **you'll have to use your brain**. Not advanced workflows or AI automations.  

## How?
I reuse long-learned pattern - messaging with a friend.  

Only now I send notes and ideas to the bot.  

And it saves everything to `.md` files nicely.  

## Telegram Bot 🤖
<img src="https://github.com/zakirullin/files.md/raw/main/web/bot.png" alt="Telegram Bot screenshot" title="Telegram Bot"/>

When we are focused and distracting information comes in, we want to get rid of it as quickly as possible.  
To do that, just send whatever is distracting you to the bot. Then choose how you want to save it - as a task, a note, or a journal entry.  

It works like a regular chat, so it's easier to use because there's less resistance.  
We're used to sending messages to friends, now we're going to send stuff to the bot.

## Another note-taking app?
Maybe. What could set it apart, is the simplicity.  
The whole project should fit into one person's head.  

1) Client app is located at `web/`. Offline-first PWA application, with no build systems whatsoever, just open `web/index.html` file.
2) Server is located at `server/`. It 



## Run your own Bot
1) Install [Go](https://go.dev/doc/install)
2) Register new telegram bot via [@BotFather](https://t.me/BotFather)
3) Copy your bot token to `.env` file (see `.env.example`)
4) Run the bot:
```bash
$ go run ./cmd/tgbot
```

Bot's artifacts can be seen in `./storage/<USER_ID>` folder

## Init Server
```bash
TOKENS_SALT=your-secret-salt-here (TODO salt)
$ make init_server host=<YOUR_SSH_HOST>
```

## Repository structure
`/cmd/server` - entrypoint for telegram bot (stable release)  
`/cmd/bot` - entrypoint for local standalone bot (beta version)  
`/internal` - bot's code (reused for both telegram/local bots)  
`/pkg` - various packages   
`/web` - standalone web application for viewing/editing files (alpha version, Chrome only)   


## App 📝
[app.files.md](https://app.files.md), is a standalone application for viewing/editing files, alpha version. Works offline. See `/web/app.html` for more details.

`cmd + k` for command palette.  
`cmd + [` to move back in history, `cmd + ]` to move forward.  
`cmd + enter` to hide/show sidebar.  
`[` to create a link.  
`ctrl + cmd + space` to show emoji dialog.  

## Storage file structure
~~We differentiate the following types of files (with `/` denoting your root folder):
- Tasks: `/today/Pay the bills.md` (`/today/*.md`, `/later/*.md`)
- Notes: `/brain/Brain is the most complex object.md` (`/*/*.md`)
- Files: `/My project.md` (`/*.md`)
- Checklists: `/_read_/How to Take Smart Notes.md` (`/_[read|watch|shop]_/*.md`)
- Journal: `/Journal/2024.08 August.md` (`/journal/<YEAR>.<MONTH> <MONTH NAME>.md`)
- Habits: `/habits/2 minute morning workout.md` (`/habits/*.md`)
- Insights: `/insights/2024 Habits.md` (`/insights/<YEAR> Habits.md`)
- Images: `/img/*`
- Pomodoro: `/today/Finished a break.md`
- Archive: `/archive/*`

## How we contribute
- No long-lived branches except `main`
- Feature branches are [short-lived](https://trunkbaseddevelopment.com/short-lived-feature-branches/)
- **We commit often, so pull `main` every once in a while**
- Once your feature is ready, open a PR to `main`

How to start a feature branch:
```bash
$ git checkout main
$ git pull
$ git checkout -b feature_name
```

## Glossary
- `filename` - a filename with extension, like "note.md" (USE THIS AS ID)
- `header` - an extension-stripped and capitalized filename, like "Note"
- `body` - file's content 
- `dir` - a dir that is meant to store notes under some category, like "happiness"
- `userID` - chatID. For the most part we're only using chatID as userID (PM with the bot)
- `ctime` for file - data blocks or metadata change time: file's ownership, location, file type and permission settings changed time. Parent folder renaming won't affect, moving the file does affect, renaming the file does affect. We need this to track file's location changes, like to understand when it was moved to archive, to track task's angry level etc
- `mtime` for file - mtime (modification time) for a file refers to the time when the contents of the file were last modified. Unlike ctime, it is not affected by changes to the file's metadata, such as ownership, permissions, or renaming
- `ctime` for dir - adding or removing files or subdirectories (similar to `mtime` plus inode changes like renaming files)

Any file can be uniquely identified by filename and dir. We only support one level of nesting.

## Performance
The project is blazing fast :) If you're afraid of using files or mutexes unnecessarily for performance reasons, take a look at this:  
```
Mutex lock/unlock = 25 ns
Read 4K randomly from SSD = 150,000 ns
1 ms = 1,000,000 ns
```

## ADRs (Architecture Decision Records)
- Tried to move web/* stuff in the root folder for simplicity. Bad decision - there should be an explicit dir which we can use as public DOCROOT on our server.
- Switched to [link] for links. The [link](full%20path) syntax is too overwhelming and clunky, plus we don't want to deal with path changes.
- Removed WASM. I had a bug when a message was removed from Inbox.txt, and was not added to a file (I pressed "move to file" button). I wasn't able to reproduce the issue, but what I found is a lot of complexity. JS -> Go (writeFile) -> Go awaiting a promise from JS -> JS Golang runtime somewhere in between -> JS (writeFile) -> Go (returning from promise) -> Sending results back to JS. And it has to be done in a separate goroutine, because both WASM and JS are running in the same thread. Also, Golang's WASM is still experimental. We have too many components and a lot of uncertainty involved. I didn't want to implement same functionality in JS back then, at the solution served for some time. Now it's time to reimplement the functionality in JS and give up all this complexity. Also, inbox.wasm is ~8MB and I wanted the application to be really small.  
- Decided to use OPFS as an initial driver for file system. Better browsers support, less hustle for users. The app starts with OPFS driver by default, if needed, user can replace the driver with Local FileSystem API by opening a local dir. DirHandle would be saved to IndexedDB in such scenario and reused every time.
- Root folder is now '/', not ''. All files in webapp are identified by path, not by 'dir' + 'filename', restricting to 1 level of nesting.
- Dropbox is changing some metadata for newfly created files, thus ctime is changed. I was thinking about moving to mtime for sync, but that wouldn't allow us detect renames (though, we detect them through a separate mechanism anyway), so mtime can be more reliable. Also sync won't be triggered by permission/ownership change etc. Migrated to mtime. Mtime is used for content-based sync, ctime is used for append-only sync log (renames/del).
- Decided to migrate every flow to Chat.md, even todo lists. Added - we can't work with multiline tasks with this flow, we may want support both files and indices. We have two ways of doing so - encode params in a uniform way, and use same command handlers with IFs. Or we can use different command handlers to handle chat/file movements. I decided to go for different command handlers. Added, if we go for different commands - move to buttons config would be complicated. Added, maybe we can move files back to Chat.md on "file move", and reuse the existing flow? Added, so far seems good. Our chat.md log acts as an append-only log. As a bonus, if we don't finish some flow (like schedule/move), the content would be saved in log and we can continue scheduling/moving from the app.
- All incoming messages go to Chat.md now by default. Before that they got moved to `/today` (and become tasks), which was good for a simple todo list, but not as convenient for other use cases. I realized that during meetings, all I needed was a simple input field where I can dump whatever stuff from my head with no further immediate action. With a possibility to review and organize it later. It can be tasks, it can be journal records, or it can be files. Also, it's better to have a really simple easy to understand default flow - we dump all the messages into one file, and that's it.  
- Default mode for chat is "One big file" now, i.e. the only thing it does is dumps all the messages into one file. Again, let's start with the simplest flow, not to overwhelm users. Added later. If we choose full mode, we'll have to create dirs upfront so that "to habits", "to read/shop" etc. would work. If users don't need it, he removes the dirs, and we don't recreate them (as we would do in "on-the-fly mode"). So, we can't use on-the-fly strategy everywhere.
- Before we created all necessary dirs upfront, now we create dirs on the fly. That way we won't clutter user's knowledge base right from the start.
- Switched to microseconds for tracking file changes during sync. Gap between consecutive files creation is more than enough - ranging from 5000μs to 1000μs. We didn't go for nanosec because js is having troubles with int64 precision. Added later. Linux is using cached kernel time, which is updated at `CONFIG_HZ` interval (`grep CONFIG_HZ /boot/config-$(uname -r)`), in my case the value is 1000 (1ms). Most real-world operations operations are spaced much further apart than 1ms due to: user interaction, network latency, disk i/o. We might only have issue if we update files inside an effective/native loop. 
- I believe it's time to make our knowledge base cross-platform, by forbidding characters like ":?<>*" in filenames. These characters aren't allowed in some environments (like Windows, PWA).
- I wanted bot-like functionality in browser. I didn't want to re-write well-tested code in TypeScript, so I used wasm~~. And it worked perfectly good.
- We use Telegram bot as distract-free write-only entrance to our knowledge base. The only issue is, it is not as wildly popular in EU/USA. I've come to the idea that we can transform app.files.md to a chat once we decrease the window size! Would be default behaviour on mobiles.
- Introduced append-only log for syncing. Stateless sync is tricky to implement - we would have to send all files in every request. Since we're only renaming on server - we'll only track renames.
- For content-only sync (no renames/deletes) we don't store any state on server, we compare hashes & last ctimes 
- Removed Wikilinks support. Only plain Markdown links, our knowledge base must be interoperable.
- Updates are now processed sequentially on per-user basis. Because there were some race conditions on concurrent file writings. Also we faced out-of-order forwarded messages processing, and it was impossible to collapse them to one message.
- **Removed fyne.io**. At first, I wanted a lightweight alternative to Electron, and fyne.io seemed to be an ideal candidate. After a few days working with it 80% of bot functionality was implemented, and I was pretty happy with it. The thing is, to implement the rest of the functionality, we would have to apply A TREMENDOUS amount of effort. I am talking tiny details such as scrolling, emojis rendering, text selecting behaviour, links support, etc. And in future we would have to implement image uploading and markdown/html renderer, which would be also painful in such non-webview based toolkit. As much as I hate using the web stack for the desktop applications, it doesn't seem like we have a choice. Let's try wails.io.
- We use vendoring for dependencies. We want all our few dependencies to be in the repo, so we don't care about blocked/removed dependencies. Our repository is the self-sufficient source of truth.
- We use granular locks (in db, journal, userconfig) instead of one global per user lock so to avoid bottlenecks. Workers might use 3rd party API like ChatGPT, and we don't want to hold user's lock all that time. **PATCHED**, we added sequential per-user updates processing, `bot` can't cause RCs on its own, but `bot` & `worker` can, so we should continue using granular locks.
- We read every userconfig value from the config file on every access. We don't need load/save whole config before/after `bot.Answer()` method. We have to reread it every time we need to change it, so we don't write back any stale data. Let's imagine we load config only once before `bot.Answer()`, next, we may have significant networking delays in `bot.Answer()` (let's say 2 seconds when making external requests), there are good changes that during those 2 seconds `worker.MoveDueTasks()` will modify `userconfig.Schedule`, causing data race (after bot's answer we write back stale data). And we don't want our schedule lost.
- Sanitize Early, we gave up sanitizing in Path method. That's an unexpected behaviour - it breaks paths. We should sanitize everything as soon as we received. Most commands work with md5 hashes, for such cases no sanitize is needed
- `gofumpt` for stricter formatting. `gofumpt` is happy with a subset of the formats that gofmt is happy with. The less we have to choose between different formating options, the better
- FS's structure should have userFS name, to reflect the fact it user user-namespaced
- Note term is way too vague. Let's try to use "file" term, without any high level abstraction (like note) 
- Gave up on AST parsing/rendering. We had lots of corner cases via AST and the code was way complex. Markdown isn't that hard to parse, we can do it via good old straigforward code. We have 3x times less code now, and it is far less mentally taxing to understand. We did the same for MD->HTML conversion. Telegram doesn't support whole range of HTML tags, so it was easier to write our own md-to-html converter.
- Adherence to Tolerant Reader principles. If enconunter gibberish during parsing - we skip it, but if we encounter flags of valid data (let's say `###`) but data itself is invalid - we panic. TODO preserve gibberish during read-write cycle.
- Usage of https://github.com/rivo/uniseg. In Go, strings are read-only slices of bytes. They can be turned into Unicode code points using the for loop or by casting: []rune(str). However, multiple code points may be combined into one user-perceived character or what the Unicode specification calls "grapheme cluster". For example, white circle "⚪" has two runes, but one grapheme cluster.
- Markdown to HTML conversion. User can have invalid Markdown in his notes, and TG API would fail to send invalid Markdown directly. So, first we escape HTML, then we convert user's Markdown to HTML and finally send it via Telegram API as HTML.
- File hashing. Everywhere where we have user input - we should use fs.hash, otherwise we get long filenames, and tg returns `INVALID_DATA` error (callbackData max 64 bytes)
- Introduced `db.go`. We had to abstract away Redis anyway (otherwise it's hard to write tests)
- Package db.go doesn't store userID (we often use it separately...) Do we? Maybe we gonna use it without userID (like global bot stats?). Added: moved userID to class. Maybe in later we'll need this class outside of user's scope, but let's stay in the future :)
- We can't ucfist filename in fs.Put - what if that was user-created file (outside the bot), i.e. it comes with lowercase

## Notes about Dropbox
- Symlink created on server will be synced on client as is (without resolving)
- To prevent symlinks attack our storage path should be mounted via `nosymfollow` flag

## Overarching design principles
- `Clarity`: The code’s purpose and rationale is clear to the reader.
- `Simplicity`: The code accomplishes its goal in the simplest way possible.
- `Concision`: The code is easy to discern the relevant details, and the naming and structure guide the reader through these details.  
- `Maintainability`: The code is easy for a future programmer to modify correctly.  
- `Consistency`: The code is consistent across the codebase  

Refer to [the following document](https://github.com/zakirullin/cognitive-load) for more comprehensive guiding rules.

## Guidelines
- We write **tests**
- eXtreme Programming and TDD are highly encouraged
- With portability in mind, everything is stored in **plain text files**
- We don't use get* prefix for methods
- No panics, errors are part of business logic
- If we are ignoring an error - we leave a WHY comment
- We wrap errors all the time, we should add method's context
- No iterators for client code
- We prefer real implementations or at least fakes over mocks and stubs
- Imports should only be renamed to avoid a name collision with other imports

## Front
- Use PATCHED keyword if you modify assets in-place
