package ping

import (
	"fmt"
	"sync"

	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"

	"github.com/Logiase/MiraiGo-Template/bot"
)

func init() {
	bot.RegisterModule(instance)
}

var instance = &A{}

// var logger = utils.GetModuleLogger("logiase.autoreply")

// var tem map[string]string

type A struct {
}

func (a *A) MiraiGoModule() bot.ModuleInfo {
	return bot.ModuleInfo{
		ID:       "MiraiQ.autoreply",
		Instance: instance,
	}
}

func (a *A) Init() {

}

func (a *A) PostInit() {
}

func (a *A) Serve(b *bot.Bot) {
	if b != nil {
		b.OnGroupMessage(func(c *client.QQClient, msg *message.GroupMessage) {
			if msg.ToString() == "ping" {
				m := message.NewSendingMessage().Append(message.NewText("pong"))
				c.SendGroupMessage(msg.GroupCode, m)
			}
		})

		b.OnPrivateMessage(func(c *client.QQClient, msg *message.PrivateMessage) {
			if msg.ToString() == "ping" {
				m := message.NewSendingMessage().Append(message.NewText("pong"))
				c.SendPrivateMessage(msg.Sender.Uin, m)
			}

		})
	} else {
		fmt.Println("bot nil")
	}

}

func (a *A) Start(bot *bot.Bot) {
}

func (a *A) Stop(bot *bot.Bot, wg *sync.WaitGroup) {
	defer wg.Done()
}
