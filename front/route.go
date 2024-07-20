package front

import (
	"encoding/json"
	"fmt"
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
func GetParam(w http.ResponseWriter, r *http.Request) {
	json := api.GetParam(w, r, true)
	jsonUserParam := json.AnyStruct.(appdata.UserParams)
	htmlComponents.GetParam(jsonUserParam).Render(r.Context(), w)
}
func PostParamAccount(w http.ResponseWriter, r *http.Request) {
	json := api.PostParamAccount(w, r, true)
	jsonParam := &appdata.Param{}
	if json.IsValidResponse {
		jsonParam = json.AnyStruct.(*appdata.Param)
	}
	htmlComponents.PostParamAccount(json.HttpStatus, jsonParam.ParamJSONstringData).Render(r.Context(), w)
}
func PostParamCategory(w http.ResponseWriter, r *http.Request) {
	json := api.PostParamCategory(w, r, true)
	jsonParam := &appdata.Param{}
	if json.IsValidResponse {
		jsonParam = json.AnyStruct.(*appdata.Param)
	}
	htmlComponents.PostParamCategory(json.HttpStatus, jsonParam.ParamJSONstringData).Render(r.Context(), w)
}
func PostParamCategoryRendering(w http.ResponseWriter, r *http.Request) {
	json := api.PostParamCategoryRendering(w, r, true)
	jsonParam := &appdata.Param{}
	if json.IsValidResponse {
		jsonParam = json.AnyStruct.(*appdata.Param)
	}
	htmlComponents.PostParamCategoryRendering(json.HttpStatus, jsonParam.ParamJSONstringData).Render(r.Context(), w)
}

func GetParamCategory(w http.ResponseWriter, r *http.Request) {
	jsonUP := api.GetParam(w, r, true)
	jsonUserParam := jsonUP.AnyStruct.(appdata.UserParams)
	for i, item := range jsonUserParam.CategoryList {
		jsonUserParam.CategoryList[i] = append(item, "input"+strconv.Itoa(i))
		jsonUserParam.CategoryList[i] = append(jsonUserParam.CategoryList[i], "icon"+strconv.Itoa(i))
	}
	// fmt.Printf("jsonUserParam.CategoryList: %#v \n", jsonUserParam.CategoryList)
	// categoryList, iconCodePointList, colorHEXList := sqlite.GetCategoryList(r.Context(), appdata.DB)
	userCategories := appdata.NewUserCategories()
	userCategories.GofiID = jsonUserParam.GofiID
	sqlite.GetFullCategoryList(r.Context(), appdata.DB, userCategories)
	userCategoriesJson, err := json.Marshal(userCategories)
	if err != nil {
		fmt.Println(err)
	}
	htmlComponents.GetParamCategory(jsonUserParam, userCategories, string(userCategoriesJson)).Render(r.Context(), w)
}

// RECORD
func GetRecordInsert(w http.ResponseWriter, r *http.Request) {
	jsonFT := api.GetRecords(w, r, true)
	jsonFTlist := jsonFT.AnyStruct.([]appdata.FinanceTracker)
	jsonUP := api.GetParam(w, r, true)
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
	jsonUP := api.GetParam(w, r, true)
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
	jsonUP := api.GetParam(w, r, true)
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
	jsonUP := api.GetParam(w, r, true)
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

	var Price float64
	var CategoryLabelList, IconCodePointList, ColorHEXList []string
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
