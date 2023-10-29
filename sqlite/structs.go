package sqlite

type Param struct {
	ID int
	GofiID string
	ParamName string
	ParamJSONstringData string
	ParamInfo string
}

type FinanceTracker struct {
	ID int
	GofiID string
	Date string `form:"date" binding:"required"`
	Year int
	Month int
	Day int
	AccountList []string
	Account string `form:"compte" binding:"required"`
	Product string `form:"designation" binding:"required"`
	FormPriceStr2Decimals string `form:"prix" binding:"required"`
	PriceIntx100 int
	CategoryList []string
	Category string `form:"categorie" binding:"required"`
	CommentInt int
	CommentString string
	Checked bool
	DateChecked string
	SentToSheets bool
}