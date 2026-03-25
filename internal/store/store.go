package store

import (
	"database/sql"
	"embed"
	"strings"

	"rustyfinancial/internal/db"

	_ "modernc.org/sqlite"
)

//go:embed migrations.sql
var migrationsFS embed.FS

type Store struct {
	*db.Queries
	DB *sql.DB
}

func New(dsn string) (*Store, error) {
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	conn.SetMaxOpenConns(1) // SQLite is single-writer
	s := &Store{DB: conn, Queries: db.New(conn)}
	if err := s.migrate(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) migrate() error {
	data, err := migrationsFS.ReadFile("migrations.sql")
	if err != nil {
		return err
	}
	for _, stmt := range strings.Split(string(data), ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := s.DB.Exec(stmt); err != nil {
			if strings.Contains(err.Error(), "duplicate column name") {
				continue
			}
			return err
		}
	}
	return nil
}
