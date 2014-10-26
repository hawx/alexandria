package response

import "github.com/hawx/alexandria/models"

type Href struct {
	Href string `json:"href"`
}

type Editions []Edition

type Edition struct {
	Id    string          `json:"id"`
	Name  string          `json:"name"`
	Links map[string]Href `json:"links"`
}

type Books []Book

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

func ConvertEdition(edition models.Edition) Edition {
	return Edition{
		Id:   edition.Id,
		Name: edition.Extension[1:],
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
	books := make([]Book, len(modelBooks))

	for i, book := range modelBooks {
		books[i] = ConvertBook(*book)
	}

	return Root{books}
}
