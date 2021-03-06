package handler

import (
	"errors"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/google/uuid"
	"hawx.me/code/alexandria/data"
	"hawx.me/code/alexandria/metadata"
	"hawx.me/code/mux"
)

func Upload(db *data.DB, es *Source, bookPath string) http.Handler {
	h := uploadHandler{db, es, bookPath}

	return mux.Method{
		"POST": h.Upload(),
	}
}

type uploadHandler struct {
	db       *data.DB
	es       *Source
	bookPath string
}

func (h uploadHandler) Upload() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(100000)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		files := r.MultipartForm.File["file"]
		for _, file := range files {
			if err := h.doUpload(file); err != nil {
				log.Println(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	})
}

func (h uploadHandler) doUpload(fileheader *multipart.FileHeader) error {
	file, err := fileheader.Open()
	defer file.Close()

	if err != nil {
		return err
	}

	contentType := fileheader.Header["Content-Type"][0]

	edition := &data.Edition{
		ID:          uuid.New().String(),
		ContentType: contentType,
	}

	dstPath := path.Join(h.bookPath, edition.Path())

	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}

	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return err
	}

	opened, err := fileheader.Open()
	if err != nil {
		return err
	}

	if contentType != data.MOBI && contentType != data.EPUB {
		return errors.New("Format not supported: " + contentType)
	}

	metaFunc := metadata.Epub
	if contentType == data.MOBI {
		metaFunc = metadata.Mobi
	}

	meta, _ := metaFunc(opened)

	newBook := data.Book{
		ID:       uuid.New().String(),
		Added:    time.Now(),
		Title:    meta.Title,
		Author:   meta.Author,
		Editions: data.Editions{edition},
	}

	h.db.Save(newBook)
	h.es.Add(newBook)

	go h.convert(contentType, newBook)

	return nil
}

func (h uploadHandler) convert(contentType string, book data.Book) {
	editions, err := h.convertAll(book)
	if err != nil {
		log.Println(err)
		return
	}

	book.Editions = append(book.Editions, editions...)

	h.db.Save(book)
	h.es.Update(book)
}

func (h uploadHandler) convertAll(book data.Book) ([]*data.Edition, error) {
	editions := []*data.Edition{}

	for _, contentType := range []string{data.MOBI, data.EPUB} {
		edition, err := h.convertEdition(book, contentType)
		if err != nil {
			return []*data.Edition{}, err
		}
		if edition != nil {
			editions = append(editions, edition)
		}
	}

	return editions, nil
}

func (h uploadHandler) convertEdition(book data.Book, contentType string) (*data.Edition, error) {
	for _, edition := range book.Editions {
		if edition.ContentType == contentType {
			return nil, nil
		}
	}

	edition := &data.Edition{
		ID:          uuid.New().String(),
		ContentType: contentType,
	}

	from := path.Join(h.bookPath, book.Editions[0].Path())
	to := path.Join(h.bookPath, edition.Path())

	cmd := exec.Command("ebook-convert", from, to)
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return edition, nil
}
