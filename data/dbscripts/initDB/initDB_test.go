package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func deleteOldTestDB() {
	dbFilesPath := filepath.Join(os.Getenv("EXE_PATH"), "data", "dbFiles")
	os.Remove(filepath.Join(dbFilesPath, "test.db"))
	os.Remove(filepath.Join(dbFilesPath, "test.db-shm"))
	os.Remove(filepath.Join(dbFilesPath, "test.db-wal"))
}

func TestUser(t *testing.T) {
	testStartTime := time.Now()

	deleteOldTestDB()
	initDB("dbFiles", "test.db")

	// fmt.Printf("response: %#v\n", response.Body.String())
	// require.Equal(t, 1, 0, "force fail")
	require.WithinDuration(t, time.Now(), testStartTime, 1*time.Second)
}
