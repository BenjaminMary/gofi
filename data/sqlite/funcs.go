package sqlite

import (
	// "fmt"
	"strconv"
	"strings"
	// "time"
)

func ConvertPriceIntToStr(i int) string {
	var PriceStr2Decimals string
	var isNegative bool = false
	if i < 0 {
		i = i * -1
		isNegative = true
	}
	switch {
	case i > 99:
		// fmt.Printf("PriceStr2Decimals: %v\n", strconv.Itoa(i)[:len(strconv.Itoa(i))-2]) // all except last 2 (stop at x-2)
		// fmt.Printf("PriceStr2Decimals: %v\n", strconv.Itoa(i)[len(strconv.Itoa(i))-2:]) // last 2 only (start at x-2)
		PriceStr2Decimals = strconv.Itoa(i)[:len(strconv.Itoa(i))-2] + "." + strconv.Itoa(i)[len(strconv.Itoa(i))-2:]
	case i > 9:
		PriceStr2Decimals = "0." + strconv.Itoa(i)
	default:
		PriceStr2Decimals = "0.0" + strconv.Itoa(i)
	}
	if isNegative {
		PriceStr2Decimals = "-" + PriceStr2Decimals
	}
	return PriceStr2Decimals
}

func ConvertPriceStrToInt(s string, csvDecimalDelimiter string) int {
	var PriceIntx100 int
	//fmt.Println("---------------")
	//fmt.Printf("s: %v\n", s)
	//fmt.Printf("csvDecimalDelimiter: %v\n", csvDecimalDelimiter)
	if !strings.Contains(s, csvDecimalDelimiter) { // add .00 if "." not present in string, equivalent of *100 with next step
		s = s + csvDecimalDelimiter + "00"
	} else {
		decimalPart := strings.Split(s, csvDecimalDelimiter)
		i := len(decimalPart[1])
		switch {
		case i == 1:
			s = s + "0"
		case i == 2:
			// pass
		default:
			// i > 2
			s = decimalPart[0] + decimalPart[1][:2] //we only keep the first 2 decimals
		}
	}
	PriceIntx100, _ = strconv.Atoi(strings.Replace(s, csvDecimalDelimiter, "", 1))
	//fmt.Printf("PriceIntx100: %v\n", PriceIntx100)
	return PriceIntx100
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
	case "-", "/":
		dateStr = "1900-12-31"
	default:
		unsuccessfullReason = "wrong separator"
		return dateStr, successfull, unsuccessfullReason
	}
	// coherence control on basic stuff, will let a 2011-02-31 valid (the time.Date func would have put it as 2011-03-03)
	if (yearInt < 1900) || (yearInt > 2200) {
		unsuccessfullReason = "year not between 1900 and 2200"
		return dateStr, successfull, unsuccessfullReason
	} else {
		dateStrSplitted = append(dateStrSplitted, strconv.Itoa(yearInt))
	}
	if (monthInt < 1) || (monthInt > 12) {
		unsuccessfullReason = "month not between 1 and 12"
		return dateStr, successfull, unsuccessfullReason
	} else {
		if monthInt < 10 {
			monthStr = "0" + strconv.Itoa(monthInt)
		} else {
			monthStr = strconv.Itoa(monthInt)
		}
		dateStrSplitted = append(dateStrSplitted, monthStr)
	}
	if (dayInt < 1) || (dayInt > 31) {
		unsuccessfullReason = "day not between 1 and 31"
		return dateStr, successfull, unsuccessfullReason
	} else {
		if dayInt < 10 {
			dayStr = "0" + strconv.Itoa(dayInt)
		} else {
			dayStr = strconv.Itoa(dayInt)
		}
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
	if (len(dateStr) > 10) || (len(dateStr) < 8) {
		unsuccessfullReason = "unhandled date length"
		return 0, 0, 0, successfull, unsuccessfullReason
	}
	switch dateSeparator {
	case "-", "/":
		dateStrSplitted = strings.Split(dateStr, dateSeparator)
		if len(dateStrSplitted) != 3 {
			unsuccessfullReason = "not splitted in 3"
			return 0, 0, 0, successfull, unsuccessfullReason
		}
	default:
		unsuccessfullReason = "wrong separator"
		return 0, 0, 0, successfull, unsuccessfullReason
	}
	for _, element := range dateStrSplitted {
		i, err = strconv.Atoi(element)
		if err != nil { // Always check errors even if they should not happen.
			unsuccessfullReason = "str to int"
			return 0, 0, 0, successfull, unsuccessfullReason
		}
		dateIntSplitted = append(dateIntSplitted, i)
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
