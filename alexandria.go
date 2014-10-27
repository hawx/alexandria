package main

import (
	"github.com/hawx/alexandria/actions"
	"github.com/hawx/alexandria/assets"
	"github.com/hawx/alexandria/database"
	"github.com/hawx/alexandria/events"
	"github.com/hawx/alexandria/models"
	"github.com/hawx/alexandria/response"
	"github.com/hawx/alexandria/views"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/hawx/phemera/cookie"
	"github.com/hawx/phemera/persona"
	"github.com/hoisie/mustache"
	"github.com/stvp/go-toml-config"

	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
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

type Ctx struct{ LoggedIn bool }

func LoggedIn(r *http.Request) bool {
	return store.Get(r) == *user
}

func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func Render(template *mustache.Template, db database.Db) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := template.Render(Ctx{LoggedIn: LoggedIn(r)})
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

	db := database.Open(*dbPath)
	defer db.Close()

	es := events.New()
	defer es.Close()

	r := mux.NewRouter()

	r.Path("/").Methods("GET").Handler(Render(views.List, db))

	r.Path("/books").Methods("GET").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")

		if LoggedIn(r) {
			json.NewEncoder(w).Encode(response.ConvertBooks(db.Get()))
		} else {
			fmt.Fprint(w, "{\"books\": []}")
		}
	})

	r.Path("/books/{id}").Methods("GET").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !LoggedIn(r) {
			w.WriteHeader(403)
			return
		}

		id := mux.Vars(r)["id"]
		book, ok := db.Find(id)

		if !ok {
			w.WriteHeader(404)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(book)
	})

	r.Path("/books/{id}").Methods("PATCH").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !LoggedIn(r) {
			w.WriteHeader(403)
			return
		}

		id := mux.Vars(r)["id"]
		book, ok := db.Find(id)

		if !ok {
			w.WriteHeader(404)
			return
		}

		var req models.Book
		json.NewDecoder(r.Body).Decode(&req)

		if req.Title != "" {
			book.Title = req.Title
		}
		if req.Author != "" {
			book.Author = req.Author
		}

		db.Save(book)
		es.Update(book)

		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(book)
	})

	r.Path("/books/{id}").Methods("DELETE").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !LoggedIn(r) {
			w.WriteHeader(403)
			return
		}

		id := mux.Vars(r)["id"]
		book, ok := db.Find(id)

		if !ok {
			w.WriteHeader(404)
			return
		}

		db.Remove(book)
		es.Delete(book)

		w.WriteHeader(204)
	})

	r.Path("/editions/{id}").Methods("GET").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !LoggedIn(r) {
			w.WriteHeader(403)
			return
		}

		id := mux.Vars(r)["id"]
		edition, book, ok := db.FindEdition(id)

		if !ok {
			w.WriteHeader(404)
			return
		}

		file, err := os.Open(edition.Path)
		if err != nil {
			w.WriteHeader(500)
		}

		stat, err := file.Stat()
		if err != nil {
			w.WriteHeader(500)
		}

		w.Header().Set("Content-Type", edition.ContentType)
		w.Header().Set("Content-Disposition", `attachment; filename=`+book.Slug(edition))
		w.Header().Set("Content-Length", strconv.Itoa(int(stat.Size())))
		io.Copy(w, file)
	})

	r.Path("/upload").Methods("POST").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !LoggedIn(r) {
			w.WriteHeader(403)
			return
		}

		err := r.ParseMultipartForm(100000)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		files := r.MultipartForm.File["file"]
		for _, file := range files {
			if err := actions.Upload(*bookPath, file, db, es); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	})

	r.Path("/sign-in").Methods("POST").Handler(persona.SignIn(store, *audience))
	r.Path("/sign-out").Methods("GET").Handler(persona.SignOut(store))

	http.Handle("/", r)
	http.Handle("/events", es)
	http.Handle("/assets/", http.StripPrefix("/assets/", assets.Server(map[string]string{
		"main.js":            assets.Main,
		"jquery.mustache.js": assets.Mustache,
		"plugins.js":         assets.Plugins,
		"tablefilter.js":     assets.Tablefilter,
		"styles.css":         assets.Styles,
	})))

	log.Print("Running on :" + *port)
	log.Fatal(http.ListenAndServe(":"+*port, context.ClearHandler(Log(http.DefaultServeMux))))
}
