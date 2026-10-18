package main

import (
	"002-crud-api/internal/controller"
	"002-crud-api/internal/infrastructure"
	"002-crud-api/internal/service"
	"encoding/json"
	"log"
	"net/http"
)

// logger adapta el paquete log estándar a la interfaz service.Logger.
type logger struct{}

func (logger) Log(message string) {
	log.Println(message)
}

func main() {
	db, err := infrastructure.OpenSQLite("books.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	svc := service.NewService(infrastructure.NewStore(db), logger{})

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/alive", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"alive": true})
	})
	controller.NewBookController(svc).RegisterRoutes(mux)

	log.Println("escuchando en http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
