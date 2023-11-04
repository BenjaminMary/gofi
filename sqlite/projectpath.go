package sqlite

import (
    "path/filepath"
	"os"
    "runtime"
    "path"
)

var (
    // this method only support the call of the variable from the same package
    // rootDirectory, _ = os.Getwd()
    // DbPath = filepath.Join(rootDirectory, "sqlite", os.Getenv("SQLITE_DB_FILENAME"))


    // this method also handle the call of the variable from other packages
    _, currentFilePath, _, _ = runtime.Caller(0)
    Dirpath = path.Dir(currentFilePath)
    DbPath = filepath.Join(Dirpath, os.Getenv("SQLITE_DB_FILENAME"))

)

func FilePath(fileName string) string {
	return filepath.Join(Dirpath, fileName)
}