package db

import (
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"os"
	"strconv"
	"time"
)

type PGConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func NewPGInstance(cfg PGConfig) *sqlx.DB {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DBName,
	)

	config, _ := pgx.ParseConfig(dsn)
	// TODO : could be improved with logger
	//config.LogLevel = pgx.LogLevelTrace
	//config.Logger = NewDbLogger(l)

	idle, _ := strconv.Atoi(os.Getenv("DB_MAX_IDLE_CONNS"))
	if idle == 0 {
		idle = 2
	}
	open, _ := strconv.Atoi(os.Getenv("DB_MAX_OPEN_CONNS"))
	if open == 0 {
		open = 2
	}

	db := stdlib.OpenDB(*config)
	ddx := sqlx.NewDb(db, "pgx")
	ddx.SetConnMaxLifetime(5 * time.Minute)
	ddx.SetMaxOpenConns(open)
	ddx.SetMaxIdleConns(idle)

	err := ddx.Ping()
	if err != nil {
		//l.Error(err.Error())
		panic(err)
	}
	return ddx.Unsafe()
}
