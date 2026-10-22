package infrastructure

import (
	"database/sql"

	_ "modernc.org/sqlite" // registra el driver "sqlite"
)

// OpenSQLite abre la base SQLite en path (la crea si no existe) y asegura que exista la tabla books.
func OpenSQLite(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	// SQLite tolera múltiples lectores; los escritores se serializan naturalmente.
	db.SetMaxOpenConns(25)

	// AUTOINCREMENT evita reutilizar el ID de un libro borrado.
	// El índice único impide dos libros con el mismo título y autor, aunque lleguen peticiones a la vez.
	// Si la base ya tiene duplicados, el índice no se crea y el arranque falla.
	for _, stmt := range []string{
		`CREATE TABLE IF NOT EXISTS books (
			id     INTEGER PRIMARY KEY AUTOINCREMENT,
			title  TEXT NOT NULL,
			author TEXT NOT NULL
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS books_title_author ON books (title, author)`,
	} {
		if _, err := db.Exec(stmt); err != nil {
			db.Close()
			return nil, err
		}
	}

	return db, nil
}
