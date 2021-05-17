package data

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
)

type DatabaseCredentials struct{
	Host string
	Port int
	Username string
	Password string
	DatabaseName string
}

func NewDB(credentials DatabaseCredentials) (*sql.DB, error) {
	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", credentials.Host, credentials.Port, credentials.Username, credentials.Password, credentials.DatabaseName)

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Printf("Error on opening sql conn: %v", err)
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		log.Printf("Error on pinging database: %v", err)
		return db, err
	}

	return db, nil
}