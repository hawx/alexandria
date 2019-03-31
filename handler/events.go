package handler

import (
	"encoding/json"
	"net/http"

	"github.com/antage/eventsource"
	"hawx.me/code/alexandria/data"
)

type Source struct {
	es eventsource.EventSource
}

func Events() *Source {
	return &Source{eventsource.New(nil, nil)}
}

func (s *Source) Close() {
	s.es.Close()
}

func (s *Source) Add(book data.Book) {
	s.send("add", convertBook(book))
}

func (s *Source) Update(book data.Book) {
	s.send("update", convertBook(book))
}

func (s *Source) Delete(book data.Book) {
	s.send("delete", struct {
		ID string `json:"id"`
	}{book.ID})
}

func (s *Source) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.es.ServeHTTP(w, r)
}

func (s *Source) send(event string, data interface{}) {
	b, _ := json.Marshal(data)
	s.es.SendEventMessage(string(b), event, "")
}
