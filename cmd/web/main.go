package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5"
)

const webPort = "80"

func main() {

	db := initDB()
	db.Ping()
}

func initDB() *sql.DB {
	conn := connectToDB()

	if conn == nil {
		log.Panic("can't connect to database!!!")
	}

	return conn
}

func connectToDB() *sql.DB {
	counts := 0

	dsn := os.Getenv("DSN")

	for {
		connection, err := openDB(dsn)
		if err != nil {
			log.Println("Postgres not ready yet...")
		} else {
			log.Println("Connected to database!")
			return connection
		}

		if counts > 10 {
			return nil
		}

		log.Println("Backing off for 1 second")
		time.Sleep(1 * time.Second)
		counts++
		continue
	}
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil

}
