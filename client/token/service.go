package token

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"

	"github.com/thanhpk/randstr"
)

const dataFolderPath = "/usr/local/telebot"
const dataFileName = "system_id"

func GetOrCreateCommandToken(commandName string) (string, error) {
	sysID, err := getOrCreateSystemID()
	if err != nil {
		return "", fmt.Errorf("failed to load an ID from the system: %w", err)
	}

	comToken := commandName + " " + sysID

	return comToken, nil
}

func getOrCreateSystemID() (string, error) {
	folderExists, err := dataFolderExists()
	if err != nil {
		return "", err
	}

	if !folderExists {
		err := os.MkdirAll(dataFolderPath, os.FileMode(os.ModePerm))
		if err != nil {
			return "", err
		}
	}

	dataFilePath := path.Join(dataFolderPath, dataFileName)
	file, err := os.OpenFile(dataFilePath, os.O_CREATE, os.FileMode(os.ModePerm))
	if err != nil {
		return "", err
	}

	defer file.Close()

	var fileData []byte
	_, err = file.Read(fileData)
	if err != nil {
		return "", err
	}

	systemID := string(fileData)
	if systemID == "" {
		systemID = generateNewID()
		_, err = file.WriteString(systemID)
		if err != nil {
			return "", err
		}
	}

	return systemID, nil
}

func generateNewID() string {
	tokenLength := 10
	id := randstr.Base62(tokenLength)
	return id
}

func dataFolderExists() (bool, error) {
	_, err := os.Stat(dataFolderPath)
	if err == nil {
		return true, nil
	}

	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}

	return false, err
}
