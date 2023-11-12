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
	"mime/multipart"
	// "time"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func CheckIfIdExists(gofiID int) {
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

func CreateUser(user User) (int64, string, error) {
	db, err := sql.Open("sqlite", DbPath)
	if err != nil { return 0, "error opening DB file", err }
	defer db.Close()
	defer db.Exec("PRAGMA optimize;") // to run just before closing each database connection.
	defer fmt.Println("defer : optimize then close DB")

	result, err := db.Exec(`
		INSERT INTO user (email, pwHash, dateCreated)
		VALUES (?,?,?);
		`, 
		user.Email, user.PwHash, user.DateCreated,
	)
	if err != nil { return 0, "error inserting row in DB", err }
	id, err := result.LastInsertId()
	if err != nil { return 0, "error to get last inserted id in DB", err }
	return id, "", nil
}

func CheckUserLogin(user User) (int, string, error) {
	db, err := sql.Open("sqlite", DbPath)
	if err != nil { return 0, "error opening DB file", err }
	defer db.Close()
	defer db.Exec("PRAGMA optimize;") // to run just before closing each database connection.
	defer fmt.Println("defer : optimize then close DB")

	q := ` 
		SELECT gofiID
		FROM user
		WHERE email = ?
			AND pwHash = ?;
	`
	rows, err := db.Query(q, user.Email, user.PwHash)
	if err != nil { return 0, "error querying DB", err }

	rows.Next()
	var gofiID int = 0
	if err := rows.Scan(&gofiID); err != nil { return 0, "error on SELECT gofiID", err }
	if (gofiID > 0) {
		_, err := db.Exec(`
			UPDATE user 
			SET numberOfRequests = numberOfRequests + 1,
				idleTimeout = strftime('%Y-%m-%dT%H:%M:%SZ', DATETIME('now', 'utc', idleDateModifier)),
				absoluteTimeout = strftime('%Y-%m-%dT%H:%M:%SZ', DATETIME('now', 'utc', absoluteDateModifier)),
				lastLoginTime = strftime('%Y-%m-%dT%H:%M:%SZ', DATETIME('now', 'utc')), 
				lastActivityTime = strftime('%Y-%m-%dT%H:%M:%SZ', DATETIME('now', 'utc')), 
				sessionID = ?, lastActivityIPaddress = ?, lastActivityUserAgent = ?, lastActivityAcceptLanguage = ?
			WHERE gofiID = ?;
			`, 
			user.SessionID, user.LastActivityIPaddress, user.LastActivityUserAgent, user.LastActivityAcceptLanguage,
			gofiID,
		)
		if err != nil { return gofiID, "error on UPDATE after login", err }
	}
	
	return gofiID, "", nil
}

func ForceNewLogin(gofiID int) (bool, string, error) {
	db, err := sql.Open("sqlite", DbPath)
	if err != nil { return false, "error opening DB file", err }
	defer db.Close()
	defer db.Exec("PRAGMA optimize;") // to run just before closing each database connection.
	defer fmt.Println("defer : optimize then close DB")

	if (gofiID > 0) {
		_, err := db.Exec(`
			UPDATE user 
			SET numberOfRequests = numberOfRequests + 1,
				idleTimeout = '1999-12-31T00:01:01Z',
				absoluteTimeout = '1999-12-31T00:01:01Z',
				lastActivityTime = strftime('%Y-%m-%dT%H:%M:%SZ', DATETIME('now', 'utc')), 
				sessionID = 'logged-out'
			WHERE gofiID = ?;
			`, 
			gofiID,
		)
		if err != nil { return false, "error on UPDATE after login", err }
	}
	return true, "", nil
}

func GetGofiID(sessionID string) (int, string, error) {
	db, err := sql.Open("sqlite", DbPath)
	if err != nil { return 0, "error opening DB file", err }
	defer db.Close()
	defer db.Exec("PRAGMA optimize;") // to run just before closing each database connection.
	defer fmt.Println("defer : optimize then close DB")

	q := ` 
		SELECT gofiID
		FROM user
		WHERE sessionID = ?;
	`
	rows, err := db.Query(q, sessionID)
	if err != nil { return 0, "error querying DB", err }

	rows.Next()
	var gofiID int = 0
	if err := rows.Scan(&gofiID); err != nil { return 0, "error on SELECT gofiID", err }
	if (gofiID > 0) { return gofiID, "", nil } else { return 0, "error no gofiID found from sessionID cookie", err }
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
		return
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

func GetLastRowsInFinanceTracker(gofiID int) []FinanceTracker {
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
		SELECT year, month, day, 
			account, product, priceIntx100, category
		FROM financeTracker
		WHERE gofiID = ?
		ORDER BY id DESC
		LIMIT 5;
	`
	rows, err := db.Query(q, gofiID)

	for rows.Next() {
		var ft FinanceTracker
		var successfull bool
		var unsuccessfullReason string
		if err := rows.Scan(&ft.Year, &ft.Month, &ft.Day, &ft.Account, &ft.Product, &ft.PriceIntx100, &ft.Category); err != nil {
			log.Fatal(err)
		}
		ft.FormPriceStr2Decimals = ConvertPriceIntToStr(ft.PriceIntx100)
		ft.Date, successfull, unsuccessfullReason = ConvertDateIntToStr(ft.Year, ft.Month, ft.Day, "FR", "/")
		if !successfull {ft.Date = "ERROR " + unsuccessfullReason}

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
		INSERT INTO financeTracker (gofiID, year, month, day, account, product, priceIntx100, category)
		VALUES (?,?,?,?,?,?,?,?);
		`, 
		ft.GofiID, ft.Year, ft.Month, ft.Day, ft.Account, ft.Product, ft.PriceIntx100, ft.Category,
	)
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func ExportCSV(gofiID int, csvSeparator rune, csvDecimalDelimiter string, dateFormat string, dateSeparator string) {
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
		SELECT id, year, month, day,
			account, product, priceIntx100, category, 
			commentInt, commentString, checked, dateChecked, sentToSheets
		FROM financeTracker
		WHERE gofiID = ?
		ORDER BY id
		LIMIT 10000;
	`
	rows, err := db.Query(q, gofiID)

	file, err := os.Create(FilePath("gofi-" + strconv.Itoa(gofiID) + ".csv"))
	defer file.Close()
	w := csv.NewWriter(file)
	w.Comma = csvSeparator //french CSV file = ;
    defer w.Flush()

	//write csv headers
	row := []string{"ID", "Date",
		"Account", "Product", "PriceStr", "Category", 
		"CommentInt", "CommentString", "Checked", "DateChecked", "SentToSheets"}
	if err := w.Write(row); err != nil {
		fmt.Printf("row error: %v\n", row)
		log.Fatalln("error writing record to file", err)
	}
	for rows.Next() {
		var ft FinanceTracker
		var successfull bool
		var unsuccessfullReason string
		if err := rows.Scan(
				&ft.ID, &ft.Year, &ft.Month, &ft.Day,
				&ft.Account, &ft.Product, &ft.PriceIntx100, &ft.Category,
				&ft.CommentInt, &ft.CommentString, &ft.Checked, &ft.DateChecked, &ft.SentToSheets,
			); err != nil {
			log.Fatal(err)
		}
		ft.FormPriceStr2Decimals = strings.Replace(ConvertPriceIntToStr(ft.PriceIntx100), ".", csvDecimalDelimiter, 1) //replace . to , for french CSV files
		ft.Date, successfull, unsuccessfullReason = ConvertDateIntToStr(ft.Year, ft.Month, ft.Day, dateFormat, dateSeparator)
		if !successfull {ft.Date = "ERROR " + unsuccessfullReason}

        row = []string{strconv.Itoa(ft.ID), ft.Date, 
			ft.Account, ft.Product, ft.FormPriceStr2Decimals, ft.Category, 
			strconv.Itoa(ft.CommentInt), ft.CommentString, strconv.FormatBool(ft.Checked), ft.DateChecked, strconv.FormatBool(ft.SentToSheets)}
        if err := w.Write(row); err != nil {
			fmt.Printf("row error: %v\n", row)
            log.Fatalln("error writing record to file", err)
        }

	}
}


func ImportCSV(gofiID int, csvSeparator rune, csvDecimalDelimiter string, dateFormat string, dateSeparator string, csvFile *multipart.FileHeader) string {
	/* take all data from the csv and put it in the DB with a specific gofiID
		1. rows without ID are new ones (INSERT)
		2. rows with ID are existing ones (UPDATE)
		3. read csv (from line 2)
		4. write row by row in DB
	*/
	var stringList string
	stringList += "traitement fichier avec ID : " + strconv.Itoa(gofiID) + "\n"
	db, err := sql.Open("sqlite", DbPath)
	if err != nil {
		log.Fatal("error opening DB file: ", err)
		stringList += "erreur de base de données, merci de réessayer plus tard."
		return stringList
	}
	defer db.Close()
	defer db.Exec("PRAGMA optimize;") // to run just before closing each database connection.
	defer fmt.Println("defer : optimize then close DB")

	if (csvFile.Size > 200000) {
		stringList += "Fichier trop lourd: " + strconv.FormatInt(csvFile.Size, 10)
		stringList += " octets.\nLa limite actuelle est fixée à 200 000 octets par fichier.\nMerci de découper le fichier et faire plusieurs traitements."
		return stringList
	}
	file, err := csvFile.Open() // For read access.
	if err != nil {
		log.Fatal("Unable to read input file : " + csvFile.Filename, err)
		stringList += "erreur d'ouverture du fichier csv, merci de vérifier le format."
		return stringList
	}
	defer file.Close() // this needs to be after the err check
	r := csv.NewReader(file)
	r.Comma = csvSeparator //french CSV file = ;
    rows, err := r.ReadAll()
    if err != nil {
        log.Fatal("Unable to parse file as CSV for : " + csvFile.Filename, err)
		stringList += "erreur de lecture d'au moins 1 ligne dans le fichier csv, merci de vérifier le contenu du fichier."
		return stringList
    }

	var ft FinanceTracker
	var lineInfo, unsuccessfullReason string
	var successfull bool
	ft.GofiID = gofiID
	stringList += "ID;Date;CommentInt;Checked;SentToSheets;NewID;Updated;\n"
	for index, row := range rows {
		if (index == 0) {continue} //skip headers
		lineInfo = ""
		ft.ID, err = strconv.Atoi(row[0])
		if err != nil { // Always check errors even if they should not happen.
			ft.ID = 0
			lineInfo += "default 0;"
		} else { lineInfo += row[0] + ";" }

		ft.Year, ft.Month, ft.Day, successfull, unsuccessfullReason = ConvertDateStrToInt(row[1], dateFormat, dateSeparator)
		if !successfull {
			lineInfo += "error " + unsuccessfullReason + ";;;;;false;"
			stringList += lineInfo + "\n"
			continue //skip this row because wrong date format
		}

		ft.Account = row[2] 
		ft.Product = row[3]
		ft.FormPriceStr2Decimals = row[4]
		safeInteger, _ := strconv.Atoi(strings.Replace(ft.FormPriceStr2Decimals, csvDecimalDelimiter, "", 1))
		ft.PriceIntx100 = safeInteger

		ft.Category = row[5]
		ft.CommentInt, err = strconv.Atoi(row[6])
		if err != nil {
			ft.CommentInt = 0
			lineInfo += "default 0;"
		} else { lineInfo += ";" }
		ft.CommentString = row[7]
		ft.Checked, err = strconv.ParseBool(row[8])
		if err != nil {
			ft.Checked = false
			lineInfo += "default 0;"
		} else { lineInfo += ";" }
		ft.DateChecked = row[9]
		ft.SentToSheets, err = strconv.ParseBool(row[10])
		if err != nil {
			ft.SentToSheets = false
			lineInfo += "default 0;"
		} else { lineInfo += ";" }

		if (ft.ID == 0) {
			// INSERT
			exec, err := db.Exec(`
				INSERT INTO financeTracker (gofiID, year, month, day, account, product, priceIntx100, category,
					commentInt, commentString, checked, dateChecked, sentToSheets)
				VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?);
				`, 
				ft.GofiID, ft.Year, ft.Month, ft.Day, ft.Account, ft.Product, ft.PriceIntx100, ft.Category,
				ft.CommentInt, ft.CommentString, ft.Checked, ft.DateChecked, ft.SentToSheets,
			)
			if err != nil {lineInfo += "error1;false;"} else {
				rowID, err := exec.LastInsertId()
				if err != nil {lineInfo += "error2;false;"} else {lineInfo += strconv.FormatInt(rowID, 10) + ";true;"}
			}
		} else {
			// UPDATE
			_, err := db.Exec(`
				UPDATE financeTracker 
				SET year = ?, month = ?, day = ?, account = ?, product = ?, priceIntx100 = ?, category = ?,
					commentInt = ?, commentString = ?, checked = ?, dateChecked = ?, sentToSheets = ?
				WHERE ID = ?
					AND gofiID = ?;
				`, 
				ft.Year, ft.Month, ft.Day, ft.Account, ft.Product, ft.PriceIntx100, ft.Category,
				ft.CommentInt, ft.CommentString, ft.Checked, ft.DateChecked, ft.SentToSheets,
				ft.ID, ft.GofiID,
			)
			if err != nil {lineInfo += "error3;false;"} else {lineInfo += ";true;"}
		}
		stringList += lineInfo + "\n"
	}
	return stringList
}
