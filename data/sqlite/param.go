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
		P.ParamJSONstringData = "Courses,Banque,Cadeaux,Entrep,Erreur,Invest,Loisirs,Loyer,Resto,Salaire,Sante,Services,Shopping,Transp,Voyage,Vehicule,Autre"
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

func GetCategoryIcon(ctx context.Context, db *sql.DB, categoryName string) (string, string) {
	q := ` 
		SELECT iconCodePoint, colorHEX
		FROM category
		WHERE category = ?
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
