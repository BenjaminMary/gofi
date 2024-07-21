package appdata

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/render"
)

type Param struct {
	ID                  int
	GofiID              int
	ParamName           string
	ParamJSONstringData string
	ParamInfo           string
}

func (a *Param) Bind(r *http.Request) error {
	// trigger an error if field = "" or is missing/wrong
	if a.ParamJSONstringData == "" {
		return errors.New("missing required field")
	}
	return nil
}

type CategoryDetails struct {
	CategoryIcon  string
	CategoryColor string
}
type DateDetails struct {
	Year     int
	Month    int
	MonthStr string
	Day      int
}
type FinanceTracker struct {
	ID                    int
	IDstr                 string `json:"-"`
	GofiID                int
	Date                  string      `form:"date" binding:"required"`
	DateDetails           DateDetails `json:"-"`
	Account               string      `form:"compte" binding:"required"`
	Product               string      `form:"designation"`
	PriceDirection        string      `form:"gain-expense" binding:"required"`
	FormPriceStr2Decimals string      `form:"prix" binding:"required"`
	PriceIntx100          int
	Category              string          `form:"categorie" binding:"required"`
	CategoryDetails       CategoryDetails `json:"-"`
	CommentInt            int
	CommentString         string
	Checked               bool
	DateChecked           string
	Exported              bool
}

func (a *FinanceTracker) Bind(r *http.Request) error {
	// trigger an error if field = "" or is missing/wrong
	// fmt.Printf("Date: %v, Account: %v, Category: %v, FormPriceStr2Decimals: %v\n", a.Date, a.Account, a.Category, a.FormPriceStr2Decimals)
	if len(a.Date) != 10 {
		return errors.New("missing required field")
	}
	if len(a.Account) == 0 {
		return errors.New("missing required field")
	}
	if !(a.PriceDirection == "expense" || a.PriceDirection == "gain") {
		return errors.New("missing required field")
	}
	if len(a.FormPriceStr2Decimals) == 0 {
		return errors.New("missing required field")
	}
	if len(a.Category) == 0 {
		return errors.New("missing required field")
	}
	return nil
}

type RecurrentRecord struct {
	ID                    int
	IDstr                 string `form:"idRRmain"`
	IDsave                string `json:"-"`
	IDedit                string `json:"-"`
	GofiID                int
	Date                  string      `form:"date" binding:"required"`
	DateDetails           DateDetails `json:"-"`
	Recurrence            string      `form:"recurrence" binding:"required"`
	Account               string      `form:"compte" binding:"required"`
	Product               string      `form:"designation" binding:"required"`
	PriceDirection        string      `form:"gain-expense" binding:"required"`
	FormPriceStr2Decimals string      `form:"prix" binding:"required"`
	PriceIntx100          int
	Category              string          `form:"categorie" binding:"required"`
	CategoryDetails       CategoryDetails `json:"-"`
}

func (a *RecurrentRecord) Bind(r *http.Request) error {
	if len(a.Date) != 10 {
		return errors.New("missing required field")
	}
	if !(a.Recurrence == "mensuelle" || a.Recurrence == "hebdomadaire" || a.Recurrence == "annuelle") {
		return errors.New("missing required field")
	}
	if len(a.Account) == 0 {
		return errors.New("missing required field")
	}
	if len(a.Category) == 0 {
		return errors.New("missing required field")
	}
	if len(a.Product) == 0 {
		return errors.New("missing required field")
	}
	if !(a.PriceDirection == "expense" || a.PriceDirection == "gain") {
		return errors.New("missing required field")
	}
	if len(a.FormPriceStr2Decimals) == 0 {
		return errors.New("missing required field")
	}
	return nil
}

type RecurrentRecordSave struct {
	ID string `form:"idRR" binding:"required"`
}

func (a *RecurrentRecordSave) Bind(r *http.Request) error {
	if len(a.ID) == 0 {
		return errors.New("missing required field")
	}
	return nil
}

type RecordValidateOrCancel struct {
	Date             string `form:"dateCopy" binding:"required"`
	IDcheckedListStr string `form:"checkedList" binding:"required"`
	IDcheckedListInt []int
}

func (a *RecordValidateOrCancel) Bind(r *http.Request) error {
	fmt.Printf("Date: %v, IDcheckedListStr: %v\n", a.Date, a.IDcheckedListStr)
	if len(a.Date) != 10 {
		return errors.New("missing required field")
	}
	if len(a.IDcheckedListStr) < 1 {
		return errors.New("missing required field")
	}
	return nil
}

type Transfer struct {
	Date                  string `form:"date" binding:"required"`
	AccountFrom           string `form:"compteDepuis" binding:"required"`
	AccountTo             string `form:"compteVers" binding:"required"`
	FormPriceStr2Decimals string `form:"prix" binding:"required"`
}

func (a *Transfer) Bind(r *http.Request) error {
	// fmt.Printf("Date: %v, AccountFrom: %v, AccountTo: %v, FormPriceStr2Decimals: %v\n", a.Date, a.AccountFrom, a.AccountTo, a.FormPriceStr2Decimals)
	if len(a.Date) != 10 {
		return errors.New("missing required field")
	}
	if len(a.AccountFrom) == 0 || a.AccountFrom == "-" {
		return errors.New("missing required field")
	}
	if len(a.AccountTo) == 0 || a.AccountTo == "-" {
		return errors.New("missing required field")
	}
	if len(a.FormPriceStr2Decimals) == 0 {
		return errors.New("missing required field")
	}
	return nil
}

type User struct {
	*UserRequest
	NumberOfRequests           int    `json:"-"`
	IdleDateModifier           string `json:"-"`
	AbsoluteDateModifier       string `json:"-"`
	IdleTimeout                string `json:"-"`
	AbsoluteTimeout            string `json:"-"`
	LastLoginTime              string `json:"lastLoginTime"`
	LastActivityTime           string `json:"-"`
	LastActivityIPaddress      string `json:"-"`
	LastActivityUserAgent      string `json:"-"`
	LastActivityAcceptLanguage string `json:"-"`
	DateCreated                string `json:"dateCreated"`
}

type UserRequest struct {
	GofiID          int    `json:"gofiID"`    // UNIQUE
	Email           string `json:"email"`     // UNIQUE
	SessionID       string `json:"sessionID"` // UNIQUE
	Password        string `json:"password"`
	PwHash          string `json:"-"` // "-" = not returned
	IsAdmin         bool   `json:"-"`
	IsAuthenticated bool   `json:"-"`
}

func (a *UserRequest) Bind(r *http.Request) error {
	// trigger an error if email or password = "" or one is missing/wrong
	if a.Email == "" || a.Password == "" {
		return errors.New("missing required field")
	}
	return nil
}

type UserParams struct {
	GofiID                   int // UNIQUE
	AccountListSingleString  string
	AccountList              []string
	CategoryListSingleString string
	CategoryList             [][]string
	CategoryRendering        string
}

func NewUserCategories() *UserCategories {
	var a UserCategories
	a.FindCategory = make(map[string]int)
	return &a
}

type UserCategories struct {
	GofiID       int // UNIQUE
	FindCategory map[string]int
	Categories   []Category
}
type Category struct {
	ID                           int
	GofiID                       int
	Name                         string
	Type                         string
	Order                        int
	InUse                        int
	InStats                      int
	Description                  string
	BudgetPrice                  int
	BudgetPeriod                 string
	BudgetType                   string
	BudgetCurrentPeriodStartDate string
	BudgetCurrentPeriodEndDate   string
	IconCodePoint                string
	ColorHEX                     string
}
type CategoryPut struct {
	ID                           int
	IDstr                        string `json:"idStrJson"`
	GofiID                       int
	Type                         string
	InStats                      int
	InStatsStr                   string
	Description                  string
	BudgetPrice                  int
	BudgetPriceStr               string
	BudgetPeriod                 string
	BudgetType                   string
	BudgetCurrentPeriodStartDate string
}

func (a *CategoryPut) Bind(r *http.Request) error {
	if a.IDstr == "" {
		fmt.Println("Bind CategoryPut err1")
		return errors.New("missing required field")
	}
	var err error
	a.ID, err = strconv.Atoi(a.IDstr)
	if err != nil || a.ID < 1 {
		fmt.Println("Bind CategoryPut err2")
		return errors.New("missing required field")
	}
	if a.Type == "" || !(a.Type == "all" || a.Type == "periodic" || a.Type == "basic") {
		fmt.Println("Bind CategoryPut err3")
		return errors.New("missing required field")
	}
	if a.InStatsStr == "on" {
		a.InStats = 1
	} else if a.InStatsStr == "" {
		a.InStats = 0
	} else {
		fmt.Println("Bind CategoryPut err5")
		return errors.New("missing required field")
	}
	if a.Description == "" {
		a.Description = "-"
	}
	if a.BudgetPriceStr == "" {
		a.BudgetPriceStr = "0"
		a.BudgetPrice = 0
	} else {
		a.BudgetPrice, err = strconv.Atoi(a.BudgetPriceStr)
		if err != nil || a.BudgetPrice < 0 {
			fmt.Println("Bind CategoryPut err6")
			return errors.New("missing required field")
		}
	}
	if a.BudgetType == "" || !(a.BudgetType == "-" || a.BudgetType == "cumulative" || a.BudgetType == "reset") {
		fmt.Println("Bind CategoryPut err7")
		return errors.New("missing required field")
	}
	if len(a.BudgetCurrentPeriodStartDate) != 10 {
		fmt.Println("Bind CategoryPut err8")
		return errors.New("missing required field")
	}
	return nil
}

type CategoryPatchInUse struct {
	ID       int
	IDstr    string `json:"idStrJson"`
	GofiID   int
	InUse    int
	InUseStr string `json:"inUseStrJson"`
}

func (a *CategoryPatchInUse) Bind(r *http.Request) error {
	if a.IDstr == "" || a.InUseStr == "" {
		fmt.Println("Bind CategoryPatchInUse err0")
		return errors.New("missing required field")
	}
	var err error
	a.ID, err = strconv.Atoi(a.IDstr)
	if err != nil || a.ID < 1 {
		fmt.Println("Bind CategoryPatchInUse err1")
		return errors.New("missing required field")
	}
	a.InUse, err = strconv.Atoi(a.InUseStr)
	if err != nil || a.InUse > 1 || a.InUse < 0 {
		fmt.Println("Bind CategoryPatchInUse err2")
		return errors.New("missing required field")
	}
	return nil
}

type CategoryPatchOrder struct {
	ID1    int
	ID1str string `json:"id1StrJson"`
	ID2    int
	ID2str string `json:"id2StrJson"`
	GofiID int
}

func (a *CategoryPatchOrder) Bind(r *http.Request) error {
	if a.ID1str == "" || a.ID2str == "" || a.ID1str == a.ID2str {
		fmt.Println("Bind CategoryPatchOrder err1")
		return errors.New("missing required field")
	}
	var err error
	a.ID1, err = strconv.Atoi(a.ID1str)
	if err != nil || a.ID1 < 1 {
		fmt.Println("Bind CategoryPatchOrder err2")
		return errors.New("missing required field")
	}
	a.ID2, err = strconv.Atoi(a.ID2str)
	if err != nil || a.ID2 < 1 {
		fmt.Println("Bind CategoryPatchOrder err3")
		return errors.New("missing required field")
	}
	return nil
}

type FilterRows struct {
	GofiID          int    // UNIQUE
	WhereAccount    string `json:"compteHidden"`
	WhereCategory   string `json:"category"`
	WhereYearStr    string `json:"annee"`
	WhereYear       int
	WhereMonthStr   string `json:"mois"`
	WhereMonth      int
	WhereCheckedStr string `json:"checked"`
	WhereChecked    int    // 0 default don't use, 1 = True, 2 = False
	OrderBy         string `json:"orderBy"`
	OrderSort       string `json:"orderSort"`
	LimitStr        string `json:"limitStr"`
	Limit           int
}

func (a *FilterRows) Bind(r *http.Request) error {
	return nil
}

type SaveBackup struct {
	ID          int    `json:"id"`
	Date        string `json:"date"`
	ExtID       string `json:"extID"`
	ExtFileName string `json:"extFileName"`
	Checkpoint  string `json:"checkpoint"`
	Tested      string `json:"tested"`
}

func (a *SaveBackup) Bind(r *http.Request) error {
	// trigger an error if email or password = "" or one is missing/wrong
	if a.ExtID == "" || a.ExtFileName == "" {
		return errors.New("missing required field")
	}
	return nil
}

func NewApexChartStats() *ApexChartStats {
	var a ApexChartStats
	a.FindSerie = make(map[string]int)
	return &a
}

type ApexChartStats struct {
	Labels    []string
	FindSerie map[string]int
	Series    []ApexChartSerie
}
type ApexChartSerie struct {
	Name   string
	Icon   string
	Color  string
	Values []string
}

type HttpStruct struct {
	IsValidResponse bool   `json:"isValidResponse"`
	HttpStatus      int    `json:"httpStatus"`
	Info            string `json:"info"`
	AnyStruct       any    `json:"jsonContent"`
}

func (a *HttpStruct) Bind(r *http.Request) error {
	return nil
}

func (a *HttpStruct) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func RenderAPIorUI(w http.ResponseWriter, r *http.Request, isFrontRequest bool, retargetError bool,
	valid bool, httpStatus int, info string, jsonContent any) *HttpStruct {
	HttpStruct := &HttpStruct{}
	HttpStruct.IsValidResponse = valid
	HttpStruct.HttpStatus = httpStatus
	HttpStruct.Info = info
	HttpStruct.AnyStruct = jsonContent
	// fmt.Printf("valid: %v, retargetError: %v\n", valid, retargetError)
	if isFrontRequest {
		if !valid && retargetError {
			w.Header().Set("HX-Retarget", "#htmxInfo")
			w.Header().Set("HX-Reswap", "innerHTML settle:300ms")
		}
	} else {
		render.Status(r, httpStatus)
		render.Render(w, r, HttpStruct)
	}
	return HttpStruct
}

func RenderFile(w http.ResponseWriter, r *http.Request, isFrontRequest bool, retargetError bool,
	valid bool, httpStatus int, info string, jsonContent any, fileData []byte) *HttpStruct {
	HttpStruct := &HttpStruct{}
	HttpStruct.IsValidResponse = valid
	HttpStruct.HttpStatus = httpStatus
	HttpStruct.Info = info
	HttpStruct.AnyStruct = jsonContent
	// fmt.Printf("valid: %v, retargetError: %v\n", valid, retargetError)
	if isFrontRequest {
		if !valid && retargetError {
			w.Header().Set("HX-Retarget", "#htmxInfo")
			w.Header().Set("HX-Reswap", "innerHTML settle:300ms")
		}
	} else {
		render.Status(r, httpStatus)
		render.Data(w, r, fileData)
	}
	return HttpStruct
}
