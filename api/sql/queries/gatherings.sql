-- name: GetGatherings :many
SELECT *
FROM gatherings
WHERE association_id = ?
ORDER BY gathering_date DESC;

-- name: GetGathering :one
SELECT *
FROM gatherings
WHERE id = ?
  AND association_id = ?;

-- name: CreateGathering :one
INSERT INTO gatherings (association_id, title, description, intent, gathering_date,
                        gathering_type, status, qualification_unit_types,
                        qualification_floors, qualification_entrances,
                        qualification_custom_rule, qualified_units_count, qualified_units_total_part,
                        qualified_units_total_area)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING *;

-- name: UpdateGathering :one
UPDATE gatherings
SET title                          = ?,
    description                    = ?,
    intent                         = ?,
    gathering_date                 = ?,
    gathering_type                 = ?,
    qualification_unit_types       = ?,
    qualification_floors           = ?,
    qualification_entrances        = ?,
    qualification_custom_rule      = ?,
    qualified_units_count          = ?,
    qualified_units_total_part     = ?,
    qualified_units_total_area     = ?,
    participating_units_count      = ?,
    participating_units_total_part = ?,
    participating_units_total_area = ?,
    updated_at                     = CURRENT_TIMESTAMP
WHERE id = ?
  AND association_id = ? RETURNING *;


-- name: GetParticipatingUnitSlots :many
SELECT us.*,
       u.unit_number,
       u.floor,
       u.entrance,
       b.name as building_name
FROM unit_slots us
         JOIN units u ON us.unit_id = u.id
         JOIN buildings b ON u.building_id = b.id
WHERE us.gathering_id = ?
  AND us.participant_id = ?
ORDER BY u.unit_number;

-- name: GetGatheringUnitSlots :many
SELECT us.*,
       u.unit_number,
       u.floor,
       u.entrance,
       b.name as building_name
FROM unit_slots us
         JOIN units u ON us.unit_id = u.id
         JOIN buildings b ON u.building_id = b.id
WHERE us.gathering_id = ?;
-- name: GetUnitSlot :one
SELECT us.*,
       u.unit_number,
       u.floor,
       u.entrance,
       b.name as building_name
FROM unit_slots us
         JOIN units u ON us.unit_id = u.id
         JOIN buildings b ON u.building_id = b.id
WHERE us.gathering_id = ?
  AND us.id = ?;
-- name: AddUnitSlot :one
INSERT INTO unit_slots (gathering_id, unit_id)
VALUES (?, ?) RETURNING *;

-- name: AssignUnitSlot :one
UPDATE unit_slots
SET participant_id = ?,
    updated_at     = CURRENT_TIMESTAMP
WHERE gathering_id = ?
  AND participant_id IS NULL
  AND id = ? RETURNING *;

-- name: RemoveUnitSlot :exec
DELETE
FROM unit_slots
WHERE gathering_id = ?;
-- name: UpdateGatheringStatus :one
UPDATE gatherings
SET status     = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
  AND association_id = ? RETURNING *;

-- name: UpdateGatheringStats :exec
UPDATE gatherings
SET qualified_units_count          = ?,
    qualified_units_total_part     = ?,
    qualified_units_total_area     = ?,
    updated_at                     = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: UpdateParticipationStats :exec
UPDATE gatherings
SET participating_units_count      = ?,
    participating_units_total_part = ?,
    participating_units_total_area = ?,
    updated_at                     = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: GetVotingMatters :many
SELECT *
FROM voting_matters
WHERE gathering_id = ?
ORDER BY order_index;

-- name: GetVotingMatter :one
SELECT *
FROM voting_matters
WHERE id = ?
  AND gathering_id = ?;

-- name: CreateVotingMatter :one
INSERT INTO voting_matters (gathering_id, order_index, title, description, matter_type, voting_config)
VALUES (?, ?, ?, ?, ?, ?) RETURNING *;

-- name: UpdateVotingMatter :one
UPDATE voting_matters
SET title         = ?,
    description   = ?,
    matter_type   = ?,
    voting_config = ?,
    updated_at    = CURRENT_TIMESTAMP
WHERE id = ?
  AND gathering_id = ? RETURNING *;

-- name: DeleteVotingMatter :exec
DELETE
FROM voting_matters
WHERE id = ?
  AND gathering_id = ?;

-- name: GetGatheringParticipants :many
SELECT gp.*,
       o.name                  as owner_name,
       o.identification_number as owner_identification,
       delo.name               as delegating_owner_name,
       u.unit_number,
       u.floor,
       u.entrance,
       b.name                  as building_name
FROM gathering_participants gp
         LEFT JOIN owners o ON gp.owner_id = o.id
         LEFT JOIN owners delo ON gp.delegating_owner_id = delo.id
         LEFT JOIN units u ON gp.unit_id = u.id
         LEFT JOIN buildings b ON u.building_id = b.id
WHERE gp.gathering_id = ?
ORDER BY gp.participant_name;

-- name: GetGatheringParticipant :one
SELECT *
FROM gathering_participants
WHERE id = ?
  AND gathering_id = ?;

-- name: CreateGatheringParticipant :one
INSERT INTO gathering_participants (gathering_id, unit_id, participant_type, participant_name,
                                    participant_identification, owner_id, delegating_owner_id,
                                    delegation_document_ref, units_info, units_area, units_part)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING *;

-- name: UpdateGatheringParticipantsUnits :one
UPDATE gathering_participants
SET units_info = ?,
    units_area = ?,
    units_part = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
  AND gathering_id = ? RETURNING *;

-- name: DeleteGatheringParticipant :exec
DELETE
FROM gathering_participants
WHERE id = ?
  AND gathering_id = ?;

-- name: CheckInParticipant :exec
UPDATE gathering_participants
SET check_in_time = CURRENT_TIMESTAMP,
    updated_at    = CURRENT_TIMESTAMP
WHERE id = ?
  AND gathering_id = ?;

-- name: GetParticipantByUnit :one
SELECT *
FROM gathering_participants
WHERE gathering_id = ?
  AND unit_id = ?;

-- name: GetQualifiedUnits :many
SELECT u.id,
       u.unit_number,
       u.cadastral_number,
       u.floor,
       u.entrance,
       u.area,
       u.part,
       u.unit_type,
       b.name    as building_name,
       b.address as building_address
FROM units u
         JOIN buildings b ON u.building_id = b.id
WHERE b.association_id = ?
  AND (false = ? OR u.unit_type IN (sqlc.slice('unit_types')))
  AND (false = ? OR u.floor IN (sqlc.slice('unit_floors')))
  AND (false = ? OR u.entrance IN (sqlc.slice('unit_entrances')))
ORDER BY b.name, u.unit_number;

-- name: GetNonParticipatingOwners :many
SELECT DISTINCT o.id,
                o.name,
                o.identification_number,
                o.contact_email,
                o.contact_phone,
                COUNT(DISTINCT own.unit_id) as units_count
FROM owners o
         JOIN ownerships own ON o.id = own.owner_id
         JOIN units u ON own.unit_id = u.id
         JOIN buildings b ON u.building_id = b.id
WHERE b.association_id = ?
  AND own.is_active = TRUE
  AND u.id IN (SELECT u2.id
               FROM units u2
                        JOIN buildings b2 ON u2.building_id = b2.id
               WHERE b2.association_id = ?
                 AND (false = ? OR u2.unit_type IN (sqlc.slice('unit_types')))
                 AND (false = ? OR u2.floor IN (sqlc.slice('unit_floors')))
                 AND (false = ? OR u2.entrance IN (sqlc.slice('unit_entrances'))))
  AND u.id NOT IN (SELECT unit_id
                   FROM gathering_participants
                   WHERE gathering_id = ?)
GROUP BY o.id, o.name, o.identification_number, o.contact_email, o.contact_phone
ORDER BY o.name;

-- name: CreateBallot :one
INSERT INTO voting_ballots (gathering_id, participant_id, ballot_content, ballot_hash,
                            submitted_ip, submitted_user_agent)
VALUES (?, ?, ?, ?, ?, ?) RETURNING *;

-- name: GetBallotByParticipant :one
SELECT *
FROM voting_ballots
WHERE gathering_id = ?
  AND participant_id = ?;

-- name: GetBallotsForGathering :many
SELECT vb.*,
       gp.participant_name,
       gp.unit_id,
       gp.units_info,
       gp.units_area,
       gp.units_part
FROM voting_ballots vb
         JOIN gathering_participants gp ON vb.participant_id = gp.id
WHERE vb.gathering_id = ?
ORDER BY vb.submitted_at;

-- name: InvalidateBallot :exec
UPDATE voting_ballots
SET is_valid            = FALSE,
    invalidated_at      = CURRENT_TIMESTAMP,
    invalidation_reason = ?
WHERE id = ?;

-- name: GetVoteTally :one
SELECT *
FROM vote_tallies
WHERE gathering_id = ?
  AND voting_matter_id = ?;

-- name: UpsertVoteTally :one
INSERT INTO vote_tallies (gathering_id, voting_matter_id, tally_data)
VALUES (?, ?, ?) ON CONFLICT (gathering_id, voting_matter_id)
DO
UPDATE SET
    tally_data = ?,
    last_updated = CURRENT_TIMESTAMP
    RETURNING *;

-- name: GetAllVoteTallies :many
SELECT vt.*,
       vm.title as matter_title,
       vm.matter_type,
       vm.voting_config
FROM vote_tallies vt
         JOIN voting_matters vm ON vt.voting_matter_id = vm.id
WHERE vt.gathering_id = ?
ORDER BY vm.order_index;

-- name: CreateAuditLog :exec
INSERT INTO voting_audit_log (gathering_id, entity_type, entity_id, action,
                              performed_by, ip_address, details)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetAuditLogs :many
SELECT *
FROM voting_audit_log
WHERE gathering_id = ?
ORDER BY performed_at DESC LIMIT ?;

-- name: CreateNotification :one
INSERT INTO voting_notifications (gathering_id, owner_id, notification_type, sent_via)
VALUES (?, ?, ?, ?) ON CONFLICT (gathering_id, owner_id, notification_type)
DO
UPDATE SET
    sent_at = CURRENT_TIMESTAMP,
    sent_via = ?
    RETURNING *;

-- name: UpdateNotificationRead :exec
UPDATE voting_notifications
SET read_at = CURRENT_TIMESTAMP
WHERE gathering_id = ?
  AND owner_id = ?
  AND notification_type = ?;

-- name: GetNotifications :many
SELECT vn.*,
       o.name as owner_name,
       o.contact_email,
       o.contact_phone
FROM voting_notifications vn
         JOIN owners o ON vn.owner_id = o.id
WHERE vn.gathering_id = ?
ORDER BY vn.sent_at DESC;

-- name: GetGatheringStats :one
SELECT *
FROM gatherings
WHERE id = ?
  AND association_id = ?;

-- name: GetGatheringParticipantCount :one
SELECT COUNT(*) as count
FROM gathering_participants
WHERE gathering_id = ?;

-- name: GetGatheringBallotCount :one
SELECT COUNT(*) as count
FROM voting_ballots
WHERE gathering_id = ? AND is_valid = TRUE;

-- name: GetActiveOwnerUnitsForGathering :many
SELECT o.id                    as owner_id,
       o.name                  as owner_name,
       o.normalized_name       as owner_normalized_name,
       o.identification_number as owner_identification,
       o.contact_email         as owner_contact_email,
       o.contact_phone         as owner_contact_phone,
       u.id                    as unit_id,
       u.cadastral_number,
       u.unit_number,
       u.area,
       u.part                  as voting_weight,
       u.unit_type,
       u.floor,
       u.entrance,
       b.id                    as building_id,
       b.name                  as building_name,
       b.address               as building_address,
       b.association_id,
       own.id                  as ownership_id
FROM owners o
         JOIN ownerships own ON o.id = own.owner_id
         JOIN units u ON own.unit_id = u.id
         JOIN buildings b ON u.building_id = b.id
WHERE own.is_active = TRUE
  AND own.is_voting = TRUE
  AND b.association_id = ?
  AND (false = ? OR u.unit_type IN (sqlc.slice('unit_types')))
  AND (false = ? OR u.floor IN (sqlc.slice('unit_floors')))
  AND (false = ? OR u.entrance IN (sqlc.slice('unit_entrances')))
ORDER BY o.name, u.unit_number;