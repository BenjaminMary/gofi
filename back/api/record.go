package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gofi/gofi/data/appdata"
	"gofi/gofi/data/sqlite"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func GetRecords(w http.ResponseWriter, r *http.Request, isFrontRequest bool) *appdata.HttpStruct {
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	var filter appdata.FilterRows
	filter.GofiID = userContext.GofiID

	orderBy := chi.URLParam(r, "orderby")
	if orderBy == "" {
		filter.OrderBy = "id"
	} else {
		filter.OrderBy = orderBy
	}
	orderSort := chi.URLParam(r, "ordersort")
	if orderSort == "" {
		filter.OrderSort = "DESC"
	} else {
		filter.OrderSort = strings.ToUpper(orderSort)
	}
	limitStr := chi.URLParam(r, "limit")
	if limitStr == "" {
		filter.Limit = 5
	} else {
		limitInt, err := strconv.Atoi(limitStr)
		if err != nil {
			fmt.Printf("GetRecords err1: %v\n", err)
			return appdata.RenderAPIorUI(w, r, isFrontRequest, true, false, http.StatusBadRequest, "invalid request, double check each field", "")
		}
		filter.Limit = limitInt
	}
	//fmt.Printf("filter: %#v\n", filter)
	var ftList []appdata.FinanceTracker
	ftList, _, _ = sqlite.GetRowsInFinanceTracker(r.Context(), appdata.DB, &filter)
	return appdata.RenderAPIorUI(w, r, isFrontRequest, true, true, http.StatusOK, "record list retrieved", ftList)
}
func GetRecordsViaPost(w http.ResponseWriter, r *http.Request, isFrontRequest bool, filterR *appdata.FilterRows) *appdata.HttpStruct {
	// filter := appdata.FilterRows{OrderBy: "date", OrderSort: "asc", LimitStr: "8", WhereCheckedStr: "2"}
	var err error
	if r.Method == "POST" {
		if err = render.Bind(r, filterR); err != nil {
			fmt.Printf("GetRecordsViaPost error1: %v\n", err.Error())
			return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusBadRequest, "invalid request, double check each field", "")
		}
	}
	var tInt int
	if filterR.LimitStr == "8" {
		filterR.Limit = 8
	} else {
		tInt = 0
		tInt, err = strconv.Atoi(filterR.LimitStr)
		if err != nil || tInt == 0 {
			if err != nil {
				fmt.Printf("GetRecordsViaPost err2: %v\n", err)
			} else {
				fmt.Println("GetRecordsViaPost limit 0")
			}
			return appdata.RenderAPIorUI(w, r, isFrontRequest, true, false, http.StatusBadRequest, "invalid request, wrong limit", "")
		}
		filterR.Limit = tInt
	}
	if filterR.WhereYearStr == "" {
		filterR.WhereYear = 0
	} else {
		tInt = 0
		tInt, err := strconv.Atoi(filterR.WhereYearStr)
		if err != nil || tInt == 0 {
			if err != nil {
				fmt.Printf("GetRecordsViaPost err3: %v\n", err)
			} else {
				fmt.Println("GetRecordsViaPost year 0")
			}
			return appdata.RenderAPIorUI(w, r, isFrontRequest, true, false, http.StatusBadRequest, "invalid request, wrong year", "")
		}
		filterR.WhereYear = tInt
	}
	if filterR.WhereMonthStr == "" {
		filterR.WhereMonth = 0
	} else {
		tInt = 0
		tInt, err := strconv.Atoi(filterR.WhereMonthStr)
		if err != nil || tInt == 0 {
			if err != nil {
				fmt.Printf("GetRecordsViaPost err4: %v\n", err)
			} else {
				fmt.Println("GetRecordsViaPost month 0")
			}
			return appdata.RenderAPIorUI(w, r, isFrontRequest, true, false, http.StatusBadRequest, "invalid request, wrong month", "")
		}
		filterR.WhereMonth = tInt
	}
	if filterR.WhereCheckedStr == "2" {
		filterR.WhereChecked = 2
	} else {
		tInt = -1
		tInt, err := strconv.Atoi(filterR.WhereCheckedStr)
		if err != nil || tInt == -1 {
			if err != nil {
				fmt.Printf("GetRecordsViaPost err5: %v\n", err)
			} else {
				fmt.Println("GetRecordsViaPost row checker -1")
			}
			return appdata.RenderAPIorUI(w, r, isFrontRequest, true, false, http.StatusBadRequest, "invalid request, wrong row checker", "")
		}
		filterR.WhereChecked = tInt
	}
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	filterR.GofiID = userContext.GofiID
	var ftList []appdata.FinanceTracker
	var totalRowsWithoutLimit int
	// fmt.Printf("filter: %#v\n", filterR)
	ftList, _, totalRowsWithoutLimit = sqlite.GetRowsInFinanceTracker(r.Context(), appdata.DB, filterR)
	r.Header.Set("totalRowsWithoutLimit", strconv.Itoa(totalRowsWithoutLimit))
	return appdata.RenderAPIorUI(w, r, isFrontRequest, true, true, http.StatusOK, "record list retrieved", ftList)
}

func PostRecordInsert(w http.ResponseWriter, r *http.Request, isFrontRequest bool) *appdata.HttpStruct {
	// WARNING when the 1st reader is used, no other read can occur
	// bytedata, _ := io.ReadAll(r.Body)
	// fmt.Printf("PostRecordInsert body: %v\n", string(bytedata))
	// fmt.Printf("PostRecordInsert Header: %v\n", r.Header)

	ft := appdata.FinanceTracker{}
	if err := render.Bind(r, &ft); err != nil {
		fmt.Printf("PostRecordInsert error1: %v\n", err.Error())
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusBadRequest, "invalid request, double check each field", "")
	}
	// check if valid date
	_, err := time.Parse(time.DateOnly, ft.Date)
	if err != nil {
		fmt.Printf("PostRecordInsert error2: %v\n", err.Error())
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusBadRequest, "invalid date", "")
	}
	var successfull bool
	var errStr string
	ft.DateDetails.Year, ft.DateDetails.Month, ft.DateDetails.Day, successfull, errStr = sqlite.ConvertDateStrToInt(ft.Date, "EN", "-")
	if !successfull {
		fmt.Printf("PostRecordInsert error3: %v\n", errStr)
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusInternalServerError, "server error", "")
	}
	if ft.PriceDirection == "expense" {
		ft.FormPriceStr2Decimals = "-" + ft.FormPriceStr2Decimals
	}
	ft.PriceIntx100 = sqlite.ConvertPriceStrToInt(ft.FormPriceStr2Decimals, ".") // always "." as decimal separator from the form
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	ft.GofiID = userContext.GofiID
	_, err = sqlite.InsertRowInFinanceTracker(r.Context(), appdata.DB, ft)
	if err != nil { // Always check errors even if they should not happen.
		fmt.Printf("PostRecordInsert error4: %v\n", err.Error())
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusInternalServerError, "server error", "")
	}
	return appdata.RenderAPIorUI(w, r, isFrontRequest, false, true, http.StatusCreated, "record saved", ft)
}

// POST Transfer.html
func PostRecordTransfer(w http.ResponseWriter, r *http.Request, isFrontRequest bool) *appdata.HttpStruct {
	tr := &appdata.Transfer{}
	if err := render.Bind(r, tr); err != nil {
		fmt.Printf("PostRecordTransfer error1: %v\n", err.Error())
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusBadRequest, "invalid request, double check each field", "")
	}
	// check if valid date
	_, err := time.Parse(time.DateOnly, tr.Date)
	if err != nil {
		fmt.Printf("PostRecordTransfer error2: %v\n", err.Error())
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusBadRequest, "invalid date", "")
	}
	ftList := []appdata.FinanceTracker{}
	ft := appdata.FinanceTracker{}
	ft.Date = tr.Date
	var successfull bool
	ft.DateDetails.Year, ft.DateDetails.Month, ft.DateDetails.Day, successfull, _ = sqlite.ConvertDateStrToInt(ft.Date, "EN", "-")
	if !successfull {
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusBadRequest, "invalid request, double check each field", "")
	}
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	ft.GofiID = userContext.GofiID
	ft.Category = "Transfert"
	ft.CategoryDetails.CategoryIcon = "e91b"
	ft.CategoryDetails.CategoryColor = "#999999"
	ft.DateDetails.MonthStr = appdata.MonthIto3A(ft.DateDetails.Month)

	//first part to add to
	ft.FormPriceStr2Decimals = tr.FormPriceStr2Decimals
	ft.PriceIntx100 = sqlite.ConvertPriceStrToInt(ft.FormPriceStr2Decimals, ".") // always "." as decimal separator from the form
	ft.Account = tr.AccountTo
	ft.Product = "Transfert+"
	ftList = append(ftList, ft)

	//second part to remove from
	ft.Account = tr.AccountFrom
	ft.Product = "Transfert-"
	ft.FormPriceStr2Decimals = "-" + ft.FormPriceStr2Decimals
	ft.PriceIntx100 = sqlite.ConvertPriceStrToInt(ft.FormPriceStr2Decimals, ".") // always "." as decimal separator from the form
	ftList = append(ftList, ft)

	// insert the amount to remove from the first account
	_, err = sqlite.InsertRowInFinanceTracker(r.Context(), appdata.DB, ftList[1])
	if err != nil { // Always check errors even if they should not happen.
		fmt.Printf("PostRecordTransfer error3: %v\n", err.Error())
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusInternalServerError, "server error", "")
	}
	// insert the amount to add to the second account
	_, err = sqlite.InsertRowInFinanceTracker(r.Context(), appdata.DB, ftList[0])
	if err != nil { // Always check errors even if they should not happen.
		fmt.Printf("PostRecordTransfer error4: %v\n", err.Error())
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusInternalServerError, "server error", "")
	}
	return appdata.RenderAPIorUI(w, r, isFrontRequest, false, true, http.StatusCreated, "transfer done", ftList)
}

// GET RecurrentRecords.html
func RecordRecurrentRead(w http.ResponseWriter, r *http.Request, isFrontRequest bool) *appdata.HttpStruct {
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	rrList := sqlite.GetRowsInRecurrentRecord(r.Context(), appdata.DB, userContext.GofiID, 0)
	if len(rrList) == 0 {
		fmt.Printf("RecordRecurrentRead empty list")
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusBadRequest, "invalid request, double check each field", "")
	}
	return appdata.RenderAPIorUI(w, r, isFrontRequest, false, true, http.StatusOK, "recurrent record selected", rrList)
}

// POST RecurrentRecords.html
func RecordRecurrentCreate(w http.ResponseWriter, r *http.Request, isFrontRequest bool) *appdata.HttpStruct {
	var rr appdata.RecurrentRecord
	if err := render.Bind(r, &rr); err != nil {
		fmt.Printf("PostCreateRecurrentRecords error1: %v\n", err.Error())
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusBadRequest, "invalid request, double check each field", "")
	}
	// check if valid date
	_, err := time.Parse(time.DateOnly, rr.Date)
	if err != nil {
		fmt.Printf("PostCreateRecurrentRecords error2: %v\n", err.Error())
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusBadRequest, "invalid date", "")
	}
	var successfull bool
	rr.DateDetails.Year, rr.DateDetails.Month, rr.DateDetails.Day, successfull, _ = sqlite.ConvertDateStrToInt(rr.Date, "EN", "-")
	if !successfull {
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusBadRequest, "invalid request, double check each field", "")
	}
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	rr.GofiID = userContext.GofiID
	if rr.PriceDirection == "expense" {
		rr.FormPriceStr2Decimals = "-" + rr.FormPriceStr2Decimals
	}
	rr.PriceIntx100 = sqlite.ConvertPriceStrToInt(rr.FormPriceStr2Decimals, ".") // always "." as decimal separator from the form
	id, err := sqlite.InsertRowInRecurrentRecord(r.Context(), appdata.DB, &rr)
	if err != nil { // Always check errors even if they should not happen.
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusInternalServerError, "server error", "")
	}
	rr.ID = int(id)
	return appdata.RenderAPIorUI(w, r, isFrontRequest, false, true, http.StatusCreated, "recurrent record created", rr)
}

func BindRecordRecurrent(r *http.Request, rr *appdata.RecurrentRecord) {
	if err := render.Bind(r, rr); err != nil {
		fmt.Printf("BindRecordRecurrent error1: %v\n", err.Error())
	}
}

func RecordRecurrentUpdate(w http.ResponseWriter, r *http.Request, isFrontRequest bool) *appdata.HttpStruct {
	var rr appdata.RecurrentRecord
	var err error
	if err := render.Bind(r, &rr); err != nil {
		fmt.Printf("RecordRecurrentUpdate error1: %v\n", err.Error())
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusBadRequest, "invalid request, double check each field", "")
	}
	// check if valid date
	_, err = time.Parse(time.DateOnly, rr.Date)
	if err != nil {
		fmt.Printf("RecordRecurrentUpdate error2: %v\n", err.Error())
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusBadRequest, "invalid date", "")
	}
	rr.ID, err = strconv.Atoi(rr.IDstr)
	if err != nil { // Always check errors even if they should not happen.
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusBadRequest, "invalid request, double check each field", "")
	}
	var successfull bool
	rr.DateDetails.Year, rr.DateDetails.Month, rr.DateDetails.Day, successfull, _ = sqlite.ConvertDateStrToInt(rr.Date, "EN", "-")
	if !successfull {
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusBadRequest, "invalid request, double check each field", "")
	}
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	rr.GofiID = userContext.GofiID
	if rr.PriceDirection == "expense" {
		rr.FormPriceStr2Decimals = "-" + rr.FormPriceStr2Decimals
	}
	rr.PriceIntx100 = sqlite.ConvertPriceStrToInt(rr.FormPriceStr2Decimals, ".") // always "." as decimal separator from the form
	rowsAffected, err := sqlite.UpdateRowInRecurrentRecord(r.Context(), appdata.DB, &rr)
	if err != nil { // Always check errors even if they should not happen.
		fmt.Printf("RecordRecurrentUpdate error3: %v\n", err.Error())
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusInternalServerError, "server error", "")
	}
	if rowsAffected != 1 {
		fmt.Printf("RecordRecurrentUpdate rowsAffected: %v\n", rowsAffected)
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusBadRequest, "invalid request, double check each field", "")
	}
	return appdata.RenderAPIorUI(w, r, isFrontRequest, false, true, http.StatusOK, "recurrent record updated", rr)
}

func RecordRecurrentDelete(w http.ResponseWriter, r *http.Request, isFrontRequest bool, idrrFuncParam string) *appdata.HttpStruct {
	idrr := getURLorFUNCparam(r, idrrFuncParam, "idrr")
	var rr appdata.RecurrentRecord
	var err error
	rr.ID, err = strconv.Atoi(idrr)
	if err != nil { // Always check errors even if they should not happen.
		fmt.Printf("RecordRecurrentDelete error1: %v\n", err.Error())
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusBadRequest, "invalid request, double check each field", "")
	}
	if rr.ID < 1 { // Always check errors even if they should not happen.
		fmt.Printf("RecordRecurrentDelete rr.ID: %v\n", rr.ID)
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusBadRequest, "invalid request, double check each field", "")
	}
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	rr.GofiID = userContext.GofiID
	rowsAffected, err := sqlite.DeleteRowInRecurrentRecord(r.Context(), appdata.DB, &rr)
	if err != nil { // Always check errors even if they should not happen.
		fmt.Printf("RecordRecurrentDelete error3: %v\n", err.Error())
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusBadRequest, "invalid request, double check each field", "")
	}
	if rowsAffected != 1 {
		fmt.Printf("RecordRecurrentDelete rowsAffected: %v\n", rowsAffected)
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusBadRequest, "invalid request, double check each field", "")
	}
	return appdata.RenderAPIorUI(w, r, isFrontRequest, false, true, http.StatusOK, "recurrent record deleted", rr)
}

func RecordRecurrentSave(w http.ResponseWriter, r *http.Request, isFrontRequest bool) *appdata.HttpStruct {
	var rrs appdata.RecurrentRecordSave
	if err := render.Bind(r, &rrs); err != nil {
		fmt.Printf("RecordRecurrentSave error1: %v\n", err.Error())
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusBadRequest, "invalid request, double check each field", "")
	}
	rowID, _ := strconv.Atoi(rrs.ID)
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	rrList := sqlite.GetRowsInRecurrentRecord(r.Context(), appdata.DB, userContext.GofiID, rowID)
	if len(rrList) == 0 {
		fmt.Println("RecordRecurrentSave empty list error")
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusBadRequest, "invalid request, double check each field", "")
	}
	rrList[0].GofiID = userContext.GofiID

	var ft appdata.FinanceTracker
	// ft.ID = rrList[0].ID
	ft.GofiID = userContext.GofiID
	ft.Date = rrList[0].Date
	ft.DateDetails.Year = rrList[0].DateDetails.Year
	ft.DateDetails.Month = rrList[0].DateDetails.Month
	ft.DateDetails.Day = rrList[0].DateDetails.Day
	ft.Account = rrList[0].Account
	ft.Product = rrList[0].Product
	ft.FormPriceStr2Decimals = rrList[0].FormPriceStr2Decimals
	ft.PriceIntx100 = rrList[0].PriceIntx100
	ft.Category = rrList[0].Category

	_, err := sqlite.InsertRowInFinanceTracker(r.Context(), appdata.DB, ft)
	if err != nil { // Always check errors even if they should not happen.
		fmt.Printf("RecordRecurrentSave error2: %v\n", err.Error())
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusInternalServerError, "server error", "")
	}
	// 1 reinit date in EN- format, 2 compute new date, 3 extract YYYYMMDD
	rrList[0].Date, _, _ = sqlite.ConvertDateIntToStr(rrList[0].DateDetails.Year, rrList[0].DateDetails.Month, rrList[0].DateDetails.Day, "EN", "-")
	baseDate, err := time.Parse(time.DateOnly, rrList[0].Date)
	if err != nil { // Always check errors even if they should not happen.
		fmt.Printf("RecordRecurrentSave error3: %v\n", err.Error())
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusBadRequest, "invalid request, double check each field", "")
	}
	switch rrList[0].Recurrence {
	case "mensuelle":
		addMonthRecalc(&baseDate, 0)
		rrList[0].Date = baseDate.Format(time.DateOnly) // Add Y,M,D
	case "hebdomadaire":
		rrList[0].Date = baseDate.AddDate(0, 0, 7).Format(time.DateOnly) // Add Y,M,D
	case "annuelle":
		rrList[0].Date = baseDate.AddDate(1, 0, 0).Format(time.DateOnly) // Add Y,M,D
	default:
		fmt.Printf("RecordRecurrentSave error4 on switch case Recurrence: %#v\n", rrList[0].Recurrence)
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusBadRequest, "invalid request, double check each field", "")
	}
	var successfull bool
	rrList[0].DateDetails.Year, rrList[0].DateDetails.Month, rrList[0].DateDetails.Day, successfull, _ = sqlite.ConvertDateStrToInt(rrList[0].Date, "EN", "-")
	if !successfull {
		fmt.Println("RecordRecurrentSave error5")
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusBadRequest, "invalid request, double check each field", "")
	}
	sqlite.UpdateDateInRecurrentRecord(r.Context(), appdata.DB, &rrList[0])
	return appdata.RenderAPIorUI(w, r, isFrontRequest, false, true, http.StatusCreated, "recurrent record saved", ft)
}

func addMonthRecalc(baseDate *time.Time, dayOffset int) {
	// with this recursive func, 2011-05-31 +1 month = 2011-06-30, would be 2011-07-01 without
	newDate := baseDate.AddDate(0, 1, dayOffset)
	// fmt.Printf("in funcA: %#v\n", newDate.Format("2006-01-02"))
	if dayOffset < -40 {
		fmt.Println("addMonthRecalc infinite loop stop")
	} else if baseDate.Day()-5 > newDate.Day() {
		addMonthRecalc(baseDate, (dayOffset - 1))
	} else {
		*baseDate = newDate
	}
}

func recordValidateOrCancel(w http.ResponseWriter, r *http.Request, isFrontRequest bool, rvc *appdata.RecordValidateOrCancel) *appdata.HttpStruct {
	if err := render.Bind(r, rvc); err != nil {
		fmt.Printf("RecordValidateOrCancel error1: %v\n", err.Error())
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusBadRequest, "invalid request, double check each field", "")
	}
	_, err := time.Parse(time.DateOnly, rvc.Date)
	if err != nil { // Always check errors even if they should not happen.
		fmt.Printf("RecordValidateOrCancel error2: %v\n", err.Error())
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusBadRequest, "invalid request, wrong date", "")
	}
	checkedListStr := strings.Split(rvc.IDcheckedListStr, ",")
	for _, strValue := range checkedListStr {
		intValue, err := strconv.Atoi(strValue)
		if err != nil { // Always check errors even if they should not happen.
			fmt.Printf("RecordValidateOrCancel error3: %v\n", err.Error())
			return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusBadRequest, "invalid request, wrong id list", "")
		}
		rvc.IDcheckedListInt = append(rvc.IDcheckedListInt, intValue)
	}
	if len(rvc.IDcheckedListInt) < 1 {
		fmt.Printf("RecordValidateOrCancel nb list: %v\n", len(rvc.IDcheckedListInt))
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusBadRequest, "invalid request, wrong id list", "")
	}
	return &appdata.HttpStruct{IsValidResponse: true}
}
func RecordValidate(w http.ResponseWriter, r *http.Request, isFrontRequest bool) *appdata.HttpStruct {
	var rvc appdata.RecordValidateOrCancel
	httpStruct := recordValidateOrCancel(w, r, isFrontRequest, &rvc)
	if !httpStruct.IsValidResponse {
		return httpStruct
	}
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	//send the list of validated id with the date to SQLite for change
	sqlite.ValidateRowsInFinanceTracker(r.Context(), appdata.DB, userContext.GofiID, rvc.IDcheckedListInt, rvc.Date, "validate")
	return appdata.RenderAPIorUI(w, r, isFrontRequest, false, true, http.StatusOK, "id list validated", rvc.IDcheckedListStr)
}
func RecordCancel(w http.ResponseWriter, r *http.Request, isFrontRequest bool) *appdata.HttpStruct {
	var rvc appdata.RecordValidateOrCancel
	httpStruct := recordValidateOrCancel(w, r, isFrontRequest, &rvc)
	if !httpStruct.IsValidResponse {
		return httpStruct
	}
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	//send the list of validated id with the date to SQLite for change
	sqlite.ValidateRowsInFinanceTracker(r.Context(), appdata.DB, userContext.GofiID, rvc.IDcheckedListInt, rvc.Date, "cancel")
	return appdata.RenderAPIorUI(w, r, isFrontRequest, false, true, http.StatusOK, "id list canceled", rvc.IDcheckedListStr)
}
