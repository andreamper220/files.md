<img src="https://github.com/zakirullin/stuff-bot/raw/main/docs/go.svg" alt="Stuff Bot logo" title="Stuff Bot" align="right" height="60" />

# Stuff Bot
A telegram bot for your personal stuff. Everything is stored in plain text files.  

[Tasks management bot showcase in Dorofeev's club](https://club.mnogosdelal.ru/post/180/)  
[Notes taking via bot](https://vas3k.club/post/18815/)

## Spin it up 🚀
1) Install [Go](https://go.dev/doc/install)
2) Register new telegram bot via [@BotFather](https://t.me/BotFather)
3) Copy your bot token to `.env` file (see `.env.example`)

```bash
$ make install && make run
```
or
```bash
$ go get ./..
$ go run ./cmd
```

Bot's artifacts can be seen in `storage/<USER_ID>` folder

## How we contribute
- No long-lived branches except `main`
- Feature branches are [short-lived](https://trunkbaseddevelopment.com/short-lived-feature-branches/)
- **We commit often, so pull `main` every once in a while**
- Once your feature is ready, open a PR to `main`

How to start a feature branch:
```bash
$ git checkout main
$ git pull
$ git checkout -b feature/feature_name
```

## Glossary
- `filename` - a filename with extension, like "note.md" (USE THIS AS ID)
- `title` - an extension-stripped and capitalized filename, like "Note"
- `content` - note's content (body/text)
- `dir` - a dir that is meant to store notes under some category, like "happiness"
- `userID` - chatID. For the most part we're only using chatID as userID (PM with the bot)
- `ctime` -  data blocks or metadata change time: file's ownership, location, file type and permission settings changed time.  Parent folder renaming won't affect, moving the file does affect, renaming the file does affect. We need this to track file's location changes, like to understand when it was moved to archive

Any file can be uniquely identified by filename and dir. We only support one level of nesting.

We differentiate the following types of files (with `/` denoting your root folder):
- Tasks: `/today/Pay the bills.md` (`/today/*.md`, `/later/*.md`)
- Files: `/My project.md` (`/*.md`)
- Notes: `/brain/Brain is the most complex object.md` (`/*/*.md` also `/inbox/*.md`)
- Checklists: `/-read-/How to Take Smart Notes.md` (`/-[read|watch|shop]-/*.md`)
- Journal: `/Journal/2024.08 August.md` (`/journal/<YEAR>.<MONTH> <MONTH NAME>.md`)
- Habits: `/habits/2 minute morning workout.md` (`/habits/*.md`)
- Insights: `/insights/2024 Habits.md` (`/insights/<YEAR> Habits.md`)
- Images: `/img/*`

## ADRs (Architecture Decision Records)
- We read every userconfig value from the config file on every access. We don't need load/save whole config before/after `bot.Answer()` method. We have to reread it every time we need to change it, so we don't write back any stale data. Let's imagine we load config only once before `bot.Answer()`, next, we may have significant networking delays in `bot.Answer()` (let's say 2 seconds when making external requests), there are good changes that during those 2 seconds `worker.MoveDueTasks()` will modify `userconfig.Schedule`, causing data race (after bot's answer we write back stale data). And we don't want our schedule lost.
- Sanitize Early, we gave up sanitizing in Path method. That's an unexpected behaviour - it breaks paths. We should sanitize everything as soon as we received. Most commands work with md5 hashes, for such cases no sanitize is needed
- `gofumpt` for stricter formatting. `gofumpt` is happy with a subset of the formats that gofmt is happy with. The less we have to choose between different formating options, the better
- FS's structure should have userFS name, to reflect the fact it user user-namespaced
- Note term is way too vague. Let's try to use "file" term, without any high level abstraction (like note) 
- Gave up on AST parsing/rendering. We had lots of corner cases via AST and the code was way complex. Markdown isn't that hard to parse, we can do it via good old straigforward code. We have 3x times less code now, and it is far less mentally taxing to understand. We did the same for MD->HTML conversion. Telegram doesn't support whole range of HTML tags, so it was easier to write our own md-to-html converter.
- Adherence to Tolerant Reader principles. If enconunter gibberish during parsing - we skip it, but if we encounter flags of valid data (let's say `###`) but data itself is invalid - we panic. TODO preserve gibberish during read-write cycle.
- Usage of https://github.com/rivo/uniseg. In Go, strings are read-only slices of bytes. They can be turned into Unicode code points using the for loop or by casting: []rune(str). However, multiple code points may be combined into one user-perceived character or what the Unicode specification calls "grapheme cluster". For example, white circle "⚪" has two runes, but one grapheme cluster.
- Markdown to HTML conversion. User can have invalid Markdown in his notes, and TG API would fail to send invalid Markdown directly. So, we need to convert user's Markdown to HTML first and then send it via Telegram as HTML.
- File hashing. Everywhere where we have user input - we should use fs.hash, otherwise we get long filenames, and tg returns `INVALID_DATA` error (callbackData max 64 bytes)
- Introduced `db.go`. We had to abstract away Redis anyway (otherwise it's hard to write tests)
- Package db.go doesn't store userID (we often use it separately...) Do we?
- We can't ucfist filename in fs.Put - what if that was user-created file (outside the bot), i.e. it comes with lowercase

## Notes about Dropbox
- Symlink created on server will be synced on client as is (without resolving)
- Typical file operations usually resolve symlinks so it is vulnerable, and we should use isSafe every time

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
- No iterators for client code
- If we are ignoring an error - we leave a WHY comment
- We wrap errors all the time, we should add method's context
- We prefer fakes/real implementations over mocks and stubs
- Imports should only be renamed to avoid a name collision with other imports
- We use Is/Has prefixes for boolean variables