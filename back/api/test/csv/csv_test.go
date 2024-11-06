package apiparamtest

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
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
func executeRequestNoPresetContentType(req *http.Request, s *routes.Server) *httptest.ResponseRecorder {
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

	fmt.Println("cleaning specificRecordsByMode table")
	_, err = appdata.DB.Exec(`
		DELETE FROM specificRecordsByMode;
		DELETE FROM SQLITE_SEQUENCE WHERE name='specificRecordsByMode';
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

	// CSV
	var body *bytes.Buffer
	var writer *multipart.Writer
	var part io.Writer
	var file *os.File

	// 7. CSV import
	body = new(bytes.Buffer)
	writer = multipart.NewWriter(body)
	part, _ = writer.CreateFormFile("csvFile", "gofi1-UTF8-LF.csv")
	file, _ = os.Open("C:/git/gofi/back/api/test/csv/gofi1-UTF8-LF.csv")
	io.Copy(part, file)
	writer.Close()
	req, _ = http.NewRequest("POST", "/api/csv/import", body)
	req.Header.Set("sessionID", fstwo)
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+writer.Boundary())
	response = executeRequestNoPresetContentType(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// GetRowsInFinanceTracker does not give info for CommentInt and CommentString fields
	// 8. GET RECORD
	req, _ = http.NewRequest("GET", "/api/record/id-asc-50", nil)
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")
	require.Equal(t,
		"{\"isValidResponse\":true,\"httpStatus\":200,\"info\":\"record list retrieved\",\"jsonContent\":[{\"ID\":1,\"GofiID\":2,\"Date\":\"1999-12-31\",\"Account\":\"-\",\"Product\":\"DELETED LINE\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"0.00\",\"PriceIntx100\":0,\"Category\":\"-\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":true,\"DateChecked\":\"2024-01-14\",\"Exported\":false},{\"ID\":2,\"GofiID\":2,\"Date\":\"2015-01-01\",\"Account\":\"Especes\",\"Product\":\"Solde Espèces\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"83.66\",\"PriceIntx100\":8366,\"Category\":\"Revenu\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":true,\"DateChecked\":\"2000-01-01\",\"Exported\":false},{\"ID\":3,\"GofiID\":2,\"Date\":\"2015-01-02\",\"Account\":\"CA\",\"Product\":\"PC Portable\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"-295.11\",\"PriceIntx100\":-29511,\"Category\":\"HighTech\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":true,\"DateChecked\":\"2015-01-05\",\"Exported\":false},{\"ID\":4,\"GofiID\":2,\"Date\":\"2015-01-04\",\"Account\":\"CA\",\"Product\":\"Gazole\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"-38.66\",\"PriceIntx100\":-3866,\"Category\":\"Vehicule\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":true,\"DateChecked\":\"2015-01-05\",\"Exported\":false},{\"ID\":5,\"GofiID\":2,\"Date\":\"2015-01-05\",\"Account\":\"CA\",\"Product\":\"Jeux Steam\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"-1.89\",\"PriceIntx100\":-189,\"Category\":\"Loisirs\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":true,\"DateChecked\":\"2015-01-06\",\"Exported\":false},{\"ID\":6,\"GofiID\":2,\"Date\":\"2015-01-06\",\"Account\":\"CA\",\"Product\":\"Série S1 à S5\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"-34.99\",\"PriceIntx100\":-3499,\"Category\":\"Cadeaux\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":true,\"DateChecked\":\"2015-01-07\",\"Exported\":false},{\"ID\":7,\"GofiID\":2,\"Date\":\"2015-01-06\",\"Account\":\"Especes\",\"Product\":\"Cafés\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"-1.60\",\"PriceIntx100\":-160,\"Category\":\"Courses\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":true,\"DateChecked\":\"2000-01-01\",\"Exported\":false},{\"ID\":8,\"GofiID\":2,\"Date\":\"2015-01-07\",\"Account\":\"CA\",\"Product\":\"VÊTEMENT HOMME SKI TOP\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"-13.95\",\"PriceIntx100\":-1395,\"Category\":\"Voyage\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":true,\"DateChecked\":\"2000-01-01\",\"Exported\":false},{\"ID\":9,\"GofiID\":2,\"Date\":\"2015-01-08\",\"Account\":\"CA\",\"Product\":\"Prêt\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"-150.00\",\"PriceIntx100\":-15000,\"Category\":\"Voyage\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":true,\"DateChecked\":\"2015-01-09\",\"Exported\":false},{\"ID\":10,\"GofiID\":2,\"Date\":\"2015-01-09\",\"Account\":\"CA\",\"Product\":\"Remboursement Prêt\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"150.00\",\"PriceIntx100\":15000,\"Category\":\"Voyage\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":true,\"DateChecked\":\"2015-01-10\",\"Exported\":false},{\"ID\":11,\"GofiID\":2,\"Date\":\"2015-01-10\",\"Account\":\"CA\",\"Product\":\"Emprunt\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"300.00\",\"PriceIntx100\":30000,\"Category\":\"Voyage\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":true,\"DateChecked\":\"2015-01-11\",\"Exported\":false},{\"ID\":12,\"GofiID\":2,\"Date\":\"2015-01-11\",\"Account\":\"CA\",\"Product\":\"Remboursement Emprunt\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"-100.00\",\"PriceIntx100\":-10000,\"Category\":\"Voyage\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":true,\"DateChecked\":\"2015-01-12\",\"Exported\":false},{\"ID\":13,\"GofiID\":2,\"Date\":\"2015-01-12\",\"Account\":\"CA\",\"Product\":\"Remboursement Emprunt\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"-200.00\",\"PriceIntx100\":-20000,\"Category\":\"Voyage\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":true,\"DateChecked\":\"2015-01-13\",\"Exported\":false}]}\n",
		response.Body.String(), "should be equal")

	// 9. CSV import
	body = new(bytes.Buffer)
	writer = multipart.NewWriter(body)
	part, _ = writer.CreateFormFile("csvFile", "gofi2-UTF8-LF.csv")
	file, _ = os.Open("C:/git/gofi/back/api/test/csv/gofi2-UTF8-LF.csv")
	io.Copy(part, file)
	writer.Close()
	req, _ = http.NewRequest("POST", "/api/csv/import", body)
	req.Header.Set("sessionID", fstwo)
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+writer.Boundary())
	response = executeRequestNoPresetContentType(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// GetRowsInFinanceTracker does not give info for CommentInt and CommentString fields
	// 10. GET RECORD
	req, _ = http.NewRequest("GET", "/api/record/id-asc-50", nil)
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")
	require.Equal(t,
		"{\"isValidResponse\":true,\"httpStatus\":200,\"info\":\"record list retrieved\",\"jsonContent\":[{\"ID\":1,\"GofiID\":2,\"Date\":\"1999-12-31\",\"Account\":\"-\",\"Product\":\"DELETED LINE\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"0.00\",\"PriceIntx100\":0,\"Category\":\"-\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":true,\"DateChecked\":\"1999-12-31\",\"Exported\":false},{\"ID\":2,\"GofiID\":2,\"Date\":\"2015-01-01\",\"Account\":\"Especes\",\"Product\":\"Solde Espèces\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"85.76\",\"PriceIntx100\":8576,\"Category\":\"Revenu\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":true,\"DateChecked\":\"2023-01-01\",\"Exported\":false},{\"ID\":3,\"GofiID\":2,\"Date\":\"2015-01-02\",\"Account\":\"CA\",\"Product\":\"PC Portable\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"-295.11\",\"PriceIntx100\":-29511,\"Category\":\"HighTech\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":true,\"DateChecked\":\"2015-01-05\",\"Exported\":false},{\"ID\":4,\"GofiID\":2,\"Date\":\"2015-01-04\",\"Account\":\"CA\",\"Product\":\"Gazole\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"-38.66\",\"PriceIntx100\":-3866,\"Category\":\"Vehicule\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":true,\"DateChecked\":\"2015-01-05\",\"Exported\":false},{\"ID\":5,\"GofiID\":2,\"Date\":\"2015-01-05\",\"Account\":\"CA\",\"Product\":\"Jeux Steam Remboursement Emprunt\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"-2.00\",\"PriceIntx100\":-200,\"Category\":\"Loisirs\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":true,\"DateChecked\":\"2015-01-06\",\"Exported\":false},{\"ID\":6,\"GofiID\":2,\"Date\":\"2015-01-06\",\"Account\":\"CA\",\"Product\":\"Série S1 à S5 Prêt\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"-34.99\",\"PriceIntx100\":-3499,\"Category\":\"Cadeaux\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":true,\"DateChecked\":\"2015-01-07\",\"Exported\":false},{\"ID\":7,\"GofiID\":2,\"Date\":\"1999-12-31\",\"Account\":\"-\",\"Product\":\"DELETED LINE\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"0.00\",\"PriceIntx100\":0,\"Category\":\"-\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":true,\"DateChecked\":\"1999-12-31\",\"Exported\":false},{\"ID\":8,\"GofiID\":2,\"Date\":\"2015-01-07\",\"Account\":\"CA\",\"Product\":\"VÊTEMENT HOMME SKI TOP\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"-13.95\",\"PriceIntx100\":-1395,\"Category\":\"Voyage\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":true,\"DateChecked\":\"2024-01-01\",\"Exported\":false},{\"ID\":9,\"GofiID\":2,\"Date\":\"2015-01-08\",\"Account\":\"CA\",\"Product\":\"Prêt\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"-150.00\",\"PriceIntx100\":-15000,\"Category\":\"Voyage\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":true,\"DateChecked\":\"2015-01-09\",\"Exported\":false},{\"ID\":10,\"GofiID\":2,\"Date\":\"2015-01-09\",\"Account\":\"CA\",\"Product\":\"Remboursement Prêt\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"150.00\",\"PriceIntx100\":15000,\"Category\":\"Voyage\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":true,\"DateChecked\":\"2015-01-10\",\"Exported\":false},{\"ID\":11,\"GofiID\":2,\"Date\":\"2015-01-10\",\"Account\":\"CA\",\"Product\":\"Emprunt\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"300.00\",\"PriceIntx100\":30000,\"Category\":\"Voyage\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":true,\"DateChecked\":\"2015-01-11\",\"Exported\":false},{\"ID\":12,\"GofiID\":2,\"Date\":\"2015-01-11\",\"Account\":\"CA\",\"Product\":\"Remboursement Emprunt\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"-100.00\",\"PriceIntx100\":-10000,\"Category\":\"Voyage\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":true,\"DateChecked\":\"2015-01-12\",\"Exported\":false},{\"ID\":13,\"GofiID\":2,\"Date\":\"1999-12-31\",\"Account\":\"-\",\"Product\":\"DELETED LINE\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"0.00\",\"PriceIntx100\":0,\"Category\":\"-\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":true,\"DateChecked\":\"1999-12-31\",\"Exported\":false}]}\n",
		response.Body.String(), "should be equal")

	// 11. CSV import CRLF
	body = new(bytes.Buffer)
	writer = multipart.NewWriter(body)
	part, _ = writer.CreateFormFile("csvFile", "gofi3-UTF8-CRLF.csv")
	file, _ = os.Open("C:/git/gofi/back/api/test/csv/gofi3-UTF8-CRLF.csv")
	io.Copy(part, file)
	writer.Close()
	req, _ = http.NewRequest("POST", "/api/csv/import", body)
	req.Header.Set("sessionID", fstwo)
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+writer.Boundary())
	response = executeRequestNoPresetContentType(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// 12. GET RECORD
	req, _ = http.NewRequest("GET", "/api/record/id-asc-8", nil)
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")
	require.Equal(t,
		"{\"isValidResponse\":true,\"httpStatus\":200,\"info\":\"record list retrieved\",\"jsonContent\":[{\"ID\":1,\"GofiID\":2,\"Date\":\"1999-12-31\",\"Account\":\"-\",\"Product\":\"DELETED LINE\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"0.00\",\"PriceIntx100\":0,\"Category\":\"-\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":true,\"DateChecked\":\"1999-12-31\",\"Exported\":false},{\"ID\":2,\"GofiID\":2,\"Date\":\"2015-01-01\",\"Account\":\"Especes\",\"Product\":\"Solde Espèces CRLF\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"85.76\",\"PriceIntx100\":8576,\"Category\":\"Revenu\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":true,\"DateChecked\":\"2023-01-01\",\"Exported\":false},{\"ID\":3,\"GofiID\":2,\"Date\":\"2015-01-02\",\"Account\":\"CA\",\"Product\":\"PC Portable\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"-295.11\",\"PriceIntx100\":-29511,\"Category\":\"HighTech\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":true,\"DateChecked\":\"2015-01-05\",\"Exported\":false},{\"ID\":4,\"GofiID\":2,\"Date\":\"2015-01-04\",\"Account\":\"CA\",\"Product\":\"Gazole\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"-38.66\",\"PriceIntx100\":-3866,\"Category\":\"Vehicule\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":true,\"DateChecked\":\"2015-01-05\",\"Exported\":false},{\"ID\":5,\"GofiID\":2,\"Date\":\"2015-01-05\",\"Account\":\"CA\",\"Product\":\"Jeux Steam Remboursement Emprunt\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"-2.00\",\"PriceIntx100\":-200,\"Category\":\"Loisirs\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":true,\"DateChecked\":\"2015-01-06\",\"Exported\":false},{\"ID\":6,\"GofiID\":2,\"Date\":\"2015-01-06\",\"Account\":\"CA\",\"Product\":\"Série S1 à S5 Prêt\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"-34.99\",\"PriceIntx100\":-3499,\"Category\":\"Cadeaux\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":true,\"DateChecked\":\"2015-01-07\",\"Exported\":false},{\"ID\":7,\"GofiID\":2,\"Date\":\"1999-12-31\",\"Account\":\"-\",\"Product\":\"DELETED LINE\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"0.00\",\"PriceIntx100\":0,\"Category\":\"-\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":true,\"DateChecked\":\"1999-12-31\",\"Exported\":false},{\"ID\":8,\"GofiID\":2,\"Date\":\"2015-01-07\",\"Account\":\"CA\",\"Product\":\"VÊTEMENT HOMME SKI TOP\",\"PriceDirection\":\"\",\"FormPriceStr2Decimals\":\"-13.95\",\"PriceIntx100\":-1395,\"Category\":\"Voyage\",\"CommentInt\":0,\"CommentString\":\"\",\"Checked\":true,\"DateChecked\":\"2024-01-01\",\"Exported\":false}]}\n",
		response.Body.String(), "should be equal")

	// 13. CSV import
	body = new(bytes.Buffer)
	writer = multipart.NewWriter(body)
	part, _ = writer.CreateFormFile("csvFile", "gofi4-UTF8-LF.csv")
	file, _ = os.Open("C:/git/gofi/back/api/test/csv/gofi4-UTF8-LF.csv")
	io.Copy(part, file)
	writer.Close()
	req, _ = http.NewRequest("POST", "/api/csv/import", body)
	req.Header.Set("sessionID", fstwo)
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+writer.Boundary())
	response = executeRequestNoPresetContentType(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// 14. CSV import
	body = new(bytes.Buffer)
	writer = multipart.NewWriter(body)
	part, _ = writer.CreateFormFile("csvFile", "gofi5-UTF8-LF.csv")
	file, _ = os.Open("C:/git/gofi/back/api/test/csv/gofi5-UTF8-LF.csv")
	io.Copy(part, file)
	writer.Close()
	req, _ = http.NewRequest("POST", "/api/csv/import", body)
	req.Header.Set("sessionID", fstwo)
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+writer.Boundary())
	response = executeRequestNoPresetContentType(req, s)
	require.Equal(t, http.StatusBadRequest, response.Code, "should be equal")

	// 15. CSV import
	body = new(bytes.Buffer)
	writer = multipart.NewWriter(body)
	part, _ = writer.CreateFormFile("csvFile", "gofi6-UTF8-LF.csv")
	file, _ = os.Open("C:/git/gofi/back/api/test/csv/gofi6-UTF8-LF.csv")
	io.Copy(part, file)
	writer.Close()
	req, _ = http.NewRequest("POST", "/api/csv/import", body)
	req.Header.Set("sessionID", fstwo)
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+writer.Boundary())
	response = executeRequestNoPresetContentType(req, s)
	require.Equal(t, http.StatusBadRequest, response.Code, "should be equal")

	// 16. CSV import
	body = new(bytes.Buffer)
	writer = multipart.NewWriter(body)
	part, _ = writer.CreateFormFile("csvFile", "gofi7-UTF8-LF.csv")
	file, _ = os.Open("C:/git/gofi/back/api/test/csv/gofi7-UTF8-LF.csv")
	io.Copy(part, file)
	writer.Close()
	req, _ = http.NewRequest("POST", "/api/csv/import", body)
	req.Header.Set("sessionID", fstwo)
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+writer.Boundary())
	response = executeRequestNoPresetContentType(req, s)
	require.Equal(t, http.StatusBadRequest, response.Code, "should be equal")

	// 17. CSV import
	body = new(bytes.Buffer)
	writer = multipart.NewWriter(body)
	part, _ = writer.CreateFormFile("csvFile", "gofi8-UTF8-LF.csv")
	file, _ = os.Open("C:/git/gofi/back/api/test/csv/gofi8-UTF8-LF.csv")
	io.Copy(part, file)
	writer.Close()
	req, _ = http.NewRequest("POST", "/api/csv/import", body)
	req.Header.Set("sessionID", fstwo)
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+writer.Boundary())
	response = executeRequestNoPresetContentType(req, s)
	require.Equal(t, http.StatusBadRequest, response.Code, "should be equal")

	fmt.Println("IMPORT ended")
	fmt.Println("EXPORT started")

	// 18. CSV export
	req, _ = http.NewRequest("POST", "/api/csv/export", nil)
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")
	require.Equal(t,
		"𫝀é ꮖꭰ;Date;Mode;Account;Product;PriceStr;Category;ThirdParty;CommentInt;CommentString;Checked;DateChecked;Exported;.\n1;31/12/1999;0;-;DELETED LINE;0,00;-;;0;;true;1999-12-31;true;.\n2;01/01/2015;0;Especes;Solde Espèces CRLF;85,76;Revenu;;0;;true;2023-01-01;true;.\n3;02/01/2015;0;CA;PC Portable Modif noms colonnes;-295,11;HighTech;;0;5heures autonomie;true;2015-01-05;true;.\n4;04/01/2015;0;CA;Gazole;-38,66;Vehicule;;200079;Kms au moment du plein;true;2015-01-05;true;.\n5;05/01/2015;3;CA;Jeux Steam Remboursement Emprunt;-2,00;Loisirs;Tier2;0;;true;2015-01-06;true;.\n6;06/01/2015;2;CA;Série S1 à S5 Prêt;-34,99;Cadeaux;Tier5;0;;true;2015-01-07;true;.\n7;31/12/1999;0;-;DELETED LINE;0,00;-;;0;;true;1999-12-31;true;.\n8;07/01/2015;0;CA;VÊTEMENT HOMME SKI TOP;-13,95;Voyage;;0;;true;2024-01-01;true;.\n9;08/01/2015;2;CA;Prêt;-150,00;Voyage;Tier1;9;id9;true;2015-01-09;true;.\n10;09/01/2015;0;CA;Remboursement Prêt;150,00;Voyage;;10;id10;true;2015-01-10;true;.\n11;10/01/2015;1;CA;Emprunt;300,00;Voyage;Tier2;11;id11;true;2015-01-11;true;.\n12;11/01/2015;3;CA;Remboursement Emprunt;-100,00;Voyage;Tier2;12;id12;true;2015-01-12;true;.\n13;31/12/1999;0;-;DELETED LINE;0,00;-;;0;;true;1999-12-31;true;.\n",
		response.Body.String(), "should be equal")
	require.Equal(t, http.Header(http.Header{"Content-Disposition": []string{"inline; filename=gofi-2.csv"}, "Content-Type": []string{"application/octet-stream"}}), response.Header(), "should be equal")

	// 19. CSV export
	req, _ = http.NewRequest("POST", "/api/csv/export", nil)
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")
	require.Equal(t,
		"Rien à télécharger\n",
		response.Body.String(), "should be equal")
	require.Equal(t, http.Header(http.Header{"Content-Disposition": []string{"inline; filename=gofi-2.csv"}, "Content-Type": []string{"application/octet-stream"}}), response.Header(), "should be equal")

	// 20. CSV export reset
	req, _ = http.NewRequest("POST", "/api/csv/export/reset", nil)
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// 21. CSV export
	req, _ = http.NewRequest("POST", "/api/csv/export", nil)
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")
	require.Equal(t,
		"𫝀é ꮖꭰ;Date;Mode;Account;Product;PriceStr;Category;ThirdParty;CommentInt;CommentString;Checked;DateChecked;Exported;.\n1;31/12/1999;0;-;DELETED LINE;0,00;-;;0;;true;1999-12-31;true;.\n2;01/01/2015;0;Especes;Solde Espèces CRLF;85,76;Revenu;;0;;true;2023-01-01;true;.\n3;02/01/2015;0;CA;PC Portable Modif noms colonnes;-295,11;HighTech;;0;5heures autonomie;true;2015-01-05;true;.\n4;04/01/2015;0;CA;Gazole;-38,66;Vehicule;;200079;Kms au moment du plein;true;2015-01-05;true;.\n5;05/01/2015;3;CA;Jeux Steam Remboursement Emprunt;-2,00;Loisirs;Tier2;0;;true;2015-01-06;true;.\n6;06/01/2015;2;CA;Série S1 à S5 Prêt;-34,99;Cadeaux;Tier5;0;;true;2015-01-07;true;.\n7;31/12/1999;0;-;DELETED LINE;0,00;-;;0;;true;1999-12-31;true;.\n8;07/01/2015;0;CA;VÊTEMENT HOMME SKI TOP;-13,95;Voyage;;0;;true;2024-01-01;true;.\n9;08/01/2015;2;CA;Prêt;-150,00;Voyage;Tier1;9;id9;true;2015-01-09;true;.\n10;09/01/2015;0;CA;Remboursement Prêt;150,00;Voyage;;10;id10;true;2015-01-10;true;.\n11;10/01/2015;1;CA;Emprunt;300,00;Voyage;Tier2;11;id11;true;2015-01-11;true;.\n12;11/01/2015;3;CA;Remboursement Emprunt;-100,00;Voyage;Tier2;12;id12;true;2015-01-12;true;.\n13;31/12/1999;0;-;DELETED LINE;0,00;-;;0;;true;1999-12-31;true;.\n",
		response.Body.String(), "should be equal")
	require.Equal(t, http.Header(http.Header{"Content-Disposition": []string{"inline; filename=gofi-2.csv"}, "Content-Type": []string{"application/octet-stream"}}), response.Header(), "should be equal")

	// fmt.Printf("response: %#v\n", response.Body.String())
	// require.Equal(t, 1, 0, "force fail")
	require.WithinDuration(t, time.Now(), testStartTime, 5*time.Second)
}
