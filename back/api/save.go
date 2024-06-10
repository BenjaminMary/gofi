package api

import (
	"fmt"
	"net/http"
	"strconv"

	"gofi/gofi/data/appdata"
	"gofi/gofi/data/sqlite"

	"github.com/go-chi/render"
)

// r.Get("/", api.SaveRead)      // GET /api/save
// r.Post("/", api.SaveCreate)   // POST /api/save
// r.Put("/", api.SaveEdit)      // PUT /api/save
// r.Delete("/", api.SaveDelete) // DELETE /api/save

func SaveRead(w http.ResponseWriter, r *http.Request) {
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	if userContext.IsAdmin {
		saveBackupList, errorStrReason, err := sqlite.SaveSelect(r.Context(), appdata.DB)
		if err != nil {
			fmt.Printf("SaveRead errorStrReason: %v\n", errorStrReason)
			fmt.Printf("SaveRead err: %v\n", err)
			appdata.RenderAPIorUI(w, r, false, false, false, http.StatusInternalServerError, "can't read backup", "")
			return
		}
		appdata.RenderAPIorUI(w, r, false, false, true, http.StatusOK, "list of backup", saveBackupList)
		return
	}
	appdata.RenderAPIorUI(w, r, false, false, false, http.StatusForbidden, "can't read backup", "")
}

func SaveCreate(w http.ResponseWriter, r *http.Request) {
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	if userContext.IsAdmin {
		saveBackup := appdata.SaveBackup{}
		if err := render.Bind(r, &saveBackup); err != nil {
			fmt.Printf("SaveCreate error1: %v\n", err.Error())
			appdata.RenderAPIorUI(w, r, false, false, false, http.StatusBadRequest, "invalid request, double check each field", "")
			return
		}
		saveBackup.Tested = "0"
		_, errorStrReason, err := sqlite.SaveCreate(r.Context(), appdata.DB, &saveBackup)
		if err != nil {
			fmt.Printf("SaveRead errorStrReason: %v\n", errorStrReason)
			fmt.Printf("SaveRead err: %v\n", err)
			appdata.RenderAPIorUI(w, r, false, false, false, http.StatusInternalServerError, "can't read backup", "")
			return
		}
		appdata.RenderAPIorUI(w, r, false, false, true, http.StatusCreated, "backup created", "")
		return
	}
	appdata.RenderAPIorUI(w, r, false, false, false, http.StatusForbidden, "can't create backup", "")
}

// func SaveEdit(w http.ResponseWriter, r *http.Request) {
// 	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
// 	if userContext.IsAdmin {

// 		// fmt.Printf("SaveEdit: %v\n", ?)
// 		appdata.RenderAPIorUI(w, r, false, false, true, http.StatusOK, "backup edited", "")
// 		return

// 	}
// 	appdata.RenderAPIorUI(w, r, false, false, false, http.StatusForbidden, "can't edit backup", "")
// }

func SaveDelete(w http.ResponseWriter, r *http.Request) {
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	if userContext.IsAdmin {
		idStr := getURLorFUNCparam(r, "", "id")
		idInt, err := strconv.Atoi(idStr)
		if err != nil { // Always check errors even if they should not happen.
			fmt.Printf("SaveDelete error1: %v\n", err.Error())
			appdata.RenderAPIorUI(w, r, false, false, false, http.StatusBadRequest, "invalid request, double check each field", "")
			return
		}
		if idInt < 1 { // Always check errors even if they should not happen.
			fmt.Printf("SaveDelete idInt: %v\n", idInt)
			appdata.RenderAPIorUI(w, r, false, false, false, http.StatusBadRequest, "invalid request, double check each field", "")
			return
		}
		errorStrReason, err := sqlite.SaveDelete(r.Context(), appdata.DB, idInt)
		if err != nil {
			fmt.Printf("SaveDelete errorStrReason: %v\n", errorStrReason)
			fmt.Printf("SaveDelete err: %v\n", err)
			appdata.RenderAPIorUI(w, r, false, false, false, http.StatusInternalServerError, "can't delete backup", "")
			return
		}
		appdata.RenderAPIorUI(w, r, false, false, true, http.StatusOK, "backup deleted", "")
		return
	}
	appdata.RenderAPIorUI(w, r, false, false, false, http.StatusForbidden, "can't delete backup", "")
}

func SaveDeleteKeepX(w http.ResponseWriter, r *http.Request) {
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	if userContext.IsAdmin {
		numStr := getURLorFUNCparam(r, "", "num")
		numInt, err := strconv.Atoi(numStr)
		if err != nil { // Always check errors even if they should not happen.
			fmt.Printf("SaveDeleteKeepX error1: %v\n", err.Error())
			appdata.RenderAPIorUI(w, r, false, false, false, http.StatusBadRequest, "invalid request, double check each field", "")
			return
		}
		if numInt < 1 { // Always check errors even if they should not happen.
			fmt.Printf("SaveDeleteKeepX numInt: %v\n", numInt)
			appdata.RenderAPIorUI(w, r, false, false, false, http.StatusBadRequest, "invalid request, double check each field", "")
			return
		}
		driveIDlist, errorStrReason, err := sqlite.SaveDeleteKeepX(r.Context(), appdata.DB, numInt)
		if err != nil {
			fmt.Printf("SaveDeleteKeepX errorStrReason: %v\n", errorStrReason)
			fmt.Printf("SaveDeleteKeepX err: %v\n", err)
			appdata.RenderAPIorUI(w, r, false, false, false, http.StatusInternalServerError, "can't delete backup", "")
			return
		}
		appdata.RenderAPIorUI(w, r, false, false, true, http.StatusOK, "backup deleted", driveIDlist)
		return
	}
	appdata.RenderAPIorUI(w, r, false, false, false, http.StatusForbidden, "can't delete backup", "")
}
