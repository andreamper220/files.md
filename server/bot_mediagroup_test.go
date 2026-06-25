package server

import (
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/zakirullin/files.md/server/fs"
	"github.com/zakirullin/files.md/server/pkg/tg"
	"github.com/zakirullin/files.md/server/pkg/txt"
	"github.com/zakirullin/files.md/server/userconfig"
)

func TestMediaGroupBuffer_FlushesCombinedAttachments(t *testing.T) {
	mediaGroups.mu.Lock()
	mediaGroups.buffers = make(map[string]*mediaGroupBuffer)
	mediaGroups.mu.Unlock()

	r := require.New(t)
	userFS, err := fs.NewFS("/", afero.NewMemMapFs())
	r.NoError(err)
	bot, fakeDB := newTestBot(tg.NewFakeTG(), userFS, userconfig.NewConfig(userFS, -1, "config.yaml"))

	part1 := txt.FormatAttachmentContent("media/a.pdf", "go-cheatsheet-1page.pdf")
	part2 := txt.FormatAttachmentContent("media/b.pdf", "go-livecoding-and-common.pdf")

	r.NoError(bot.bufferMediaGroupContent("grp1", part1, "", nil))
	r.NoError(bot.bufferMediaGroupContent("grp1", part2, "Go docs", nil))

	time.Sleep(mediaGroupCollectDelay + 300*time.Millisecond)

	r.NotEmpty(fakeDB.PendingDrafts)
	var combined string
	for _, content := range fakeDB.PendingDrafts {
		combined = content
	}
	r.Contains(combined, "Go docs")
	r.Contains(combined, "go-cheatsheet-1page.pdf")
	r.Contains(combined, "go-livecoding-and-common.pdf")
}
