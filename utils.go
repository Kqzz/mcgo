package mcgo

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

type Droptime struct {
	Droptime time.Time
	Username string
}

// https://stackoverflow.com/a/68240817/13312615
func SameErrorMessage(err, target error) bool {
	if target == nil || err == nil {
		return err == target
	}
	return err.Error() == target.Error()
}

func NameAvailability(username string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.mojang.com/user/profile/agent/minecraft/name/%v", username))

	if err != nil {
		return "", err
	}

	if resp.StatusCode == 200 {
		return "claimed", nil
	}

	if resp.StatusCode == 400 {
		return "invalid", nil
	}

	if resp.StatusCode == 429 {
		return "", errors.New("mojang API ratelimit reached")
	}

	return "", fmt.Errorf("this should not be possible! | Got status %v on request for name availability", resp.StatusCode)
}
