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
		--DROP TABLE IF EXISTS modes;
		CREATE TABLE IF NOT EXISTS modes (
			mode INTEGER NOT NULL, 
			info TEXT NOT NULL
		);
		INSERT INTO modes
		VALUES 
			(0, '+- standard'),
			(1, '+ emprunt'),
			(2, '- pret'),
			(3, '- remboursement emprunt'),
			(4, '+ remboursement pret')
		;

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
			catWhereToUse TEXT NOT NULL, /* 'all'= everywhere, 'periodic'= recurrent record, 'specific'= transfer/lendborrow, 'basic'= standard record */
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
		/* https://personal.sron.nl/~pault/ figure 4 */
		/*
		DELETE FROM category;
		INSERT INTO category (gofiID, category, catWhereToUse, catOrder, inUse,
			iconName, iconCodePoint, colorName, colorHSL, colorHEX)
		VALUES 
			(-1, 'Besoin', 		'all', 		1, 1, 'bed', 'e91f', 'needfix-teal', '(170,43,47)', '#44AA99'),
			(-1, 'Envie', 		'all', 		2, 1, 'film', 'e920', 'wantko-wine', '(330,60,33)', '#882255'),
			(-1, 'Revenu', 		'periodic', 3, 1, 'credit-card', 'e903', 'invest-cyan', '(200,75,73)', '#88CCEE'),
			(-1, 'Epargne', 	'all', 		4, 1, 'line-chart', 'e904', 'invest-cyan', '(200,75,73)', '#88CCEE'),
			(-1, 'Habitude-', 	'all', 		5, 0, 'thumbs-down', 'e91e', 'wantko-wine', '(330,60,33)', '#882255'),
			(-1, 'Vehicule', 	'all', 		6, 0, 'car-front', 'e900', 'needvar-olive', '(60,50,40)', '#999933'),
			(-1, 'Transport', 	'all', 		7, 0, 'train-front', 'e913', 'needvar-olive', '(60,50,40)', '#999933'),
			(-1, 'Shopping', 	'basic', 	8, 0, 'shopping-cart', 'e918', 'wantko-wine', '(330,60,33)', '#882255'),
			(-1, 'Cadeaux', 	'basic', 	9, 0, 'gift', 'e91a', 'wantko-wine', '(330,60,33)', '#882255'),
			(-1, 'Courses', 	'all', 		10, 0, 'carrot', 'e916', 'needvar-olive', '(60,50,40)', '#999933'),
			(-1, 'Resto', 		'basic', 	11, 0, 'chef-hat', 'e914', 'wantok-purple', '(310,43,47)', '#AA4499'),
			(-1, 'Loisirs', 	'all', 		12, 0, 'drama', 'e901', 'wantok-purple', '(310,43,47)', '#AA4499'),
			(-1, 'Voyage', 		'basic', 	13, 0, 'earth', 'e902', 'wantok-purple', '(310,43,47)', '#AA4499'),
			(-1, 'Enfants', 	'all', 		14, 0, 'baby', 'e91d', 'needvar-olive', '(60,50,40)', '#999933'),
			(-1, 'Banque', 		'all', 		15, 0, 'landmark', 'e919', 'needvar-olive', '(60,50,40)', '#999933'),
			(-1, 'Societe', 	'all', 		16, 0, 'briefcase', 'e905', 'invest-cyan', '(200,75,73)', '#88CCEE'),
			(-1, 'Loyer', 		'periodic', 17, 0, 'home', 'e906', 'needfix-teal', '(170,43,47)', '#44AA99'),
			(-1, 'Services', 	'periodic', 18, 0, 'receipt-text', 'e924', 'needfix-teal', '(170,43,47)', '#44AA99'),
			(-1, 'Sante', 		'all', 		19, 0, 'heart-pulse', 'e908', 'needvar-olive', '(60,50,40)', '#999933'),
			(-1, 'Animaux', 	'all', 		20, 0, 'paw-print', 'e91c', 'wantko-wine', '(330,60,33)', '#882255'),
			(-1, 'Taxes', 		'periodic', 21, 0, 'calculator', 'e925', 'needvar-olive', '(60,50,40)', '#999933'),
			(-1, 'Assurance', 	'periodic', 22, 0, 'shield-check', 'e926', 'needfix-teal', '(170,43,47)', '#44AA99'),
			(-1, 'Telecom', 	'periodic', 23, 0, 'wifi', 'e927', 'needfix-teal', '(170,43,47)', '#44AA99'),
			(-1, 'Energie', 	'periodic', 24, 0, 'cable', 'e928', 'needfix-teal', '(170,43,47)', '#44AA99'),
			(-1, 'Eau', 		'periodic', 25, 0, 'droplet', 'e929', 'needfix-teal', '(170,43,47)', '#44AA99')
		;
		INSERT INTO category (gofiID, category, catWhereToUse, catOrder, inUse, defaultInStats,
			description,
			iconName, iconCodePoint, colorName, colorHSL, colorHEX)
		VALUES
			(-1, 'Autre', 		'basic', 	26, 0, 1, 
				'Permet de ranger un élément qu''on ne sait pas où placer, temporairement ou définitivement.',
				'more-horizontal', 'e90c', 'system-lightgrey', '(0,0,87)', '#DDDDDD'),
			(-1, 'Erreur', 		'basic', 	27, 0, 1, 
				'Utile lorsqu''on souhaite corriger un montant global sans savoir réellement quel était l''achat en question.',
				'bug', 'e909', 'system-lightgrey', '(0,0,87)', '#DDDDDD'),
			(-1, 'Pret', 	'specific', -2, 1, 0, 
				'Utilisable uniquement par le système lors de l''utilisation de la fonction prêt.',
				'handshake', 'e922', 'system-lightgrey', '(0,0,87)', '#DDDDDD'),
			(-1, 'Emprunt', 	'specific', -1, 1, 0, 
				'Utilisable uniquement par le système lors de l''utilisation de la fonction emprunt.',
				'handshake', 'e922', 'system-lightgrey', '(0,0,87)', '#DDDDDD'),
			(-1, 'Transfert', 	'specific', 97, 1, 0, 
				'Utilisé uniquement par le système lors de l''utilisation de la fonction transfert.',
				'arrow-right-left', 'e91b', 'system-lightgrey', '(0,0,87)', '#DDDDDD'),
			(-1, '?', 			'specific', 98, 1, 0, 
				'Utilisé uniquement comme icône par le système lorsqu''aucune icône ne correspond à la catégorie demandée.',
				'help-circle', 'e90a', 'system-lightgrey', '(0,0,87)', '#DDDDDD'),
			(-1, '-', 			'specific', 99, 1, 0, 
				'Utilisé uniquement par le système lorsqu''on supprime une ligne.',
				'trash-2', 'e90b', 'system-lightgrey', '(0,0,87)', '#DDDDDD')
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
			mode INTEGER DEFAULT 0,
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

		--DROP TABLE IF EXISTS lenderBorrower;
		CREATE TABLE IF NOT EXISTS lenderBorrower (
			id INTEGER PRIMARY KEY AUTOINCREMENT, 
			gofiID INTEGER NOT NULL,
			name TEXT NOT NULL,
			isActive INTEGER DEFAULT 1
			
			-- dateFirstLentBorrowed TEXT DEFAULT '9999-12-31',
			-- dateLastLentBorrowed TEXT DEFAULT '9999-12-31',
			-- numberLentBorrowed INTEGER DEFAULT 0,
			-- sumIntx100lentBorrowed INTEGER DEFAULT 0,
			-- sumIntx100refunded INTEGER DEFAULT 0
		);

		--DROP TABLE IF EXISTS specificRecordsByMode;
		CREATE TABLE IF NOT EXISTS specificRecordsByMode (
			id INTEGER PRIMARY KEY AUTOINCREMENT, 
			gofiID INTEGER NOT NULL,
			mode INTEGER NOT NULL,
			idFinanceTracker INTEGER NOT NULL,
			idLenderBorrower INTEGER DEFAULT 0

			-- parentIdIfRefund INTEGER DEFAULT 0,
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
