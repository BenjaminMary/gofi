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
		// REAL are not exacts, even 0.01 through a form is 0.00999999977648258 in DB
		// instead, store money as cents, so 0.01 = 1 and 1.01 = 101
		`
		--DROP TABLE IF EXISTS user;
		CREATE TABLE IF NOT EXISTS user (
			gofiID INTEGER PRIMARY KEY AUTOINCREMENT, 
			email TEXT NOT NULL UNIQUE,
			sessionID TEXT UNIQUE,
			pwHash TEXT NOT NULL,
			numberOfRequests INTEGER DEFAULT 0,
			idleDateModifier TEXT DEFAULT '5 minutes',
			absoluteDateModifier TEXT DEFAULT '1 months',
			idleTimeout TEXT,
			absoluteTimeout TEXT,
			lastLoginTime TEXT DEFAULT '1999-12-31T00:01:01Z',
			lastActivityTime TEXT DEFAULT '1999-12-31T00:01:01',
			lastActivityIPaddress TEXT,
			lastActivityUserAgent TEXT,
			lastActivityAcceptLanguage TEXT,
			dateCreated TEXT NOT NULL
		);

		--DROP TABLE IF EXISTS param;
		CREATE TABLE IF NOT EXISTS param (
			id INTEGER PRIMARY KEY AUTOINCREMENT, 
			gofiID INTEGER NOT NULL,
			paramName TEXT NOT NULL,
			paramJSONstringData TEXT NOT NULL,
			paramInfo TEXT NOT NULL,
			UNIQUE(gofiID, paramName)
		);

		--DROP TABLE IF EXISTS financeTracker;
		CREATE TABLE IF NOT EXISTS financeTracker (
			id INTEGER PRIMARY KEY AUTOINCREMENT, 
			gofiID INTEGER NOT NULL,
			year INTEGER DEFAULT (strftime('%Y','now')),
			month INTEGER DEFAULT (strftime('%m','now')),
			day INTEGER DEFAULT (strftime('%d','now')),
			account TEXT DEFAULT 'CB',
			product TEXT NOT NULL,
			priceIntx100 INTEGER NOT NULL,
			category TEXT NOT NULL,
			commentInt INTEGER DEFAULT 0,
			commentString TEXT DEFAULT '',
			checked INTEGER DEFAULT 0,
			dateChecked TEXT DEFAULT '9999-12-31',
			sentToSheets INTEGER DEFAULT 0
		);
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
