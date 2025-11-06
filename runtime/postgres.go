package runtime

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Postgres struct {
	*gorm.DB
	sqlDB *sql.DB
}

func newPostgres(config *Config) (*Postgres, error) {
	ssl := "disable"
	if config.Deps.Postgres.SSL {
		ssl = "require"
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s timezone=%s",
		config.Deps.Postgres.Host,
		config.Deps.Postgres.User,
		config.Deps.Postgres.Password,
		config.Deps.Postgres.DB,
		config.Deps.Postgres.Port,
		ssl,
		config.Deps.Postgres.TimeZone,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(config.Deps.Postgres.Pool.MaxIdle)
	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(config.Deps.Postgres.Pool.MaxOpen)
	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(time.Duration(config.Deps.Postgres.Pool.MaxLife) * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping postgres failed: %w", err)
	}

	return &Postgres{
		DB:    db,
		sqlDB: sqlDB,
	}, nil
}

func (p *Postgres) Close() error {
	return p.sqlDB.Close()
}
