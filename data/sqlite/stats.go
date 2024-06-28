package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"

	"gofi/gofi/data/appdata"
)

func GetStatsInFinanceTracker(ctx context.Context, db *sql.DB, gofiID int, checkedDataOnly int, year int) (
	[][]string, [][]string, []string, []string, appdata.ApexChartStats) {
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
		statsRow = append(statsRow, account, ConvertPriceIntToStr(sum, true), strconv.Itoa(count))
		statsAccountList = append(statsAccountList, statsRow)
	}
	totalAccountList = append(totalAccountList, ConvertPriceIntToStr(totalPriceIntx100, true), strconv.Itoa(totalRows))
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
		statsRow = append(statsRow, category, ConvertPriceIntToStr(sum, true), strconv.Itoa(count), iconCodePoint, colorHEX)
		statsCategoryList = append(statsCategoryList, statsRow)
	}
	totalCategoryList = append(totalCategoryList, ConvertPriceIntToStr(totalPriceIntx100, true), strconv.Itoa(totalRows))
	rows.Close()

	// initialize all categories with values to 0
	q3 := ` 
		SELECT DISTINCT fT.category, ifnull(c.iconCodePoint,'e90a') AS icp, ifnull(c.colorHEX,'#000000') AS ch
		FROM financeTracker AS fT
			LEFT JOIN category AS c ON c.category = fT.category
		WHERE gofiID = ?
			AND checked IN (1, ?)
			AND year > ?
			AND year <= ?
			AND priceIntx100 < 0
		ORDER BY fT.category
	`
	rows, err = db.QueryContext(ctx, q3, gofiID, checkedDataOnly, (year - 6), year)
	if err != nil {
		log.Fatal("error on DB query3: ", err)
	}
	apexChartStats := appdata.NewApexChartStats()
	yearMin := year - 5
	apexChartStats.Labels = append(apexChartStats.Labels,
		strconv.Itoa(yearMin), strconv.Itoa(year-4), strconv.Itoa(year-3), strconv.Itoa(year-2), strconv.Itoa(year-1), strconv.Itoa(year))
	loop := -1
	for rows.Next() {
		loop += 1
		var apexChartSerie appdata.ApexChartSerie
		var category, iconCodePoint, colorHEX string
		if err := rows.Scan(&category, &iconCodePoint, &colorHEX); err != nil {
			log.Fatal(err)
		}
		apexChartSerie.Name = category
		apexChartSerie.Icon = "&#x" + iconCodePoint + ";"
		apexChartSerie.Color = colorHEX
		for i := 0; i < len(apexChartStats.Labels); i++ {
			apexChartSerie.Values = append(apexChartSerie.Values, "0")
		}
		apexChartStats.FindSerie[category] = loop
		apexChartStats.Series = append(apexChartStats.Series, apexChartSerie)
	}
	rows.Close()

	// update a category value for the current year in each loop
	q4 := ` 
		SELECT category, year, SUM(priceIntx100) AS sum
		FROM financeTracker
		WHERE gofiID = ?
			AND checked IN (1, ?)
			AND year > ?
			AND year <= ?
			AND priceIntx100 < 0
		GROUP BY category, year
	`
	rows, err = db.QueryContext(ctx, q4, gofiID, checkedDataOnly, (year - 6), year)
	if err != nil {
		log.Fatal("error on DB query4: ", err)
	}
	for rows.Next() {
		var category string
		var sum, yearQ, index int
		if err := rows.Scan(&category, &yearQ, &sum); err != nil {
			log.Fatal(err)
		}
		// find the Index on which the current category is stored
		index = apexChartStats.FindSerie[category]
		// update the Value corresponding to the current year Index inside the right Serie
		apexChartStats.Series[index].Values[yearQ-yearMin] = ConvertPriceIntToStr(sum*-1, false)
	}
	rows.Close()

	fmt.Printf("apexChartStats: %#v\n", apexChartStats)
	return statsAccountList, statsCategoryList, totalAccountList, totalCategoryList, *apexChartStats
}
