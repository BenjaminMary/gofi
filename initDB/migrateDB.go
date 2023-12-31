package main

import (
	"database/sql"
	"fmt"
	"log"

	"example.com/sqlite"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func main() {
	var err error
	fmt.Println(sqlite.DbPath)

	if len(sqlite.DbPath) == 0 {
		log.Fatal("specify the SQLITE_DB_FILENAME environment variable")
	}
	db, err = sql.Open("sqlite", sqlite.DbPath)
	if err != nil {
		log.Fatal("error opening the SQLite file")
	}
	db.Exec("PRAGMA journal_mode = WAL;")
	db.Exec("PRAGMA synchronous = normal;")
	db.Exec("PRAGMA vacuum;")
	
	defer fmt.Println("1st defer : db.Close(), played last")
	defer db.Close()
	defer fmt.Println("2nd defer : played 1st, PRAGMA optimize to run just before closing each database connection.")
	defer db.Exec("PRAGMA optimize;") // to run just before closing each database connection.

	_, err = db.Exec(
		// LAST MIGRATION TO PLAY
		`
		ALTER TABLE financeTracker 
		RENAME COLUMN sentToSheets TO exported;
		`,
	)
	if err != nil {
		log.Fatal("error executing requests: ", err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal("error initializing DB connection: ping error: ", err)
	}
	fmt.Println("database initialized..")
}
