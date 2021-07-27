package mcgo

import (
	"fmt"
	"os"
	"testing"
)

func TestClaimNamemc(t *testing.T) {
	email := os.Getenv("MJ_EMAIL")
	pass := os.Getenv("MJ_PASS")
	acc := MCaccount{
		Email:    email,
		Password: pass,
	}
	err := acc.MojangAuthenticate()
	fmt.Println(err)
	acc.ClaimNamemc()
}
