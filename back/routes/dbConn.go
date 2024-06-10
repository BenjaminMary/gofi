package routes

import (
	"database/sql"
	"fmt"
	"log"

	"gofi/gofi/data/appdata"

	_ "modernc.org/sqlite"
)

func OpenDbCon() *sql.DB {
	db, err := sql.Open("sqlite", appdata.DbPath)
	if err != nil {
		log.Fatal("error opening DB file: ", err)
	}
	err = db.Ping()
	if err != nil {
		panic("can't ping DB")
	}
	db.SetMaxIdleConns(1) //default 2
	db.SetMaxOpenConns(3) //default 0 = infinite
	return db
}

func CloseDbCon(db *sql.DB) error {
	fmt.Println("defer : PRAGMA optimize then db.Close() called from main")
	db.Exec("PRAGMA optimize;") // to run just before closing each database connection.
	return db.Close()
}
