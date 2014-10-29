package handlers

import (
	"github.com/hawx/alexandria/data"

	"github.com/gorilla/mux"

	"io"
	"net/http"
	"os"
	"strconv"
)

func Editions(db data.Db) EditionsHandler {
	h := editionsHandler{db}

	return EditionsHandler{
	  Get: h.Get(),
	}
}

type EditionsHandler struct {
	Get http.Handler
}

type editionsHandler struct {
	db data.Db
}

func (h editionsHandler) Get() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	})
}
