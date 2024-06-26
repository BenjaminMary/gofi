package api

import (
	"fmt"
	"net/http"
	"strings"

	"gofi/gofi/data/appdata"
	"gofi/gofi/data/sqlite"

	"github.com/go-chi/render"
)

func GetParamSetup(w http.ResponseWriter, r *http.Request, isFrontRequest bool) *appdata.HttpStruct {
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	var userParams appdata.UserParams
	userParams.GofiID = userContext.GofiID
	sqlite.GetList(r.Context(), appdata.DB, &userParams)
	return appdata.RenderAPIorUI(w, r, isFrontRequest, true, true, http.StatusOK, "user params retrieved", userParams)
}

func cleanStringList(stringList string) string {
	var list []string
	var cleanedString string
	var cleanedStringResult string = ""
	list = strings.Split(stringList, ",")
	for _, element := range list {
		cleanedString = strings.Trim(element, " ,")
		if len(cleanedString) > 0 {
			if cleanedStringResult != "" {
				cleanedStringResult += ","
			}
			cleanedStringResult += cleanedString
		}
	}
	return cleanedStringResult
}
func postParamSetup(w http.ResponseWriter, r *http.Request, isFrontRequest bool, paramName string, paramInfo string) *appdata.HttpStruct {
	param := &appdata.Param{}
	if err := render.Bind(r, param); err != nil {
		fmt.Printf("error: %v\n", err.Error())
		return appdata.RenderAPIorUI(w, r, isFrontRequest, true, false, http.StatusBadRequest, "invalid request, double check each field", "")
	}
	userContext := r.Context().Value(appdata.ContextUserKey).(*appdata.UserRequest)
	param.GofiID = userContext.GofiID
	param.ParamName = paramName
	param.ParamJSONstringData = cleanStringList(param.ParamJSONstringData)
	param.ParamInfo = paramInfo
	_, err := sqlite.InsertRowInParam(r.Context(), appdata.DB, param)
	if err != nil { // Always check errors even if they should not happen.
		fmt.Printf("error: %v\n", err.Error())
		return appdata.RenderAPIorUI(w, r, isFrontRequest, true, false, http.StatusInternalServerError, "server error", "")
	}
	return appdata.RenderAPIorUI(w, r, isFrontRequest, true, true, http.StatusOK, "user param updated", param)
}
func PostParamSetupAccount(w http.ResponseWriter, r *http.Request, isFrontRequest bool) *appdata.HttpStruct {
	return postParamSetup(w, r, isFrontRequest,
		"accountList",
		"Liste des comptes (séparer par des , sans espaces)")
}
func PostParamSetupCategory(w http.ResponseWriter, r *http.Request, isFrontRequest bool) *appdata.HttpStruct {
	return postParamSetup(w, r, isFrontRequest,
		"categoryList",
		"Liste des catégories (séparer par des , sans espaces)")
}
func PostParamSetupCategoryRendering(w http.ResponseWriter, r *http.Request, isFrontRequest bool) *appdata.HttpStruct {
	return postParamSetup(w, r, isFrontRequest,
		"categoryRendering",
		"Affichage des catégories: icons | names")
}

func GetCategoryIcon(w http.ResponseWriter, r *http.Request, isFrontRequest bool, categoryNameFuncParam string, cd *appdata.CategoryDetails) *appdata.HttpStruct {
	categoryName := getURLorFUNCparam(r, categoryNameFuncParam, "")
	cd.CategoryIcon = "e909"
	iconCodePoint, colorHEX := sqlite.GetCategoryIcon(r.Context(), appdata.DB, categoryName)
	// fmt.Printf("GetCategoryIcon iconCodePoint: %v, colorHEX: %v \n", iconCodePoint, colorHEX)
	if iconCodePoint == "" || colorHEX == "" {
		return appdata.RenderAPIorUI(w, r, isFrontRequest, false, false, http.StatusNotFound, "category not found", "")
	}
	cd.CategoryIcon = iconCodePoint
	cd.CategoryColor = colorHEX
	return appdata.RenderAPIorUI(w, r, isFrontRequest, false, true, http.StatusOK, "category info found", cd)
}

/*
// GET CategorySetup.html
func getCategorySetup() {
	// var CategoryList, IconCodePointList, ColorHEXList []string
	// CategoryList, IconCodePointList, ColorHEXList = sqlite.GetCategoryList(ctx, db)
	// c.HTML(http.StatusOK, "2.2.categorySetup.html", gin.H{
	//     "CategoryList": CategoryList,
	//     "IconCodePointList": IconCodePointList,
	//     "ColorHEXList": ColorHEXList,
	// })
}

// POST CategorySetup.html
func postCategorySetup() {
	var returnedString string
	returnedString = "empty"
	// c.String(200, returnedString)
}
*/
/*
```powershell
curl -X GET -H "Content-Type: application/json" --include --location "http://localhost:8083/api/param/setup"
```
*/
