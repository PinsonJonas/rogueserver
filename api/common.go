package api

import (
	"encoding/base64"
	"fmt"

	"github.com/pagefaultgames/pokerogue-server/api/account"
	"github.com/pagefaultgames/pokerogue-server/api/daily"
	"github.com/pagefaultgames/pokerogue-server/db"
)

func Init() {
	scheduleStatRefresh()
	daily.Init()
}

func usernameFromTokenHeader(token string) (string, error) {
	if token == "" {
		return "", fmt.Errorf("missing token")
	}

	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return "", fmt.Errorf("failed to decode token: %s", err)
	}

	if len(decoded) != account.TokenSize {
		return "", fmt.Errorf("invalid token length: got %d, expected %d", len(token), account.TokenSize)
	}

	username, err := db.FetchUsernameFromToken(decoded)
	if err != nil {
		return "", fmt.Errorf("failed to validate token: %s", err)
	}

	return username, nil
}

func uuidFromTokenHeader(token string) ([]byte, error) {
	if token == "" {
		return nil, fmt.Errorf("missing token")
	}

	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, fmt.Errorf("failed to decode token: %s", err)
	}

	if len(decoded) != account.TokenSize {
		return nil, fmt.Errorf("invalid token length: got %d, expected %d", len(token), account.TokenSize)
	}

	uuid, err := db.FetchUUIDFromToken(decoded)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %s", err)
	}

	return uuid, nil
}
