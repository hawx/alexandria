package handlers

import (
	"github.com/hawx/alexandria/data"

	"github.com/gorilla/mux"

	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
)

func Editions(db data.Db, bookPath string) EditionsHandler {
	h := editionsHandler{db, bookPath}

	return EditionsHandler{
		Get: h.Get(),
	}
}

type EditionsHandler struct {
	Get http.Handler
}

type editionsHandler struct {
	db       data.Db
	bookPath string
}

func (h editionsHandler) Get() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		edition, book, ok := h.db.FindEdition(id)

		if !ok {
			w.WriteHeader(404)
			return
		}

		file, err := os.Open(path.Join(h.bookPath, edition.Path()))
		if err != nil {
			log.Println(err)
			w.WriteHeader(500)
			return
		}

		stat, err := file.Stat()
		if err != nil {
			log.Println(err)
			w.WriteHeader(500)
			return
		}

		w.Header().Set("Content-Type", edition.ContentType)
		w.Header().Set("Content-Disposition", `attachment; filename=`+book.Slug(edition))
		w.Header().Set("Content-Length", strconv.Itoa(int(stat.Size())))
		io.Copy(w, file)
	})
}
