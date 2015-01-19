package metadata

import (
	"github.com/hawx/alexandria/metadata/epub"
	"github.com/hawx/alexandria/metadata/mobi"

	"io"
)

type Metadata struct {
	Title  string
	Author string
}

type readSeekCloser interface {
	io.Reader
	io.ReaderAt
	io.Seeker
	io.Closer
}

func Epub(file readSeekCloser) (Metadata, error) {
	book, err := epub.Open(file)
	if err != nil {
		return Metadata{}, err
	}

	meta, err := book.Metadata()
	if err != nil {
		return Metadata{}, err
	}

	return Metadata{meta.Title[0], meta.Creator[0].Value}, nil
}

func Mobi(file readSeekCloser) (Metadata, error) {
	book, err := mobi.Open(file)
	if err != nil {
		return Metadata{}, err
	}

	meta, err := book.Metadata()
	if err != nil {
		return Metadata{}, err
	}

	return Metadata{meta.Title, meta.Creator}, nil
}
