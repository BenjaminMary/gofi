package api

import (
	"fmt"
	"net/http"
	"strings"

	"gofi/gofi/data/appdata"
	"gofi/gofi/data/sqlite"

	"github.com/go-chi/render"
)

func GetParam(w http.ResponseWriter, r *http.Request, isFrontRequest bool,
	categoryTypeFilter string, categoryTypeFilterValue string, firstEmptyCategory bool) *appdata.HttpStruct {
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	var userParams appdata.UserParams
	userParams.GofiID = userContext.GofiID
	userCategories := appdata.NewUserCategories()
	userCategories.GofiID = userContext.GofiID
	sqlite.GetList(r.Context(), appdata.DB, &userParams, userCategories, categoryTypeFilter, categoryTypeFilterValue, firstEmptyCategory)
	userParams.AccountListUnhandled = sqlite.GetUnhandledAccountList(r.Context(), appdata.DB, userContext.GofiID, userParams.AccountList)
	userParams.Categories = userCategories
	return appdata.RenderAPIorUI(w, r, isFrontRequest, true, true, http.StatusOK, "user params retrieved", userParams)
}

func cleanStringList(stringList string, minStringLength int) string {
	var list []string
	var cleanedString string
	var cleanedStringResult string = ""
	list = strings.Split(stringList, ",")
	for _, element := range list {
		// remove all " " and "," from beginning and end
		cleanedString = strings.Trim(element, " ,")
		if len(cleanedString) >= minStringLength {
			if cleanedStringResult != "" {
				cleanedStringResult += ","
			}
			cleanedStringResult += cleanedString
		}
	}
	return cleanedStringResult
}
func postParam(w http.ResponseWriter, r *http.Request, isFrontRequest bool, paramName string, paramInfo string, minStringLength int) *appdata.HttpStruct {
	param := &appdata.Param{}
	if err := render.Bind(r, param); err != nil {
		fmt.Printf("error: %v\n", err.Error())
		return appdata.RenderAPIorUI(w, r, isFrontRequest, true, false, http.StatusBadRequest, "invalid request, double check each field", "")
	}
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	param.GofiID = userContext.GofiID
	param.ParamName = paramName
	param.ParamJSONstringData = cleanStringList(param.ParamJSONstringData, minStringLength)
	param.ParamInfo = paramInfo
	// fmt.Printf("param: %#v\n", param)
	_, err := sqlite.InsertRowInParam(r.Context(), appdata.DB, param)
	if err != nil { // Always check errors even if they should not happen.
		fmt.Printf("error: %v\n", err.Error())
		return appdata.RenderAPIorUI(w, r, isFrontRequest, true, false, http.StatusInternalServerError, "server error", "")
	}
	return appdata.RenderAPIorUI(w, r, isFrontRequest, true, true, http.StatusOK, "user param updated", param)
}
func PostParamAccount(w http.ResponseWriter, r *http.Request, isFrontRequest bool) *appdata.HttpStruct {
	return postParam(w, r, isFrontRequest,
		"accountList",
		"Liste des comptes (séparer par des , sans espaces)",
		2) // minStringLength to handle the value
}
func PostParamCategoryRendering(w http.ResponseWriter, r *http.Request, isFrontRequest bool) *appdata.HttpStruct {
	return postParam(w, r, isFrontRequest,
		"categoryRendering",
		"Affichage des catégories: icons | names",
		1) // minStringLength to handle the value
}
func PostParamOnboardingCheckList(w http.ResponseWriter, r *http.Request, isFrontRequest bool) *appdata.HttpStruct {
	return postParam(w, r, isFrontRequest,
		"onboardingCheckList",
		"Liste des étapes (séparer par des , sans espaces)",
		1) // minStringLength to handle the value
}
func UpdateParamCheckList(w http.ResponseWriter, r *http.Request, isFrontRequest bool, 
	gofiID int, onboardingCheckListSingleString string) *appdata.HttpStruct {
	param := &appdata.Param{}
	param.GofiID = gofiID
	param.ParamName = "onboardingCheckList"
	param.ParamJSONstringData = onboardingCheckListSingleString
	param.ParamInfo = "Liste des étapes (séparer par des , sans espaces)"
	_, err := sqlite.InsertRowInParam(r.Context(), appdata.DB, param)
	if err != nil { // Always check errors even if they should not happen.
		fmt.Printf("error: %v\n", err.Error())
		return appdata.RenderAPIorUI(w, r, isFrontRequest, true, false, http.StatusInternalServerError, "server error", "")
	}
	return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusOK, "onboardingCheckList updated", "")
}

func GetCategoryIcon(w http.ResponseWriter, r *http.Request, isFrontRequest bool, categoryNameFuncParam string, cd *appdata.CategoryDetails) *appdata.HttpStruct {
	categoryName := getURLorFUNCparam(r, categoryNameFuncParam, "")
	cd.CategoryIcon = "e909"
	cd.CategoryColorName = "system-lightgrey"
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	iconCodePoint, colorHEX, colorName := sqlite.GetCategoryIcon(r.Context(), appdata.DB, categoryName, userContext.GofiID)
	// fmt.Printf("GetCategoryIcon iconCodePoint: %v, colorHEX: %v \n", iconCodePoint, colorHEX)
	if iconCodePoint == "" || colorHEX == "" || colorName == "" {
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusNotFound, "category not found", "")
	}
	cd.CategoryIcon = iconCodePoint
	cd.CategoryColor = colorHEX
	cd.CategoryColorName = colorName
	return appdata.RenderAPIorUI(w, r, isFrontRequest, false, true, http.StatusOK, "category info found", cd)
}

func PutParamCategory(w http.ResponseWriter, r *http.Request, isFrontRequest bool) *appdata.HttpStruct {
	categoryPut := &appdata.CategoryPut{}
	if err := render.Bind(r, categoryPut); err != nil {
		fmt.Printf("error: %v\n", err.Error())
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusBadRequest, "invalid request, double check each field", "")
	}
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	categoryPut.GofiID = userContext.GofiID
	successBool := sqlite.PutCategory(r.Context(), appdata.DB, categoryPut)
	if !successBool {
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusNotFound, "category not updated", categoryPut.ID)
	}
	return appdata.RenderAPIorUI(w, r, isFrontRequest, false, true, http.StatusOK, "category updated", categoryPut.ID)
}

func PatchParamCategoryInUse(w http.ResponseWriter, r *http.Request, isFrontRequest bool) *appdata.HttpStruct {
	categoryInUse := &appdata.CategoryPatchInUse{}
	if err := render.Bind(r, categoryInUse); err != nil {
		fmt.Printf("error: %v\n", err.Error())
		return appdata.RenderAPIorUI(w, r, isFrontRequest, true, false, http.StatusBadRequest, "invalid request, double check each field", "")
	}
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	categoryInUse.GofiID = userContext.GofiID
	successBool := sqlite.PatchCategoryInUse(r.Context(), appdata.DB, categoryInUse)
	if !successBool {
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusNotFound, "category inUse not updated", categoryInUse.ID)
	}
	return appdata.RenderAPIorUI(w, r, isFrontRequest, false, true, http.StatusOK, "category inUse updated", categoryInUse.ID)
}

func PatchParamCategoryOrder(w http.ResponseWriter, r *http.Request, isFrontRequest bool) *appdata.HttpStruct {
	categoryOrder := &appdata.CategoryPatchOrder{}
	if err := render.Bind(r, categoryOrder); err != nil {
		fmt.Printf("error: %v\n", err.Error())
		return appdata.RenderAPIorUI(w, r, isFrontRequest, true, false, http.StatusBadRequest, "invalid request, double check each field", "")
	}
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	categoryOrder.GofiID = userContext.GofiID
	successBool := sqlite.PatchCategoryOrder(r.Context(), appdata.DB, categoryOrder)
	if !successBool {
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusNotFound, "category order not updated", "")
	}
	return appdata.RenderAPIorUI(w, r, isFrontRequest, false, true, http.StatusOK, "category order updated", "")
}
