package front

import (
	"net/http"
)

/*
type Cookie struct {
	Name  string
	Value string

	Path       string    // optional
	Domain     string    // optional
	Expires    time.Time // optional
	RawExpires string    // for reading cookies only

	// MaxAge=0 means no 'Max-Age' attribute specified.
	// MaxAge<0 means delete cookie now, equivalently 'Max-Age: 0'
	// MaxAge>0 means Max-Age attribute present and given in seconds
	MaxAge   int
	Secure   bool
	HttpOnly bool
	SameSite SameSite
	Raw      string
	Unparsed []string // Raw text of unparsed attribute-value pairs
}

*/
// FUNC set cookie
func FrontSetCookie(w http.ResponseWriter, cookieValue string) {
	// c.SetSameSite(http.SameSiteLaxMode)
	// c.SetCookie("gofiID", "", -1, "/", "", true, true)
	// c.SetCookie("gofiID", cookie, 2592000, "/", "", true, true) // 30d duration

	cookieStruct := &http.Cookie{}
	cookieStruct.Name = "gofiID"
	cookieStruct.Value = "empty"
	cookieStruct.Path = "/"
	cookieStruct.Domain = ""
	cookieStruct.MaxAge = -1
	cookieStruct.Secure = true
	cookieStruct.HttpOnly = true
	cookieStruct.SameSite = http.SameSiteLaxMode
	http.SetCookie(w, cookieStruct)

	cookieStruct.Value = cookieValue
	cookieStruct.MaxAge = 2592000
	http.SetCookie(w, cookieStruct)
}
