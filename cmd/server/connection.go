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

	log.Println("Connect() - successfully connected to db", connStr)

	c, ioErr := os.ReadFile("../../migrations/create_db.sql")
	if ioErr != nil {
		log.Fatal("Connect()- io error ", ioErr)
	}

	sql := string(c)

	_, err = db.Exec(sql)
	if err != nil {
		log.Fatal("Connect() - execution error ", err)
	}

	return db
}
