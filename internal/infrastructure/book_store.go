package infrastructure

import (
	"database/sql"
	"errors"

	"002-crud-api/internal/model"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

// ErrDuplicateBook indica que ya existe un libro con el mismo título y autor.
var ErrDuplicateBook = errors.New("a book with the same title and author already exists")

// isDuplicate indica si err es la violación del índice único de título y autor.
func isDuplicate(err error) bool {
	var se *sqlite.Error
	return errors.As(err, &se) && se.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE
}

// Si la consulta o la lectura de alguna fila falla, devuelve el error y ningún libro.
func (s *store) GetAllBooks() ([]*model.Book, error) {
	query := "SELECT id, title, author FROM books"
	rows, err := s.db.Query(query)

	if err != nil {
		return nil, err
	}

	// si es rows, hay que cerrar para liberar recursos
	defer rows.Close()

	// Se inicializa vacío para que la API devuelva [] y no null cuando no hay libros.
	books := []*model.Book{}

	for rows.Next() {
		var b model.Book
		// El Scan sirve para leer los valores de la fila actual en las variables que le pasemos.
		if err := rows.Scan(&b.ID, &b.Title, &b.Author); err != nil {
			return nil, err
		}

		books = append(books, &b)
	}

	// rows.Next() también termina cuando hay un error; hay que revisarlo aquí.
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return books, nil
}

// Si no existe ningún libro con ese ID, devuelve sql.ErrNoRows.
func (s *store) GetBookByID(id int) (*model.Book, error) {
	query := `SELECT id, title, author FROM books WHERE id = ?`

	var b model.Book

	/*
		Scan devuelve sql.ErrNoRows si no hay ninguna fila que leer, y ese error se propaga
	*/
	err := s.db.QueryRow(query, id).Scan(&b.ID, &b.Title, &b.Author)
	if err != nil {
		return nil, err
	}

	return &b, nil
}

// Si el INSERT o el ID generado fallan, devuelve el error y ningún libro.
// Si ya existe un libro con el mismo título y autor, devuelve ErrDuplicateBook.
func (s *store) CreateBook(book *model.Book) (*model.Book, error) {
	res, err := s.db.Exec("INSERT INTO books (title, author) VALUES (?, ?)", book.Title, book.Author)
	if isDuplicate(err) {
		return nil, ErrDuplicateBook
	}
	if err != nil {
		return nil, err
	}

	// LastInsertId devuelve el ID generado por la base de datos para la fila insertada.
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	book.ID = int(id)
	return book, nil
}

// Cambia solo los campos que no sean nil; el resto conserva su valor. Es una sola sentencia,
// así dos actualizaciones al mismo libro no se pisan. Si no existe ningún libro con ese ID, devuelve sql.ErrNoRows.
// Si el cambio dejaría dos libros con el mismo título y autor, devuelve ErrDuplicateBook.
func (s *store) UpdateBook(id int, title, author *string) (*model.Book, error) {
	var b model.Book
	err := s.db.QueryRow(
		`UPDATE books SET title = COALESCE(?, title), author = COALESCE(?, author) WHERE id = ? RETURNING id, title, author`,
		title, author, id,
	).Scan(&b.ID, &b.Title, &b.Author)
	if isDuplicate(err) {
		return nil, ErrDuplicateBook
	}
	if err != nil {
		return nil, err
	}

	return &b, nil
}

// Si no existe ningún libro con ese ID, devuelve sql.ErrNoRows.
func (s *store) DeleteBook(id int) error {
	res, err := s.db.Exec("DELETE FROM books WHERE id = ?", id)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}

	return nil
}
