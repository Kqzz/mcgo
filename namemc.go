package mcgo

import (
	"fmt"
	"log"

	"github.com/Tnze/go-mc/bot"
	"github.com/Tnze/go-mc/bot/basic"
	"github.com/Tnze/go-mc/bot/screen"
	"github.com/Tnze/go-mc/chat"
	"github.com/google/uuid"
)

var client *bot.Client
var player *basic.Player
var screenManager *screen.Manager

func (account *MCaccount) ClaimNamemc() {
	client = bot.NewClient()

	client.Auth.AsTk = account.Bearer
	client.Auth.Name = account.Username
	client.Auth.UUID = account.UUID

	player = basic.NewPlayer(client, basic.Settings{
		ChatMode:   1,
		ChatColors: false,
	})

	basic.EventsListener{
		GameStart: func() error {
			fmt.Println("Hey")
			return nil
		},
		ChatMsg: chatMsgHandlerd,
		Disconnect: func(chat.Message) error {
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

	err := client.JoinServer("blockmania.com")
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
