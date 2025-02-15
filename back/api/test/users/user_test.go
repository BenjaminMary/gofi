package apiusertest

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"os"

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

	// 2. CREATE a second time the same Test User
	req, _ = http.NewRequest("POST", "/api/user/create", strings.NewReader(`{
		"email": "test@test.test", 
		"password": "test"
	}`))
	response = executeRequest(req, s)
	require.Equal(t, http.StatusInternalServerError, response.Code, "should be equal")

	// 3. LOGIN Test User
	req, _ = http.NewRequest("POST", "/api/user/login", strings.NewReader(`{
		"email": "test@test.test", 
		"password": "test"
	}`))
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// 4. Force a specific sessionID
	fsone := generateFakeSessionID("1")
	_, err = appdata.DB.Exec(`UPDATE user SET sessionID = ? WHERE gofiID = 1;`, fsone)
	if err != nil {
		panic(err)
	}

	// 5. DELETE the Test User
	req, _ = http.NewRequest("DELETE", "/api/user/1/delete", nil)
	req.Header.Set("sessionID", fsone)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusForbidden, response.Code, "should be equal")

	// 6. CREATE with bad email JSON field
	req, _ = http.NewRequest("POST", "/api/user/create", strings.NewReader(`{
		"BADemail": "testb@test.test", 
		"password": "testb"
	}`))
	response = executeRequest(req, s)
	require.Equal(t, http.StatusBadRequest, response.Code, "should be equal")

	// 7. CREATE with bad password JSON field
	req, _ = http.NewRequest("POST", "/api/user/create", strings.NewReader(`{
		"email": "testb@test.test", 
		"BADpassword": "testb"
	}`))
	response = executeRequest(req, s)
	require.Equal(t, http.StatusBadRequest, response.Code, "should be equal")

	// 8. CREATE with empty email
	req, _ = http.NewRequest("POST", "/api/user/create", strings.NewReader(`{
		"email": "", 
		"password": "testb"
	}`))
	response = executeRequest(req, s)
	require.Equal(t, http.StatusBadRequest, response.Code, "should be equal")

	// 9. CREATE with empty password
	req, _ = http.NewRequest("POST", "/api/user/create", strings.NewReader(`{
		"email": "testb@test.test", 
		"password": ""
	}`))
	response = executeRequest(req, s)
	require.Equal(t, http.StatusBadRequest, response.Code, "should be equal")

	// 10. CREATE a 2nd user
	req, _ = http.NewRequest("POST", "/api/user/create", strings.NewReader(`{
		"email": "testb@test.test", 
		"password": "testb"
	}`))
	response = executeRequest(req, s)
	require.Equal(t, http.StatusCreated, response.Code, "should be equal")

	// 11. LOGIN Test User
	req, _ = http.NewRequest("POST", "/api/user/login", strings.NewReader(`{
		"email": "testb@test.test", 
		"password": "testb"
	}`))
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// 12. Force a specific sessionID
	fstwo := generateFakeSessionID("2")
	_, err = appdata.DB.Exec(`UPDATE user SET sessionID = ? WHERE gofiID = 2;`, fstwo)
	if err != nil {
		panic(err)
	}

	// 13. SHUTDOWN
	req, _ = http.NewRequest("GET", "/api/shutdown", nil)
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusForbidden, response.Code, "should be equal")

	// 14. DELETE the first user with the second
	req, _ = http.NewRequest("DELETE", "/api/user/1/delete", nil)
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusForbidden, response.Code, "should be equal")

	// 15. LOGIN Test User
	req, _ = http.NewRequest("POST", "/api/user/login", strings.NewReader(`{
		"BADemail": "test@test.test", 
		"password": "test"
	}`))
	response = executeRequest(req, s)
	require.Equal(t, http.StatusBadRequest, response.Code, "should be equal")

	// 16. LOGIN Test User
	req, _ = http.NewRequest("POST", "/api/user/login", strings.NewReader(`{
		"email": "test@test.test", 
		"BADpassword": "test"
	}`))
	response = executeRequest(req, s)
	require.Equal(t, http.StatusBadRequest, response.Code, "should be equal")

	// 17. LOGIN Test User
	req, _ = http.NewRequest("POST", "/api/user/login", strings.NewReader(`{
		"email": "test@test.test", 
		"password": ""
	}`))
	response = executeRequest(req, s)
	require.Equal(t, http.StatusBadRequest, response.Code, "should be equal")

	// 18. LOGIN Test User
	req, _ = http.NewRequest("POST", "/api/user/login", strings.NewReader(`{
		"email": "test@test.test", 
		"password": "wrong"
	}`))
	response = executeRequest(req, s)
	require.Equal(t, http.StatusUnauthorized, response.Code, "should be equal")

	// 19. LOGOUT Test User
	req, _ = http.NewRequest("GET", "/api/user/logout", nil)
	req.Header.Set("sessionID", "wrong one")
	response = executeRequest(req, s)
	require.Equal(t, http.StatusUnauthorized, response.Code, "should be equal")

	// 20. LOGOUT Test User
	req, _ = http.NewRequest("GET", "/api/user/logout", nil)
	req.Header.Set("sessionID", fsone)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// 21. LOGIN Test User
	req, _ = http.NewRequest("POST", "/api/user/login", strings.NewReader(`{
		"email": "test@test.test", 
		"password": "test"
	}`))
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// 22. LOGOUT Test User
	req, _ = http.NewRequest("GET", "/api/user/logout", nil)
	req.Header.Set("sessionID", fsone)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusUnauthorized, response.Code, "should be equal")

	// 23. Force a specific sessionID
	// fsone := generateFakeSessionID("1")
	_, err = appdata.DB.Exec(`UPDATE user SET sessionID = ? WHERE gofiID = 1;`, fsone)
	if err != nil {
		panic(err)
	}

	// 24. REFRESH session Test User
	req, _ = http.NewRequest("GET", "/api/user/2/refresh", nil)
	req.Header.Set("sessionID", fsone)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusUnauthorized, response.Code, "should be equal")

	// 25. REFRESH session Test User
	req, _ = http.NewRequest("GET", "/api/user/1/refresh", nil)
	req.Header.Set("sessionID", "wrong one")
	response = executeRequest(req, s)
	require.Equal(t, http.StatusUnauthorized, response.Code, "should be equal")

	// 26. REFRESH session Test User
	req, _ = http.NewRequest("GET", "/api/user/a/refresh", nil)
	req.Header.Set("sessionID", fsone)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusBadRequest, response.Code, "should be equal")

	// 27. REFRESH session Test User
	req, _ = http.NewRequest("GET", "/api/user/1/refresh", nil)
	req.Header.Set("sessionID", fsone)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// 28. LOGOUT Test User
	req, _ = http.NewRequest("GET", "/api/user/logout", nil)
	req.Header.Set("sessionID", fsone)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusUnauthorized, response.Code, "should be equal")

	// 29. Force a specific sessionID
	// fsone := generateFakeSessionID("1")
	_, err = appdata.DB.Exec(`UPDATE user SET sessionID = ? WHERE gofiID = 1;`, fsone)
	if err != nil {
		panic(err)
	}

	// 30. CHECK IsAdmin first user
	req, _ = http.NewRequest("GET", "/api/isadmin", nil)
	req.Header.Set("sessionID", fsone)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// 31. CHECK IsAuthenticated first user
	req, _ = http.NewRequest("GET", "/api/isauthenticated", nil)
	req.Header.Set("sessionID", fsone)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// 32. CHECK IsAdmin second user
	req, _ = http.NewRequest("GET", "/api/isadmin", nil)
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusUnauthorized, response.Code, "should be equal")

	// 33. CHECK IsAuthenticated second user
	req, _ = http.NewRequest("GET", "/api/isauthenticated", nil)
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// 34. CHECK IsAuthenticated wrong user
	req, _ = http.NewRequest("GET", "/api/isauthenticated", nil)
	req.Header.Set("sessionID", "wrong one")
	response = executeRequest(req, s)
	require.Equal(t, http.StatusUnauthorized, response.Code, "should be equal")

	// 35. Update to force idle timeout
	_, err = appdata.DB.Exec(`UPDATE user SET idleTimeout='1999-12-31T00:01:02Z' WHERE gofiID = 2;`)
	if err != nil {
		panic(err)
	}

	// 36. CHECK IsAuthenticated second user
	req, _ = http.NewRequest("GET", "/api/isauthenticated", nil)
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal and update the SessionID")

	// 37. CHECK IsAuthenticated second user
	req, _ = http.NewRequest("GET", "/api/isauthenticated", nil)
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusUnauthorized, response.Code, "should be equal")

	// 39. Update to force absoluteTimeout timeout
	_, err = appdata.DB.Exec(`UPDATE user SET absoluteTimeout='1999-12-31T00:01:02Z' WHERE gofiID = 2;`)
	if err != nil {
		panic(err)
	}

	// 40. Force a specific sessionID
	// fstwo := generateFakeSessionID("2")
	_, err = appdata.DB.Exec(`UPDATE user SET sessionID = ? WHERE gofiID = 2;`, fstwo)
	if err != nil {
		panic(err)
	}

	// 41. CHECK IsAuthenticated second user
	req, _ = http.NewRequest("GET", "/api/isauthenticated", nil)
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusUnauthorized, response.Code, "should be equal")
	// require.Equal(t,
	// 	"{\"isValidResponse\":false,\"httpStatus\":401,\"info\":\"authenticated ko\",\"jsonContent\":\"\"}\n",
	// 	response.Body.String(), "should be equal")

	// 42. DELETE the second user
	req, _ = http.NewRequest("DELETE", "/api/user/2/delete", nil)
	req.Header.Set("sessionID", fsone)
	response = executeRequest(req, s)
	fmt.Printf("response: %#v\n", response.Body.String())
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// 43. SHUTDOWN
	req, _ = http.NewRequest("GET", "/api/shutdown", nil)
	req.Header.Set("sessionID", "wrong one")
	response = executeRequest(req, s)
	require.Equal(t, http.StatusUnauthorized, response.Code, "should be equal")

	// 44. GET DB PATH
	req, _ = http.NewRequest("GET", "/api/dbpath", nil)
	req.Header.Set("sessionID", "wrong one")
	response = executeRequest(req, s)
	require.Equal(t, http.StatusUnauthorized, response.Code, "should be equal")

	// 45. GET DB PATH
	req, _ = http.NewRequest("GET", "/api/dbpath", nil)
	req.Header.Set("sessionID", fsone)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// 46. CREATE a 3rd user
	req, _ = http.NewRequest("POST", "/api/user/create", strings.NewReader(`{
		"email": "testc@test.test", 
		"password": "testc"
	}`))
	response = executeRequest(req, s)
	require.Equal(t, http.StatusCreated, response.Code, "should be equal")

	// 47. LOGIN Test User
	req, _ = http.NewRequest("POST", "/api/user/login", strings.NewReader(`{
		"email": "testc@test.test", 
		"password": "testc"
	}`))
	req.Header.Set(os.Getenv("HEADER_IP"), "1.2.3.4")
	req.Header.Set("User-Agent", "test")
	req.Header.Set("Accept-Language", "fr-en")
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// 48. Force a specific sessionID
	fsthree := generateFakeSessionID("3")
	_, err = appdata.DB.Exec(`UPDATE user SET sessionID = ? WHERE gofiID = 3;`, fsthree)
	if err != nil {
		panic(err)
	}

	// 49. LOGOUT Test User without headers
	req, _ = http.NewRequest("GET", "/api/user/logout", nil)
	req.Header.Set("sessionID", fsthree)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusUnauthorized, response.Code, "should be equal")

	// 49. LOGOUT Test User with headers but session reseted previously
	req, _ = http.NewRequest("GET", "/api/user/logout", nil)
	req.Header.Set("sessionID", fsthree)
	req.Header.Set(os.Getenv("HEADER_IP"), "1.2.3.4")
	req.Header.Set("User-Agent", "test")
	req.Header.Set("Accept-Language", "fr-en")
	response = executeRequest(req, s)
	require.Equal(t, http.StatusUnauthorized, response.Code, "should be equal")

	// 50. LOGIN Test User
	req, _ = http.NewRequest("POST", "/api/user/login", strings.NewReader(`{
		"email": "testc@test.test", 
		"password": "testc"
	}`))
	req.Header.Set(os.Getenv("HEADER_IP"), "1.2.3.4.b")
	req.Header.Set("User-Agent", "testc")
	req.Header.Set("Accept-Language", "fr-en-b")
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// 51. Force a specific sessionID
	_, err = appdata.DB.Exec(`UPDATE user SET sessionID = ? WHERE gofiID = 3;`, fsthree)
	if err != nil {
		panic(err)
	}

	// 52. LOGOUT Test User err IP StatusOK because param set to "0"
	req, _ = http.NewRequest("GET", "/api/user/logout", nil)
	req.Header.Set("sessionID", fsthree)
	req.Header.Set(os.Getenv("HEADER_IP"), "-")
	req.Header.Set("User-Agent", "testc")
	req.Header.Set("Accept-Language", "fr-en-b")
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// 52a. LOGIN Test User
	req, _ = http.NewRequest("POST", "/api/user/login", strings.NewReader(`{
		"email": "testc@test.test", 
		"password": "testc"
	}`))
	req.Header.Set(os.Getenv("HEADER_IP"), "1.2.3.4.b")
	req.Header.Set("User-Agent", "testc")
	req.Header.Set("Accept-Language", "fr-en-b")
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// 52b. Force a specific absoluteTimeout + sessionID
	_, err = appdata.DB.Exec(`UPDATE user SET sessionID = ? WHERE gofiID = 3;`, fsthree)
	if err != nil {
		panic(err)
	}

	// 52c. UPDATE PARAM forceNewLoginOnIPchange to "1"
	req, _ = http.NewRequest("POST", "/api/param/force-new-login-on-ip-change", strings.NewReader(`{
		"ParamJSONstringData": "1"
	}`))
	req.Header.Set("sessionID", fsthree)
	req.Header.Set(os.Getenv("HEADER_IP"), "-")
	req.Header.Set("User-Agent", "testc")
	req.Header.Set("Accept-Language", "fr-en-b")
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// 52d. LOGOUT Test User err IP StatusUnauthorized because param set to "1"
	req, _ = http.NewRequest("GET", "/api/user/logout", nil)
	req.Header.Set("sessionID", fsthree)
	req.Header.Set(os.Getenv("HEADER_IP"), "xyz")
	req.Header.Set("User-Agent", "testc")
	req.Header.Set("Accept-Language", "fr-en-b")
	response = executeRequest(req, s)
	require.Equal(t, http.StatusUnauthorized, response.Code, "should be equal")

	// 53. Force a specific absoluteTimeout + sessionID
	_, err = appdata.DB.Exec(`UPDATE user SET absoluteTimeout = '3999-12-31T00:01:01Z', sessionID = ? WHERE gofiID = 3;`, fsthree)
	if err != nil {
		panic(err)
	}

	// 54. LOGOUT Test User err userAgent
	req, _ = http.NewRequest("GET", "/api/user/logout", nil)
	req.Header.Set("sessionID", fsthree)
	req.Header.Set(os.Getenv("HEADER_IP"), "1.2.3.4.b")
	req.Header.Set("User-Agent", "-")
	req.Header.Set("Accept-Language", "fr-en-b")
	response = executeRequest(req, s)
	require.Equal(t, http.StatusUnauthorized, response.Code, "should be equal")

	// 55. Force a specific absoluteTimeout + sessionID
	_, err = appdata.DB.Exec(`UPDATE user SET absoluteTimeout = '3999-12-31T00:01:01Z', sessionID = ? WHERE gofiID = 3;`, fsthree)
	if err != nil {
		panic(err)
	}

	// 56. LOGOUT Test User err acceptLanguage
	req, _ = http.NewRequest("GET", "/api/user/logout", nil)
	req.Header.Set("sessionID", fsthree)
	req.Header.Set(os.Getenv("HEADER_IP"), "1.2.3.4.b")
	req.Header.Set("User-Agent", "testc")
	req.Header.Set("Accept-Language", "-")
	response = executeRequest(req, s)
	require.Equal(t, http.StatusUnauthorized, response.Code, "should be equal")

	// 57. Force a specific absoluteTimeout + sessionID
	_, err = appdata.DB.Exec(`UPDATE user SET absoluteTimeout = '3999-12-31T00:01:01Z', sessionID = ? WHERE gofiID = 3;`, fsthree)
	if err != nil {
		panic(err)
	}

	// 58. LOGOUT Test User ok
	req, _ = http.NewRequest("GET", "/api/user/logout", nil)
	req.Header.Set("sessionID", fsthree)
	req.Header.Set(os.Getenv("HEADER_IP"), "1.2.3.4.b")
	req.Header.Set("User-Agent", "testc")
	req.Header.Set("Accept-Language", "fr-en-b")
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// fmt.Printf("response: %#v\n", response.Body.String())
	// require.Equal(t, 1, 0, "force fail")
	require.WithinDuration(t, time.Now(), testStartTime, 1*time.Second)
}
