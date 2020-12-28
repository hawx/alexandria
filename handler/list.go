package handler

import (
	"io"
	"log"
	"net/http"

	"hawx.me/code/indieauth/v2"
)

type Templates interface {
	ExecuteTemplate(w io.Writer, name string, data interface{}) error
}

type Sessions interface {
	SignedIn(*http.Request) (*indieauth.Response, bool)
}

func List(me string, templates Templates, sessions Sessions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		response, ok := sessions.SignedIn(r)

		tmpl := "signin.gotmpl"
		if ok && response.Me == me {
			tmpl = "list.gotmpl"
		}

		if err := templates.ExecuteTemplate(w, tmpl, nil); err != nil {
			log.Println(err)
		}
	}
}
