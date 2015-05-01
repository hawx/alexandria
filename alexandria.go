package main

import (
	"hawx.me/code/alexandria/data"
	"hawx.me/code/alexandria/web/assets"
	"hawx.me/code/alexandria/web/events"
	"hawx.me/code/alexandria/web/filters"
	"hawx.me/code/alexandria/web/handlers"

	"github.com/BurntSushi/toml"
	"hawx.me/code/mux"
	"hawx.me/code/persona"
	"hawx.me/code/route"
	"hawx.me/code/serve"

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

	route.Handle("/", mux.Method{"GET": persona.Switch(handlers.List(true), handlers.List(false))})
	route.Handle("/books", persona.Protect(handlers.AllBooks(db, es)))
	route.Handle("/books/:id", persona.Protect(handlers.Books(db, es)))
	route.Handle("/editions/:id", persona.Protect(handlers.Editions(db, conf.BooksPath)))
	route.Handle("/upload", persona.Protect(handlers.Upload(db, es, conf.BooksPath)))

	route.Handle("/sign-in", mux.Method{"POST": persona.SignIn})
	route.Handle("/sign-out", mux.Method{"GET": persona.SignOut})
	route.Handle("/events", es)
	route.Handle("/assets/*filepath", http.StripPrefix("/assets/", assets.Server(map[string]string{
		"main.js":        assets.MainJs,
		"mustache.js":    assets.MustacheJs,
		"tablesorter.js": assets.TablesorterJs,
		"tablefilter.js": assets.TablefilterJs,
		"styles.css":     assets.StylesCss,
	})))

	serve.Serve(*port, *socket, filters.Log(route.Default))
}
