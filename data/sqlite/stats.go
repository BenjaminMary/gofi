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

	q3 := ` 
		SELECT fT.category, fT.year, ifnull(c.iconCodePoint,'e90a') AS icp, ifnull(c.colorHEX,'#000000') AS ch, SUM(priceIntx100) AS sum, COUNT(1) AS count
		FROM financeTracker AS fT
			LEFT JOIN category AS c ON c.category = fT.category
		WHERE gofiID = ?
			AND checked IN (1, ?)
			AND year > ?
		GROUP BY fT.category, fT.year
		ORDER BY fT.category, fT.year
	`
	// AND fT.category NOT IN ('Salaire','Invest')
	rows, err = db.QueryContext(ctx, q3, gofiID, checkedDataOnly, (year - 6))
	if err != nil {
		log.Fatal("error on DB query3: ", err)
	}
	var apexChartStats appdata.ApexChartStats
	apexChartStats.Labels = append(apexChartStats.Labels,
		strconv.Itoa(year-5), strconv.Itoa(year-4), strconv.Itoa(year-3), strconv.Itoa(year-2), strconv.Itoa(year-1), strconv.Itoa(year))
	var apexChartSerie appdata.ApexChartSerie
	for rows.Next() {
		var category, iconCodePoint, colorHEX string
		var sum, count, yearQ, yearDif int
		if err := rows.Scan(&category, &yearQ, &iconCodePoint, &colorHEX, &sum, &count); err != nil {
			log.Fatal(err)
		}
		if sum < 0 {
			sum = sum * -1

			// if category != "Salaire" && category != "Invest" {
			if apexChartSerie.Name == "" {
				apexChartSerie.Name = "&#x" + iconCodePoint + ";"
				apexChartSerie.Color = colorHEX
				apexChartSerie.Year = (year - 5)
			}

			if apexChartSerie.Year != yearQ {
				yearDif = yearQ - apexChartSerie.Year
				if yearDif > 0 {
					// for each missing year, add a 0
					for i := 1; i <= yearDif; i++ {
						apexChartSerie.Values = append(apexChartSerie.Values, "0")
						apexChartSerie.Year += 1
					}
				} else {
					// < 0 = new category add missing years until year variable
					yearDif = year - apexChartSerie.Year
					// for each missing year, add a 0
					for i := 1; i <= yearDif; i++ {
						apexChartSerie.Values = append(apexChartSerie.Values, "0")
						apexChartSerie.Year += 1
					}
				}
			}

			if apexChartSerie.Name != "&#x"+iconCodePoint+";" {
				apexChartSerie.SumStr = ConvertPriceIntToStr(apexChartSerie.SumInt, false)
				apexChartSerie.CountStr = strconv.Itoa(apexChartSerie.CountInt)
				apexChartStats.Series = append(apexChartStats.Series, apexChartSerie)
				apexChartSerie = appdata.ApexChartSerie{Name: "&#x" + iconCodePoint + ";", Color: colorHEX, Year: (year - 5)}
			}
			if apexChartSerie.Year < yearQ {
				yearDif = yearQ - apexChartSerie.Year
				if yearDif > 0 {
					// for each missing year, add a 0
					for i := 1; i <= yearDif; i++ {
						apexChartSerie.Values = append(apexChartSerie.Values, "0")
						apexChartSerie.Year += 1
					}
				}
			}
			if apexChartSerie.Year == yearQ {
				apexChartSerie.Values = append(apexChartSerie.Values, ConvertPriceIntToStr(sum, false))
				apexChartSerie.Year += 1
			} else {
				fmt.Printf("err Year: %v, category: %v\n", yearQ, category)
			}
			apexChartSerie.SumInt += sum
			apexChartSerie.CountInt += count
			// }
		}
	}
	rows.Close()

	fmt.Printf("apexChartStats: %#v\n", apexChartStats)
	return statsAccountList, statsCategoryList, totalAccountList, totalCategoryList, apexChartStats
}
