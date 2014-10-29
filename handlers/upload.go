package handlers

import (
	"github.com/hawx/alexandria/database"
	"github.com/hawx/alexandria/epub"
	"github.com/hawx/alexandria/events"
	"github.com/hawx/alexandria/mobi"
	"github.com/hawx/alexandria/models"

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

const (
	EPUB = "application/epub+zip"
	MOBI = "application/x-mobipocket-ebook"
)

func extension(contentType string) string {
	switch contentType {
	case EPUB:
		return ".epub"
	case MOBI:
		return ".mobi"
	}
	return ""
}

func Upload(db database.Db, es *events.Source, bookPath string) uploadHandler {
	return uploadHandler{db, es, bookPath}
}

type uploadHandler struct {
	db       database.Db
	es       *events.Source
	bookPath string
}

func (h uploadHandler) Upload(w http.ResponseWriter, r *http.Request) {
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
}

func (h uploadHandler) editionPath(id, contentType string) string {
	return path.Join(h.bookPath, id+extension(contentType))
}

func (h uploadHandler) doUpload(fileheader *multipart.FileHeader) error {
	file, err := fileheader.Open()
	defer file.Close()

	if err != nil {
		return err
	}

	contentType := fileheader.Header["Content-Type"][0]
	newBook := models.Book{Id: uuid.New(), Added: time.Now()}
	editionId := uuid.New()
	dstPath := h.editionPath(editionId, contentType)

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

	switch contentType {
	case EPUB:
		book, _ := epub.Open(opened)
		meta, _ := book.Metadata()

		newBook.Title = meta.Title[0]
		newBook.Author = meta.Creator[0].Value
		newBook.Editions = models.Editions{
			{
				Id:          editionId,
				Path:        dstPath,
				ContentType: EPUB,
				Extension:   extension(EPUB),
			},
		}

	case MOBI:
		book, _ := mobi.Open(opened)
		meta, _ := book.Metadata()

		newBook.Title = meta.Title
		newBook.Author = meta.Creator
		newBook.Editions = models.Editions{
			{
				Id:          editionId,
				Path:        dstPath,
				ContentType: MOBI,
				Extension:   extension(MOBI),
			},
		}

	default:
		return errors.New("Format not supported: " + contentType)
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

	for _, contentType := range []string{MOBI, EPUB} {
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

	editionId := uuid.New()
	from := book.Editions[0].Path
	to := h.editionPath(editionId, contentType)

	cmd := exec.Command("ebook-convert", from, to)
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return &models.Edition{
		Id:          editionId,
		Path:        to,
		ContentType: contentType,
		Extension:   extension(contentType),
	}, nil
}
