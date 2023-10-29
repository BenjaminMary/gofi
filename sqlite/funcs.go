package sqlite

import (
	// "fmt"
	"strconv"
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