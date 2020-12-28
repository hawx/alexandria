package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"

	"hawx.me/code/alexandria/data"
	"hawx.me/code/alexandria/handler"
	"hawx.me/code/indieauth/v2"
	"hawx.me/code/mux"
	"hawx.me/code/route"
	"hawx.me/code/serve"
)

func main() {
	var (
		secret    = flag.String("secret", "plschange", "")
		dbPath    = flag.String("db", "./db", "")
		booksPath = flag.String("books", "./books", "")
		webPath   = flag.String("web", "web", "")
		url       = flag.String("url", "http://localhost:8080/", "")
		me        = flag.String("me", "", "")
		port      = flag.String("port", "8080", "")
		socket    = flag.String("socket", "", "")
	)
	flag.Parse()

	session, err := indieauth.NewSessions(*secret, &indieauth.Config{
		ClientID:    *url,
		RedirectURL: *url + "callback",
	})
	if err != nil {
		log.Fatal(err)
	}

	templates, err := template.ParseGlob(*webPath + "/template/*")
	if err != nil {
		log.Fatal("could not load templates:", err)
	}

	db, err := data.Open(*dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	es := handler.Events()
	defer es.Close()

	signedIn := func(a http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if response, ok := session.SignedIn(r); ok && response.Me == *me {
				a.ServeHTTP(w, r)
			} else {
				http.NotFound(w, r)
			}
		}
	}

	route.Handle("/", mux.Method{"GET": handler.List(*me, templates, session)})
	route.Handle("/books", signedIn(handler.AllBooks(db, es)))
	route.Handle("/books/:id", signedIn(handler.Books(db, es)))
	route.Handle("/editions/:id", signedIn(handler.Editions(db, *booksPath)))
	route.Handle("/upload", signedIn(handler.Upload(db, es, *booksPath)))

	route.HandleFunc("/sign-in", func(w http.ResponseWriter, r *http.Request) {
		if err := session.RedirectToSignIn(w, r, *me); err != nil {
			log.Println(err)
		}
	})
	route.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if err := session.Verify(w, r); err != nil {
			log.Println(err)
		}
		http.Redirect(w, r, "/", http.StatusFound)
	})
	route.HandleFunc("/sign-out", func(w http.ResponseWriter, r *http.Request) {
		if err := session.SignOut(w, r); err != nil {
			log.Println(err)
		}
		http.Redirect(w, r, "/", http.StatusFound)
	})

	route.Handle("/events", es)
	route.Handle("/public/*path", http.StripPrefix("/public", http.FileServer(http.Dir(*webPath+"/static"))))

	serve.Serve(*port, *socket, route.Default)
}
