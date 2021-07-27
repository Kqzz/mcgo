package mcgo

import (
	"os"
	"testing"
)

func TestMsa(t *testing.T) {
	email := os.Getenv("MS_EMAIL")
	pass := os.Getenv("MS_PASSWORD")
	acc := MCaccount{
		Email:    email,
		Password: pass,
	}
	if err := acc.MicrosoftAuthenticate(); err != nil || acc.Bearer == "" {
		t.Fatal(err)
	}
}
