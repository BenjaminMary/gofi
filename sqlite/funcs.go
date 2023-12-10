package sqlite

import (
	// "fmt"
	"strconv"
	"strings"
	// "time"
)

func ConvertPriceIntToStr(i int) string {
	var FormPriceStr2Decimals string
	switch {
		case i > 99:
			// fmt.Printf("FormPriceStr2Decimals: %v\n", strconv.Itoa(i)[:len(strconv.Itoa(i))-2]) // all except last 2 (stop at x-2)
			// fmt.Printf("FormPriceStr2Decimals: %v\n", strconv.Itoa(i)[len(strconv.Itoa(i))-2:]) // last 2 only (start at x-2)
			FormPriceStr2Decimals = strconv.Itoa(i)[:len(strconv.Itoa(i))-2] + "." + strconv.Itoa(i)[len(strconv.Itoa(i))-2:]
		case i > 9:
			FormPriceStr2Decimals = "0." + strconv.Itoa(i)
		default:
			FormPriceStr2Decimals = "0.0" + strconv.Itoa(i)
	}
	return FormPriceStr2Decimals
}


func ConvertDateIntToStr(yearInt int, monthInt int, dayInt int, dateLanguageFormat string, dateSeparator string) (string, bool, string) {
	/*
		input integers : Year, Month, Day
		output date string : YYYY-MM-DD, YYYY/MM/DD, DD/MM/YYYY, DD-MM-YYYY
		output boolean true if successfull, false otherwise
	*/
	var dateStr string = "1900-12-31"
	var successfull bool = false
	var unsuccessfullReason string = ""
	var monthStr, dayStr string
	var dateStrSplitted []string

	switch dateSeparator {
		case "-","/":
			dateStr = "1900-12-31"
		default:
			unsuccessfullReason = "wrong separator"
			return dateStr, successfull, unsuccessfullReason
	}
	// coherence control on basic stuff, will let a 2011-02-31 valid (the time.Date func would have put it as 2011-03-03)
	if (yearInt < 1900) || (yearInt > 2200) {
		unsuccessfullReason = "year not between 1900 and 2200"
		return dateStr, successfull, unsuccessfullReason
	} else {dateStrSplitted = append(dateStrSplitted,strconv.Itoa(yearInt))}
	if (monthInt < 1) || (monthInt > 12) {
		unsuccessfullReason = "month not between 1 and 12"
		return dateStr, successfull, unsuccessfullReason
	} else {
		if (monthInt < 10) {monthStr = "0" + strconv.Itoa(monthInt)} else {monthStr = strconv.Itoa(monthInt)}
		dateStrSplitted = append(dateStrSplitted, monthStr)
	}
	if (dayInt < 1) || (dayInt > 31) {
		unsuccessfullReason = "day not between 1 and 31"
		return dateStr, successfull, unsuccessfullReason
	} else {
		if (dayInt < 10) {dayStr = "0" + strconv.Itoa(dayInt)} else {dayStr = strconv.Itoa(dayInt)}
		dateStrSplitted = append(dateStrSplitted, dayStr)
	}
	switch dateLanguageFormat {
		case "FR": // DD/MM/YYYY
			dateStr = dateStrSplitted[2] + dateSeparator + dateStrSplitted[1] + dateSeparator + dateStrSplitted[0]
		case "EN":
			dateStr = dateStrSplitted[0] + dateSeparator + dateStrSplitted[1] + dateSeparator + dateStrSplitted[2]
		default:
			unsuccessfullReason = "wrong date language format"
			return dateStr, successfull, unsuccessfullReason
	}

	// fmt.Println(time.Date(yearInt, time.Month(monthInt), dayInt, 0, 0, 0, 0, time.UTC))
	// fmt.Printf("yearInt, monthInt, dayInt: %v, %v, %v\n", yearInt, monthInt, dayInt)
	successfull = true
	return dateStr, successfull, unsuccessfullReason
}

func ConvertDateStrToInt(dateStr string, dateLanguageFormat string, dateSeparator string) (int, int, int, bool, string) {
	/*
		input date string : YYYY-MM-DD, YYYY/MM/DD, DD/MM/YYYY, DD-MM-YYYY
		output integers : Year, Month, Day
		output boolean true if successfull, false otherwise
	*/
	var yearInt, monthInt, dayInt, i int = 0, 0, 0, 0
	var successfull bool = false
	var unsuccessfullReason string = ""
	var err error
	var dateStrSplitted []string
	var dateIntSplitted []int

	switch dateSeparator {
		case "-","/":
			dateStrSplitted = strings.Split(dateStr, dateSeparator)
			if (len(dateStrSplitted) != 3) {
				unsuccessfullReason = "not splitted in 3"
				return 0, 0, 0, successfull, unsuccessfullReason
			} 
		default:
			unsuccessfullReason = "wrong separator"
			return 0, 0, 0, successfull, unsuccessfullReason
	}
	if (len(dateStr) > 10) || (len(dateStr) < 8) {
		unsuccessfullReason = "unhandled date length"
		return 0, 0, 0, successfull, unsuccessfullReason
	}
	for _, element := range dateStrSplitted {
		i, err = strconv.Atoi(element)
		if err != nil { // Always check errors even if they should not happen.
			unsuccessfullReason = "str to int"
			return 0, 0, 0, successfull, unsuccessfullReason
		}
		dateIntSplitted = append(dateIntSplitted,i)
	}
	switch dateLanguageFormat {
		case "FR":
			dayInt = dateIntSplitted[0]
			monthInt = dateIntSplitted[1]
			yearInt = dateIntSplitted[2]
		case "EN":
			yearInt = dateIntSplitted[0]
			monthInt = dateIntSplitted[1]
			dayInt = dateIntSplitted[2]
		default:
			unsuccessfullReason = "wrong date language format"
			return 0, 0, 0, successfull, unsuccessfullReason
	}
	// coherence control on basic stuff, will let a 2011-02-31 valid (the time.Date func would have put it as 2011-03-03)
	if (yearInt < 1900) || (yearInt > 2200) {
		unsuccessfullReason = "year not between 1900 and 2200"
		return 0, 0, 0, successfull, unsuccessfullReason
	}
	if (monthInt < 1) || (monthInt > 12) {
		unsuccessfullReason = "month not between 1 and 12"
		return 0, 0, 0, successfull, unsuccessfullReason
	}
	if (dayInt < 1) || (dayInt > 31) {
		unsuccessfullReason = "day not between 1 and 31"
		return 0, 0, 0, successfull, unsuccessfullReason
	}

	// fmt.Println(time.Date(yearInt, time.Month(monthInt), dayInt, 0, 0, 0, 0, time.UTC))
	// fmt.Printf("yearInt, monthInt, dayInt: %v, %v, %v\n", yearInt, monthInt, dayInt)
	successfull = true
	return yearInt, monthInt, dayInt, successfull, unsuccessfullReason
}