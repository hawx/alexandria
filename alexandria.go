package main

import (
	"github.com/hawx/alexandria/assets"
	"github.com/hawx/alexandria/database"
	"github.com/hawx/alexandria/epub"
	"github.com/hawx/alexandria/mobi"
	"github.com/hawx/alexandria/models"
	"github.com/hawx/alexandria/views"

	"code.google.com/p/go-uuid/uuid"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/hawx/phemera/cookie"
	"github.com/hawx/phemera/persona"
	"github.com/hoisie/mustache"
	"github.com/stvp/go-toml-config"

	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"time"
)

const (
	EPUB = "application/epub+zip"
	MOBI = "application/x-mobipocket-ebook"
)

func extension(contentType string) string {
	switch contentType {
	case EPUB: return ".epub"
	case MOBI: return ".mobi"
	}
	return ""
}

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

func Upload(fileheader *multipart.FileHeader, db database.Db) error {
	file, err := fileheader.Open()
	defer file.Close()

	if err != nil {
		return err
	}

	contentType := fileheader.Header["Content-Type"][0]
	newBook := models.Book{Id: uuid.New(), Added: time.Now()}
	editionId := uuid.New()
	dstPath := path.Join(*bookPath, editionId + extension(contentType))

	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}

	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return err
	}

	opened, err := fileheader.Open()
	if err != nil {
		return err
	}

	switch contentType {
	case EPUB:
		book, _ := epub.Open(opened)
		meta, _ := book.Metadata()

		newBook.Title = meta.Title[0]
		newBook.Author = meta.Creator[0].Value
		newBook.Editions = models.Editions{
			{
			  Id:          editionId,
			  Path:        dstPath,
			  ContentType: EPUB,
			  Extension:   extension(EPUB),
			},
		}

	case MOBI:
		book, _ := mobi.Open(opened)
		meta, _ := book.Metadata()

		newBook.Title = meta.Title
		newBook.Author = meta.Creator
		newBook.Editions = models.Editions{
			{
			  Id:          editionId,
			  Path:        dstPath,
			  ContentType: MOBI,
			  Extension:   extension(MOBI),
			},
		}

	default:
		return errors.New("Format not supported: " + contentType)
	}

	log.Println("Uploaded")
	db.Save(newBook)

	go func(contentType string, book models.Book, db database.Db) {
		if contentType != EPUB {
			log.Println("Converting to EPUB")
			editionId := uuid.New()
			from := book.Editions[0].Path
			to := path.Join(*bookPath, editionId + extension(EPUB))

			cmd := exec.Command("ebook-convert", from, to)
			if err := cmd.Run(); err != nil {
				log.Println(err)
				return
			}

			book.Editions = append(book.Editions, &models.Edition{
			  Id: editionId,
			  Path: to,
			  ContentType: EPUB,
			  Extension: extension(EPUB),
			})
		}

		if contentType != MOBI {
			log.Println("Converting to MOBI")
			editionId := uuid.New()
			from := book.Editions[0].Path
			to := path.Join(*bookPath, editionId + extension(MOBI))

			cmd := exec.Command("ebook-convert", from, to)
			if err := cmd.Run(); err != nil {
				log.Println(err)
				return
			}

			book.Editions = append(book.Editions, &models.Edition{
			  Id: editionId,
			  Path: to,
			  ContentType: MOBI,
			  Extension: extension(MOBI),
			})
		}

		log.Println("Converted")
		db.Save(book)
	}(contentType, newBook, db)

	return nil
}

/*
   def upload(temp_path, content_type)
    temp_type = Alexandria::Edition.get_type(content_type)

    initial = Library.create_edition(temp_path, temp_type)
    book = Alexandria::Book.new([initial], initial.meta)
    Library.add(book)

    notify :add, Representer::Book.new(book)

    Thread.new {
      missing_types = Alexandria::Edition::TYPES - [temp_type]
      missing_types.each do |type|
        begin
          other_path = "#{initial.path}#{type.extname}"
          convert initial.path, other_path
          book << Library.create_edition(other_path, type, false)
        rescue => err
          warn "Failed to convert #{initial.path} to #{type}"
          warn err
        end
      end

      Library.update(book)
      notify :update, Representer::Book.new(book)
    }
  end
*/

type Href struct {
	Href string `json:"href"`
}

type Editions []*Edition

type Edition struct {
	Id    string          `json:"id"`
	Name  string          `json:"name"`
	Links map[string]Href `json:"links"`
}

type Books []*Book

type Book struct {
	Id       string          `json:"id"`
	Title    string          `json:"title"`
	Author   string          `json:"author"`
	Added    string          `json:"added"`
	Editions Editions        `json:"editions"`
	Links    map[string]Href `json:"links"`
}

type Root struct {
	Books Books `json:"books"`
}

func Decorate(modelBooks models.Books) Root {
	books := make([]*Book, len(modelBooks))

	for i, book := range modelBooks {
		editions := make([]*Edition, len(book.Editions))

		for j, edition := range book.Editions {
			editions[j] = &Edition{
				Id:   edition.Id,
				Name: edition.Extension[1:],
				Links: map[string]Href{
					"self": {"/editions/" + edition.Id},
				},
			}
		}

		books[i] = &Book{
			Id:       book.Id,
			Title:    book.Title,
			Author:   book.Author,
			Added:    book.Added.Format("2006-01-02"),
			Editions: editions,
			Links: map[string]Href{
				"self": {"/books/" + book.Id},
			},
		}
	}

	return Root{books}
}

func main() {
	flag.Parse()

	if err := config.Parse(*settingsPath); err != nil {
		log.Fatal("toml:", err)
	}

	store = cookie.NewStore(*cookieSecret)

	db := database.Open(*dbPath)
	defer db.Close()

	r := mux.NewRouter()

	r.Path("/").Methods("GET").Handler(Render(views.List, db))

	r.Path("/books").Methods("GET").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")

		if LoggedIn(r) {
			json.NewEncoder(w).Encode(Decorate(db.Get()))
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
		w.Header().Set("Content-Disposition", `attachment; filename=`+ book.Slug(edition))
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
			if err := Upload(file, db); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	})

	r.Path("/sign-in").Methods("POST").Handler(persona.SignIn(store, *audience))
	r.Path("/sign-out").Methods("GET").Handler(persona.SignOut(store))

	http.Handle("/", r)
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
