package mcgo

import (
	"strings"
	"time"

	"github.com/Tnze/go-mc/bot"
	"github.com/Tnze/go-mc/bot/basic"
	"github.com/Tnze/go-mc/bot/screen"
	"github.com/Tnze/go-mc/chat"
	pk "github.com/Tnze/go-mc/net/packet"
	"github.com/Tnze/go-mc/yggdrasil"
	"github.com/google/uuid"
)

var client *bot.Client
var player *basic.Player
var screenManager *screen.Manager

func (account *MCaccount) ClaimNamemc() (string, error) {
	client = bot.NewClient()

	resp, err := yggdrasil.Authenticate(account.Email, account.Password)
	if err != nil {
		return "", err
	}

	id, name := resp.SelectedProfile()
	client.Auth.Name = name
	client.Auth.UUID = id
	client.Auth.AsTk = resp.AccessToken()

	player = basic.NewPlayer(client, basic.Settings{
		ChatMode:   1,
		ChatColors: false,
	})

	claimUrlChan := make(chan string)

	basic.EventsListener{
		GameStart: func() error {
			go func() {
				// sleep and send /namemc cmd
				time.Sleep(time.Millisecond * 500)
				err = client.Conn.WritePacket(pk.Marshal(
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

	err = client.JoinServer("blockmania.com")
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
