package api

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"gofi/gofi/data/appdata"
	"gofi/gofi/data/sqlite"
)

// refaire avec https://pkg.go.dev/net/http#example-Server.Shutdown
func Shutdown(w http.ResponseWriter, r *http.Request) {
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	checkpointStr := "-"
	if userContext.IsAdmin {
		fmt.Printf("Shutting down... ShutDownFlag: %v\n", appdata.ShutDownFlag)
		if err := appdata.DB.Ping(); err != nil {
			appdata.ShutDownFlag = true
			fmt.Println("Cannot Ping")
			appdata.DB.Close()
			appdata.RenderAPIorUI(w, r, false, false, false, http.StatusInternalServerError, "Cannot Ping, application already closing", checkpointStr)
			return
		}
		appdata.ShutDownFlag = true
		fmt.Println("PRAGMA optimize then db.Close() called from /api/shutdown")
		appdata.DB.Exec("PRAGMA optimize;") // to run just before closing each database connection.
		appdata.DB.Close()
		ctx := r.Context()
		checkpointReturn := sqlite.WalCheckpoint(ctx) // checkpointReturn = 0 if OK
		if checkpointReturn == 0 {
			fmt.Println("Checkpoint réalisé avec succès!")
			checkpointStr = "Checkpoint done"
		} else {
			fmt.Println("Checkpoint non réalisé.")
			checkpointStr = "Checkpoint failed"
		}
		appdata.RenderAPIorUI(w, r, false, false, true, http.StatusOK, "closing application in 5 secs", checkpointStr)
		go exit()
		return
	}
	appdata.RenderAPIorUI(w, r, false, false, false, http.StatusForbidden, "can't shutdown the app", checkpointStr)
}

func exit() {
	fmt.Println("Shutting down in 5 secs")
	duration := time.Duration(5) * time.Second
	time.Sleep(duration)
	os.Exit(0)
}

/*
```powershell
curl -X GET -H "Content-Type: application/json" --include --location "http://localhost:8083/api/shutdown"
```
*/
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

${jsonb} = curl -X GET -H "sessionID: ${sessionID}" "${url}/api/dbpath" | ConvertFrom-Json
${dbpath} = ${jsonb}.jsonContent
${dbpath}

curl -H "Content-Type: application/json" -H "sessionID: ${sessionID}" "${url}/api/shutdown"
```
*/
