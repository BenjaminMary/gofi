package sqlite

import (
    "path/filepath"
	"os"
)

var (
    rootDirectory, _ = os.Getwd()
    DbPath = filepath.Join(rootDirectory, "sqlite", os.Getenv("SQLITE_DB_FILENAME"))
)

// Then use projectpath.Root in this package