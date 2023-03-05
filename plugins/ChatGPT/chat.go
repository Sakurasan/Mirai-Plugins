package chatgpt

import (
	"Mirai-Plugins/pkg/config"
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"

	"github.com/Logiase/MiraiGo-Template/bot"
)

func init() {
	bot.RegisterModule(instance)
}

var (
	ctx, cancel = context.WithCancel(context.Background())
	chatgpt     *ChatGPT
)

var instance = &A{}

// var logger = utils.GetModuleLogger("logiase.autoreply")

// var tem map[string]string

type A struct {
}

func (a *A) MiraiGoModule() bot.ModuleInfo {
	return bot.ModuleInfo{
		ID:       "Mirai-Plugins.ChatGPT",
		Instance: instance,
	}
}

func (a *A) Init() {
	authToken := config.PluginConfig.GetString("plugins.chatgpt.authToken")

	proxyUrl := config.PluginConfig.GetString("plugins.chatgpt.proxyUrl")
	if len(proxyUrl) > 2 {
		p, err := url.Parse(proxyUrl)
		if err != nil {
			panic(err)
		}
		tr := &http.Transport{
			Proxy:           http.ProxyURL(p),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		httpc := http.DefaultClient
		httpc.Transport = tr

		chatgpt = New(ctx, authToken, WithHttpClient(httpc))
	} else {
		chatgpt = New(ctx, authToken)
	}
}

func (a *A) PostInit() {
}

func (a *A) Serve(b *bot.Bot) {
	if b != nil {
		b.OnGroupMessage(func(c *client.QQClient, msg *message.GroupMessage) {
			answer, err := chatgpt.Chat(msg.ToString())
			if err != nil {
				m := message.NewSendingMessage().Append(message.NewText(err.Error()))
				c.SendGroupMessage(msg.GroupCode, m)
				return
			}
			m := message.NewSendingMessage().Append(message.NewText(strings.TrimPrefix(answer, "\n")))
			sm := c.SendGroupMessage(msg.GroupCode, m)
			if sm.Id != 0 {
				log.Println("发送消息失败")
			}
		})
		// fixAt := func(elem []message.IMessageElement) {
		// 	for _, e := range elem {
		// 		if at, ok := e.(*message.AtElement); ok && at.Target != 0 && at.Display == "" {
		// 			mem := group.FindMember(at.Target)
		// 			if mem != nil {
		// 				at.Display = "@" + mem.DisplayName()
		// 			} else {
		// 				at.Display = "@" + strconv.FormatInt(at.Target, 10)
		// 			}
		// 		}
		// 	}
		// }
		b.OnGroupMessage(func(q *client.QQClient, gm *message.GroupMessage) {
			for _, e := range gm.Elements {
				switch elem := e.(type) {
				case *message.AtElement:
					// group := q.FindGroup(gm.GroupCode)
					// mem := group.FindMember(elem.Target)
					if elem.Target == q.Uin {
						sm := message.NewSendingMessage().Append(message.NewReply(gm)).Append(message.NewText("在！"))
						q.SendGroupMessage(gm.GroupCode, sm)
					}
				}
			}
		})

		b.OnPrivateMessage(func(c *client.QQClient, msg *message.PrivateMessage) {
			answer, err := chatgpt.Chat(msg.ToString())
			if err != nil {
				m := message.NewSendingMessage().Append(message.NewText(err.Error()))
				c.SendPrivateMessage(msg.Sender.Uin, m)
			}
			m := message.NewSendingMessage().Append(message.NewText(strings.TrimPrefix(answer, "\n\n")))
			c.SendPrivateMessage(msg.Sender.Uin, m)
		})
	} else {
		fmt.Println("bot nil")
	}

}

func (a *A) Start(bot *bot.Bot) {
}

func (a *A) Stop(bot *bot.Bot, wg *sync.WaitGroup) {
	defer wg.Done()
	cancel()
}
