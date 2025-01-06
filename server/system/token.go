package system

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"

	"github.com/thanhpk/randstr"
)

const dataFolderPath = "/var/lib/telebot"
const dataFileName = "system_id"

// The method gets or creates a token containing the topic name and a unique string for each system.
// The unique topic name is used to restrict the data to only one system in which the application works.
// The system ID will be saved to a file which is stored in the system.
func GetOrCreateTopicToken(topicName string) (string, error) {
	sysID, err := getOrCreateSystemID()
	if err != nil {
		return "", fmt.Errorf("failed to load an ID from the system: %w", err)
	}

	token := topicName + "-" + sysID

	return token, nil
}

func getOrCreateSystemID() (string, error) {
	folderExists, err := dataFolderExists()
	if err != nil {
		return "", err
	}

	if !folderExists {
		err := os.MkdirAll(dataFolderPath, os.FileMode(os.O_RDWR))
		if err != nil {
			return "", err
		}
	}

	dataFilePath := path.Join(dataFolderPath, dataFileName)
	file, err := os.OpenFile(dataFilePath, os.O_CREATE|os.O_RDWR, os.FileMode(os.O_RDWR))
	if err != nil {
		return "", err
	}

	defer file.Close()

	fileData, err := io.ReadAll(file)
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
