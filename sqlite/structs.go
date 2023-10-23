package sqlite

type FinanceTracker struct {
	ID int
	GofiID string
	Date string `form:"date" binding:"required"`
	Year int
	Month int
	Day int
	PaymentMethod string
	Product string `form:"designation" binding:"required"`
	FormPriceStr2Decimals string `form:"prix" binding:"required"`
	PriceIntx100 int
	Category string `form:"categorie" binding:"required"`
	CommentFloat float32
	CommentString string
	Checked bool
	DateChecked string
	SentToSheets bool
}