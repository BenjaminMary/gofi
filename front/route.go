package front

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

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
func GetParamSetup(w http.ResponseWriter, r *http.Request) {
	json := api.GetParamSetup(w, r, true)
	jsonUserParam := json.AnyStruct.(appdata.UserParams)
	htmlComponents.GetParamSetup(jsonUserParam).Render(r.Context(), w)
}
func PostParamSetupAccount(w http.ResponseWriter, r *http.Request) {
	json := api.PostParamSetupAccount(w, r, true)
	jsonParam := &appdata.Param{}
	if json.IsValidResponse {
		jsonParam = json.AnyStruct.(*appdata.Param)
	}
	htmlComponents.PostParamSetupAccount(json.HttpStatus, jsonParam.ParamJSONstringData).Render(r.Context(), w)
}
func PostParamSetupCategory(w http.ResponseWriter, r *http.Request) {
	json := api.PostParamSetupCategory(w, r, true)
	jsonParam := &appdata.Param{}
	if json.IsValidResponse {
		jsonParam = json.AnyStruct.(*appdata.Param)
	}
	htmlComponents.PostParamSetupCategory(json.HttpStatus, jsonParam.ParamJSONstringData).Render(r.Context(), w)
}
func PostParamSetupCategoryRendering(w http.ResponseWriter, r *http.Request) {
	json := api.PostParamSetupCategoryRendering(w, r, true)
	jsonParam := &appdata.Param{}
	if json.IsValidResponse {
		jsonParam = json.AnyStruct.(*appdata.Param)
	}
	htmlComponents.PostParamSetupCategoryRendering(json.HttpStatus, jsonParam.ParamJSONstringData).Render(r.Context(), w)
}

// RECORD
func GetRecordInsert(w http.ResponseWriter, r *http.Request) {
	jsonFT := api.GetRecords(w, r, true)
	jsonFTlist := jsonFT.AnyStruct.([]appdata.FinanceTracker)
	jsonUP := api.GetParamSetup(w, r, true)
	jsonUserParam := jsonUP.AnyStruct.(appdata.UserParams)
	for i, item := range jsonUserParam.CategoryList {
		jsonUserParam.CategoryList[i] = append(item, "input"+strconv.Itoa(i))
		jsonUserParam.CategoryList[i] = append(jsonUserParam.CategoryList[i], "icon"+strconv.Itoa(i))
	}
	currentTime := time.Now()
	currentDate := currentTime.Format(time.DateOnly) // YYYY-MM-DD
	htmlComponents.GetRecordInsert(jsonFTlist, jsonUserParam, currentDate).Render(r.Context(), w)
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
func GetRecordTransfer(w http.ResponseWriter, r *http.Request) {
	jsonFT := api.GetRecords(w, r, true)
	jsonFTlist := jsonFT.AnyStruct.([]appdata.FinanceTracker)
	jsonUP := api.GetParamSetup(w, r, true)
	jsonUserParam := jsonUP.AnyStruct.(appdata.UserParams)
	for i, item := range jsonUserParam.CategoryList {
		jsonUserParam.CategoryList[i] = append(item, "input"+strconv.Itoa(i))
		jsonUserParam.CategoryList[i] = append(jsonUserParam.CategoryList[i], "icon"+strconv.Itoa(i))
	}
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
	jsonUP := api.GetParamSetup(w, r, true)
	jsonUserParam := jsonUP.AnyStruct.(appdata.UserParams)
	for i, item := range jsonUserParam.CategoryList {
		jsonUserParam.CategoryList[i] = append(item, "input"+strconv.Itoa(i))
		jsonUserParam.CategoryList[i] = append(jsonUserParam.CategoryList[i], "icon"+strconv.Itoa(i))
	}
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

func GetRecordValidateOrCancel(w http.ResponseWriter, r *http.Request) {
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
	jsonUP := api.GetParamSetup(w, r, true)
	jsonUserParam := jsonUP.AnyStruct.(appdata.UserParams)
	forceSelectCategory := [][]string{{"", "e90a", "#808080", "input0", "icon0"}} // ? icon for default value
	currentCategoryList := jsonUserParam.CategoryList
	jsonUserParam.CategoryList = forceSelectCategory
	for i, item := range currentCategoryList {
		j := i + 1 //gap with the added element at the first position 0
		jsonUserParam.CategoryList = append(jsonUserParam.CategoryList, item)
		jsonUserParam.CategoryList[j] = append(item, "input"+strconv.Itoa(j), "icon"+strconv.Itoa(j))
	}
	// fmt.Printf("jsonUserParam.CategoryList: %#v\n", jsonUserParam.CategoryList)
	currentTime := time.Now()
	currentDate := currentTime.Format(time.DateOnly) // YYYY-MM-DD
	htmlComponents.GetRecordValidateOrCancel(jsonFTlist, jsonUserParam, currentDate, r.Header.Get("totalRowsWithoutLimit")).Render(r.Context(), w)
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
	htmlComponents.PostFullRecordRefresh(jsonFTlist, r.Header.Get("totalRowsWithoutLimit")).Render(r.Context(), w)
}

func GetStats(w http.ResponseWriter, r *http.Request) {
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)

	checkedStr := chi.URLParam(r, "checked")
	var CheckedBool bool
	var checkedDataOnly int
	switch checkedStr {
	case "", "false":
		CheckedBool = false
		checkedDataOnly = 0
	case "true":
		CheckedBool = true
		checkedDataOnly = 1
	default:
		CheckedBool = false
		checkedDataOnly = 0
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

	var AccountList, CategoryList [][]string
	var TotalAccount, TotalCategory []string
	AccountList, CategoryList, TotalAccount, TotalCategory = sqlite.GetStatsInFinanceTracker(r.Context(), appdata.DB, userContext.GofiID, checkedDataOnly, YearInt)

	var m appdata.PieChartD3js
	var CategoryListJsonBinary []appdata.PieChartD3js
	var CategoryLabelList, IconCodePointList, ColorHEXList []string
	var CategoryValueList []float64
	for _, element := range CategoryList {
		m.Price, _ = strconv.ParseFloat(element[1], 64)
		if m.Price < 0 {
			m.Category = element[0]
			m.Price = m.Price * -1
			//m.Quantity = element[2]
			CategoryListJsonBinary = append(CategoryListJsonBinary, m)
			CategoryLabelList = append(CategoryLabelList, element[0])
			CategoryValueList = append(CategoryValueList, m.Price)
			IconCodePointList = append(IconCodePointList, element[3])
			ColorHEXList = append(ColorHEXList, element[4])
		}
	}
	ResponseJsonBinary, _ := json.Marshal(CategoryListJsonBinary)
	htmlComponents.GetStats(YearInt,
		TotalAccount, TotalCategory,
		AccountList, CategoryList,
		string(ResponseJsonBinary), // array of dict [{},{}] for d3.js
		CheckedBool,
		CategoryLabelList, CategoryValueList, IconCodePointList, ColorHEXList,
	).Render(r.Context(), w)
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
