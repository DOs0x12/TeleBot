package system

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/thanhpk/randstr"
)

const dataFolderPath = "/var/lib/telebot"
const dataFileName = "system_id"

var systemID string

// The method gets the topic name with a unique string for each system. The topic contains messages for a server.
func GetDataToken() (string, error) {
	dataPrefix := "telebot-data"

	err := generateSystemIDIfNotExist()
	if err != nil {
		return "", fmt.Errorf("failed to generate a system ID for the data token: %w", err)
	}

	return getToken(dataPrefix), nil
}

// The method gets the topic name with a unique string for each system. The topic contains messages for a service.
func GetServiceToken(serviceName string) (string, error) {
	err := generateSystemIDIfNotExist()
	if err != nil {
		return "", fmt.Errorf("failed to generate a system ID for the service token: %w", err)
	}

	return getToken(serviceName), nil
}

func getToken(prefix string) string {
	token := prefix + "-" + systemID

	return token
}

func generateSystemIDIfNotExist() error {
	if systemID != "" {
		return nil
	}

	folderExists, err := dataFolderExists()
	if err != nil {
		return err
	}

	fPermissions := os.FileMode.Perm(0644)

	if !folderExists {
		err := os.MkdirAll(dataFolderPath, fPermissions)
		if err != nil {
			return err
		}
	}

	dataFilePath := path.Join(dataFolderPath, dataFileName)
	file, err := os.OpenFile(dataFilePath, os.O_CREATE|os.O_RDWR, fPermissions)
	if err != nil {
		return err
	}

	defer file.Close()

	fileData, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	systemID = strings.Trim(string(fileData), "\n")
	if systemID == "" {
		systemID = generateNewID()
		_, err = file.WriteString(systemID)
		if err != nil {
			return err
		}
	}

	return nil
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
