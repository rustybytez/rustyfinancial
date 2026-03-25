-- name: ListTransactionsByAccount :many
SELECT * FROM transactions WHERE account_id = ? ORDER BY date DESC, id DESC;

-- name: CreateTransaction :one
INSERT INTO transactions (account_id, amount, description, date)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: DeleteTransaction :exec
DELETE FROM transactions WHERE id = ?;
