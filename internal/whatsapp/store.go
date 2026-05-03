package whatsapp

import (
	"context"
	"fmt"
	"sync"

	"go.mau.fi/whatsmeow/store/sqlstore"
)

var (
	waStore     *sqlstore.Container
	waStoreErr  error
	waStoreOnce sync.Once
)

func getWAStore(ctx context.Context) (*sqlstore.Container, error) {
	waStoreOnce.Do(func() {
		dsn := "file:./data/whatsapp.db?cache=shared&_foreign_keys=on&_journal_mode=WAL&_busy_timeout=10000"
		container, err := sqlstore.New(ctx, "sqlite3", dsn, nil)
		if err != nil {
			waStoreErr = fmt.Errorf("open whatsapp store: %w", err)
			return
		}
		waStore = container
	})
	return waStore, waStoreErr
}
