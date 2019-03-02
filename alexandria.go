package main

import (
	"crypto/rand"
	"encoding/base64"
	"flag"
	"io"
	"log"
	"net/http"

	"github.com/BurntSushi/toml"
	"github.com/gorilla/sessions"
	"hawx.me/code/alexandria/data"
	"hawx.me/code/alexandria/web/assets"
	"hawx.me/code/alexandria/web/events"
	"hawx.me/code/alexandria/web/filters"
	"hawx.me/code/alexandria/web/handlers"
	"hawx.me/code/indieauth"
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

	if conf.Me == "" {
		log.Fatal("me must be set to sign-in")
	}

	auth, err := indieauth.Authentication(conf.URL, conf.URL+"/callback")
	if err != nil {
		log.Fatal(err)
	}

	endpoints, err := indieauth.FindEndpoints(conf.Me)
	if err != nil {
		log.Fatal(err)
	}

	db := data.Open(conf.DbPath)
	defer db.Close()

	es := events.New()
	defer es.Close()

	store := NewStore(conf.Secret)

	protect := func(good, bad http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if addr := store.Get(r); addr == conf.Me {
				good.ServeHTTP(w, r)
			} else {
				bad.ServeHTTP(w, r)
			}
		})
	}

	shield := func(h http.Handler) http.Handler {
		return protect(h, http.NotFoundHandler())
	}

	route.Handle("/", mux.Method{"GET": protect(handlers.List(true), handlers.List(false))})
	route.Handle("/books", shield(handlers.AllBooks(db, es)))
	route.Handle("/books/:id", shield(handlers.Books(db, es)))
	route.Handle("/editions/:id", shield(handlers.Editions(db, conf.BooksPath)))
	route.Handle("/upload", shield(handlers.Upload(db, es, conf.BooksPath)))

	route.HandleFunc("/sign-in", func(w http.ResponseWriter, r *http.Request) {
		state, err := store.SetState(w, r)
		if err != nil {
			http.Error(w, "could not start auth", http.StatusInternalServerError)
			return
		}

		redirectURL := auth.RedirectURL(endpoints, conf.Me, state)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, redirectURL, http.StatusFound)
	})

	route.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		state := store.GetState(r)

		if r.FormValue("state") != state {
			http.Error(w, "state is bad", http.StatusBadRequest)
			return
		}

		me, err := auth.Exchange(endpoints, r.FormValue("code"))
		if err != nil || me != conf.Me {
			http.Error(w, "nope", http.StatusForbidden)
			return
		}

		store.Set(w, r, me)
		http.Redirect(w, r, "/", http.StatusFound)
	})

	route.HandleFunc("/sign-out", func(w http.ResponseWriter, r *http.Request) {
		store.Set(w, r, "")
		http.Redirect(w, r, "/", http.StatusFound)
	})

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

type meStore struct {
	store sessions.Store
}

func NewStore(secret string) *meStore {
	return &meStore{sessions.NewCookieStore([]byte(secret))}
}

func (s meStore) Get(r *http.Request) string {
	session, _ := s.store.Get(r, "session")

	if v, ok := session.Values["me"].(string); ok {
		return v
	}

	return ""
}

func (s meStore) Set(w http.ResponseWriter, r *http.Request, me string) {
	session, _ := s.store.Get(r, "session")
	session.Values["me"] = me
	session.Save(r, w)
}

func (s meStore) SetState(w http.ResponseWriter, r *http.Request) (string, error) {
	bytes := make([]byte, 32)

	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		return "", err
	}

	state := base64.StdEncoding.EncodeToString(bytes)

	session, _ := s.store.Get(r, "session")
	session.Values["state"] = state
	return state, session.Save(r, w)
}

func (s meStore) GetState(r *http.Request) string {
	session, _ := s.store.Get(r, "session")

	if v, ok := session.Values["state"].(string); ok {
		return v
	}

	return ""
}
