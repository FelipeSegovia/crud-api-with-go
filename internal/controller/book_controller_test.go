package controller

import (
	"002-crud-api/internal/infrastructure"
	"002-crud-api/internal/model"
	"002-crud-api/internal/service"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// fakeStore implementa infrastructure.Store en memoria para probar las rutas sin base de datos.
type fakeStore struct {
	books  map[int]*model.Book
	nextID int
}

func (f *fakeStore) GetAllBooks() ([]*model.Book, error) {
	books := []*model.Book{}
	for _, b := range f.books {
		books = append(books, b)
	}
	return books, nil
}

func (f *fakeStore) GetBookByID(id int) (*model.Book, error) {
	b, ok := f.books[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	// Devuelve una copia, como el store real: modificarla no cambia lo guardado.
	cp := *b
	return &cp, nil
}

// duplicate indica si otro libro distinto de id ya tiene ese título y autor.
func (f *fakeStore) duplicate(id int, title, author string) bool {
	for _, b := range f.books {
		if b.ID != id && b.Title == title && b.Author == author {
			return true
		}
	}
	return false
}

func (f *fakeStore) CreateBook(book *model.Book) (*model.Book, error) {
	if f.duplicate(0, book.Title, book.Author) {
		return nil, infrastructure.ErrDuplicateBook
	}
	f.nextID++
	book.ID = f.nextID
	f.books[book.ID] = book
	return book, nil
}

func (f *fakeStore) UpdateBook(id int, title, author *string) (*model.Book, error) {
	b, ok := f.books[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	newTitle, newAuthor := b.Title, b.Author
	if title != nil {
		newTitle = *title
	}
	if author != nil {
		newAuthor = *author
	}
	if f.duplicate(id, newTitle, newAuthor) {
		return nil, infrastructure.ErrDuplicateBook
	}
	b.Title, b.Author = newTitle, newAuthor
	cp := *b
	return &cp, nil
}

func (f *fakeStore) DeleteBook(id int) error {
	if _, err := f.GetBookByID(id); err != nil {
		return err
	}
	delete(f.books, id)
	return nil
}

type nopLogger struct{}

func (nopLogger) Log(string) {}

func TestBookRoutes(t *testing.T) {
	mux := http.NewServeMux()
	store := &fakeStore{books: map[int]*model.Book{}}
	NewBookController(service.NewService(store, nopLogger{})).RegisterRoutes(mux)

	// Los casos se ejecutan en orden: cada uno depende del estado que dejó el anterior.
	tests := []struct {
		method, path, body string
		want               int
	}{
		{"GET", "/books", "", http.StatusOK},
		{"POST", "/books", `{"title":"Go","author":"Pike"}`, http.StatusCreated},
		{"POST", "/books", `{"title":"Go","author":"Pike"}`, http.StatusConflict},
		{"POST", "/books", `{"title":"Go","author":"Donovan"}`, http.StatusCreated},
		{"POST", "/books", `{"title":"Otro","author":"Pike"}`, http.StatusCreated},
		{"POST", "/books", `{"title":""}`, http.StatusBadRequest},
		{"POST", "/books", `{`, http.StatusBadRequest},
		{"GET", "/books/1", "", http.StatusOK},
		{"GET", "/books/99", "", http.StatusNotFound},
		{"GET", "/books/abc", "", http.StatusBadRequest},
		{"PATCH", "/books/1", `{"title":"Go 2"}`, http.StatusOK},
		{"PATCH", "/books/1", `{"title":"Go","author":"Donovan"}`, http.StatusConflict},
		{"PATCH", "/books/99", `{"title":"X"}`, http.StatusNotFound},
		{"PATCH", "/books/1", `{"title":""}`, http.StatusBadRequest},
		{"DELETE", "/books/1", "", http.StatusNoContent},
		{"DELETE", "/books/1", "", http.StatusNotFound},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != tt.want {
			t.Errorf("%s %s: got %d, want %d", tt.method, tt.path, rec.Code, tt.want)
		}
	}
}

func TestPatchKeepsUnsentFields(t *testing.T) {
	store := &fakeStore{books: map[int]*model.Book{1: {ID: 1, Title: "Go", Author: "Pike"}}}
	mux := http.NewServeMux()
	NewBookController(service.NewService(store, nopLogger{})).RegisterRoutes(mux)

	req := httptest.NewRequest("PATCH", "/books/1", strings.NewReader(`{"title":"Go 2"}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusOK)
	}

	var got model.Book
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Title != "Go 2" || got.Author != "Pike" {
		t.Errorf("got %+v, want Title \"Go 2\" and Author \"Pike\"", got)
	}
}
