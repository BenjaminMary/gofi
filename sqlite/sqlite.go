package sqlite

import (
	// "context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"strconv"

	"encoding/csv"
	"os"

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

func GetLastRowsInFinanceTracker(gofiID string) []FinanceTracker {
	var ftList []FinanceTracker
	db, err := sql.Open("sqlite", DbPath)
	if err != nil {
		log.Fatal("error opening DB file: ", err)
		return ftList
	}
	defer db.Close()
	defer db.Exec("PRAGMA optimize;") // to run just before closing each database connection.
	defer fmt.Println("defer : optimize then close DB")
	
	q := ` 
		SELECT year || '-' || month || '-' || day AS date, 
			account, product, priceIntx100, category
		FROM financeTracker
		WHERE gofiID = ?
		ORDER BY id DESC
		LIMIT 5;
	`
	rows, err := db.Query(q, gofiID)

	for rows.Next() {
		var ft FinanceTracker
		if err := rows.Scan(&ft.Date, &ft.Account, &ft.Product, &ft.PriceIntx100, &ft.Category); err != nil {
			log.Fatal(err)
		}
		ft.FormPriceStr2Decimals = ConvertPriceIntToStr(ft.PriceIntx100)
		// fmt.Printf("ft: %#v\n", ft)
		ftList = append(ftList, ft)
	}
	return ftList
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

func ExportCSV(gofiID string, csvSeparator rune, csvDecimalDelimiter string) {
	/* take all data from the DB for a specific gofiID and put it in a csv file 
		1. read database with gofiID
		2. write row by row in a csv (include headers)
	*/
	db, err := sql.Open("sqlite", DbPath)
	if err != nil {
		log.Fatal("error opening DB file: ", err)
		return
	}
	defer db.Close()
	defer db.Exec("PRAGMA optimize;") // to run just before closing each database connection.
	defer fmt.Println("defer : optimize then close DB")
	
	q := ` 
		SELECT id, gofiID, year || '-' || month || '-' || day AS date,
			account, product, priceIntx100, category, 
			commentInt, commentString, checked, dateChecked, sentToSheets
		FROM financeTracker
		WHERE gofiID = ?
		ORDER BY id
		LIMIT 10000;
	`
	rows, err := db.Query(q, gofiID)

	file, err := os.Create(FilePath("gofi-" + gofiID + ".csv"))
	defer file.Close()
	w := csv.NewWriter(file)
	w.Comma = csvSeparator //french CSV file = ;
    defer w.Flush()

	//write csv headers
	row := []string{"ID", "GofiID", "Date",
		"Account", "Product", "PriceStr", "Category", 
		"CommentInt", "CommentString", "Checked", "DateChecked", "SentToSheets"}
	if err := w.Write(row); err != nil {
		fmt.Printf("row error: %v\n", row)
		log.Fatalln("error writing record to file", err)
	}
	for rows.Next() {
		var ft FinanceTracker
		if err := rows.Scan(
				&ft.ID, &ft.GofiID, &ft.Date,
				&ft.Account, &ft.Product, &ft.PriceIntx100, &ft.Category,
				&ft.CommentInt, &ft.CommentString, &ft.Checked, &ft.DateChecked, &ft.SentToSheets,
			); err != nil {
			log.Fatal(err)
		}
		ft.FormPriceStr2Decimals = strings.Replace(ConvertPriceIntToStr(ft.PriceIntx100), ".", csvDecimalDelimiter, 1) //replace . to , for french CSV files
	
        row = []string{strconv.Itoa(ft.ID), ft.GofiID, ft.Date, 
			ft.Account, ft.Product, ft.FormPriceStr2Decimals, ft.Category, 
			strconv.Itoa(ft.CommentInt), ft.CommentString, strconv.FormatBool(ft.Checked), ft.DateChecked, strconv.FormatBool(ft.SentToSheets)}
        if err := w.Write(row); err != nil {
			fmt.Printf("row error: %v\n", row)
            log.Fatalln("error writing record to file", err)
        }

	}
}
