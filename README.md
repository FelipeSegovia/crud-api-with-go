# 002-crud-api

API REST de libros (CRUD) escrita en Go. Usa `net/http` de la biblioteca estándar para las rutas y SQLite para guardar los datos.

## Estructura

- `main.go`: abre la base de datos, arma las capas y levanta el servidor en el puerto 8080.
- `internal/model/`: el tipo `Book`.
- `internal/infrastructure/`: acceso a SQLite. `db.go` abre la base y crea la tabla; `book_store.go` hace las consultas.
- `internal/service/`: lógica de negocio y validaciones.
- `internal/controller/`: rutas HTTP y respuestas JSON.

## Requisitos

- Go 1.27.2 o superior (lo indica `go.mod`).

## Cómo ejecutar

1. Ubícate en la raíz del proyecto.
2. Arranca el servidor con `go run .`. Las dependencias se descargan solas la primera vez.
3. La API queda en `http://localhost:8080`.

Reglas:

- Ejecuta siempre desde la raíz del proyecto. La base `books.db` se crea en la carpeta desde donde ejecutas el programa.
- No borres `books.db` si tiene datos que quieras conservar.
- No edites `go.sum` a mano.
- Corre los tests con `go test ./...`.
- Antes de subir cambios, revisa el formato con `gofmt -l .`. Si no imprime nada, está bien.

## Rutas

| Método | Ruta          | Descripción                               | Respuestas        |
|--------|---------------|-------------------------------------------|-------------------|
| GET    | `/books`      | Lista todos los libros                    | 200               |
| GET    | `/books/{id}` | Obtiene un libro                          | 200, 404          |
| POST   | `/books`      | Crea un libro (`title` y `author`)        | 201, 400, 409     |
| PATCH  | `/books/{id}` | Actualiza solo los campos enviados        | 200, 400, 404, 409 |
| DELETE | `/books/{id}` | Borra un libro                            | 204, 404          |

Ejemplo:

```bash
curl -X POST localhost:8080/books -H 'Content-Type: application/json' -d '{"title":"Go","author":"Pike"}'
curl localhost:8080/books
```

## Reglas de negocio

- No pueden existir dos libros con el mismo título y el mismo autor. La comparación distingue mayúsculas y espacios.
- Sí pueden existir dos libros con el mismo título y distinto autor, o con el mismo autor y distinto título.
- Si una petición crea o deja dos libros con el mismo título y autor, responde 409 Conflict.

## Salud del servidor

`GET /health/alive` responde `{"alive":true}` con 200 mientras el servidor esté activo.

```bash
curl localhost:8080/health/alive
```

## Despliegue en Railway

El `Dockerfile` de la raíz lo usa Railway para construir la imagen.

1. Crea un servicio en Railway desde este repositorio de GitHub.
2. Agrega un volumen al servicio con punto de montaje `/data`. Sin volumen, los libros se pierden en cada despliegue.
3. Genera un dominio público desde la configuración de red del servicio.

Variables de entorno:

- `PORT`: la asigna Railway. En local la API usa 8080.
- `DB_PATH`: ruta de la base. El `Dockerfile` la fija en `/data/books.db`. En local usa `books.db`.

## go.mod y go.sum

- **`go.mod`** declara el nombre del módulo (`002-crud-api`), la versión de Go y las dependencias con su versión. Las que no usas directamente aparecen marcadas con `// indirect`.
- **`go.sum`** guarda un hash de cada dependencia. Go lo usa para comprobar que el código descargado es el mismo. Se genera con `go get` o `go mod tidy`; no se edita a mano.

Ambos archivos se suben al control de versiones.
