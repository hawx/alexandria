package handler

import (
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"

	"hawx.me/code/alexandria/data"
	"hawx.me/code/mux"
	"hawx.me/code/route"
)

func Editions(db *data.DB, bookPath string) http.Handler {
	h := editionsHandler{db, bookPath}

	return mux.Method{
		"GET": h.Get(),
	}
}

type editionsHandler struct {
	db       *data.DB
	bookPath string
}

func (h editionsHandler) Get() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := route.Vars(r)["id"]
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
