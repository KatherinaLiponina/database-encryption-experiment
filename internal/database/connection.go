package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type ConnectionConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	DatabaseName string
}

type Connection interface {
	Select(query string, args ...any) (*sql.Rows, error)
	Insert(query string, args ...any) error
	Exec(query string) error
	Close() error
}

type connection struct {
	conn *sql.DB
}

func NewConnection(cfg ConnectionConfig) (Connection, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DatabaseName)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return &connection{}, err
	}
	return &connection{conn: db}, nil
}

func (c *connection) Close() error {
	return c.conn.Close()
}

func (c *connection) Select(query string, args ...any) (*sql.Rows, error) {
	rows, err := c.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return rows, nil
}

func (c *connection) Insert(query string, args ...any) error {
	_, err := c.conn.Exec(query, args...)
	return err
}

func (c *connection) Exec(query string) error {
	_, err := c.conn.Exec(query)
	return err
}
