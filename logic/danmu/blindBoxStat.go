package danmu

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

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
		return
	}

	now := carbon.Now(carbon.Local)
	var year, month, day int
	var err error

	reg := `^(今日|\d{1,2}月(?:\d{1,2}日)?|\d{4}年\d{1,2}月(?:\d{1,2}日)?)盲盒$`
	re := regexp.MustCompile(reg)
	match := re.FindStringSubmatch(msg)

	if len(match) == 0 {
		return
	}

	dateStr := match[1]

	if dateStr == "今日" {
		year = now.Year()
		month = now.Month()
		day = now.Day()
	} else {
		// 解析年月日
		parts := strings.Split(strings.TrimSuffix(dateStr, "日"), "月")
		if strings.Contains(parts[0], "年") {
			// 包含年份
			yearMonth := strings.Split(parts[0], "年")
			year, err = strconv.Atoi(yearMonth[0])
			if err != nil || year < 2000 || year > 9999 {
				logic.PushToBulletSender(fmt.Sprintf("年份「%s」不正确!", yearMonth[0]), reply...)
				return
			}
			month, err = strconv.Atoi(yearMonth[1])
		} else {
			// 不包含年份
			year = now.Year()
			month, err = strconv.Atoi(parts[0])
		}

		if err != nil || month < 1 || month > 12 {
			logic.PushToBulletSender(fmt.Sprintf("月份「%d」不正确!", month), reply...)
			return
		}

		if len(parts) > 1 && parts[1] != "" {
			day, err = strconv.Atoi(parts[1])
			if err != nil || day < 1 || day > 31 {
				logic.PushToBulletSender(fmt.Sprintf("日期「%s」不正确!", parts[1]), reply...)
				return
			}
		}
	}

	id, err := strconv.ParseInt(uid, 10, 64)
	if err != nil {
		logx.Errorf("UID解析错误: %v", err)
		logic.PushToBulletSender(errInfo, reply...)
		return
	}

	logx.Infof("开始查询盲盒数据 - 年份: %d, 月份: %d, 日期: %d, UID: %s", year, month, day, uid)

	var ret *model.Result
	if svcCtx.UserID == id {
		logx.Info("查询主播全部数据")
		ret, err = svcCtx.BlindBoxStatModel.GetTotal(context.Background(), int16(year), int16(month), int16(day))
	} else {
		logx.Info("查询用户个人数据")
		ret, err = svcCtx.BlindBoxStatModel.GetTotalOnePersion(context.Background(), id, int16(year), int16(month), int16(day))
	}

	// 删除重复的查询代码，直接使用上面的查询结果
	if err == nil {
		r := float64(ret.R) / float64(1000.0)
		var dateStr string
		if msg == "今日盲盒" {
			dateStr = fmt.Sprintf("%d年%d月%d日", year, month, day)
		} else if day > 0 {
			dateStr = fmt.Sprintf("%d年%d月%d日", year, month, day)
		} else {
			dateStr = fmt.Sprintf("%d年%d月", year, month)
		}

		if ret.R > 0 {
			logic.PushToBulletSender(
				fmt.Sprintf(
					"%s共开%d个, 赚了＋%.2f元",
					dateStr,
					ret.C,
					r,
				),
				reply...,
			)
		} else if ret.R == 0 {
			logic.PushToBulletSender(
				fmt.Sprintf(
					"%s共开%d个, 没亏没赚!",
					dateStr,
					ret.C,
				),
				reply...,
			)
		} else {
			logic.PushToBulletSender(
				fmt.Sprintf(
					"%s共开%d个, 亏了－%.2f元",
					dateStr,
					ret.C,
					math.Abs(r),
				),
				reply...,
			)
		}
	} else {
		logic.PushToBulletSender(errInfo, reply...)
		logx.Alert("盲盒统计出错了!" + err.Error())
	}
}

func DoBlindBoxStatByType(msg, uid, username string, svcCtx *svc.ServiceContext, reply ...*entity.DanmuMsgTextReplyInfo) {
	if !svcCtx.Config.BlindBoxStat {
		return
	}

	now := carbon.Now(carbon.Local)
	var year, month, day int
	var err error
	var boxType string

	reg := `^(今日|\d{1,2}月(?:\d{1,2}日)?|\d{4}年\d{1,2}月(?:\d{1,2}日)?)(.+?)盲盒$`
	re := regexp.MustCompile(reg)
	match := re.FindStringSubmatch(msg)

	if len(match) == 0 {
		return
	}

	dateStr := match[1]
	boxType = match[2]

	if dateStr == "今日" {
		year = now.Year()
		month = now.Month()
		day = now.Day()
	} else {
		// 解析年月日
		parts := strings.Split(strings.TrimSuffix(dateStr, "日"), "月")
		if strings.Contains(parts[0], "年") {
			// 包含年份
			yearMonth := strings.Split(parts[0], "年")
			year, err = strconv.Atoi(yearMonth[0])
			if err != nil || year < 2000 || year > 9999 {
				logic.PushToBulletSender(fmt.Sprintf("年份「%s」不正确!", yearMonth[0]), reply...)
				return
			}
			month, err = strconv.Atoi(yearMonth[1])
		} else {
			// 不包含年份
			year = now.Year()
			month, err = strconv.Atoi(parts[0])
		}

		if err != nil || month < 1 || month > 12 {
			logic.PushToBulletSender(fmt.Sprintf("月份「%d」不正确!", month), reply...)
			return
		}

		if len(parts) > 1 && parts[1] != "" {
			day, err = strconv.Atoi(parts[1])
			if err != nil || day < 1 || day > 31 {
				logic.PushToBulletSender(fmt.Sprintf("日期「%s」不正确!", parts[1]), reply...)
				return
			}
		}
	}

	id, err := strconv.ParseInt(uid, 10, 64)
	if err != nil {
		logx.Error(err)
		logic.PushToBulletSender(errInfo, reply...)
		return
	}

	now = carbon.Now(carbon.Local)
	var ret *model.Result

	// 主播查询本月所有数据
	if svcCtx.UserID == id {
		ret, err = svcCtx.BlindBoxStatModel.GetTotalByType(context.Background(), boxType, int16(year), int16(month), int16(day))
	} else {
		ret, err = svcCtx.BlindBoxStatModel.GetTotalOnePersonByType(context.Background(), id, boxType, int16(year), int16(month), int16(day))
	}

	if err == nil {
		r := float64(ret.R) / float64(1000.0)
		var dateStr string
		if strings.HasPrefix(msg, "今日") {
			dateStr = fmt.Sprintf("%d年%d月%d日", year, month, day)
		} else if day > 0 {
			dateStr = fmt.Sprintf("%d年%d月%d日", year, month, day)
		} else {
			dateStr = fmt.Sprintf("%d年%d月", year, month)
		}

		if ret.R > 0 {
			logic.PushToBulletSender(
				fmt.Sprintf(
					"%s%s盲盒共开%d个, 赚了＋%.2f元",
					dateStr,
					boxType,
					ret.C,
					r,
				),
				reply...,
			)
		} else if ret.R == 0 {
			logic.PushToBulletSender(
				fmt.Sprintf(
					"%s%s盲盒共开%d个, 没亏没赚!",
					dateStr,
					boxType,
					ret.C,
				),
				reply...,
			)
		} else {
			logic.PushToBulletSender(
				fmt.Sprintf(
					"%s%s盲盒共开%d个, 亏了－%.2f元",
					dateStr,
					boxType,
					ret.C,
					math.Abs(r),
				),
				reply...,
			)
		}
	} else {
		logic.PushToBulletSender(errInfo, reply...)
		logx.Alert("盲盒统计出错了!" + err.Error())
	}
}
