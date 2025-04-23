-- name: GetAccount :one
SELECT *
FROM accounts
WHERE id = ? LIMIT 1;

-- name: CreateExpense :one
INSERT INTO expenses (amount, description, destination, date,
                      month, year, category_id, account_id)
VALUES (?, ?, ?, ?, ?, ?, ?, ?) RETURNING *;

-- name: GetExpenseWithAssociation :one
SELECT e.*
FROM expenses e
         JOIN categories c ON e.category_id = c.id
WHERE e.id = ?
  AND c.association_id = ? LIMIT 1;

-- name: UpdateExpense :one
UPDATE expenses
SET amount      = ?,
    description = ?,
    destination = ?,
    date        = ?,
    month       = ?,
    year        = ?,
    category_id = ?,
    account_id  = ?,
    updated_at  = CURRENT_TIMESTAMP
WHERE id = ? RETURNING *;

-- name: DeleteExpense :exec
DELETE
FROM expenses
WHERE id = ?;

-- name: GetExpensesByDateRange :many
SELECT e.id,
       e.amount,
       e.description,
       e.destination,
       e.date,
       e.month,
       e.year,
       e.category_id,
       e.account_id,
       c.type        as category_type,
       c.family      as category_family,
       c.name        as category_name,
       a.number      as account_number,
       a.description as account_name
FROM expenses e
         JOIN categories c ON e.category_id = c.id
         JOIN accounts a ON e.account_id = a.id
WHERE c.association_id = ? AND e.date > ? AND e.date < ?
ORDER BY e.date DESC;