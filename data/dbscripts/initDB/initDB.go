/*
$Env:EXE_PATH="C:\git\gofi"
$Env:SQLITE_DB_FILENAME="gofi.db"
cd c:\git\gofi\
go run ./data/dbscripts/initDB
*/

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func initDB(folder string, dbName string) {
	var err error
	dataFilePath := filepath.Join(os.Getenv("EXE_PATH"), "data")
	file, err := os.Create(filepath.Join(dataFilePath, folder, dbName))
	if err != nil {
		fmt.Printf("err: %v\n", err)
		log.Fatal("error creating the SQLite file")
	}
	file.Close()
	dbPath := filepath.Join(dataFilePath, folder, dbName)
	fmt.Println(dbPath)

	if len(dbPath) == 0 {
		log.Fatal("specify the SQLITE_DB_FILENAME environment variable")
	}
	db, err = sql.Open("sqlite", dbPath)
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
			idleTimeout TEXT DEFAULT '1999-12-31T00:01:01Z',
			absoluteTimeout TEXT DEFAULT '1999-12-31T00:01:01Z',
			lastLoginTime TEXT DEFAULT '1999-12-31T00:01:01Z',
			lastActivityTime TEXT DEFAULT '1999-12-31T00:01:01Z',
			lastActivityIPaddress TEXT DEFAULT '-',
			lastActivityUserAgent TEXT DEFAULT '-',
			lastActivityAcceptLanguage TEXT DEFAULT '-',
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

		--DROP TABLE IF EXISTS category;
		CREATE TABLE IF NOT EXISTS category (
			id INTEGER PRIMARY KEY AUTOINCREMENT, 
			--gofiID INTEGER NOT NULL,
			category TEXT NOT NULL,
			iconName TEXT NOT NULL,
			iconCodePoint TEXT NOT NULL, /* "error" icon exemple: e909*/
			colorName TEXT NOT NULL,
			colorHSL TEXT NOT NULL,
			colorHEX TEXT NOT NULL
		);
		DELETE FROM category;
		INSERT INTO category (category, iconName, iconCodePoint, colorName, colorHSL, colorHEX)
		VALUES 
			('Vehicule', 'car-front', 'e900', 'orange', '(15,60,50)', '#CC5933'),
			('Transp', 'train-front', 'e913', 'orange', '(25,60,50)', '#CC7633'),
			('Courses', 'carrot', 'e916', 'yellow', '(50,40,50)', '#B3A24D'),
			('Shopping', 'shopping-cart', 'e918', 'yellow', '(55,40,50)', '#B3AA4D'),
			('Cadeaux', 'gift', 'e91a', 'yellow', '(60,40,50)', '#B3B34D'),
			('Resto', 'chef-hat', 'e914', 'green', '(120,60,50)', '#33CC33'),
			('Loisirs', 'drama', 'e901', 'green', '(125,60,50)', '#33CC40'),
			('Voyage', 'earth', 'e902', 'green', '(130,60,50)', '#33CC4C'),
			('Salaire', 'credit-card', 'e903', 'teal', '(160,60,50)', '#33CC99'),
			('Banque', 'landmark', 'e919', 'light blue', '(190,60,50)', '#33B3CC'),
			('Invest', 'line-chart', 'e904', 'light blue', '(200,60,50)', '#3399CC'),
			('Societe', 'briefcase', 'e905', 'blue', '(230,60,50)', '#334CCC'),
			('Loyer', 'home', 'e906', 'purple', '(260,60,50)', '#6633CC'),
			('Services', 'plug-zap', 'e907', 'purple', '(270,60,50)', '#8033CC'),
			('Sante', 'heart-pulse', 'e908', 'pink', '(300,60,50)', '#CC33CC'),
			('Erreur', 'bug', 'e909', 'red', '(335,60,50)', '#CC3373'),
			('-', 'trash-2', 'e90b', 'red', '(1,60,50)', '#CC3633'),
			('Autre', 'more-horizontal', 'e90c', 'gris', '(0,0,50)', '#808080')
		;

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
			exported INTEGER DEFAULT 0
		);

		--DROP TABLE IF EXISTS recurrentRecord;
		CREATE TABLE IF NOT EXISTS recurrentRecord (
			id INTEGER PRIMARY KEY AUTOINCREMENT, 
			gofiID INTEGER NOT NULL,
			year INTEGER NOT NULL,
			month INTEGER NOT NULL,
			day INTEGER NOT NULL,
			recurrence TEXT NOT NULL,
			account TEXT NOT NULL,
			product TEXT NOT NULL,
			priceIntx100 INTEGER NOT NULL,
			category TEXT NOT NULL
		);

		--DROP TABLE IF EXISTS backupSave;
		CREATE TABLE IF NOT EXISTS backupSave (
			id INTEGER PRIMARY KEY AUTOINCREMENT, 
			date TEXT DEFAULT '1999-12-31T00:01:01Z',
			extID TEXT NOT NULL,
			extFileName TEXT NOT NULL,
			checkpoint INTEGER DEFAULT 0,
			tested INTEGER DEFAULT 0
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

func main() {
	dbName := os.Getenv("SQLITE_DB_FILENAME")
	initDB("dbscripts", dbName)
}
