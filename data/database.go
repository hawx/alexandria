package data

import (
	"encoding/json"
	"fmt"

	"github.com/boltdb/bolt"
)

type DB struct {
	me *bolt.DB
}

const bucketName = "alexandria"

func Open(path string) (*DB, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	return &DB{db}, err
}

func (db *DB) Get() (Books, error) {
	list := Books{}

	err := db.me.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var book Book
			if err := json.Unmarshal(v, &book); err != nil {
				return err
			}
			list = append(list, &book)
		}

		return nil
	})

	return list, err
}

func (db *DB) Find(id string) (book Book, ok bool) {
	if err := db.me.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		v := b.Get([]byte(id))
		if v != nil {
			json.Unmarshal(v, &book)
			ok = true
		}
		return nil
	}); err != nil {
		return book, false
	}

	return book, ok
}

func (db *DB) FindEdition(id string) (*Edition, *Book, bool) {
	books, err := db.Get()
	if err != nil {
		return nil, nil, false
	}

	for _, book := range books {
		for _, edition := range book.Editions {
			if id == edition.ID {
				return edition, book, true
			}
		}
	}

	return nil, nil, false
}

func (db *DB) Save(book Book) error {
	key := book.ID
	serialised, err := json.Marshal(book)
	if err != nil {
		return err
	}

	return db.me.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		return b.Put([]byte(key), serialised)
	})
}

func (db *DB) Remove(book Book) error {
	return db.me.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		return b.Delete([]byte(book.ID))
	})
}

func (db *DB) Close() error {
	return db.me.Close()
}
