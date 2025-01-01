package token

import (
	"fmt"

	"github.com/thanhpk/randstr"
)

func GetOrCreateCommandToken(commandName string) (string, error) {
	sysID, err := getIDFromSystem()
	if err != nil {
		return "", fmt.Errorf("failed to load an ID from the system: %w", err)
	}

	comToken := commandName + " " + sysID

	return comToken, nil
}

func getIDFromSystem() (string, error) {
	return "", nil
}

func generateNewID() string {
	tokenLength := 10
	id := randstr.Base62(tokenLength)
	return id
}
