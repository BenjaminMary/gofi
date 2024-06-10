package sqlite

import (
	"context"
	"database/sql"
	"log"
	"strconv"
)

func GetStatsInFinanceTracker(ctx context.Context, db *sql.DB, gofiID int, checkedDataOnly int, year int) (
	[][]string, [][]string, []string, []string) {
	var statsAccountList, statsCategoryList [][]string // [account1, sum1, count1], [...,] | [category1, sum1, count1, icon1, color1], [...,]
	var totalAccountList, totalCategoryList []string   // [total, total, sum, count]
	q1 := ` 
		SELECT account, SUM(priceIntx100) AS sum, COUNT(1) AS c
		FROM financeTracker
		WHERE gofiID = ?
			AND checked IN (1, ?)
			AND year <= ?
		GROUP BY account
		ORDER BY sum DESC
	`
	q2 := ` 
		SELECT fT.category, ifnull(c.iconCodePoint,'e90a') AS icp, ifnull(c.colorHEX,'#000000') AS ch, SUM(priceIntx100) AS sum, COUNT(1) AS count
		FROM financeTracker AS fT
			LEFT JOIN category AS c ON c.category = fT.category
		WHERE gofiID = ?
			AND checked IN (1, ?)
			AND year = ?
		GROUP BY fT.category
		ORDER BY sum ASC
	`
	rows, err := db.QueryContext(ctx, q1, gofiID, checkedDataOnly, year)
	if err != nil {
		log.Fatal("error on DB query1: ", err)
	}
	var totalPriceIntx100 int = 0
	var totalRows int = 0
	for rows.Next() {
		var statsRow []string
		var account string
		var sum, count int
		if err := rows.Scan(&account, &sum, &count); err != nil {
			log.Fatal(err)
		}
		totalPriceIntx100 += sum
		totalRows += count
		statsRow = append(statsRow, account, ConvertPriceIntToStr(sum), strconv.Itoa(count))
		statsAccountList = append(statsAccountList, statsRow)
	}
	totalAccountList = append(totalAccountList, ConvertPriceIntToStr(totalPriceIntx100), strconv.Itoa(totalRows))
	// fmt.Printf("statsList: %#v\n", statsList)
	rows.Close()

	rows, err = db.QueryContext(ctx, q2, gofiID, checkedDataOnly, year)
	if err != nil {
		log.Fatal("error on DB query2: ", err)
	}
	totalPriceIntx100 = 0
	totalRows = 0
	for rows.Next() {
		var statsRow []string
		var category, iconCodePoint, colorHEX string
		var sum, count int
		if err := rows.Scan(&category, &iconCodePoint, &colorHEX, &sum, &count); err != nil {
			log.Fatal(err)
		}
		totalPriceIntx100 += sum
		totalRows += count
		statsRow = append(statsRow, category, ConvertPriceIntToStr(sum), strconv.Itoa(count), iconCodePoint, colorHEX)
		statsCategoryList = append(statsCategoryList, statsRow)
	}
	totalCategoryList = append(totalCategoryList, ConvertPriceIntToStr(totalPriceIntx100), strconv.Itoa(totalRows))
	rows.Close()
	return statsAccountList, statsCategoryList, totalAccountList, totalCategoryList
}
