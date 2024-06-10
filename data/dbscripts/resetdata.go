package main

import (
	"database/sql"
	"fmt"

	"gofi/gofi/data/appdata"

	_ "modernc.org/sqlite"
)

func main() {
	var err error
	var db *sql.DB
	fmt.Println(appdata.DbPath)
	if len(appdata.DbPath) < 4 {
		panic("specify the SQLITE_DB_FILENAME environment variable")
	}
	db, err = sql.Open("sqlite", appdata.DbPath)
	if err != nil {
		panic("error opening the SQLite file")
	}
	err = db.Ping()
	if err != nil {
		panic("can't ping DB")
	}

	fmt.Println("cleaning user table")
	_, err = db.Exec(`
		DELETE FROM user;
		DELETE FROM SQLITE_SEQUENCE WHERE name='user';
		`,
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("cleaning param table")
	_, err = db.Exec(`
		DELETE FROM param;
		DELETE FROM SQLITE_SEQUENCE WHERE name='param';
		`,
	)
	if err != nil {
		panic(err)
	}
}
