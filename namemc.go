package mcgo

import (
	"fmt"
	"log"
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

func (account *MCaccount) ClaimNamemc() {
	client = bot.NewClient()

	resp, err := yggdrasil.Authenticate(account.Email, account.Password)
	if err != nil {
		log.Fatal(err)
	}

	id, name := resp.SelectedProfile()
	client.Auth.Name = name
	client.Auth.UUID = id
	client.Auth.AsTk = resp.AccessToken()

	player = basic.NewPlayer(client, basic.Settings{
		ChatMode:   1,
		ChatColors: false,
	})

	basic.EventsListener{
		GameStart: func() error {
			go func() {
				time.Sleep(time.Second * 2)
				err = client.Conn.WritePacket(pk.Marshal(
					0x03,
					pk.String(chat.Message{Text: "/namemc"}.String()),
					pk.Byte(0),
				))
				fmt.Println(err)
			}()
			return nil
		},
		ChatMsg: chatMsgHandlerd,
		Disconnect: func(m chat.Message) error {

			fmt.Println(m.ClearString())
			fmt.Println("disconnecting...")
			return nil
		},
		HealthChange: func(float32) error {
			fmt.Println("rip i took damage")
			return nil
		},
		Death: func() error {
			fmt.Println(
				"BRO WHO KILLED ME WTH",
			)
			return player.Respawn()
		},
	}.Attach(client)

	err = client.JoinServer("blockmania.com")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Login success")

	//JoinGame
	if err = client.HandleGame(); err == nil {
		panic("HandleGame never return nil")
	}

}

func chatMsgHandlerd(c chat.Message, _ byte, _ uuid.UUID) error {
	fmt.Println("Chat:", c.ClearString()) // output chat message without any format code (like color or bold)
	return nil
}

type DisconnectErr struct {
	Reason chat.Message
}
