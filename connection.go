package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func Connect() *sql.DB {
	connStr := "postgres://postgres:1234@localhost:5432/postgres?sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("connected to ", db)
	c, ioErr := os.ReadFile("create_db.sql")
	if ioErr != nil {
		log.Fatal(ioErr)
	}

	sql := string(c)

	_, err = db.Exec(sql)
	if err != nil {
		log.Fatal(err)
	}

	return db
}
