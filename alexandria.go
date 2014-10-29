package main

import (
	"github.com/hawx/alexandria/data"
	"github.com/hawx/alexandria/web/assets"
	"github.com/hawx/alexandria/web/events"
	"github.com/hawx/alexandria/web/handlers"
	"github.com/hawx/alexandria/web/views"
	"github.com/hawx/alexandria/web/filters"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/hawx/phemera/cookie"
	"github.com/hawx/wwwhat/persona"
	"github.com/hoisie/mustache"
	"github.com/stvp/go-toml-config"

	"flag"
	"fmt"
	"log"
	"net/http"
)

var store cookie.Store

var (
	settingsPath = flag.String("settings", "./settings.toml", "")
	port         = flag.String("port", "8080", "")

	user         = config.String("user", "someone@example.com")
	cookieSecret = config.String("secret", "some-secret-plz-change")
	audience     = config.String("audience", "localhost")
	dbPath       = config.String("db", "./alexandria-db")
	bookPath     = config.String("books", "./alexandria-books")
)

func loggedIn(r *http.Request) bool {
	return store.Get(r) == *user
}

func Render(template *mustache.Template, db data.Db) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := template.Render(struct{ LoggedIn bool }{loggedIn(r)})
		w.Header().Add("Content-Type", "text/html")
		fmt.Fprintf(w, body)
	})
}

func main() {
	flag.Parse()

	if err := config.Parse(*settingsPath); err != nil {
		log.Fatal("toml:", err)
	}

	store = cookie.NewStore(*cookieSecret)
	protect := persona.Protector(store, []string{*user})

	db := data.Open(*dbPath)
	defer db.Close()

	es := events.New()
	defer es.Close()

	r := mux.NewRouter()

	r.Path("/").Methods("GET").Handler(Render(views.List, db))

	booksHandler := handlers.Books(db, es)
	editionsHandler := handlers.Editions(db)
	uploadHandler := handlers.Upload(db, es, *bookPath)

	r.Path("/books").Methods("GET").Handler(protect(booksHandler.GetAll))
	r.Path("/books/{id}").Methods("GET").Handler(protect(booksHandler.Get))
	r.Path("/books/{id}").Methods("PATCH").Handler(protect(booksHandler.Update))
	r.Path("/books/{id}").Methods("DELETE").Handler(protect(booksHandler.Delete))

	r.Path("/editions/{id}").Methods("GET").Handler(protect(editionsHandler.Get))

	r.Path("/upload").Methods("POST").Handler(protect(uploadHandler.Upload))

	r.Path("/sign-in").Methods("POST").Handler(persona.SignIn(store, *audience))
	r.Path("/sign-out").Methods("GET").Handler(persona.SignOut(store))

	http.Handle("/", r)
	http.Handle("/events", es)
	http.Handle("/assets/", http.StripPrefix("/assets/", assets.Server(map[string]string{
		"main.js":        assets.MainJs,
		"mustache.js":    assets.MustacheJs,
		"tablesorter.js": assets.TablesorterJs,
		"tablefilter.js": assets.TablefilterJs,
		"styles.css":     assets.StylesCss,
	})))

	log.Print("Running on :" + *port)
	log.Fatal(http.ListenAndServe(":"+*port, context.ClearHandler(filters.Log(http.DefaultServeMux))))
}
