package sqlite

import (
	"context"
	"database/sql"

	"fmt"
	"log"
	"strings"
	"errors"
	"net/http"

	"gofi/gofi/data/appdata"
)

func getOrCompleteVarsWithAnyListAndQuestionMarks(offset int, addQuestionMarks bool, nbQuestionMarks string, intListIn *[]int,
	anyListOut []any) (string, []any) {
	for i, v := range *intListIn {
        anyListOut[i+offset] = v
		if len(nbQuestionMarks) == 0 && addQuestionMarks {
			nbQuestionMarks = "?"
		} else if addQuestionMarks {
			nbQuestionMarks = nbQuestionMarks + ",?"
		}
    }
	return nbQuestionMarks, anyListOut
}

func GetRowsInFinanceTracker(ctx context.Context, db *sql.DB, filter *appdata.FilterRows) ([]appdata.FinanceTracker, string, int) {
	var ftList []appdata.FinanceTracker
	var totalPriceStr2Decimals string
	var queryValues, totalRowsWithoutLimit int = 0, 0
	var err error
	if filter.Limit > 500 {
		filter.Limit = 500
	}
	//fmt.Println("inside GetRowsInFinanceTracker")

	//fmt.Printf("filter.WhereAccount: %#v, type:%T\n", filter.WhereAccount, filter.WhereAccount) // check default value and type
	//fmt.Printf("filter.WhereYear: %#v, type:%T\n", filter.WhereYear, filter.WhereYear) // check default value and type

	// start building query
	// (golang sql package does not support dynamic sql on other things than values)
	q := ` 
		SELECT COUNT(1) 
		FROM financeTracker AS fT
			LEFT JOIN category AS c ON c.category = fT.category AND c.gofiID = fT.gofiID
		WHERE fT.gofiID = ?
	`
	// others where on 3 fields max = 7 possibilities
	if filter.ID > 0 {
		queryValues = 999
		fmt.Println("filter.ID is used")
		q += ` AND ft.ID = ? `
	} else {
		if filter.WhereAccount != "" { //1
			queryValues += 1
			fmt.Println("filter.WhereAccount is used")
			q += ` AND account = ? `
		}
		if filter.WhereCategory != "" && filter.WhereCategory != "Toutes" { //2
			queryValues += 2
			fmt.Println("filter.WhereCategory is used")
			q += ` AND fT.category = ? `
		}
		if filter.WhereYear != 0 { //4
			queryValues += 4
			fmt.Println("filter.WhereYear is used")
			q += ` AND year = ? `
		}
		if filter.WhereMonth != 0 { // month used alone
			switch filter.WhereMonth {
			case 1:
				q += ` AND month =  1 `
			case 2:
				q += ` AND month =  2 `
			case 3:
				q += ` AND month =  3 `
			case 4:
				q += ` AND month =  4 `
			case 5:
				q += ` AND month =  5 `
			case 6:
				q += ` AND month =  6 `
			case 7:
				q += ` AND month =  7 `
			case 8:
				q += ` AND month =  8 `
			case 9:
				q += ` AND month =  9 `
			case 10:
				q += ` AND month = 10 `
			case 11:
				q += ` AND month = 11 `
			case 12:
				q += ` AND month = 12 `
			default:
				q += ` `
			}
			fmt.Println("filter.WhereMonth is used")
		}
		if filter.WhereChecked != 0 { // checked used alone
			if filter.WhereChecked == 2 {
				q += ` AND checked = 0 `
			} else {
				q += ` AND checked = 1 `
			}
			fmt.Println("filter.WhereChecked is used", filter.WhereChecked)
		}

		// order by column and type
		q += ` ORDER BY `
		switch filter.OrderBy {
		case "id":
			q += ` fT.id `
		case "date":
			fmt.Println("case date is used")
			q += ` (year*10000 + month*100 + day) `
			if filter.OrderSort == "DESC" {
				q += ` DESC `
			} else {
				q += ` ASC `
			}
			q += ` , fT.id `
		case "price":
			q += ` priceIntx100 `
			if filter.OrderSort == "DESC" {
				q += ` DESC `
			} else {
				q += ` ASC `
			}
			q += ` , fT.id `
		default:
			q += ` fT.id `
		}
		if filter.OrderSort == "DESC" {
			q += ` DESC `
		} else {
			q += ` ASC `
		}

		// finally, add limit
		q += ` LIMIT ?;`
	}
	//fmt.Printf("q: %v\n", q)
	// end building query
	q2 := strings.Replace(q, `COUNT(1)`,
		`fT.id, fT.gofiID, year, month, day, account, product, priceIntx100, 
			fT.category, ifnull(c.iconCodePoint,'e90a') AS icp, ifnull(c.colorHEX,'#808080') AS ch, 
			checked, dateChecked, mode`, 1)

	row := execSingleRow(queryValues, db, ctx, q, filter)
	if err := row.Scan(&totalRowsWithoutLimit); err != nil {
		fmt.Printf("GetRowsInFinanceTracker err1: %v\n", err)
		log.Fatal(err)
	}
	var totalPriceIntx100 int = 0
	var successfull bool
	var unsuccessfullReason string
	if totalRowsWithoutLimit == 0 {
		fmt.Println("in 0 row")
	} else if totalRowsWithoutLimit == 1 || filter.Limit == 1 {
		fmt.Println("in single row")
		row = execSingleRow(queryValues, db, ctx, q2, filter)
		var ft appdata.FinanceTracker
		if err := row.Scan(&ft.ID, &ft.GofiID, &ft.DateDetails.Year, &ft.DateDetails.Month, &ft.DateDetails.Day, &ft.Account, &ft.Product, &ft.PriceIntx100,
			&ft.Category, &ft.CategoryDetails.CategoryIcon, &ft.CategoryDetails.CategoryColor, &ft.Checked, &ft.DateChecked, &ft.Mode); err != nil {
			fmt.Printf("GetRowsInFinanceTracker err2: %v\n", err)
			log.Fatal(err)
		}
		ft.FormPriceStr2Decimals = ConvertPriceIntToStr(ft.PriceIntx100, true)
		totalPriceIntx100 += ft.PriceIntx100
		ft.Date, successfull, unsuccessfullReason = ConvertDateIntToStr(ft.DateDetails.Year, ft.DateDetails.Month, ft.DateDetails.Day, "EN", "-")
		if !successfull {
			ft.Date = "ERROR " + unsuccessfullReason
		}
		ft.DateDetails.MonthStr = appdata.MonthIto3A(ft.DateDetails.Month)
		// fmt.Printf("ft: %#v\n", ft)
		ftList = append(ftList, ft)
	} else {
		fmt.Printf("in multiple rows, nb: %v\n", totalRowsWithoutLimit)
		var rows *sql.Rows
		rows, err = execMultipleRow(queryValues, db, ctx, q2, filter)
		if err != nil {
			fmt.Printf("GetRowsInFinanceTracker err3: %v\n", err)
			log.Fatal("error on DB query: ", err)
		}
		for rows.Next() {
			var ft appdata.FinanceTracker
			if err := rows.Scan(&ft.ID, &ft.GofiID, &ft.DateDetails.Year, &ft.DateDetails.Month, &ft.DateDetails.Day, &ft.Account, &ft.Product, &ft.PriceIntx100,
				&ft.Category, &ft.CategoryDetails.CategoryIcon, &ft.CategoryDetails.CategoryColor, &ft.Checked, &ft.DateChecked, &ft.Mode); err != nil {
				fmt.Printf("GetRowsInFinanceTracker err4: %v\n", err)
				log.Fatal(err)
			}
			ft.FormPriceStr2Decimals = ConvertPriceIntToStr(ft.PriceIntx100, true)
			totalPriceIntx100 += ft.PriceIntx100
			ft.Date, successfull, unsuccessfullReason = ConvertDateIntToStr(ft.DateDetails.Year, ft.DateDetails.Month, ft.DateDetails.Day, "EN", "-")
			if !successfull {
				ft.Date = "ERROR " + unsuccessfullReason
			}
			ft.DateDetails.MonthStr = appdata.MonthIto3A(ft.DateDetails.Month)
			// fmt.Printf("ft: %#v\n", ft)
			ftList = append(ftList, ft)
		}
		// fmt.Printf("totalPriceIntx100: %v, inStr: %v\n", totalPriceIntx100, totalPriceStr2Decimals)
		rows.Close()
	}
	totalPriceStr2Decimals = ConvertPriceIntToStr(totalPriceIntx100, true)
	return ftList, totalPriceStr2Decimals, totalRowsWithoutLimit
}
func execSingleRow(queryValues int, db *sql.DB, ctx context.Context, q string, filter *appdata.FilterRows) *sql.Row {
	switch queryValues {
	case 0:
		return db.QueryRowContext(ctx, q, filter.GofiID, 1)
	case 1:
		return db.QueryRowContext(ctx, q, filter.GofiID, filter.WhereAccount, 1)
	case 2:
		return db.QueryRowContext(ctx, q, filter.GofiID, filter.WhereCategory, 1)
	case 3:
		return db.QueryRowContext(ctx, q, filter.GofiID, filter.WhereAccount, filter.WhereCategory, 1)
	case 4:
		return db.QueryRowContext(ctx, q, filter.GofiID, filter.WhereYear, 1)
	case 5:
		return db.QueryRowContext(ctx, q, filter.GofiID, filter.WhereAccount, filter.WhereYear, 1)
	case 6:
		return db.QueryRowContext(ctx, q, filter.GofiID, filter.WhereCategory, filter.WhereYear, 1)
	case 7:
		return db.QueryRowContext(ctx, q, filter.GofiID, filter.WhereAccount, filter.WhereCategory, filter.WhereYear, 1)
	case 999:
		return db.QueryRowContext(ctx, q, filter.GofiID, filter.ID)
	default:
		return db.QueryRowContext(ctx, q, filter.GofiID, 1)
	}
}
func execMultipleRow(queryValues int, db *sql.DB, ctx context.Context, q string, filter *appdata.FilterRows) (*sql.Rows, error) {
	switch queryValues {
	case 0:
		return db.QueryContext(ctx, q, filter.GofiID, filter.Limit)
	case 1:
		return db.QueryContext(ctx, q, filter.GofiID, filter.WhereAccount, filter.Limit)
	case 2:
		return db.QueryContext(ctx, q, filter.GofiID, filter.WhereCategory, filter.Limit)
	case 3:
		return db.QueryContext(ctx, q, filter.GofiID, filter.WhereAccount, filter.WhereCategory, filter.Limit)
	case 4:
		return db.QueryContext(ctx, q, filter.GofiID, filter.WhereYear, filter.Limit)
	case 5:
		return db.QueryContext(ctx, q, filter.GofiID, filter.WhereAccount, filter.WhereYear, filter.Limit)
	case 6:
		return db.QueryContext(ctx, q, filter.GofiID, filter.WhereCategory, filter.WhereYear, filter.Limit)
	case 7:
		return db.QueryContext(ctx, q, filter.GofiID, filter.WhereAccount, filter.WhereCategory, filter.WhereYear, filter.Limit)
	default:
		return db.QueryContext(ctx, q, filter.GofiID, filter.Limit)
	}
}

func addDataInRR(rr *appdata.RecurrentRecord) {
	var successfull bool
	var unsuccessfullReason string
	rr.FormPriceStr2Decimals = ConvertPriceIntToStr(rr.PriceIntx100, true)
	rr.Date, successfull, unsuccessfullReason = ConvertDateIntToStr(rr.DateDetails.Year, rr.DateDetails.Month, rr.DateDetails.Day, "EN", "-")
	if !successfull {
		rr.Date = "ERROR " + unsuccessfullReason
	}
}
func GetRowsInRecurrentRecord(ctx context.Context, db *sql.DB, gofiID int, rowID int) []appdata.RecurrentRecord {
	var rrList []appdata.RecurrentRecord
	var err error
	q := ` 
		SELECT id, year, month, day, recurrence, account, product, priceIntx100, category
		FROM recurrentRecord
		WHERE gofiID = ?
			AND id > ?
		ORDER BY year*10000 + month*100 + day, id DESC
	`
	//insert into recurrentRecord values (1, 5, 2024, 3, 5, 'mensuelle', 'CB', 'test1', 1000, 'Supermarché')
	if rowID > 0 {
		q = strings.Replace(q, `AND id >`,
			`AND id =`, 1)
		var singlerr appdata.RecurrentRecord
		row := db.QueryRowContext(ctx, q, gofiID, rowID)
		if err := row.Scan(&singlerr.ID, &singlerr.DateDetails.Year, &singlerr.DateDetails.Month, &singlerr.DateDetails.Day, &singlerr.Recurrence, &singlerr.Account, &singlerr.Product,
			&singlerr.PriceIntx100, &singlerr.Category); err != nil {
			fmt.Printf("GetRowsInRecurrentRecord single row error on DB query: %v\n", err)
			return rrList
		} else {
			addDataInRR(&singlerr)
			rrList = append(rrList, singlerr)
		}
	} else {
		var rows *sql.Rows
		rows, err = db.QueryContext(ctx, q, gofiID, rowID)
		if err != nil {
			log.Fatal("error on DB query: ", err)
		}
		for rows.Next() {
			var rr appdata.RecurrentRecord
			if err := rows.Scan(&rr.ID, &rr.DateDetails.Year, &rr.DateDetails.Month, &rr.DateDetails.Day, &rr.Recurrence, &rr.Account, &rr.Product,
				&rr.PriceIntx100, &rr.Category); err != nil {
				log.Fatal(err)
			}
			addDataInRR(&rr)
			// fmt.Printf("rr: %#v\n", rr)
			rrList = append(rrList, rr)
		}
		rows.Close()
	}
	return rrList
}

func ValidateRowsInFinanceTracker(ctx context.Context, db *sql.DB, gofiID int, checkedListInt []int, dateValidated string, mode string) {
	var query string
	if mode == "validate" {
		query = `
			UPDATE financeTracker 
			SET checked = 1,
				dateChecked = ?,
				exported = 0
			WHERE gofiID = ?
				AND id = ?;
			`
	} else if mode == "cancel" {
		query = `
			UPDATE financeTracker 
			SET dateIn = '1999-12-31', year = 1999, month = 12, day = 31, account = '-', product = 'DELETED LINE', 
				priceIntx100 = 0, category = '-', commentInt = 0, commentString = '-', 
				checked = 1, dateChecked = ?, exported = 0
			WHERE gofiID = ?
				AND id = ?;
			`
	} else {
		return
	}
	for _, intValue := range checkedListInt {
		_, err := db.ExecContext(ctx, query,
			dateValidated, gofiID, intValue,
		)
		if err != nil {
			fmt.Printf("error on UPDATE financeTracker with mode: %v, id: %v, err: %#v\n", mode, intValue, err)
		}
	}
}

func InsertRowInFinanceTracker(ctx context.Context, db *sql.DB, ft *appdata.FinanceTracker) (int64, error) {
	result, err := db.ExecContext(ctx, `
		INSERT INTO financeTracker (gofiID, dateIn, year, month, day, 
			account, product, priceIntx100, category, mode,
			commentInt, commentString, checked, dateChecked)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?);
		`,
		ft.GofiID, ft.Date, ft.DateDetails.Year, ft.DateDetails.Month, ft.DateDetails.Day,
		ft.Account, ft.Product, ft.PriceIntx100, ft.Category, ft.Mode,
		ft.CommentInt, ft.CommentString, ft.Checked, ft.DateChecked,
	)
	if err != nil {
		fmt.Printf("InsertRowInFinanceTracker err1: %#v\n", err)
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		fmt.Printf("InsertRowInFinanceTracker err2: %#v\n", err)
		return 0, err
	}
	return id, nil
}

func UpdateRowInFinanceTrackerFull(ctx context.Context, db *sql.DB, ft *appdata.FinanceTracker) (bool, error) {
	if ft.ID < 1 || ft.GofiID < 1 {
		fmt.Printf("UpdateRowInFinanceTrackerFull err1, id: %v, gofiID: %v\n", ft.ID, ft.GofiID)
		return true, errors.New("wrong ids")
	}
	_, err := db.ExecContext(ctx, `
		UPDATE financeTracker 
		SET dateIn = ?, year = ?, month = ?, day = ?, mode = ?, account = ?, product = ?, priceIntx100 = ?, category = ?,
			commentInt = ?, commentString = ?, checked = ?, dateChecked = ?, exported = 0
		WHERE ID = ?
			AND gofiID = ?;
		`,
		ft.Date, ft.DateDetails.Year, ft.DateDetails.Month, ft.DateDetails.Day, ft.Mode, ft.Account, ft.Product, ft.PriceIntx100, ft.Category,
		ft.CommentInt, ft.CommentString, ft.Checked, ft.DateChecked,
		ft.ID, ft.GofiID,
	)
	if err != nil {
		fmt.Printf("UpdateRowInFinanceTrackerFull err2: %#v\n", err)
		return true, err
	}
	return false, nil
}
func UpdateRowInFinanceTrackerLite(ctx context.Context, db *sql.DB, ft *appdata.FinanceTracker) (bool, error) {
	if ft.ID < 1 || ft.GofiID < 1 {
		fmt.Printf("UpdateRowInFinanceTrackerLite err1, id: %v, gofiID: %v\n", ft.ID, ft.GofiID)
		return true, errors.New("wrong ids")
	}
	_, err := db.ExecContext(ctx, `
		UPDATE financeTracker 
		SET dateIn = ?, year = ?, month = ?, day = ?, mode = ?, account = ?, product = ?, priceIntx100 = ?, category = ?,
			exported = 0
		WHERE ID = ?
			AND gofiID = ?;
		`,
		ft.Date, ft.DateDetails.Year, ft.DateDetails.Month, ft.DateDetails.Day, ft.Mode, ft.Account, ft.Product, ft.PriceIntx100, ft.Category,
		ft.ID, ft.GofiID,
	)
	if err != nil {
		fmt.Printf("UpdateRowInFinanceTrackerLite err2: %#v\n", err)
		return true, err
	}
	return false, nil
}

func UpdateRowsInFinanceTrackerToMode0(ctx context.Context, db *sql.DB, gofiID int, intList *[]int) bool {
	q := `
		UPDATE financeTracker
		SET mode = 0
		WHERE gofiID = ?
			AND id IN (XnumberOf?);
	`
	nbParams := ""
    anyList := make([]any, len(*intList)+1)
	nbParams, anyList = getOrCompleteVarsWithAnyListAndQuestionMarks(0, false, nbParams, &[]int{gofiID}, anyList)
	nbParams, anyList = getOrCompleteVarsWithAnyListAndQuestionMarks(1, true, nbParams, intList, anyList)
	if nbParams == "" {
		fmt.Println("UpdateRowsInFinanceTrackerToMode0 err1: no param")
		return true	
	}
	q = strings.Replace(q, `XnumberOf?`, nbParams, 1)
	// fmt.Printf("q: %v\n", q)
	_, err := db.ExecContext(ctx, q, anyList...,)
	if err != nil {
		fmt.Printf("UpdateRowsInFinanceTrackerToMode0 err2: %#v\n", err)
		return true
	}
	return false
}

func InsertRowInRecurrentRecord(ctx context.Context, db *sql.DB, rr *appdata.RecurrentRecord) (int64, error) {
	result, _ := db.ExecContext(ctx, `
		INSERT INTO recurrentRecord (gofiID, year, month, day, recurrence, account, product, priceIntx100, category)
		VALUES (?,?,?,?,?,?,?,?,?);
		`,
		rr.GofiID, rr.DateDetails.Year, rr.DateDetails.Month, rr.DateDetails.Day, rr.Recurrence, rr.Account, rr.Product, rr.PriceIntx100, rr.Category,
	)
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}
func UpdateRowInRecurrentRecord(ctx context.Context, db *sql.DB, rr *appdata.RecurrentRecord) (int64, error) {
	result, err := db.ExecContext(ctx, `
		UPDATE recurrentRecord
		SET year = ?, month = ?, day = ?, recurrence = ?, account = ?, product = ?, priceIntx100 = ?, category = ?
		WHERE gofiID = ?
			AND id = ?;
		`,
		rr.DateDetails.Year, rr.DateDetails.Month, rr.DateDetails.Day, rr.Recurrence, rr.Account, rr.Product, rr.PriceIntx100, rr.Category,
		rr.GofiID, rr.ID,
	)
	if err != nil {
		return 0, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rowsAffected, nil
}
func DeleteRowInRecurrentRecord(ctx context.Context, db *sql.DB, rr *appdata.RecurrentRecord) (int64, error) {
	result, err := db.ExecContext(ctx, `
		DELETE FROM recurrentRecord
		WHERE gofiID = ?
			AND id = ?;
		`,
		rr.GofiID, rr.ID,
	)
	if err != nil {
		return 0, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rowsAffected, nil
}
func UpdateDateInRecurrentRecord(ctx context.Context, db *sql.DB, rr *appdata.RecurrentRecord) {
	_, err := db.ExecContext(ctx, `
		UPDATE recurrentRecord
		SET year = ?, month = ?, day = ?
		WHERE gofiID = ?
			AND id = ?;
		`,
		rr.DateDetails.Year, rr.DateDetails.Month, rr.DateDetails.Day, rr.GofiID, rr.ID,
	)
	if err != nil {
		fmt.Printf("error on UPDATE recurrentRecord err: %#v\n", err)
	}
}

func InsertUpdateInLenderBorrower(ctx context.Context, db *sql.DB, lb *appdata.LendBorrow) (bool, int) {
	q := ` 
		SELECT COALESCE(MIN(id), 0), 
			COUNT(1), 
			COALESCE(SUM(lb.isActive), 0)
		FROM lenderBorrower AS lb
		WHERE lb.gofiID = ?
			AND name = ?;
	`
	var err error
	var lbID, nbRows, sumActive int = 0, 0, 0
	var lbName string
	var id int64
	var result sql.Result
	if lb.Who == "-" {
		lbName = lb.CreateLenderBorrowerName
	} else {
		lbName = lb.Who
	}
	row := db.QueryRowContext(ctx, q, lb.FT.GofiID, lbName)
	if err := row.Scan(&lbID, &nbRows, &sumActive); err != nil {
		fmt.Printf("InsertUpdateInLenderBorrower err1: %v\n", err)
		return true, http.StatusInternalServerError
	}
	lb.ID = lbID
	if nbRows == 0 {
		if len(lb.CreateLenderBorrowerName) > 0 {
			if lb.ModeInt == 1 || lb.ModeInt == 2 {
				fmt.Println("InsertUpdateInLenderBorrower in 0 row, create")
				q = ` 
					INSERT INTO lenderBorrower (gofiID, name)
					VALUES (?,?);
				`
				result, err = db.ExecContext(ctx, q,
					lb.FT.GofiID, lb.CreateLenderBorrowerName, //lb.FT.Date, lb.FT.Date, 1, lb.FT.PriceIntx100,
				)
				if err != nil {
					fmt.Printf("InsertUpdateInLenderBorrower err2: %v\n", err)
					return true, http.StatusInternalServerError
				}
				id, err = result.LastInsertId()
				if err != nil {
					fmt.Printf("InsertUpdateInLenderBorrower err3: %v\n", err)
					return true, http.StatusInternalServerError
				}
				lb.ID = int(id)
			} else {
				fmt.Println("InsertUpdateInLenderBorrower in 0 row, error wrong mode")
				return true, http.StatusBadRequest
			}
		} else {
			fmt.Println("InsertUpdateInLenderBorrower in 0 row, error no name")
			return true, http.StatusBadRequest
		}
	} else if nbRows == 1 && sumActive == 1 {
		fmt.Println("InsertUpdateInLenderBorrower in single row active")
	} else if nbRows == 1 && sumActive == 0 {
		fmt.Println("InsertUpdateInLenderBorrower in single row inactive")
		_, err := db.ExecContext(ctx, `
			UPDATE lenderBorrower
			SET isActive = 1
			WHERE id = ?;
			`,
			lbID,
		)
		if err != nil {
			fmt.Printf("InsertUpdateInLenderBorrower err4: %#v\n", err)
			return true, http.StatusInternalServerError
		}
	} else {
		fmt.Printf("InsertUpdateInLenderBorrower in multiple rows, nb: %v\n", nbRows)
		return true, http.StatusInternalServerError
	}
	return false, http.StatusOK
}

func FindLenderBorrowerFromFTid(ctx context.Context, db *sql.DB, lb *appdata.LendBorrow) bool {
	q := ` 
		SELECT COALESCE(MIN(id), 0), 
			COUNT(1), 
			COALESCE(MIN(srm.idLenderBorrower), 0)
		FROM specificRecordsByMode AS srm
		WHERE srm.gofiID = ?
			AND srm.idFinanceTracker = ?;
	`
	var srmID, nbRows, lbID int = 0, 0, 0
	row := db.QueryRowContext(ctx, q, lb.FT.GofiID, lb.FT.ID)
	if err := row.Scan(&srmID, &nbRows, &lbID); err != nil {
		fmt.Printf("FindLenderBorrowerFromFTid err1: %v\n", err)
		return true
	}
	if nbRows == 1 {
		fmt.Println("FindLenderBorrowerFromFTid found row")
		lb.ID = lbID
		q := ` 
			SELECT COALESCE(MIN(id), 0), 
				COUNT(1), 
				COALESCE(MIN(lb.name), 0)
			FROM lenderBorrower AS lb
			WHERE lb.gofiID = ?
				AND id = ?;
		`
		var lbName string
		row = db.QueryRowContext(ctx, q, lb.FT.GofiID, lbID)
		if err := row.Scan(&lbID, &nbRows, &lbName); err != nil {
			fmt.Printf("FindLenderBorrowerFromFTid err2: %v\n", err)
			return true
		}
		if nbRows == 1 {
			lb.Who = lbName
		} else {
			fmt.Printf("FindLenderBorrowerFromFTid err3 in rows, nb: %v\n", nbRows)
			return true
		}
	} else {
		fmt.Printf("FindLenderBorrowerFromFTid err4 in rows, nb: %v\n", nbRows)
		return true
	}
	return false
}

func UpdateStateInLenderBorrower(ctx context.Context, db *sql.DB, lb *appdata.LenderBorrower) bool {
	_, err := db.ExecContext(ctx, `
		UPDATE lenderBorrower
		SET isActive = ?
		WHERE id = ?;
		`,
		lb.IsActive, lb.ID,
	)
	if err != nil {
		fmt.Printf("UpdateStateInLenderBorrower err1: %#v\n", err)
		return true
	}
	return false
}

func DeleteSpecificRecordsByMode(ctx context.Context, db *sql.DB, gofiID int, intList *[]int) bool {
	q := `
		DELETE FROM specificRecordsByMode
		WHERE gofiID = ?
			AND idFinanceTracker IN (XnumberOf?);
	`
	nbParams := ""
    anyList := make([]any, len(*intList)+1)
	nbParams, anyList = getOrCompleteVarsWithAnyListAndQuestionMarks(0, false, nbParams, &[]int{gofiID}, anyList)
	nbParams, anyList = getOrCompleteVarsWithAnyListAndQuestionMarks(1, true, nbParams, intList, anyList)
	if nbParams == "" {
		fmt.Println("UpdateRowsInFinanceTrackerToMode0 err1: no param")
		return true	
	}
	q = strings.Replace(q, `XnumberOf?`, nbParams, 1)
	// fmt.Printf("q: %v\n", q)
	_, err := db.ExecContext(ctx, q, anyList...,)
	if err != nil {
		fmt.Printf("DeleteSpecificRecordsByMode err2: %#v\n", err)
		return true
	}
	return false
}

func InsertInSpecificRecordsByMode(ctx context.Context, db *sql.DB, lb *appdata.LendBorrow) bool {
	q := ` 
		INSERT INTO specificRecordsByMode (gofiID, idFinanceTracker, idLenderBorrower, mode)
		VALUES (?,?,?,?);
	`
	_, err := db.ExecContext(ctx, q,
		lb.FT.GofiID, lb.FT.ID, lb.ID, lb.ModeInt,
	)
	if err != nil {
		fmt.Printf("InsertInSpecificRecordsByMode err1: %v\n", err)
		return true
	}
	return false
}
