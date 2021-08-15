package mcgo

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestPrename(t *testing.T) {
	bearer := os.Getenv("BEARER")
	if bearer == "" {
		bearer = "TestToken"
	}
	acc := MCaccount{Bearer: bearer}

	nameChangeRet, err := acc.ChangeName("test", time.Now().Add(time.Second*1), true)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(nameChangeRet.ChangedName)
	fmt.Println(nameChangeRet.SendTime)
	fmt.Println(nameChangeRet.ReceiveTime)
	fmt.Println(nameChangeRet.StatusCode)
}
