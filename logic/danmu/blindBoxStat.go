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

	// 修改正则表达式以更好地支持可选的年月日
	reg := `^(?:今日|(?:(?P<year>\d{4})年)?(?P<month>\d{1,2})月(?:(?P<day>\d{1,2})日)?)盲盒$`
	re := regexp.MustCompile(reg)
	match := re.FindStringSubmatch(msg)

	if len(match) == 0 {
		return
	}

	// 获取当前时间
	now := carbon.Now(carbon.Local)
	var year, month, day int
	var err error

	if msg == "今日盲盒" {
		year = now.Year()
		month = now.Month()
		day = now.Day()
	} else {
		// 解析年份（可选）
		if match[1] != "" {
			year, err = strconv.Atoi(match[1])
			if err != nil || year < 2000 || year > 9999 {
				logic.PushToBulletSender(fmt.Sprintf("年份「%s」不正确!", match[1]), reply...)
				return
			}
		} else {
			year = now.Year()
		}

		// 解析月份（必需）
		month, err = strconv.Atoi(match[2])
		if err != nil || month < 1 || month > 12 {
			logic.PushToBulletSender(fmt.Sprintf("月份「%s」不正确!", match[2]), reply...)
			return
		}

		// 解析日期（可选）
		if match[3] != "" {
			day, err = strconv.Atoi(match[3])
			if err != nil || day < 1 || day > 31 {
				logic.PushToBulletSender(fmt.Sprintf("日期「%s」不正确!", match[3]), reply...)
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

	if err != nil {
		logx.Errorf("查询盲盒数据出错: %v", err)
		logic.PushToBulletSender(errInfo, reply...)
		return
	}

	logx.Infof("查询结果 - 数量: %d, 盈亏: %.2f", ret.C, float64(ret.R)/1000.0)

	// 使用解析出的年份查询数据
	if svcCtx.UserID == id {
		ret, err = svcCtx.BlindBoxStatModel.GetTotal(context.Background(), int16(year), int16(month), 0)
	} else {
		ret, err = svcCtx.BlindBoxStatModel.GetTotalOnePersion(context.Background(), id, int16(year), int16(month), 0)
	}

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

	// 修改正则表达式以支持特定类型盲盒
	reg := `^(?:今日|(?:(?P<year>\d{4})年)?(?P<month>\d{1,2})月(?:(?P<day>\d{1,2})日)?)(?P<type>[^盲]+)盲盒$`
	re := regexp.MustCompile(reg)
	match := re.FindStringSubmatch(msg)

	if len(match) == 0 {
		return
	}

	now := carbon.Now(carbon.Local)
	var year, month, day int
	var err error

	if strings.HasPrefix(msg, "今日") {
		year = now.Year()
		month = now.Month()
		day = now.Day()
	} else {
		// 解析年份（可选）
		if match[1] != "" {
			year, err = strconv.Atoi(match[1])
			if err != nil || year < 2000 || year > 9999 {
				logic.PushToBulletSender(fmt.Sprintf("年份「%s」不正确!", match[1]), reply...)
				return
			}
		} else {
			year = now.Year()
		}

		// 解析月份（必需）
		month, err = strconv.Atoi(match[2])
		if err != nil || month < 1 || month > 12 {
			logic.PushToBulletSender(fmt.Sprintf("月份「%s」不正确!", match[2]), reply...)
			return
		}

		// 解析日期（可选）
		if match[3] != "" {
			day, err = strconv.Atoi(match[3])
			if err != nil || day < 1 || day > 31 {
				logic.PushToBulletSender(fmt.Sprintf("日期「%s」不正确!", match[3]), reply...)
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
		ret, err = svcCtx.BlindBoxStatModel.GetTotalByType(context.Background(), match[4], int16(year), int16(month), int16(day))
	} else {
		// 用户查询自己的数据
		ret, err = svcCtx.BlindBoxStatModel.GetTotalOnePersonByType(context.Background(), id, match[4], int16(year), int16(month), int16(day))
	}

	if err == nil {
		r := float64(ret.R) / float64(1000.0)
		var dateStr string
		if strings.HasPrefix(msg, "今日") {
			dateStr = fmt.Sprintf("%d年%d月%d日", year, month, day)
		} else {
			dateStr = fmt.Sprintf("%d年%d月", year, month)
			if day > 0 {
				dateStr = fmt.Sprintf("%d年%d月%d日", year, month, day)
			}
		}

		if ret.R > 0 {
			logic.PushToBulletSender(
				fmt.Sprintf(
					"%s%s盲盒共开%d个, 赚了＋%.2f元",
					dateStr,
					match[4],
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
					match[4],
					ret.C,
				),
				reply...,
			)
		} else {
			logic.PushToBulletSender(
				fmt.Sprintf(
					"%s%s盲盒共开%d个, 亏了－%.2f元",
					dateStr,
					match[4],
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
