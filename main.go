package main

import (
    "fmt"

    "net/http"
    "time"
    "strings"
    "strconv"
    "html/template"
    "os"
    "encoding/hex"
    "crypto/sha256"
    "crypto/rand"
    "math/big"
    // "encoding/json"

    "example.com/sqlite"

    "github.com/gin-gonic/gin"
    // "github.com/gin-gonic/gin/render"
)

// FUNC generateRandomString returns a securely generated random string
func generateRandomString(n int) (string, error) {
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
func setCookie(c *gin.Context, cookie string) {
    c.SetSameSite(http.SameSiteLaxMode)
    c.SetCookie("gofiID", "", -1, "/", "", true, true)
    c.SetCookie("gofiID", cookie, 2592000, "/", "", true, true) // 30d duration
    return
}

// FUNC check cookie sessionID to get the gofiID
func checkCookie(c *gin.Context) (int) {
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

// index.html
func index(c *gin.Context) {
    c.HTML(http.StatusOK, "0.index.html", "")
}

// GET createUser.html
func getCreateUser(c *gin.Context) {
    c.HTML(http.StatusOK, "0.createUser.html", "")
}

// POST createUser.html
func postCreateUser(c *gin.Context) {
    var User sqlite.User
    User.Email = c.PostForm("email")
    User.DateCreated = time.Now().Format(time.DateOnly)

    password := c.PostForm("password")
	h := sha256.New()
	h.Write([]byte(password))
	byteSlice := h.Sum(nil)
    User.PwHash = hex.EncodeToString(byteSlice)

    gofiID, errorStrReason, err := sqlite.CreateUser(User)
    if err != nil {
        fmt.Printf("errorStrReason: %v\n", errorStrReason)
        fmt.Printf("err: %v\n", err)
        c.Header("HX-Retarget", "#forbidden")
        c.Header("HX-Reswap", "innerHTML settle:500ms")
        c.String(http.StatusForbidden, `
            <div id="forbidden">
                <p>
                    ERREUR: Impossible de créer le compte.<br> 
                    Requête annulée, merci de recommencer.<br> 
                    Si l'erreur persiste, merci de changer d'email.
                </p>
            </div>
        `)
        return
    }
    sqlite.CheckIfIdExists(int(gofiID))
    c.String(http.StatusOK, "<div>Création du compte terminée.<br>Merci de procéder à la connexion.</div>")
}

// GET login.html
func getLogin(c *gin.Context) {
    c.HTML(http.StatusOK, "1.login.html", "")
}

// POST login.html
func postLogin(c *gin.Context) {
    var User sqlite.User
    User.Email = c.PostForm("email")

    password := c.PostForm("password")
	h := sha256.New()
	h.Write([]byte(password))
	byteSlice := h.Sum(nil)
    User.PwHash = hex.EncodeToString(byteSlice)

    sessionID, err := generateRandomString(32)
    if err != nil {
        fmt.Printf("err: %v\n", err)
        c.Header("HX-Retarget", "#forbidden")
        c.Header("HX-Reswap", "innerHTML settle:500ms")
        c.String(http.StatusForbidden, `
            <div id="forbidden">
                <p>
                    ERREUR: problème interne à l'application.<br> 
                    Requête annulée, merci de recommencer.
                </p>
            </div>
        `)
        return
    }
    User.SessionID = sessionID

    currentTimeStr := time.Now().Format(time.RFC3339)
    User.LastLoginTime = currentTimeStr
    User.LastActivityTime = currentTimeStr

	User.LastActivityIPaddress = c.Request.RemoteAddr
	User.LastActivityUserAgent = c.Request.Header.Get("User-Agent")
	User.LastActivityAcceptLanguage = c.Request.Header.Get("Accept-Language")

    gofiID, errorStrReason, err := sqlite.CheckUserLogin(User)
    if err != nil {
        fmt.Printf("errorStrReason: %v\n", errorStrReason)
        fmt.Printf("err: %v\n", err)
        c.Header("HX-Retarget", "#forbidden")
        c.Header("HX-Reswap", "innerHTML settle:500ms")
        c.String(http.StatusForbidden, `
            <div id="forbidden">
                <p>
                    ERREUR: email ou mot de passe incorrect.<br> 
                    Requête annulée, merci de recommencer.<br> 
                    Si l'erreur persiste, merci de réessayer plus tard, il pourrait y avoir un problème sur le serveur.
                </p>
            </div>
        `)
        return
    }
    User.GofiID = gofiID

    setCookie(c, User.SessionID)
    c.String(http.StatusOK, "<div>Login terminé.</div>")
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
    fmt.Printf("before FormPriceStr2Decimals: %s \n", Form.FormPriceStr2Decimals)
    if !strings.Contains(Form.FormPriceStr2Decimals, "."){ // add .00 if "." not present in string, equivalent of *100 with next step
        Form.FormPriceStr2Decimals = Form.FormPriceStr2Decimals + ".00"
    } else {
        decimalPart := strings.Split(Form.FormPriceStr2Decimals, ".")
        if len(decimalPart[1]) == 1 { Form.FormPriceStr2Decimals = Form.FormPriceStr2Decimals + "0" }
    }
    safeInteger, _ := strconv.Atoi(strings.Replace(Form.FormPriceStr2Decimals, ".", "", 1))
    Form.PriceIntx100 = safeInteger
    fmt.Printf("after PriceIntx100: %s \n", Form.PriceIntx100)

    var successfull bool
    Form.Year, Form.Month, Form.Day, successfull, _ = sqlite.ConvertDateStrToInt(Form.Date, "EN", "-")
    if !successfull {return}

    Form.GofiID = cookieGofiID
    // fmt.Printf("before sqlite insert, form: %#s \n", &Form) // form: {2023-09-13 désig Supermarche 5.03}
    _, err := sqlite.InsertRowInFinanceTracker(&Form)
	if err != nil { // Always check errors even if they should not happen.
		panic(err)
	}
    tmpl := template.Must(template.ParseFiles("./html/templates/3.insertrows.html"))
    tmpl.ExecuteTemplate(c.Writer, "lastInsert", Form)
}

// GET exportCsv.html
func getExportCsv(c *gin.Context) {
    cookieGofiID := checkCookie(c)
    if c.IsAborted() {return}

    FileName := "gofi-" + strconv.Itoa(cookieGofiID) + ".csv"
    c.HTML(http.StatusOK, "100.exportCsv.html", gin.H{
        "FileName": FileName,
    })
}

// POST exportCsv.html
func postExportCsv(c *gin.Context) {
    cookieGofiID := checkCookie(c)
    if c.IsAborted() {return}

    csvSeparator := c.PostForm("csvSeparator")
    csvDecimalDelimiter := c.PostForm("csvDecimalDelimiter")
    dateFormat := c.PostForm("dateFormat")
    dateSeparator := c.PostForm("dateSeparator")

    var csvSeparatorRune rune
    for _, runeValue := range csvSeparator {csvSeparatorRune = runeValue}

    fileName := "gofi-" + strconv.Itoa(cookieGofiID) + ".csv"
    filePathWithName := sqlite.FilePath(fileName)
    defer os.Remove(filePathWithName)
    sqlite.ExportCSV(cookieGofiID, csvSeparatorRune, csvDecimalDelimiter, dateFormat, dateSeparator)

    c.Header("Content-Disposition", "attachment; filename=" + fileName)
    c.Header("Content-Type", "text/plain")
    c.FileAttachment(filePathWithName, fileName)
}

// GET importCsv.html
func getImportCsv(c *gin.Context) {
    cookieGofiID := checkCookie(c)
    if c.IsAborted() {return}

    c.HTML(http.StatusOK, "101.importCsv.html", gin.H{
        "cookieGofiID": cookieGofiID,
    })
}

// POST importCsv.html
func postImportCsv(c *gin.Context) {
    cookieGofiID := checkCookie(c)
    if c.IsAborted() {return}

    csvSeparator := c.PostForm("csvSeparator")
    csvDecimalDelimiter := c.PostForm("csvDecimalDelimiter")
    dateFormat := c.PostForm("dateFormat")
    dateSeparator := c.PostForm("dateSeparator")
    csvFile, err := c.FormFile("csvFile")
	if err != nil { // Always check errors even if they should not happen.
        c.String(http.StatusBadRequest, `
            ERREUR: Problème de chargement du fichier csv.
            Merci de vérifier le format du fichier et réessayer.
        `)
        return
	}

    var csvSeparatorRune rune
    for _, runeValue := range csvSeparator {csvSeparatorRune = runeValue}

    stringList := sqlite.ImportCSV(cookieGofiID, csvSeparatorRune, csvDecimalDelimiter, dateFormat, dateSeparator, csvFile)

    c.String(http.StatusOK, stringList)
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

    router.GET("/param-setup", getParamSetup)
    router.POST("/param-setup", postParamSetup)

    router.GET("/insertrows", getinsertrows)
    router.POST("/insertrows", postinsertrows)

    router.GET("/export-csv", getExportCsv)
    router.POST("/export-csv", postExportCsv)

    router.GET("/import-csv", getImportCsv)
    router.POST("/import-csv", postImportCsv)

    router.GET("/login", getLogin)
    router.POST("/login", postLogin)

    router.GET("/createUser", getCreateUser)
    router.POST("/createUser", postCreateUser)

    router.Run("0.0.0.0:8082")
}
