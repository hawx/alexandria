package handlers

import (
	"github.com/hawx/alexandria/data"
	"github.com/hawx/alexandria/data/models"
	"github.com/hawx/alexandria/web/events"
	"github.com/hawx/alexandria/web/response"

	"github.com/gorilla/mux"

	"encoding/json"
	"net/http"
)

func Books(db data.Db, es *events.Source) BooksHandler {
	h := booksHandler{db, es}

	return BooksHandler{
		GetAll: h.GetAll(),
		Get:    h.Get(),
		Update: h.Update(),
		Delete: h.Delete(),
	}
}

type BooksHandler struct {
	GetAll http.Handler
	Get    http.Handler
	Update http.Handler
	Delete http.Handler
}

type booksHandler struct {
	db data.Db
	es *events.Source
}

func (h booksHandler) GetAll() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response.ConvertBooks(h.db.Get()))
	})
}

func (h booksHandler) Get() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		book, ok := h.db.Find(id)

		if !ok {
			w.WriteHeader(404)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(book)
	})
}

func (h booksHandler) Update() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		book, ok := h.db.Find(id)

		if !ok {
			w.WriteHeader(404)
			return
		}

		var req models.Book
		json.NewDecoder(r.Body).Decode(&req)

		if req.Title != "" {
			book.Title = req.Title
		}
		if req.Author != "" {
			book.Author = req.Author
		}

		h.db.Save(book)
		h.es.Update(book)

		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(book)
	})
}

func (h booksHandler) Delete() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		book, ok := h.db.Find(id)

		if !ok {
			w.WriteHeader(404)
			return
		}

		h.db.Remove(book)
		h.es.Delete(book)

		w.WriteHeader(204)
	})
}
