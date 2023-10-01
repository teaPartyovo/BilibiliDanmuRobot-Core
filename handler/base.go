package handler

import (
	"context"
	"github.com/Akegarasu/blivedm-go/client"
	_ "github.com/Akegarasu/blivedm-go/utils"
	"github.com/robfig/cron/v3"
	"github.com/xbclub/BilibiliDanmuRobot-Core/entity"
	"github.com/xbclub/BilibiliDanmuRobot-Core/http"
	"github.com/xbclub/BilibiliDanmuRobot-Core/logic"
	"github.com/xbclub/BilibiliDanmuRobot-Core/svc"
	"github.com/xbclub/BilibiliDanmuRobot-Core/utiles"
	"github.com/zeromicro/go-zero/core/logx"
	"math/rand"
	"os"
)

type wsHandler struct {
	client *client.Client
	svc    *svc.ServiceContext
	// 机器人
	robotBulletCtx    context.Context
	robotBulletCancel context.CancelFunc
	// 弹幕发送
	sendBulletCtx    context.Context
	sendBulletCancel context.CancelFunc
	// 特效欢迎
	ineterractCtx    context.Context
	ineterractCancel context.CancelFunc
	//礼物感谢
	thanksGiftCtx   context.Context
	thankGiftCancel context.CancelFunc
	//pk提醒
	pkCtx    context.Context
	pkCancel context.CancelFunc
	//定时弹幕
	corndanmu           *cron.Cron
	mapCronDanmuSendIdx map[int]int
}

func NewWsHandler(svc *svc.ServiceContext) WsHandler {
	ws := new(wsHandler)
	ws.starthttp()
	ws.client = client.NewClient(svc.Config.RoomId)
	ws.client.SetCookie(http.CookieStr)
	ws.svc = svc
	//初始化定时弹幕
	ws.corndanmu = cron.New(cron.WithParser(cron.NewParser(
		cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow,
	)))
	ws.mapCronDanmuSendIdx = make(map[int]int)
	return ws
}

type WsHandler interface {
	StartWsClient() error
	StopWsClient()
	starthttp()
}

func (w *wsHandler) StartWsClient() error {
	w.startLogic()
	if w.svc.Config.EntryMsg != "off" {
		err := http.Send(w.svc.Config.EntryMsg, w.svc)
		if err != nil {
			logx.Error(err)
		}
	}
	return w.client.Start()
}
func (w *wsHandler) StopWsClient() {
	if w.sendBulletCancel != nil {
		w.sendBulletCancel()
	}
	if w.robotBulletCancel != nil {
		w.robotBulletCancel()
	}
	if w.thankGiftCancel != nil {
		w.thankGiftCancel()
	}
	if w.ineterractCancel != nil {
		w.ineterractCancel() // 关闭弹幕姬goroutine
	}
	if w.pkCancel != nil {
		w.pkCancel()
	}
	for _, i := range w.corndanmu.Entries() {
		w.corndanmu.Remove(i.ID)
	}
	w.corndanmu.Stop()
	w.client.Stop()
}
func (w *wsHandler) startLogic() {
	w.sendBulletCtx, w.sendBulletCancel = context.WithCancel(context.Background())
	go logic.StartSendBullet(w.sendBulletCtx, w.svc)
	logx.Info("弹幕推送已开启...")
	// 机器人
	w.robotBulletCtx, w.robotBulletCancel = context.WithCancel(context.Background())
	go logic.StartBulletRobot(w.robotBulletCtx, w.svc)
	w.robot()
	logx.Info("弹幕机器人已开启")
	// 特效欢迎
	w.ineterractCtx, w.ineterractCancel = context.WithCancel(context.Background())
	go logic.Interact(w.ineterractCtx)
	w.welcomeEntryEffect()
	w.welcomeInteractWord()
	logx.Info("欢迎模块已开启")
	// 礼物感谢
	w.thanksGiftCtx, w.thankGiftCancel = context.WithCancel(context.Background())
	go logic.ThanksGift(w.thanksGiftCtx, w.svc)
	w.thankGifts()
	logx.Info("礼物感谢已开启")
	// pk提醒
	w.pkCtx, w.pkCancel = context.WithCancel(context.Background())
	go logic.PK(w.pkCtx, w.svc)
	w.pkBattleStart()
	w.pkBattleEnd()
	logx.Info("pk提醒已开启")
	//定时弹幕
	w.corndanmuStart()

}
func (w *wsHandler) starthttp() {
	var err error
	http.InitHttpClient()
	// 判断是否存在历史cookie
	if http.FileExists("token/bili_token.txt") && http.FileExists("token/bili_token.json") {
		err = http.SetHistoryCookie()
		if err != nil {
			logx.Error("用户登录失败")
			os.Exit(1)
		}
		logx.Info("用户登录成功")
	} else {
		//if err = w.userlogin(); err != nil {
		//	logx.Errorf("用户登录失败：%v", err)
		//	return
		//}
		//logx.Info("用户登录成功")
		logx.Error("用户登录失败")
		os.Exit(1)
	}
}
func (w *wsHandler) userlogin() error {
	var err error
	http.InitHttpClient()
	var loginUrl *entity.LoginUrl
	if loginUrl, err = http.GetLoginUrl(); err != nil {
		logx.Error(err)
		return err
	}

	if err = utiles.GenerateQr(loginUrl.Data.Url); err != nil {
		logx.Error(err)
		return err
	}

	if _, err = http.GetLoginInfo(loginUrl.Data.OauthKey); err != nil {
		logx.Error(err)
		return err
	}

	return err
}
func (w *wsHandler) corndanmuStart() {
	if w.svc.Config.CronDanmu == false {
		return
	}
	for n, danmu := range w.svc.Config.CronDanmuList {
		if danmu.Danmu != nil {
			i := n
			danmus := danmu
			_, err := w.corndanmu.AddFunc(danmus.Cron, func() {
				if len(danmus.Danmu) > 0 {
					if danmus.Random {
						logic.PushToBulletSender(danmus.Danmu[rand.Intn(len(danmus.Danmu))])
					} else {
						_, ok := w.mapCronDanmuSendIdx[i]
						if !ok {
							w.mapCronDanmuSendIdx[i] = 0
						}
						w.mapCronDanmuSendIdx[i] = w.mapCronDanmuSendIdx[i] + 1
						logic.PushToBulletSender(danmus.Danmu[w.mapCronDanmuSendIdx[i]%len(danmus.Danmu)])
					}
				}
			})
			if err != nil {
				logx.Errorf("第%d条定时弹幕配置出现错误: %v", i+1, err)
			}
		}
	}
	w.corndanmu.Start()
}
