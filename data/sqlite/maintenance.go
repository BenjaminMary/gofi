package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"gofi/gofi/data/appdata"
	"log"
	"time"
)

// func OpenDbCon() *sql.DB {
// 	db, err := sql.Open("sqlite", appdata.DbPath)
// 	if err != nil {
// 		log.Fatal("error opening DB file: ", err)
// 	}
// 	db.SetMaxIdleConns(1) //default 2
// 	db.SetMaxOpenConns(3) //default 0 = infinite
// 	return db
// }

// func WalCheckpointB() {
// 	sql.Open("sqlite", "file:///gofi.db?_pragma=foreign_keys(1)&_time_format=sqlite")
// 	// sqlite.Open("file:///tmp/mydata.sqlite?_pragma=foreign_keys(1)&_time_format=sqlite")
// }

func WalCheckpoint(ctx context.Context) int {
	db, err := sql.Open("sqlite", appdata.DbPath)
	if err != nil {
		log.Fatal("error opening DB file: ", err)
		return -1
	}
	defer db.Close()
	defer fmt.Println("defer : db.Close()")
	db.SetMaxIdleConns(1) //default 2
	db.SetMaxOpenConns(1) //default 0 = infinite

	conn, err := db.Conn(ctx)
	if err != nil {
		log.Fatal("error connecting to DB file: ", err)
		return -1
	}
	defer conn.Close() // Return the connection to the pool.
	defer fmt.Println("defer : conn.Close()")

	// fmt.Println("optimize, vacuum, checkpoint TRUNCATE then close DB")
	fmt.Println("optimize, vacuum, checkpoint TRUNCATE")
	conn.ExecContext(ctx, "PRAGMA optimize;") // to run just before closing each database connection.

	var journalMode string
	err = conn.QueryRowContext(ctx, "PRAGMA journal_mode;").Scan(&journalMode)
	if err != nil {
		log.Fatal("error PRAGMA journal_mode: ", err)
		return -1
	}
	fmt.Printf("journalMode: %v\n", journalMode)

	conn.ExecContext(ctx, "VACUUM;") // to run just before closing each database connection.

	var busyTimeout string
	err = conn.QueryRowContext(ctx, "PRAGMA busy_timeout;").Scan(&busyTimeout)
	if err != nil {
		log.Fatal("error PRAGMA busyTimeout 1: ", err)
		return -1
	}
	//fmt.Printf("busyTimeout 1: %v\n", busyTimeout)
	err = conn.QueryRowContext(ctx, "PRAGMA busy_timeout = 2000;").Scan(&busyTimeout)
	if err != nil {
		log.Fatal("error PRAGMA busyTimeout 2: ", err)
		return -1
	}
	//fmt.Printf("busyTimeout 2: %v\n", busyTimeout)

	db.SetConnMaxIdleTime(100 * time.Millisecond)
	db.SetConnMaxLifetime(100 * time.Millisecond)
	time.Sleep(3 * time.Second)

	//stats := db.Stats()
	//fmt.Printf("stats: %#v\n", stats)

	conn.ExecContext(ctx, "COMMIT;")
	conn.Close()
	conn, err = db.Conn(ctx)
	if err != nil {
		log.Fatal("error connecting to DB file: ", err)
		return -1
	}

	// wal_checkpoint doc: https://www.sqlite.org/pragma.html#pragma_wal_checkpoint
	// checkpointReturn = 0 if OK, pagestoWal AND pagesFromWalToDb -1 if not in WAL mode
	var checkpointReturn, pagestoWal, pagesFromWalToDb int
	err = conn.QueryRowContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE);").Scan(&checkpointReturn, &pagestoWal, &pagesFromWalToDb)
	if err != nil {
		log.Fatal("error PRAGMA wal_checkpoint(TRUNCATE): ", err)
		return -1
	}
	//fmt.Printf("checkpointReturn: %v\n", strconv.Itoa(checkpointReturn))
	//fmt.Printf("pagestoWal: %v\n", strconv.Itoa(pagestoWal))
	//fmt.Printf("pagesFromWalToDb: %v\n", strconv.Itoa(pagesFromWalToDb))
	if checkpointReturn == 1 {
		// conn.Close()
		return 1
	}
	conn.Close()
	db.Close()
	time.Sleep(1 * time.Second)

	return checkpointReturn
}
