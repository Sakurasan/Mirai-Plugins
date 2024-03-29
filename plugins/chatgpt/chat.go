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
	"github.com/Sakurasan/to"
	openai "github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"

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
		//优先处理@
		b.OnGroupMessage(func(c *client.QQClient, msg *message.GroupMessage) {
			if !chatgpt.hasdb {
				logrus.WithField("Plugins", "chatgpt").Warningln("数据库配置失败,该功能暂停使用")
				return
			}
			var isat bool
			var ele []message.IMessageElement
			for _, elem := range msg.Elements {
				switch e := elem.(type) {
				case *message.AtElement:
					if e.Target == c.Uin { //被@
						if !isat {
							isat = true
							ele = append(ele, elem)
						}
					}
				case *message.TextElement:
					e.Content = strings.TrimSpace(e.Content)
					if len(e.Content) == 0 {
						continue
					}
					if isat {
						ele = append(ele, elem)
						var (
							answer string
							err    error
						)
						storemsg, _ := chatgpt.GetMessage(to.String(msg.Sender.Uin))
						if storemsg == nil {
							answer, err = chatgpt.ChatWithMessage([]openai.ChatCompletionMessage{{Role: "user", Content: e.Content}}, WithUser(to.String(msg.Sender.Uin)))
						} else if len(storemsg) == 10 {
							storemsg = storemsg[1:]
							storemsg = append(storemsg, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: e.Content})
							answer, err = chatgpt.ChatWithMessage(storemsg, WithUser(to.String(msg.Sender.Uin)))
						} else if len(storemsg) >= 0 {
							storemsg = append(storemsg, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: e.Content})
							answer, err = chatgpt.ChatWithMessage(storemsg, WithUser(to.String(msg.Sender.Uin)))
						}
						// answer, err := chatgpt.Chat(msg.ToString())
						if err != nil {
							sm := message.NewSendingMessage().Append(message.NewText(err.Error()))
							c.SendGroupMessage(msg.GroupCode, sm)
							return
						}
						answer = strings.TrimPrefix(answer, "\n\n")
						sm := message.NewSendingMessage().Append(message.NewReply(&message.GroupMessage{Id: msg.Id, Sender: msg.Sender, Time: msg.Time, Elements: ele})).Append(message.NewText(answer))
						c.SendGroupMessage(msg.GroupCode, sm)
					}
				}
			}
		})

		b.OnGroupMessage(func(c *client.QQClient, msg *message.GroupMessage) {
			log.Println(msg.ToString())
			var str string
			for _, elem := range msg.Elements {
				if e, ok := elem.(*message.TextElement); ok {
					str += e.Content
				}
				if _, ok := elem.(*message.AtElement); ok {
					return
				}
			}
			str = strings.TrimSpace(str)
			answer, err := chatgpt.Chat(str)
			// answer, err := chatgpt.ChatWithMessage([]openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: msg.ToString()}}, WithUser(to.String(msg.Sender.Uin)))
			if err != nil {
				m := message.NewSendingMessage().Append(message.NewText(err.Error()))
				c.SendGroupMessage(msg.GroupCode, m)
				return
			}
			m := message.NewSendingMessage().Append(message.NewText(strings.TrimPrefix(answer, "\n")))
			sm := c.SendGroupMessage(msg.GroupCode, m)
			if sm == nil || sm.Id == -1 {
				logrus.WithField("chatgpt", "OnGroupMessage").Error("发送消息失败")
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
