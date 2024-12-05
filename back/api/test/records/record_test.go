package apirecordtest

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

	fmt.Println("cleaning financeTracker table")
	_, err = appdata.DB.Exec(`
		DELETE FROM financeTracker;
		DELETE FROM SQLITE_SEQUENCE WHERE name='financeTracker';
		`,
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("cleaning recurrentRecord table")
	_, err = appdata.DB.Exec(`
		DELETE FROM recurrentRecord;
		DELETE FROM SQLITE_SEQUENCE WHERE name='recurrentRecord';
		`,
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("cleaning lenderBorrower table")
	_, err = appdata.DB.Exec(`
		DELETE FROM lenderBorrower;
		DELETE FROM SQLITE_SEQUENCE WHERE name='lenderBorrower';
		`,
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("cleaning specificRecordsByMode table")
	_, err = appdata.DB.Exec(`
		DELETE FROM specificRecordsByMode;
		DELETE FROM SQLITE_SEQUENCE WHERE name='specificRecordsByMode';
		`,
	)
	if err != nil {
		panic(err)
	}
}

func TestRecord(t *testing.T) {
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

	// 7. POST RECORD INSERT
	req, _ = http.NewRequest("POST", "/api/record/insert", strings.NewReader(`{
		"date": "2001-01-01",
		"compte": "CB",
		"designation": "test",
		"gain-expense": "expense",
		"prix": "1.00",
		"categorie": "Courses"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusBadRequest, response.Code, "should be equal")

	// 8. POST RECORD INSERT
	req, _ = http.NewRequest("POST", "/api/record/insert", strings.NewReader(`{
		"Date": "2011-11-11",
		"Account": "CB",
		"Product": "test",
		"PriceDirection": "expense",
		"FormPriceStr2Decimals": "1.00",
		"Category": "Courses"
	}`))
	req.Header.Set("sessionID", "wrong one")
	response = executeRequest(req, s)
	require.Equal(t, http.StatusUnauthorized, response.Code, "should be equal")

	// 9. POST RECORD INSERT
	req, _ = http.NewRequest("POST", "/api/record/insert", strings.NewReader(`{
		"Date": "2011-11-11",
		"Account": "CB",
		"Product": "testb",
		"PriceDirection": "expense",
		"FormPriceStr2Decimals": "1.00",
		"Category": "Courses"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusCreated, response.Code, "should be equal")

	// 10. GET RECORD
	req, _ = http.NewRequest("GET", "/api/record/id-desc-5", nil)
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")
	require.Equal(t,
		"{\"isValidResponse\":true,\"httpStatus\":200,\"info\":\"record list retrieved\",\"jsonContent\":[{\"ID\":1,\"GofiID\":2,\"Date\":\"2011-11-11\",\"Account\":\"CB\",\"Product\":\"testb\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"-1.00\",\"PriceIntx100\":-100,\"Category\":\"Courses\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":false,\"DateChecked\":\"9999-12-31\",\"Mode\":0,\"Exported\":false}]}\n",
		response.Body.String(), "should be equal")

	// 11. POST RECORD INSERT
	req, _ = http.NewRequest("POST", "/api/record/insert", strings.NewReader(`{
		"Date": "2011-11-11",
		"Account": "CB",
		"Product": "test",
		"PriceDirection": "expense",
		"FormPriceStr2Decimals": "1.00",
		"Category": "Courses"
	}`))
	req.Header.Set("sessionID", fsone)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusCreated, response.Code, "should be equal")

	// 12. GET RECORD
	req, _ = http.NewRequest("GET", "/api/record/id-desc-5", nil)
	req.Header.Set("sessionID", fsone)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")
	require.Equal(t,
		"{\"isValidResponse\":true,\"httpStatus\":200,\"info\":\"record list retrieved\",\"jsonContent\":[{\"ID\":2,\"GofiID\":1,\"Date\":\"2011-11-11\",\"Account\":\"CB\",\"Product\":\"test\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"-1.00\",\"PriceIntx100\":-100,\"Category\":\"Courses\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":false,\"DateChecked\":\"9999-12-31\",\"Mode\":0,\"Exported\":false}]}\n",
		response.Body.String(), "should be equal")

	// 13. POST RECORD INSERT
	req, _ = http.NewRequest("POST", "/api/record/insert", strings.NewReader(`{
		"Date": "2011-13-11",
		"Account": "CB",
		"Product": "testberr",
		"PriceDirection": "expense",
		"FormPriceStr2Decimals": "1.00",
		"Category": "Courses"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusBadRequest, response.Code, "should be equal")
	require.Equal(t,
		"{\"isValidResponse\":false,\"httpStatus\":400,\"info\":\"invalid date\",\"jsonContent\":\"\"}\n",
		response.Body.String(), "should be equal")

	// 14. POST RECORD INSERT
	req, _ = http.NewRequest("POST", "/api/record/insert", strings.NewReader(`{
		"Date": "2010-12-31",
		"Account": "CB",
		"Product": "testb",
		"PriceDirection": "expense",
		"FormPriceStr2Decimals": "1.00",
		"Category": "Courses"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusCreated, response.Code, "should be equal")

	// 15. POST RECORD INSERT
	req, _ = http.NewRequest("POST", "/api/record/insert", strings.NewReader(`{
		"Date": "2011-04-31",
		"Account": "CB",
		"Product": "testb err 11-4-31 should fail",
		"PriceDirection": "expense",
		"FormPriceStr2Decimals": "1.00",
		"Category": "Courses"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusBadRequest, response.Code, "should be equal")
	require.Equal(t,
		"{\"isValidResponse\":false,\"httpStatus\":400,\"info\":\"invalid date\",\"jsonContent\":\"\"}\n",
		response.Body.String(), "should be equal")

	// 16. GET RECORD
	req, _ = http.NewRequest("GET", "/api/record/date-asc-1", nil)
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")
	require.Equal(t,
		"{\"isValidResponse\":true,\"httpStatus\":200,\"info\":\"record list retrieved\",\"jsonContent\":[{\"ID\":3,\"GofiID\":2,\"Date\":\"2010-12-31\",\"Account\":\"CB\",\"Product\":\"testb\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"-1.00\",\"PriceIntx100\":-100,\"Category\":\"Courses\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":false,\"DateChecked\":\"9999-12-31\",\"Mode\":0,\"Exported\":false}]}\n",
		response.Body.String(), "should be equal")

	// 17. POST RECORD TRANSFER
	req, _ = http.NewRequest("POST", "/api/record/transfer", strings.NewReader(`{
		"Date": "2012-05-11",
		"AccountFrom": "A",
		"AccountTo": "CB",
		"FormPriceStr2Decimals": "143.70"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusCreated, response.Code, "should be equal")

	// 18. POST RECORD TRANSFER
	req, _ = http.NewRequest("POST", "/api/record/transfer", strings.NewReader(`{
		"Date": "2012-08-19",
		"AccountFrom": "-",
		"AccountTo": "CB",
		"FormPriceStr2Decimals": "117.40"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusBadRequest, response.Code, "should be equal")

	// 19. POST RECORD TRANSFER
	req, _ = http.NewRequest("POST", "/api/record/transfer", strings.NewReader(`{
		"Date": "2012-08-19",
		"AccountFrom": "A",
		"AccountTo": "-",
		"FormPriceStr2Decimals": "117.40"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusBadRequest, response.Code, "should be equal")

	// 20. POST RECORD TRANSFER
	req, _ = http.NewRequest("POST", "/api/record/transfer", strings.NewReader(`{
		"Date": "2012-08-19",
		"AccountFrom": "A",
		"AccountTo": "CB",
		"FormPriceStr2Decimals": "117.40"
	}`))
	req.Header.Set("sessionID", "wrong one")
	response = executeRequest(req, s)
	require.Equal(t, http.StatusUnauthorized, response.Code, "should be equal")

	// 21. GET RECORD
	req, _ = http.NewRequest("GET", "/api/record/id-desc-2", nil)
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")
	require.Equal(t,
		"{\"isValidResponse\":true,\"httpStatus\":200,\"info\":\"record list retrieved\",\"jsonContent\":[{\"ID\":5,\"GofiID\":2,\"Date\":\"2012-05-11\",\"Account\":\"CB\",\"Product\":\"Transfert+\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"143.70\",\"PriceIntx100\":14370,\"Category\":\"Transfert\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":false,\"DateChecked\":\"9999-12-31\",\"Mode\":0,\"Exported\":false},{\"ID\":4,\"GofiID\":2,\"Date\":\"2012-05-11\",\"Account\":\"A\",\"Product\":\"Transfert-\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"-143.70\",\"PriceIntx100\":-14370,\"Category\":\"Transfert\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":false,\"DateChecked\":\"9999-12-31\",\"Mode\":0,\"Exported\":false}]}\n",
		response.Body.String(), "should be equal")

	// 22. POST RECORD RECURRENT
	req, _ = http.NewRequest("POST", "/api/record/recurrent/create", strings.NewReader(`{
		"Date": "2011-05-31",
		"Recurrence": "mensuelle",
		"Account": "A",
		"Category": "Banque",
		"Product": "Revenu",
		"PriceDirection": "gain",
		"FormPriceStr2Decimals": "1183.16"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusCreated, response.Code, "should be equal")

	// 22. POST RECORD RECURRENT
	req, _ = http.NewRequest("POST", "/api/record/recurrent/create", strings.NewReader(`{
		"Date": "2012-01-31",
		"Recurrence": "mensuelle",
		"Account": "A",
		"Category": "Loyer",
		"Product": "Loyer",
		"PriceDirection": "expense",
		"FormPriceStr2Decimals": "641.09"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusCreated, response.Code, "should be equal")

	// 23. GET RECORD RECURRENT
	req, _ = http.NewRequest("GET", "/api/record/recurrent", nil)
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")
	require.Equal(t,
		"{\"isValidResponse\":true,\"httpStatus\":200,\"info\":\"recurrent record selected\",\"jsonContent\":[{\"ID\":1,\"IDstr\":\"\",\"GofiID\":0,\"Date\":\"2011-05-31\",\"Recurrence\":\"mensuelle\",\"Account\":\"A\",\"Product\":\"Revenu\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"1183.16\",\"PriceIntx100\":118316,\"Category\":\"Banque\"},{\"ID\":2,\"IDstr\":\"\",\"GofiID\":0,\"Date\":\"2012-01-31\",\"Recurrence\":\"mensuelle\",\"Account\":\"A\",\"Product\":\"Loyer\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"-641.09\",\"PriceIntx100\":-64109,\"Category\":\"Loyer\"}]}\n",
		response.Body.String(), "should be equal")

	// 24. POST RECORD RECURRENT RecordRecurrentSave
	req, _ = http.NewRequest("POST", "/api/record/recurrent/save", strings.NewReader(`{
		"ID": "1"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusCreated, response.Code, "should be equal")

	// 25. POST RECORD RECURRENT RecordRecurrentSave
	req, _ = http.NewRequest("POST", "/api/record/recurrent/save", strings.NewReader(`{
		"ID": "2"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusCreated, response.Code, "should be equal")

	// 26. POST RECORD RECURRENT RecordRecurrentSave
	req, _ = http.NewRequest("POST", "/api/record/recurrent/save", strings.NewReader(`{
		"ID": "2"
	}`))
	req.Header.Set("sessionID", fsone)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusBadRequest, response.Code, "should be equal")

	// 27. POST RECORD RECURRENT
	req, _ = http.NewRequest("POST", "/api/record/recurrent/create", strings.NewReader(`{
		"Date": "2012-01-31",
		"Recurrence": "hebdomadaire",
		"Account": "CB",
		"Category": "Courses",
		"Product": "ToDelete",
		"PriceDirection": "expense",
		"FormPriceStr2Decimals": "34.53"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusCreated, response.Code, "should be equal")

	// 28. POST RECORD RECURRENT RecordRecurrentSave
	req, _ = http.NewRequest("POST", "/api/record/recurrent/save", strings.NewReader(`{
		"ID": "3"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusCreated, response.Code, "should be equal")

	// 29. POST RECORD RECURRENT RecordRecurrentSave
	req, _ = http.NewRequest("POST", "/api/record/recurrent/save", strings.NewReader(`{
		"ID": "3"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusCreated, response.Code, "should be equal")

	// 30. POST RECORD RECURRENT RecordRecurrentSave
	req, _ = http.NewRequest("POST", "/api/record/recurrent/save", strings.NewReader(`{
		"ID": "3"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusCreated, response.Code, "should be equal")

	// 31. GET RECORD
	req, _ = http.NewRequest("GET", "/api/record/id-desc-5", nil)
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")
	require.Equal(t,
		"{\"isValidResponse\":true,\"httpStatus\":200,\"info\":\"record list retrieved\",\"jsonContent\":[{\"ID\":10,\"GofiID\":2,\"Date\":\"2012-02-14\",\"Account\":\"CB\",\"Product\":\"ToDelete\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"-34.53\",\"PriceIntx100\":-3453,\"Category\":\"Courses\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":false,\"DateChecked\":\"9999-12-31\",\"Mode\":0,\"Exported\":false},{\"ID\":9,\"GofiID\":2,\"Date\":\"2012-02-07\",\"Account\":\"CB\",\"Product\":\"ToDelete\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"-34.53\",\"PriceIntx100\":-3453,\"Category\":\"Courses\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":false,\"DateChecked\":\"9999-12-31\",\"Mode\":0,\"Exported\":false},{\"ID\":8,\"GofiID\":2,\"Date\":\"2012-01-31\",\"Account\":\"CB\",\"Product\":\"ToDelete\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"-34.53\",\"PriceIntx100\":-3453,\"Category\":\"Courses\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":false,\"DateChecked\":\"9999-12-31\",\"Mode\":0,\"Exported\":false},{\"ID\":7,\"GofiID\":2,\"Date\":\"2012-01-31\",\"Account\":\"A\",\"Product\":\"Loyer\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"-641.09\",\"PriceIntx100\":-64109,\"Category\":\"Loyer\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":false,\"DateChecked\":\"9999-12-31\",\"Mode\":0,\"Exported\":false},{\"ID\":6,\"GofiID\":2,\"Date\":\"2011-05-31\",\"Account\":\"A\",\"Product\":\"Revenu\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"1183.16\",\"PriceIntx100\":118316,\"Category\":\"Banque\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":false,\"DateChecked\":\"9999-12-31\",\"Mode\":0,\"Exported\":false}]}\n",
		response.Body.String(), "should be equal")

	// 32. GET RECORD RECURRENT
	req, _ = http.NewRequest("GET", "/api/record/recurrent", nil)
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")
	require.Equal(t,
		"{\"isValidResponse\":true,\"httpStatus\":200,\"info\":\"recurrent record selected\",\"jsonContent\":[{\"ID\":1,\"IDstr\":\"\",\"GofiID\":0,\"Date\":\"2011-06-30\",\"Recurrence\":\"mensuelle\",\"Account\":\"A\",\"Product\":\"Revenu\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"1183.16\",\"PriceIntx100\":118316,\"Category\":\"Banque\"},{\"ID\":3,\"IDstr\":\"\",\"GofiID\":0,\"Date\":\"2012-02-21\",\"Recurrence\":\"hebdomadaire\",\"Account\":\"CB\",\"Product\":\"ToDelete\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"-34.53\",\"PriceIntx100\":-3453,\"Category\":\"Courses\"},{\"ID\":2,\"IDstr\":\"\",\"GofiID\":0,\"Date\":\"2012-02-29\",\"Recurrence\":\"mensuelle\",\"Account\":\"A\",\"Product\":\"Loyer\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"-641.09\",\"PriceIntx100\":-64109,\"Category\":\"Loyer\"}]}\n",
		response.Body.String(), "should be equal")

	// 33. DELETE RECORD RECURRENT
	req, _ = http.NewRequest("DELETE", "/api/record/recurrent/a/delete", nil)
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusBadRequest, response.Code, "should be equal")

	// 34. DELETE RECORD RECURRENT
	req, _ = http.NewRequest("DELETE", "/api/record/recurrent/0/delete", nil)
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusBadRequest, response.Code, "should be equal")

	// 35. DELETE RECORD RECURRENT
	req, _ = http.NewRequest("DELETE", "/api/record/recurrent/3/delete", nil)
	req.Header.Set("sessionID", fsone)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusBadRequest, response.Code, "should be equal")

	// 36. DELETE RECORD RECURRENT
	req, _ = http.NewRequest("DELETE", "/api/record/recurrent/3/delete", nil)
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// 37. PUT RECORD RECURRENT
	req, _ = http.NewRequest("PUT", "/api/record/recurrent/update", strings.NewReader(`{
		"IDstr": "2",
		"Date": "2012-03-31",
		"Recurrence": "annuelle",
		"Account": "CB",
		"Category": "Loyer",
		"Product": "Loyer v2",
		"PriceDirection": "expense",
		"FormPriceStr2Decimals": "6521.04"
	}`))
	req.Header.Set("sessionID", fsone)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusBadRequest, response.Code, "should be equal")

	// 38. PUT RECORD RECURRENT
	req, _ = http.NewRequest("PUT", "/api/record/recurrent/update", strings.NewReader(`{
		"IDstr": "2",
		"Date": "2012-03-31",
		"Recurrence": "annuelle",
		"Account": "CB",
		"Category": "Loyer",
		"Product": "Loyer v2",
		"PriceDirection": "expense",
		"FormPriceStr2Decimals": "6521.04"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// 39. GET RECORD RECURRENT
	req, _ = http.NewRequest("GET", "/api/record/recurrent", nil)
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")
	require.Equal(t,
		"{\"isValidResponse\":true,\"httpStatus\":200,\"info\":\"recurrent record selected\",\"jsonContent\":[{\"ID\":1,\"IDstr\":\"\",\"GofiID\":0,\"Date\":\"2011-06-30\",\"Recurrence\":\"mensuelle\",\"Account\":\"A\",\"Product\":\"Revenu\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"1183.16\",\"PriceIntx100\":118316,\"Category\":\"Banque\"},{\"ID\":2,\"IDstr\":\"\",\"GofiID\":0,\"Date\":\"2012-03-31\",\"Recurrence\":\"annuelle\",\"Account\":\"CB\",\"Product\":\"Loyer v2\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"-6521.04\",\"PriceIntx100\":-652104,\"Category\":\"Loyer\"}]}\n",
		response.Body.String(), "should be equal")

	// 40. PUT RECORD VALIDATE
	req, _ = http.NewRequest("PUT", "/api/record/validate", strings.NewReader(`{
		"Date": "2013-05-31",
		"IDcheckedListStr": "1,2,3,4"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// 41. PUT RECORD VALIDATE
	req, _ = http.NewRequest("PUT", "/api/record/validate", strings.NewReader(`{
		"Date": "2014-07-31",
		"IDcheckedListStr": "a,2,3,4"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusBadRequest, response.Code, "should be equal")

	// 42. GET RECORD
	req, _ = http.NewRequest("GET", "/api/record/id-asc-4", nil)
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")
	require.Equal(t,
		"{\"isValidResponse\":true,\"httpStatus\":200,\"info\":\"record list retrieved\",\"jsonContent\":[{\"ID\":1,\"GofiID\":2,\"Date\":\"2011-11-11\",\"Account\":\"CB\",\"Product\":\"testb\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"-1.00\",\"PriceIntx100\":-100,\"Category\":\"Courses\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":true,\"DateChecked\":\"2013-05-31\",\"Mode\":0,\"Exported\":false},{\"ID\":3,\"GofiID\":2,\"Date\":\"2010-12-31\",\"Account\":\"CB\",\"Product\":\"testb\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"-1.00\",\"PriceIntx100\":-100,\"Category\":\"Courses\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":true,\"DateChecked\":\"2013-05-31\",\"Mode\":0,\"Exported\":false},{\"ID\":4,\"GofiID\":2,\"Date\":\"2012-05-11\",\"Account\":\"A\",\"Product\":\"Transfert-\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"-143.70\",\"PriceIntx100\":-14370,\"Category\":\"Transfert\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":true,\"DateChecked\":\"2013-05-31\",\"Mode\":0,\"Exported\":false},{\"ID\":5,\"GofiID\":2,\"Date\":\"2012-05-11\",\"Account\":\"CB\",\"Product\":\"Transfert+\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"143.70\",\"PriceIntx100\":14370,\"Category\":\"Transfert\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":false,\"DateChecked\":\"9999-12-31\",\"Mode\":0,\"Exported\":false}]}\n",
		response.Body.String(), "should be equal")

	// 43. PUT RECORD CANCEL
	req, _ = http.NewRequest("PUT", "/api/record/cancel", strings.NewReader(`{
		"Date": "2013-06-02",
		"IDcheckedListStr": "9,10"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// 44. GET RECORD
	req, _ = http.NewRequest("GET", "/api/record/id-desc-3", nil)
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")
	require.Equal(t,
		"{\"isValidResponse\":true,\"httpStatus\":200,\"info\":\"record list retrieved\",\"jsonContent\":[{\"ID\":10,\"GofiID\":2,\"Date\":\"1999-12-31\",\"Account\":\"-\",\"Product\":\"DELETED LINE\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"0.00\",\"PriceIntx100\":0,\"Category\":\"-\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":true,\"DateChecked\":\"2013-06-02\",\"Mode\":0,\"Exported\":false},{\"ID\":9,\"GofiID\":2,\"Date\":\"1999-12-31\",\"Account\":\"-\",\"Product\":\"DELETED LINE\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"0.00\",\"PriceIntx100\":0,\"Category\":\"-\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":true,\"DateChecked\":\"2013-06-02\",\"Mode\":0,\"Exported\":false},{\"ID\":8,\"GofiID\":2,\"Date\":\"2012-01-31\",\"Account\":\"CB\",\"Product\":\"ToDelete\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"-34.53\",\"PriceIntx100\":-3453,\"Category\":\"Courses\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":false,\"DateChecked\":\"9999-12-31\",\"Mode\":0,\"Exported\":false}]}\n",
		response.Body.String(), "should be equal")

	// 45. POST RECORD
	req, _ = http.NewRequest("POST", "/api/record/getviapost", strings.NewReader(`{
		"compteHidden": "A",
		"category": "Banque",
		"annee": "2011",
		"mois": "5",
		"checked": "0",
		"OrderBy": "id",
		"OrderSort": "ASC",
		"LimitStr": "2"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")
	require.Equal(t,
		"{\"isValidResponse\":true,\"httpStatus\":200,\"info\":\"record list retrieved\",\"jsonContent\":[{\"ID\":6,\"GofiID\":2,\"Date\":\"2011-05-31\",\"Account\":\"A\",\"Product\":\"Revenu\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"1183.16\",\"PriceIntx100\":118316,\"Category\":\"Banque\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":false,\"DateChecked\":\"9999-12-31\",\"Mode\":0,\"Exported\":false}]}\n",
		response.Body.String(), "should be equal")

	// 46. POST RECORD
	req, _ = http.NewRequest("POST", "/api/record/getviapost", strings.NewReader(`{
		"compteHidden": "CB",
		"category": "Courses",
		"annee": "",
		"mois": "",
		"checked": "0",
		"OrderBy": "id",
		"OrderSort": "ASC",
		"LimitStr": "10"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")
	require.Equal(t,
		"{\"isValidResponse\":true,\"httpStatus\":200,\"info\":\"record list retrieved\",\"jsonContent\":[{\"ID\":1,\"GofiID\":2,\"Date\":\"2011-11-11\",\"Account\":\"CB\",\"Product\":\"testb\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"-1.00\",\"PriceIntx100\":-100,\"Category\":\"Courses\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":true,\"DateChecked\":\"2013-05-31\",\"Mode\":0,\"Exported\":false},{\"ID\":3,\"GofiID\":2,\"Date\":\"2010-12-31\",\"Account\":\"CB\",\"Product\":\"testb\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"-1.00\",\"PriceIntx100\":-100,\"Category\":\"Courses\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":true,\"DateChecked\":\"2013-05-31\",\"Mode\":0,\"Exported\":false},{\"ID\":8,\"GofiID\":2,\"Date\":\"2012-01-31\",\"Account\":\"CB\",\"Product\":\"ToDelete\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"-34.53\",\"PriceIntx100\":-3453,\"Category\":\"Courses\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":false,\"DateChecked\":\"9999-12-31\",\"Mode\":0,\"Exported\":false}]}\n",
		response.Body.String(), "should be equal")

	// 47. POST RECORD LEND BORROW
	req, _ = http.NewRequest("POST", "/api/record/lend-or-borrow", strings.NewReader(`{
		"ModeStr": "1",
		"Who": "-",
		"CreateLenderBorrowerName": "Mr X",
		"FT":{
			"Date": "2011-11-11",
			"Account": "CB",
			"Product": "+ emprunt a1",
			"PriceDirection": "gain",
			"FormPriceStr2Decimals": "1000.00",
			"Category": "Cadeaux"
		}
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusCreated, response.Code, "should be equal")

	// 48. POST RECORD LEND BORROW
	req, _ = http.NewRequest("POST", "/api/record/lend-or-borrow", strings.NewReader(`{
		"ModeStr": "1",
		"Who": "Mr X",
		"FT":{
			"Date": "2010-11-11",
			"Account": "CB",
			"Product": "+ emprunt a2",
			"PriceDirection": "gain",
			"FormPriceStr2Decimals": "1200.00",
			"Category": "Cadeaux"
		}
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusCreated, response.Code, "should be equal")

	// 49. POST RECORD LEND BORROW
	req, _ = http.NewRequest("POST", "/api/record/lend-or-borrow", strings.NewReader(`{
		"ModeStr": "3",
		"Who": "Mr X",
		"FT":{
			"Date": "2010-11-11",
			"Account": "CB",
			"Product": "- remboursement emprunt a3",
			"PriceDirection": "expense",
			"FormPriceStr2Decimals": "2200.00",
			"Category": "Cadeaux"
		}
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusCreated, response.Code, "should be equal")

	// 50. POST RECORD LEND BORROW
	req, _ = http.NewRequest("POST", "/api/record/lend-or-borrow", strings.NewReader(`{
		"ModeStr": "2",
		"Who": "-",
		"CreateLenderBorrowerName": "Mr Y",
		"FT":{
			"Date": "2010-11-11",
			"Account": "CB",
			"Product": "- pret b1 puis dissocié",
			"PriceDirection": "expense",
			"FormPriceStr2Decimals": "1600.00",
			"Category": "Cadeaux"
		}
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusCreated, response.Code, "should be equal")

	// 51. POST RECORD LEND BORROW
	req, _ = http.NewRequest("POST", "/api/record/lend-or-borrow", strings.NewReader(`{
		"ModeStr": "4",
		"Who": "Mr Y",
		"FT":{
			"Date": "2010-11-11",
			"Account": "CB",
			"Product": "+ remb. pret b2 puis dissocié",
			"PriceDirection": "gain",
			"FormPriceStr2Decimals": "1100.00",
			"Category": "Cadeaux"
		}
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusCreated, response.Code, "should be equal")

	// 52. POST UNLINK LEND BORROW
	req, _ = http.NewRequest("POST", "/api/record/lend-or-borrow-unlink", strings.NewReader(`{
		"idsInOneString": "14,15"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// 53. GET RECORD
	req, _ = http.NewRequest("GET", "/api/record/id-desc-5", nil)
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")
	require.Equal(t,
		"{\"isValidResponse\":true,\"httpStatus\":200,\"info\":\"record list retrieved\",\"jsonContent\":[{\"ID\":15,\"GofiID\":2,\"Date\":\"2010-11-11\",\"Account\":\"CB\",\"Product\":\"+ remb. pret b2 puis dissocié\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"1100.00\",\"PriceIntx100\":110000,\"Category\":\"Cadeaux\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":false,\"DateChecked\":\"9999-12-31\",\"Mode\":0,\"Exported\":false},{\"ID\":14,\"GofiID\":2,\"Date\":\"2010-11-11\",\"Account\":\"CB\",\"Product\":\"- pret b1 puis dissocié\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"-1600.00\",\"PriceIntx100\":-160000,\"Category\":\"Cadeaux\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":false,\"DateChecked\":\"9999-12-31\",\"Mode\":0,\"Exported\":false},{\"ID\":13,\"GofiID\":2,\"Date\":\"2010-11-11\",\"Account\":\"CB\",\"Product\":\"- remboursement emprunt a3\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"-2200.00\",\"PriceIntx100\":-220000,\"Category\":\"Cadeaux\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":false,\"DateChecked\":\"9999-12-31\",\"Mode\":3,\"Exported\":false},{\"ID\":12,\"GofiID\":2,\"Date\":\"2010-11-11\",\"Account\":\"CB\",\"Product\":\"+ emprunt a2\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"1200.00\",\"PriceIntx100\":120000,\"Category\":\"Cadeaux\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":false,\"DateChecked\":\"9999-12-31\",\"Mode\":1,\"Exported\":false},{\"ID\":11,\"GofiID\":2,\"Date\":\"2011-11-11\",\"Account\":\"CB\",\"Product\":\"+ emprunt a1\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"1000.00\",\"PriceIntx100\":100000,\"Category\":\"Cadeaux\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":false,\"DateChecked\":\"9999-12-31\",\"Mode\":1,\"Exported\":false}]}\n",
		response.Body.String(), "should be equal")

	// fmt.Printf("response: %#v\n", response.Body.String())
	// require.Equal(t, 1, 0, "force fail")
	require.WithinDuration(t, time.Now(), testStartTime, 5*time.Second)
}
