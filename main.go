package main

import (
    // "fmt"

    "net/http"
    "time"
    "strings"
    "strconv"
    "html/template"
    "os"
    // "encoding/json"

    "example.com/sqlite"

    "github.com/gin-gonic/gin"
    // "github.com/gin-gonic/gin/render"
)

// FUNC check cookieGofiID
func checkCookie(c *gin.Context) (string) {
    // try to read if a cookie exists, return to setup cookie otherwise
    cookieGofiID, err := c.Cookie("gofiID")
    if err != nil {
        if c.Request.Method == "GET" {
            c.Redirect(http.StatusSeeOther, "/cookie-setup")
            c.Abort()
            return ""
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
            return ""
        }
    } 
    return cookieGofiID
}

// index.html
func index(c *gin.Context) {
    c.HTML(http.StatusOK, "0.index.html", "")
}

// GET cookie
func getCookieSetup(c *gin.Context) {
    // try to read if a cookie exists, return "Aucun" otherwise
    cookieGofiID, err := c.Cookie("gofiID")
    if err != nil {
        cookieGofiID = "Aucun"
    }    
    c.HTML(http.StatusOK, "1.cookieSetup.html", gin.H{
        "Cookie": cookieGofiID,
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

    //read existing param with this gofiID, and create default param if none 
    sqlite.CheckIfIdExists(gofiID)

    c.SetSameSite(http.SameSiteLaxMode)
    c.SetCookie("gofiID", "", -1, "/", "", false, true)
    c.SetCookie("gofiID", gofiID, cookieDurationInt, "/", "", false, true)
    c.String(200, `<p>Cookie: %s</p><p id="hx-swap-oob1" hx-swap-oob="true">Nouveau Cookie enregistré.</p>`, gofiID)
}

// GET ParamSetup.html
func getParamSetup(c *gin.Context) {
    cookieGofiID := checkCookie(c)
    if c.IsAborted() {return}

    var Form sqlite.FinanceTracker
    Form.GofiID = cookieGofiID
    sqlite.GetList(&Form)
    c.HTML(http.StatusOK, "2.paramSetup.html", gin.H{
        "Form": Form,
    })
}

func cleanStringList(stringList string) string {
    var list []string
    var cleanedString string
    var cleanedStringResult string = ""
    list = strings.Split(stringList, ",")
    for _, element := range list {
        cleanedString = strings.Trim(element, " ,")
        if len(cleanedString) > 0 {
            if cleanedStringResult != "" {cleanedStringResult += ","}
            cleanedStringResult += cleanedString
        }
    }
    return cleanedStringResult
}
// POST ParamSetup.html
func postParamSetup(c *gin.Context) {
    cookieGofiID := checkCookie(c)
    if c.IsAborted() {return}

    var Form sqlite.Param
    var returnedString string
    Form.GofiID = cookieGofiID
    accountList := c.PostForm("accountList")
    categoryList := c.PostForm("categoryList")
    if accountList != "" {
        Form.ParamName = "accountList"
        Form.ParamJSONstringData = cleanStringList(accountList)
        Form.ParamInfo = "Liste des comptes (séparer par des , sans espaces)"
        returnedString = `<input type="text" id="accountList" name="accountList" value="` + accountList + `" aria-invalid="false" disabled />`
    }
    if categoryList != "" {
        Form.ParamName = "categoryList"
        Form.ParamJSONstringData = cleanStringList(categoryList)
        Form.ParamInfo = "Liste des catégories (séparer par des , sans espaces)"
        returnedString = `<input type="text" id="categoryList" name="categoryList" value="` + categoryList + `" aria-invalid="false" disabled />`
    }
    _, err := sqlite.InsertRowInParam(&Form)
	if err != nil { // Always check errors even if they should not happen.
		panic(err)
	}
    c.String(200, returnedString)
}

// GET InsertRows.html
func getinsertrows(c *gin.Context) {
    cookieGofiID := checkCookie(c)
    if c.IsAborted() {return}

    var Form sqlite.FinanceTracker
    var FTlist []sqlite.FinanceTracker
    Form.GofiID = cookieGofiID
    const DateOnly = "2006-01-02" // YYYY-MM-DD
    currentTime := time.Now()
    Form.Date = currentTime.Format(DateOnly) // YYYY-MM-DD
    sqlite.GetList(&Form)
    FTlist = sqlite.GetLastRowsInFinanceTracker(cookieGofiID)
	// fmt.Printf("\naccountList: %v\n", Form.AccountList)
	// fmt.Printf("\ncategoryList: %v\n", Form.CategoryList)
    c.HTML(http.StatusOK, "3.insertrows.html", gin.H{
        "Form": Form,
        "FTlist": FTlist,
    })
}

// POST InsertRows.html
func postinsertrows(c *gin.Context) {
    cookieGofiID := checkCookie(c)
    if c.IsAborted() {return}

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
    Form.GofiID = cookieGofiID
    // fmt.Printf("before sqlite insert, form: %#s \n", &Form) // form: {2023-09-13 désig Supermarche 5.03}
    _, err = sqlite.InsertRowInFinanceTracker(&Form)
	if err != nil { // Always check errors even if they should not happen.
		panic(err)
	}
    tmpl := template.Must(template.ParseFiles("./html/templates/3.insertrows.html"))
    tmpl.ExecuteTemplate(c.Writer, "lastInsert", Form)
}

// GET exportCsv.html
func getExportCsv(c *gin.Context) {
    checkCookie(c)
    if c.IsAborted() {return}

    c.HTML(http.StatusOK, "100.exportCsv.html", "")
}

// POST exportCsv.html
func postExportCsv(c *gin.Context) {
    cookieGofiID := checkCookie(c)
    if c.IsAborted() {return}

    csvSeparator := c.PostForm("csvSeparator")
    csvDecimalDelimiter := c.PostForm("csvDecimalDelimiter")

    var csvSeparatorRune rune
    for _, runeValue := range csvSeparator {csvSeparatorRune = runeValue}

    fileName := "gofi-" + cookieGofiID + ".csv"
    filePathWithName := sqlite.FilePath(fileName)
    defer os.Remove(filePathWithName)
    sqlite.ExportCSV(cookieGofiID, csvSeparatorRune, csvDecimalDelimiter)

    c.Header("Content-Disposition", "attachment; filename=" + fileName)
    c.Header("Content-Type", "text/plain")
    c.FileAttachment(filePathWithName, fileName)
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

    router.GET("/param-setup", getParamSetup)
    router.POST("/param-setup", postParamSetup)

    router.GET("/insertrows", getinsertrows)
    router.POST("/insertrows", postinsertrows)

    router.GET("/export-csv", getExportCsv)
    router.POST("/export-csv", postExportCsv)

    router.Run("0.0.0.0:8082")
}
