package epub

// Based on: https://gitorious.org/go-pkg/epubgo

import (
	"archive/zip"
	"encoding/xml"
	"errors"
	"io"
)

type Epub interface {
	Metadata() (Metadata, error)
	Close() error
}

type epub struct {
	r *zip.Reader
	c io.Closer
}

type readSeekCloser interface {
	io.ReaderAt
	io.Seeker
	io.Closer
}

func Open(file readSeekCloser) (Epub, error) {
	size, err := file.Seek(0, 2)
	if err != nil {
		return nil, err
	}

	file.Seek(0, 0)

	r, err := zip.NewReader(file, size)
	if err != nil {
		return nil, err
	}

	return &epub{r, file}, nil
}

func (e *epub) Close() error {
	return e.c.Close()
}

func (e *epub) Metadata() (Metadata, error) {
	root, err := e.getRoot()
	if err != nil {
		return Metadata{}, err
	}

	return e.parse(root)
}

func (e *epub) getRoot() (string, error) {
	file, err := e.open("META-INF/container.xml")
	if err != nil {
		return "", err
	}

	var c struct {
		XMLName   xml.Name `xml:"container"`
		Rootfiles struct {
			FullPath string `xml:"full-path,attr"`
		} `xml:"rootfiles>rootfile"`
	}

	decoder := xml.NewDecoder(file)
	if err := decoder.Decode(&c); err != nil {
		return "", err
	}

	return c.Rootfiles.FullPath, nil
}

func (e *epub) parse(name string) (Metadata, error) {
	file, err := e.open(name)
	if err != nil {
		return Metadata{}, err
	}

	var c opf

	decoder := xml.NewDecoder(file)
	if err := decoder.Decode(&c); err != nil {
		return Metadata{}, err
	}

	return c.Metadata, nil
}

func (e *epub) open(name string) (io.ReadCloser, error) {
	for _, f := range e.r.File {
		if f.Name == name {
			return f.Open()
		}
	}

	return nil, errors.New("open: " + name + " not found")
}
