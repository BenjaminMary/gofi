package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func getURLorFUNCparam(r *http.Request, funcParam string, urlParamCode string) string {
	urlParam := chi.URLParam(r, urlParamCode)
	if urlParam == "" && funcParam == "" {
		return ""
	}
	returnStr := funcParam
	if returnStr == "" {
		returnStr = urlParam
	}
	return returnStr
}
