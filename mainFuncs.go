package main

import (
    "fmt"
	"database/sql"
    "context"

    "net/http"
    "crypto/rand"
    "math/big"
    "os"
    "strconv"
    "slices"

    "example.com/sqlite"

    "github.com/gin-gonic/gin"
)

var (
    cookieLengthOs = os.Getenv("COOKIE_LENGTH")
    CookieLength, _ = strconv.Atoi(cookieLengthOs)
)

// FUNC generateRandomString returns a securely generated random string
func GenerateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_.~"
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil { return "", err }
		ret[i] = letters[num.Int64()]
	}
	return string(ret), nil
}

// FUNC set cookie
func SetCookie(c *gin.Context, cookie string) {
    c.SetSameSite(http.SameSiteLaxMode)
    c.SetCookie("gofiID", "", -1, "/", "", true, true)
    c.SetCookie("gofiID", cookie, 2592000, "/", "", true, true) // 30d duration
    return
}

// FUNC check cookie sessionID to get the gofiID
func CheckCookie(ctx context.Context, c *gin.Context, db *sql.DB) (int, string) {
    // try to read if a cookie exists, return to login otherwise
    sessionID, err := c.Cookie("gofiID")
    uriList := []string{"/", "/logout", "/login"}
    if ( err != nil || len(sessionID) != CookieLength ) {
        if slices.Contains(uriList, c.Request.RequestURI) {return 0, ""}
        if c.Request.Method == "GET" {
            c.Redirect(http.StatusSeeOther, "/login")
            c.Abort()
            return 0, ""
        } else if c.Request.Method == "POST" {
            c.Header("HX-Retarget", "#forbidden")
            c.Header("HX-Reswap", "innerHTML")
            c.String(http.StatusForbidden, `
                <div id="forbidden">
                    <p>
                        ERREUR: Aucun identifiant trouvé (Cookie gofiID).<br> 
                        Requête annulée, il est nécessaire de se reconnecter pour reprendre.<br> 
                        C'est par ici: <a href="/login">Login</a>
                    </p>
                </div>
            `)
            c.Abort()
            return 0, ""
        }
    }
    gofiID, email, errorStrReason, err := sqlite.GetGofiID(ctx, db, sessionID)
    if (gofiID > 0) { 
        if (errorStrReason == "idleTimeout, change cookie") {
            newSessionID, err := GenerateRandomString(CookieLength)
            if (err != nil) {
                fmt.Printf("err GenerateRandomString: %v\n", err)
                c.Redirect(http.StatusSeeOther, "/login")
                c.Abort()
                return 0, ""
            }
            errorStrReason, err := sqlite.UpdateSessionID(ctx, db, gofiID, newSessionID)
            if (err != nil) {
                fmt.Printf("errorStrReason: %v\n", errorStrReason)
                fmt.Printf("err: %v\n", err)
                c.Redirect(http.StatusSeeOther, "/login")
                c.Abort()
                return 0, ""
            }
            SetCookie(c, newSessionID)
            fmt.Println("auto cookie update")
        }
        return gofiID, email
    } else {
        fmt.Printf("errorStrReason: %v\n", errorStrReason)
        fmt.Printf("err: %v\n", err)
        c.Redirect(http.StatusSeeOther, "/login")
        c.Abort()
        return 0, ""
    }
}


// FUNC check admin 
func CheckAdmin(c *gin.Context, email string) {
    // check if user is an admin to access the required page, redirect otherwise
    if (email != os.Getenv("ADMIN_EMAIL")) {
        c.Redirect(http.StatusSeeOther, "/")
        c.Abort()
        return
    }
}
