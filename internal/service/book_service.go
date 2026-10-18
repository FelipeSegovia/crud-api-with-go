package service

import (
	"002-crud-api/internal/infrastructure"
	"002-crud-api/internal/model"
	"errors"
)

var ErrTitleAndAuthorRequired = errors.New("title and author are required")

// ErrDuplicateBook indica que ya existe un libro con el mismo título y autor.
var ErrDuplicateBook = infrastructure.ErrDuplicateBook

type Logger interface {
	Log(message string)
}

type Service struct {
	store  infrastructure.Store
	logger Logger
}

func NewService(store infrastructure.Store, logger Logger) *Service {
	return &Service{store: store, logger: logger}
}

func (s *Service) GetAllBooks() ([]*model.Book, error) {
	s.logger.Log("Getting all books")
	return s.store.GetAllBooks()
}

func (s *Service) GetBookByID(id int) (*model.Book, error) {
	s.logger.Log("Getting book by ID")
	return s.store.GetBookByID(id)
}

func (s *Service) CreateBook(book *model.Book) (*model.Book, error) {
	s.logger.Log("Creating book")
	if book.Title == "" || book.Author == "" {
		s.logger.Log("Failed to create book: title and author are required")
		return nil, ErrTitleAndAuthorRequired
	}

	return s.store.CreateBook(book)
}

// UpdateBook cambia solo los campos enviados (no nil); los demás conservan su valor.
func (s *Service) UpdateBook(id int, title, author *string) (*model.Book, error) {
	s.logger.Log("Updating book")
	if (title != nil && *title == "") || (author != nil && *author == "") {
		s.logger.Log("Failed to update book: title and author are required")
		return nil, ErrTitleAndAuthorRequired
	}

	return s.store.UpdateBook(id, title, author)
}

func (s *Service) DeleteBook(id int) error {
	s.logger.Log("Deleting book")
	return s.store.DeleteBook(id)
}
