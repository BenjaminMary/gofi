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

	// RESET IN CASE 2 SIMULTANEOUS USES
	req, _ = http.NewRequest("PUT", "/api/param/category", strings.NewReader(`{
		"idStrJson": "27",
		"type": "all",
		"inStatsStr": "on",
		"description": "-",
		"budgetPriceStr": "0",
		"budgetPeriod": "-",
		"budgetType": "-",
		"budgetCurrentPeriodStartDate": "9999-12-31"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	req, _ = http.NewRequest("PUT", "/api/param/category", strings.NewReader(`{
		"idStrJson": "28",
		"type": "periodic",
		"inStatsStr": "on",
		"description": "-",
		"budgetPriceStr": "0",
		"budgetPeriod": "-",
		"budgetType": "-",
		"budgetCurrentPeriodStartDate": "9999-12-31"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	req, _ = http.NewRequest("PUT", "/api/param/category", strings.NewReader(`{
		"idStrJson": "26",
		"type": "all",
		"inStatsStr": "on",
		"description": "-",
		"budgetPriceStr": "0",
		"budgetPeriod": "-",
		"budgetType": "-",
		"budgetCurrentPeriodStartDate": "9999-12-31"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

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
		"id1StrJson": "28",
		"id2StrJson": "26"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// 18. PATCH PARAM Order
	req, _ = http.NewRequest("PATCH", "/api/param/category/order", strings.NewReader(`{
		"id1StrJson": "29",
		"id2StrJson": "28"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// 19. GET PARAM
	req, _ = http.NewRequest("GET", "/api/param", nil)
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")
	require.Equal(t,
		"{\"isValidResponse\":true,\"httpStatus\":200,\"info\":\"user params retrieved\",\"jsonContent\":{\"GofiID\":2,\"AccountListSingleString\":\"acc1,acc2,acc3\",\"AccountList\":[\"acc1\",\"acc2\",\"acc3\"],\"CategoryListSingleString\":\"cat1,cat2,cat3\",\"CategoryList\":[[\"cat1\",\"e90a\",\"#808080\"],[\"cat2\",\"e90a\",\"#808080\"],[\"cat3\",\"e90a\",\"#808080\"]],\"CategoryRendering\":\"names\",\"Categories\":{\"GofiID\":2,\"FindCategory\":{\"-\":5,\"?\":4,\"Animaux\":22,\"Autre\":23,\"Banque\":17,\"Besoin\":6,\"Cadeaux\":11,\"Courses\":12,\"Enfants\":16,\"Envie\":1,\"Epargne\":0,\"Erreur\":24,\"Habitude-\":7,\"Loisirs\":14,\"Loyer\":19,\"Resto\":13,\"Revenu\":2,\"Sante\":21,\"Services\":20,\"Shopping\":10,\"Societe\":18,\"Transfert\":3,\"Transport\":9,\"Vehicule\":8,\"Voyage\":15},\"Categories\":[{\"ID\":29,\"GofiID\":2,\"Name\":\"Epargne\",\"Type\":\"all\",\"Order\":1,\"InUse\":1,\"InStats\":0,\"Description\":\"-\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e904\",\"ColorHEX\":\"#3380CC\"},{\"ID\":27,\"GofiID\":2,\"Name\":\"Envie\",\"Type\":\"all\",\"Order\":2,\"InUse\":1,\"InStats\":1,\"Description\":\"-\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e920\",\"ColorHEX\":\"#CC8033\"},{\"ID\":28,\"GofiID\":2,\"Name\":\"Revenu\",\"Type\":\"periodic\",\"Order\":4,\"InUse\":1,\"InStats\":1,\"Description\":\"-\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e903\",\"ColorHEX\":\"#33CC99\"},{\"ID\":48,\"GofiID\":2,\"Name\":\"Transfert\",\"Type\":\"specific\",\"Order\":97,\"InUse\":1,\"InStats\":0,\"Description\":\"Utilisé uniquement par le système lors de l'utilisation de la fonction transfert.\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e91b\",\"ColorHEX\":\"#666666\"},{\"ID\":49,\"GofiID\":2,\"Name\":\"?\",\"Type\":\"specific\",\"Order\":98,\"InUse\":1,\"InStats\":0,\"Description\":\"Utilisé uniquement comme icône par le système lorsqu'aucune icône ne correspond à la catégorie demandée.\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e90a\",\"ColorHEX\":\"#808080\"},{\"ID\":50,\"GofiID\":2,\"Name\":\"-\",\"Type\":\"specific\",\"Order\":99,\"InUse\":1,\"InStats\":0,\"Description\":\"Utilisé uniquement par le système lorsqu'on supprime une ligne.\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e90b\",\"ColorHEX\":\"#CC3633\"},{\"ID\":26,\"GofiID\":2,\"Name\":\"Besoin\",\"Type\":\"all\",\"Order\":3,\"InUse\":0,\"InStats\":1,\"Description\":\"-\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e91f\",\"ColorHEX\":\"#33CC4C\"},{\"ID\":30,\"GofiID\":2,\"Name\":\"Habitude-\",\"Type\":\"all\",\"Order\":5,\"InUse\":0,\"InStats\":1,\"Description\":\"-\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e91e\",\"ColorHEX\":\"#CC3633\"},{\"ID\":31,\"GofiID\":2,\"Name\":\"Vehicule\",\"Type\":\"all\",\"Order\":6,\"InUse\":0,\"InStats\":1,\"Description\":\"-\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e900\",\"ColorHEX\":\"#CC5933\"},{\"ID\":32,\"GofiID\":2,\"Name\":\"Transport\",\"Type\":\"all\",\"Order\":7,\"InUse\":0,\"InStats\":1,\"Description\":\"-\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e913\",\"ColorHEX\":\"#CC8033\"},{\"ID\":33,\"GofiID\":2,\"Name\":\"Shopping\",\"Type\":\"basic\",\"Order\":8,\"InUse\":0,\"InStats\":1,\"Description\":\"-\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e918\",\"ColorHEX\":\"#B3994D\"},{\"ID\":34,\"GofiID\":2,\"Name\":\"Cadeaux\",\"Type\":\"basic\",\"Order\":9,\"InUse\":0,\"InStats\":1,\"Description\":\"-\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e91a\",\"ColorHEX\":\"#B3B34D\"},{\"ID\":35,\"GofiID\":2,\"Name\":\"Courses\",\"Type\":\"all\",\"Order\":10,\"InUse\":0,\"InStats\":1,\"Description\":\"-\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e916\",\"ColorHEX\":\"#AABF40\"},{\"ID\":36,\"GofiID\":2,\"Name\":\"Resto\",\"Type\":\"basic\",\"Order\":11,\"InUse\":0,\"InStats\":1,\"Description\":\"-\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e914\",\"ColorHEX\":\"#80CC33\"},{\"ID\":37,\"GofiID\":2,\"Name\":\"Loisirs\",\"Type\":\"all\",\"Order\":12,\"InUse\":0,\"InStats\":1,\"Description\":\"-\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e901\",\"ColorHEX\":\"#4DCC33\"},{\"ID\":38,\"GofiID\":2,\"Name\":\"Voyage\",\"Type\":\"basic\",\"Order\":13,\"InUse\":0,\"InStats\":1,\"Description\":\"-\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e902\",\"ColorHEX\":\"#33CC4C\"},{\"ID\":39,\"GofiID\":2,\"Name\":\"Enfants\",\"Type\":\"all\",\"Order\":14,\"InUse\":0,\"InStats\":1,\"Description\":\"-\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e91d\",\"ColorHEX\":\"#33CCBF\"},{\"ID\":40,\"GofiID\":2,\"Name\":\"Banque\",\"Type\":\"all\",\"Order\":15,\"InUse\":0,\"InStats\":0,\"Description\":\"-\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e919\",\"ColorHEX\":\"#33B3CC\"},{\"ID\":41,\"GofiID\":2,\"Name\":\"Societe\",\"Type\":\"all\",\"Order\":16,\"InUse\":0,\"InStats\":1,\"Description\":\"-\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e905\",\"ColorHEX\":\"#334CCC\"},{\"ID\":42,\"GofiID\":2,\"Name\":\"Loyer\",\"Type\":\"periodic\",\"Order\":17,\"InUse\":0,\"InStats\":1,\"Description\":\"-\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e906\",\"ColorHEX\":\"#6633CC\"},{\"ID\":43,\"GofiID\":2,\"Name\":\"Services\",\"Type\":\"periodic\",\"Order\":18,\"InUse\":0,\"InStats\":1,\"Description\":\"-\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e907\",\"ColorHEX\":\"#8033CC\"},{\"ID\":44,\"GofiID\":2,\"Name\":\"Sante\",\"Type\":\"all\",\"Order\":19,\"InUse\":0,\"InStats\":1,\"Description\":\"-\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e908\",\"ColorHEX\":\"#CC33CC\"},{\"ID\":45,\"GofiID\":2,\"Name\":\"Animaux\",\"Type\":\"all\",\"Order\":20,\"InUse\":0,\"InStats\":1,\"Description\":\"-\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e91c\",\"ColorHEX\":\"#CC3399\"},{\"ID\":46,\"GofiID\":2,\"Name\":\"Autre\",\"Type\":\"basic\",\"Order\":21,\"InUse\":0,\"InStats\":1,\"Description\":\"Permet de ranger un élément qu'on ne sait pas où placer, temporairement ou définitivement.\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e90c\",\"ColorHEX\":\"#999999\"},{\"ID\":47,\"GofiID\":2,\"Name\":\"Erreur\",\"Type\":\"basic\",\"Order\":22,\"InUse\":0,\"InStats\":1,\"Description\":\"Utile lorsqu'on souhaite corriger un montant global sans savoir réellement quel était l'achat en question.\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e909\",\"ColorHEX\":\"#CC3373\"}]}}}\n",
		response.Body.String(), "should be equal")

	// roll back order changes to be able to play this test 2 times in a row with exactly the same result
	// 20. PATCH PARAM Order
	req, _ = http.NewRequest("PATCH", "/api/param/category/order", strings.NewReader(`{
		"id1StrJson": "28",
		"id2StrJson": "29"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// 21. PATCH PARAM Order
	req, _ = http.NewRequest("PATCH", "/api/param/category/order", strings.NewReader(`{
		"id1StrJson": "26",
		"id2StrJson": "28"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// 22. PATCH PARAM Order
	req, _ = http.NewRequest("PUT", "/api/param/category", strings.NewReader(`{
		"idStrJson": "27",
		"type": "basic",
		"inStatsStr": "",
		"description": "test description category envie",
		"budgetPriceStr": "100",
		"budgetPeriod": "month",
		"budgetType": "reset",
		"budgetCurrentPeriodStartDate": "2000-01-01"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// 23. PATCH PARAM Order
	req, _ = http.NewRequest("PUT", "/api/param/category", strings.NewReader(`{
		"idStrJson": "28",
		"type": "periodic",
		"inStatsStr": "on",
		"description": "revenus",
		"budgetPriceStr": "0",
		"budgetPeriod": "-",
		"budgetType": "-",
		"budgetCurrentPeriodStartDate": "9999-12-31"
	}`))
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")

	// 24. GET PARAM
	req, _ = http.NewRequest("GET", "/api/param", nil)
	req.Header.Set("sessionID", fstwo)
	response = executeRequest(req, s)
	require.Equal(t, http.StatusOK, response.Code, "should be equal")
	require.Equal(t,
		"{\"isValidResponse\":true,\"httpStatus\":200,\"info\":\"user params retrieved\",\"jsonContent\":{\"GofiID\":2,\"AccountListSingleString\":\"acc1,acc2,acc3\",\"AccountList\":[\"acc1\",\"acc2\",\"acc3\"],\"CategoryListSingleString\":\"cat1,cat2,cat3\",\"CategoryList\":[[\"cat1\",\"e90a\",\"#808080\"],[\"cat2\",\"e90a\",\"#808080\"],[\"cat3\",\"e90a\",\"#808080\"]],\"CategoryRendering\":\"names\",\"Categories\":{\"GofiID\":2,\"FindCategory\":{\"-\":5,\"?\":4,\"Animaux\":22,\"Autre\":23,\"Banque\":17,\"Besoin\":6,\"Cadeaux\":11,\"Courses\":12,\"Enfants\":16,\"Envie\":0,\"Epargne\":2,\"Erreur\":24,\"Habitude-\":7,\"Loisirs\":14,\"Loyer\":19,\"Resto\":13,\"Revenu\":1,\"Sante\":21,\"Services\":20,\"Shopping\":10,\"Societe\":18,\"Transfert\":3,\"Transport\":9,\"Vehicule\":8,\"Voyage\":15},\"Categories\":[{\"ID\":27,\"GofiID\":2,\"Name\":\"Envie\",\"Type\":\"basic\",\"Order\":2,\"InUse\":1,\"InStats\":0,\"Description\":\"test description category envie\",\"BudgetPrice\":100,\"BudgetPeriod\":\"month\",\"BudgetType\":\"reset\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e920\",\"ColorHEX\":\"#CC8033\"},{\"ID\":28,\"GofiID\":2,\"Name\":\"Revenu\",\"Type\":\"periodic\",\"Order\":3,\"InUse\":1,\"InStats\":1,\"Description\":\"revenus\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e903\",\"ColorHEX\":\"#33CC99\"},{\"ID\":29,\"GofiID\":2,\"Name\":\"Epargne\",\"Type\":\"all\",\"Order\":4,\"InUse\":1,\"InStats\":0,\"Description\":\"-\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e904\",\"ColorHEX\":\"#3380CC\"},{\"ID\":48,\"GofiID\":2,\"Name\":\"Transfert\",\"Type\":\"specific\",\"Order\":97,\"InUse\":1,\"InStats\":0,\"Description\":\"Utilisé uniquement par le système lors de l'utilisation de la fonction transfert.\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e91b\",\"ColorHEX\":\"#666666\"},{\"ID\":49,\"GofiID\":2,\"Name\":\"?\",\"Type\":\"specific\",\"Order\":98,\"InUse\":1,\"InStats\":0,\"Description\":\"Utilisé uniquement comme icône par le système lorsqu'aucune icône ne correspond à la catégorie demandée.\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e90a\",\"ColorHEX\":\"#808080\"},{\"ID\":50,\"GofiID\":2,\"Name\":\"-\",\"Type\":\"specific\",\"Order\":99,\"InUse\":1,\"InStats\":0,\"Description\":\"Utilisé uniquement par le système lorsqu'on supprime une ligne.\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e90b\",\"ColorHEX\":\"#CC3633\"},{\"ID\":26,\"GofiID\":2,\"Name\":\"Besoin\",\"Type\":\"all\",\"Order\":1,\"InUse\":0,\"InStats\":1,\"Description\":\"-\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e91f\",\"ColorHEX\":\"#33CC4C\"},{\"ID\":30,\"GofiID\":2,\"Name\":\"Habitude-\",\"Type\":\"all\",\"Order\":5,\"InUse\":0,\"InStats\":1,\"Description\":\"-\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e91e\",\"ColorHEX\":\"#CC3633\"},{\"ID\":31,\"GofiID\":2,\"Name\":\"Vehicule\",\"Type\":\"all\",\"Order\":6,\"InUse\":0,\"InStats\":1,\"Description\":\"-\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e900\",\"ColorHEX\":\"#CC5933\"},{\"ID\":32,\"GofiID\":2,\"Name\":\"Transport\",\"Type\":\"all\",\"Order\":7,\"InUse\":0,\"InStats\":1,\"Description\":\"-\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e913\",\"ColorHEX\":\"#CC8033\"},{\"ID\":33,\"GofiID\":2,\"Name\":\"Shopping\",\"Type\":\"basic\",\"Order\":8,\"InUse\":0,\"InStats\":1,\"Description\":\"-\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e918\",\"ColorHEX\":\"#B3994D\"},{\"ID\":34,\"GofiID\":2,\"Name\":\"Cadeaux\",\"Type\":\"basic\",\"Order\":9,\"InUse\":0,\"InStats\":1,\"Description\":\"-\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e91a\",\"ColorHEX\":\"#B3B34D\"},{\"ID\":35,\"GofiID\":2,\"Name\":\"Courses\",\"Type\":\"all\",\"Order\":10,\"InUse\":0,\"InStats\":1,\"Description\":\"-\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e916\",\"ColorHEX\":\"#AABF40\"},{\"ID\":36,\"GofiID\":2,\"Name\":\"Resto\",\"Type\":\"basic\",\"Order\":11,\"InUse\":0,\"InStats\":1,\"Description\":\"-\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e914\",\"ColorHEX\":\"#80CC33\"},{\"ID\":37,\"GofiID\":2,\"Name\":\"Loisirs\",\"Type\":\"all\",\"Order\":12,\"InUse\":0,\"InStats\":1,\"Description\":\"-\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e901\",\"ColorHEX\":\"#4DCC33\"},{\"ID\":38,\"GofiID\":2,\"Name\":\"Voyage\",\"Type\":\"basic\",\"Order\":13,\"InUse\":0,\"InStats\":1,\"Description\":\"-\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e902\",\"ColorHEX\":\"#33CC4C\"},{\"ID\":39,\"GofiID\":2,\"Name\":\"Enfants\",\"Type\":\"all\",\"Order\":14,\"InUse\":0,\"InStats\":1,\"Description\":\"-\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e91d\",\"ColorHEX\":\"#33CCBF\"},{\"ID\":40,\"GofiID\":2,\"Name\":\"Banque\",\"Type\":\"all\",\"Order\":15,\"InUse\":0,\"InStats\":0,\"Description\":\"-\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e919\",\"ColorHEX\":\"#33B3CC\"},{\"ID\":41,\"GofiID\":2,\"Name\":\"Societe\",\"Type\":\"all\",\"Order\":16,\"InUse\":0,\"InStats\":1,\"Description\":\"-\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e905\",\"ColorHEX\":\"#334CCC\"},{\"ID\":42,\"GofiID\":2,\"Name\":\"Loyer\",\"Type\":\"periodic\",\"Order\":17,\"InUse\":0,\"InStats\":1,\"Description\":\"-\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e906\",\"ColorHEX\":\"#6633CC\"},{\"ID\":43,\"GofiID\":2,\"Name\":\"Services\",\"Type\":\"periodic\",\"Order\":18,\"InUse\":0,\"InStats\":1,\"Description\":\"-\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e907\",\"ColorHEX\":\"#8033CC\"},{\"ID\":44,\"GofiID\":2,\"Name\":\"Sante\",\"Type\":\"all\",\"Order\":19,\"InUse\":0,\"InStats\":1,\"Description\":\"-\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e908\",\"ColorHEX\":\"#CC33CC\"},{\"ID\":45,\"GofiID\":2,\"Name\":\"Animaux\",\"Type\":\"all\",\"Order\":20,\"InUse\":0,\"InStats\":1,\"Description\":\"-\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e91c\",\"ColorHEX\":\"#CC3399\"},{\"ID\":46,\"GofiID\":2,\"Name\":\"Autre\",\"Type\":\"basic\",\"Order\":21,\"InUse\":0,\"InStats\":1,\"Description\":\"Permet de ranger un élément qu'on ne sait pas où placer, temporairement ou définitivement.\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e90c\",\"ColorHEX\":\"#999999\"},{\"ID\":47,\"GofiID\":2,\"Name\":\"Erreur\",\"Type\":\"basic\",\"Order\":22,\"InUse\":0,\"InStats\":1,\"Description\":\"Utile lorsqu'on souhaite corriger un montant global sans savoir réellement quel était l'achat en question.\",\"BudgetPrice\":0,\"BudgetPeriod\":\"-\",\"BudgetType\":\"-\",\"BudgetCurrentPeriodStartDate\":\"\",\"BudgetCurrentPeriodEndDate\":\"\",\"IconCodePoint\":\"e909\",\"ColorHEX\":\"#CC3373\"}]}}}\n",
		response.Body.String(), "should be equal")

	// fmt.Printf("response: %#v\n", response.Body.String())
	// require.Equal(t, 1, 0, "force fail")
	require.WithinDuration(t, time.Now(), testStartTime, 5*time.Second)
}
