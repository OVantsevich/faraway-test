package config

import (
	"fmt"
)

// SqliteMode - Access Mode of the database.
type SqliteMode string

const (
	// SQLITE_OPEN_MEMORY - The database will be opened as an in-memory database. The database is named by the "filename" argument for the purposes of cache-sharing, if shared cache mode is enabled, but the "filename" is otherwise ignored.
	SQLITE_OPEN_MEMORY SqliteMode = "memory"
	// SQLITE_OPEN_READONLY - The database is opened in read-only mode. If the database does not already exist, an error is returned.
	SQLITE_OPEN_READONLY SqliteMode = "ro"
	// SQLITE_OPEN_READWRITE SQLITE_OPEN_READWRITE - The database is opened for reading and writing if possible, or reading only if the file is write protected by the operating system.
	SQLITE_OPEN_READWRITE SqliteMode = "rw"
	// SQLITE_OPEN_CREATE - The database is opened for reading and writing, and is created if it does not already exist. This is the behavior that is always used for sqlite3_open() and sqlite3_open16().
	SQLITE_OPEN_CREATE SqliteMode = "rwc"
)

// Sqlite - config for Sqlite
type Sqlite struct {
	SqliteDirectory string     `env:"DB_DIR,notEmpty" envDefault:"db"`
	SqliteName      string     `env:"DB_NAME,notEmpty" envDefault:"test"`
	SQLiteMode      SqliteMode `env:"SQLITE_MODE,notEmpty" envDefault:"rwc"`
}

// SqliteConn - connection line to ent sqlite
func (s *Sqlite) SqliteConn() string {
	return fmt.Sprint("file:", s.SqliteDirectory, "//", s.SqliteName, ".db?",
		"mode=", s.SQLiteMode,
		"&cache=shared&_fk=1")
}
