package main

import (
	"github.com/hawx/alexandria/data"
	"github.com/hawx/alexandria/web/assets"
	"github.com/hawx/alexandria/web/events"
	"github.com/hawx/alexandria/web/filters"
	"github.com/hawx/alexandria/web/handlers"

	"github.com/BurntSushi/toml"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/hawx/persona"
	"github.com/hawx/serve"

	"flag"
	"log"
	"net/http"
)

var (
	settingsPath = flag.String("settings", "./settings.toml", "")
	port         = flag.String("port", "8080", "")
	socket       = flag.String("socket", "", "")
)

func main() {
	flag.Parse()

	var conf struct {
		Users     []string
		Secret    string
		Audience  string
		DbPath    string `toml:"database"`
		BooksPath string `toml:"library"`
	}
	if _, err := toml.DecodeFile(*settingsPath, &conf); err != nil {
		log.Fatal("toml:", err)
	}

	store := persona.NewStore(conf.Secret)
	persona := persona.New(store, conf.Audience, conf.Users)

	db := data.Open(conf.DbPath)
	defer db.Close()

	es := events.New()
	defer es.Close()

	r := mux.NewRouter()

	r.Path("/").Methods("GET").Handler(persona.Switch(handlers.List(true), handlers.List(false)))

	booksHandler := handlers.Books(db, es)
	r.Path("/books").Methods("GET").Handler(persona.Protect(booksHandler.GetAll))
	r.Path("/books/{id}").Methods("GET").Handler(persona.Protect(booksHandler.Get))
	r.Path("/books/{id}").Methods("PATCH").Handler(persona.Protect(booksHandler.Update))
	r.Path("/books/{id}").Methods("DELETE").Handler(persona.Protect(booksHandler.Delete))

	editionsHandler := handlers.Editions(db, conf.BooksPath)
	r.Path("/editions/{id}").Methods("GET").Handler(persona.Protect(editionsHandler.Get))

	uploadHandler := handlers.Upload(db, es, conf.BooksPath)
	r.Path("/upload").Methods("POST").Handler(persona.Protect(uploadHandler.Upload))

	r.Path("/sign-in").Methods("POST").Handler(persona.SignIn)
	r.Path("/sign-out").Methods("GET").Handler(persona.SignOut)

	http.Handle("/", r)
	http.Handle("/events", es)
	http.Handle("/assets/", http.StripPrefix("/assets/", assets.Server(map[string]string{
		"main.js":        assets.MainJs,
		"mustache.js":    assets.MustacheJs,
		"tablesorter.js": assets.TablesorterJs,
		"tablefilter.js": assets.TablefilterJs,
		"styles.css":     assets.StylesCss,
	})))

	serve.Serve(*port, *socket, context.ClearHandler(filters.Log(http.DefaultServeMux)))
}
