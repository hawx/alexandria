package models

import (
	"net/url"
	"time"
)

const (
	EPUB = "application/epub+zip"
	MOBI = "application/x-mobipocket-ebook"
)

type Editions []*Edition

type Edition struct {
	Id          string `json:"id"`
	ContentType string `json:"content-type"`
}

func (e Edition) Extension() string {
	switch e.ContentType {
	case EPUB:
		return ".epub"
	case MOBI:
		return ".mobi"
	}
	return ""
}

func (e Edition) Path() string {
	return e.Id + e.Extension()
}

type Books []*Book

type Book struct {
	Id       string    `json:"id"`
	Title    string    `json:"title"`
	Author   string    `json:"author"`
	Added    time.Time `json:"added"`
	Editions Editions  `json:"editions"`
}

func (b Book) Slug(edition *Edition) string {
	return url.QueryEscape(b.Title + edition.Extension())
}
