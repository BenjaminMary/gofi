package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"gofi/gofi/data/appdata"
	"log"
	"strings"
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
		log.Fatalf("query error: %v\n", err)
		//default:
	}
	if nbRows != 1 {
		db.ExecContext(ctx, "DELETE FROM param WHERE gofiID = ? AND paramName = 'accountList';", gofiID)
		P.GofiID = gofiID
		P.ParamName = "accountList"
		P.ParamJSONstringData = "CB,A"
		P.ParamInfo = "Liste des comptes (séparer par des , sans espaces)"
		InsertRowInParam(ctx, db, &P)
	}

	err = db.QueryRowContext(ctx, q, gofiID, "categoryList").Scan(&nbRows)
	switch {
	case err == sql.ErrNoRows:
		nbRows = 0
	case err != nil:
		log.Fatalf("query error param categoryList: %v\n", err)
	}
	if nbRows != 1 {
		db.ExecContext(ctx, "DELETE FROM param WHERE gofiID = ? AND paramName = 'categoryList';", gofiID)
		P.GofiID = gofiID
		P.ParamName = "categoryList"
		P.ParamJSONstringData = "Courses,Banque,Cadeaux,Entrep,Erreur,Epargne,Loisirs,Loyer,Resto,Revenu,Sante,Services,Shopping,Transport,Voyage,Vehicule,Autre"
		P.ParamInfo = "Liste des catégories (séparer par des , sans espaces)"
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
		q := `
			INSERT INTO category (gofiID, category, catWhereToUse, catOrder, inUse,
				iconName, iconCodePoint, colorName, colorHSL, colorHEX)
			VALUES 
				(?1, 'Besoin', 		'all', 		1, 1, 'bed', 'e91f', 'green', '(130,60,50)', '#33CC4C'),
				(?1, 'Envie', 		'all', 		2, 1, 'film', 'e920', 'orange', '(30,60,50)', '#CC8033'),
				(?1, 'Revenu', 		'periodic', 3, 1, 'credit-card', 'e903', 'teal', '(160,60,50)', '#33CC99'),
				(?1, 'Epargne', 	'all', 		4, 1, 'line-chart', 'e904', 'light blue', '(210,60,50)', '#3380CC'),
				(?1, 'Habitude-', 	'all', 		5, 0, 'thumbs-down', 'e91e', 'red', '(1,60,50)', '#CC3633'),
				(?1, 'Vehicule', 	'all', 		6, 0, 'car-front', 'e900', 'orange', '(15,60,50)', '#CC5933'),
				(?1, 'Transport', 	'all', 		7, 0, 'train-front', 'e913', 'orange', '(30,60,50)', '#CC8033'),
				(?1, 'Shopping', 	'basic', 	8, 0, 'shopping-cart', 'e918', 'yellow', '(45,40,50)', '#B3994D'),
				(?1, 'Cadeaux', 	'basic', 	9, 0, 'gift', 'e91a', 'yellow', '(60,40,50)', '#B3B34D'),
				(?1, 'Courses', 	'all', 		10, 0, 'carrot', 'e916', 'yellow', '(70,50,50)', '#AABF40'),
				(?1, 'Resto', 		'basic', 	11, 0, 'chef-hat', 'e914', 'green', '(90,60,50)', '#80CC33'),
				(?1, 'Loisirs', 	'all', 		12, 0, 'drama', 'e901', 'green', '(110,60,50)', '#4DCC33'),
				(?1, 'Voyage', 		'basic', 	13, 0, 'earth', 'e902', 'green', '(130,60,50)', '#33CC4C'),
				(?1, 'Enfants', 	'all', 		14, 0, 'baby', 'e91d', 'teal', '(175,60,50)', '#33CCBF'),
				(?1, 'Banque', 		'all', 		15, 0, 'landmark', 'e919', 'light blue', '(190,60,50)', '#33B3CC'),
				(?1, 'Societe', 	'all', 		16, 0, 'briefcase', 'e905', 'blue', '(230,60,50)', '#334CCC'),
				(?1, 'Loyer', 		'periodic', 17, 0, 'home', 'e906', 'purple', '(260,60,50)', '#6633CC'),
				(?1, 'Services', 	'periodic', 18, 0, 'plug-zap', 'e907', 'purple', '(270,60,50)', '#8033CC'),
				(?1, 'Sante', 		'all', 		19, 0, 'heart-pulse', 'e908', 'pink', '(300,60,50)', '#CC33CC'),
				(?1, 'Animaux', 	'all', 		20, 0, 'paw-print', 'e91c', 'pink', '(320,60,50)', '#CC3399')
			;
		`
		result, err := db.ExecContext(ctx, q, gofiID)
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
	}
}

func GetCategoryList(ctx context.Context, db *sql.DB) ([]string, []string, []string) {
	q := ` 
		SELECT category, iconCodePoint, colorHEX
		FROM category
		ORDER BY id
	`
	rows, _ := db.QueryContext(ctx, q)

	var category, iconCodePoint, colorHEX string
	var categoryList, iconCodePointList, colorHEXList []string
	for rows.Next() {
		if err := rows.Scan(&category, &iconCodePoint, &colorHEX); err != nil {
			log.Fatal(err)
			return categoryList, iconCodePointList, colorHEXList
		}
		categoryList = append(categoryList, category)
		iconCodePointList = append(iconCodePointList, iconCodePoint)
		colorHEXList = append(colorHEXList, colorHEX)
	}
	// fmt.Printf("\naccountList: %v\n", up.AccountList)
	rows.Close()
	return categoryList, iconCodePointList, colorHEXList
}

func GetFullCategoryList(ctx context.Context, db *sql.DB, uc *appdata.UserCategories) {
	q := ` 
		SELECT id, gofiID, category, catWhereToUse, catOrder, inUse, defaultInStats, description, 
			budgetPrice, budgetPeriod, budgetType, iconCodePoint, colorHEX
		FROM category
		WHERE gofiID IN (0, ?)
		ORDER BY inUse DESC, catOrder, id
	`
	rows, err := db.QueryContext(ctx, q, uc.GofiID)
	if err != nil {
		fmt.Printf("error in GetFullCategoryList QueryContext: %v\n", err)
		log.Fatal(err)
		return
	}

	loop := -1
	for rows.Next() {
		loop += 1
		var category appdata.Category
		err := rows.Scan(&category.ID, &category.GofiID, &category.Name, &category.Type, &category.Order, &category.InUse, &category.InStats, &category.Description,
			&category.BudgetPrice, &category.BudgetPeriod, &category.BudgetType, &category.IconCodePoint, &category.ColorHEX)
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

func GetCategoryIcon(ctx context.Context, db *sql.DB, categoryName string) (string, string) {
	q := ` 
		SELECT iconCodePoint, colorHEX
		FROM category
		WHERE category = ?
			AND gofiID = ?;
	`
	var iconCodePoint, colorHEX string
	err := db.QueryRowContext(ctx, q, categoryName).Scan(&iconCodePoint, &colorHEX)
	switch {
	case err == sql.ErrNoRows:
		fmt.Printf("GetCategoryIcon error no row returned, category: %v\n", categoryName)
		return "", ""
	case err != nil:
		fmt.Printf("GetCategoryIcon error: %v\n", err)
		return "", ""
	default:
		return iconCodePoint, colorHEX
	}
}

func GetList(ctx context.Context, db *sql.DB, up *appdata.UserParams) {
	q := ` 
		SELECT paramJSONstringData
		FROM param
		WHERE gofiID = ?
			AND paramName = ?;
	`
	rows, _ := db.QueryContext(ctx, q, up.GofiID, "accountList")
	rows.Next()
	var accountList string
	if err := rows.Scan(&accountList); err != nil {
		fmt.Printf("error in GetList accountList, err: %v\n", err)
		log.Fatal(err)
	}
	rows.Close()
	up.AccountListSingleString = accountList
	up.AccountList = strings.Split(accountList, ",")

	rows, _ = db.QueryContext(ctx, q, up.GofiID, "categoryList")
	rows.Next()
	var categoryListStr string
	if err := rows.Scan(&categoryListStr); err != nil {
		fmt.Printf("error in GetList categoryList, err: %v\n", err)
		log.Fatal(err)
	}
	rows.Close()
	up.CategoryListSingleString = categoryListStr

	var categoryListA, categoryListB, iconCodePointList, colorHEXList []string
	categoryListA = strings.Split(categoryListStr, ",")
	categoryListB, iconCodePointList, colorHEXList = GetCategoryList(ctx, db)
	for i, v := range categoryListA {
		var found bool = false
		var stringToAppend []string
		if i < len(categoryListB) {
			for i2, v2 := range categoryListB {
				if v == v2 {
					stringToAppend = append(stringToAppend, v, iconCodePointList[i2], colorHEXList[i2])
					found = true
				}
			}
		}
		if !found {
			stringToAppend = append(stringToAppend, v, "e90a", "#808080")
		}
		up.CategoryList = append(up.CategoryList, stringToAppend)
	}

	rows, _ = db.QueryContext(ctx, q, up.GofiID, "categoryRendering")
	rows.Next()
	var categoryRendering string
	if err := rows.Scan(&categoryRendering); err != nil {
		fmt.Printf("error in GetList categoryRendering, err: %v\n", err)
		log.Fatal(err)
	}
	rows.Close()
	up.CategoryRendering = categoryRendering
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
		category.BudgetPrice, category.BudgetPrice, category.BudgetType, category.BudgetCurrentPeriodStartDate,
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
