# Usa la misma versión de Go que go.mod.
FROM golang:1.27.2-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# modernc.org/sqlite es Go puro, así que no hace falta CGO.
RUN CGO_ENABLED=0 go build -o /out/server .

FROM alpine:3.20
# Railway monta un volumen en /data; ahí vive la base para que sobreviva a los despliegues.
RUN mkdir /data
COPY --from=build /out/server /server
ENV DB_PATH=/data/books.db
CMD ["/server"]
