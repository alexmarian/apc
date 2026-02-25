-- name: GetCategory :one
SELECT *
FROM categories
WHERE id = ? LIMIT 1;

-- name: GetAssociationCategory :one
SELECT *
FROM categories
WHERE id = ? and association_id = ? LIMIT 1;

-- name: DeactivateCategory :exec
UPDATE categories
SET is_deleted = TRUE,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- When retrieving categories, filter out deleted ones
-- name: GetActiveCategories :many
SELECT *
FROM categories
WHERE association_id = ?
  AND is_deleted = FALSE
ORDER BY type, family, name;

-- When retrieving categories for dropdown lists, filter out deleted ones
-- name: GetCategoriesForDropdown :many
SELECT id, name
FROM categories
WHERE association_id = ?
  AND is_deleted = FALSE
ORDER BY name;

-- In all expense endpoints, validate that the category is not deleted
-- name: IsCategoryActive :one
SELECT EXISTS(SELECT 1
              FROM categories
              WHERE id = ?
                AND is_deleted = FALSE) as is_active;

-- name: CreateCategory :one
INSERT INTO categories (type, family, name, association_id, original_labels)
VALUES (?, ?, ?, ?, ?) RETURNING *;

-- name: GetAllCategories :many
SELECT *
FROM categories
WHERE association_id = ?
  AND (? = TRUE OR is_deleted = FALSE)
ORDER BY is_deleted ASC, type, family, name;

-- name: UpdateCategory :one
UPDATE categories
SET type = ?,
    family = ?,
    name = ?,
    original_labels = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
  AND association_id = ?
RETURNING *;

-- name: ReactivateCategory :exec
UPDATE categories
SET is_deleted = FALSE,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
  AND association_id = ?;

-- name: GetCategoryUsageCount :one
SELECT COUNT(*) as usage_count
FROM expenses
WHERE category_id = ?;

-- name: GetCategoryUsageDetails :many
SELECT id, description, amount, date, created_at
FROM expenses
WHERE category_id = ?
ORDER BY date DESC
LIMIT 10;

-- name: BulkDeactivateCategories :exec
UPDATE categories
SET is_deleted = TRUE,
    updated_at = CURRENT_TIMESTAMP
WHERE id IN (sqlc.slice('category_ids'))
  AND association_id = ?;

-- name: BulkReactivateCategories :exec
UPDATE categories
SET is_deleted = FALSE,
    updated_at = CURRENT_TIMESTAMP
WHERE id IN (sqlc.slice('category_ids'))
  AND association_id = ?;

-- name: CheckCategoryUniqueness :one
SELECT COUNT(*) as count
FROM categories
WHERE association_id = ?
  AND type = ?
  AND family = ?
  AND name = ?
  AND is_deleted = FALSE
  AND (? = 0 OR id != ?);