package api

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"os"

	"gofi/gofi/data/appdata"
	"gofi/gofi/data/sqlite"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

/* USELESS
func UserRead(w http.ResponseWriter, r *http.Request) {
	if gofiID := chi.URLParam(r, "userID"); gofiID != "" {
		fmt.Printf("In UserRead GofiID: %v\n", gofiID)

	}
}
*/
/*
```powershell
curl -X GET -H "Content-Type: application/json" -H "sessionID: 1" --include --location "http://localhost:8083/api/user/1"
```
*/

func UserLogin(w http.ResponseWriter, r *http.Request, isFrontRequest bool) *appdata.HttpStruct {
	// WARNING when the 1st reader is used, no other read can occur
	// bytedata, _ := io.ReadAll(r.Body)
	// fmt.Printf("UserLogin body: %v\n", string(bytedata))
	// fmt.Printf("UserLogin Header: %v\n", r.Header)

	userRequest := &appdata.UserRequest{}
	if err := render.Bind(r, userRequest); err != nil {
		// trigger if email or password = "" or one is missing/wrong
		fmt.Printf("err: %v\n", err)
		return appdata.RenderAPIorUI(w, r, isFrontRequest, true, false, http.StatusBadRequest, "invalid request, double check each field", "")
	}
	user := &appdata.User{}
	user.UserRequest = userRequest
	sessionID, err := appdata.GenerateRandomString(appdata.CookieLength)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return appdata.RenderAPIorUI(w, r, isFrontRequest, true, false, http.StatusInternalServerError, "server error", "")
	}
	user.SessionID = sessionID
	currentTimeStr := time.Now().Format(time.RFC3339)
	user.LastLoginTime = currentTimeStr
	user.LastActivityTime = currentTimeStr
	user.LastActivityIPaddress = r.RemoteAddr
	user.LastActivityUserAgent = r.Header.Get("User-Agent")
	user.LastActivityAcceptLanguage = r.Header.Get("Accept-Language")
	h := sha256.New()
	h.Write([]byte(user.Password))
	byteSlice := h.Sum(nil)
	user.PwHash = hex.EncodeToString(byteSlice)
	gofiID, errorStrReason, err := sqlite.CheckUserLogin(r.Context(), appdata.DB, user)
	if err != nil || gofiID < 1 {
		fmt.Println("error after CheckUserLogin in postLogin")
		fmt.Printf("errorStrReason: %v\n", errorStrReason)
		fmt.Printf("err: %v\n", err)
		return appdata.RenderAPIorUI(w, r, isFrontRequest, true, false, http.StatusUnauthorized, "login failed", "")
	}
	user.GofiID = gofiID
	user.Password = "*"
	return appdata.RenderAPIorUI(w, r, isFrontRequest, true, true, http.StatusOK, "login success", user)
}

/*
```powershell
# with password, trigger CheckUserLogin
curl -X POST -H "Content-Type: application/json" --include --location "http://localhost:8083/api/user/login" `
	-d '{"email":"abcd@d.e","password":"pw"}'

# with empty password
curl -X POST -H "Content-Type: application/json" --include --location "http://localhost:8083/api/user/login" `
	-d '{"email":"abcd@d.e","password":""}'
```
*/

func UserRefreshSession(w http.ResponseWriter, r *http.Request, isFrontRequest bool) *appdata.HttpStruct {
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	gofiIDstr := chi.URLParam(r, "userID")
	if gofiIDstr == "" {
		return appdata.RenderAPIorUI(w, r, isFrontRequest, true, false, http.StatusBadRequest, "refresh failed, please login", "")
	}
	gofiID, err := strconv.Atoi(gofiIDstr)
	if err != nil {
		fmt.Printf("err1: %v\n", err)
		return appdata.RenderAPIorUI(w, r, isFrontRequest, true, false, http.StatusBadRequest, "refresh failed, please login", "")
	}
	if gofiID > 0 && gofiID == userContext.GofiID {
		fmt.Printf("In UserRefreshSession GofiID: %v\n", gofiIDstr)
	} else {
		return appdata.RenderAPIorUI(w, r, isFrontRequest, true, false, http.StatusUnauthorized, "refresh failed, please login", "")
	}
	newSessionID, err := appdata.GenerateRandomString(appdata.CookieLength)
	if err != nil {
		fmt.Printf("err2: %v\n", err)
		return appdata.RenderAPIorUI(w, r, isFrontRequest, true, false, http.StatusInternalServerError, "server error", "")
	}
	userContext.SessionID = newSessionID
	errorStrReason, err := sqlite.UpdateSessionID(r.Context(), appdata.DB, userContext.GofiID, userContext.SessionID)
	if err != nil {
		fmt.Printf("errorStrReason: %v\n", errorStrReason)
		fmt.Printf("err3: %v\n", err)
		return appdata.RenderAPIorUI(w, r, isFrontRequest, true, false, http.StatusUnauthorized, "refresh failed, please login", "")
	}
	userContext.Password = "*"
	return appdata.RenderAPIorUI(w, r, isFrontRequest, true, true, http.StatusOK, "session ID refresh", userContext)
}

/*
```powershell
# sessionID only, trigger UpdateSessionID
curl -X GET -H "Content-Type: application/json" -H "sessionID: 1" --include --location "http://localhost:8083/api/user/refresh"
```
*/

func UserLogout(w http.ResponseWriter, r *http.Request, isFrontRequest bool) *appdata.HttpStruct {
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	sqlite.Logout(r.Context(), appdata.DB, userContext.GofiID)
	userContext.Password = "*"
	return appdata.RenderAPIorUI(w, r, isFrontRequest, true, true, http.StatusOK, "logged out", userContext)
}

/*
```powershell
# with password, trigger CheckUserLogin
curl -X GET -H "Content-Type: application/json" -H "sessionID: 1" --include --location "http://localhost:8083/api/user/logout"
```
*/

func UserCreate(w http.ResponseWriter, r *http.Request, isFrontRequest bool) *appdata.HttpStruct {
	var User appdata.User
	data := &appdata.UserRequest{}
	if err := render.Bind(r, data); err != nil {
		fmt.Printf("error1: %v\n", err.Error())
		return appdata.RenderAPIorUI(w, r, isFrontRequest, true, false, http.StatusBadRequest, "invalid request, double check each field", "")
	}
	User.UserRequest = data
	User.DateCreated = time.Now().Format(time.DateOnly)
	password := data.Password
	hash := sha256.New()
	hash.Write([]byte(password))
	byteSlice := hash.Sum(nil)
	User.PwHash = hex.EncodeToString(byteSlice)

	ctx := r.Context()
	gofiID, errorStrReason, err := sqlite.CreateUser(ctx, appdata.DB, User)
	if err != nil {
		fmt.Printf("errorStrReason: %v\n", errorStrReason)
		fmt.Printf("error2: %v\n", err.Error())
		return appdata.RenderAPIorUI(w, r, isFrontRequest, true, false, http.StatusInternalServerError, "can't create the requested user", "")
	}
	sqlite.CheckIfIdExists(ctx, appdata.DB, int(gofiID))
	sqlite.InitCategoriesForUser(ctx, appdata.DB, int(gofiID))
	if os.Getenv("NOTIFICATION_FLAG") == "1" {
		go notificationPost(os.Getenv("NOTIFICATION_URL") + "-newuser", 
			"Email: " + User.UserRequest.Email,
			"Creation",
			"new,partying_face")
	}
	return appdata.RenderAPIorUI(w, r, isFrontRequest, true, true, http.StatusCreated, "user created", User)
}

/*
```powershell
curl -X POST -H "Content-Type: application/json" -H "sessionID: 123test" --include --location "http://localhost:8083/api/user/create" `
	-d '{"email":"abcd@d.e","password":"pw"}'

curl -X POST -H "Content-Type: application/json" -H "sessionID: 123test"           --location "http://localhost:8083/api/user/create" `
	-d '{"email":"abc@d.e","password":"pw"}' | convertfrom-json | convertto-json -depth 20
```
*/

func UserDelete(w http.ResponseWriter, r *http.Request, isFrontRequest bool) *appdata.HttpStruct {
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	if userContext.IsAdmin {
		if gofiIDstr := chi.URLParam(r, "userID"); gofiIDstr != "" {
			fmt.Printf("GofiID: %v\n", gofiIDstr)
			gofiID, _ := strconv.Atoi(gofiIDstr)
			if gofiID != userContext.GofiID {
				fmt.Printf("GofiID DELETED: %v\n", gofiIDstr)
				errorStrReason, err := sqlite.DeleteUser(r.Context(), appdata.DB, gofiID)
				if err != nil {
					fmt.Printf("errorStrReason: %v\n", errorStrReason)
					fmt.Printf("error: %v\n", err.Error())
					return appdata.RenderAPIorUI(w, r, isFrontRequest, true, false, http.StatusInternalServerError, "can't delete the requested user", "")
				}
				return appdata.RenderAPIorUI(w, r, isFrontRequest, true, true, http.StatusOK, "user deleted", userContext)
			}
		}
	}
	return appdata.RenderAPIorUI(w, r, isFrontRequest, true, false, http.StatusForbidden, "can't delete the requested user", "")
}

/*
```powershell
curl -X DELETE -H "Content-Type: application/json" -H "sessionID: 1" --include --location "http://localhost:8083/api/user/2/delete"
```
*/
