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
INSERT INTO categories (type, family, name, association_id)
VALUES (?, ?, ?, ?) RETURNING *;