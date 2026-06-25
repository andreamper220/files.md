package server

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/zakirullin/files.md/server/pkg/txt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const mediaGroupCollectDelay = 600 * time.Millisecond

type mediaGroupCollector struct {
	mu      sync.Mutex
	buffers map[string]*mediaGroupBuffer
}

var mediaGroups mediaGroupCollector

func init() {
	mediaGroups.buffers = make(map[string]*mediaGroupBuffer)
}

type mediaGroupBuffer struct {
	mu      sync.Mutex
	bot     *Bot
	parts   []string
	caption string
	timer   *time.Timer
}

func (b *Bot) bufferMediaGroupContent(groupID, part, caption string, captionEntities []tgbotapi.MessageEntity) error {
	if caption != "" {
		caption = strings.TrimSpace(txt.TelegramEntitiesToMarkdown(caption, captionEntities))
	}

	key := fmt.Sprintf("%d:%s", b.userID, groupID)
	mediaGroups.mu.Lock()
	buf, ok := mediaGroups.buffers[key]
	if !ok {
		buf = &mediaGroupBuffer{bot: b}
		mediaGroups.buffers[key] = buf
	}
	mediaGroups.mu.Unlock()

	buf.mu.Lock()
	defer buf.mu.Unlock()

	buf.parts = append(buf.parts, part)
	if caption != "" {
		buf.caption = caption
	}
	if buf.timer != nil {
		buf.timer.Stop()
	}
	buf.timer = time.AfterFunc(mediaGroupCollectDelay, func() {
		buf.flush(key)
	})
	return nil
}

func (buf *mediaGroupBuffer) flush(key string) {
	buf.mu.Lock()
	parts := append([]string(nil), buf.parts...)
	caption := buf.caption
	bot := buf.bot
	buf.mu.Unlock()

	mediaGroups.mu.Lock()
	delete(mediaGroups.buffers, key)
	mediaGroups.mu.Unlock()

	var lines []string
	if caption != "" {
		lines = append(lines, caption)
	}
	lines = append(lines, parts...)
	content := strings.TrimSpace(strings.Join(lines, "\n"))
	if content == "" {
		return
	}
	_ = bot.queueIncomingContent(content)
}
