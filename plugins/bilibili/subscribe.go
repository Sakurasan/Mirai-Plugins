package bilibili

import (
	"Mirai-Plugins/pkg/config"
	"Mirai-Plugins/pkg/miraitool"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sakurasan/to"
	// "golang.org/x/exp/rand"
)

var (
	instance = new()
	// logger      = utils.GetModuleLogger("MiraiGo-Q")
	ctx, cancel = context.WithCancel(context.Background())
	// Sub_map      = make(map[string]Vlist)
	// subUrlformat = "https://api.bilibili.com/x/space/arc/search?mid=697166795&pn=1&ps=5&order=pubdate&index=1"
	subUrlformat = "https://api.bilibili.com/x/space/arc/search?mid=%s&ps=5&tid=0&pn=1&keyword=&order=pubdate"
	tpl          = `
  bilibili:
    name: "bilibili/sub"
    channel:
      - 7584632 #频道号
    sub:
	  1234567: #群号
	    - 7584632 #频道号
      
`
	ua = map[int]string{
		0: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.6.1 Safari/605.1.15",
		1: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:99.0) Gecko/20100101 Firefox/99.0",
		2: "bilibili",
		3: "PostmanRuntime/7.29.0",
		// 4: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36",
	}
)

func suburl(mid string) string {
	return fmt.Sprintf(subUrlformat, mid)
}

type SubBiLi struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
	Data    struct {
		List struct {
			Tlist interface{} `json:"tlist,omitempty"`
			Vlist []Vlist     `json:"vlist,omitempty"`
		} `json:"list,omitempty"`
		Page struct {
			Pn    int `json:"pn,omitempty"`
			Ps    int `json:"ps,omitempty"`
			Count int `json:"count,omitempty"`
		} `json:"page,omitempty"`
		EpisodicButton struct {
			Text string `json:"text,omitempty"`
			URI  string `json:"uri,omitempty"`
		} `json:"episodic_button,omitempty"`
	} `json:"data"`
}

type Vlist struct {
	Comment        int    `json:"comment,omitempty"`
	Typeid         int    `json:"typeid,omitempty"`
	Play           int    `json:"play,omitempty"`
	Pic            string `json:"pic,omitempty"`
	Subtitle       string `json:"subtitle,omitempty"`
	Description    string `json:"description,omitempty"`
	Copyright      string `json:"copyright,omitempty"`
	Title          string `json:"title,omitempty"`
	Review         int    `json:"review,omitempty"`
	Author         string `json:"author,omitempty"`
	Mid            int    `json:"mid,omitempty"`
	Created        int64  `json:"created,omitempty"`
	Length         string `json:"length,omitempty"`
	VideoReview    int    `json:"video_review,omitempty"`
	Aid            int    `json:"aid,omitempty"`
	Bvid           string `json:"bvid,omitempty"`
	HideClick      bool   `json:"hide_click,omitempty"`
	IsPay          int    `json:"is_pay,omitempty"`
	IsUnionVideo   int    `json:"is_union_video,omitempty"`
	IsSteinsGate   int    `json:"is_steins_gate,omitempty"`
	IsLivePlayback int    `json:"is_live_playback,omitempty"`
}

type bili struct {
	sub_bili map[string]int
	sub_map  map[string]Vlist
	sync.Mutex
	Jar http.CookieJar
}

func new() *bili {
	var b bili
	b.sub_bili = make(map[string]int, 10)
	b.sub_map = make(map[string]Vlist)
	return &b
}

func init() {
	bot.RegisterModule(instance)
}

func (a *bili) MiraiGoModule() bot.ModuleInfo {
	return bot.ModuleInfo{
		ID:       "internal/bilibili",
		Instance: instance,
	}
}
func (a *bili) getCookie() error {
	if a.Jar != nil {
		return nil
	}
	client := http.DefaultClient
	client.Timeout = 10 * time.Second
	req, _ := http.NewRequest(http.MethodGet, "https://www.bilibili.com", nil)
	seed := time.Now().Unix()
	req.Header.Set("User-Agent", ua[rand.New(rand.NewSource(seed)).Intn(len(ua))])
	// req.Header.Set("User-Agent", ua[0])
	// req.Header.Set("User-Agent", "PostmanRuntime/7.29.0")
	_, err := client.Do(req)
	if err != nil {
		return err
	}
	a.Jar = client.Jar
	return nil
}

func (a *bili) getBiliSub(mid string) (*Vlist, error) {
	if err := a.getCookie(); err != nil {
		return nil, err
	}
	client := http.DefaultClient
	client.Jar = a.Jar
	client.Timeout = 10 * time.Second
	req, _ := http.NewRequest(http.MethodGet, suburl(mid), nil)
	seed := time.Now().Unix()
	randnum := rand.New(rand.NewSource(seed)).Intn(len(ua))
	req.Header.Set("User-Agent", ua[randnum])
	rsp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	bytersp, _ := ioutil.ReadAll(rsp.Body)

	var sb SubBiLi

	if err := json.NewDecoder(bytes.NewReader(bytersp)).Decode(&sb); err != nil {
		return nil, err
	}
	if len(sb.Data.List.Vlist) <= 1 {
		log.Println("getbili fail:", ua[randnum])
		return nil, errors.New(fmt.Sprintf("bilibili vlist[%d] lehgth  less than 5", len(sb.Data.List.Vlist)))
	}
	if _, ok := a.sub_map[mid]; ok {
		localpubtime := time.Unix(a.sub_map[mid].Created, 0)
		for i := len(sb.Data.List.Vlist) - 1; i >= 0; i-- {
			pubtime := time.Unix(sb.Data.List.Vlist[i].Created, 0)
			if pubtime.After(localpubtime) {
				return &sb.Data.List.Vlist[i], nil
			}
		}
	}

	// debug
	// for _, v := range sb.Data.List.Vlist {
	// 	fmt.Println("debug vlist:", v)
	// }

	return &sb.Data.List.Vlist[0], nil
}

func (a *bili) subscribeBiliChannel() {
	for _, b_channel := range config.PluginConfig.GetStringSlice("plugins.bilibili.channel") {
		vlist, err := a.getBiliSub(b_channel)
		if err != nil {
			continue
		}
		a.Lock()
		a.sub_map[b_channel] = *vlist
		a.Unlock()
	}
}

func (a *bili) Init() {
	if !config.PluginConfig.IsSet("plugins.bilibili.channel") {
		f, _ := os.OpenFile("./plugins.yaml", os.O_CREATE|os.O_RDWR|os.O_APPEND, os.ModePerm)
		defer f.Close()
		f.WriteString(tpl)
	}
}

func (a *bili) PostInit() {}

func (a *bili) Serve(bot *bot.Bot) {
	ticker := time.NewTicker(3 * time.Second)
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				defer ticker.Stop()
				return
			case <-ticker.C:
				ticker.Reset(60 * time.Second)
				for _, v := range config.PluginConfig.GetStringSlice("plugins.bilibili.channel") {
					vlist, err := a.getBiliSub(v)
					if err != nil {
						fmt.Println("Query_Bili channel:", err)
						continue
					}
					a.Lock()
					a.sub_map[v] = *vlist
					a.Unlock()
				}
				// case <-time.NewTicker(60 * time.Second).C:
				for groupCode, vlist := range config.PluginConfig.GetStringMapStringSlice("plugins.bilibili.sub") {
					for _, video_channnel := range vlist {
						if _, ok := a.sub_bili[video_channnel]; !ok {
							if a.sub_map[video_channnel].Aid != 0 {
								a.sub_bili[video_channnel] = a.sub_map[video_channnel].Aid
							}
							continue
						} else {
							if a.sub_bili[video_channnel] != a.sub_map[video_channnel].Aid {
								t := time.Unix(a.sub_map[video_channnel].Created, 0)
								img, err := miraitool.UpGroupImgByUrl(bot.QQClient, to.Int64(groupCode), a.sub_map[video_channnel].Pic)
								if err != nil {
									img, _ = miraitool.UpGroupImgByUrl(bot.QQClient, to.Int64(groupCode), a.sub_map[video_channnel].Pic)
								}
							retry:
								var n = 5
								ret := bot.SendGroupMessage(to.Int64(groupCode), message.NewSendingMessage().Append(message.NewText(a.sub_map[video_channnel].Author+" "+video_channnel+"\n")).Append(img).Append(message.NewText("\n"+a.sub_map[video_channnel].Title+"\n"+a.sub_map[video_channnel].Description+"\n")).Append(message.NewText(fmt.Sprintf("https://www.bilibili.com/av%d \n --- \n 时长:%s \n %ss ago", a.sub_map[video_channnel].Aid, a.sub_map[video_channnel].Length, strings.Split(time.Now().Sub(t).String(), ".")[0]))))
								if ret == nil || ret.Id == -1 {
									if n >= 0 {
										goto retry
									}
									log.Printf("retry %d ,bilibili:%d | %d\n", n, a.sub_bili[video_channnel], a.sub_map[video_channnel].Aid)
									log.Printf("%#v \n %#v", a.sub_bili, a.sub_map)
								} else {
									a.sub_bili[video_channnel] = a.sub_map[video_channnel].Aid
								}

							}
						}
					}
				}

			}
		}
	}(ctx)

	bot.OnGroupMessage(func(c *client.QQClient, msg *message.GroupMessage) {
		s := msg.ToString()
		for _, m := range msg.Elements {
			switch m.(type) {
			case *message.TextElement:
				if strings.HasPrefix(s, "subili") {
					// log.Printf("sub: %#v \n%#v", a.sub_bili, a.sub_map)
					var s1, s2 string
					for k, v := range a.sub_bili {
						s1 += fmt.Sprintf("%s-%d \n", k, v)
					}
					for k, v := range a.sub_map {
						s2 += fmt.Sprintf("%s-%d ,%s \n", k, v.Aid, v.Title)
					}
				retry:
					var i = 5
					ret := c.SendGroupMessage(msg.GroupCode, message.NewSendingMessage().Append(message.NewText(s1+"\n---\n"+s2)))
					if i >= 0 && ret == nil || ret.Id == -1 {
						i--
						fmt.Println("retry:", 5-i)
						goto retry
					}
				}
				if strings.HasPrefix(s, "pubili") {
					if blist := strings.Split(s, " "); len(blist) == 2 {
						a.sub_bili[blist[1]] = 233
					}
				}
			}
		}

	})
}

func (a *bili) Start(bot *bot.Bot) {

}

func (a *bili) Stop(bot *bot.Bot, wg *sync.WaitGroup) {
	defer wg.Done()
	cancel()
	time.Sleep(100 * time.Microsecond)
}
