package events

import (
	"github.com/hawx/alexandria/data/models"
	"github.com/hawx/alexandria/web/response"

	"github.com/antage/eventsource"

	"encoding/json"
	"net/http"
)

type Source struct {
	es eventsource.EventSource
}

func New() *Source {
	return &Source{eventsource.New(nil, nil)}
}

func (s *Source) Close() {
	s.es.Close()
}

func (s *Source) Add(book models.Book) {
	s.send("add", response.ConvertBook(book))
}

func (s *Source) Update(book models.Book) {
	s.send("update", response.ConvertBook(book))
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
