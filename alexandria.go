package main

import (
	"hawx.me/code/alexandria/data"
	"hawx.me/code/alexandria/web/events"
	"hawx.me/code/alexandria/web/filters"
	"hawx.me/code/alexandria/web/handlers"

	"github.com/BurntSushi/toml"
	"hawx.me/code/route"
	"hawx.me/code/serve"
	"hawx.me/code/uberich"

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
		Secret    string
		DbPath    string `toml:"database"`
		BooksPath string `toml:"library"`
		Uberich   struct {
			AppName    string
			AppURL     string
			UberichURL string
			Secret     string
		}
	}
	if _, err := toml.DecodeFile(*settingsPath, &conf); err != nil {
		log.Fatal("toml:", err)
	}

	store := uberich.NewStore(conf.Secret)
	uberich := uberich.NewClient(conf.Uberich.AppName, conf.Uberich.AppURL, conf.Uberich.UberichURL, conf.Uberich.Secret, store)

	db := data.Open(conf.DbPath)
	defer db.Close()

	es := events.New()
	defer es.Close()

	shield := func(h http.Handler) http.Handler {
		return uberich.Protect(h, http.NotFoundHandler())
	}

	// route.Handle("/", mux.Method{"GET": uberich.Protect(handlers.List(true), handlers.List(false))})
	route.Handle("/books", shield(handlers.AllBooks(db, es)))
	route.Handle("/books/:id", shield(handlers.Books(db, es)))
	route.Handle("/editions/:id", shield(handlers.Editions(db, conf.BooksPath)))
	route.Handle("/upload", shield(handlers.Upload(db, es, conf.BooksPath)))

	route.Handle("/sign-in", uberich.SignIn("/"))
	route.Handle("/sign-out", uberich.SignOut("/"))
	route.Handle("/events", es)
	// route.Handle("/assets/*filepath", http.StripPrefix("/assets/", assets.Server(map[string]string{
	// 	"main.js":        assets.MainJs,
	// 	"mustache.js":    assets.MustacheJs,
	// 	"tablesorter.js": assets.TablesorterJs,
	// 	"tablefilter.js": assets.TablefilterJs,
	// 	"styles.css":     assets.StylesCss,
	// })))

	route.Handle("/", http.FileServer(http.Dir("/home/hawx/dev/go/src/hawx.me/code/alexandria/app/dist")))

	serve.Serve(*port, *socket, filters.Log(route.Default))
}
