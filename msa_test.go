package mcgo

import (
	"fmt"
	"testing"
)

func TestMsa(t *testing.T) {
	acc := MCaccount{
		Email:    "replace with valid email",
		Password: "replace with valid password",
	}
	if err := acc.MicrosoftAuthenticate(); err != nil {
		t.Fatal(err)
	}
	fmt.Println(acc.Bearer)
}
