package gorm_driver

import (
	"context"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Driver struct {
	*gorm.DB
}

type MysqlConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
}

func (c MysqlConfig) ToDsn() string {
	return fmt.Sprintf(
		`%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local`,
		c.Username,
		c.Password,
		c.Host,
		c.Port,
		c.Database)
}

type Option struct {
	DbOptions []gorm.Option
}

func NewDriver(config MysqlConfig, options Option) (*Driver, error) {
	db, err := gorm.Open(mysql.Open(config.ToDsn()), options.DbOptions...)
	if err != nil {
		return nil, err
	}
	return &Driver{
		db,
	}, nil
}

const (
	talentIdKey = "talent-id"
)

func WithTalentID(db *gorm.DB) *gorm.DB {
	ctx := db.Statement.Context
	tid, ok := ctx.Value(talentIdKey).(string)
	if ok {
		db = db.Where("talent_id = ?", tid)
	}
	return db
}

func (d *Driver) Begin(ctx context.Context, callback func(ctx context.Context, db *gorm.DB) error) error {
	tx, ok := ctx.Value("tx").(*gorm.DB)
	if !ok {
		tx = d.DB.Begin()
		ctx = context.WithValue(ctx, "tx", tx)
		tx = tx.WithContext(ctx)
	}

	if err := callback(ctx, tx); err != nil {
		if err := tx.Rollback(); err != nil {
			return tx.Error
		}
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}
