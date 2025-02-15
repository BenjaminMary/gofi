package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"gofi/gofi/data/appdata"
	"log"
	"strings"
	"slices"
)

func InsertRowInParam(ctx context.Context, db *sql.DB, p *appdata.Param) (int64, error) {
	result, err := db.ExecContext(ctx, ` 
		INSERT OR REPLACE INTO param (gofiID, paramName, paramJSONstringData, paramInfo)
		VALUES (?,?,?,?);
		`,
		p.GofiID, p.ParamName, p.ParamJSONstringData, p.ParamInfo,
	)
	if err != nil {
		fmt.Printf("error INSERT OR REPLACE: %v\n", err)
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		fmt.Printf("error LastInsertId: %v\n", err)
		return 0, err
	}
	return id, nil
}

func CheckIfIdExists(ctx context.Context, db *sql.DB, gofiID int) {
	//if new ID, create default params
	var nbRows int = 0
	var P appdata.Param

	q := ` 
		SELECT COUNT(1)
		FROM param
		WHERE gofiID = ?
			AND paramName = ?;
	`
	err := db.QueryRowContext(ctx, q, gofiID, "accountList").Scan(&nbRows)
	switch {
	case err == sql.ErrNoRows:
		nbRows = 0
	case err != nil:
		log.Fatalf("query error param accountList: %v\n", err)
	}
	if nbRows != 1 {
		db.ExecContext(ctx, "DELETE FROM param WHERE gofiID = ? AND paramName = 'accountList';", gofiID)
		P.GofiID = gofiID
		P.ParamName = "accountList"
		P.ParamJSONstringData = "CB,LA,PEA"
		P.ParamInfo = "Liste des comptes (séparer par des , sans espaces)"
		InsertRowInParam(ctx, db, &P)
	}

	err = db.QueryRowContext(ctx, q, gofiID, "onboardingCheckList").Scan(&nbRows)
	switch {
	case err == sql.ErrNoRows:
		nbRows = 0
	case err != nil:
		log.Fatalf("query error param onboardingCheckList: %v\n", err)
	}
	if nbRows != 1 {
		db.ExecContext(ctx, "DELETE FROM param WHERE gofiID = ? AND paramName = 'onboardingCheckList';", gofiID)
		P.GofiID = gofiID
		P.ParamName = "onboardingCheckList"
		P.ParamJSONstringData = ""
		P.ParamInfo = "Liste des étapes (séparer par des , sans espaces)"
		InsertRowInParam(ctx, db, &P)
	}

	err = db.QueryRowContext(ctx, q, gofiID, "categoryRendering").Scan(&nbRows)
	switch {
	case err == sql.ErrNoRows:
		nbRows = 0
	case err != nil:
		log.Fatalf("query error param categoryRendering: %v\n", err)
	}
	if nbRows != 1 {
		db.ExecContext(ctx, "DELETE FROM param WHERE gofiID = ? AND paramName = 'categoryRendering';", gofiID)
		P.GofiID = gofiID
		P.ParamName = "categoryRendering"
		P.ParamJSONstringData = "icons"
		P.ParamInfo = "Affichage des catégories: icons | names"
		InsertRowInParam(ctx, db, &P)
	}
}

func InitCategoriesForUser(ctx context.Context, db *sql.DB, gofiID int) {
	var nbRows int = 0
	q := ` 
		SELECT COUNT(1)
		FROM category
		WHERE gofiID = ?;
	`
	err := db.QueryRowContext(ctx, q, gofiID).Scan(&nbRows)
	switch {
	case err == sql.ErrNoRows:
		nbRows = 0
	case err != nil:
		log.Fatalf("query error: %v\n", err)
		//default:
	}
	if nbRows == 0 {
		q1 := `
			INSERT INTO category (gofiID, category, catWhereToUse, catOrder, inUse, defaultInStats,
				iconName, iconCodePoint, colorName, colorHSL, colorHEX)
			VALUES 
				(?1, 'Besoin', 		'all', 		1, 1, 1, 'bed', 'e91f', 'needfix-teal', '(170,43,47)', '#44AA99'),
				(?1, 'Envie', 		'all', 		2, 1, 1, 'film', 'e920', 'wantko-wine', '(330,60,33)', '#882255'),
				(?1, 'Revenu', 		'periodic', 3, 1, 1, 'credit-card', 'e903', 'invest-cyan', '(200,75,73)', '#88CCEE'),
				(?1, 'Epargne', 	'all', 		4, 1, 0, 'line-chart', 'e904', 'invest-cyan', '(200,75,73)', '#88CCEE'),
				(?1, 'Habitude-', 	'all', 		5, 0, 1, 'thumbs-down', 'e91e', 'wantko-wine', '(330,60,33)', '#882255'),
				(?1, 'Vehicule', 	'all', 		6, 0, 1, 'car-front', 'e900', 'needvar-olive', '(60,50,40)', '#999933'),
				(?1, 'Transport', 	'all', 		7, 0, 1, 'train-front', 'e913', 'needvar-olive', '(60,50,40)', '#999933'),
				(?1, 'Shopping', 	'basic', 	8, 0, 1, 'shopping-cart', 'e918', 'wantko-wine', '(330,60,33)', '#882255'),
				(?1, 'Cadeaux', 	'basic', 	9, 0, 1, 'gift', 'e91a', 'wantko-wine', '(330,60,33)', '#882255'),
				(?1, 'Courses', 	'all', 		10, 0, 1, 'carrot', 'e916', 'needvar-olive', '(60,50,40)', '#999933'),
				(?1, 'Resto', 		'basic', 	11, 0, 1, 'chef-hat', 'e914', 'wantok-purple', '(310,43,47)', '#AA4499'),
				(?1, 'Loisirs', 	'all', 		12, 0, 1, 'drama', 'e901', 'wantok-purple', '(310,43,47)', '#AA4499'),
				(?1, 'Voyage', 		'basic', 	13, 0, 1, 'earth', 'e902', 'wantok-purple', '(310,43,47)', '#AA4499'),
				(?1, 'Enfants', 	'all', 		14, 0, 1, 'baby', 'e91d', 'needvar-olive', '(60,50,40)', '#999933'),
				(?1, 'Banque', 		'all', 		15, 0, 0, 'landmark', 'e919', 'needvar-olive', '(60,50,40)', '#999933'),
				(?1, 'Societe', 	'all', 		16, 0, 1, 'briefcase', 'e905', 'invest-cyan', '(200,75,73)', '#88CCEE'),
				(?1, 'Loyer', 		'periodic', 17, 0, 1, 'home', 'e906', 'needfix-teal', '(170,43,47)', '#44AA99'),
				(?1, 'Services', 	'periodic', 18, 0, 1, 'plug-zap', 'e907', 'needfix-teal', '(170,43,47)', '#44AA99'),
				(?1, 'Sante', 		'all', 		19, 0, 1, 'heart-pulse', 'e908', 'needvar-olive', '(60,50,40)', '#999933'),
				(?1, 'Animaux', 	'all', 		20, 0, 1, 'paw-print', 'e91c', 'wantko-wine', '(330,60,33)', '#882255')
			;`
		q2 := `
			INSERT INTO category (gofiID, category, catWhereToUse, catOrder, inUse, defaultInStats,
				description,
				iconName, iconCodePoint, colorName, colorHSL, colorHEX)
			VALUES
				(?1, 'Autre', 		'basic', 	21, 0, 1, 
					'Permet de ranger un élément qu''on ne sait pas où placer, temporairement ou définitivement.',
					'more-horizontal', 'e90c', 'system-lightgrey', '(0,0,87)', '#DDDDDD'),
				(?1, 'Erreur', 		'basic', 	22, 0, 1, 
					'Utile lorsqu''on souhaite corriger un montant global sans savoir réellement quel était l''achat en question.',
					'bug', 'e909', 'system-lightgrey', '(0,0,87)', '#DDDDDD'),
				(?1, 'Pret', 	'specific', -2, 1, 0, 
					'Utilisable uniquement par le système lors de l''utilisation de la fonction prêt.',
					'lend-hand-coin', 'e921', 'system-lightgrey', '(0,0,87)', '#DDDDDD'),
				(?1, 'Emprunt', 	'specific', -1, 1, 0, 
					'Utilisable uniquement par le système lors de l''utilisation de la fonction emprunt.',
					'borrow-hand-coin', 'e922', 'system-lightgrey', '(0,0,87)', '#DDDDDD'),
				(?1, 'Transfert', 	'specific', 97, 1, 0, 
					'Utilisé uniquement par le système lors de l''utilisation de la fonction transfert.',
					'arrow-right-left', 'e91b', 'system-lightgrey', '(0,0,87)', '#DDDDDD'),
				(?1, '?', 			'specific', 98, 1, 0, 
					'Utilisé uniquement comme icône par le système lorsqu''aucune icône ne correspond à la catégorie demandée.',
					'help-circle', 'e90a', 'system-lightgrey', '(0,0,87)', '#DDDDDD'),
				(?1, '-', 			'specific', 99, 1, 0, 
					'Utilisé uniquement par le système lorsqu''on supprime une ligne.',
					'trash-2', 'e90b', 'system-lightgrey', '(0,0,87)', '#DDDDDD')
			;`
		result, err := db.ExecContext(ctx, q1, gofiID)
		if err != nil {
			fmt.Printf("error1 on InitCategoriesForUser err: %#v\n", err)
			log.Fatalf("InitCategoriesForUser query error1: %v\n", err)
		}
		rowsAffected, err := result.RowsAffected()
		switch {
		case err == sql.ErrNoRows:
			rowsAffected = 0
		case err != nil:
			log.Fatalf("InitCategoriesForUser query error2: %v\n", err)
			//default:
		}
		if rowsAffected != 20 {
			log.Fatalf("InitCategoriesForUser query error3: %v\n", err)
		}
		result, err = db.ExecContext(ctx, q2, gofiID)
		if err != nil {
			fmt.Printf("error4 on InitCategoriesForUser err: %#v\n", err)
			log.Fatalf("InitCategoriesForUser query error4: %v\n", err)
		}
		rowsAffected, err = result.RowsAffected()
		switch {
		case err == sql.ErrNoRows:
			rowsAffected = 0
		case err != nil:
			log.Fatalf("InitCategoriesForUser query error5: %v\n", err)
			//default:
		}
		if rowsAffected != 7 {
			log.Fatalf("InitCategoriesForUser query error6 rowsAffected: %v\n", rowsAffected)
		}
	}
}

func GetFullCategoryList(ctx context.Context, db *sql.DB, uc *appdata.UserCategories, filterName string, filterValue any, firstEmptyCategory bool) {
	q := ` 
		SELECT id, gofiID, category, catWhereToUse, catOrder, inUse, defaultInStats, description, 
			budgetPrice, budgetPeriod, budgetType, budgetCurrentPeriodStartDate, iconCodePoint, colorHEX, colorName
		FROM category
		WHERE gofiID = ?
			AND OTHER FILTERS
		ORDER BY inUse DESC, catOrder, id
	`
	var err error
	var rows *sql.Rows
	switch filterName {
	case "":
		q = strings.Replace(q, `AND OTHER FILTERS`,
			` `, 1)
		rows, err = db.QueryContext(ctx, q, uc.GofiID)
	case "allinuse":
		q = strings.Replace(q, `OTHER FILTERS`,
			` inUse = 1 `, 1)
		 rows, err = db.QueryContext(ctx, q, uc.GofiID, filterValue)
	case "type":
		q = strings.Replace(q, `OTHER FILTERS`,
			` catWhereToUse IN ('all', ?) 
			 AND inUse = 1 `, 1)
		rows, err = db.QueryContext(ctx, q, uc.GofiID, filterValue)
	case "stats":
		q = strings.Replace(q, `OTHER FILTERS`,
			` defaultInStats = 1 `, 1)
		rows, err = db.QueryContext(ctx, q, uc.GofiID)
	case "budget":
		q = strings.Replace(q, `OTHER FILTERS`,
			` budgetPrice <> 0 
			 AND inUse = 1 `, 1)
		rows, err = db.QueryContext(ctx, q, uc.GofiID)
	case "lendborrow":
		q = strings.Replace(q, `OTHER FILTERS`,
			` catWhereToUse IN ('all', 'basic')
			AND inUse = 1
			OR (
				gofiID = ? 
				AND category IN ('Pret', 'Emprunt')) `, 1)
		rows, err = db.QueryContext(ctx, q, uc.GofiID, uc.GofiID)
	}
	if err != nil {
		fmt.Printf("error in GetFullCategoryList QueryContext: %v\n", err)
		log.Fatal(err)
		return
	}
	loop := -1
	if firstEmptyCategory {
		loop += 1
		var firstCategory appdata.Category
		firstCategory.GofiID = uc.GofiID
		firstCategory.Name = "Toutes"
		firstCategory.Type = "all"
		firstCategory.Order = 0
		firstCategory.InUse = 1
		firstCategory.IconCodePoint = "e90a"
		uc.FindCategory[firstCategory.Name] = loop
		uc.Categories = append(uc.Categories, firstCategory)
	}
	for rows.Next() {
		loop += 1
		var category appdata.Category
		err := rows.Scan(&category.ID, &category.GofiID, &category.Name, &category.Type, &category.Order, &category.InUse, &category.InStats, &category.Description,
			&category.BudgetPrice, &category.BudgetPeriod, &category.BudgetType, &category.BudgetCurrentPeriodStartDate, 
			&category.IconCodePoint, &category.ColorHEX, &category.ColorName)
		if err != nil {
			fmt.Printf("error in category loop: %v, category: %v\n", loop, category.Name)
			log.Fatal(err)
			return
		}
		uc.FindCategory[category.Name] = loop
		uc.Categories = append(uc.Categories, category)
	}
	rows.Close()
}

func GetUnhandledCategoryList(ctx context.Context, db *sql.DB, gofiID int) []string {
	var unhandledCategoryList []string
	q := ` 
		SELECT DISTINCT fT.category
		FROM financeTracker AS fT
			LEFT JOIN category AS c ON c.category = fT.category AND c.gofiID = fT.gofiID
		WHERE fT.gofiID = ?
			AND c.category IS NULL
			AND fT.category NOT IN (SELECT category FROM category WHERE gofiID = 0)
		ORDER BY fT.category
	`
	rows, err := db.QueryContext(ctx, q, gofiID)
	if err != nil {
		fmt.Printf("error1 in GetUnhandledCategoryList QueryContext: %v\n", err)
		return unhandledCategoryList
	}
	for rows.Next() {
		var category string
		err := rows.Scan(&category)
		if err != nil {
			fmt.Printf("error2 in GetUnhandledCategoryList category: %v\n", err)
			return unhandledCategoryList
		}
		unhandledCategoryList = append(unhandledCategoryList, category)
	}
	rows.Close()
	return unhandledCategoryList
}

func GetCategoryIcon(ctx context.Context, db *sql.DB, categoryName string, gofiID int) (string, string, string) {
	q := ` 
		SELECT iconCodePoint, colorHEX, colorName
		FROM category
		WHERE category = ?
			AND gofiID = ?;
	`
	var iconCodePoint, colorHEX, colorName string
	err := db.QueryRowContext(ctx, q, categoryName, gofiID).Scan(&iconCodePoint, &colorHEX, &colorName)
	switch {
	case err == sql.ErrNoRows:
		fmt.Printf("GetCategoryIcon error1 no row returned, category: %v\n", categoryName)
		return "", "", ""
	case err != nil:
		fmt.Printf("GetCategoryIcon error2: %v\n", err)
		return "", "", ""
	default:
		return iconCodePoint, colorHEX, colorName
	}
}

func GetList(ctx context.Context, db *sql.DB, up *appdata.UserParams, uc *appdata.UserCategories,
	categoryTypeFilter string, categoryTypeFilterValue string, firstEmptyCategory bool) {
	q := ` 
		SELECT paramJSONstringData
		FROM param
		WHERE gofiID = ?
			AND paramName = ?;
	`
	paramList := [3]string{"accountList", "onboardingCheckList", "categoryRendering"}
	var paramListResult []string
	for i := 0; i < len(paramList); i++ {
		var param, result string 
		param = paramList[i]
		rows, _ := db.QueryContext(ctx, q, up.GofiID, param)
		rows.Next()
		if err := rows.Scan(&result); err != nil {
			fmt.Printf("error in GetList %v, err: %v\n", param, err)
			log.Fatal(err)
		}
		rows.Close()
		paramListResult = append(paramListResult, result)
	}
	up.AccountListSingleString = paramListResult[0]
	up.AccountList = strings.Split(up.AccountListSingleString, ",")
	up.OnboardingCheckListSingleString = paramListResult[1]
	up.OnboardingCheckList = strings.Split(up.OnboardingCheckListSingleString, ",")
	up.CategoryRendering = paramListResult[2]

	GetFullCategoryList(ctx, db, uc, categoryTypeFilter, categoryTypeFilterValue, firstEmptyCategory)
}

func PutCategory(ctx context.Context, db *sql.DB, category *appdata.CategoryPut) bool {
	q := ` 
		UPDATE category 
		SET catWhereToUse = ?, defaultInStats = ?, description = ?,
			budgetPrice = ?, budgetPeriod = ?, budgetType = ?, budgetCurrentPeriodStartDate = ?
		WHERE id = ?
			AND gofiID = ?
	`
	result, err := db.ExecContext(ctx, q, category.Type, category.InStats, category.Description,
		category.BudgetPrice, category.BudgetPeriod, category.BudgetType, category.BudgetCurrentPeriodStartDate,
		category.ID, category.GofiID)
	if err != nil {
		fmt.Printf("error1 PutCategory categoryID: %v, gofiID: %v, err: %#v\n", category.ID, category.GofiID, err)
		return false
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected != 1 {
		fmt.Printf("error2 PutCategory categoryID: %v, gofiID: %v, rowsAffected: %v\n", category.ID, category.GofiID, rowsAffected)
		return false
	}
	return true
}

func PatchCategoryInUse(ctx context.Context, db *sql.DB, category *appdata.CategoryPatchInUse) bool {
	q := ` 
		UPDATE category 
		SET inUse = ?
		WHERE id = ?
			AND gofiID = ?
	`
	result, err := db.ExecContext(ctx, q, category.InUse, category.ID, category.GofiID)
	if err != nil {
		fmt.Printf("error1 on UPDATE PatchCategoryInUse categoryID: %v, gofiID: %v, err: %#v\n", category.ID, category.GofiID, err)
		return false
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected != 1 {
		fmt.Printf("error2 on UPDATE PatchCategoryInUse categoryID: %v, gofiID: %v, rowsAffected: %v\n", category.ID, category.GofiID, rowsAffected)
		return false
	}
	return true
}

func PatchCategoryOrder(ctx context.Context, db *sql.DB, categoriesSwitch *appdata.CategoryPatchOrder) bool {
	// need to change the order of 2 categories: +1 for one and -1 for the other
	// switch the order between 2 categories
	var categoryOrder1, categoryOrder2 int = 0, 0
	q1 := `
		SELECT catOrder
		FROM category
		WHERE gofiID = ?
			AND id = ?;
	`
	err := db.QueryRowContext(ctx, q1, categoriesSwitch.GofiID, categoriesSwitch.ID1).Scan(&categoryOrder1)
	switch {
	case err == sql.ErrNoRows:
		fmt.Printf("PatchCategoryOrder query error1: %v\n", err)
		return false
	case err != nil:
		fmt.Printf("PatchCategoryOrder query error2: %v\n", err)
		return false
	}
	err = db.QueryRowContext(ctx, q1, categoriesSwitch.GofiID, categoriesSwitch.ID2).Scan(&categoryOrder2)
	switch {
	case err == sql.ErrNoRows:
		fmt.Printf("PatchCategoryOrder query error3: %v\n", err)
		return false
	case err != nil:
		fmt.Printf("PatchCategoryOrder query error4: %v\n", err)
		return false
	}

	q2 := ` 
		UPDATE category 
		SET catOrder = ? 
		WHERE gofiID = ?
			AND id = ?;
	`
	result, err := db.ExecContext(ctx, q2, categoryOrder2, categoriesSwitch.GofiID, categoriesSwitch.ID1)
	if err != nil {
		fmt.Printf("error5 PatchCategoryOrder categoryID: %v, gofiID: %v, err: %#v\n", categoriesSwitch.ID1, categoriesSwitch.GofiID, err)
		return false
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected != 1 {
		fmt.Printf("error6 PatchCategoryOrder categoryID: %v, gofiID: %v, rowsAffected: %v\n", categoriesSwitch.ID1, categoriesSwitch.GofiID, rowsAffected)
		return false
	}

	result, err = db.ExecContext(ctx, q2, categoryOrder1, categoriesSwitch.GofiID, categoriesSwitch.ID2)
	if err != nil {
		fmt.Printf("error7 PatchCategoryOrder categoryID: %v, gofiID: %v, err: %#v\n", categoriesSwitch.ID2, categoriesSwitch.GofiID, err)
		return false
	}
	rowsAffected, err = result.RowsAffected()
	if err != nil || rowsAffected != 1 {
		fmt.Printf("error8 PatchCategoryOrder categoryID: %v, gofiID: %v, rowsAffected: %v\n", categoriesSwitch.ID2, categoriesSwitch.GofiID, rowsAffected)
		return false
	}

	return true
}

func GetUnhandledAccountList(ctx context.Context, db *sql.DB, gofiID int, accountList []string) []string {
	var accountListUnhandled []string
	q := ` 
		SELECT DISTINCT fT.account
		FROM financeTracker AS fT
		WHERE fT.gofiID = ?
			AND fT.account NOT IN ('-')
		ORDER BY fT.account
	`
	rows, err := db.QueryContext(ctx, q, gofiID)
	if err != nil {
		fmt.Printf("error1 in GetUnhandledAccountList QueryContext: %v\n", err)
		return accountListUnhandled
	}
	for rows.Next() {
		var account string
		err := rows.Scan(&account)
		if err != nil {
			fmt.Printf("error2 in GetUnhandledAccountList account: %v\n", err)
			return accountListUnhandled
		}
		if !slices.Contains(accountList, account) {
			accountListUnhandled = append(accountListUnhandled, account)
		}
	}
	rows.Close()
	return accountListUnhandled
}