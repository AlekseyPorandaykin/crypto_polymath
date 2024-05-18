package database

import (
	"fmt"
	"github.com/jmoiron/sqlx"

	//drivers
	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/mattn/go-sqlite3"
)

type Config struct {
	Driver             string
	Username           string
	Password           string
	Host               string
	Port               string
	Database           string
	MaxOpenConnections int
	MaxIdleConnections int

	PathToDB string //for sqlite
}

func CreateConnection(conf Config) (*sqlx.DB, error) {
	conn, err := connection(conf)
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return conn, nil
}

func connection(conf Config) (*sqlx.DB, error) {
	switch conf.Driver {
	case "postgres":
		return postgresConnection(conf)
	case "sqlite":
		return sqliteConnection(conf)
	default:
		return nil, fmt.Errorf("not found connection for driver: %s", conf.Driver)
	}
}

func postgresConnection(conf Config) (*sqlx.DB, error) {
	conn, err := sqlx.Connect(
		"pgx",
		fmt.Sprintf(
			"%s://%s:%s@%s:%s/%s",
			conf.Driver,
			conf.Username,
			conf.Password,
			conf.Host,
			conf.Port,
			conf.Database,
		),
	)
	if err != nil {
		return nil, err
	}
	if conf.MaxOpenConnections > 0 {
		conn.SetMaxOpenConns(conf.MaxOpenConnections)
	}
	if conf.MaxIdleConnections > 0 {
		conn.SetMaxIdleConns(conf.MaxIdleConnections)
	}
	return conn, nil
}

func sqliteConnection(conf Config) (*sqlx.DB, error) {
	conn, err := sqlx.Open("sqlite3", conf.PathToDB)
	if err != nil {
		return nil, err
	}
	if conf.MaxOpenConnections > 0 {
		conn.SetMaxOpenConns(conf.MaxOpenConnections)
	}
	if conf.MaxIdleConnections > 0 {
		conn.SetMaxIdleConns(conf.MaxIdleConnections)
	}
	return conn, nil
}
