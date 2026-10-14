-- name: CreateOwnerUser :one
INSERT INTO users (login, password_hash, topt_secret, is_admin, account_type)
VALUES (?, ?, '', FALSE, 'owner') RETURNING *;

-- name: LinkUserOwner :exec
INSERT INTO user_owners (user_id, owner_id)
VALUES (?, ?);

-- name: UnlinkOwner :execrows
DELETE
FROM user_owners
WHERE owner_id = ?;

-- name: CountUserOwners :one
SELECT COUNT(*)
FROM user_owners
WHERE user_id = ?;

-- name: GetOwnerAccount :one
SELECT users.*
FROM user_owners
         JOIN users ON users.id = user_owners.user_id
WHERE user_owners.owner_id = ?;

-- name: ListUserOwners :many
SELECT owners.id,
       owners.name,
       owners.identification_number,
       owners.contact_phone,
       owners.contact_email,
       owners.association_id,
       associations.name AS association_name
FROM user_owners
         JOIN owners ON owners.id = user_owners.owner_id
         JOIN associations ON associations.id = owners.association_id
WHERE user_owners.user_id = ?
ORDER BY associations.name, owners.name;

-- name: GetUserOwnerInAssociation :one
SELECT owners.*
FROM user_owners
         JOIN owners ON owners.id = user_owners.owner_id
WHERE user_owners.user_id = ?
  AND owners.association_id = ?;

-- name: ListAssociationOwnerAccounts :many
SELECT owners.id AS owner_id,
       users.id  AS user_id,
       users.login,
       users.disabled_at,
       users.photo_updated_at
FROM owners
         JOIN user_owners ON user_owners.owner_id = owners.id
         JOIN users ON users.id = user_owners.user_id
WHERE owners.association_id = ?;

-- name: CreateOwnerActivation :one
INSERT INTO owner_activations (owner_id, token_hash, expires_at, created_by)
VALUES (?, ?, ?, ?) RETURNING *;

-- name: GetOwnerActivationByHash :one
SELECT *
FROM owner_activations
WHERE token_hash = ?;

-- name: ListPendingOwnerActivations :many
-- Not used nor revoked (expiry is checked by the caller).
SELECT owner_activations.*
FROM owner_activations
         JOIN owners ON owners.id = owner_activations.owner_id
WHERE owners.association_id = ?
  AND owner_activations.used_at IS NULL
  AND owner_activations.revoked_at IS NULL;

-- name: UseOwnerActivation :execrows
UPDATE owner_activations
SET used_at = datetime('now')
WHERE id = ?
  AND used_at IS NULL
  AND revoked_at IS NULL;

-- name: RevokeOwnerActivations :execrows
UPDATE owner_activations
SET revoked_at = datetime('now')
WHERE owner_id = ?
  AND used_at IS NULL
  AND revoked_at IS NULL;

-- name: UpdateOwnerContact :exec
UPDATE owners
SET contact_phone = ?,
    contact_email = ?,
    updated_at    = datetime()
WHERE id = ?;

-- name: SetUserPhoto :exec
UPDATE users
SET photo_path       = ?,
    photo_updated_at = ?,
    updated_at       = datetime('now')
WHERE id = ?;

-- name: CountUserAssociationLinks :one
SELECT COUNT(*)
FROM users_associations
WHERE user_id = ?;

-- name: ListOwnerActiveUnits :many
SELECT ownerships.id         AS ownership_id,
       ownerships.is_voting,
       ownerships.start_date,
       units.id              AS unit_id,
       units.unit_number,
       units.unit_type,
       units.floor,
       units.entrance,
       units.area,
       units.part,
       units.address,
       buildings.id          AS building_id,
       buildings.name        AS building_name,
       buildings.address     AS building_address
FROM ownerships
         JOIN units ON units.id = ownerships.unit_id
         JOIN buildings ON buildings.id = units.building_id
WHERE ownerships.owner_id = ?
  AND ownerships.is_active = TRUE
ORDER BY buildings.name, units.unit_number;

-- name: ListUnitsCoOwners :many
SELECT ownerships.unit_id, owners.name
FROM ownerships
         JOIN owners ON owners.id = ownerships.owner_id
WHERE ownerships.owner_id != ?
  AND ownerships.is_active = TRUE
  AND ownerships.unit_id IN (sqlc.slice('unit_ids'))
ORDER BY owners.name;

-- name: ListAllUserOwners :many
SELECT user_owners.user_id,
       owners.id,
       owners.name,
       owners.identification_number,
       owners.contact_phone,
       owners.contact_email,
       owners.association_id,
       associations.name AS association_name
FROM user_owners
         JOIN owners ON owners.id = user_owners.owner_id
         JOIN associations ON associations.id = owners.association_id
ORDER BY associations.name, owners.name;
