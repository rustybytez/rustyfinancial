CREATE TABLE IF NOT EXISTS accounts (
    id          INTEGER  PRIMARY KEY AUTOINCREMENT,
    name        TEXT     NOT NULL,
    type        TEXT     NOT NULL CHECK(type IN ('cash', 'investment', 'real_estate', 'liability', 'other')),
    currency    TEXT     NOT NULL DEFAULT 'USD',
    institution    TEXT     NOT NULL DEFAULT '',
    manual_balance INTEGER,
    created_at     DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS account_snapshots (
    id         INTEGER  PRIMARY KEY AUTOINCREMENT,
    account_id INTEGER  NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    date       DATE     NOT NULL,
    balance    INTEGER  NOT NULL,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    UNIQUE(account_id, date)
);

CREATE TABLE IF NOT EXISTS transactions (
    id          INTEGER  PRIMARY KEY AUTOINCREMENT,
    account_id  INTEGER  NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    amount      INTEGER  NOT NULL, -- stored in cents; positive = credit, negative = debit
    description TEXT     NOT NULL DEFAULT '',
    date        DATE     NOT NULL,
    created_at  DATETIME NOT NULL DEFAULT (datetime('now'))
);
