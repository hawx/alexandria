package handlers

import (
	"github.com/hawx/alexandria/database"
	"github.com/hawx/alexandria/events"
	"github.com/hawx/alexandria/response"
	"github.com/hawx/alexandria/models"

	"github.com/gorilla/mux"

	"net/http"
	"encoding/json"
)

func Books(db database.Db, es *events.Source) booksHandler {
	return booksHandler{db, es}
}

type booksHandler struct {
	db database.Db
	es *events.Source
}

func (h booksHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response.ConvertBooks(h.db.Get()))
}

func (h booksHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	book, ok := h.db.Find(id)

	if !ok {
		w.WriteHeader(404)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(book)
}

func (h booksHandler) Update(w http.ResponseWriter, r *http.Request) {
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
}

func (h booksHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	book, ok := h.db.Find(id)

	if !ok {
		w.WriteHeader(404)
		return
	}

	h.db.Remove(book)
	h.es.Delete(book)

	w.WriteHeader(204)
}
