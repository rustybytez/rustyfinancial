package store

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"time"

	"rustyfinancial/internal/db"

	_ "modernc.org/sqlite"
)

// migrations is an ordered list of SQL migration blocks. Each block is applied
// exactly once and tracked in the schema_migrations table by its 1-based index.
// Statements within a block are separated by semicolons.
var migrations = []string{
	// v1: initial schema
	`CREATE TABLE IF NOT EXISTS accounts (
		id         INTEGER  PRIMARY KEY AUTOINCREMENT,
		name       TEXT     NOT NULL,
		type       TEXT     NOT NULL CHECK(type IN ('retirement', 'investment', 'cash', 'debt', 'other')),
		currency   TEXT     NOT NULL DEFAULT 'USD',
		created_at DATETIME NOT NULL DEFAULT (datetime('now'))
	);
	CREATE TABLE IF NOT EXISTS transactions (
		id          INTEGER  PRIMARY KEY AUTOINCREMENT,
		account_id  INTEGER  NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
		amount      INTEGER  NOT NULL,
		description TEXT     NOT NULL DEFAULT '',
		date        DATE     NOT NULL,
		created_at  DATETIME NOT NULL DEFAULT (datetime('now'))
	)`,

	// v2: add institution field, migrate retired/debt types, update CHECK constraint
	`ALTER TABLE accounts ADD COLUMN institution TEXT NOT NULL DEFAULT '';
	UPDATE accounts SET type = 'investment' WHERE type = 'retirement';
	UPDATE accounts SET type = 'liability' WHERE type = 'debt';
	CREATE TABLE accounts_v2 (
		id          INTEGER  PRIMARY KEY AUTOINCREMENT,
		name        TEXT     NOT NULL,
		type        TEXT     NOT NULL CHECK(type IN ('cash', 'investment', 'real_estate', 'liability', 'other')),
		currency    TEXT     NOT NULL DEFAULT 'USD',
		institution TEXT     NOT NULL DEFAULT '',
		created_at  DATETIME NOT NULL DEFAULT (datetime('now'))
	);
	INSERT INTO accounts_v2 (id, name, type, currency, institution, created_at)
		SELECT id, name, type, currency, institution, created_at FROM accounts;
	DROP TABLE accounts;
	ALTER TABLE accounts_v2 RENAME TO accounts`,

	// v3: add manual_balance field (nullable, overrides transaction sum when set)
	`ALTER TABLE accounts ADD COLUMN manual_balance INTEGER`,

	// v4: monthly balance snapshots
	`CREATE TABLE IF NOT EXISTS account_snapshots (
		id         INTEGER  PRIMARY KEY AUTOINCREMENT,
		account_id INTEGER  NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
		date       DATE     NOT NULL,
		balance    INTEGER  NOT NULL,
		created_at DATETIME NOT NULL DEFAULT (datetime('now')),
		UNIQUE(account_id, date)
	)`,
}

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
	if _, err := s.DB.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (id INTEGER PRIMARY KEY)`); err != nil {
		return err
	}
	for i, migration := range migrations {
		version := i + 1
		var n int
		s.DB.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE id = ?`, version).Scan(&n) //nolint:errcheck
		if n > 0 {
			continue
		}
		for _, stmt := range strings.Split(migration, ";") {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if _, err := s.DB.Exec(stmt); err != nil {
				return err
			}
		}
		if _, err := s.DB.Exec(`INSERT INTO schema_migrations (id) VALUES (?)`, version); err != nil {
			return err
		}
	}
	return nil
}

// MaybeSnapshot takes a snapshot only if today is the 1st and none exists yet.
func (s *Store) MaybeSnapshot(ctx context.Context) error {
	if time.Now().Day() != 1 {
		return nil
	}
	return s.TakeSnapshot(ctx)
}

// TakeSnapshot records current balances for all accounts, keyed to the 1st of
// the current month. Safe to call multiple times — overwrites any existing row.
func (s *Store) TakeSnapshot(ctx context.Context) error {
	now := time.Now()
	snapshotDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	accounts, err := s.ListAccounts(ctx)
	if err != nil {
		return err
	}

	for _, a := range accounts {
		var bal int64
		if a.ManualBalance.Valid {
			bal = a.ManualBalance.Int64
		} else {
			raw, err := s.GetAccountBalance(ctx, a.ID)
			if err != nil {
				return err
			}
			bal = balanceToInt64(raw)
		}
		if err := s.CreateSnapshot(ctx, db.CreateSnapshotParams{
			AccountID: a.ID,
			Date:      snapshotDate,
			Balance:   bal,
		}); err != nil {
			return err
		}
	}

	log.Printf("snapshot: recorded %d accounts for %s", len(accounts), snapshotDate.Format("2006-01-02"))
	return nil
}

// StartSnapshotWorker spawns a goroutine that calls MaybeSnapshot once at
// startup and then again every day just after midnight.
func (s *Store) StartSnapshotWorker(ctx context.Context) {
	go func() {
		if err := s.MaybeSnapshot(ctx); err != nil {
			log.Printf("snapshot: %v", err)
		}
		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 1, 0, 0, now.Location())
			select {
			case <-time.After(time.Until(next)):
				if err := s.MaybeSnapshot(ctx); err != nil {
					log.Printf("snapshot: %v", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

func balanceToInt64(v interface{}) int64 {
	if v == nil {
		return 0
	}
	switch n := v.(type) {
	case int64:
		return n
	case float64:
		return int64(n)
	}
	return 0
}
