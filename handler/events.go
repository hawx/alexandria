package handler

import (
	"encoding/json"
	"net/http"

	"github.com/antage/eventsource"
	"hawx.me/code/alexandria/data/models"
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

func (s *Source) Add(book models.Book) {
	s.send("add", convertBook(book))
}

func (s *Source) Update(book models.Book) {
	s.send("update", convertBook(book))
}

func (s *Source) Delete(book models.Book) {
	s.send("delete", struct {
		Id string `json:"id"`
	}{book.Id})
}

func (s *Source) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.es.ServeHTTP(w, r)
}

func (s *Source) send(event string, data interface{}) {
	b, _ := json.Marshal(data)
	s.es.SendEventMessage(string(b), event, "")
}
