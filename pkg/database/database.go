package database

import (
	"context"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

type Provider interface {
	GetDB(ctx context.Context) *gorm.DB
}

type DB interface {
	Provider
	Close() error
	Migrate(dialect string) error
}

type Database struct {
	db *gorm.DB
}

func NewDatabase(uri string, opts ...Option) (*Database, error) {
	logger := gormLogger.New(log.New(os.Stderr, "\r\n", log.LstdFlags), gormLogger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  gormLogger.Warn,
		IgnoreRecordNotFoundError: true,
		Colorful:                  false,
	})

	database, err := gorm.Open(postgres.Open(uri), &gorm.Config{
		Logger: logger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := database.DB()
	if err != nil {
		return nil, err
	}

	opt := getOption(opts...)

	sqlDB.SetMaxOpenConns(opt.MaxOpenConns)
	sqlDB.SetMaxIdleConns(opt.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(opt.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(opt.ConnMaxIdleTime)

	return &Database{
		db: database,
	}, nil
}

func (d *Database) Close() error {
	db, err := d.db.DB()
	if err != nil {
		return err
	}
	return db.Close()
}

func (d *Database) GetDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return d.db.WithContext(ctx)
}

func (d *Database) Migrate(dialect string) error {
	db, err := d.db.DB()
	if err != nil {
		return err
	}

	return Migrate(db, dialect)
}
