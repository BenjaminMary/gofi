package main

import (
    "fmt"
	"database/sql"
    "context"

    "net/http"
    "time"
    "strings"
    "strconv"
    "html/template"
    "os"
    "encoding/hex"
    "crypto/sha256"

    "example.com/sqlite"
    "example.com/drive"

    "github.com/gin-gonic/gin"
)

var db *sql.DB
/*  //deadline with gin context
	deadline := time.Now().Add(1500 * time.Millisecond)
	ctx, cancelCtx := context.WithDeadline(ctx, deadline)
*/

//getBackup
func getBackup(c *gin.Context) {
    ctx, cancel := context.WithTimeout(context.TODO(), 9*time.Second)
    defer cancel()

    _, email := CheckCookie(ctx, c, db)
    if c.IsAborted() {return}

    //reserved admin page
    CheckAdmin(c, email)
    if c.IsAborted() {return}

    //get the URL param value named "save"
    backup := c.DefaultQuery("save", "0")

    driveSaveEnabledStr := os.Getenv("DRIVE_SAVE_ENABLED")
    DriveSaveEnabled, _ := strconv.Atoi(driveSaveEnabledStr)
    var DriveFileMetaData drive.DriveFileMetaData
    DriveFileMetaData.DriveFileID = ""
    DriveFileMetaData.Name = ""
    var DriveFileMetaDataList drive.DriveFileMetaDataList

    // checkpointReturn = 0 if OK
    checkpointReturn := sqlite.WalCheckpoint(ctx)
    var CheckpointReturnInfo string
    if (checkpointReturn == 0) {
        CheckpointReturnInfo = "Checkpoint réalisé avec succès!"
    } else {CheckpointReturnInfo = "Checkpoint non réalisé."}
    if (checkpointReturn == 0 && DriveSaveEnabled == 1 && backup == "1") {
        //backup db
        DriveFileMetaData = drive.UploadWithDrivePostRequestAPI(sqlite.DbPath)
        today := time.Now().Format(time.DateOnly)
        DriveFileMetaData.Name = today + "-gofi.db"
        drive.UpdateMetaDataDriveFile(DriveFileMetaData)
    }
    if (DriveSaveEnabled == 1) {
        DriveFileMetaDataList = drive.ListFileInDrive()
    }
    db = sqlite.OpenDbCon()

    c.HTML(http.StatusOK, "1.checkpoint.html", gin.H{
        "DriveFileMetaData": DriveFileMetaData,
        "DriveFileMetaDataList": DriveFileMetaDataList,
        "CheckpointReturnInfo": CheckpointReturnInfo,
        "DriveSaveEnabled": DriveSaveEnabled,
    })
}
func postBackup(c *gin.Context) {
    ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second)
    defer cancel()

    _, email := CheckCookie(ctx, c, db)
    if c.IsAborted() {return}

    //reserved admin page
    CheckAdmin(c, email)
    if c.IsAborted() {return}

    method := c.PostForm("method")
    driveID := c.PostForm("driveID")
    fmt.Println("method: ", method)
    // fmt.Println("driveID: ", driveID)

    if (method == "DELETE" && driveID != "none") {
        drive.DeleteFileInDrive(driveID)
        c.String(http.StatusOK, "Suppression effectuée")
        return
    }
    if (method == "DOWNLOAD" && driveID != "none") {
        filePathWithName := sqlite.FilePath("downloaded-gofi.db")
        fileName := "downloaded-gofi.db"
        drive.GetFileInDrive(driveID, filePathWithName)
        //c.String(http.StatusOK, "Téléchargement terminé")
        c.Header("Content-Disposition", "attachment; filename=" + fileName)
        c.File(filePathWithName)
        return
    }
    c.String(http.StatusBadRequest, "Erreur")
}

// index.html
func index(c *gin.Context) {
    ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second)
    defer cancel()

    gofiID, Email := CheckCookie(ctx, c, db)
    var Logged, Admin bool
    Admin = false
    if gofiID > 0 {
        Logged = true
        if Email == os.Getenv("ADMIN_EMAIL") {Admin = true}
    } else {Logged = false}
    c.HTML(http.StatusOK, "0.index.html", gin.H{
        "Logged": Logged,
        "Admin": Admin,
        "Email": Email,
    })
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

// GET logout.html
func getLogout(c *gin.Context) {
    ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second)
    defer cancel()

    gofiID, Email := CheckCookie(ctx, c, db)

    SetCookie(c, "logged-out")
    successfull, errorStrReason, err := sqlite.ForceNewLogin(gofiID)
    var Info string
    if successfull {Info = "Déconnexion réussie, à très vite"} else {
        fmt.Printf("errorStrReason: %v\n", errorStrReason)
        fmt.Printf("err: %v\n", err)
        Info = "Déjà déconnecté au moment de la demande"
    }
    c.HTML(http.StatusOK, "1.logout.html", gin.H{
        "Info": Info,
        "Email": Email,
    })
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

    sessionID, err := GenerateRandomString(CookieLength)
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

    SetCookie(c, User.SessionID)
    c.String(http.StatusOK, "<div>Login terminé, bienvenue <code>" + User.Email + "</code>.</div>")
}

// GET ParamSetup.html
func getParamSetup(c *gin.Context) {
    ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second)
    defer cancel()

    cookieGofiID, _ := CheckCookie(ctx, c, db)
    if c.IsAborted() {return}

    var UserParams sqlite.UserParams
    UserParams.GofiID = cookieGofiID
    sqlite.GetList(ctx, db, &UserParams)
    c.HTML(http.StatusOK, "2.paramSetup.html", gin.H{
        "UserParams": UserParams,
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
    ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second)
    defer cancel()

    cookieGofiID, _ := CheckCookie(ctx, c, db)
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
    ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second)
    defer cancel()

    cookieGofiID, _ := CheckCookie(ctx, c, db)
    if c.IsAborted() {return}

    var UserParams sqlite.UserParams
    UserParams.GofiID = cookieGofiID
    sqlite.GetList(ctx, db, &UserParams)

    var Form sqlite.FinanceTracker
    const DateOnly = "2006-01-02" // YYYY-MM-DD
    currentTime := time.Now()
    Form.Date = currentTime.Format(DateOnly) // YYYY-MM-DD

    var Filter sqlite.FilterRows
    Filter.GofiID = cookieGofiID
    Filter.OrderBy = "id"
    Filter.OrderByType = "DESC"
    Filter.Limit = 5
    var FTlist []sqlite.FinanceTracker
    FTlist, _ = sqlite.GetRowsInFinanceTracker(ctx, db, &Filter)

    c.HTML(http.StatusOK, "3.insertrows.html", gin.H{
        "Form": Form,
        "FTlist": FTlist,
        "UserParams": UserParams,
    })
}

// POST InsertRows.html
func postinsertrows(c *gin.Context) {
    ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second)
    defer cancel()

    cookieGofiID, _ := CheckCookie(ctx, c, db)
    if c.IsAborted() {return}

    // time.Sleep(299999999 * time.Nanosecond) // to simulate 300ms of loading in the front when submiting form
    var Form sqlite.FinanceTracker // PostInsertRows
    if err := c.ShouldBind(&Form); err != nil {
        c.String(http.StatusBadRequest, "bad request: %v", err)
        return
    }
    Form.PriceIntx100 = sqlite.ConvertPriceStrToInt(Form.FormPriceStr2Decimals, ".") // always "." as decimal separator from the form

    var successfull bool
    Form.Year, Form.Month, Form.Day, successfull, _ = sqlite.ConvertDateStrToInt(Form.Date, "EN", "-")
    if !successfull {return}

    Form.GofiID = cookieGofiID
    // fmt.Printf("before sqlite insert, form: %#s \n", &Form) // form: {2023-09-13 désig Supermarche 5.03}
    _, err := sqlite.InsertRowInFinanceTracker(ctx, db, &Form)
	if err != nil { // Always check errors even if they should not happen.
		panic(err)
	}
    tmpl := template.Must(template.ParseFiles("./html/templates/3.insertrows.html"))
    tmpl.ExecuteTemplate(c.Writer, "lastInsert", Form)
}


// GET EditRows.html
func getEditRows(c *gin.Context) {
    ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second)
    defer cancel()

    cookieGofiID, _ := CheckCookie(ctx, c, db)
    if c.IsAborted() {return}

    var UserParams sqlite.UserParams
    UserParams.GofiID = cookieGofiID
    sqlite.GetList(ctx, db, &UserParams)

    var FTlist []sqlite.FinanceTracker
    var TotalPriceStr2Decimals string
    var Filter sqlite.FilterRows
    Filter.GofiID = cookieGofiID
    Filter.OrderBy = "id"
    Filter.OrderByType = "DESC"
    Filter.Limit = 20
    FTlist, TotalPriceStr2Decimals = sqlite.GetRowsInFinanceTracker(ctx, db, &Filter)

    c.HTML(http.StatusOK, "4.editrows.html", gin.H{
        "UserParams": UserParams,
        "FTlist": FTlist,
        "TotalPriceStr2Decimals": TotalPriceStr2Decimals,
    })
}
// POST EditRows.html
func postEditRows(c *gin.Context) {
    ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second)
    defer cancel()

    cookieGofiID, _ := CheckCookie(ctx, c, db)
    if c.IsAborted() {return}

    var FTlistPost []sqlite.FinanceTracker
    var TotalPriceStr2Decimals string
    var Filter sqlite.FilterRows
    Filter.GofiID = cookieGofiID

    Filter.WhereAccount = c.PostForm("compte")
    Filter.WhereCategory = c.PostForm("categorie")

    whereYearStr := c.PostForm("annee")
	fmt.Printf("whereYearStr: %#v, type:%T\n", whereYearStr, whereYearStr) // check default value and type
    if whereYearStr != "" {Filter.WhereYear, _ = strconv.Atoi(whereYearStr)}
    whereMonthStr := c.PostForm("mois")
    if whereMonthStr != "" {Filter.WhereMonth, _ = strconv.Atoi(whereMonthStr)}
    whereCheckedStr := c.PostForm("checked")
    if whereCheckedStr != "" {Filter.WhereChecked, _ = strconv.Atoi(whereCheckedStr)}

    Filter.OrderBy = c.PostForm("orderBy")
    Filter.OrderByType = c.PostForm("orderByType")

    limitStr := c.PostForm("limit")
    Filter.Limit, _ = strconv.Atoi(limitStr)

    FTlistPost, TotalPriceStr2Decimals = sqlite.GetRowsInFinanceTracker(ctx, db, &Filter)

    tmpl := template.Must(template.ParseFiles("./html/templates/4.editrows.html"))
    tmpl.ExecuteTemplate(c.Writer, "listEditRows", gin.H{
        "FTlistPost": FTlistPost,
        "TotalPriceStr2Decimals": TotalPriceStr2Decimals,
    })
}

// GET stats
// func getStats(c *gin.Context) {
//     ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second)
//     defer cancel()

//     cookieGofiID, _ := CheckCookie(ctx, c, db)
//     if c.IsAborted() {return}

//     var TotalPerAccount map[string]int
//     TotalPerAccount = make(map[string]string)
//     TotalPerAccount = sqlite.GetRowsInFinanceTracker(ctx, db, &Filter)

//     c.HTML(http.StatusOK, "2.stats.html", gin.H{
//         "FileName": FileName,
//     })
// }

// GET exportCsv.html
func getExportCsv(c *gin.Context) {
    ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second)
    defer cancel()

    cookieGofiID, _ := CheckCookie(ctx, c, db)
    if c.IsAborted() {return}

    FileName := "gofi-" + strconv.Itoa(cookieGofiID) + ".csv"
    c.HTML(http.StatusOK, "100.exportCsv.html", gin.H{
        "FileName": FileName,
    })
}

// POST exportCsv.html
func postExportCsv(c *gin.Context) {
    ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second)
    defer cancel()

    cookieGofiID, _ := CheckCookie(ctx, c, db)
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
    ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second)
    defer cancel()

    cookieGofiID, _ := CheckCookie(ctx, c, db)
    if c.IsAborted() {return}

    c.HTML(http.StatusOK, "101.importCsv.html", gin.H{
        "cookieGofiID": cookieGofiID,
    })
}

// POST importCsv.html
func postImportCsv(c *gin.Context) {
    ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second)
    defer cancel()

    cookieGofiID, _ := CheckCookie(ctx, c, db)
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

    fmt.Println("------------------ROUTER START HERE------------------")
    fmt.Println("start db from main")
    db = sqlite.OpenDbCon()
	defer db.Close()
	defer fmt.Println("defer : db.Close() from main")

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

    router.GET("/editrows", getEditRows)
    router.POST("/editrows", postEditRows)

    // router.GET("/stats", getStats)

    router.GET("/export-csv", getExportCsv)
    router.POST("/export-csv", postExportCsv)

    router.GET("/import-csv", getImportCsv)
    router.POST("/import-csv", postImportCsv)

    router.GET("/admin/backup", getBackup)
    router.POST("/admin/backup", postBackup)

    router.GET("/login", getLogin)
    router.POST("/login", postLogin)
    router.GET("/logout", getLogout)

    router.GET("/createUser", getCreateUser)
    router.POST("/createUser", postCreateUser)

    router.Run("0.0.0.0:8083")
}
