package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

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
		SELECT DISTINCT fT.category, ifnull(c.iconCodePoint,'e90a') AS icp, ifnull(c.colorHEX,'#808080') AS ch, ifnull(c.defaultInStats,1) AS inst
		FROM financeTracker AS fT
			LEFT JOIN category AS c ON c.category = fT.category AND c.gofiID = fT.gofiID
		WHERE fT.gofiID = ?
			AND checked IN (1, ?)
			AND year > 1999
			AND year >= ?
			AND year <= ?
		ORDER BY c.catOrder
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
		var defaultInStats int
		if err := rows.Scan(&category, &iconCodePoint, &colorHEX, &defaultInStats); err != nil {
			log.Fatal(err)
		}
		apexChartSerie.Name = category
		apexChartSerie.Icon = "&#x" + iconCodePoint + ";"
		apexChartSerie.Color = colorHEX
		apexChartSerie.InStats = defaultInStats
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
			GROUP BY category, year
			HAVING SUM(priceIntx100) < 0;
		`
		if checkedGainsStats == 1 {
			q4 = strings.Replace(q4, `SUM(priceIntx100) < 0`,
				`SUM(priceIntx100) > 0`, 1)
		}
		rows, err = db.QueryContext(ctx, q4, gofiID, checkedValidData, yearMin, year)
	} else {
		q4 = ` 
			SELECT category, month, SUM(priceIntx100) AS sum
			FROM financeTracker
			WHERE gofiID = ?
				AND checked IN (1, ?)
				AND year = ?
			GROUP BY category, year, month
			HAVING SUM(priceIntx100) < 0;
		`
		if checkedGainsStats == 1 {
			q4 = strings.Replace(q4, `SUM(priceIntx100) < 0`,
				`SUM(priceIntx100) > 0`, 1)
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

func GetBudgetStats(ctx context.Context, db *sql.DB, uc *appdata.UserCategories) {
	// for the current user, get the budget stats for each category GetBudget
	currentTime := time.Now()
	currentDate := currentTime.Format(time.DateOnly) // YYYY-MM-DD
	// fmt.Printf("currentDate: %v\n", currentDate)
	startDateCurrentWeek := currentTime.AddDate(0, 0, int(currentTime.Weekday())*-1) // 1=monday, 7=sunday
	startDateCurrentMonth := time.Date(currentTime.Year(), currentTime.Month(), 1, 0, 0, 0, 0, time.UTC)
	startDateCurrentYear := time.Date(currentTime.Year(), time.January, 1, 0, 0, 0, 0, time.UTC)
	endDateCurrentWeek := startDateCurrentWeek.AddDate(0, 0, 6).Format(time.DateOnly)
	endDateCurrentMonth := startDateCurrentMonth.AddDate(0, 1, -1).Format(time.DateOnly)
	endDateCurrentYear := startDateCurrentYear.AddDate(1, 0, -1).Format(time.DateOnly)
	startDatePreviousWeek := startDateCurrentWeek.AddDate(0, 0, -7).Format(time.DateOnly)
	startDatePreviousMonth := startDateCurrentMonth.AddDate(0, -1, 0).Format(time.DateOnly)
	startDatePreviousYear := startDateCurrentYear.AddDate(-1, 0, 0).Format(time.DateOnly)
	endDatePreviousWeek := startDateCurrentWeek.AddDate(0, 0, -1).Format(time.DateOnly)
	endDatePreviousMonth := startDateCurrentMonth.AddDate(0, 0, -1).Format(time.DateOnly)
	endDatePreviousYear := startDateCurrentYear.AddDate(0, 0, -1).Format(time.DateOnly)
	startDateCurrentWeekFormated := startDateCurrentWeek.Format(time.DateOnly)
	startDateCurrentMonthFormated := startDateCurrentMonth.Format(time.DateOnly)
	startDateCurrentYearFormated := startDateCurrentYear.Format(time.DateOnly)
	q := ` 
		SELECT IFNULL(SUM(priceIntx100)*-1, 0) AS sum
		FROM financeTracker
		WHERE gofiID = ?
			AND category = ?
			AND dateIn BETWEEN ? AND ?;
	`
	var err error
	for i := 0; i < len(uc.Categories); i++ {
		// fmt.Printf("uc.Categories[i]: %#v\n", uc.Categories[i])
		var sumCurrent, sumPrevious int = 0, 0
		switch uc.Categories[i].BudgetType {
		case "reset":
			switch uc.Categories[i].BudgetPeriod {
			case "hebdomadaire":
				uc.Categories[i].BudgetCurrentPeriodStartDate = startDateCurrentWeekFormated
				uc.Categories[i].BudgetCurrentPeriodEndDate = endDateCurrentWeek
				uc.Categories[i].BudgetPreviousPeriodStartDate = startDatePreviousWeek
				uc.Categories[i].BudgetPreviousPeriodEndDate = endDatePreviousWeek
			case "mensuelle":
				uc.Categories[i].BudgetCurrentPeriodStartDate = startDateCurrentMonthFormated
				uc.Categories[i].BudgetCurrentPeriodEndDate = endDateCurrentMonth
				uc.Categories[i].BudgetPreviousPeriodStartDate = startDatePreviousMonth
				uc.Categories[i].BudgetPreviousPeriodEndDate = endDatePreviousMonth
			case "annuelle":
				uc.Categories[i].BudgetCurrentPeriodStartDate = startDateCurrentYearFormated
				uc.Categories[i].BudgetCurrentPeriodEndDate = endDateCurrentYear
				uc.Categories[i].BudgetPreviousPeriodStartDate = startDatePreviousYear
				uc.Categories[i].BudgetPreviousPeriodEndDate = endDatePreviousYear
			default:
				fmt.Println("err0 BudgetPeriod")
				continue
			}
			err = db.QueryRowContext(ctx, q, uc.GofiID, uc.Categories[i].Name,
				uc.Categories[i].BudgetCurrentPeriodStartDate, uc.Categories[i].BudgetCurrentPeriodEndDate).Scan(&sumCurrent)
			switch {
			case err == sql.ErrNoRows:
				fmt.Println("err1 GetBudgetStats query")
				continue
			case err != nil:
				fmt.Println("err2 GetBudgetStats query")
				continue
			}
			err = db.QueryRowContext(ctx, q, uc.GofiID, uc.Categories[i].Name,
				uc.Categories[i].BudgetPreviousPeriodStartDate, uc.Categories[i].BudgetPreviousPeriodEndDate).Scan(&sumPrevious)
			switch {
			case err == sql.ErrNoRows:
				fmt.Println("err3 GetBudgetStats query")
				continue
			case err != nil:
				fmt.Println("err4 GetBudgetStats query")
				continue
			}
		case "cumulative":
			t, _ := time.Parse(time.DateOnly, uc.Categories[i].BudgetCurrentPeriodStartDate)
			duration := currentTime.Sub(t)
			switch uc.Categories[i].BudgetPeriod {
			case "hebdomadaire":
				// fmt.Println("Weeks : ", int(duration.Hours()/168))
				// uc.Categories[i].BudgetPriceXPeriod = int(duration.Hours()/168) * uc.Categories[i].BudgetPrice
				uc.Categories[i].BudgetPrice = int(duration.Hours()/168) * uc.Categories[i].BudgetPrice
			case "mensuelle":
				// fmt.Println("Months : ", int(duration.Hours()/730))
				// uc.Categories[i].BudgetPriceXPeriod = int(duration.Hours()/730) * uc.Categories[i].BudgetPrice
				uc.Categories[i].BudgetPrice = int(duration.Hours()/730) * uc.Categories[i].BudgetPrice
			case "annuelle":
				// fmt.Println("Years : ", int(duration.Hours()/8760))
				// uc.Categories[i].BudgetPriceXPeriod = int(duration.Hours()/8760) * uc.Categories[i].BudgetPrice
				uc.Categories[i].BudgetPrice = int(duration.Hours()/8760) * uc.Categories[i].BudgetPrice
			default:
				// fmt.Println("Days : ", int(duration.Hours()/24))
				fmt.Println("err0 BudgetPeriod")
				continue
			}
			uc.Categories[i].BudgetCurrentPeriodEndDate = currentDate
			err = db.QueryRowContext(ctx, q, uc.GofiID, uc.Categories[i].Name,
				uc.Categories[i].BudgetCurrentPeriodStartDate, uc.Categories[i].BudgetCurrentPeriodEndDate).Scan(&sumCurrent)
			switch {
			case err == sql.ErrNoRows:
				fmt.Println("err5 GetBudgetStats query")
				continue
			case err != nil:
				fmt.Printf("err6 GetBudgetStats query: %v\n", err)
				continue
			}
			err = db.QueryRowContext(ctx, q, uc.GofiID, uc.Categories[i].Name,
				uc.Categories[i].BudgetPreviousPeriodStartDate, uc.Categories[i].BudgetPreviousPeriodEndDate).Scan(&sumPrevious)
			switch {
			case err == sql.ErrNoRows:
				fmt.Println("err7 GetBudgetStats query")
				continue
			case err != nil:
				fmt.Printf("err8 GetBudgetStats query: %v\n", err)
				continue
			}
		default:
			continue
		}
		uc.Categories[i].IntBudgetAmount = sumCurrent
		uc.Categories[i].BudgetAmount = ConvertPriceIntToStr(sumCurrent, false)
		uc.Categories[i].IntBudgetPreviousAmount = sumPrevious
		uc.Categories[i].BudgetPreviousAmount = ConvertPriceIntToStr(sumPrevious, false)
		// fmt.Printf("Name: %v, Amount: %v\n", uc.Categories[i].Name, uc.Categories[i].BudgetAmount)
	}
}

func GetLenderBorrowerStats(ctx context.Context, db *sql.DB, gofiID int, activeListOnly bool) ([]appdata.LenderBorrower, []appdata.LenderBorrower) {
	// list the lenders and borrowers
	var lbListActive, lbListInactive []appdata.LenderBorrower
	q1 := ` 
		SELECT id, name, isActive
		FROM lenderBorrower
		WHERE gofiID = ?;
	`
	q2 := ` 
		SELECT SUM(ft.priceIntx100)
		FROM lenderBorrower AS lb
			INNER JOIN specificRecordsByMode AS srm ON lb.id = srm.idLenderBorrower
			INNER JOIN financeTracker AS ft ON srm.idFinanceTracker = ft.id
		WHERE lb.id = ?
			AND lb.gofiID = ?
			AND srm.mode IN (?,?)
		GROUP BY lb.id;
	`
	rows, err := db.QueryContext(ctx, q1, gofiID)
	if err != nil {
		fmt.Printf("error on GetLenderBorrowerStats query1: %v\n", err)
	}
	for rows.Next() {
		var lb appdata.LenderBorrower
		var isActive int
		if err := rows.Scan(&lb.ID, &lb.Name, &isActive); err != nil {
			fmt.Println("error in loop on GetLenderBorrowerStats query2")
			log.Fatal(err)
		}
		if activeListOnly {
			if isActive == 1 {
				lbListActive = append(lbListActive, lb)
			}
		} else {
			err = db.QueryRowContext(ctx, q2, lb.ID, gofiID, 1, 2).Scan(&lb.AmountLentBorrowedIntx100)
			switch {
			case err == sql.ErrNoRows:
				fmt.Println("GetLenderBorrowerStats err3")
				continue
			case err != nil:
				fmt.Printf("GetLenderBorrowerStats err4: %v\n", err)
				continue
			}
			err = db.QueryRowContext(ctx, q2, lb.ID, gofiID, 3, 4).Scan(&lb.AmountSentReceivedIntx100)
			switch {
			case err == sql.ErrNoRows:
				fmt.Println("GetLenderBorrowerStats no row in the received part, set to 0")
				lb.AmountSentReceivedIntx100 = 0
			case err != nil:
				fmt.Printf("GetLenderBorrowerStats err5: %v\n", err)
				continue
			}
			lb.AmountLentBorrowedStr2Decimals = ConvertPriceIntToStr(lb.AmountLentBorrowedIntx100, false)
			lb.AmountSentReceivedStr2Decimals = ConvertPriceIntToStr(lb.AmountSentReceivedIntx100, false)
			if isActive == 1 {
				lbListActive = append(lbListActive, lb)
			} else if isActive == 0 {
				lbListInactive = append(lbListInactive, lb)
			}
		}
	}
	rows.Close()
	// fmt.Printf("lbList: %#v\n", lbList)
	return lbListActive, lbListInactive
}

func GetLenderBorrowerDetailedStats(ctx context.Context, db *sql.DB, gofiID int, lbID int) ([]appdata.FinanceTracker, []appdata.FinanceTracker, string) {
	var ftList1, ftList2 []appdata.FinanceTracker
	var lbName string
	q1 := ` 
		SELECT name
		FROM lenderBorrower
		WHERE gofiID = ?
			AND isActive = 1
			AND id = ?;
	`
	err := db.QueryRowContext(ctx, q1, gofiID, lbID).Scan(&lbName)
	switch {
	case err == sql.ErrNoRows:
		fmt.Println("GetLenderBorrowerDetailedStats err1")
		return ftList1, ftList2, lbName
	case err != nil:
		fmt.Printf("GetLenderBorrowerDetailedStats err2: %v\n", err)
		return ftList1, ftList2, lbName
	}
	q2 := ` 
		SELECT ft.id, ft.dateIn, ft.account, ft.category, ft.priceIntx100, ft.product, ft.mode,
			ft.year, ft.month, ft.day,
			ifnull(c.iconCodePoint,'e90a') AS icp, ifnull(c.colorHEX,'#808080') AS ch
		FROM specificRecordsByMode AS srm
			INNER JOIN financeTracker AS ft ON srm.idFinanceTracker = ft.id
			LEFT JOIN category AS c ON c.category = fT.category AND c.gofiID = fT.gofiID
		WHERE srm.idLenderBorrower = ?
			AND srm.gofiID = ?
			AND srm.mode IN (?,?)
		ORDER BY ft.dateIn DESC, ft.id DESC;
	`
	rows, err := db.QueryContext(ctx, q2, lbID, gofiID, 1, 2)
	if err != nil {
		fmt.Printf("error on GetLenderBorrowerDetailedStats query1: %v\n", err)
	}
	for rows.Next() {
		var ft appdata.FinanceTracker
		if err := rows.Scan(&ft.ID, &ft.Date, &ft.Account, &ft.Category, &ft.PriceIntx100, &ft.Product, &ft.Mode,
			&ft.DateDetails.Year, &ft.DateDetails.Month, &ft.DateDetails.Day,
			&ft.CategoryDetails.CategoryIcon, &ft.CategoryDetails.CategoryColor); err != nil {
			fmt.Println("error in loop on GetLenderBorrowerDetailedStats query2")
			log.Fatal(err)
		}
		ft.DateDetails.MonthStr = appdata.MonthIto3A(ft.DateDetails.Month)
		ft.FormPriceStr2Decimals = ConvertPriceIntToStr(ft.PriceIntx100, true)
		ftList1 = append(ftList1, ft)
	}
	rows.Close()
	rows, err = db.QueryContext(ctx, q2, lbID, gofiID, 3, 4)
	if err != nil {
		fmt.Printf("error on GetLenderBorrowerDetailedStats query3: %v\n", err)
	}
	for rows.Next() {
		var ft appdata.FinanceTracker
		if err := rows.Scan(&ft.ID, &ft.Date, &ft.Account, &ft.Category, &ft.PriceIntx100, &ft.Product, &ft.Mode,
			&ft.DateDetails.Year, &ft.DateDetails.Month, &ft.DateDetails.Day,
			&ft.CategoryDetails.CategoryIcon, &ft.CategoryDetails.CategoryColor); err != nil {
			fmt.Println("error in loop on GetLenderBorrowerDetailedStats query4")
			log.Fatal(err)
		}
		ft.DateDetails.MonthStr = appdata.MonthIto3A(ft.DateDetails.Month)
		ft.FormPriceStr2Decimals = ConvertPriceIntToStr(ft.PriceIntx100, true)
		ftList2 = append(ftList2, ft)
	}
	rows.Close()
	return ftList1, ftList2, lbName
}
