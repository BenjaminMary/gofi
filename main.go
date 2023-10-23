package main

import (
    "net/http"
    "fmt"
    "time"
    "strings"
    "strconv"
    "html/template"
    // "os"
    // "encoding/json"

    "example.com/sqlite"

    "github.com/gin-gonic/gin"
    // "github.com/gin-gonic/gin/render"
)

// index.html
func index(c *gin.Context) {
    c.HTML(http.StatusOK, "0.index.html", "")
}

// GET cookie
func getCookieSetup(c *gin.Context) {
    // try to read if a cookie exists, return "Aucun" otherwise
    cookie, err := c.Cookie("gofiID")
    if err != nil {
        cookie = "Aucun"
    }    
    c.HTML(http.StatusOK, "1.cookieSetup.html", gin.H{
        "Cookie": cookie,
    })
}
// POST cookie
func postCookieSetup(c *gin.Context) {
    // name (string): The name of the cookie to be set.
    // value (string): The value of the cookie.
    // maxAge (int): The maximum age of the cookie in seconds. If set to 0, the cookie will be deleted immediately. If set to a negative value, the cookie will be a session cookie and will be deleted when the browser is closed.
    // path (string): The URL path for which the cookie is valid. Defaults to "/", meaning the cookie is valid for all URLs.
    // domain (string): The domain for which the cookie is valid. Defaults to the current domain with "".
    // secure (bool): If set to true, the cookie will only be sent over secure (HTTPS) connections.
    // httpOnly (bool): If set to true, the cookie will be inaccessible to JavaScript and can only be sent over HTTP(S) connections.
    
    // set a cookie
    gofiID := c.PostForm("gofiID")
    cookieDurationStr := c.PostForm("cookieDuration")
    var cookieDurationInt64 int64
    var cookieDurationInt int
	cookieDurationInt64, err := strconv.ParseInt(cookieDurationStr, 10, 0)
    if err != nil {
        c.String(http.StatusBadRequest, "cookieDurationStr conversion en int64 KO: %v", err)
        return
    }	
    cookieDurationInt = int(cookieDurationInt64)

    c.SetSameSite(http.SameSiteLaxMode)
    c.SetCookie("gofiID", "", -1, "/", "", false, true)
    c.SetCookie("gofiID", gofiID, cookieDurationInt, "/", "", false, true)
    c.String(200, `<p>Cookie: %s</p><p id="hx-swap-oob1" hx-swap-oob="true">Nouveau Cookie enregistré.</p>`, gofiID)
}

// GET InsertRows.html
func getinsertrows(c *gin.Context) {
    currentTime := time.Now()
    currentDate := currentTime.Format("2006-01-02") // YYYY-MM-DD
    _, err := c.Cookie("gofiID")
    if err != nil {
        c.Redirect(http.StatusSeeOther, "/cookie-setup")
        return
    }    
    c.HTML(http.StatusOK, "3.insertrows.html", gin.H{
        "currentDate": currentDate,
    })
}

// POST InsertRows.html
func postinsertrows(c *gin.Context) {
    cookieGofiID, err := c.Cookie("gofiID")
    if err != nil {
        // cookieGofiID = "Aucun"
        // <tr><td>ERREUR</td><td>ERREUR</td><td>ERREUR</td><td>ERREUR</td></tr>

        //FONCTIONNEL avec oob
        // c.String(http.StatusOK, `
        //     <p>ERREUR</p>
        //     <p id="hx-swap-oob1" hx-swap-oob="true">Aucun identifiant, requête annulée, <a href="/cookie-setup">Setup Gofi Cookie</a></p>
        // `)

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
        
        // response := `<div id="forbidden">ERREUR: Aucun identifiant</div>`
        // c.Render(
        //     http.StatusForbidden, render.Data{
        //         "HX-Retarget": "#forbidden",
        //         "HX-Reswap": "innerHTML",
        //         Data:        []byte(response),
        // })

        // c.Redirect(http.StatusSeeOther, "/cookie-setup")
        // r.HandleContext(c)
        // c.HTML(http.StatusOK, "1.cookieSetup.html", gin.H{
        //     "Cookie": cookieGofiID,
        // })
        return
    }  

    // time.Sleep(299999999 * time.Nanosecond) // to simulate 300ms of loading in the front when submiting form
    var Form sqlite.FinanceTracker // PostInsertRows
    if err := c.ShouldBind(&Form); err != nil {
        c.String(http.StatusBadRequest, "bad request: %v", err)
        return
    }
    const DateOnly = "2006-01-02" // YYYY-MM-DD
    t, err := time.Parse(DateOnly, Form.Date)
	if err != nil { // Always check errors even if they should not happen.
		panic(err)
	}
    if !strings.Contains(Form.FormPriceStr2Decimals, "."){ // add .00 if "." not present in string, equivalent of *100 with next step
        Form.FormPriceStr2Decimals = Form.FormPriceStr2Decimals + ".00"
    }
    safeInteger, _ := strconv.Atoi(strings.Replace(Form.FormPriceStr2Decimals, ".", "", 1))
    Form.PriceIntx100 = safeInteger
    Form.Year, _ = strconv.Atoi(t.Format("2006")) // YYYY
    Form.Month, _ = strconv.Atoi(t.Format("01")) // MM
    Form.Day, _ = strconv.Atoi(t.Format("02")) // DD
    Form.PaymentMethod = "CBtest"
    Form.GofiID = cookieGofiID
    fmt.Printf("before sqlite insert, form: %#s \n", &Form) // form: {2023-09-13 désig Supermarche 5.03}
    _, err = sqlite.InsertRowInFinanceTracker(&Form)

    tmpl := template.Must(template.ParseFiles("./html/templates/3.insertrows.html"))
    tmpl.ExecuteTemplate(c.Writer, "lastInsert", Form)
}

// sqlite
func sqliteInit(c *gin.Context) {
    sqlite.RunSQLite()
    c.HTML(http.StatusOK, "0.index.html", "")
}

func main() {
    router := gin.Default()

    // render HTML
    // https://gin-gonic.com/docs/examples/html-rendering/
	router.LoadHTMLGlob("html/**/*")
	//router.LoadHTMLFiles("templates/template1.html", "templates/template2.html")
    
    // SERVE STATICS
    router.StaticFile("/favicon.ico", "./img/favicon.ico")
    router.StaticFile("/favicon.png", "./img/favicon.png") // 32x32
    router.Static("/img", "./img")

    router.GET("/", index)

    router.GET("/cookie-setup", getCookieSetup)
    router.POST("/cookie-setup", postCookieSetup)

    router.GET("/insertrows", getinsertrows)
    router.POST("/insertrows", postinsertrows)

    router.GET("/sqlite", sqliteInit)

    router.Run("0.0.0.0:8082")
}
