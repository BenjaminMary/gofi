package apiparamtest

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gofi/gofi/back/routes"
	"gofi/gofi/data/appdata"

	"github.com/stretchr/testify/require"
)

// executeRequest, creates a new ResponseRecorder
// then executes the request by calling ServeHTTP in the router
// after which the handler writes the response to the response recorder
// which we can then inspect.
func executeRequest(req *http.Request, s *routes.Server) *httptest.ResponseRecorder {
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	s.Router.ServeHTTP(rr, req)
	return rr
}

func generateFakeSessionID(str string) string {
	var sb strings.Builder
	for i := 0; i < appdata.CookieLength; i++ {
		sb.WriteString(str)
	}
	return sb.String()
}

func resetData() {
	var err error
	fmt.Println(appdata.DbPath)
	if len(appdata.DbPath) < 4 {
		panic("specify the SQLITE_DB_FILENAME environment variable")
	}
	appdata.DB, err = sql.Open("sqlite", appdata.DbPath)
	if err != nil {
		panic("error opening the SQLite file")
	}
	err = appdata.DB.Ping()
	if err != nil {
		panic("can't ping DB")
	}

	fmt.Println("cleaning user table")
	_, err = appdata.DB.Exec(`
		DELETE FROM user;
		DELETE FROM SQLITE_SEQUENCE WHERE name='user';
		`,
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("cleaning param table")
	_, err = appdata.DB.Exec(`
		DELETE FROM param;
		DELETE FROM SQLITE_SEQUENCE WHERE name='param';
		`,
	)
	if err != nil {
		panic(err)
	}
}

func TestUser(t *testing.T) {
	testStartTime := time.Now()
	s := routes.CreateNewServer()
	s.MountBackHandlers()

	resetData()
	var req *http.Request
	var response *httptest.ResponseRecorder
	var err error

	// 1. CREATE a New Test User (ADMIN)
	req, _ = http.NewRequest("POST", "/api/user/create", strings.NewReader(`{
			"email": "test@test.test", 
			"password": "test"
		}`))
	// Execute Request
	response = executeRequest(req, s)
	// Check the response code
	require.Equal(t, http.StatusCreated, response.Code, "should be equal")

	// 2. LOGIN Test User
	req, _ = http.NewRequest("POST", "/api/user/login", strings.NewReader(`{
		"email": "test@test.test", 
		"password": "test"
	}`))
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// 3. Force a specific sessionID
	fsone := generateFakeSessionID("1")
	_, err = appdata.DB.Exec(`UPDATE user SET sessionID = ? WHERE gofiID = 1;`, fsone)
	if err != nil {
		panic(err)
	}

	// 4. CREATE a 2nd user
	req, _ = http.NewRequest("POST", "/api/user/create", strings.NewReader(`{
		"email": "testb@test.test", 
		"password": "testb"
	}`))
	response = executeRequest(req, s)
	require.Equal(t, http.StatusCreated, response.Code, "should be equal")

	// 5. LOGIN Test User
	req, _ = http.NewRequest("POST", "/api/user/login", strings.NewReader(`{
		"email": "testb@test.test", 
		"password": "testb"
	}`))
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// 6. Force a specific sessionID
	fstwo := generateFakeSessionID("2")
	_, err = appdata.DB.Exec(`UPDATE user SET sessionID = ? WHERE gofiID = 2;`, fstwo)
	if err != nil {
		panic(err)
	}

	// 7. GET PARAM
	req, _ = http.NewRequest("GET", "/api/param", nil)
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// 8. POST PARAM ACCOUNT
	req, _ = http.NewRequest("POST", "/api/param/account", strings.NewReader(`{
		"ParamJSONstringData": "acc1,acc2,acc3"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// 9. POST PARAM CATEGORY
	req, _ = http.NewRequest("POST", "/api/param/category", strings.NewReader(`{
		"ParamJSONstringData": "cat1,cat2,cat3"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// 10. POST PARAM ACCOUNT
	req, _ = http.NewRequest("POST", "/api/param/category-rendering", strings.NewReader(`{
		"ParamJSONstringData": "names"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	fmt.Println("-----------------11. GET PARAM")
	// 11. GET PARAM
	req, _ = http.NewRequest("GET", "/api/param", nil)
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")
	// require.Equal(t,
	// 	"{\"isValidResponse\":true,\"httpStatus\":200,\"info\":\"user params retrieved\",\"jsonContent\":{\"GofiID\":2,\"AccountListSingleString\":\"acc1,acc2,acc3\",\"AccountList\":[\"acc1\",\"acc2\",\"acc3\"],\"CategoryListSingleString\":\"cat1,cat2,cat3\",\"CategoryList\":[[\"cat1\",\"e90a\",\"#808080\"],[\"cat2\",\"e90a\",\"#808080\"],[\"cat3\",\"e90a\",\"#808080\"]],\"CategoryRendering\":\"names\"}}\n",
	// 	response.Body.String(), "should be equal")

	// 12. GET PARAM
	req, _ = http.NewRequest("GET", "/api/param", nil)
	req.Header.Set("sessionID", "wrong one")
	response = executeRequest(req, s)
	require.Equal(t, http.StatusUnauthorized, response.Code, "should be equal")

	// 13. POST PARAM ACCOUNT
	req, _ = http.NewRequest("POST", "/api/param/account", strings.NewReader(`{
		"ParamJSONstringData": "fail"
	}`))
	req.Header.Set("sessionID", "wrong one")
	response = executeRequest(req, s)
	require.Equal(t, http.StatusUnauthorized, response.Code, "should be equal")

	// 14. PATCH PARAM InUse
	req, _ = http.NewRequest("PATCH", "/api/param/category/in-use", strings.NewReader(`{
		"idstrjson": "6",
		"inusestrjson": "0"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusNotFound, response.Code, "should be equal")

	// 15. PATCH PARAM InUse
	req, _ = http.NewRequest("PATCH", "/api/param/category/in-use", strings.NewReader(`{
		"idstrjson": "0",
		"inusestrjson": "0"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusBadRequest, response.Code, "should be equal")

	// 16. PATCH PARAM InUse
	req, _ = http.NewRequest("PATCH", "/api/param/category/in-use", strings.NewReader(`{
		"idstrjson": "26",
		"inusestrjson": "0"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// 17. PATCH PARAM Order
	req, _ = http.NewRequest("PATCH", "/api/param/category/order", strings.NewReader(`{
		"idStrJson": "28",
		"orderStrJson": "3",
		"directionStrJson": "up"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// 18. PATCH PARAM Order
	req, _ = http.NewRequest("PATCH", "/api/param/category/order", strings.NewReader(`{
		"idStrJson": "28",
		"orderStrJson": "2",
		"directionStrJson": "up"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// 19. PATCH PARAM Order
	req, _ = http.NewRequest("PATCH", "/api/param/category/order", strings.NewReader(`{
		"idStrJson": "26",
		"orderStrJson": "2",
		"directionStrJson": "down"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// 20. GET PARAM
	req, _ = http.NewRequest("GET", "/api/param", nil)
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")
	require.Equal(t,
		"{\"isValidResponse\":true,\"httpStatus\":200,\"info\":\"user params retrieved\",\"jsonContent\":{\"GofiID\":2,\"AccountListSingleString\":\"acc1,acc2,acc3\",\"AccountList\":[\"acc1\",\"acc2\",\"acc3\"],\"CategoryListSingleString\":\"cat1,cat2,cat3\",\"CategoryList\":[[\"cat1\",\"e90a\",\"#808080\"],[\"cat2\",\"e90a\",\"#808080\"],[\"cat3\",\"e90a\",\"#808080\"]],\"CategoryRendering\":\"names\"}}\n",
		response.Body.String(), "should be equal")

	// roll back order changes to be able to play this test 2 times in a row
	// 21. PATCH PARAM Order
	req, _ = http.NewRequest("PATCH", "/api/param/category/order", strings.NewReader(`{
		"idStrJson": "28",
		"orderStrJson": "1",
		"directionStrJson": "down"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// 22. PATCH PARAM Order
	req, _ = http.NewRequest("PATCH", "/api/param/category/order", strings.NewReader(`{
		"idStrJson": "28",
		"orderStrJson": "2",
		"directionStrJson": "down"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// 23. PATCH PARAM Order
	req, _ = http.NewRequest("PATCH", "/api/param/category/order", strings.NewReader(`{
		"idStrJson": "26",
		"orderStrJson": "2",
		"directionStrJson": "up"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// 24. PATCH PARAM Order
	req, _ = http.NewRequest("PATCH", "/api/param/category/order", strings.NewReader(`{
		"idStrJson": "26",
		"orderStrJson": "3",
		"directionStrJson": "down"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusNotFound, response.Code, "should be equal")

	// fmt.Printf("response: %#v\n", response.Body.String())
	// require.Equal(t, 1, 0, "force fail")
	require.WithinDuration(t, time.Now(), testStartTime, 5*time.Second)
}
