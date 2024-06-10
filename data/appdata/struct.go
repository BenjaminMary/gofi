package appdata

import (
	"errors"
	"fmt"
	"net/http"

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

type PieChartD3js struct {
	Category string  `json:"name"`
	Price    float64 `json:"value"`
	//Quantity string `json:"quantity"`
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
