-- name: CreateMemberInvitation :one
INSERT INTO member_invitations (
    gathering_id,
    owner_id,
    token_hash,
    expires_at
) VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetMemberInvitationByTokenHash :one
SELECT *
FROM member_invitations
WHERE token_hash = ?
  AND revoked_at IS NULL
  AND expires_at > datetime('now')
LIMIT 1;

-- name: ListMemberInvitationsByGathering :many
SELECT *
FROM member_invitations
WHERE gathering_id = ?
ORDER BY created_at DESC;

-- name: GetActiveMemberInvitationByOwnerAndGathering :one
SELECT *
FROM member_invitations
WHERE gathering_id = ?
  AND owner_id = ?
  AND revoked_at IS NULL
LIMIT 1;

-- name: RevokeMemberInvitation :exec
UPDATE member_invitations
SET revoked_at = datetime('now'),
    updated_at = datetime('now')
WHERE id = ?
  AND revoked_at IS NULL;
