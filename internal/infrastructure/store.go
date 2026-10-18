package infrastructure

import (
	"database/sql"

	"002-crud-api/internal/model"
)

// Store es el contrato de las operaciones sobre libros.
type Store interface {
	GetAllBooks() ([]*model.Book, error)
	GetBookByID(id int) (*model.Book, error)
	CreateBook(book *model.Book) (*model.Book, error)
	UpdateBook(id int, title, author *string) (*model.Book, error)
	DeleteBook(id int) error
}

// store es la implementación de Store respaldada por una base de datos SQL.
type store struct {
	db *sql.DB
}

// NewStore crea un Store que usa la conexión db.
//
// Devuelve un *store oculto tras la interfaz Store: quien lo usa solo ve los
// métodos del contrato, no la implementación concreta.
func NewStore(db *sql.DB) Store {
	return &store{db: db}
}
