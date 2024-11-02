package tg

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	MarkupMarkdown = "MarkdownV2"
	MarkupHTML     = "Html"
)

// FakeTG is a simple wrapper over Telegram API.
// It can send/edit messages with Keyboard attached.
type TG struct {
	api *tgbotapi.BotAPI
}

func NewTG(api *tgbotapi.BotAPI) *TG {
	return &TG{api}
}

func (tg *TG) SendHTML(userID int64, text string, kb *Keyboard) (int, error) {
	return tg.Send(userID, text, kb, MarkupHTML)
}

func (tg *TG) Send(userID int64, text string, kb *Keyboard, markup string) (int, error) {
	msg := tgbotapi.NewMessage(userID, text)
	msg.ReplyMarkup = tg.buildInlineKeyboard(kb)
	msg.ParseMode = markup

	resp, err := tg.api.Send(msg)
	if err != nil {
		js, _ := json.Marshal(msg)
		return 0, fmt.Errorf("tg send: can't send json %s: %w", js, err)
	}

	return resp.MessageID, nil
}

func (tg *TG) SendImages(userID int64, text string, images []string) (int, error) {
	var files []interface{}
	for _, img := range images {
		files = append(files, tgbotapi.NewInputMediaPhoto(tgbotapi.FileID(img)))
	}

	msg := tgbotapi.NewMediaGroup(userID, files)

	resp, err := tg.api.Send(msg)
	if err != nil {
		js, _ := json.Marshal(msg)
		return 0, fmt.Errorf("tg send images: can't send json %s: %w", js, err)
	}

	return resp.MessageID, nil
}

func (tg *TG) Edit(userID int64, msgID int, text string, kb *Keyboard, markup string) error {
	msg := tgbotapi.NewEditMessageText(userID, msgID, text)

	msg.ReplyMarkup = tg.buildInlineKeyboard(kb)
	msg.ParseMode = markup

	_, err := tg.api.Send(msg)
	if err != nil {
		return fmt.Errorf("tg edit: %w", err)
	}

	return nil
}

func (tg *TG) Del(userID int64, msgID int) error {
	del := tgbotapi.NewDeleteMessage(userID, msgID)

	_, err := tg.api.Send(del)
	if err != nil {
		return fmt.Errorf("tg del: %w", err)
	}

	return nil
}

// AnswerCallbackQuery answers to incoming callbacks (keyboard's button clicks)
func (tg *TG) AnswerCallbackQuery(queryID string, text string) error {
	_, err := tg.api.Send(tgbotapi.CallbackConfig{CallbackQueryID: queryID, Text: text})
	if err != nil {
		return fmt.Errorf("tg can't answer to callback query: %w", err)
	}

	return nil
}

// AnswerInlineQuery answers to incoming inline queries (@BotMention <text>)
func (tg *TG) AnswerInlineQuery(queryID string, results []interface{}, cacheTime int, offset string) error {
	config := tgbotapi.InlineConfig{
		InlineQueryID: queryID,
		Results:       results,
		CacheTime:     cacheTime,
		NextOffset:    offset,
		IsPersonal:    true,
	}
	_, err := tg.api.Send(config)
	if err != nil {
		return fmt.Errorf("tg can't answer to inline query: %w", err)
	}

	return nil
}

// DownloadFile returns downloaded file extension
func (tg *TG) DownloadFile(fileID string, outFile io.Writer) (string, error) {
	file, err := tg.api.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		return "", fmt.Errorf("tg can't get file: %w", err)
	}

	fileURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", tg.api.Token, file.FilePath)
	resp, err := http.Get(fileURL)
	if err != nil {
		return "", fmt.Errorf("tg can't download file: %w", err)
	}
	defer resp.Body.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("tg can't save file: %w", err)
	}

	return path.Ext(file.FilePath), nil
}

func (tg *TG) buildInlineKeyboard(kb *Keyboard) *tgbotapi.InlineKeyboardMarkup {
	if kb == nil || len(kb.Btns) == 0 {
		return nil
	}

	inlineKb := tgbotapi.NewInlineKeyboardMarkup()
	for _, row := range kb.Btns {
		switch typedRow := row.(type) {
		case Btn:
			btn := tg.buildBtn(typedRow)
			inlineKb.InlineKeyboard = append(inlineKb.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(btn))
		case []Btn:
			inlineRow := tgbotapi.NewInlineKeyboardRow()
			for _, b := range typedRow {
				inlineRow = append(inlineRow, tg.buildBtn(b))
			}
			inlineKb.InlineKeyboard = append(inlineKb.InlineKeyboard, inlineRow)
		}
	}

	return &inlineKb
}

func (tg *TG) buildBtn(btn Btn) tgbotapi.InlineKeyboardButton {
	serializedCmd, _ := json.Marshal(btn.Cmd)
	button := tgbotapi.InlineKeyboardButton{
		Text: btn.Name,
	}
	switch btn.Cmd.Type {

	case CmdTypeInlineQueryCurrentChat:
		{
			param := ""
			if len(btn.Cmd.Params) > 0 {
				param = btn.Cmd.Params[0]
			}
			button.SwitchInlineQueryCurrentChat = &param
		}
	case CmdTypeWebApp:
		{
			param := ""
			if len(btn.Cmd.Params) > 0 {
				param = btn.Cmd.Params[0]
			}
			button.WebApp = &tgbotapi.WebAppInfo{URL: param}
		}
	default:
		{
			str := string(serializedCmd)
			button.CallbackData = &str
		}
	}

	return button
}
