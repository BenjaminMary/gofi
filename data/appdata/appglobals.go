package appdata

import (
	"crypto/rand"
	"database/sql"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
)

type ContextKey string

var (
	DB           *sql.DB // https://www.alexedwards.net/blog/organising-database-access
	ShutDownFlag = false

	executablePath = os.Getenv("EXE_PATH")
	DbDirPath      = filepath.Join(executablePath, "data", "dbFiles")
	DbPath         = filepath.Join(DbDirPath, os.Getenv("SQLITE_DB_FILENAME"))

	cookieLengthOs  = os.Getenv("COOKIE_LENGTH")
	CookieLength, _ = strconv.Atoi(cookieLengthOs)

	// ServerDirpath  = filepath.Join(executablePath, "back", "server")
)

const (
	ContextUserKey ContextKey = "userRequestCtx"
)

func SQLiteFilePath(fileName string) string {
	return filepath.Join(DbDirPath, fileName)
}

// FUNC generateRandomString returns a securely generated random string
func GenerateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_.~"
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}
	return string(ret), nil
}
