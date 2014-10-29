package handlers

import (
	"github.com/hawx/alexandria/database"

	"github.com/gorilla/mux"

	"net/http"
	"io"
	"strconv"
	"os"
)

func Editions(db database.Db) editionsHandler {
	return editionsHandler{db}
}

type editionsHandler struct {
	db database.Db
}

func (h editionsHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	edition, book, ok := h.db.FindEdition(id)

	if !ok {
		w.WriteHeader(404)
		return
	}

	file, err := os.Open(edition.Path)
	if err != nil {
		w.WriteHeader(500)
	}

	stat, err := file.Stat()
	if err != nil {
		w.WriteHeader(500)
	}

	w.Header().Set("Content-Type", edition.ContentType)
	w.Header().Set("Content-Disposition", `attachment; filename=`+book.Slug(edition))
	w.Header().Set("Content-Length", strconv.Itoa(int(stat.Size())))
	io.Copy(w, file)
}
