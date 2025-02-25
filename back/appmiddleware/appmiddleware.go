package appmiddleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"gofi/gofi/data/appdata"
	"gofi/gofi/data/sqlite"
	"gofi/gofi/front"
	"gofi/gofi/front/htmlComponents"
)

func MaintenanceMode(next http.Handler) http.Handler {
	// code here run only one time at the start of the server (initialize)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// fmt.Printf("MaintenanceMode GlobalShutDownFlag: %v\n", api.GlobalShutDownFlag)
		if appdata.ShutDownFlag {
			fmt.Println("MaintenanceMode canceled the request")
			appdata.RenderAPIorUI(w, r, false, false, false, http.StatusInternalServerError, "the application is shutting down", "")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func AddContextUserAndTimeout(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userContext := &appdata.UserRequest{}
		userContext.IsAdmin = false
		userContext.IsAuthenticated = false
		userContext.IsFront = false
		// r.URL.Path is used while testing where r.RequestURI isn't
		// fmt.Printf("r.URL.Path: %v\n", r.URL.Path)
		// if len(r.URL.Path) < 5 {
		// 	userContext.IsFrontRequest = true
		// } else {
		// 	if r.URL.Path[0:5] == "/api/" {
		// 		userContext.IsFrontRequest = false
		// 	} else {
		// 		userContext.IsFrontRequest = true
		// 	}
		// }
		ctx := context.WithValue(r.Context(), appdata.ContextUserKey, userContext)
		ctx, cancel := context.WithTimeout(ctx, time.Duration(20*time.Second))
		// TODO: timeout lower than this except for csv part 
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func completeUserInfos(userContext *appdata.UserRequest, w http.ResponseWriter, r *http.Request, isCookie bool) {
	gofiID, email, errStr, err := sqlite.GetGofiID(r.Context(), appdata.DB, userContext.SessionID,
		r.Header.Get("User-Agent"), r.Header.Get("Accept-Language"), r.Header.Get(os.Getenv("HEADER_IP")))
	if err == nil && gofiID > 0 {
		validAuth := true
		if errStr == "idleTimeout, change cookie" {
			fmt.Println("middleware: auto refresh session ID")
			newSessionID, err := appdata.GenerateRandomString(appdata.CookieLength)
			if err == nil {
				errorStrReason, err := sqlite.UpdateSessionID(r.Context(), appdata.DB, gofiID, newSessionID)
				if err != nil {
					fmt.Printf("middleware: errorStrReason: %v\n", errorStrReason)
					fmt.Printf("middleware: err3: %v\n", err)
					validAuth = false
				}
				if isCookie {
					front.FrontSetCookie(w, newSessionID)
				}
			} else {
				fmt.Printf("middleware: err2: %v\n", err)
				validAuth = false
			}
		}
		if validAuth {
			if email == os.Getenv("ADMIN_EMAIL") || email == os.Getenv("ADMIN_EMAIL_B") {
				userContext.IsAdmin = true
			}
			userContext.IsAuthenticated = true
			userContext.GofiID = gofiID
			userContext.Email = email
		}
	} else if err != nil {
		fmt.Printf("middleware: errStr: %v\n", errStr)
		fmt.Printf("middleware: err: %v\n", err)
	}
	if !userContext.IsAuthenticated && gofiID > 0 {
		sqlite.Logout(r.Context(), appdata.DB, gofiID)
	}
}

func CheckHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
		sessionID := ""
		sessionID = r.Header.Get("sessionID")
		if sessionID == "" {
			// handle specific API request to download file from front
			if r.URL.Path[0:15] == "/api/csv/export" {
				cookieStruct, err := r.Cookie("gofiID")
				if err == nil {
					err = cookieStruct.Valid()
					if err == nil {
						sessionID = cookieStruct.Value
					}
				}
			}
		}
		if sessionID != "" && len(sessionID) == appdata.CookieLength {
			userContext.SessionID = sessionID
		}
		if len(userContext.SessionID) > 0 {
			completeUserInfos(userContext, w, r, false)
		}
		// fmt.Printf("CH Full userContext: %#v\n", userContext)
		next.ServeHTTP(w, r)
	})
}

func CheckCookie(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
		userContext.IsFront = true
		cookieStruct, err := r.Cookie("gofiID")
		if err == nil {
			err = cookieStruct.Valid()
			if err == nil && len(cookieStruct.Value) == appdata.CookieLength {
				userContext.SessionID = cookieStruct.Value
			}
		}
		if len(userContext.SessionID) > 0 {
			completeUserInfos(userContext, w, r, true)
		}
		// fmt.Printf("CC Full userContext: %#v\n", userContext)
		next.ServeHTTP(w, r)
	})
}

func HeaderContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("Content-Type", "application/json")
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func AuthenticatedUserOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
		if !userContext.IsAuthenticated {
			appdata.RenderAPIorUI(w, r, userContext.IsFront, false, false, http.StatusUnauthorized, "please login", "")
			if userContext.IsFront {
				// TODO: return a 401
				htmlComponents.Lost(false).Render(r.Context(), w)
			}
			return
		}
		next.ServeHTTP(w, r)
	})
}
