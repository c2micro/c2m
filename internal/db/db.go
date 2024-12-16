package db

import (
	"context"

	"github.com/c2micro/c2msrv/internal/ent"

	"github.com/go-faster/errors"
)

// инициализация БД в зависимости от конфига
func (c ConfigV1) Init(ctx context.Context) (*ent.Client, error) {
	// инициализация sqlite3
	if c.Sqlite != nil {
		// инициализируем sqlite БД
		db, err := c.Sqlite.Init(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "init sqlite")
		}
		return db, nil
	}

	// инициализация psql
	if c.Postgresql != nil {
		// TODO инициализируем postgresql БД
		return nil, errors.New("postgresql not supported yet")
	}

	// сюда мы не должны попасть, т.к. есть валидатор
	return nil, errors.New("no db selected for processing")
}
