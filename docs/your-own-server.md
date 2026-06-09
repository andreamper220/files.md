# Run your own server

## Containerized deployment (Docker/Podman)

### Quick start

```bash
cp .env.example .env
# Edit .env: BOT_API_TOKEN, KIE_API_KEY, TOKENS_SALT

docker compose up --build
```

Open [http://localhost](http://localhost) for the PWA. Message your Telegram bot to link your account.

| Variable | Purpose |
|----------|---------|
| `BOT_API_TOKEN` | Telegram bot from [@BotFather](https://t.me/BotFather) |
| `KIE_API_KEY` | Voice transcription via [kie.ai](https://kie.ai/api-key) |
| `TOKENS_SALT` | Random string for PWA sync tokens (`openssl rand -base64 32`) |
| `APP_URL` / `API_URL` | Public URL, `http://localhost` for local Docker |
| `HTTP_PORT` | Host port (default `80`) |
| `STORAGE_QUOTA_KB` | `0` = unlimited (recommended for self-hosted) |

Data persists in Docker volumes `storage` and `tokens`.

### Import Obsidian vault into Docker storage

After the first message to the bot, note your Telegram user ID (folder name in `storage/`).

**Windows (PowerShell):**
```powershell
docker compose build
docker compose run --rm `
  -v "${PWD}/storage:/app/storage" `
  -v "C:\Users\ADMIN\Documents\obsidian:/obsidian:ro" `
  --entrypoint /app/importobsidian `
  files-md `
  --src /obsidian --dst /app/storage/YOUR_TELEGRAM_ID
```

**Linux/macOS:**
```bash
make docker_import OBSIDIAN_SRC=/path/to/obsidian USER_ID=YOUR_TELEGRAM_ID
```

Add `--dry-run` before `--src` to preview without writing.

### Enable HTTPS
In `compose.yaml`: set `CERT_DIR` to a persistent path, uncomment the `"443:443"` port.


## Deploy on your own server (manual)

Install [Go](https://go.dev/doc/install) on your host machine.  

Initialize server with folders and systemd service. Tested on Debian-based systems:
```bash
$ make init_server host=user@example.com salt=$(head -c 32 /dev/urandom | base64)
```

Configure the `/app/.env` file:
```
BOT_API_TOKEN=<TELEGRAM_API_TOKEN_IF_NEEDED>
STORAGE_DIR=/app/storage
CERT_DIR=/opt/files.md
TOKENS_DIR=/opt/files.md/tokens
LOG_FILE=server.log
API_URL=https://api.yourdomain.com
APP_URL=https://app.youdomain.com
```

Deploy a systemd service:
```bash
$ make deploy_systemd host=<YOUR_SSH_HOST>
```

That's all :)  

## Run your own Telegram Bot
1) Install [Go](https://go.dev/doc/install)
2) Register new telegram bot via [@BotFather](https://t.me/BotFather)
3) Add `BOT_API_TOKEN=<YOUR_TELEGRAM_API_TOKEN>` line to `.env` file
4) Redeploy/relaunch the server

Bot's artifacts can be seen in `./storage/<USER_ID>` folder.  

## Linking a new device
1) Open telegram bot
2) Open `/app`
3) Open the link in your browser
4) Device is now linked

### Additional bot's settings
1) For search functionality, enable `Inline Mode` for your bot in [@BotFather](https://t.me/BotFather)
2) Press "Edit Commands", and send the following list:
```
chat - 🏠 Home
files - 📄 Files
dirs - 🗂 Dirs
checklists - ☑️ Checklists
schedule - 📆 Schedule
postpone - 🦥 Postpone
rename - ✏️ Rename
move - ➡️ Move
app - 🔗 Open in app
settings - ⚙️ Settings
help - 📕 Help
```

## Hosting the bot on you local computer
You can host the bot locally, because it doesn't expose any ports to the outside world (if you don't use habits functionality).  
It communicates with Telegram using pull API.

Create a symlink to your local folder with `.md` files for convenience:  
`ln -s <YOUR_EXISTING_DIR_WITH_MD_FILES> storage/<USER_ID>`

## Transfer files to another server

1) Backup your data (`/app/storage`)
2) Be sure that all client app fully synced with the server (bring the app in the focus)
3) Stop bot on old server, so no new files would be created.
4) Compress all the files on one server: `tar -czvf storage.tar.gz storage`
5) `scp` the file to your host machine: `scp SSH_HOST:/app/storage.tar.gz .`
6) `scp` the file to your target machine

Synchronization is relying on `mtime`, so after compressing/decompressing the flag wouldn't be lost.

1) `cd /opt/files.md`
2) `tar -czvf tokens.tar.gz tokens`
3) `scp` to same dir on target machine

We don't need to transfer fslog (renames), if we're certain that all clients read the log.

1) Extract all files on new server
2) Transfer `BOT_API_TOKEN`
3) Launch server
4) Execute `localStorage.setItem('ApiHost', 'YOUR_NEW_API_HOST');` in your PWA applications
5) Make sure that all files are available
6) Cleanup the oldserver

## Maintenance notes
Add this to your crontab (`crontab -e`) for daily git backups:
`0 0 * * * cd /app/storage/<YOUR_TELEGRAM_ID> && git add . && git commit -m "$(date +\%d.\%m.\%Y)"`

Execute `git init` in your folder before that, to init a git repository.

If you have non-ASCI character in filenames, disable quoting:
`git config --global core.quotePath false`

Systemd journal:  
`sudo journalctl -u filesmd`

Find forbidden character in filenames (can be executed in user's storage folder):
`find . -name '*[<>:"|\?*]*'`

Remove forbidden filename characters:
```bash
find . -type f -name '*[<>:"|\?*]*' -print0 | while IFS= read -r -d '' f; do
  dir=$(dirname "$f")
  base=$(basename "$f")
  newbase="${base//[<>:\"|\\?*]/}"
  [ "$base" != "$newbase" ] && [ -n "$newbase" ] && mv -n -- "$f" "$dir/$newbase"
done
```
