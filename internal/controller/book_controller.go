package controller

import (
	"002-crud-api/internal/model"
	"002-crud-api/internal/service"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
)

type BookController struct {
	service *service.Service
}

func NewBookController(service *service.Service) *BookController {
	return &BookController{service: service}
}

// RegisterRoutes registra las rutas CRUD de libros en el mux.
func (c *BookController) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /books", c.getAllBooks)
	mux.HandleFunc("GET /books/{id}", c.getBookByID)
	mux.HandleFunc("POST /books", c.createBook)
	mux.HandleFunc("PATCH /books/{id}", c.patchBook)
	mux.HandleFunc("DELETE /books/{id}", c.deleteBook)
}

func (c *BookController) getAllBooks(w http.ResponseWriter, r *http.Request) {
	books, err := c.service.GetAllBooks()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, books)
}

func (c *BookController) getBookByID(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	book, err := c.service.GetBookByID(id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, book)
}

func (c *BookController) createBook(w http.ResponseWriter, r *http.Request) {
	var book model.Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	created, err := c.service.CreateBook(&book)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// patchBook actualiza solo los campos que vengan en el JSON; los demás conservan su valor.
func (c *BookController) patchBook(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	// Los punteros distinguen un campo ausente (nil) de uno enviado con cadena vacía.
	var fields struct {
		Title  *string `json:"title"`
		Author *string `json:"author"`
	}
	if err := json.NewDecoder(r.Body).Decode(&fields); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	updated, err := c.service.UpdateBook(id, fields.Title, fields.Author)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (c *BookController) deleteBook(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := c.service.DeleteBook(id); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseID(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid book id", http.StatusBadRequest)
		return 0, false
	}
	return id, true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// Al cliente no se le muestra el error interno; se registra en el log del servidor.
func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		http.Error(w, "book not found", http.StatusNotFound)
	case errors.Is(err, service.ErrTitleAndAuthorRequired):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, service.ErrDuplicateBook):
		http.Error(w, err.Error(), http.StatusConflict)
	default:
		log.Printf("book request failed: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
