package main

import (
	"002-crud-api/internal/controller"
	"002-crud-api/internal/infrastructure"
	"002-crud-api/internal/service"
	"encoding/json"
	"log"
	"net/http"
	"os"
)

// logger adapta el paquete log estándar a la interfaz service.Logger.
type logger struct{}

func (logger) Log(message string) {
	log.Println(message)
}

func main() {
	db, err := infrastructure.OpenSQLite(env("DB_PATH", "books.db"))
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

	port := env("PORT", "8080") // Railway asigna PORT en cada despliegue
	log.Println("escuchando en el puerto " + port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

// env devuelve el valor de la variable de entorno key, o fallback si no está definida.
func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
