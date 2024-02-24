package sqlite

type Param struct {
	ID int
	GofiID int
	ParamName string
	ParamJSONstringData string
	ParamInfo string
}

type FinanceTracker struct {
	ID int
	GofiID int
	Date string `form:"date" binding:"required"`
	Year int
	Month int
	Day int
	Account string `form:"compte" binding:"required"`
	Product string `form:"designation"`
	FormPriceStr2Decimals string `form:"prix" binding:"required"`
	PriceIntx100 int
	Category string `form:"categorie" binding:"required"`
	CommentInt int
	CommentString string
	Checked bool
	DateChecked string
	Exported bool
}

type RecurrentRecord struct {
	ID int
	GofiID int
	Date string `form:"date" binding:"required"`
	Year int
	Month int
	Day int
	Recurrence string `form:"recurrence" binding:"required"`
	Account string `form:"compte" binding:"required"`
	Product string `form:"designation" binding:"required"`
	FormPriceStr2Decimals string `form:"prix" binding:"required"`
	PriceIntx100 int
	Category string `form:"categorie" binding:"required"`
}

type User struct {
	GofiID int // UNIQUE
	Email string // UNIQUE
	SessionID string // UNIQUE
	PwHash string
	NumberOfRequests int
	IdleDateModifier string
	AbsoluteDateModifier string
	IdleTimeout string
	AbsoluteTimeout string
	LastLoginTime string
	LastActivityTime string
	LastActivityIPaddress string
	LastActivityUserAgent string
	LastActivityAcceptLanguage string
	DateCreated string
}

type UserParams struct {
	GofiID int // UNIQUE
	AccountListSingleString string
	AccountList []string
	CategoryListSingleString string
	CategoryList [][]string
}

type FilterRows struct {
	GofiID int // UNIQUE
	WhereAccount string
	WhereCategory string
	WhereYear int
	WhereMonth int
	WhereChecked int // 0 default don't use, 1 = True, 2 = False
	OrderBy string
	OrderByType string
	Limit int
}

type PieChartD3js struct {
	Category string  `json:"name"`
	Price float64 `json:"value"`
	//Quantity string `json:"quantity"`
}