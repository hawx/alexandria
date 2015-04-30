package handlers

import (
	"hawx.me/code/alexandria/data"
	"hawx.me/code/alexandria/data/models"
	"hawx.me/code/alexandria/metadata"
	"hawx.me/code/alexandria/web/events"
	"hawx.me/code/mux"

	"code.google.com/p/go-uuid/uuid"

	"errors"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path"
	"time"
)

func Upload(db data.Db, es *events.Source, bookPath string) http.Handler {
	h := uploadHandler{db, es, bookPath}

	return mux.Method{
		"POST": h.Upload(),
	}
}

type uploadHandler struct {
	db       data.Db
	es       *events.Source
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

	edition := &models.Edition{
		Id:          uuid.New(),
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

	if contentType != models.MOBI && contentType != models.EPUB {
		return errors.New("Format not supported: " + contentType)
	}

	metaFunc := metadata.Epub
	if contentType == models.MOBI {
		metaFunc = metadata.Mobi
	}

	meta, _ := metaFunc(opened)

	newBook := models.Book{
		Id:       uuid.New(),
		Added:    time.Now(),
		Title:    meta.Title,
		Author:   meta.Author,
		Editions: models.Editions{edition},
	}

	h.db.Save(newBook)
	h.es.Add(newBook)

	go h.convert(contentType, newBook)

	return nil
}

func (h uploadHandler) convert(contentType string, book models.Book) {
	editions, err := h.convertAll(book)
	if err != nil {
		log.Println(err)
		return
	}

	book.Editions = append(book.Editions, editions...)

	h.db.Save(book)
	h.es.Update(book)
}

func (h uploadHandler) convertAll(book models.Book) ([]*models.Edition, error) {
	editions := []*models.Edition{}

	for _, contentType := range []string{models.MOBI, models.EPUB} {
		edition, err := h.convertEdition(book, contentType)
		if err != nil {
			return []*models.Edition{}, err
		}
		if edition != nil {
			editions = append(editions, edition)
		}
	}

	return editions, nil
}

func (h uploadHandler) convertEdition(book models.Book, contentType string) (*models.Edition, error) {
	for _, edition := range book.Editions {
		if edition.ContentType == contentType {
			return nil, nil
		}
	}

	edition := &models.Edition{
		Id:          uuid.New(),
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
