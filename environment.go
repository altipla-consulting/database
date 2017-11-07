package database

func isDebug() bool {
	return os.Getenv("DEBUG") == "true" || os.Getenv("DEBUG") == "1"
}
