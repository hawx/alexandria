package main

import (
	"github.com/hawx/alexandria/data"
	"github.com/hawx/alexandria/web/assets"
	"github.com/hawx/alexandria/web/events"
	"github.com/hawx/alexandria/web/filters"
	"github.com/hawx/alexandria/web/handlers"
	"github.com/hawx/alexandria/web/persona"

	"github.com/BurntSushi/toml"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"

	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
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

	editionsHandler := handlers.Editions(db)
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

	if *socket == "" {
		go func() {
			log.Print("listening on :" + *port)
			log.Fatal(http.ListenAndServe(":"+*port, context.ClearHandler(filters.Log(http.DefaultServeMux))))
		}()

	} else {
		l, err := net.Listen("unix", *socket)
		if err != nil {
			log.Fatal(err)
		}

		defer l.Close()

		go func() {
			log.Println("listening on", *socket)
			log.Fatal(http.Serve(l, context.ClearHandler(filters.Log(http.DefaultServeMux))))
		}()
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	s := <-c
	log.Printf("caught %s: shutting down", s)
}
