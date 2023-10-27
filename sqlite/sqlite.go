package sqlite

import (
	// "context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func CheckIfIdExists(gofiID string) {
	//if new ID, create default params
	var nbRows int = 0

	db, err := sql.Open("sqlite", DbPath)
	if err != nil {
		log.Fatal("error opening DB file: ", err)
		return
	}
	defer db.Close()
	defer db.Exec("PRAGMA optimize;") // to run just before closing each database connection.
	defer fmt.Println("defer : optimize then close DB")

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
	}

	qB := ` 
		SELECT COUNT(1)
		FROM param
		WHERE gofiID = ?
			AND paramName = 'categoryList';
	`
	err = db.QueryRow(qB, gofiID).Scan(&nbRows)
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
	}

	return
}

func GetList(ft *FinanceTracker) {
	db, err := sql.Open("sqlite", DbPath)
	if err != nil {
		log.Fatal("error opening DB file: ", err)
		return
	}
	defer db.Close()
	defer db.Exec("PRAGMA optimize;") // to run just before closing each database connection.
	defer fmt.Println("defer : optimize then close DB")
	
	q := ` 
		SELECT paramJSONstringData
		FROM param
		WHERE gofiID = ?
			AND paramName = ?;
	`
	rows, err := db.Query(q, ft.GofiID, "accountList")

	rows.Next()
	var accountList string
	if err := rows.Scan(&accountList); err != nil {
		log.Fatal(err)
	}
	ft.Account = accountList
	ft.AccountList = strings.Split(accountList, ",")
	// fmt.Printf("\naccountList: %v\n", ft.AccountList)

	rows, err = db.Query(q, ft.GofiID, "categoryList")
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
	db, err := sql.Open("sqlite", DbPath)
	if err != nil {
		log.Fatal("error opening DB file: ", err)
		return 0, err
	}
	defer db.Close()
	defer db.Exec("PRAGMA optimize;") // to run just before closing each database connection.
	defer fmt.Println("defer : optimize then close DB")

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
	return id, nil
}

func InsertRowInFinanceTracker(ft *FinanceTracker) (int64, error) {
	db, err := sql.Open("sqlite", DbPath)
	if err != nil {
		log.Fatal("error opening DB file: ", err)
		return 0, err
	}
	defer db.Close()
	defer db.Exec("PRAGMA optimize;") // to run just before closing each database connection.
	defer fmt.Println("defer : optimize then close DB")

	result, err := db.Exec(`
		INSERT INTO financeTracker (gofiID, account, product, priceIntx100, category)
		VALUES (?,?,?,?,?);
		`, 
		ft.GofiID, ft.Account, ft.Product, ft.PriceIntx100, ft.Category,
	)
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}
