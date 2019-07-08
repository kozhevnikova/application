package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func CreateDBConnection(config Config) (*sql.DB, error) {

	dbinfo := fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s sslmode=disable",
		config.Database.User,
		config.Database.Password,
		config.Database.Name,
		config.Database.Host,
	)

	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
