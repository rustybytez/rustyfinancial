-- name: CreateSnapshot :exec
INSERT INTO account_snapshots (account_id, date, balance)
VALUES (?, ?, ?)
ON CONFLICT(account_id, date) DO UPDATE SET balance = excluded.balance;

-- name: HasSnapshotForDate :one
SELECT COUNT(*) FROM account_snapshots WHERE date = ?;

-- name: ListAllSnapshots :many
SELECT * FROM account_snapshots ORDER BY date DESC, account_id;
