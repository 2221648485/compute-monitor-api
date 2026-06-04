package mysql

import (
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Options struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// New creates a GORM MySQL database handle and configures the underlying pool.
func New(opts Options) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(opts.DSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(opts.MaxOpenConns)
	sqlDB.SetMaxIdleConns(opts.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(opts.ConnMaxLifetime)

	return db, nil
}

// Close closes the underlying sql.DB owned by GORM.
func Close(db *gorm.DB) error {
	if db == nil {
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}
