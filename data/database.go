package data

import (
	"hawx.me/code/alexandria/data/models"

	"github.com/boltdb/bolt"

	"encoding/json"
	"fmt"
	"log"
)

type Db interface {
	Get() models.Books
	Find(string) (models.Book, bool)
	FindEdition(string) (*models.Edition, *models.Book, bool)
	Save(models.Book)
	Remove(models.Book)
	Close()
}

type BoltDb struct {
	me *bolt.DB
}

const bucketName = "alexandria"

func Open(path string) Db {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(bucketName))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	return BoltDb{db}
}

func (db BoltDb) Get() models.Books {
	list := models.Books{}

	db.me.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var book models.Book
			json.Unmarshal(v, &book)
			list = append(list, &book)
		}

		return nil
	})

	return list
}

func (db BoltDb) Find(id string) (book models.Book, ok bool) {
	db.me.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		v := b.Get([]byte(id))
		if v != nil {
			json.Unmarshal(v, &book)
			ok = true
		}
		return nil
	})

	return book, ok
}

func (db BoltDb) FindEdition(id string) (*models.Edition, *models.Book, bool) {
	books := db.Get()

	for _, book := range books {
		for _, edition := range book.Editions {
			if id == edition.Id {
				return edition, book, true
			}
		}
	}

	return nil, nil, false
}

func (db BoltDb) Save(book models.Book) {
	key := book.Id
	serialised, _ := json.Marshal(book)

	db.me.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		return b.Put([]byte(key), serialised)
	})
}

func (db BoltDb) Remove(book models.Book) {
	db.me.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		return b.Delete([]byte(book.Id))
	})
}

func (db BoltDb) Close() {
	db.me.Close()
}
