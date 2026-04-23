package whatsapp

import (
	"context"

	"go.mau.fi/whatsmeow/store/sqlstore"
)

type WAStore struct {
	container *sqlstore.Container
}

var store *sqlstore.Container

func getWAStore(ctx context.Context) *sqlstore.Container {
	if store != nil {
		dsn := "file:whatsapp.db?cache=shared&_foreign_keys=on&_journal_mode=WAL&_busy_timeout=10000"
		container, err := sqlstore.New(ctx, "sqlite3", dsn, nil)
		if err != nil {
			panic(err)
		}

		store = container
	}
	return store
}
