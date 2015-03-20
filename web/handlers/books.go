package handlers

import (
	"github.com/hawx/alexandria/data"
	"github.com/hawx/alexandria/data/models"
	"github.com/hawx/alexandria/web/events"
	"github.com/hawx/alexandria/web/response"
	"github.com/hawx/route"

	"github.com/hawx/mux"

	"encoding/json"
	"net/http"
)

func AllBooks(db data.Db, es *events.Source) http.Handler {
	return mux.Method{
		"GET": booksHandler{db, es}.GetAll(),
	}
}

func Books(db data.Db, es *events.Source) http.Handler {
	h := booksHandler{db, es}

	return mux.Method{
		"GET":    h.Get(),
		"PATCH":  h.Update(),
		"DELETE": h.Delete(),
	}
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
		id := route.Vars(r)["id"]
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
		id := route.Vars(r)["id"]
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
		id := route.Vars(r)["id"]
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
