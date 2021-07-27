package mcgo

import (
	"strings"
	"time"

	"github.com/Tnze/go-mc/bot"
	"github.com/Tnze/go-mc/bot/basic"
	"github.com/Tnze/go-mc/chat"
	pk "github.com/Tnze/go-mc/net/packet"
	"github.com/google/uuid"
)

var client *bot.Client

func (account *MCaccount) ClaimNamemc() (string, error) {
	client = bot.NewClient()

	client.Auth.Name = account.Username
	client.Auth.UUID = account.UUID
	client.Auth.AsTk = account.Bearer

	claimUrlChan := make(chan string)

	basic.EventsListener{
		GameStart: func() error {
			go func() {
				// sleep and send /namemc cmd
				time.Sleep(time.Millisecond * 500)
				client.Conn.WritePacket(pk.Marshal(
					0x03,
					pk.String("/namemc"),
				))
			}()
			return nil
		},
		ChatMsg: func(c chat.Message, pos byte, uuid uuid.UUID) error {
			cStr := c.ClearString()
			if strings.Contains(cStr, "https://namemc.com/claim?key=") {
				go func() {
					claimUrlChan <- cStr
				}()
			}
			return nil
		},
	}.Attach(client)

	err := client.JoinServer("blockmania.com")
	if err != nil {
		return "", err
	}

	go func() error {
		//JoinGame
		err = client.HandleGame()
		if err == nil {
			return err
		}
		return nil
	}()

	claimUrl := <-claimUrlChan
	return claimUrl, nil

}
