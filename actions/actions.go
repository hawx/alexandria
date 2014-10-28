package actions

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

func editionPath(bookPath, id, contentType string) string {
	return path.Join(bookPath, id+extension(contentType))
}

func Upload(bookPath string, fileheader *multipart.FileHeader, db database.Db, es *events.Source) error {
	file, err := fileheader.Open()
	defer file.Close()

	if err != nil {
		return err
	}

	contentType := fileheader.Header["Content-Type"][0]
	newBook := models.Book{Id: uuid.New(), Added: time.Now()}
	editionId := uuid.New()
	dstPath := editionPath(bookPath, editionId, contentType)

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

	db.Save(newBook)
	es.Add(newBook)

	go convert(bookPath, contentType, newBook, db, es)

	return nil
}

func convert(bookPath string, contentType string, book models.Book, db database.Db, es *events.Source) {
	editions, err := convertAll(bookPath, book)
	if err != nil {
		log.Println(err)
		return
	}

	book.Editions = append(book.Editions, editions...)

	db.Save(book)
	es.Update(book)
}

func convertAll(bookPath string, book models.Book) ([]*models.Edition, error) {
	editions := []*models.Edition{}

	for _, contentType := range []string{MOBI, EPUB} {
		edition, err := convertEdition(bookPath, book, contentType)
		if err != nil {
			return []*models.Edition{}, err
		}
		if edition != nil {
			editions = append(editions, edition)
		}
	}

	return editions, nil
}

func convertEdition(bookPath string, book models.Book, contentType string) (*models.Edition, error) {
	for _, edition := range book.Editions {
		if edition.ContentType == contentType {
			return nil, nil
		}
	}

	editionId := uuid.New()
	from := book.Editions[0].Path
	to := editionPath(bookPath, editionId, contentType)

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