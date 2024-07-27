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
			gofiID INTEGER, -- NOT NULL,
			category TEXT NOT NULL,
			catWhereToUse TEXT NOT NULL, /* 'all'= everywhere, 'periodic'= recurrent record, 'specific'= transfer, 'basic'= standard record */
			catOrder INTEGER, -- NOT NULL,
			inUse INTEGER DEFAULT 1,
			defaultInStats INTEGER DEFAULT 1,
			description TEXT DEFAULT '-',
			budgetPrice INTEGER DEFAULT 0,
			budgetPeriod TEXT DEFAULT '-', /* -,month,year,week */
			budgetType TEXT DEFAULT '-', /* -,cumulative,reset */
			budgetCurrentPeriodStartDate TEXT DEFAULT '9999-12-30',
			budgetCurrentPeriodEndDate TEXT DEFAULT '9999-12-31',
			iconName TEXT NOT NULL,
			iconCodePoint TEXT NOT NULL, /* "error" icon exemple: e909 */
			colorName TEXT NOT NULL,
			colorHSL TEXT NOT NULL,
			colorHEX TEXT NOT NULL
		);
		/*
		DELETE FROM category;
		INSERT INTO category (gofiID, category, catWhereToUse, catOrder, inUse,
			iconName, iconCodePoint, colorName, colorHSL, colorHEX)
		VALUES 
			(-1, 'Besoin', 		'all', 		1, 1, 'bed', 'e91f', 'green', '(130,60,50)', '#33CC4C'),
			(-1, 'Envie', 		'all', 		2, 1, 'film', 'e920', 'orange', '(30,60,50)', '#CC8033'),
			(-1, 'Revenu', 		'periodic', 3, 1, 'credit-card', 'e903', 'teal', '(160,60,50)', '#33CC99'),
			(-1, 'Epargne', 	'all', 		4, 1, 'line-chart', 'e904', 'light blue', '(210,60,50)', '#3380CC'),
			(-1, 'Habitude-', 	'all', 		5, 0, 'thumbs-down', 'e91e', 'red', '(1,60,50)', '#CC3633'),
			(-1, 'Vehicule', 	'all', 		6, 0, 'car-front', 'e900', 'orange', '(15,60,50)', '#CC5933'),
			(-1, 'Transport', 	'all', 		7, 0, 'train-front', 'e913', 'orange', '(30,60,50)', '#CC8033'),
			(-1, 'Shopping', 	'basic', 	8, 0, 'shopping-cart', 'e918', 'yellow', '(45,40,50)', '#B3994D'),
			(-1, 'Cadeaux', 	'basic', 	9, 0, 'gift', 'e91a', 'yellow', '(60,40,50)', '#B3B34D'),
			(-1, 'Courses', 	'all', 		10, 0, 'carrot', 'e916', 'yellow', '(70,50,50)', '#AABF40'),
			(-1, 'Resto', 		'basic', 	11, 0, 'chef-hat', 'e914', 'green', '(90,60,50)', '#80CC33'),
			(-1, 'Loisirs', 	'all', 		12, 0, 'drama', 'e901', 'green', '(110,60,50)', '#4DCC33'),
			(-1, 'Voyage', 		'basic', 	13, 0, 'earth', 'e902', 'green', '(130,60,50)', '#33CC4C'),
			(-1, 'Enfants', 	'all', 		14, 0, 'baby', 'e91d', 'teal', '(175,60,50)', '#33CCBF'),
			(-1, 'Banque', 		'all', 		15, 0, 'landmark', 'e919', 'light blue', '(190,60,50)', '#33B3CC'),
			(-1, 'Societe', 	'all', 		16, 0, 'briefcase', 'e905', 'blue', '(230,60,50)', '#334CCC'),
			(-1, 'Loyer', 		'periodic', 17, 0, 'home', 'e906', 'purple', '(260,60,50)', '#6633CC'),
			(-1, 'Services', 	'periodic', 18, 0, 'plug-zap', 'e907', 'purple', '(270,60,50)', '#8033CC'),
			(-1, 'Sante', 		'all', 		19, 0, 'heart-pulse', 'e908', 'pink', '(300,60,50)', '#CC33CC'),
			(-1, 'Animaux', 	'all', 		20, 0, 'paw-print', 'e91c', 'pink', '(320,60,50)', '#CC3399')
		;
		INSERT INTO category (gofiID, category, catWhereToUse, catOrder, inUse, defaultInStats,
			description,
			iconName, iconCodePoint, colorName, colorHSL, colorHEX)
		VALUES
			(-1, 'Autre', 		'basic', 	21, 0, 1, 
				'Permet de ranger un élément qu''on ne sait pas où placer, temporairement ou définitivement.',
				'more-horizontal', 'e90c', 'grey', '(0,0,60)', '#999999'),
			(-1, 'Erreur', 		'basic', 	22, 0, 1, 
				'Utile lorsqu''on souhaite corriger un montant global sans savoir réellement quel était l''achat en question.',
				'bug', 'e909', 'red', '(335,60,50)', '#CC3373'),
			(-1, 'Transfert', 	'specific', 97, 1, 0, 
				'Utilisé uniquement par le système lors de l''utilisation de la fonction transfert.',
				'arrow-right-left', 'e91b', 'grey', '(0,0,40)', '#666666'),
			(-1, '?', 			'specific', 98, 1, 0, 
				'Utilisé uniquement comme icône par le système lorsqu''aucune icône ne correspond à la catégorie demandée.',
				'help-circle', 'e90a', 'grey', '(0,0,50)', '#808080'),
			(-1, '-', 			'specific', 99, 1, 0, 
				'Utilisé uniquement par le système lorsqu''on supprime une ligne.',
				'trash-2', 'e90b', 'red', '(1,60,50)', '#CC3633')
		;
		*/

		--DROP TABLE IF EXISTS financeTracker;
		CREATE TABLE IF NOT EXISTS financeTracker (
			id INTEGER PRIMARY KEY AUTOINCREMENT, 
			gofiID INTEGER NOT NULL,
			dateIn TEXT DEFAULT '1999-12-31',
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
