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

// WARNING when the 1st reader is used, no other read can occur
// bytedata, _ := io.ReadAll(r.Body)
// fmt.Printf("PostRecordInsert body: %v\n", string(bytedata))
// fmt.Printf("PostRecordInsert Header: %v\n", r.Header)
// r.Body.Close() //  must close
// RESET the body reader
// r.Body = io.NopCloser(bytes.NewBuffer(bytedata))

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
	ft := appdata.FinanceTracker{}
	if err := render.Bind(r, &ft); err != nil {
		fmt.Printf("PostRecordInsert error0: %v\n", err.Error())
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusBadRequest, "invalid request, double check each field", "")
	}
	_, isErr, httpCode, info := handleFTinsert(r, &ft)
	if isErr {
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, httpCode, info, "")
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
	ft.Mode = 0
	ft.Category = "Transfert"
	ft.CategoryDetails.CategoryIcon = "e91b"
	ft.CategoryDetails.CategoryColor = "#999999"
	ft.DateDetails.MonthStr = appdata.MonthIto3A(ft.DateDetails.Month)
	ft.FormPriceStr2Decimals = tr.FormPriceStr2Decimals

	//first part to add to
	ft.PriceDirection = "gain"
	ft.Account = tr.AccountTo
	ft.Product = "Transfert+"
	ftList = append(ftList, ft)

	//second part to remove from
	ft.PriceDirection = "expense"
	ft.Account = tr.AccountFrom
	ft.Product = "Transfert-"
	ftList = append(ftList, ft)

	_, isErr, httpCode, info := handleFTinsert(r, &ftList[1])
	if isErr {
		fmt.Println("PostRecordTransfer error3")
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, httpCode, info, "")
	}
	_, isErr, httpCode, info = handleFTinsert(r, &ftList[0])
	if isErr {
		fmt.Println("PostRecordTransfer error4")
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, httpCode, info, "")
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
	ft.Mode = 0
	ft.DateDetails.Year = rrList[0].DateDetails.Year
	ft.DateDetails.Month = rrList[0].DateDetails.Month
	ft.DateDetails.Day = rrList[0].DateDetails.Day
	ft.Account = rrList[0].Account
	ft.Product = rrList[0].Product
	ft.Category = rrList[0].Category
	if rrList[0].PriceIntx100 < 0 {
		ft.PriceDirection = "expense"
		ft.FormPriceStr2Decimals = rrList[0].FormPriceStr2Decimals[1:] //remove the first "-"
	} else {
		ft.PriceDirection = "gain"
		ft.FormPriceStr2Decimals = rrList[0].FormPriceStr2Decimals
	}
	_, isErr, httpCode, info := handleFTinsert(r, &ft)
	if isErr {
		fmt.Println("RecordRecurrentSave error2")
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, httpCode, info, "")
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

func PostLendOrBorrowRecords(w http.ResponseWriter, r *http.Request, isFrontRequest bool) *appdata.HttpStruct {
	/*
		1. check lenderBorrower, already exist or to create
			- if it's to create, accept only in mode 1 (lend) or 2 (borrow), stop the process otherwise
		2. create row in financeTracker
		3. create row to match financeTracker ID and lenderBorrower ID
	*/
	lb := appdata.LendBorrow{}
	if err := render.Bind(r, &lb); err != nil {
		fmt.Printf("PostLendOrBorrowRecords error0: %v\n", err.Error())
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusBadRequest, "invalid request, double check each field", "")
	}
	lb.FT.Mode = lb.ModeInt
	isErr := sqlite.InsertUpdateInLenderBorrower(r.Context(), appdata.DB, &lb) // 1.
	if isErr {
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusInternalServerError, "error", "")
	}
	idFT, isErr, httpCode, info := handleFTinsert(r, &lb.FT) // 2.
	if isErr {
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, httpCode, info, "")
	}
	lb.FT.ID = int(idFT)
	isErr = sqlite.InsertInSpecificRecordsByMode(r.Context(), appdata.DB, &lb) // 3.
	if isErr {
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusInternalServerError, "error", "")
	}
	return appdata.RenderAPIorUI(w, r, isFrontRequest, false, true, http.StatusCreated, "record saved", lb.FT)
}

func handleFTinsert(r *http.Request, ft *appdata.FinanceTracker) (int64, bool, int, string) {
	// if err := render.Bind(r, ft); err != nil {
	// 	fmt.Printf("handleFTinsert error1: %v\n", err.Error())
	// 	return 0, true, http.StatusBadRequest, "invalid request, double check each field"
	// }
	// fmt.Printf("ft: %#v\n", ft)
	_, err := time.Parse(time.DateOnly, ft.Date)
	if err != nil {
		fmt.Printf("handleFTinsert error2 invalid date: %v\n", err.Error())
		return 0, true, http.StatusBadRequest, "invalid date"
	}
	var successfull bool
	var errStr string
	ft.DateDetails.Year, ft.DateDetails.Month, ft.DateDetails.Day, successfull, errStr = sqlite.ConvertDateStrToInt(ft.Date, "EN", "-")
	if !successfull {
		fmt.Printf("handleFTinsert error3 invalid convert date str to int: %v\n", errStr)
		return 0, true, http.StatusInternalServerError, "server error"
	}
	if (ft.Mode == 1 || ft.Mode == 4) && ft.PriceDirection == "expense" {
		fmt.Println("handleFTinsert error4 invalid PriceDirection and Mode combinaison")
		return 0, true, http.StatusBadRequest, "invalid PriceDirection and Mode combinaison"
	}
	if (ft.Mode == 2 || ft.Mode == 3) && ft.PriceDirection == "gain" {
		fmt.Println("handleFTinsert error5 invalid PriceDirection and Mode combinaison")
		return 0, true, http.StatusBadRequest, "invalid PriceDirection and Mode combinaison"
	}
	if ft.PriceDirection == "" {
		fmt.Println("handleFTinsert error6 invalid PriceDirection")
		return 0, true, http.StatusBadRequest, "invalid PriceDirection"
	} else if ft.PriceDirection == "expense" {
		ft.FormPriceStr2Decimals = "-" + ft.FormPriceStr2Decimals
	}
	if ft.DateChecked == "" {
		ft.DateChecked = "9999-12-31"
	}
	ft.PriceIntx100 = sqlite.ConvertPriceStrToInt(ft.FormPriceStr2Decimals, ".") // always "." as decimal separator from the form
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	ft.GofiID = userContext.GofiID
	idInserted, err := sqlite.InsertRowInFinanceTracker(r.Context(), appdata.DB, ft)
	if err != nil { // Always check errors even if they should not happen.
		fmt.Printf("handleFTinsert error sqlite.InsertRowInFinanceTracker: %v\n", err.Error())
		return 0, true, http.StatusInternalServerError, "server error"
	}
	return idInserted, false, 0, ""
}

func PostLenderBorrowerStateChange(w http.ResponseWriter, r *http.Request, isFrontRequest bool) *appdata.HttpStruct {
	lb := appdata.LenderBorrower{}
	if err := render.Bind(r, &lb); err != nil {
		fmt.Printf("PostLenderBorrowerStateChange error1: %v\n", err.Error())
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusBadRequest, "invalid request, double check each field", "")
	}
	b := sqlite.UpdateStateInLenderBorrower(r.Context(), appdata.DB, &lb)
	if b {
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusInternalServerError, "can't update the state", "")
	}
	return appdata.RenderAPIorUI(w, r, isFrontRequest, false, true, http.StatusOK, "record state changed", "")
}

// TODO voir pour retirer le lien d'une ligne DELETE partie prêt et emprunt lorsqu'une annulation de ligne est faite 
func PostUnlinkLendOrBorrowRecords(w http.ResponseWriter, r *http.Request, isFrontRequest bool) *appdata.HttpStruct {
	/*
		input: list of fT ID
		specificRecordsByMode (delete the row)
		financeTracker (put back mode to 0)
	*/
	idL := appdata.IDlist{}
	if err := render.Bind(r, &idL); err != nil {
		fmt.Printf("PostUnlinkLendOrBorrowRecords error1: %v\n", err.Error())
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusBadRequest, "invalid request, double check each field", "")
	}
	idL.IDlistStr = strings.Split(idL.IDsInOneString, ",")
	var err error
	for _, element := range idL.IDlistStr {
		var idInt int
		idInt, err = strconv.Atoi(element)
		if err != nil { // Always check errors even if they should not happen.
			fmt.Printf("PostUnlinkLendOrBorrowRecords error2: %v\n", err.Error())
			return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusBadRequest, "invalid request, double check each field", "")
		}
		if idInt < 1 { // Always check errors even if they should not happen.
			fmt.Printf("PostUnlinkLendOrBorrowRecords error3: %v\n", element)
			return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusBadRequest, "invalid request, double check each field", "")
		}
		idL.IDlistInt = append(idL.IDlistInt, idInt)
	}
	// fmt.Printf("idL: %#v\n", idL)
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	b := sqlite.DeleteSpecificRecordsByMode(r.Context(), appdata.DB, userContext.GofiID, &idL.IDlistInt)
	if b {
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusInternalServerError, "can't update the state", "")
	}
	b = sqlite.UpdateRowsInFinanceTrackerToMode0(r.Context(), appdata.DB, userContext.GofiID, &idL.IDlistInt)
	if b {
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusInternalServerError, "can't update the state", "")
	}
	return appdata.RenderAPIorUI(w, r, isFrontRequest, false, true, http.StatusOK, "records unlinked", "")
}
