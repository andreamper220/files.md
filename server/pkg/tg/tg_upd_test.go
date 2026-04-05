package tg

import (
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/require"
)

func TestCmdNil(t *testing.T) {
	r := require.New(t)

	m := tgbotapi.Message{}
	m.Text = "j new journal record"
	rawUpdate := tgbotapi.Update{
		UpdateID: 0,
		Message:  &m,
	}

	u := NewTGUpd(rawUpdate)
	cmd := u.Cmd()

	r.Nil(cmd)
}

func TestCmdInTheBeginning(t *testing.T) {
	r := require.New(t)

	m := tgbotapi.Message{}
	m.Text = "/j new journal record"
	m.Entities = []tgbotapi.MessageEntity{{
		Type:   "bot_command",
		Offset: 0,
		Length: 2,
	}}
	rawUpdate := tgbotapi.Update{
		UpdateID: 0,
		Message:  &m,
	}

	u := NewTGUpd(rawUpdate)
	cmd := u.Cmd()

	r.NotNil(cmd)
	r.Equal("j", cmd.Name)
	r.Equal("New journal record", cmd.Params[0])
}

func TestCmdAtTheEnd(t *testing.T) {
	r := require.New(t)

	m := tgbotapi.Message{}
	m.Text = "new journal record /j"
	m.Entities = []tgbotapi.MessageEntity{{
		Type:   "bot_command",
		Offset: 19,
		Length: 2,
	}}
	rawUpdate := tgbotapi.Update{
		UpdateID: 0,
		Message:  &m,
	}

	u := NewTGUpd(rawUpdate)
	cmd := u.Cmd()

	r.NotNil(cmd)
	r.Equal("j", cmd.Name)
	r.Equal("New journal record", cmd.Params[0])
}

func TestUserID(t *testing.T) {
	r := require.New(t)

	// Test case with a regular message
	m := tgbotapi.Message{Chat: &tgbotapi.Chat{ID: 12345}}
	rawUpdate := tgbotapi.Update{Message: &m}
	u := NewTGUpd(rawUpdate)
	userID := u.UserID()
	r.Equal(int64(12345), userID)

	// Test case with a callback query
	cbq := tgbotapi.CallbackQuery{Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: 67890}}}
	rawUpdate = tgbotapi.Update{CallbackQuery: &cbq}
	u = NewTGUpd(rawUpdate)
	userID = u.UserID()
	r.Equal(int64(67890), userID)

	// Test case with an inline query
	iq := tgbotapi.InlineQuery{From: &tgbotapi.User{ID: 112233}}
	rawUpdate = tgbotapi.Update{InlineQuery: &iq}
	u = NewTGUpd(rawUpdate)
	userID = u.UserID()
	r.Equal(int64(112233), userID)
}

func TestMsgText(t *testing.T) {
	r := require.New(t)

	// Test with a regular message
	m := tgbotapi.Message{Text: "Hello, world!"}
	rawUpdate := tgbotapi.Update{Message: &m}
	u := NewTGUpd(rawUpdate)
	r.Equal("Hello, world!", u.MsgText())

	// Test with no message
	rawUpdate = tgbotapi.Update{}
	u = NewTGUpd(rawUpdate)
	r.Equal("", u.MsgText())
}

func TestIsForwarded(t *testing.T) {
	r := require.New(t)

	// Test case where the message is forwarded
	m := tgbotapi.Message{ForwardFromMessageID: 1}
	rawUpdate := tgbotapi.Update{Message: &m}
	u := NewTGUpd(rawUpdate)
	r.True(u.IsForwarded())

	// Test case where the message is not forwarded
	m = tgbotapi.Message{}
	rawUpdate = tgbotapi.Update{Message: &m}
	u = NewTGUpd(rawUpdate)
	r.False(u.IsForwarded())
}

func TestIsSentViaBot(t *testing.T) {
	r := require.New(t)

	// Test case where the message is sent via a bot
	m := tgbotapi.Message{ViaBot: &tgbotapi.User{ID: 1}}
	rawUpdate := tgbotapi.Update{Message: &m}
	u := NewTGUpd(rawUpdate)
	r.True(u.IsSentViaBot())

	// Test case where the message is not sent via a bot
	m = tgbotapi.Message{}
	rawUpdate = tgbotapi.Update{Message: &m}
	u = NewTGUpd(rawUpdate)
	r.False(u.IsSentViaBot())
}

func TestReplyToMsgID(t *testing.T) {
	r := require.New(t)

	// Test case where the message is a reply
	m := tgbotapi.Message{ReplyToMessage: &tgbotapi.Message{MessageID: 101}}
	rawUpdate := tgbotapi.Update{Message: &m}
	u := NewTGUpd(rawUpdate)
	replyID, found := u.ReplyToMsgID()
	r.True(found)
	r.Equal(101, replyID)

	// Test case where the message is not a reply
	m = tgbotapi.Message{}
	rawUpdate = tgbotapi.Update{Message: &m}
	u = NewTGUpd(rawUpdate)
	replyID, found = u.ReplyToMsgID()
	r.False(found)
	r.Equal(0, replyID)
}

func TestPhotoOrImageID(t *testing.T) {
	r := require.New(t)

	// Test case with a photo
	m := tgbotapi.Message{Photo: []tgbotapi.PhotoSize{{FileID: "photo_id_123", FileSize: 200}}}
	rawUpdate := tgbotapi.Update{Message: &m}
	u := NewTGUpd(rawUpdate)
	id, found := u.PhotoOrImageID()
	r.True(found)
	r.Equal("photo_id_123", id)

	// Test case with an image
	m = tgbotapi.Message{Document: &tgbotapi.Document{FileID: "image_id_456", MimeType: "image/png"}}
	rawUpdate = tgbotapi.Update{Message: &m}
	u = NewTGUpd(rawUpdate)
	id, found = u.PhotoOrImageID()
	r.True(found)
	r.Equal("image_id_456", id)

	// Test case with no photo or image
	m = tgbotapi.Message{}
	rawUpdate = tgbotapi.Update{Message: &m}
	u = NewTGUpd(rawUpdate)
	id, found = u.PhotoOrImageID()
	r.False(found)
	r.Equal("", id)
}
