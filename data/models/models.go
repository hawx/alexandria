package models

import (
	"net/url"
	"time"
)

type Editions []*Edition

type Edition struct {
	Id          string  `json:"id"`
	Path        string  `json:"path"`
	ContentType string  `json:"content-type"`
	Extension   string  `json:"extension"`
}

type Books []*Book

type Book struct {
	Id       string     `json:"id"`
	Title    string     `json:"title"`
	Author   string     `json:"author"`
	Added    time.Time  `json:"added"`
  Editions Editions   `json:"editions"`
}

func (b Book) Slug(edition *Edition) string {
	return url.QueryEscape(b.Title + edition.Extension)
}
