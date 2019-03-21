package handler

import (
	"io"
	"log"
	"net/http"
)

func List(loggedIn bool, templates interface {
	ExecuteTemplate(w io.Writer, name string, data interface{}) error
}) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		if err := templates.ExecuteTemplate(w, "list.gotmpl", !loggedIn); err != nil {
			log.Println(err)
		}
	})
}
