package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"strings"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/c2micro/c2msrv/internal/ent"
	"github.com/go-faster/errors"
	"modernc.org/sqlite"

	"github.com/go-faster/sdk/zctx"
	"go.uber.org/zap"
)

// структура для sqlite3 драйвера
type ConfigSqliteV1 struct {
	Path string `json:"path" validate:"required"`
}

// инициализация sqlite3
func (c *ConfigSqliteV1) Init(ctx context.Context) (*ent.Client, error) {
	lg := zctx.From(ctx).Named("sqlite")

	// составление dsn подключения
	var d strings.Builder
	// TODO добавить больше возможных ключей через конфиг
	d.WriteString("file:")
	d.WriteString(c.Path)
	d.WriteString("?cache=shared&_fk=1")

	lg.Debug("connection dsn", zap.String("dsn", d.String()))

	// открытие кона к БД
	db, err := sql.Open("sqlite3", d.String())
	if err != nil {
		return nil, err
	}
	// чтобы был одновременный доступ к БД без постоянных ошибок "database is locked"
	db.SetMaxOpenConns(1)

	// создаем клиента
	drv := entsql.OpenDB(dialect.SQLite, db)
	client := ent.NewClient(ent.Driver(drv))
	lg.Debug("connected to database")

	// применение автомиграций
	if err = client.Schema.Create(ctx); err != nil {
		// закрываем кон к БД
		_ = client.Close()
		return nil, err
	}
	lg.Debug("schema migrated")

	return client, nil
}

type sqlite3Driver struct {
	*sqlite.Driver
}

// открытие кона к sqlite3 БД
func (d sqlite3Driver) Open(name string) (driver.Conn, error) {
	// создаем conn
	conn, err := d.Driver.Open(name)
	if err != nil {
		return conn, err
	}
	c := conn.(interface {
		Exec(stmt string, args []driver.Value) (driver.Result, error)
	})
	// включение необходимых фичей
	if _, err = c.Exec("PRAGMA foreign_keys = on;", nil); err != nil {
		_ = conn.Close()
		return nil, errors.Wrap(err, "failed to enable foreign keys")
	}
	if _, err = c.Exec("PRAGMA journal_mode = WAL;", nil); err != nil {
		_ = conn.Close()
		return nil, errors.Wrap(err, "failed to enable WAL mode")
	}
	if _, err = c.Exec("PRAGMA synchronous=NORMAL;", nil); err != nil {
		_ = conn.Close()
		return nil, errors.Wrap(err, "failed to enable normal synchronous")
	}
	return conn, nil
}

func init() {
	// регистрация sqlite3 драйвера для использования с ent
	sql.Register("sqlite3", sqlite3Driver{Driver: &sqlite.Driver{}})
}
