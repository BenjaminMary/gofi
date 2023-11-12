package main

import (
    "fmt"

    "net/http"
    "crypto/rand"
    "math/big"

    "example.com/sqlite"

    "github.com/gin-gonic/gin"
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
func CheckCookie(c *gin.Context) (int) {
    // try to read if a cookie exists, return to login otherwise
    sessionID, err := c.Cookie("gofiID")
    if err != nil {
        if c.Request.Method == "GET" {
            c.Redirect(http.StatusSeeOther, "/login")
            c.Abort()
            return 0
        } else if c.Request.Method == "POST" {
            c.Header("HX-Retarget", "#forbidden")
            c.Header("HX-Reswap", "innerHTML")
            c.String(http.StatusForbidden, `
                <div id="forbidden">
                    <p>
                        ERREUR: Aucun identifiant trouvé (Cookie gofiID).<br> 
                        Requête annulée, il est nécessaire de réenregistrer un identifiant pour reprendre.<br> 
                        C'est par ici: <a href="/cookie-setup">Setup Gofi Cookie</a>
                    </p>
                </div>
            `)
            c.Abort()
            return 0
        }
    }
    gofiID, errorStrReason, err := sqlite.GetGofiID(sessionID)
    if (gofiID > 0) { return gofiID } else {
        fmt.Printf("errorStrReason: %v\n", errorStrReason)
        fmt.Printf("err: %v\n", err)
        c.Redirect(http.StatusSeeOther, "/login")
        c.Abort()
        return 0
    }
}
