package appdata

import (
	"strconv"
	"strings"
)

func MonthIto3A(monthI int) string {
	switch monthI {
	case 1:
		return "jan"
	case 2:
		return "fev"
	case 3:
		return "mar"
	case 4:
		return "avr"
	case 5:
		return "mai"
	case 6:
		return "juin"
	case 7:
		return "juil"
	case 8:
		return "aou"
	case 9:
		return "sep"
	case 10:
		return "oct"
	case 11:
		return "nov"
	case 12:
		return "dec"
	default:
		return "---"
	}
}

func ParseDateSVGfront(dateStr string, dateSVGfront *DateDetails) {
	if (len(dateStr) > 10) || (len(dateStr) < 8) {
		return
	}
	dateStrSplitted := strings.Split(dateStr, "-")
	if len(dateStrSplitted) != 3 {
		return
	}
	var err error
	dateSVGfront.Year, err = strconv.Atoi(dateStrSplitted[0])
	if err != nil { // Always check errors even if they should not happen.
		return
	}
	dateSVGfront.Month, err = strconv.Atoi(dateStrSplitted[1])
	if err != nil { // Always check errors even if they should not happen.
		return
	}
	dateSVGfront.Day, err = strconv.Atoi(dateStrSplitted[2])
	if err != nil { // Always check errors even if they should not happen.
		return
	}
	switch dateSVGfront.Month {
	case 1:
		dateSVGfront.MonthStr = "jan"
	case 2:
		dateSVGfront.MonthStr = "fev"
	case 3:
		dateSVGfront.MonthStr = "mar"
	case 4:
		dateSVGfront.MonthStr = "avr"
	case 5:
		dateSVGfront.MonthStr = "mai"
	case 6:
		dateSVGfront.MonthStr = "juin"
	case 7:
		dateSVGfront.MonthStr = "juil"
	case 8:
		dateSVGfront.MonthStr = "aou"
	case 9:
		dateSVGfront.MonthStr = "sep"
	case 10:
		dateSVGfront.MonthStr = "oct"
	case 11:
		dateSVGfront.MonthStr = "nov"
	case 12:
		dateSVGfront.MonthStr = "dec"
	default:
		dateSVGfront.MonthStr = "---"
	}
}
