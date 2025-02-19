package api

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"gofi/gofi/data/appdata"
	"gofi/gofi/data/sqlite"
)

// POST importCsv.html
func PostCSVimport(w http.ResponseWriter, r *http.Request, isFrontRequest bool) *appdata.HttpStruct {
	// ctx, cancel := context.WithTimeout(context.?(), 9*time.Second) // 9 sec in case of 10k rows import
	// defer cancel()

	// WARNING when the 1st reader is used, no other read can occur
	// bytedata, _ := io.ReadAll(r.Body)
	// fmt.Printf("PostCSVimport body: %v\n", string(bytedata))
	// fmt.Printf("PostCSVimport Header: %v\n", r.Header)

	// TODO add csv params in user params
	var Separator, DecimalDelimiter, DateFormat, DateSeparator string
	Separator = ";"
	DecimalDelimiter = ","
	DateFormat = "FR"
	DateSeparator = "/"

	_, csvHeader, err := r.FormFile("csvFile")
	if err != nil { // Always check errors even if they should not happen.
		fmt.Printf("PostCSVimport error on file: %v\n", err.Error())
		return appdata.RenderAPIorUI(w, r, isFrontRequest, true, false, http.StatusBadRequest, "invalid file", "error on file")
	}
	var csvSeparatorRune rune
	for _, runeValue := range Separator {
		csvSeparatorRune = runeValue
	}
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	stringList, errorBool := sqlite.ImportCSV(r.Context(), appdata.DB, userContext.GofiID, userContext.Email,
		csvSeparatorRune, DecimalDelimiter, DateFormat, DateSeparator, csvHeader)
	if errorBool {
		return appdata.RenderAPIorUI(w, r, isFrontRequest, true, false, http.StatusBadRequest, "Error, review file content", stringList)
	}
	if os.Getenv("NOTIFICATION_FLAG") == "1" {
		go notificationPost(os.Getenv("NOTIFICATION_URL") + "-csv", 
			"Email: " + userContext.Email,
			"CSV import",
			"memo,arrow_up")
	}
	return appdata.RenderAPIorUI(w, r, isFrontRequest, false, true, http.StatusOK, "CSV rows list", stringList)
}

/*
```powershell
cd C:\git\gofi\back\api\test\csv

${url} = "http://localhost:8083"
${email} = "x@example.com"
${mdp} = ""
${body} = @"
{`"email`": `"${email}`", `"password`": `"${mdp}`"}
"@
${json} = curl -X POST --location "${url}/api/user/login" -d ${body} | ConvertFrom-Json
${sessionID} = ${json}.jsonContent.sessionID
${sessionID}

curl -X POST -H "sessionID: ${sessionID}" "http://localhost:8083/csv/import" `
	-F "csvFile=@2024-02-24-export-gofi-id5-UTF8-LF.csv"
```
*/

// POST exportCsv.html
func PostCSVexport(w http.ResponseWriter, r *http.Request, isFrontRequest bool) *appdata.HttpStruct {
	// TODO add csv params in user params
	var Separator, DecimalDelimiter, DateFormat, DateSeparator string
	Separator = ";"
	DecimalDelimiter = ","
	DateFormat = "FR"
	DateSeparator = "/"

	var csvSeparatorRune rune
	for _, runeValue := range Separator {
		csvSeparatorRune = runeValue
	}
	var fileData []byte
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	fileName := "gofi-" + strconv.Itoa(userContext.GofiID) + ".csv"
	filePathWithName := appdata.SQLiteFilePath(fileName)
	fmt.Printf("filePathWithName: %v\n", filePathWithName)
	defer os.Remove(filePathWithName)
	_, fileData = sqlite.ExportCSV(r.Context(), appdata.DB, userContext.GofiID, csvSeparatorRune, DecimalDelimiter, DateFormat, DateSeparator)
	w.Header().Set("Content-Disposition", "inline; filename="+fileName)
	if os.Getenv("NOTIFICATION_FLAG") == "1" {
		go notificationPost(os.Getenv("NOTIFICATION_URL") + "-csv", 
			"Email: " + userContext.Email,
			"CSV export",
			"memo,arrow_down")
	}
	return appdata.RenderFile(w, r, isFrontRequest, false, true, http.StatusOK, "CSV file", "", fileData)
}

/*
```powershell
cd C:\git\gofi\back\api\test\csv
Invoke-Item .

${url} = "http://localhost:8083"
${email} = "x@example.com"
${mdp} = ""
${body} = @"
{`"email`": `"${email}`", `"password`": `"${mdp}`"}
"@
${json} = curl -X POST --location "${url}/api/user/login" -d ${body} | ConvertFrom-Json
${sessionID} = ${json}.jsonContent.sessionID
${sessionID}

curl -X POST -H "sessionID: ${sessionID}" "http://localhost:8083/csv/export" --output fileDownloadedWithCurl.csv
```
*/

func PostCSVexportReset(w http.ResponseWriter, r *http.Request, isFrontRequest bool) *appdata.HttpStruct {
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	sqlite.ExportCSVreset(r.Context(), appdata.DB, userContext.GofiID)
	return appdata.RenderAPIorUI(w, r, isFrontRequest, false, true, http.StatusOK, "CSV exported flag reseted", "")
}
