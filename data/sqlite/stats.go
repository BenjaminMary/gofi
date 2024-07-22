package sqlite

import (
	"context"
	"database/sql"
	"log"
	"strconv"
	"strings"

	"gofi/gofi/data/appdata"
)

func GetStatsForLineChartInFinanceTracker(ctx context.Context, db *sql.DB,
	gofiID int, checkedValidData int, year int) appdata.ApexChartStats {
	// initialize all years
	q := ` 
		SELECT DISTINCT year
		FROM financeTracker
		WHERE gofiID = ?
			AND checked IN (1, ?)
			AND year > 1999
			AND year <= ?
		ORDER BY year
	`
	rows, err := db.QueryContext(ctx, q, gofiID, checkedValidData, year)
	if err != nil {
		log.Fatal("GetStatsForLineChartInFinanceTracker error on DB query1: ", err)
	}
	apexChartStats := appdata.NewApexChartStats()
	loop := -1
	var yearMin int // yearMin used as an index
	for rows.Next() {
		loop += 1
		var yearQ int
		if err := rows.Scan(&yearQ); err != nil {
			log.Fatal(err)
		}
		if loop == 0 {
			yearMin = yearQ
		}
		apexChartStats.Labels = append(apexChartStats.Labels, strconv.Itoa(yearQ))
	}
	rows.Close()

	// initialize all accounts with values to 0
	q = ` 
		SELECT DISTINCT account
		FROM financeTracker
		WHERE gofiID = ?
			AND checked IN (1, ?)
			AND year > 1999
			AND year <= ?
		ORDER BY account
	`
	rows, err = db.QueryContext(ctx, q, gofiID, checkedValidData, year)
	if err != nil {
		log.Fatal("GetStatsForLineChartInFinanceTracker error on DB query2: ", err)
	}
	loop = -1
	for rows.Next() {
		loop += 1
		var apexChartSerie appdata.ApexChartSerie
		var account string
		if err := rows.Scan(&account); err != nil {
			log.Fatal(err)
		}
		apexChartSerie.Name = account
		for i := 0; i < len(apexChartStats.Labels); i++ {
			apexChartSerie.Values = append(apexChartSerie.Values, "-")
		}
		apexChartStats.FindSerie[account] = loop
		apexChartStats.Series = append(apexChartStats.Series, apexChartSerie)
	}
	rows.Close()

	// update an account value for the current year in each loop
	q = ` 
			SELECT DISTINCT year, account, 
				SUM(priceIntx100) OVER (
					PARTITION BY account
					ORDER BY account, year -- Window ordering (not necessarily the same as result ordering!)
					GROUPS BETWEEN -- Window for the SUM includes these rows:
						UNBOUNDED PRECEDING -- all rows before current one in window ordering
						AND CURRENT ROW -- up to and including current row.
					) AS cumulativeSum
			FROM financeTracker
			WHERE gofiID = ?
				AND checked IN (1, ?)
				AND year > 1999
				AND year <= ?
		`
	rows, err = db.QueryContext(ctx, q, gofiID, checkedValidData, year)
	if err != nil {
		log.Fatal("GetStatsForLineChartInFinanceTracker error on DB query3: ", err)
	}
	for rows.Next() {
		var account string
		var sum, yearQ, index int
		if err := rows.Scan(&yearQ, &account, &sum); err != nil {
			log.Fatal(err)
		}
		// find the Index on which the current account is stored
		index = apexChartStats.FindSerie[account]
		// update the Value corresponding to the current year Index inside the right Serie
		apexChartStats.Series[index].Values[yearQ-yearMin] = ConvertPriceIntToStr(sum, true)
	}
	rows.Close()
	// rework the years without values (can't be set to 0 due to cumulative values)
	for i := 0; i < len(apexChartStats.Series); i++ {
		currentValue := "0"
		for j := 0; j < len(apexChartStats.Labels); j++ {
			if apexChartStats.Series[i].Values[j] == "-" {
				apexChartStats.Series[i].Values[j] = currentValue
			} else {
				currentValue = apexChartStats.Series[i].Values[j]
			}
		}
	}
	// fmt.Printf("apexChartStats: %#v\n", apexChartStats)
	return *apexChartStats
}

func GetStatsInFinanceTracker(ctx context.Context, db *sql.DB, gofiID int,
	checkedValidData int, year int, checkedYearStats int, checkedGainsStats int) (
	[][]string, [][]string, []string, []string, appdata.ApexChartStats) {
	var statsAccountList, statsCategoryList [][]string // [account1, sum1, count1], [...,] | [category1, sum1, count1, icon1, color1], [...,]
	var totalAccountList, totalCategoryList []string   // [total, total, sum, count]
	q1 := ` 
		SELECT account, SUM(priceIntx100) AS sum, COUNT(1) AS c
		FROM financeTracker
		WHERE gofiID = ?
			AND checked IN (1, ?)
			AND year > 1999
			AND year <= ?
		GROUP BY account
		ORDER BY sum DESC
	`
	q2 := ` 
		SELECT fT.category, ifnull(c.iconCodePoint,'e90a') AS icp, ifnull(c.colorHEX,'#808080') AS ch, SUM(priceIntx100) AS sum, COUNT(1) AS count
		FROM financeTracker AS fT
			LEFT JOIN category AS c ON c.category = fT.category AND c.gofiID = fT.gofiID
		WHERE fT.gofiID = ?
			AND checked IN (1, ?)
			AND year = ?
		GROUP BY fT.category
		ORDER BY sum ASC
	`
	rows, err := db.QueryContext(ctx, q1, gofiID, checkedValidData, year)
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

	rows, err = db.QueryContext(ctx, q2, gofiID, checkedValidData, year)
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
		SELECT DISTINCT fT.category, ifnull(c.iconCodePoint,'e90a') AS icp, ifnull(c.colorHEX,'#808080') AS ch
		FROM financeTracker AS fT
			LEFT JOIN category AS c ON c.category = fT.category AND c.gofiID = fT.gofiID
		WHERE fT.gofiID = ?
			AND checked IN (1, ?)
			AND year > 1999
			AND year >= ?
			AND year <= ?
		ORDER BY fT.category
	`
	yearMin := year - 11
	rows, err = db.QueryContext(ctx, q3, gofiID, checkedValidData, yearMin, year)
	if err != nil {
		log.Fatal("error on DB query3: ", err)
	}
	apexChartStats := appdata.NewApexChartStats()
	if checkedYearStats == 1 {
		apexChartStats.Labels = append(apexChartStats.Labels,
			strconv.Itoa(yearMin), strconv.Itoa(year-10), strconv.Itoa(year-9), strconv.Itoa(year-8),
			strconv.Itoa(year-7), strconv.Itoa(year-6), strconv.Itoa(year-5), strconv.Itoa(year-4),
			strconv.Itoa(year-3), strconv.Itoa(year-2), strconv.Itoa(year-1), strconv.Itoa(year))
	} else {
		apexChartStats.Labels = append(apexChartStats.Labels,
			strconv.Itoa(year)+"-01", strconv.Itoa(year)+"-02", strconv.Itoa(year)+"-03", strconv.Itoa(year)+"-04",
			strconv.Itoa(year)+"-05", strconv.Itoa(year)+"-06", strconv.Itoa(year)+"-07", strconv.Itoa(year)+"-08",
			strconv.Itoa(year)+"-09", strconv.Itoa(year)+"-10", strconv.Itoa(year)+"-11", strconv.Itoa(year)+"-12")
	}
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
	var q4 string
	if checkedYearStats == 1 {
		q4 = ` 
			SELECT category, year, SUM(priceIntx100) AS sum
			FROM financeTracker
			WHERE gofiID = ?
				AND checked IN (1, ?)
				AND year > 1999
				AND year >= ?
				AND year <= ?
				AND priceIntx100 < 0
			GROUP BY category, year
		`
		if checkedGainsStats == 1 {
			q4 = strings.Replace(q4, `AND priceIntx100 < 0`,
				`AND priceIntx100 > 0`, 1)
		}
		rows, err = db.QueryContext(ctx, q4, gofiID, checkedValidData, yearMin, year)
	} else {
		q4 = ` 
			SELECT category, month, SUM(priceIntx100) AS sum
			FROM financeTracker
			WHERE gofiID = ?
				AND checked IN (1, ?)
				AND year = ?
				AND priceIntx100 < 0
			GROUP BY category, year, month
		`
		if checkedGainsStats == 1 {
			q4 = strings.Replace(q4, `AND priceIntx100 < 0`,
				`AND priceIntx100 > 0`, 1)
		}
		rows, err = db.QueryContext(ctx, q4, gofiID, checkedValidData, year)
	}
	if err != nil {
		log.Fatal("error on DB query4: ", err)
	}
	for rows.Next() {
		var category string
		var sum, dateQ, index int
		if err := rows.Scan(&category, &dateQ, &sum); err != nil {
			log.Fatal(err)
		}
		// find the Index on which the current category is stored
		index = apexChartStats.FindSerie[category]
		// update the Value corresponding to the current date Index inside the right Serie
		var dateIndex int
		if checkedYearStats == 1 {
			dateIndex = dateQ - yearMin //year
		} else {
			dateIndex = dateQ - 1 //month already 1-12
		}
		var sumStr string
		if checkedGainsStats == 1 {
			sumStr = ConvertPriceIntToStr(sum, false)
		} else {
			sumStr = ConvertPriceIntToStr(sum*-1, false)
		}
		apexChartStats.Series[index].Values[dateIndex] = sumStr
	}
	rows.Close()

	// fmt.Printf("apexChartStats: %#v\n", apexChartStats)
	return statsAccountList, statsCategoryList, totalAccountList, totalCategoryList, *apexChartStats
}
