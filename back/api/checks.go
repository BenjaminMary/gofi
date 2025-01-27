package api

import (
	"net/http"

	"gofi/gofi/data/appdata"
)

func IsAdmin(w http.ResponseWriter, r *http.Request) {
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	if userContext.IsAdmin {
		appdata.RenderAPIorUI(w, r, false, false, true, http.StatusOK, "admin ok", "")
		return
	}
	appdata.RenderAPIorUI(w, r, false, false, false, http.StatusUnauthorized, "admin ko", "")
}

func IsAuthenticated(w http.ResponseWriter, r *http.Request) {
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	if userContext.IsAuthenticated {
		appdata.RenderAPIorUI(w, r, false, false, true, http.StatusOK, "authenticated ok", "")
		return
	}
	appdata.RenderAPIorUI(w, r, false, false, false, http.StatusUnauthorized, "authenticated ko", "")
}

/*
```powershell
curl -X GET -H "Content-Type: application/json" --include --location "http://localhost:8083/api/shutdown"
```
*/

func GetFullDbPath(w http.ResponseWriter, r *http.Request) {
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	if userContext.IsAdmin {
		appdata.RenderAPIorUI(w, r, false, false, true, http.StatusOK, "DB path ok", appdata.DbPath)
		return
	}
	appdata.RenderAPIorUI(w, r, false, false, false, http.StatusUnauthorized, "DB path ko", "")
}

// WINDOWS: "{\"isValidResponse\":true,...,\"jsonContent\":\"C:\\\\git\\\\gofi\\\\back\\\\data\\\\dbFiles\\\\test.db\"}"
// LINUX: {"isValidResponse":true,...,"jsonContent":"/.../back/data/dbFiles/gofi.db"}
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
```
*/
