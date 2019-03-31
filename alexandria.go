package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"

	"hawx.me/code/alexandria/data"
	"hawx.me/code/alexandria/handler"
	"hawx.me/code/indieauth"
	"hawx.me/code/indieauth/sessions"
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

	auth, err := indieauth.Authentication(*url, *url+"callback")
	if err != nil {
		log.Fatal(err)
	}

	session, err := sessions.New(*me, *secret, auth)
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

	route.Handle("/", mux.Method{"GET": session.Choose(handler.List(true, templates), handler.List(false, templates))})
	route.Handle("/books", session.Shield(handler.AllBooks(db, es)))
	route.Handle("/books/:id", session.Shield(handler.Books(db, es)))
	route.Handle("/editions/:id", session.Shield(handler.Editions(db, *booksPath)))
	route.Handle("/upload", session.Shield(handler.Upload(db, es, *booksPath)))

	route.Handle("/sign-in", session.SignIn())
	route.Handle("/callback", session.Callback())
	route.Handle("/sign-out", session.SignOut())

	route.Handle("/events", es)
	route.Handle("/public/*path", http.StripPrefix("/public", http.FileServer(http.Dir(*webPath+"/static"))))

	serve.Serve(*port, *socket, route.Default)
}
