package alarmclock

import (
	"sync"
	"time"

	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
)

var (
	instance = &ac{}
	// logger   = utils.GetModuleLogger("MiraiGo-Q")
	// exitchan = make(chan bool)
	// ctx, cancel = context.WithCancel(context.Background())
	// timeto = make(chan time.Time)
)

type ac struct{}

func init() {
	bot.RegisterModule(instance)
}

func (a *ac) MiraiGoModule() bot.ModuleInfo {
	return bot.ModuleInfo{
		ID:       "internal/alarmclock",
		Instance: instance,
	}
}

func (a *ac) Init() {
}

func (a *ac) PostInit() {
	// qq = bot.Instance
}

func (a *ac) Serve(bot *bot.Bot) {
	// 群消息
	go func() {
		for {
			// case <-ctx.Done():
			// return
			<-time.NewTicker(1 * time.Minute).C
			msg := message.NewSendingMessage().Append(message.NewText(time.Now().Format("2006-01-02 15:04:05")))
			bot.SendGroupMessage(808468274, msg)
		}
	}()
	bot.OnGroupMessage(func(q *client.QQClient, gm *message.GroupMessage) {
		q.SendGroupMessage(gm.GroupCode, message.NewSendingMessage().Append(message.NewText(gm.ToString())))
	})

}

func (a *ac) Start(bot *bot.Bot) {

}

func (a *ac) Stop(bot *bot.Bot, wg *sync.WaitGroup) {
	defer wg.Done()
	// exitchan <- true
	// cancel()
}
