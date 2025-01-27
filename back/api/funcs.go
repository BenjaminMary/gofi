package api

import (
	"net/http"
	"strings"

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

func notificationPost(url string, text string, title string, tags string) {
	req, _ := http.NewRequest("POST", url,
		strings.NewReader(text))
	req.Header.Set("Title", title)
	req.Header.Set("Tags", tags)
	http.DefaultClient.Do(req)
}