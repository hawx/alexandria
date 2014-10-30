package handlers

import (
	"github.com/hawx/alexandria/web/views"

	"net/http"
)

func List(loggedIn bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		views.List.Execute(w, !loggedIn)
	})
}
