package front

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"slices"

	"gofi/gofi/back/api"
	"gofi/gofi/data/appdata"
	"gofi/gofi/data/sqlite"
	"gofi/gofi/front/htmlComponents"

	"github.com/go-chi/chi/v5"
)

func TemplIndex(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("in front TemplIndex")
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	htmlComponents.IndexHtmlContent(userContext).Render(r.Context(), w)
}
func Checklist(w http.ResponseWriter, r *http.Request) {
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	json := api.GetParam(w, r, true, "", "", false)
	jsonUserParam := json.AnyStruct.(appdata.UserParams)
	// fmt.Printf("jsonUserParam: %#v\n", jsonUserParam)
	htmlComponents.Checklist(userContext, &jsonUserParam).Render(r.Context(), w)
}
func Checklist1(w http.ResponseWriter, r *http.Request) {
	strNumber := "1"
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	json := api.GetParam(w, r, true, "", "", false)
	jsonUserParam := json.AnyStruct.(appdata.UserParams)
	if !slices.Contains(jsonUserParam.OnboardingCheckList, strNumber) {
		api.UpdateParamCheckList(w, r, true, userContext.GofiID, (jsonUserParam.OnboardingCheckListSingleString + "," + strNumber))
	}
	// fmt.Printf("jsonUserParam: %#v\n", jsonUserParam)
	htmlComponents.Checklist1(userContext, &jsonUserParam).Render(r.Context(), w)
}
func Checklist2(w http.ResponseWriter, r *http.Request) {
	strNumber := "2"
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	json := api.GetParam(w, r, true, "", "", false)
	jsonUserParam := json.AnyStruct.(appdata.UserParams)
	if !slices.Contains(jsonUserParam.OnboardingCheckList, strNumber) {
		api.UpdateParamCheckList(w, r, true, userContext.GofiID, (jsonUserParam.OnboardingCheckListSingleString + "," + strNumber))
	}
	// fmt.Printf("jsonUserParam: %#v\n", jsonUserParam)
	htmlComponents.Checklist2(userContext, &jsonUserParam).Render(r.Context(), w)
}
func Checklist3(w http.ResponseWriter, r *http.Request) {
	strNumber := "3"
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	json := api.GetParam(w, r, true, "", "", false)
	jsonUserParam := json.AnyStruct.(appdata.UserParams)
	if !slices.Contains(jsonUserParam.OnboardingCheckList, strNumber) {
		api.UpdateParamCheckList(w, r, true, userContext.GofiID, (jsonUserParam.OnboardingCheckListSingleString + "," + strNumber))
	}
	// fmt.Printf("jsonUserParam.Categories: %#v\n", jsonUserParam.Categories)
	htmlComponents.Checklist3(userContext, &jsonUserParam).Render(r.Context(), w)
}
func Checklist4(w http.ResponseWriter, r *http.Request) {
	strNumber := "4"
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	json := api.GetParam(w, r, true, "", "", false)
	jsonUserParam := json.AnyStruct.(appdata.UserParams)
	if !slices.Contains(jsonUserParam.OnboardingCheckList, strNumber) {
		api.UpdateParamCheckList(w, r, true, userContext.GofiID, (jsonUserParam.OnboardingCheckListSingleString + "," + strNumber))
	}
	// fmt.Printf("jsonUserParam: %#v\n", jsonUserParam)
	htmlComponents.Checklist4(userContext, &jsonUserParam).Render(r.Context(), w)
}
func Checklist5(w http.ResponseWriter, r *http.Request) {
	strNumber := "5"
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	json := api.GetParam(w, r, true, "", "", false)
	jsonUserParam := json.AnyStruct.(appdata.UserParams)
	if !slices.Contains(jsonUserParam.OnboardingCheckList, strNumber) {
		api.UpdateParamCheckList(w, r, true, userContext.GofiID, (jsonUserParam.OnboardingCheckListSingleString + "," + strNumber))
	}
	// fmt.Printf("jsonUserParam: %#v\n", jsonUserParam)
	htmlComponents.Checklist5(userContext, &jsonUserParam).Render(r.Context(), w)
}
func Checklist6(w http.ResponseWriter, r *http.Request) {
	strNumber := "6"
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	json := api.GetParam(w, r, true, "", "", false)
	jsonUserParam := json.AnyStruct.(appdata.UserParams)
	if !slices.Contains(jsonUserParam.OnboardingCheckList, strNumber) {
		api.UpdateParamCheckList(w, r, true, userContext.GofiID, (jsonUserParam.OnboardingCheckListSingleString + "," + strNumber))
	}
	// fmt.Printf("jsonUserParam: %#v\n", jsonUserParam)
	htmlComponents.Checklist6(userContext, &jsonUserParam).Render(r.Context(), w)
}
func Checklist7(w http.ResponseWriter, r *http.Request) {
	strNumber := "7"
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	json := api.GetParam(w, r, true, "", "", false)
	jsonUserParam := json.AnyStruct.(appdata.UserParams)
	if !slices.Contains(jsonUserParam.OnboardingCheckList, strNumber) {
		api.UpdateParamCheckList(w, r, true, userContext.GofiID, (jsonUserParam.OnboardingCheckListSingleString + "," + strNumber))
	}
	// fmt.Printf("jsonUserParam: %#v\n", jsonUserParam)
	htmlComponents.Checklist7(userContext, &jsonUserParam).Render(r.Context(), w)
}
func Checklist8(w http.ResponseWriter, r *http.Request) {
	strNumber := "8"
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	json := api.GetParam(w, r, true, "", "", false)
	jsonUserParam := json.AnyStruct.(appdata.UserParams)
	if !slices.Contains(jsonUserParam.OnboardingCheckList, strNumber) {
		api.UpdateParamCheckList(w, r, true, userContext.GofiID, (jsonUserParam.OnboardingCheckListSingleString + "," + strNumber))
	}
	// fmt.Printf("jsonUserParam: %#v\n", jsonUserParam)
	htmlComponents.Checklist8(userContext, &jsonUserParam).Render(r.Context(), w)
}
func Lost(w http.ResponseWriter, r *http.Request) {
	// TODO: return a 404
	htmlComponents.Lost().Render(r.Context(), w)
}

// USER
func GetCreateUser(w http.ResponseWriter, r *http.Request) {
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	// fmt.Printf("RequestURI: %v\n", r.RequestURI)
	htmlComponents.GetCreateUser(userContext).Render(r.Context(), w)
}
func PostCreateUser(w http.ResponseWriter, r *http.Request) {
	json := api.UserCreate(w, r, true)
	// fmt.Printf("PostCreateUser: %#v\n", json)
	htmlComponents.PostCreateUser(json.HttpStatus).Render(r.Context(), w)
}
func GetLogin(w http.ResponseWriter, r *http.Request) {
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	htmlComponents.GetLogin(userContext).Render(r.Context(), w)
}
func PostLogin(w http.ResponseWriter, r *http.Request) {
	json := api.UserLogin(w, r, true)
	if json.IsValidResponse {
		jsonUser := json.AnyStruct.(*appdata.User)
		FrontSetCookie(w, jsonUser.SessionID)
	}
	htmlComponents.PostLogin(json).Render(r.Context(), w)
}
func GetLogout(w http.ResponseWriter, r *http.Request) {
	api.UserLogout(w, r, true)
	htmlComponents.GetLogout().Render(r.Context(), w)
}

// PARAM
func GetParam(w http.ResponseWriter, r *http.Request) {
	json := api.GetParam(w, r, true, "", "", false)
	jsonUserParam := json.AnyStruct.(appdata.UserParams)
	htmlComponents.GetParam(jsonUserParam).Render(r.Context(), w)
}
func GetParamAccount(w http.ResponseWriter, r *http.Request) {
	json := api.GetParam(w, r, true, "", "", false)
	jsonUserParam := json.AnyStruct.(appdata.UserParams)
	htmlComponents.GetParamAccount(jsonUserParam).Render(r.Context(), w)
}
func PostParamAccount(w http.ResponseWriter, r *http.Request) {
	json := api.PostParamAccount(w, r, true)
	jsonParam := &appdata.Param{}
	if json.IsValidResponse {
		jsonParam = json.AnyStruct.(*appdata.Param)
	}
	jsonb := api.GetParam(w, r, true, "", "", false)
	jsonUserParam := jsonb.AnyStruct.(appdata.UserParams)
	htmlComponents.PostParamAccount(json.HttpStatus, jsonParam.ParamJSONstringData, jsonUserParam.AccountList).Render(r.Context(), w)
}
func PostParamCategoryRendering(w http.ResponseWriter, r *http.Request) {
	json := api.PostParamCategoryRendering(w, r, true)
	jsonParam := &appdata.Param{}
	if json.IsValidResponse {
		jsonParam = json.AnyStruct.(*appdata.Param)
	}
	htmlComponents.PostParamCategoryRendering(json.HttpStatus, jsonParam.ParamJSONstringData).Render(r.Context(), w)
}
func PostParamOnboardingCheckList(w http.ResponseWriter, r *http.Request) {
	json := api.PostParamOnboardingCheckList(w, r, true)
	jsonParam := &appdata.Param{}
	if json.IsValidResponse {
		jsonParam = json.AnyStruct.(*appdata.Param)
	}
	htmlComponents.PostParamOnboardingCheckList(json.HttpStatus, jsonParam.ParamJSONstringData).Render(r.Context(), w)
}

func GetParamCategory(w http.ResponseWriter, r *http.Request) {
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	userCategories := appdata.NewUserCategories()
	userCategories.GofiID = userContext.GofiID
	sqlite.GetFullCategoryList(r.Context(), appdata.DB, userCategories, "", "", false)
	userCategoriesJson, err := json.Marshal(userCategories)
	if err != nil {
		fmt.Println(err)
	}
	unhandledCategoryList := sqlite.GetUnhandledCategoryList(r.Context(), appdata.DB, userCategories.GofiID)
	htmlComponents.GetParamCategory(userCategories, string(userCategoriesJson), unhandledCategoryList).Render(r.Context(), w)
}
func PutParamCategory(w http.ResponseWriter, r *http.Request) {
	api.PutParamCategory(w, r, true)
	htmlComponents.PutParamCategory().Render(r.Context(), w)
}
func PatchParamCategoryInUse(w http.ResponseWriter, r *http.Request) {
	api.PatchParamCategoryInUse(w, r, true)
	htmlComponents.PatchParamCategoryInUse().Render(r.Context(), w)
}
func PatchParamCategoryOrder(w http.ResponseWriter, r *http.Request) {
	api.PatchParamCategoryOrder(w, r, true)
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	userCategories := appdata.NewUserCategories()
	userCategories.GofiID = userContext.GofiID
	sqlite.GetFullCategoryList(r.Context(), appdata.DB, userCategories, "", "", false)
	htmlComponents.PatchParamCategoryOrder(userCategories).Render(r.Context(), w)
}

// RECORD
func GetRecordInsert(w http.ResponseWriter, r *http.Request) {
	jsonFT := api.GetRecords(w, r, true)
	jsonFTlist := jsonFT.AnyStruct.([]appdata.FinanceTracker)
	jsonUP := api.GetParam(w, r, true, "type", "basic", false)
	jsonUserParam := jsonUP.AnyStruct.(appdata.UserParams)
	currentTime := time.Now()
	currentDate := currentTime.Format(time.DateOnly) // YYYY-MM-DD
	jsonFrontURLdefaultValues := appdata.FrontURLdefaultValues{UseNewDefaultValues: false, CategoryNumber: 0, PriceDirection: "expense"}
	urlParams := chi.URLParam(r, "*")
	if urlParams != "" {
		listURLparams := strings.Split(urlParams, "/")
		if len(listURLparams) != 5 {
			fmt.Println("error len!=5 in url params")
		} else {
			jsonFrontURLdefaultValues.UseNewDefaultValues = true
			jsonFrontURLdefaultValues.Account = listURLparams[0]
			jsonFrontURLdefaultValues.Category = listURLparams[1]
			jsonFrontURLdefaultValues.CategoryNumber = jsonUserParam.Categories.FindCategory[listURLparams[1]]
			jsonFrontURLdefaultValues.Product = listURLparams[2]
			if listURLparams[3] == "+" || listURLparams[3] == "gain" {
				// default "expense"
				jsonFrontURLdefaultValues.PriceDirection = "gain"
			}
			jsonFrontURLdefaultValues.FormPriceStr2Decimals = listURLparams[4]
		}
		// /{account}/{category}/{product}/{priceDirection}/{price}
		// urlsafe special characters: $-_.+!*'(),
		// http://localhost:8083/record/insert/LA/Epargne/desi/+/56.78
		// fmt.Printf("jsonFrontURLdefaultValues: %#v\n", jsonFrontURLdefaultValues)
	}
	htmlComponents.GetRecordInsert(jsonFTlist, jsonUserParam, currentDate, jsonFrontURLdefaultValues).Render(r.Context(), w)
}
func PostRecordInsert(w http.ResponseWriter, r *http.Request) {
	json := api.PostRecordInsert(w, r, true)
	jsonFT := appdata.FinanceTracker{}
	if json.IsValidResponse {
		jsonFT = json.AnyStruct.(appdata.FinanceTracker)
		jsonFT.DateDetails.MonthStr = appdata.MonthIto3A(jsonFT.DateDetails.Month)
		api.GetCategoryIcon(w, r, true, jsonFT.Category, &jsonFT.CategoryDetails)
	}
	htmlComponents.PostRecordSingle(json.HttpStatus, jsonFT).Render(r.Context(), w)
}
func GetLenderBorrowerStats(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("in front part")
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	lbListActive, lbListInactive := sqlite.GetLenderBorrowerStats(r.Context(), appdata.DB, userContext.GofiID, false)
	lbIDstr := chi.URLParam(r, "lbID")
	var lbIDint int
	if len(lbListActive) == 0 {
		lbIDint = 0
	} else if lbIDstr == "" || lbIDstr == "0" {
		lbIDint = lbListActive[0].ID
	} else {
		lbIDint, _ = strconv.Atoi(lbIDstr)
	}
	ftList1, ftList2, lbName := sqlite.GetLenderBorrowerDetailedStats(r.Context(), appdata.DB, userContext.GofiID, lbIDint)
	htmlComponents.GetLenderBorrowerStats(lbListActive, lbListInactive, ftList1, ftList2, lbName, lbIDint).Render(r.Context(), w)
}
func GetLendBorrowRecord(w http.ResponseWriter, r *http.Request) {
	jsonFT := api.GetRecords(w, r, true)
	jsonFTlist := jsonFT.AnyStruct.([]appdata.FinanceTracker)
	jsonUP := api.GetParam(w, r, true, "lendborrow", "", false)
	jsonUserParam := jsonUP.AnyStruct.(appdata.UserParams)
	currentTime := time.Now()
	currentDate := currentTime.Format(time.DateOnly) // YYYY-MM-DD
	lbListActive, _ := sqlite.GetLenderBorrowerStats(r.Context(), appdata.DB, jsonUserParam.GofiID, true)
	htmlComponents.GetLendBorrowRecord(jsonFTlist, jsonUserParam, currentDate, lbListActive).Render(r.Context(), w)
}
func PostLendOrBorrowRecord(w http.ResponseWriter, r *http.Request) {
	json := api.PostLendOrBorrowRecords(w, r, true)
	jsonFT := appdata.FinanceTracker{}
	if json.IsValidResponse {
		jsonFT = json.AnyStruct.(appdata.FinanceTracker)
		jsonFT.DateDetails.MonthStr = appdata.MonthIto3A(jsonFT.DateDetails.Month)
		api.GetCategoryIcon(w, r, true, jsonFT.Category, &jsonFT.CategoryDetails)
	}
	htmlComponents.PostRecordSingle(json.HttpStatus, jsonFT).Render(r.Context(), w)
}
func PostLenderBorrowerStateChange(w http.ResponseWriter, r *http.Request) {
	// json := api.PostLenderBorrowerStateChange(w, r, true)
	// htmlComponents.PostLenderBorrowerStateChange(json.HttpStatus).Render(r.Context(), w)

	api.PostLenderBorrowerStateChange(w, r, true)
	//reload the get page
	GetLenderBorrowerStats(w, r)
}
func PostUnlinkLendOrBorrowRecords(w http.ResponseWriter, r *http.Request) {
	api.PostUnlinkLendOrBorrowRecords(w, r, true)
	//reload the get page
	GetLenderBorrowerStats(w, r)
}

func GetRecordTransfer(w http.ResponseWriter, r *http.Request) {
	jsonFT := api.GetRecords(w, r, true)
	jsonFTlist := jsonFT.AnyStruct.([]appdata.FinanceTracker)
	jsonUP := api.GetParam(w, r, true, "", "", false)
	jsonUserParam := jsonUP.AnyStruct.(appdata.UserParams)
	currentTime := time.Now()
	currentDate := currentTime.Format(time.DateOnly) // YYYY-MM-DD
	htmlComponents.GetRecordTransfer(jsonFTlist, jsonUserParam, currentDate).Render(r.Context(), w)
}
func PostRecordTransfer(w http.ResponseWriter, r *http.Request) {
	json := api.PostRecordTransfer(w, r, true)
	jsonFTlist := []appdata.FinanceTracker{}
	if json.IsValidResponse {
		jsonFTlist = json.AnyStruct.([]appdata.FinanceTracker)
		//fmt.Printf("jsonFTlist: %#v \n", jsonFTlist)
	}
	htmlComponents.PostRecordDouble(json.HttpStatus, jsonFTlist).Render(r.Context(), w)
}

func GetRecordRecurrent(w http.ResponseWriter, r *http.Request) {
	jsonFT := api.GetRecords(w, r, true)
	jsonFTlist := []appdata.FinanceTracker{}
	if jsonFT.IsValidResponse {
		jsonFTlist = jsonFT.AnyStruct.([]appdata.FinanceTracker)
	}
	jsonUP := api.GetParam(w, r, true, "type", "periodic", false)
	jsonUserParam := jsonUP.AnyStruct.(appdata.UserParams)
	jsonRR := api.RecordRecurrentRead(w, r, true)
	jsonRRlist := []appdata.RecurrentRecord{}
	if jsonRR.IsValidResponse {
		jsonRRlist = jsonRR.AnyStruct.([]appdata.RecurrentRecord)
		for i, item := range jsonRRlist {
			// generate new id for each row
			jsonRRlist[i].IDstr = strconv.Itoa(item.ID)
			jsonRRlist[i].IDsave = "s" + jsonRRlist[i].IDstr
			jsonRRlist[i].IDedit = "e" + jsonRRlist[i].IDstr
			appdata.ParseDateSVGfront(jsonRRlist[i].Date, &jsonRRlist[i].DateDetails)
			// get category icon for each row
			api.GetCategoryIcon(w, r, true, jsonRRlist[i].Category, &jsonRRlist[i].CategoryDetails)
		}
	}
	currentTime := time.Now()
	currentDate := currentTime.Format(time.DateOnly) // YYYY-MM-DD
	htmlComponents.GetRecordRecurrent(jsonRRlist, jsonFTlist, jsonUserParam, currentDate).Render(r.Context(), w)
}
func PostRecordRecurrentCreate(w http.ResponseWriter, r *http.Request) {
	jsonRR := appdata.RecurrentRecord{}
	json := api.RecordRecurrentCreate(w, r, true)
	if json.IsValidResponse {
		jsonRR = json.AnyStruct.(appdata.RecurrentRecord)
		appdata.ParseDateSVGfront(jsonRR.Date, &jsonRR.DateDetails)
		api.GetCategoryIcon(w, r, true, jsonRR.Category, &jsonRR.CategoryDetails)
	}
	htmlComponents.PostRecordRecurrent(json.HttpStatus, jsonRR).Render(r.Context(), w)
}
func PostRecordRecurrentSave(w http.ResponseWriter, r *http.Request) {
	jsonFT := appdata.FinanceTracker{}
	json := api.RecordRecurrentSave(w, r, true)
	if json.IsValidResponse {
		jsonFT = json.AnyStruct.(appdata.FinanceTracker)
		appdata.ParseDateSVGfront(jsonFT.Date, &jsonFT.DateDetails)
		api.GetCategoryIcon(w, r, true, jsonFT.Category, &jsonFT.CategoryDetails)
	}
	// fmt.Printf("jsonFT: %#v\n", jsonFT)
	htmlComponents.PostRecordRecurrentSave(json.HttpStatus, jsonFT).Render(r.Context(), w)
}
func PostRecordRecurrentUpdate(w http.ResponseWriter, r *http.Request) {
	json := api.RecordRecurrentUpdate(w, r, true)
	jsonRR := appdata.RecurrentRecord{}
	if json.IsValidResponse {
		jsonRR = json.AnyStruct.(appdata.RecurrentRecord)
		appdata.ParseDateSVGfront(jsonRR.Date, &jsonRR.DateDetails)
		api.GetCategoryIcon(w, r, true, jsonRR.Category, &jsonRR.CategoryDetails)
	}
	htmlComponents.PostRecordRecurrent(json.HttpStatus, jsonRR).Render(r.Context(), w)
}
func PostRecordRecurrentDelete(w http.ResponseWriter, r *http.Request) {
	rr := &appdata.RecurrentRecord{}
	api.BindRecordRecurrent(r, rr)
	json := api.RecordRecurrentDelete(w, r, true, rr.IDstr)
	htmlComponents.DeleteRecordRecurrent(json.HttpStatus).Render(r.Context(), w)
}
func PostRecordEdit(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "idft")
	json := api.PostRecordEdit(w, r, true, idStr)
	htmlComponents.PostRecordEdit(json.HttpStatus).Render(r.Context(), w)
}

func GetRecordEdit(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	var idInt int
	if idStr == "" || idStr == "0" {
		idInt = 0
	} else {
		idInt, _ = strconv.Atoi(idStr)
	}
	filterR := appdata.FilterRows{ID: idInt}
	jsonFT := api.GetRecordsViaPost(w, r, true, &filterR)
	jsonFTlist := []appdata.FinanceTracker{}
	if jsonFT.IsValidResponse {
		jsonFTlist = jsonFT.AnyStruct.([]appdata.FinanceTracker)
		for i, item := range jsonFTlist {
			jsonFTlist[i].IDstr = "v" + strconv.Itoa(item.ID)
		}
	}
	lenJsonFTlist := len(jsonFTlist)
	lb := appdata.LendBorrow{ModeStr: strconv.Itoa(jsonFTlist[0].Mode), ModeInt: jsonFTlist[0].Mode, Who: "-", FT: jsonFTlist[0]}
	if lb.ModeInt > 0 && lb.ModeInt < 5 {
		isErr := sqlite.FindLenderBorrowerFromFTid(r.Context(), appdata.DB, &lb)
		if isErr {
			fmt.Println("GetRecordEdit err in FindLenderBorrowerFromFTid")
		}
	}
	if string(lb.FT.FormPriceStr2Decimals[0]) == "-" {
		lb.FT.FormPriceStr2Decimals = lb.FT.FormPriceStr2Decimals[1:]
		lb.FT.PriceDirection = "expense"
	} else {
		lb.FT.PriceDirection = "gain"
	}
	// fmt.Printf("lb GetRecordEdit: %#v\n", lb)
	jsonUP := api.GetParam(w, r, true, "allinuse", "", false)
	jsonUserParam := jsonUP.AnyStruct.(appdata.UserParams)
	lbListActive, _ := sqlite.GetLenderBorrowerStats(r.Context(), appdata.DB, jsonUserParam.GofiID, true)
	htmlComponents.GetRecordEdit(lenJsonFTlist, lb, jsonUserParam, lbListActive).Render(r.Context(), w)
}

func GetRecordAlter(w http.ResponseWriter, r *http.Request) {
	alterMode := chi.URLParam(r, "alterMode")
	if alterMode != "validate" && alterMode != "cancel" {
		alterMode = "edit"
	}
	filterR := appdata.FilterRows{WhereAccount: "", WhereCategory: "", WhereYearStr: "", WhereMonthStr: "",
		WhereCheckedStr: "2",
		OrderBy:         "date",
		OrderSort:       "ASC",
		LimitStr:        "8",
	}
	jsonFT := api.GetRecordsViaPost(w, r, true, &filterR)
	jsonFTlist := []appdata.FinanceTracker{}
	if jsonFT.IsValidResponse {
		jsonFTlist = jsonFT.AnyStruct.([]appdata.FinanceTracker)
		for i, item := range jsonFTlist {
			jsonFTlist[i].IDstr = "v" + strconv.Itoa(item.ID)
		}
	}
	jsonUP := api.GetParam(w, r, true, "", "", true)
	jsonUserParam := jsonUP.AnyStruct.(appdata.UserParams)
	currentTime := time.Now()
	currentDate := currentTime.Format(time.DateOnly) // YYYY-MM-DD
	htmlComponents.GetRecordAlter(jsonFTlist, jsonUserParam, currentDate, r.Header.Get("totalRowsWithoutLimit"), alterMode).Render(r.Context(), w)
}
func PostRecordValidate(w http.ResponseWriter, r *http.Request) {
	json := api.RecordValidate(w, r, true)
	htmlComponents.PostRecordValidate(json.HttpStatus).Render(r.Context(), w)
}
func PostRecordCancel(w http.ResponseWriter, r *http.Request) {
	json := api.RecordCancel(w, r, true)
	htmlComponents.PostRecordCancel(json.HttpStatus).Render(r.Context(), w)
}
func PostFullRecordRefresh(w http.ResponseWriter, r *http.Request) {
	filterR := appdata.FilterRows{}
	jsonFT := api.GetRecordsViaPost(w, r, true, &filterR)
	jsonFTlist := []appdata.FinanceTracker{}
	if jsonFT.IsValidResponse {
		jsonFTlist = jsonFT.AnyStruct.([]appdata.FinanceTracker)
		for i, item := range jsonFTlist {
			jsonFTlist[i].IDstr = "v" + strconv.Itoa(item.ID)
		}
	}
	botScriptAutoReloadOnAccountChange := r.Header.Get("botScriptAutoReloadOnAccountChange")
	// fmt.Printf("botScriptAutoReloadOnAccountChange: %v\n", botScriptAutoReloadOnAccountChange)
	htmlComponents.PostFullRecordRefresh(jsonFTlist, r.Header.Get("totalRowsWithoutLimit"), botScriptAutoReloadOnAccountChange).Render(r.Context(), w)
}

func GetStats(w http.ResponseWriter, r *http.Request) {
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)

	checkedValidDataStr := chi.URLParam(r, "checkedValidData")
	var CheckedValidDataBool bool
	var checkedValidData int
	switch checkedValidDataStr {
	case "", "false":
		CheckedValidDataBool = false
		checkedValidData = 0
	case "true":
		CheckedValidDataBool = true
		checkedValidData = 1
	default:
		CheckedValidDataBool = false
		checkedValidData = 0
	}
	yearStr := chi.URLParam(r, "year")
	var YearInt int
	const yearOnly = "2006" // YYYY
	if yearStr == "" || yearStr == "0" {
		currentTime := time.Now()
		YearInt, _ = strconv.Atoi(currentTime.Format(yearOnly)) // YYYY
	} else {
		YearInt, _ = strconv.Atoi(yearStr) // YYYY
	}
	checkedYearStatsStr := chi.URLParam(r, "checkedYearStats")
	var CheckedYearStatsBool bool
	var checkedYearStats int
	switch checkedYearStatsStr {
	case "", "false":
		CheckedYearStatsBool = false
		checkedYearStats = 0
	case "true":
		CheckedYearStatsBool = true
		checkedYearStats = 1
	default:
		CheckedYearStatsBool = false
		checkedYearStats = 0
	}
	checkedGainsStatsStr := chi.URLParam(r, "checkedGainsStats")
	var CheckedGainsStatsBool bool
	var checkedGainsStats int
	switch checkedGainsStatsStr {
	case "", "false":
		CheckedGainsStatsBool = false
		checkedGainsStats = 0
	case "true":
		CheckedGainsStatsBool = true
		checkedGainsStats = 1
	default:
		CheckedGainsStatsBool = false
		checkedGainsStats = 0
	}

	ApexLineChartStats := sqlite.GetStatsForLineChartInFinanceTracker(
		r.Context(), appdata.DB, userContext.GofiID, checkedValidData, YearInt)

	var AccountList, CategoryList [][]string
	var TotalAccount, TotalCategory []string
	var ApexChartStats appdata.ApexChartStats
	AccountList, CategoryList, TotalAccount, TotalCategory, ApexChartStats = sqlite.GetStatsInFinanceTracker(
		r.Context(), appdata.DB, userContext.GofiID, checkedValidData, YearInt, checkedYearStats, checkedGainsStats)
	// fmt.Printf("CategoryList: %#v\n", CategoryList)

	var Price float64
	var CategoryLabelList, IconCodePointList, ColorHEXList, ColorNameList []string
	var CategoryValueList []float64
	for _, element := range CategoryList {
		Price, _ = strconv.ParseFloat(element[1], 64)
		if Price < 0 {
			// Category = element[0]
			Price = Price * -1
			//Quantity = element[2]
			CategoryLabelList = append(CategoryLabelList, element[0])
			CategoryValueList = append(CategoryValueList, Price)
			IconCodePointList = append(IconCodePointList, element[3])
			ColorHEXList = append(ColorHEXList, element[4])
			ColorNameList = append(ColorNameList, element[5])
		}
	}

	ApexLineChartStatsJson, err := json.Marshal(ApexLineChartStats)
	if err != nil {
		fmt.Println(err)
	}
	ApexChartStatsJson, err := json.Marshal(ApexChartStats)
	if err != nil {
		fmt.Println(err)
	}

	htmlComponents.GetStats(YearInt,
		TotalAccount, TotalCategory,
		AccountList, CategoryList,
		CheckedValidDataBool, CheckedYearStatsBool, CheckedGainsStatsBool,
		CategoryLabelList, CategoryValueList, IconCodePointList, ColorHEXList,
		string(ApexLineChartStatsJson), string(ApexChartStatsJson),
	).Render(r.Context(), w)
}

func GetBudget(w http.ResponseWriter, r *http.Request) {
	jsonUP := api.GetParam(w, r, true, "budget", "", false)
	jsonUserParam := jsonUP.AnyStruct.(appdata.UserParams)
	sqlite.GetBudgetStats(r.Context(), appdata.DB, jsonUserParam.Categories)
	// fmt.Printf("jsonUserParam.Categories: %#v\n", jsonUserParam.Categories)
	htmlComponents.GetBudget(jsonUserParam.Categories).Render(r.Context(), w)
}

// fmt.Printf("json1: %#v\n", json)
// fmt.Printf("json2: %#v\n", json.AnyStruct)

func GetCSVexport(w http.ResponseWriter, r *http.Request) {
	htmlComponents.GetCSVexport().Render(r.Context(), w)
}
func PostCSVexportReset(w http.ResponseWriter, r *http.Request) {
	json := api.PostCSVexportReset(w, r, true)
	htmlComponents.PostCSVexportReset(json.HttpStatus).Render(r.Context(), w)
}

func GetCSVimport(w http.ResponseWriter, r *http.Request) {
	htmlComponents.GetCSVimport().Render(r.Context(), w)
}
func PostCSVimport(w http.ResponseWriter, r *http.Request) {
	json := api.PostCSVimport(w, r, true)
	stringFile := json.AnyStruct.(string)
	htmlComponents.PostCSVimport(json.HttpStatus, json.Info, stringFile).Render(r.Context(), w)
}
