package mcgo

import (
	"os"
	"testing"
)

func TestMsa(t *testing.T) {
	email := os.Getenv("EMAIL")
	pass := os.Getenv("PASSWORD")
	acc := MCaccount{
		Email:    email,
		Password: pass,
	}
	if err := acc.MicrosoftAuthenticate(); err != nil || acc.Bearer == "" {
		t.Fatal(err)
	}
}
