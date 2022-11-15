package main

import (
	"GOUA/internal/data"
	"GOUA/internal/driver"
	"log"
	"net/http"
	"os"
)

type config struct {
	port string
}

type application struct {
	config      config
	infoLog     *log.Logger
	errorLog    *log.Logger
	models      data.Models
	environment string
}

func main() {
	// CONFIG
	var cfg config
	cfg.port = ":8082"

	// LOG
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// DB
	dsn := os.Getenv("DSN")
	environment := os.Getenv("ENV")

	db, err := driver.ConnectPostgres(dsn)
	if err != nil {
		log.Fatal("Cannot connect to DB")
	}

	// APPLICATION STRUCTURE
	app := application{
		config:      cfg,
		infoLog:     infoLog,
		errorLog:    errorLog,
		models:      data.New(db.SQL),
		environment: environment,
	}

	// SERVE
	err = app.serve()
	if err != nil {
		log.Fatal(err)
	}
}

func (app application) serve() error {
	app.infoLog.Println("API listening on port", app.config.port)

	srv := &http.Server{
		Addr:    app.config.port,
		Handler: app.routes(),
	}

	return srv.ListenAndServe() // method attached to &http.Server
}
