package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/BurntSushi/toml"
	"hawx.me/code/alexandria/data"
	"hawx.me/code/alexandria/web/assets"
	"hawx.me/code/alexandria/web/events"
	"hawx.me/code/alexandria/web/filters"
	"hawx.me/code/alexandria/web/handlers"
	"hawx.me/code/indieauth"
	"hawx.me/code/indieauth/sessions"
	"hawx.me/code/mux"
	"hawx.me/code/route"
	"hawx.me/code/serve"
)

var (
	settingsPath = flag.String("settings", "./settings.toml", "")
	port         = flag.String("port", "8080", "")
	socket       = flag.String("socket", "", "")
)

func main() {
	flag.Parse()

	var conf struct {
		Secret    string
		DbPath    string `toml:"database"`
		BooksPath string `toml:"library"`
		URL       string
		Me        string
	}
	if _, err := toml.DecodeFile(*settingsPath, &conf); err != nil {
		log.Fatal("toml:", err)
	}

	auth, err := indieauth.Authentication(conf.URL, conf.URL+"callback")
	if err != nil {
		log.Fatal(err)
	}

	session, err := sessions.New(conf.Me, conf.Secret, auth)
	if err != nil {
		log.Fatal(err)
	}

	db := data.Open(conf.DbPath)
	defer db.Close()

	es := events.New()
	defer es.Close()

	route.Handle("/", mux.Method{"GET": session.Choose(handlers.List(true), handlers.List(false))})
	route.Handle("/books", session.Shield(handlers.AllBooks(db, es)))
	route.Handle("/books/:id", session.Shield(handlers.Books(db, es)))
	route.Handle("/editions/:id", session.Shield(handlers.Editions(db, conf.BooksPath)))
	route.Handle("/upload", session.Shield(handlers.Upload(db, es, conf.BooksPath)))

	route.Handle("/sign-in", session.SignIn())
	route.Handle("/callback", session.Callback())
	route.Handle("/sign-out", session.SignOut())

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
