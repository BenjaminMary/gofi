package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func InitDatabase(DbPath string) error {
	var err error
	if len(DbPath) == 0 {
		log.Fatal("specify the SQLITE_DB_FILENAME environment variable")
	}
	fmt.Println(DbPath)
	db, err = sql.Open("sqlite", DbPath)
	if err != nil {
		return err
	}
	db.Exec("PRAGMA journal_mode = WAL;")
	db.Exec("PRAGMA synchronous = normal;")
	db.Exec("PRAGMA vacuum;")
	// db.Exec("PRAGMA optimize;") // to run just before closing each database connection.

	_, err = db.ExecContext(
		// REAL are not exacts, even 0.01 through a form is 0.00999999977648258 in DB
		// instead, store money as cents, so 0.01 = 1 and 1.01 = 101
		context.Background(),`
		--DROP TABLE IF EXISTS param;
		CREATE TABLE IF NOT EXISTS param (
			id INTEGER PRIMARY KEY AUTOINCREMENT, 
			gofiID TEXT NOT NULL,
			paramName TEXT NOT NULL,
			paramJSONstringData TEXT NOT NULL,
			paramInfo TEXT NOT NULL,
			UNIQUE(gofiID, paramName)
		);

		--DROP TABLE IF EXISTS financeTracker;
		CREATE TABLE IF NOT EXISTS financeTracker (
			id INTEGER PRIMARY KEY AUTOINCREMENT, 
			gofiID TEXT NOT NULL,
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
		return err
	}
	err = db.Ping()
	if err != nil {
		log.Fatal("error initializing DB connection: ping error: ", err)
	}
	fmt.Println("database initialized..")

	return nil
}


func CheckIfIdExists(gofiID string) {
	//if new ID, create default params
	err := InitDatabase(DbPath)
	if err != nil {
		log.Fatal("error initializing DB connection: ", err)
		panic("database init error")
	}
	db.Close()
	var nbRows int = 0

	db, err = sql.Open("sqlite", DbPath)
	defer db.Close()
	qA := ` 
		SELECT COUNT(1)
		FROM param
		WHERE gofiID = ?
			AND paramName = 'accountList';
	`
	err = db.QueryRow(qA, gofiID).Scan(&nbRows)
	switch {
		case err == sql.ErrNoRows:
			nbRows = 0
		case err != nil:
			log.Fatalf("query error: %v\n", err)
		//default:
	}
	if nbRows != 1 {
		db.QueryRow("DELETE FROM param WHERE gofiID = ? AND paramName = 'accountList';", gofiID)
		var P1 Param
		P1.GofiID = gofiID
        P1.ParamName = "accountList"
        P1.ParamJSONstringData = "CB,A"
        P1.ParamInfo = "Liste des comptes (séparer par des , sans espaces)"
		InsertRowInParam(&P1)
		db, err = sql.Open("sqlite", DbPath)
		defer db.Close()
	}

	qB := ` 
		SELECT COUNT(1)
		FROM param
		WHERE gofiID = ?
			AND paramName = 'categoryList';
	`
	err = db.QueryRow(qB, gofiID).Scan(&nbRows)
	// defer db.Close()
	switch {
		case err == sql.ErrNoRows:
			nbRows = 0
		case err != nil:
			log.Fatalf("query error param categoryList: %v\n", err)
		//default:
	}
	if nbRows != 1 {
		db.QueryRow("DELETE FROM param WHERE gofiID = ? AND paramName = 'categoryList';", gofiID)
		var P2 Param
		P2.GofiID = gofiID
        P2.ParamName = "categoryList"
        P2.ParamJSONstringData = "Supermarché,Restaurant,Loisir"
        P2.ParamInfo = "Liste des catégories (séparer par des , sans espaces)"
		InsertRowInParam(&P2)
		db, err = sql.Open("sqlite", DbPath)
		defer db.Close()
	}

	return
}

func GetList(ft *FinanceTracker) {
	err := InitDatabase(DbPath)
	if err != nil {
		log.Fatal("error initializing DB connection: ", err)
	}
	q := ` 
		SELECT paramJSONstringData
		FROM param
		WHERE gofiID = ?
			AND paramName = ?;
	`
	rows, err := db.Query(q, ft.GofiID, "accountList")
	defer rows.Close()

	rows.Next()
	var accountList string
	if err := rows.Scan(&accountList); err != nil {
		log.Fatal(err)
	}
	ft.Account = accountList
	ft.AccountList = strings.Split(accountList, ",")
	// fmt.Printf("\naccountList: %v\n", ft.AccountList)

	rows, err = db.Query(q, ft.GofiID, "categoryList")
	defer rows.Close()
	rows.Next()
	var categoryList string
	if err := rows.Scan(&categoryList); err != nil {
		log.Fatal(err)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	ft.Category = categoryList
	ft.CategoryList = strings.Split(categoryList, ",")
	// fmt.Printf("\ncategoryList: %v\n", ft.CategoryList)
	return
}

func InsertRowInParam(p *Param) (int64, error) {
	err := InitDatabase(DbPath)
	if err != nil {
		log.Fatal("error initializing DB connection: ", err)
	}
	result, err := db.Exec(` 
		INSERT OR REPLACE INTO param (gofiID, paramName, paramJSONstringData, paramInfo)
		VALUES (?,?,?,?);
		`, 
		p.GofiID, p.ParamName, p.ParamJSONstringData, p.ParamInfo,
	)
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	defer db.Close()
	return id, nil
}

func InsertRowInFinanceTracker(ft *FinanceTracker) (int64, error) {
	err := InitDatabase(DbPath)
	if err != nil {
		log.Fatal("error initializing DB connection: ", err)
	}
	result, err := db.ExecContext(
		context.Background(),`
		INSERT INTO financeTracker (gofiID, account, product, priceIntx100, category)
		VALUES (?,?,?,?,?);
		`, 
		ft.GofiID, ft.Account, ft.Product, ft.PriceIntx100, ft.Category,
	)
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	defer db.Close()
	return id, nil
}

func RunSQLite() {
	if len(DbPath) == 0 {
		log.Fatal("specify the SQLITE_DB_FILENAME environment variable")
	}
	fmt.Println(DbPath)

	err := InitDatabase(DbPath)
	if err != nil {
		log.Fatal("error initializing DB connection: ", err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		log.Fatal("error initializing DB connection: ping error: ", err)
	}
	fmt.Println("database initialized..")

	fmt.Println("--------------financeTrackers----------------")

	result, err := db.Exec(`INSERT INTO financeTracker(gofiID, account, product, priceIntx100, category) SELECT 'sheet1', 'cb', 'test', 16.52, 'catego';`,)
	id, err := result.LastInsertId()
	if err != nil {
		fmt.Printf("err: %v\n", err)
		fmt.Printf("id : %v\n", id)
	}




	var financeTrackers []FinanceTracker
	rows, err := db.QueryContext(
		context.Background(),
		// `SELECT id, day, account, product, priceIntx100, category, commentFloat, CommentString
		// FROM financeTracker ORDER BY id DESC LIMIT 2;`)
		`
			SELECT id, gofiID, year, month, day, account, product, priceIntx100, category, commentFloat, CommentString, checked, dateChecked, sentToSheets
			FROM financeTracker 
			ORDER BY id DESC 
			LIMIT 3
		;`)
		// `SELECT 999,0,0,0,'cb', 'test', 1.52, 'catego',0,'',0,'';`)
	defer rows.Close()
	for rows.Next() {
		var financeTracker FinanceTracker 

		// err = rows.Scan(&financeTracker.ID, &financeTracker.Day, &financeTracker.Account, &financeTracker.Product, &financeTracker.PriceIntx100, &financeTracker.Category,
		// 	&financeTracker.CommentFloat, &financeTracker.CommentString)
		err = rows.Scan(&financeTracker.ID, &financeTracker.GofiID, &financeTracker.Year, &financeTracker.Month, &financeTracker.Day, 
			&financeTracker.Account, &financeTracker.Product, &financeTracker.PriceIntx100, &financeTracker.Category, 
			&financeTracker.CommentFloat, &financeTracker.CommentString, &financeTracker.Checked, &financeTracker.DateChecked, &financeTracker.SentToSheets)
		if err != nil {
			fmt.Printf("err: %v\n", err)
		}
	
		fmt.Printf("line: %v %v %v %v %v\n", financeTracker.ID, financeTracker.Account, financeTracker.Product, financeTracker.PriceIntx100, financeTracker.Category)
		financeTrackers = append(financeTrackers, financeTracker)
	}
	fmt.Printf("financeTrackers nb rows returned: %v\n", len(financeTrackers))
	fmt.Printf("%#v\n", financeTrackers[0])
	fmt.Printf("%#v\n", financeTrackers)

}