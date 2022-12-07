package main

import (
	"os"
	"os/signal"

	pconfig "Mirai-Plugins/pkg/config"
	_ "Mirai-Plugins/plugins/bilibili"
	_ "Mirai-Plugins/plugins/ping"

	"github.com/Logiase/MiraiGo-Template/bot"
	_ "github.com/Logiase/MiraiGo-Template/modules/logging"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/duke-git/lancet/v2/fileutil"
)

func init() {
	utils.WriteLogToFS(utils.LogInfoLevel, utils.LogWithStack)
	pconfig.GlobalConfigInit(pconfig.WithConfigName("qBot.yaml"))
	pconfig.Init()
	if !fileutil.IsExist("./device.json") {
		bot.GenRandomDevice()
	}

}

func main() {
	// 快速初始化
	bot.Init()

	// 初始化 Modules
	bot.StartService()

	// 使用协议
	// 不同协议可能会有部分功能无法使用
	// 在登陆前切换协议
	bot.UseProtocol(bot.AndroidPhone)

	// 登录
	bot.Login()

	// 刷新好友列表，群列表
	bot.RefreshList()
	bot.SaveToken()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
	bot.Stop()
}
