package bot

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	botEnt "github.com/DOs0x12/TeleBot/server/v2/internal/entities/bot"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

type FileDto struct {
	Name,
	Data string
}

func (t telebot) receiveInData(
	ctx context.Context,
	updChan tgbot.UpdatesChannel,
	botInDataChan chan<- botEnt.Data,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case upd := <-updChan:
			t.processBotData(botInDataChan, upd)
		}
	}
}

func (t telebot) processBotData(botInDataChan chan<- botEnt.Data, upd tgbot.Update) {
	if upd.Message == nil {
		return
	}

	isComm := upd.Message.Command() != ""

	var mesVal string
	if upd.Message.Document != nil {
		var err error
		fileData, err := getFileData(t, upd.Message.Document.FileID)
		if err != nil {
			logrus.Errorf("failed to get a file data from the bot API: %v", err)

			return
		}

		fileDto := FileDto{Name: upd.Message.Document.FileName, Data: fileData}
		rawMesVal, err := json.Marshal(fileDto)
		if err != nil {
			logrus.Errorf("failed to marshal a file DTO: %v", err)
		}

		mesVal = string(rawMesVal)
	} else {
		mesVal = upd.Message.Text
	}

	botInDataChan <- botEnt.Data{
		ChatID:    upd.Message.Chat.ID,
		Value:     mesVal,
		IsCommand: isComm,
	}
}

func getFileData(t telebot, fileID string) (string, error) {
	fileUrl, err := t.bot.GetFileDirectURL(fileID)
	if err != nil {
		return "", err
	}

	resp, err := http.Get(fileUrl)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	fileData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(fileData), nil
}
