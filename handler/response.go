package handler

import (
	"sort"

	"hawx.me/code/alexandria/data/models"
)

type hrefResponse struct {
	Href string `json:"href"`
}

type editionResponse struct {
	Id    string                  `json:"id"`
	Name  string                  `json:"name"`
	Links map[string]hrefResponse `json:"links"`
}

type editionsResponse []editionResponse

type bookResponse struct {
	Id       string                  `json:"id"`
	Title    string                  `json:"title"`
	Author   string                  `json:"author"`
	Added    string                  `json:"added"`
	Editions editionsResponse        `json:"editions"`
	Links    map[string]hrefResponse `json:"links"`
}

type booksResponse []bookResponse

func (books booksResponse) Len() int {
	return len(books)
}

func (books booksResponse) Swap(i, j int) {
	books[i], books[j] = books[j], books[i]
}

func (books booksResponse) Less(i, j int) bool {
	return books[i].Added < books[j].Added
}

type rootResponse struct {
	Books booksResponse `json:"books"`
}

func convertEdition(edition models.Edition) editionResponse {
	return editionResponse{
		Id:   edition.Id,
		Name: edition.Extension()[1:],
		Links: map[string]hrefResponse{
			"self": {"/editions/" + edition.Id},
		},
	}
}

func convertBook(book models.Book) bookResponse {
	editions := make([]editionResponse, len(book.Editions))

	for j, edition := range book.Editions {
		editions[j] = convertEdition(*edition)
	}

	return bookResponse{
		Id:       book.Id,
		Title:    book.Title,
		Author:   book.Author,
		Added:    book.Added.Format("2006-01-02"),
		Editions: editions,
		Links: map[string]hrefResponse{
			"self": {"/books/" + book.Id},
		},
	}
}

func convertBooks(modelBooks models.Books) rootResponse {
	books := booksResponse{}

	for _, book := range modelBooks {
		books = append(books, convertBook(*book))
	}

	sort.Sort(sort.Reverse(books))
	return rootResponse{books}
}
