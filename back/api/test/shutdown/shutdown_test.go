package apishutdowntest

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

// TestMaintenance covers the read-only middleware + toggle endpoint.
// MUST run before TestUser: TestUser calls /api/shutdown which sets ShutDownFlag
// and closes the DB, after which no further requests succeed. Go test runs
// functions in source order within a file, so as long as this stays above
// TestUser it works correctly.
func TestMaintenance(t *testing.T) {
	s := routes.CreateNewServer()
	s.MountBackHandlers()

	resetData()
	// guarantee flag state is clean across runs and at end of test
	appdata.ReadOnlyFlag = false
	t.Cleanup(func() {
		appdata.ReadOnlyFlag = false
		appdata.ShutDownFlag = false
	})

	var req *http.Request
	var response *httptest.ResponseRecorder
	var err error

	// 1. Create admin user (email matches ADMIN_EMAIL env var → IsAdmin = true)
	req, _ = http.NewRequest("POST", "/api/user/create", strings.NewReader(`{
			"email": "test@test.test",
			"password": "test"
		}`))
	response = executeRequest(req, s)
	require.Equal(t, http.StatusCreated, response.Code, "user create should succeed")

	// 2. Login
	req, _ = http.NewRequest("POST", "/api/user/login", strings.NewReader(`{
			"email": "test@test.test",
			"password": "test"
		}`))
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "login should succeed")

	// 3. Force a known sessionID for the admin user
	adminSession := generateFakeSessionID("1")
	_, err = appdata.DB.Exec(`UPDATE user SET sessionID = ? WHERE gofiID = 1;`, adminSession)
	require.NoError(t, err)

	// 4. ReadOnlyFlag defaults to false
	require.False(t, appdata.ReadOnlyFlag, "flag should default to false")

	// 5. Bad state value → 400, flag unchanged
	req, _ = http.NewRequest("GET", "/api/maintenance/readonly/foo", nil)
	req.Header.Set("sessionID", adminSession)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusBadRequest, response.Code, "bad state should be 400")
	require.False(t, appdata.ReadOnlyFlag, "flag should not change on bad input")

	// 6. Enable read-only
	req, _ = http.NewRequest("GET", "/api/maintenance/readonly/on", nil)
	req.Header.Set("sessionID", adminSession)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "enable read-only should succeed")
	require.True(t, appdata.ReadOnlyFlag, "flag should be true after enable")

	// 7. Writes are blocked while read-only is on
	req, _ = http.NewRequest("POST", "/api/user/create", strings.NewReader(`{
			"email": "blocked@test.test",
			"password": "x"
		}`))
	response = executeRequest(req, s)
	require.Equal(t, http.StatusServiceUnavailable, response.Code, "POST should be blocked with 503")

	// 8. Reads still work in read-only mode
	req, _ = http.NewRequest("GET", "/api/isauthenticated", nil)
	req.Header.Set("sessionID", adminSession)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "GET should work in read-only")

	// 9. The toggle endpoint stays reachable (it's a GET, naturally allowed)
	req, _ = http.NewRequest("GET", "/api/maintenance/readonly/off", nil)
	req.Header.Set("sessionID", adminSession)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "toggle off should work while flag is on")
	require.False(t, appdata.ReadOnlyFlag, "flag should be false after disable")

	// 10. Writes work again after disable
	req, _ = http.NewRequest("POST", "/api/user/create", strings.NewReader(`{
			"email": "after@test.test",
			"password": "x"
		}`))
	response = executeRequest(req, s)
	require.Equal(t, http.StatusCreated, response.Code, "POST should work after disable")

	// 11. Non-admin user can't toggle (gets 403).
	// Must log in first: idleTimeout/absoluteTimeout are only set by CheckUserLogin,
	// not by user creation. Without a real login, GetGofiID rejects the session as
	// "absoluteTimeout" before the admin check is even reached.
	req, _ = http.NewRequest("POST", "/api/user/login", strings.NewReader(`{
			"email": "after@test.test",
			"password": "x"
		}`))
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "non-admin login should succeed")
	nonAdminSession := generateFakeSessionID("2")
	_, err = appdata.DB.Exec(`UPDATE user SET sessionID = ? WHERE email = 'after@test.test';`, nonAdminSession)
	require.NoError(t, err)
	req, _ = http.NewRequest("GET", "/api/maintenance/readonly/on", nil)
	req.Header.Set("sessionID", nonAdminSession)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusForbidden, response.Code, "non-admin should get 403")
	require.False(t, appdata.ReadOnlyFlag, "flag should not change for non-admin")

	// 12. Unauthenticated request gets 401 (handled by AuthenticatedUserOnly middleware)
	req, _ = http.NewRequest("GET", "/api/maintenance/readonly/on", nil)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusUnauthorized, response.Code, "no session should get 401")

	// 13. ShutDownFlag still takes precedence over ReadOnlyFlag (full block on shutdown)
	appdata.ReadOnlyFlag = true
	appdata.ShutDownFlag = true
	req, _ = http.NewRequest("GET", "/api/isauthenticated", nil)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusInternalServerError, response.Code, "shutdown should block GETs too")
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

	// 4. SHUTDOWN
	req, _ = http.NewRequest("GET", "/api/shutdown", nil)
	req.Header.Set("sessionID", fsone)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// fmt.Printf("response: %#v\n", response.Body.String())
	// require.Equal(t, 1, 0, "force fail")
	require.WithinDuration(t, time.Now(), testStartTime, 5*time.Second)
}
