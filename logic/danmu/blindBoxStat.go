package danmu

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strconv"

	_ "github.com/glebarez/go-sqlite"
	"github.com/golang-module/carbon/v2"
	"github.com/xbclub/BilibiliDanmuRobot-Core/entity"
	"github.com/xbclub/BilibiliDanmuRobot-Core/logic"
	"github.com/xbclub/BilibiliDanmuRobot-Core/model"
	"github.com/xbclub/BilibiliDanmuRobot-Core/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

var errInfo string = "盲盒统计服务异常"

func SaveBlindBoxStat(g *entity.SendGiftText, svcCtx *svc.ServiceContext) {
	logx.Info(g.Data.BlindGift.OriginalGiftName)
	if g.Data.BlindGift.OriginalGiftName == "" {
		return
	}
	now := carbon.Now(carbon.Local)
	err := svcCtx.BlindBoxStatModel.Insert(context.Background(), nil, &model.BlindBoxStatBase{
		Uid:               int64(g.Data.UID),
		BlindBoxName:      g.Data.BlindGift.OriginalGiftName,
		Price:             int32(g.Data.Price),
		OriginalGiftPrice: int32(g.Data.BlindGift.OriginalGiftPrice),
		Cnt:               int32(g.Data.Num),
		Year:              int16(now.Year()),
		Month:             int16(now.Month()),
		Day:               int16(now.Day()),
	})
	if err != nil {
		logx.Alert("保存盲盒数据出错!!! " + err.Error())
	} else {
		logx.Info("盲盒数据保存成功!!! ")
	}
}

func DoBlindBoxStat(msg, uid, username string, svcCtx *svc.ServiceContext, reply ...*entity.DanmuMsgTextReplyInfo) {
	if !svcCtx.Config.BlindBoxStat {
		logx.Debug("盲盒统计功能未启用")
		return
	}

	logx.Infof("收到盲盒统计请求 - 用户: %s, UID: %s, 消息: %s", username, uid, msg)

	reg := `^(?:([0-9]{4})年)?([0-9]+)月盲盒$`
	re := regexp.MustCompile(reg)
	match := re.FindStringSubmatch(msg)

	if len(match) != 3 {
		logx.Debugf("消息格式不匹配: %s", msg)
		return
	}

	now := carbon.Now(carbon.Local)
	var year int
	var err error

	if match[1] != "" {
		year, err = strconv.Atoi(match[1])
		if err != nil || year < 2000 || year > 9999 {
			logx.Errorf("年份解析错误: %s", match[1])
			logic.PushToBulletSender(fmt.Sprintf("年份「%s」不正确!", match[1]), reply...)
			return
		}
		logx.Infof("使用指定年份: %d", year)
	} else {
		year = now.Year()
		logx.Infof("使用当前年份: %d", year)
	}

	month, err := strconv.Atoi(match[2])
	if err != nil || month < 1 || month > 12 {
		logx.Errorf("月份解析错误: %s", match[2])
		logic.PushToBulletSender(fmt.Sprintf("月份「%s」不正确!", match[2]), reply...)
		return
	}

	id, err := strconv.ParseInt(uid, 10, 64)
	if err != nil {
		logx.Errorf("UID解析错误: %v", err)
		logic.PushToBulletSender(errInfo, reply...)
		return
	}

	logx.Infof("开始查询盲盒数据 - 年份: %d, 月份: %d, UID: %s", year, month, uid)

	var ret *model.Result
	if svcCtx.UserID == id {
		logx.Info("查询主播全部数据")
		ret, err = svcCtx.BlindBoxStatModel.GetTotal(context.Background(), int16(year), int16(month), 0)
	} else {
		logx.Info("查询用户个人数据")
		ret, err = svcCtx.BlindBoxStatModel.GetTotalOnePersion(context.Background(), id, int16(year), int16(month), 0)
	}

	if err != nil {
		logx.Errorf("查询盲盒数据出错: %v", err)
		logic.PushToBulletSender(errInfo, reply...)
		return
	}

	logx.Infof("查询结果 - 数量: %d, 盈亏: %.2f", ret.C, float64(ret.R)/1000.0)

	r := float64(ret.R) / float64(1000.0)
	if ret.R > 0 {
		logic.PushToBulletSender(
			fmt.Sprintf("%d年%s月共开%d个, 赚了＋%.2f元", year, match[2], ret.C, r),
			reply...,
		)
	} else if ret.R == 0 {
		logic.PushToBulletSender(
			fmt.Sprintf("%d年%s月共开%d个, 没亏没赚!", year, match[2], ret.C),
			reply...,
		)
	} else {
		logic.PushToBulletSender(
			fmt.Sprintf("%d年%s月共开%d个, 亏了－%.2f元", year, match[2], ret.C, math.Abs(r)),
			reply...,
		)
	}
}

func DoBlindBoxStatByType(msg, uid, username string, svcCtx *svc.ServiceContext, reply ...*entity.DanmuMsgTextReplyInfo) {
	if !svcCtx.Config.BlindBoxStat {
		logx.Debug("盲盒统计功能未启用")
		return
	}

	logx.Infof("收到指定类型盲盒统计请求 - 用户: %s, UID: %s, 消息: %s", username, uid, msg)

	reg := `^(?:([0-9]{4})年)?([0-9]+)月([^盲]+)盲盒$`
	re := regexp.MustCompile(reg)
	match := re.FindStringSubmatch(msg)

	if len(match) != 4 {
		logx.Debugf("消息格式不匹配: %s", msg)
		return
	}

	now := carbon.Now(carbon.Local)
	var year int
	var err error

	if match[1] != "" {
		year, err = strconv.Atoi(match[1])
		if err != nil || year < 2000 || year > 9999 {
			logx.Errorf("年份解析错误: %s", match[1])
			logic.PushToBulletSender(fmt.Sprintf("年份「%s」不正确!", match[1]), reply...)
			return
		}
		logx.Infof("使用指定年份: %d", year)
	} else {
		year = now.Year()
		logx.Infof("使用当前年份: %d", year)
	}

	month, err := strconv.Atoi(match[2])
	if err != nil || month < 1 || month > 12 {
		logx.Errorf("月份解析错误: %s", match[2])
		logic.PushToBulletSender(fmt.Sprintf("月份「%s」不正确!", match[2]), reply...)
		return
	}

	id, err := strconv.ParseInt(uid, 10, 64)
	if err != nil {
		logx.Errorf("UID解析错误: %v", err)
		logic.PushToBulletSender(errInfo, reply...)
		return
	}

	logx.Infof("开始查询指定类型盲盒数据 - 年份: %d, 月份: %d, 类型: %s, UID: %s", year, month, match[3], uid)

	var ret *model.Result
	if svcCtx.UserID == id {
		logx.Info("查询主播指定类型数据")
		ret, err = svcCtx.BlindBoxStatModel.GetTotalByType(context.Background(), match[3], int16(year), int16(month), 0)
	} else {
		logx.Info("查询用户个人指定类型数据")
		ret, err = svcCtx.BlindBoxStatModel.GetTotalOnePersonByType(context.Background(), id, match[3], int16(year), int16(month), 0)
	}

	if err != nil {
		logx.Errorf("查询盲盒数据出错: %v", err)
		logic.PushToBulletSender(errInfo, reply...)
		return
	}

	logx.Infof("查询结果 - 类型: %s, 数量: %d, 盈亏: %.2f", match[3], ret.C, float64(ret.R)/1000.0)

	r := float64(ret.R) / float64(1000.0)
	if ret.R > 0 {
		logic.PushToBulletSender(
			fmt.Sprintf("%d年%s月%s盲盒共开%d个, 赚了＋%.2f元", year, match[2], match[3], ret.C, r),
			reply...,
		)
	} else if ret.R == 0 {
		logic.PushToBulletSender(
			fmt.Sprintf("%d年%s月%s盲盒共开%d个, 没亏没赚!", year, match[2], match[3], ret.C),
			reply...,
		)
	} else {
		logic.PushToBulletSender(
			fmt.Sprintf("%d年%s月%s盲盒共开%d个, 亏了－%.2f元", year, match[2], match[3], ret.C, math.Abs(r)),
			reply...,
		)
	}
}
