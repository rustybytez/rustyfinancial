-- name: ListAccounts :many
SELECT * FROM accounts ORDER BY type, name;

-- name: GetAccount :one
SELECT * FROM accounts WHERE id = ?;

-- name: CreateAccount :one
INSERT INTO accounts (name, type, currency)
VALUES (?, ?, ?)
RETURNING *;

-- name: UpdateAccount :one
UPDATE accounts SET name = ?, type = ?, currency = ?
WHERE id = ?
RETURNING *;

-- name: DeleteAccount :exec
DELETE FROM accounts WHERE id = ?;

-- name: GetAccountBalance :one
SELECT COALESCE(SUM(amount), 0) AS balance
FROM transactions
WHERE account_id = ?;
