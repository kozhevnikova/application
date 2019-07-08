package main

import (
	"database/sql"
	"net/http"
)

type Server struct {
	Database *sql.DB
	Static   http.Handler
	Dynamic  http.Handler
}

func NewServer(config Config) (*Server, error) {

	dbp, err := CreateDBConnection(config)
	if err != nil {
		return nil, err
	}

	server := &Server{
		Database: dbp,
		Static:   http.FileServer(http.Dir("templates/styles")),
		Dynamic:  http.FileServer(http.Dir("templates/js")),
	}

	return server, nil
}
