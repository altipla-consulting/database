package database

import (
	"os"
)

func isDebug() bool {
	return os.Getenv("DEBUG") == "true" || os.Getenv("DEBUG") == "1"
}
