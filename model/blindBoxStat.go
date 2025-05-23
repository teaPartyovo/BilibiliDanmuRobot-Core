package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type (
	BlindBoxStatModel interface {
		Insert(ctx context.Context, tx *gorm.DB, data *BlindBoxStatBase) error
		GetTotalOnePersion(ctx context.Context, uid int64, year, month, day int16) (*Result, error)
		GetTotal(ctx context.Context, year, month, day int16) (*Result, error)
		GetTotalByType(ctx context.Context, boxType string, year, month, day int16) (*Result, error)
		GetTotalOnePersonByType(ctx context.Context, uid int64, boxType string, year, month, day int16) (*Result, error)
	}
	defaultBlindBoxStatModel struct {
		conn  *gorm.DB
		table string
	}
	BlindBoxStatBase struct {
		ID                int64 `gorm:"primaryKey;autoIncrement"`
		Uid               int64
		BlindBoxName      string
		Price             int32 // 爆出的礼物价格
		OriginalGiftPrice int32 // 原始盲盒价格
		Cnt               int32
		Year              int16
		Month             int16
		Day               int16
	}

	Result struct {
		C int
		R int64
	}
)

func NewBlindBoxStatModel(conn *gorm.DB, RoomID int64) BlindBoxStatModel {
	err := conn.Table(fmt.Sprintf("blind_%v", RoomID)).AutoMigrate(&BlindBoxStatBase{})
	if err != nil {
		logx.Error(err)
	}
	return &defaultBlindBoxStatModel{
		conn:  conn,
		table: fmt.Sprintf("blind_%v", RoomID),
	}
}

func (m *defaultBlindBoxStatModel) Insert(ctx context.Context, tx *gorm.DB, data *BlindBoxStatBase) error {
	db := m.conn
	if tx != nil {
		db = tx
	}
	err := db.WithContext(ctx).Table(m.table).Save(&data).Error
	return err
}

func (m *defaultBlindBoxStatModel) GetTotalOnePersion(ctx context.Context, uid int64, year, month, day int16) (*Result, error) {
	var resp Result

	d := m.conn.WithContext(ctx).Table(m.table).Model(&BlindBoxStatBase{}).Select(`sum(cnt) as C, (sum(cnt*Price)-sum(cnt*original_gift_price)) as R`).Where("uid = ?", uid)
	if year > 0 {
		d = d.Where("year = ?", year)
	}
	if month > 0 {
		d = d.Where("month = ?", month)
	}
	if day > 0 {
		d = d.Where("day = ?", day)
	}
	err := d.Take(&resp).Error

	switch err {
	case nil:
		return &resp, nil
	case ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultBlindBoxStatModel) GetTotal(ctx context.Context, year, month, day int16) (*Result, error) {
	var resp Result

	d := m.conn.WithContext(ctx).Table(m.table).Model(&BlindBoxStatBase{}).Select(`sum(cnt) as C, (sum(cnt*Price)-sum(cnt*original_gift_price)) as R`)
	if year > 0 {
		d = d.Where("year = ?", year)
	}
	if month > 0 {
		d = d.Where("month = ?", month)
	}
	if day > 0 {
		d = d.Where("day = ?", day)
	}
	err := d.Take(&resp).Error

	switch err {
	case nil:
		return &resp, nil
	case ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// 实现新增的方法
func (m *defaultBlindBoxStatModel) GetTotalByType(ctx context.Context, boxType string, year, month, day int16) (*Result, error) {
	var resp Result
	db := m.conn.WithContext(ctx).Table(m.table).
		Select("sum(cnt) as C, sum(cnt * Price - cnt * original_gift_price) as R").
		Where("blind_box_name LIKE ?", "%"+boxType+"%")

	if year > 0 {
		db = db.Where("year = ?", year)
	}
	if month > 0 {
		db = db.Where("month = ?", month)
	}
	if day > 0 {
		db = db.Where("day = ?", day)
	}

	err := db.Take(&resp).Error
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (m *defaultBlindBoxStatModel) GetTotalOnePersonByType(ctx context.Context, uid int64, boxType string, year, month, day int16) (*Result, error) {
	var resp Result
	db := m.conn.WithContext(ctx).Table(m.table).
		Select("sum(cnt) as C, sum(cnt * Price - cnt * original_gift_price) as R").
		Where("uid = ?", uid).
		Where("blind_box_name LIKE ?", "%"+boxType+"%")

	if year > 0 {
		db = db.Where("year = ?", year)
	}
	if month > 0 {
		db = db.Where("month = ?", month)
	}
	if day > 0 {
		db = db.Where("day = ?", day)
	}

	err := db.Take(&resp).Error
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
