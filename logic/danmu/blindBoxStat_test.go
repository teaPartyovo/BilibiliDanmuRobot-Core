package danmu

import (
	"testing"

	"github.com/xbclub/BilibiliDanmuRobot-Core/entity"
	"github.com/xbclub/BilibiliDanmuRobot-Core/svc"
)

func TestBlindBoxStatCommands(t *testing.T) {
	// 模拟 ServiceContext
	svcCtx := &svc.ServiceContext{
		Config: struct{ BlindBoxStat bool }{BlindBoxStat: true},
		UserID: 12345,
	}

	// 测试用例
	tests := []struct {
		name     string
		msg      string
		uid      string
		username string
	}{
		{"今日盲盒", "今日盲盒", "12345", "测试用户"},
		{"月份盲盒", "3月盲盒", "12345", "测试用户"},
		{"日期盲盒", "3月15日盲盒", "12345", "测试用户"},
		{"年月盲盒", "2024年3月盲盒", "12345", "测试用户"},
		{"完整日期盲盒", "2024年3月15日盲盒", "12345", "测试用户"},
		{"今日特定盲盒", "今日心动盲盒", "12345", "测试用户"},
		{"月份特定盲盒", "3月星月盲盒", "12345", "测试用户"},
		{"日期特定盲盒", "3月15日星月盲盒", "12345", "测试用户"},
		{"年月特定盲盒", "2024年3月心动盲盒", "12345", "测试用户"},
		{"完整日期特定盲盒", "2024年3月15日心动盲盒", "12345", "测试用户"},
	}

	// 执行测试
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试普通盲盒统计
			DoBlindBoxStat(tt.msg, tt.uid, tt.username, svcCtx)

			// 如果包含特定类型，测试特定类型盲盒统计
			if tt.name != "今日盲盒" && tt.name != "月份盲盒" && tt.name != "日期盲盒" && tt.name != "年月盲盒" && tt.name != "完整日期盲盒" {
				DoBlindBoxStatByType(tt.msg, tt.uid, tt.username, svcCtx)
			}
		})
	}
}
