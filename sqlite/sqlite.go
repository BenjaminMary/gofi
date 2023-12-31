package sqlite

import (
	"context"
	"database/sql"

	"fmt"
	"log"
	"strings"
	"strconv"

	"encoding/csv"
	"os"
	"mime/multipart"
	"time"

	_ "modernc.org/sqlite"
)

func OpenDbCon() *sql.DB {
	db, err := sql.Open("sqlite", DbPath)
	if err != nil {
		log.Fatal("error opening DB file: ", err)
	}
	db.SetMaxIdleConns(1) //default 2
	db.SetMaxOpenConns(3) //default 0 = infinite
	return db
}

// func WalCheckpointB() {
// 	sql.Open("sqlite", "file:///gofi.db?_pragma=foreign_keys(1)&_time_format=sqlite")
// 	// sqlite.Open("file:///tmp/mydata.sqlite?_pragma=foreign_keys(1)&_time_format=sqlite")
// }

func WalCheckpoint(ctx context.Context) int {
	db, err := sql.Open("sqlite", DbPath)
	if err != nil {
		log.Fatal("error opening DB file: ", err)
		return -1
	}
	defer db.Close()
	defer fmt.Println("defer : db.Close()")
	db.SetMaxIdleConns(1) //default 2
	db.SetMaxOpenConns(1) //default 0 = infinite

	conn, err := db.Conn(ctx)
	if err != nil {
		log.Fatal("error connecting to DB file: ", err)
		return -1
	}
	defer conn.Close() // Return the connection to the pool.
	defer fmt.Println("defer : conn.Close()")

	// fmt.Println("optimize, vacuum, checkpoint TRUNCATE then close DB")
	fmt.Println("optimize, vacuum, checkpoint TRUNCATE")
	conn.ExecContext(ctx, "PRAGMA optimize;") // to run just before closing each database connection.

	var journalMode string
	err = conn.QueryRowContext(ctx, "PRAGMA journal_mode;").Scan(&journalMode)
	if err != nil {
		log.Fatal("error PRAGMA journal_mode: ", err)
		return -1
	}
	fmt.Printf("journalMode: %v\n", journalMode)

	conn.ExecContext(ctx, "VACUUM;") // to run just before closing each database connection.

	var busyTimeout string
	err = conn.QueryRowContext(ctx, "PRAGMA busy_timeout;").Scan(&busyTimeout)
	if err != nil {
		log.Fatal("error PRAGMA busyTimeout 1: ", err)
		return -1
	}
	//fmt.Printf("busyTimeout 1: %v\n", busyTimeout)
	err = conn.QueryRowContext(ctx, "PRAGMA busy_timeout = 2000;").Scan(&busyTimeout)
	if err != nil {
		log.Fatal("error PRAGMA busyTimeout 2: ", err)
		return -1
	}
	//fmt.Printf("busyTimeout 2: %v\n", busyTimeout)

	db.SetConnMaxIdleTime(100 * time.Millisecond)
	db.SetConnMaxLifetime(100 * time.Millisecond)
	time.Sleep(3 * time.Second)

	//stats := db.Stats()
	//fmt.Printf("stats: %#v\n", stats)

	conn.ExecContext(ctx, "COMMIT;")
	conn.Close()
	conn, err = db.Conn(ctx)
	if err != nil {
		log.Fatal("error connecting to DB file: ", err)
		return -1
	}

	// wal_checkpoint doc: https://www.sqlite.org/pragma.html#pragma_wal_checkpoint
	// checkpointReturn = 0 if OK, pagestoWal AND pagesFromWalToDb -1 if not in WAL mode
	var checkpointReturn, pagestoWal, pagesFromWalToDb int
	err = conn.QueryRowContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE);").Scan(&checkpointReturn, &pagestoWal, &pagesFromWalToDb)
	if err != nil {
		log.Fatal("error PRAGMA wal_checkpoint(TRUNCATE): ", err)
		return -1
	}
	//fmt.Printf("checkpointReturn: %v\n", strconv.Itoa(checkpointReturn))
	//fmt.Printf("pagestoWal: %v\n", strconv.Itoa(pagestoWal))
	//fmt.Printf("pagesFromWalToDb: %v\n", strconv.Itoa(pagesFromWalToDb))
	if checkpointReturn == 1 {
		// conn.Close()
		return 1
	}
	conn.Close()
	db.Close()
	time.Sleep(1 * time.Second)

	return checkpointReturn
}

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
	rows.Close()
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

func UpdateSessionID(ctx context.Context, db *sql.DB, gofiID int, sessionID string) (string, error) {
	_, err := db.ExecContext(ctx, `
		UPDATE user 
		SET idleTimeout = strftime('%Y-%m-%dT%H:%M:%SZ', DATETIME('now', 'utc', idleDateModifier)),
			lastActivityTime = strftime('%Y-%m-%dT%H:%M:%SZ', DATETIME('now', 'utc')), 
			sessionID = ?
			--, lastActivityIPaddress = ?, lastActivityUserAgent = ?, lastActivityAcceptLanguage = ?
		WHERE gofiID = ?;
		`, 
		sessionID, //user.LastActivityIPaddress, user.LastActivityUserAgent, user.LastActivityAcceptLanguage,
		gofiID,
	)
	if err != nil { return "error on UPDATE for UpdateSessionID", err }

	return "", nil
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

func GetGofiID(ctx context.Context, db *sql.DB, sessionID string) (int, string, string, error) {
	q := ` 
		SELECT gofiID, email, idleTimeout, absoluteTimeout, strftime('%Y-%m-%dT%H:%M:%SZ', DATETIME('now', 'utc')) AS currentTimeUTC
		FROM user
		WHERE sessionID = ?;
	`
	rows, err := db.QueryContext(ctx, q, sessionID)
	if err != nil { return 0, "", "error querying DB", err }

	rows.Next()
	var gofiID int = 0
	var email, idleTimeout, absoluteTimeout, currentTimeUTC string
	if err := rows.Scan(&gofiID,&email,&idleTimeout,&absoluteTimeout,&currentTimeUTC); err != nil { return 0, "", "error on SELECT gofiID", err }
	rows.Close()

	timeCurrentTimeUTC, err := time.Parse(time.RFC3339, currentTimeUTC)
	// fmt.Printf("timeCurrentTimeUTC: %v\n", timeCurrentTimeUTC)
	if err != nil { return -1, "", "error parsing currentTimeUTC, force new login", err }

	timeAbsoluteTimeout, err := time.Parse(time.RFC3339, absoluteTimeout)
	// fmt.Printf("timeAbsoluteTimeout: %v\n", timeAbsoluteTimeout)
	if err != nil { return -1, "", "error parsing absoluteTimeout, force new login", err }
	differenceAbsolute := timeCurrentTimeUTC.Sub(timeAbsoluteTimeout)
	// fmt.Printf("differenceAbsolute: %v\n", differenceAbsolute)
	if (differenceAbsolute > 0) { return -1, "", "absoluteTimeout, force new login", nil }

	timeIdleTimeout, err := time.Parse(time.RFC3339, idleTimeout)
	// fmt.Printf("timeIdleTimeout: %v\n", timeIdleTimeout)
	if err != nil { return -1, "", "error parsing idleTimeout, force new login", err }
	differenceIdle := timeCurrentTimeUTC.Sub(timeIdleTimeout)
	// fmt.Printf("differenceIdle: %v\n", differenceIdle)
	if (differenceIdle > 0) { return gofiID, email, "idleTimeout, change cookie", nil }

	if (gofiID > 0) { return gofiID, email, "", nil } else { return 0, "", "error no gofiID found from sessionID cookie", err }
}

func GetList(ctx context.Context, db *sql.DB, up *UserParams) {
	q := ` 
		SELECT paramJSONstringData
		FROM param
		WHERE gofiID = ?
			AND paramName = ?;
	`
	rows, _ := db.QueryContext(ctx, q, up.GofiID, "accountList")

	rows.Next()
	var accountList string
	if err := rows.Scan(&accountList); err != nil {
		log.Fatal(err)
		return
	}
	up.AccountListSingleString = accountList
	up.AccountList = strings.Split(accountList, ",")
	// fmt.Printf("\naccountList: %v\n", up.AccountList)
	rows.Close()

	rows, _ = db.QueryContext(ctx, q, up.GofiID, "categoryList")
	rows.Next()
	var categoryList string
	if err := rows.Scan(&categoryList); err != nil {
		log.Fatal(err)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	up.CategoryListSingleString = categoryList
	up.CategoryList = strings.Split(categoryList, ",")
	rows.Close()
	// fmt.Printf("\ncategoryList: %v\n", up.CategoryList)
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

func GetRowsInFinanceTracker(ctx context.Context, db *sql.DB, filter *FilterRows) ([]FinanceTracker, string) {
	var ftList []FinanceTracker
	var totalPriceStr2Decimals string
	var queryValues int = 0
	var err error
	if filter.Limit > 100 {filter.Limit = 100}
	//fmt.Printf("filter.WhereAccount: %#v, type:%T\n", filter.WhereAccount, filter.WhereAccount) // check default value and type
	//fmt.Printf("filter.WhereYear: %#v, type:%T\n", filter.WhereYear, filter.WhereYear) // check default value and type
	
	// start building query 
	// (golang sql package does not support dynamic sql on other things than values)
	q := ` 
		SELECT id, year, month, day, 
			account, product, priceIntx100, category, checked, dateChecked
		FROM financeTracker
		WHERE gofiID = ?
	`
	// others where on 3 fields max = 7 possibilities
	if filter.WhereAccount != "" { //1
		queryValues += 1
		fmt.Println("filter.WhereAccount is used")
		q += ` AND account = ? `
	} 
	if filter.WhereCategory != "" { //2
		queryValues += 2
		fmt.Println("filter.WhereCategory is used")
		q += ` AND category = ? `
	} 
	if filter.WhereYear != 0 { //4
		queryValues += 4
		fmt.Println("filter.WhereYear is used")
		q += ` AND year = ? `
	}
	if filter.WhereMonth != 0 { // month used alone
		switch filter.WhereMonth {
			case  1: q += ` AND month =  1 `
			case  2: q += ` AND month =  2 `
			case  3: q += ` AND month =  3 `
			case  4: q += ` AND month =  4 `
			case  5: q += ` AND month =  5 `
			case  6: q += ` AND month =  6 `
			case  7: q += ` AND month =  7 `
			case  8: q += ` AND month =  8 `
			case  9: q += ` AND month =  9 `
			case 10: q += ` AND month = 10 `
			case 11: q += ` AND month = 11 `
			case 12: q += ` AND month = 12 `
			default: q += ` `
		}
		fmt.Println("filter.WhereMonth is used")
	} 
	if filter.WhereChecked != 0 { // checked used alone
		if filter.WhereChecked == 2 {q += ` AND checked = 0 `} else {q += ` AND checked = 1 `}
		fmt.Println("filter.WhereChecked is used", filter.WhereChecked)
	} 

	// order by column
	q += ` ORDER BY `
	switch filter.OrderBy {
		case "id":
			q += ` id `
		case "date":
			q += ` year*10000 + month*100 + day `
		case "price":
			q += ` priceIntx100 `
		default:
			q += ` id `
	}
	// order by type
	if (filter.OrderByType == "DESC") {q += ` DESC `} else {q += ` ASC `}

	// finally, add limit
	q += ` LIMIT ?;`
	//fmt.Printf("q: %v\n", q)
	// end building query

	var rows *sql.Rows
	switch queryValues {
		case 0:
			rows, err = db.QueryContext(ctx, q, filter.GofiID, filter.Limit)
		case 1:
			rows, err = db.QueryContext(ctx, q, filter.GofiID, filter.WhereAccount, filter.Limit)
		case 2:
			rows, err = db.QueryContext(ctx, q, filter.GofiID, filter.WhereCategory, filter.Limit)
		case 3:
			rows, err = db.QueryContext(ctx, q, filter.GofiID, filter.WhereAccount, filter.WhereCategory, filter.Limit)
		case 4:
			rows, err = db.QueryContext(ctx, q, filter.GofiID, filter.WhereYear, filter.Limit)
		case 5:
			rows, err = db.QueryContext(ctx, q, filter.GofiID, filter.WhereAccount, filter.WhereYear, filter.Limit)
		case 6:
			rows, err = db.QueryContext(ctx, q, filter.GofiID, filter.WhereCategory, filter.WhereYear, filter.Limit)
		case 7:
			rows, err = db.QueryContext(ctx, q, filter.GofiID, filter.WhereAccount, filter.WhereCategory, filter.WhereYear, filter.Limit)
		default:
			rows, err = db.QueryContext(ctx, q, filter.GofiID, filter.Limit)
	}

	if err != nil {
		log.Fatal("error on DB query: ", err)
	}
	var totalPriceIntx100 int = 0
	for rows.Next() {
		var ft FinanceTracker
		var successfull bool
		var unsuccessfullReason string
		if err := rows.Scan(&ft.ID, &ft.Year, &ft.Month, &ft.Day, &ft.Account, &ft.Product, &ft.PriceIntx100, 
			&ft.Category, &ft.Checked, &ft.DateChecked); err != nil {
			log.Fatal(err)
		}
		ft.FormPriceStr2Decimals = ConvertPriceIntToStr(ft.PriceIntx100)
		totalPriceIntx100 += ft.PriceIntx100
		ft.Date, successfull, unsuccessfullReason = ConvertDateIntToStr(ft.Year, ft.Month, ft.Day, "FR", "/")
		if !successfull {ft.Date = "ERROR " + unsuccessfullReason}

		// fmt.Printf("ft: %#v\n", ft)
		ftList = append(ftList, ft)
	}
	// fmt.Printf("totalPriceIntx100: %v, inStr: %v\n", totalPriceIntx100, totalPriceStr2Decimals)
	totalPriceStr2Decimals = ConvertPriceIntToStr(totalPriceIntx100)
	rows.Close()
	return ftList, totalPriceStr2Decimals
}

func InsertRowInFinanceTracker(ctx context.Context, db *sql.DB, ft *FinanceTracker) (int64, error) {
	result, err := db.ExecContext(ctx, `
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
			commentInt, commentString, checked, dateChecked, exported
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
	row := []string{"𫝀é ꮖꭰ", "Date",
		"Account", "Product", "PriceStr", "Category", 
		"CommentInt", "CommentString", "Checked", "DateChecked", "Exported", 
		""} //keeping an empty column at the end will handle the LF and CRLF cases
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
				&ft.CommentInt, &ft.CommentString, &ft.Checked, &ft.DateChecked, &ft.Exported,
			); err != nil {
			log.Fatal(err)
		}
		ft.FormPriceStr2Decimals = strings.Replace(ConvertPriceIntToStr(ft.PriceIntx100), ".", csvDecimalDelimiter, 1) //replace . to , for french CSV files
		ft.Date, successfull, unsuccessfullReason = ConvertDateIntToStr(ft.Year, ft.Month, ft.Day, dateFormat, dateSeparator)
		if !successfull {ft.Date = "ERROR " + unsuccessfullReason}

        row = []string{strconv.Itoa(ft.ID), ft.Date, 
			ft.Account, ft.Product, ft.FormPriceStr2Decimals, ft.Category, 
			strconv.Itoa(ft.CommentInt), ft.CommentString, strconv.FormatBool(ft.Checked), ft.DateChecked, strconv.FormatBool(ft.Exported), ""}
        if err := w.Write(row); err != nil {
			fmt.Printf("row error: %v\n", row)
            log.Fatalln("error writing record to file", err)
        }

	}
	rows.Close()
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

	if (csvFile.Size > 1000000) {
		stringList += "Fichier trop lourd: " + strconv.FormatInt(csvFile.Size, 10)
		stringList += " octets.\nLa limite actuelle est fixée à 1 000 000 octets par fichier.\nMerci de découper le fichier et faire plusieurs traitements."
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
	var lineInfo, unsuccessfullReason, controlEncoding string
	var successfull bool
	ft.GofiID = gofiID
	stringList += "𫝀é ꮖꭰ;Date;CommentInt;Checked;exported;NewID;Updated;\n"
	for index, row := range rows {
		if (index == 0) { //control UTF-8 on headers
			controlEncoding = row[0]
			if (controlEncoding == "𫝀é ꮖꭰ"){
				continue //skip the row
			} else {
				stringList = /* "EN: CANCELED IMPORT.\n" +
					"ERROR on encoding format of the file.\n" +
					"The system only accept UTF-8.\n\n" +
					"INFO: specifics characters are in the 1st column of the header and must be kept.\n" +
					"1st column = '𫝀é ꮖꭰ'\n\n" + */
					"IMPORTATION ANNULEE.\n" +
					"ERREUR sur le format d'encodage du fichier.\n" +
					"Le système accepte uniquement du UTF-8.\n\n" +
					"INFO: des caractères spécifiques sont présents en en-tête de la 1ere colonne et doivent être gardés.\n" +
					"1ere colonne = '𫝀é ꮖꭰ'\n" +
					"Un exemple de données d'import valide est disponible plus bas sur cette page."
				break //stop
			}
		}
		lineInfo = ""
		ft.ID, err = strconv.Atoi(row[0])
		if err != nil { // Always check errors even if they should not happen.
			ft.ID = 0
			lineInfo += "INSERT;"
		} else { lineInfo += "UPDATE " + row[0] + ";" }

		ft.Year, ft.Month, ft.Day, successfull, unsuccessfullReason = ConvertDateStrToInt(row[1], dateFormat, dateSeparator)
		if !successfull {
			lineInfo += "error " + unsuccessfullReason + ";;;;;false;"
			stringList += lineInfo + "\n"
			continue //skip this row because wrong date format
		}

		ft.Account = row[2] 
		ft.Product = row[3]
		ft.FormPriceStr2Decimals = row[4]
		ft.PriceIntx100 = ConvertPriceStrToInt(ft.FormPriceStr2Decimals, csvDecimalDelimiter)

		ft.Category = row[5]
		ft.CommentInt, err = strconv.Atoi(row[6])
		if err != nil {
			ft.CommentInt = 0
			lineInfo += "comment i 0;"
		} else { lineInfo += ";" }
		ft.CommentString = row[7]

		// Checked
		ft.Checked, err = strconv.ParseBool(row[8])
		if err != nil {
			ft.Checked = false
			lineInfo += "checked 0;"
		} else { lineInfo += ";" }

		// DateChecked
		ft.DateChecked = "9999-12-31"
		if len(row[9]) == 10 {
			yearInt, monthInt, dayInt, successfull, _ := ConvertDateStrToInt(row[9], dateFormat, dateSeparator)
			// fmt.Println("---------------")
			// fmt.Printf("ft.DateChecked: %v\n", ft.DateChecked)
			// fmt.Printf("yearInt %v, monthInt %v, dayInt %v, successfull %v, unsuccessfullReason %v\n", yearInt, monthInt, dayInt, successfull, unsuccessfullReason)
			if successfull {
				dateForDB, successfull, _ := ConvertDateIntToStr(yearInt, monthInt, dayInt, "EN", "-") //force YYYY-MM-DD inside DB
				//fmt.Printf("dateForDB %v, successfull %v, unsuccessfullReason %v\n", dateForDB, successfull, unsuccessfullReason)
				if successfull {ft.DateChecked = dateForDB}
			}	
		}
		ft.Exported, err = strconv.ParseBool(row[10])
		if err != nil {
			ft.Exported = false
			lineInfo += "sent 0;"
		} else { lineInfo += ";" }

		if (ft.ID == 0) {
			// INSERT
			exec, err := db.Exec(`
				INSERT INTO financeTracker (gofiID, year, month, day, account, product, priceIntx100, category,
					commentInt, commentString, checked, dateChecked, exported)
				VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?);
				`, 
				ft.GofiID, ft.Year, ft.Month, ft.Day, ft.Account, ft.Product, ft.PriceIntx100, ft.Category,
				ft.CommentInt, ft.CommentString, ft.Checked, ft.DateChecked, ft.Exported,
			)
			if err != nil {lineInfo += "error1;false;"} else {
				rowID, err := exec.LastInsertId()
				if err != nil {lineInfo += "error2;false;"} else {lineInfo += strconv.FormatInt(rowID, 10) + ";true;"}
			}
		} else {
			// UPDATE
			result, err := db.Exec(`
				UPDATE financeTracker 
				SET year = ?, month = ?, day = ?, account = ?, product = ?, priceIntx100 = ?, category = ?,
					commentInt = ?, commentString = ?, checked = ?, dateChecked = ?, exported = ?
				WHERE ID = ?
					AND gofiID = ?;
				`, 
				ft.Year, ft.Month, ft.Day, ft.Account, ft.Product, ft.PriceIntx100, ft.Category,
				ft.CommentInt, ft.CommentString, ft.Checked, ft.DateChecked, ft.Exported,
				ft.ID, ft.GofiID,
			)
			if err != nil {lineInfo += "error3;false;"} else {
				rows, err := result.RowsAffected()
				if err != nil {lineInfo += "error4;false;"} else {
					if rows != 1 {lineInfo += "unknown ID;false;"} else {lineInfo += ";true;"}
				}
			}
		}
		stringList += lineInfo + "\n"
	}
	return stringList
}
