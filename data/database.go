package data

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"time"
)

type DatabaseCredentials struct{
	Host string
	Port int
	Username string
	Password string
	DatabaseName string
}

type PoolSettings struct{
	MaxOpenConns int
	MaxIdleConns int
	ConnMaxLifeTime time.Duration
}

func NewDB(credentials DatabaseCredentials, settings PoolSettings) (*sql.DB, error) {
	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", credentials.Host, credentials.Port, credentials.Username, credentials.Password, credentials.DatabaseName)

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Printf("Error on opening sql conn: %v", err)
		return nil, err
	}

	/*
	db.SetMaxOpenConns(settings.MaxOpenConns)
	db.SetMaxIdleConns(settings.MaxIdleConns)
	db.SetConnMaxLifetime(settings.ConnMaxLifeTime)

	 */

	err = db.Ping()
	if err != nil {
		log.Printf("Error on pinging database: %v", err)
		return db, err
	}

	return db, nil
}