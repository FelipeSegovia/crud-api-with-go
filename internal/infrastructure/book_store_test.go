package infrastructure

import (
	"002-crud-api/internal/model"
	"database/sql"
	"errors"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
)

// newTestStore usa una base SQLite temporal, para probar las consultas reales.
func newTestStore(t *testing.T) Store {
	t.Helper()
	db, err := OpenSQLite(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return NewStore(db)
}

func TestUpdateBookKeepsUnsentField(t *testing.T) {
	s := newTestStore(t)
	book, err := s.CreateBook(&model.Book{Title: "Go", Author: "Pike"})
	if err != nil {
		t.Fatal(err)
	}

	title := "Go 2"
	got, err := s.UpdateBook(book.ID, &title, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "Go 2" || got.Author != "Pike" {
		t.Errorf("got %+v, want Title \"Go 2\" and Author \"Pike\"", got)
	}
}

func TestUpdateBookMissingIDIsNoRows(t *testing.T) {
	s := newTestStore(t)
	title := "X"
	if _, err := s.UpdateBook(99, &title, nil); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("got %v, want sql.ErrNoRows", err)
	}
}

// Mismo título y distinto autor, o mismo autor y distinto título, sí se permiten; ambos iguales no.
func TestCreateBookRejectsOnlySamePair(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.CreateBook(&model.Book{Title: "Go", Author: "Pike"}); err != nil {
		t.Fatal(err)
	}

	if _, err := s.CreateBook(&model.Book{Title: "Go", Author: "Pike"}); !errors.Is(err, ErrDuplicateBook) {
		t.Errorf("mismo título y autor: got %v, want ErrDuplicateBook", err)
	}
	if _, err := s.CreateBook(&model.Book{Title: "Go", Author: "Donovan"}); err != nil {
		t.Errorf("mismo título, distinto autor: %v", err)
	}
	if _, err := s.CreateBook(&model.Book{Title: "Otro", Author: "Pike"}); err != nil {
		t.Errorf("mismo autor, distinto título: %v", err)
	}
}

func TestUpdateBookRejectsDuplicatePair(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.CreateBook(&model.Book{Title: "Go", Author: "Pike"}); err != nil {
		t.Fatal(err)
	}
	other, err := s.CreateBook(&model.Book{Title: "Go", Author: "Donovan"})
	if err != nil {
		t.Fatal(err)
	}

	// Cambiar el autor del segundo libro a Pike lo haría igual al primero.
	author := "Pike"
	if _, err := s.UpdateBook(other.ID, nil, &author); !errors.Is(err, ErrDuplicateBook) {
		t.Errorf("got %v, want ErrDuplicateBook", err)
	}

	// Dejar el libro con sus propios valores no es un duplicado.
	title := "Go"
	if _, err := s.UpdateBook(other.ID, &title, nil); err != nil {
		t.Errorf("mismos valores: %v", err)
	}
}

// Dos actualizaciones de campos distintos al mismo libro no deben pisarse.
func TestUpdateBookConcurrentFieldsDontOverwrite(t *testing.T) {
	s := newTestStore(t)
	book, err := s.CreateBook(&model.Book{Title: "T0", Author: "A0"})
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 1; i <= 100; i++ {
			title := "T" + strconv.Itoa(i)
			if _, err := s.UpdateBook(book.ID, &title, nil); err != nil {
				t.Error(err)
			}
		}
	}()
	go func() {
		defer wg.Done()
		for i := 1; i <= 100; i++ {
			author := "A" + strconv.Itoa(i)
			if _, err := s.UpdateBook(book.ID, nil, &author); err != nil {
				t.Error(err)
			}
		}
	}()
	wg.Wait()

	got, err := s.GetBookByID(book.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "T100" || got.Author != "A100" {
		t.Errorf("got %+v, want T100 and A100: una actualización pisó a la otra", got)
	}
}
