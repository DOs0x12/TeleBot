package bot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	botEnt "github.com/DOs0x12/TeleBot/server/v3/internal/entities/bot"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type FileDto struct {
	Name string
	Data []byte
}

func (t telebot) receiveInData(
	ctx context.Context,
	updChan tgbot.UpdatesChannel,
	botInDataChan chan<- botEnt.Data,
	botErrChan chan<- error,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case upd := <-updChan:
			err := t.processBotData(botInDataChan, upd)
			if err != nil {
				botErrChan <- err
			}
		}
	}
}

func (t telebot) processBotData(botInDataChan chan<- botEnt.Data, upd tgbot.Update) error {
	if upd.Message == nil {
		return errors.New("got an empty message from the bot")
	}

	isComm := upd.Message.Command() != ""

	var mesVal []byte
	isFile := false
	if upd.Message.Document != nil {
		var err error
		fileData, err := getFileData(t, upd.Message.Document.FileID)
		if err != nil {
			return fmt.Errorf("failed to get a file data from the bot API: %w", err)
		}

		fileDto := FileDto{Name: upd.Message.Document.FileName, Data: fileData}
		rawMesVal, err := json.Marshal(fileDto)
		if err != nil {
			return fmt.Errorf("failed to marshal a file DTO: %w", err)
		}

		mesVal = rawMesVal
		isFile = true
	} else {
		mesVal = []byte(upd.Message.Text)
	}

	botInDataChan <- botEnt.Data{
		ChatID:    upd.Message.Chat.ID,
		Value:     mesVal,
		IsCommand: isComm,
		IsFile:    isFile,
	}

	return nil
}

func getFileData(t telebot, fileID string) ([]byte, error) {
	fileUrl, err := t.bot.GetFileDirectURL(fileID)
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(fileUrl)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	fileData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return fileData, nil
}
