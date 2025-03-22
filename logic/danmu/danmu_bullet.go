package danmu

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/xbclub/BilibiliDanmuRobot-Core/entity"
	"github.com/xbclub/BilibiliDanmuRobot-Core/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"regexp"
	"strconv"
	"strings"
)

var danmuHandler *DanmuLogic

type DanmuLogic struct {
	danmuChan chan string
}

func PushToBDanmuLogic(bullet string) {
	danmuHandler.danmuChan <- bullet
}

func StartDanmuLogic(ctx context.Context, svcCtx *svc.ServiceContext) {
	var err error

	danmuHandler = &DanmuLogic{
		danmuChan: make(chan string, 1000),
	}

	var msg string
	for {
		select {
		case <-ctx.Done():
			goto END
		case msg = <-danmuHandler.danmuChan:
			danmu := &entity.DanmuMsgText{}
			err = json.Unmarshal([]byte(msg), danmu)
			if err != nil {
				logx.Error(err)
			}

			danmumsg := danmu.Info[1].(string)
			from := danmu.Info[2].([]interface{})
			uid := fmt.Sprintf("%.0f", from[0].(float64))
			re := regexp.MustCompile("\\[(.*?)\\]")
			danmumsg = re.ReplaceAllString(danmumsg, "")

			// 解析Info[0]
			extras := danmu.Info[0].([]interface{})
			extra := extras[15].(map[string]interface{})
			jsonExtra := strings.ReplaceAll(extra["extra"].(string), `\"`, `"`)
			tagExtra := &entity.DanmuMsgTextInfo0Extra{}
			e := json.Unmarshal([]byte(jsonExtra), tagExtra)
			if e != nil {
				logx.Alert(e.Error())
				// } else {
				// 	logx.Info("idStr: ", tagExtra.IdStr, " ", uid)
			}

			reply := &entity.DanmuMsgTextReplyInfo{
				ReplyUid:   uid,
				ReplyMsgId: tagExtra.IdStr,
			}

			cardLv := "0"
			card := "无信仰"
			if len(danmu.Info) > 3 {
				cardInfo := danmu.Info[3].([]interface{})
				if len(cardInfo) > 1 {
					cardLv = fmt.Sprintf("%.0f", cardInfo[0].(float64))
					card = cardInfo[1].(string)
				}
			}
			if len(danmumsg) > 0 && uid != svcCtx.RobotID {
				// 机器人相关
				go DoDanmuProcess(danmumsg, svcCtx, reply)
				// 弹幕统计
				if svcCtx.Config.DanmuCntEnable {
					go BadgeActiveCheckProcess(danmumsg, uid, from[1].(string), svcCtx, reply)
				}
				// 关键词回复
				if svcCtx.Config.KeywordReply {
					go KeywordReply(danmumsg, svcCtx, reply)
				}
			}
			// 签到
			if svcCtx.Config.SignInEnable {
				go DosignInProcess(danmumsg, uid, from[1].(string), svcCtx, reply)
			}
			// 抽签
			if svcCtx.Config.DrawByLot {
				go DodrawByLotProcess(danmumsg, from[1].(string), svcCtx, reply)
			}
			// 修改调用方式
			if svcCtx.Config.BlindBoxStat {
				go DoBlindBoxStat(danmumsg, uid, from[1].(string), svcCtx, reply)
				go DoBlindBoxStatByType(danmumsg, uid, from[1].(string), svcCtx, reply)
			}
			if len(danmumsg) > 0 && uid == strconv.FormatInt(svcCtx.UserID, 10) {
				// 主播指令控制
				go DoCMDProcess(danmumsg, uid, svcCtx)
			}
			// 实时输出弹幕消息
			if tagExtra.ReplyMid > 0 {
				danmumsg = fmt.Sprintf("@%s %s", tagExtra.ReplyUname, danmumsg)
			}
			logx.Infof("%v 「%s %s」%s:%s", uid, cardLv, card, from[1], danmumsg)
		}

	}
END:
}
