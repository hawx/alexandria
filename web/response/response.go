package response

import (
	"sort"

	"hawx.me/code/alexandria/data/models"
)

type Href struct {
	Href string `json:"href"`
}

type Edition struct {
	Id    string          `json:"id"`
	Name  string          `json:"name"`
	Links map[string]Href `json:"links"`
}

type Editions []Edition

type Book struct {
	Id       string          `json:"id"`
	Title    string          `json:"title"`
	Author   string          `json:"author"`
	Added    string          `json:"added"`
	Editions Editions        `json:"editions"`
	Links    map[string]Href `json:"links"`
}

type Books []Book

func (books Books) Len() int {
	return len(books)
}

func (books Books) Swap(i, j int) {
	books[i], books[j] = books[j], books[i]
}

func (books Books) Less(i, j int) bool {
	return books[i].Added < books[j].Added
}

type Root struct {
	Books Books `json:"books"`
}

func ConvertEdition(edition models.Edition) Edition {
	return Edition{
		Id:   edition.Id,
		Name: edition.Extension()[1:],
		Links: map[string]Href{
			"self": {"/editions/" + edition.Id},
		},
	}
}

func ConvertBook(book models.Book) Book {
	editions := make([]Edition, len(book.Editions))

	for j, edition := range book.Editions {
		editions[j] = ConvertEdition(*edition)
	}

	return Book{
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

func ConvertBooks(modelBooks models.Books) Root {
	books := Books{}

	for _, book := range modelBooks {
		books = append(books, ConvertBook(*book))
	}

	sort.Sort(sort.Reverse(books))
	return Root{books}
}
