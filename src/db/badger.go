package db

import (
	"encoding/json"
	"log"

	"github.com/dgraph-io/badger/v4"
)

var conn *badger.DB

func Open() {
	opts := badger.DefaultOptions("./data/")
	opts.Logger = nil
	var err error
	conn, err = badger.Open(opts)
	if err != nil {
		log.Fatalf("error opening db: %v", err)
	}
}

func Close() {
	err := conn.Close()
	if err != nil {
		log.Fatalf("error closing db: %v", err)
	}
}

func Set(k string, v interface{}) error {
	return conn.Update(func(txn *badger.Txn) error {
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		err = txn.Set([]byte(k), b)
		return err
	})
}

func Get(k string, v interface{}) error {
	return conn.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(k))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, v)
		})
	})
}

func Exists(k string) error {
	return conn.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte(k))
		return err
	})
}

func Delete(k string) error {
	return conn.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(k))
	})
}
