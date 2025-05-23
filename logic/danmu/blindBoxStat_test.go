package danmu

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/xbclub/BilibiliDanmuRobot-Core/config"
	"github.com/xbclub/BilibiliDanmuRobot-Core/entity"
	"github.com/xbclub/BilibiliDanmuRobot-Core/model"
	"github.com/xbclub/BilibiliDanmuRobot-Core/svc"
)

func TestDoBlindBoxStat(t *testing.T) {
	// 模拟 ServiceContext
	svcCtx := &svc.ServiceContext{
		Config: &config.Config{
			BlindBoxStat: true,
		},
		UserID: 12345, // 模拟主播ID
	}

	// 模拟弹幕消息
	danmuMsgs := []struct {
		name     string
		danmuMsg string
		uid      string
		username string
		want     string
	}{
		{
			name:     "测试普通盲盒查询",
			danmuMsg: "1月盲盒",
			uid:      "12345",
			username: "主播",
			want:     "1月共开",
		},
		{
			name:     "测试特定盲盒查询",
			danmuMsg: "1月心动盲盒",
			uid:      "54321",
			username: "观众A",
			want:     "1月心动盲盒共开",
		},
		{
			name:     "测试错误格式",
			danmuMsg: "一月盲盒",
			uid:      "54321",
			username: "观众B",
			want:     "",
		},
	}

	// 模拟弹幕发送
	for _, dm := range danmuMsgs {
		t.Run(dm.name, func(t *testing.T) {
			// 构造弹幕消息
			danmuText := &entity.DanmuMsgText{
				Info: []interface{}{
					[]interface{}{
						0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
						map[string]interface{}{
							"extra": `{"reply_uname":"","reply_mid":0}`,
						},
					},
					dm.danmuMsg,
					[]interface{}{
						dm.uid,
						dm.username,
					},
				},
			}

			// 转换为JSON
			jsonData, err := json.Marshal(danmuText)
			if err != nil {
				t.Errorf("JSON序列化失败: %v", err)
				return
			}

			// 发送到弹幕处理通道
			PushToBDanmuLogic(string(jsonData))

			// 等待处理完成
			time.Sleep(100 * time.Millisecond)
		})
	}
}

func TestDoBlindBoxStatByType(t *testing.T) {
	svcCtx := &svc.ServiceContext{
		Config: &config.Config{
			BlindBoxStat: true,
		},
		UserID: 12345,
	}

	// 模拟弹幕消息
	danmuMsgs := []struct {
		name     string
		danmuMsg string
		uid      string
		username string
		want     string
	}{
		{
			name:     "测试普通盲盒查询",
			danmuMsg: "1月盲盒",
			uid:      "12345",
			username: "主播",
			want:     "1月共开",
		},
		{
			name:     "测试特定盲盒查询",
			danmuMsg: "1月心动盲盒",
			uid:      "54321",
			username: "观众A",
			want:     "1月心动盲盒共开",
		},
		{
			name:     "测试错误格式",
			danmuMsg: "一月盲盒",
			uid:      "54321",
			username: "观众B",
			want:     "",
		},
	}

	// 模拟弹幕发送
	for _, dm := range danmuMsgs {
		t.Run(dm.name, func(t *testing.T) {
			// 构造弹幕消息
			danmuText := &entity.DanmuMsgText{
				Info: []interface{}{
					[]interface{}{
						0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
						map[string]interface{}{
							"extra": `{"reply_uname":"","reply_mid":0}`,
						},
					},
					dm.danmuMsg,
					[]interface{}{
						dm.uid,
						dm.username,
					},
				},
			}

			// 转换为JSON
			jsonData, err := json.Marshal(danmuText)
			if err != nil {
				t.Errorf("JSON序列化失败: %v", err)
				return
			}

			// 发送到弹幕处理通道
			PushToBDanmuLogic(string(jsonData))

			// 等待处理完成
			time.Sleep(100 * time.Millisecond)
		})
	}
}

func TestSaveBlindBoxStat(t *testing.T) {
	svcCtx := &svc.ServiceContext{
		Config: &config.Config{
			BlindBoxStat: true,
		},
	}

	testGift := &entity.SendGiftText{
		Data: entity.GiftData{
			UID:   12345,
			Price: 1000,
			Num:   1,
			BlindGift: entity.BlindGift{
				OriginalGiftName:  "心动盲盒",
				OriginalGiftPrice: 500,
			},
		},
	}

	SaveBlindBoxStat(testGift, svcCtx)
}

func TestDanmuBlindBoxStat(t *testing.T) {
	// 初始化数据库连接
	db := &model.BlindBoxStatModel{} // 这里需要实际的数据库连接

	// 模拟 ServiceContext
	svcCtx := &svc.ServiceContext{
		Config: &config.Config{
			BlindBoxStat: true,
		},
		UserID:            12345,
		BlindBoxStatModel: db,
	}

	// 启动弹幕处理服务
	ctx := context.Background()
	go StartDanmuLogic(ctx, svcCtx)

	// 等待服务启动
	time.Sleep(100 * time.Millisecond)

	danmuMsgs := []struct {
		name     string
		danmuMsg string
		uid      string
		username string
		want     string
	}{
		{
			name:     "测试普通盲盒查询",
			danmuMsg: "1月盲盒",
			uid:      "12345",
			username: "主播",
			want:     "1月共开",
		},
		{
			name:     "测试特定盲盒查询",
			danmuMsg: "1月心动盲盒",
			uid:      "54321",
			username: "观众A",
			want:     "1月心动盲盒共开",
		},
		{
			name:     "测试错误格式",
			danmuMsg: "一月盲盒",
			uid:      "54321",
			username: "观众B",
			want:     "",
		},
	}

	for _, dm := range danmuMsgs {
		t.Run(dm.name, func(t *testing.T) {
			danmuText := &entity.DanmuMsgText{
				Info: []interface{}{
					[]interface{}{
						0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
						map[string]interface{}{
							"extra": `{"reply_uname":"","reply_mid":0}`,
						},
					},
					dm.danmuMsg,
					[]interface{}{
						dm.uid,
						dm.username,
					},
				},
			}

			jsonData, err := json.Marshal(danmuText)
			if err != nil {
				t.Errorf("JSON序列化失败: %v", err)
				return
			}

			PushToBDanmuLogic(string(jsonData))
			time.Sleep(100 * time.Millisecond)
		})
	}
}