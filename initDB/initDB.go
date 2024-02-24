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

		--DROP TABLE IF EXISTS category;
		CREATE TABLE IF NOT EXISTS category (
			id INTEGER PRIMARY KEY AUTOINCREMENT, 
			--gofiID INTEGER NOT NULL,
			category TEXT NOT NULL,
			iconName TEXT NOT NULL,
			iconCodePoint TEXT NOT NULL, /* "error" icon exemple: HTML = &#xe000;  |  JS = \ue000 */
			colorName TEXT NOT NULL,
			colorHSL TEXT NOT NULL,
			colorHEX TEXT NOT NULL
		);
		DELETE FROM category;
		INSERT INTO category (category, iconName, iconCodePoint, colorName, colorHSL, colorHEX)
		VALUES 
			('Transport', 'train', 'e570', 'orange', '(15,60,50)', '#CC5933'),
			('Véhicule', 'directions_car', 'e531', 'orange', '(25,60,50)', '#CC7633'),
			('Courses', 'grocery', 'ef97', 'yellow', '(50,40,50)', '#B3A24D'),
			('Shopping', 'shopping_cart', 'e8cc', 'yellow', '(55,40,50)', '#B3AA4D'),
			('Cadeaux', 'redeem', 'e8b1', 'yellow', '(60,40,50)', '#B3B34D'),
			('Restaurant', 'tapas', 'f1e9', 'green', '(120,60,50)', '#33CC33'),
			('Loisirs', 'theater_comedy', 'ea66', 'green', '(125,60,50)', '#33CC40'),
			('Voyage', 'travel_explore', 'e2db', 'green', '(130,60,50)', '#33CC4C'),
			('Salaire', 'add_card', 'eb86', 'teal', '(160,60,50)', '#33CC99'),
			('Banque', 'account_balance', 'e84f', 'light blue', '(190,60,50)', '#33B3CC'),
			('Investissement', 'real_estate_agent', 'e73a', 'light blue', '(200,60,50)', '#3399CC'),
			('Entreprise', 'enterprise', 'e70e', 'blue', '(230,60,50)', '#334CCC'),
			('Loyer', 'cottage', 'e587', 'purple', '(260,60,50)', '#6633CC'),
			('Services', 'wifi_home', 'f671', 'purple', '(270,60,50)', '#8033CC'),
			('Santé', 'heart_plus', 'f884', 'pink', '(300,60,50)', '#CC33CC'),
			('Erreur', 'error', 'e000', 'red', '(335,60,50)', '#CC3373')
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
