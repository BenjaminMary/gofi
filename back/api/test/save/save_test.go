package apiusertest

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

	fmt.Println("cleaning param table")
	_, err = appdata.DB.Exec(`
		DELETE FROM backupSave;
		DELETE FROM SQLITE_SEQUENCE WHERE name='backupSave';
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

	// 7. POST save
	req, _ = http.NewRequest("POST", "/api/save", strings.NewReader(`{
		"date": "2021-06-05",
		"extID": "edfgEd52qedEDQd",
		"extFileName": "2021-06-05-test.db",
		"checkpoint": "1"
	}`))
	req.Header.Set("sessionID", fsone)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusCreated, response.Code, "should be equal")

	// 8. GET save
	req, _ = http.NewRequest("GET", "/api/save", nil)
	req.Header.Set("sessionID", fsone)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")
	require.Equal(t,
		"{\"isValidResponse\":true,\"httpStatus\":200,\"info\":\"list of backup\",\"jsonContent\":[{\"id\":1,\"date\":\"2021-06-05\",\"extID\":\"edfgEd52qedEDQd\",\"extFileName\":\"2021-06-05-test.db\",\"checkpoint\":\"1\",\"tested\":\"0\"}]}\n",
		response.Body.String(), "should be equal")

	// 9. DELETE save
	req, _ = http.NewRequest("DELETE", "/api/save/delete/1", nil)
	req.Header.Set("sessionID", fsone)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// 10. POST save
	req, _ = http.NewRequest("POST", "/api/save", strings.NewReader(`{
		"date": "2021-07-01",
		"extID": "driveID1",
		"extFileName": "save1.db",
		"checkpoint": "1"
	}`))
	req.Header.Set("sessionID", fsone)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusCreated, response.Code, "should be equal")

	// 11. POST save
	req, _ = http.NewRequest("POST", "/api/save", strings.NewReader(`{
		"date": "2021-07-02",
		"extID": "driveID2",
		"extFileName": "save2.db",
		"checkpoint": "1"
	}`))
	req.Header.Set("sessionID", fsone)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusCreated, response.Code, "should be equal")

	// 12. POST save
	req, _ = http.NewRequest("POST", "/api/save", strings.NewReader(`{
		"date": "2021-07-03",
		"extID": "driveID3",
		"extFileName": "save3.db",
		"checkpoint": "1"
	}`))
	req.Header.Set("sessionID", fsone)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusCreated, response.Code, "should be equal")

	// 13. POST save
	req, _ = http.NewRequest("POST", "/api/save", strings.NewReader(`{
		"date": "2021-07-04",
		"extID": "driveID4",
		"extFileName": "save4.db",
		"checkpoint": "1"
	}`))
	req.Header.Set("sessionID", fsone)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusCreated, response.Code, "should be equal")

	// 14. GET save
	req, _ = http.NewRequest("GET", "/api/save", nil)
	req.Header.Set("sessionID", fsone)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")
	require.Equal(t,
		"{\"isValidResponse\":true,\"httpStatus\":200,\"info\":\"list of backup\",\"jsonContent\":[{\"id\":5,\"date\":\"2021-07-04\",\"extID\":\"driveID4\",\"extFileName\":\"save4.db\",\"checkpoint\":\"1\",\"tested\":\"0\"},{\"id\":4,\"date\":\"2021-07-03\",\"extID\":\"driveID3\",\"extFileName\":\"save3.db\",\"checkpoint\":\"1\",\"tested\":\"0\"},{\"id\":3,\"date\":\"2021-07-02\",\"extID\":\"driveID2\",\"extFileName\":\"save2.db\",\"checkpoint\":\"1\",\"tested\":\"0\"},{\"id\":2,\"date\":\"2021-07-01\",\"extID\":\"driveID1\",\"extFileName\":\"save1.db\",\"checkpoint\":\"1\",\"tested\":\"0\"}]}\n",
		response.Body.String(), "should be equal")

	// 15. DELETE save
	req, _ = http.NewRequest("DELETE", "/api/save/keep/2", nil)
	req.Header.Set("sessionID", fsone)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")
	require.Equal(t,
		"{\"isValidResponse\":true,\"httpStatus\":200,\"info\":\"backup deleted\",\"jsonContent\":[\"\",\"driveID1\",\"driveID2\"]}\n",
		response.Body.String(), "should be equal")

	// 16. GET save
	req, _ = http.NewRequest("GET", "/api/save", nil)
	req.Header.Set("sessionID", fsone)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")
	require.Equal(t,
		"{\"isValidResponse\":true,\"httpStatus\":200,\"info\":\"list of backup\",\"jsonContent\":[{\"id\":5,\"date\":\"2021-07-04\",\"extID\":\"driveID4\",\"extFileName\":\"save4.db\",\"checkpoint\":\"1\",\"tested\":\"0\"},{\"id\":4,\"date\":\"2021-07-03\",\"extID\":\"driveID3\",\"extFileName\":\"save3.db\",\"checkpoint\":\"1\",\"tested\":\"0\"}]}\n",
		response.Body.String(), "should be equal")

	// fmt.Printf("response: %#v\n", response.Body.String())
	// require.Equal(t, 1, 0, "force fail")
	require.WithinDuration(t, time.Now(), testStartTime, 1*time.Second)
}
