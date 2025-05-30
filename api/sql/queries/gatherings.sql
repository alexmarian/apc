-- name: GetGatherings :many
SELECT * FROM gatherings
WHERE association_id = ?
ORDER BY gathering_date DESC;
--

-- name: GetGathering :one
SELECT * FROM gatherings
WHERE id = ? AND association_id = ?;
--

-- name: CreateGathering :one
INSERT INTO gatherings (
    association_id, title, description, intent, gathering_date,
    gathering_type, status, qualification_unit_types,
    qualification_floors, qualification_entrances,
    qualification_custom_rule
) VALUES (
             ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
         ) RETURNING *;
--

-- name: UpdateGathering :one
UPDATE gatherings SET
                      title = ?,
                      description = ?,
                      intent = ?,
                      gathering_date = ?,
                      gathering_type = ?,
                      qualification_unit_types = ?,
                      qualification_floors = ?,
                      qualification_entrances = ?,
                      qualification_custom_rule = ?,
                      updated_at = CURRENT_TIMESTAMP
WHERE id = ? AND association_id = ?
    RETURNING *;
--

-- name: UpdateGatheringStatus :one
UPDATE gatherings SET
                      status =?,
                      updated_at = CURRENT_TIMESTAMP
WHERE id = ? AND association_id = ?
    RETURNING *;

-- name: UpdateGatheringStats :exec
UPDATE gatherings SET
                      qualified_units_count = ?,
                      qualified_units_total_part = ?,
                      participating_units_count = ?,
                      participating_units_total_part = ?,
                      updated_at = CURRENT_TIMESTAMP
WHERE id = ?;
--

-- name: GetVotingMatters :many
SELECT * FROM voting_matters
WHERE gathering_id = ?
ORDER BY order_index;
--

-- name: GetVotingMatter :one
SELECT * FROM voting_matters
WHERE id = ? AND gathering_id = ?;
--

-- name: CreateVotingMatter :one
INSERT INTO voting_matters (
    gathering_id, order_index, title, description, matter_type,
    voting_type, required_majority_type, required_majority_value,
    required_quorum, is_anonymous, show_results_during_voting,
    allow_abstention
) VALUES (
             ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
         ) RETURNING *;
--

-- name: UpdateVotingMatter :one
UPDATE voting_matters SET
                          title = ?,
                          description = ?,
                          matter_type = ?,
                          voting_type = ?,
                          required_majority_type = ?,
                          required_majority_value = ?,
                          required_quorum = ?,
                          is_anonymous = ?,
                          show_results_during_voting = ?,
                          allow_abstention = ?,
                          updated_at = CURRENT_TIMESTAMP
WHERE id = ? AND gathering_id = ? AND is_locked = FALSE
    RETURNING *;
--

-- name: LockVotingMatters :exec
UPDATE voting_matters SET
                          is_locked = TRUE,
                          locked_at = CURRENT_TIMESTAMP
WHERE gathering_id = ?;
--

-- name: GetVotingOptions :many
SELECT * FROM voting_options
WHERE voting_matter_id = ?
ORDER BY order_index;
--

-- name: CreateVotingOption :one
INSERT INTO voting_options (
    voting_matter_id, option_text, order_index
) VALUES (
             ?, ?, ?
         ) RETURNING *;
--

-- name: DeleteVotingOptions :exec
DELETE  FROM voting_options
WHERE voting_matter_id = ? ;
--

-- name: GetGatheringParticipants :many
SELECT
    gp.*,
    o.name as owner_name,
    o.identification_number as owner_identification,
    delo.name as delegating_owner_name,
    u.unit_number,
    u.floor,
    u.entrance,
    b.name as building_name
FROM gathering_participants gp
         LEFT JOIN owners o ON gp.owner_id = o.id
         LEFT JOIN owners delo ON gp.delegating_owner_id = delo.id
         LEFT JOIN units u ON gp.unit_id = u.id
         LEFT JOIN buildings b ON u.building_id = b.id
WHERE gp.gathering_id = ?
ORDER BY gp.participant_name;
--

-- name: GetGatheringParticipant :one
SELECT * FROM gathering_participants
WHERE id = ? AND gathering_id = ?;
--

-- name: CreateGatheringParticipant :one
INSERT INTO gathering_participants (
    gathering_id, unit_id, participant_type, participant_name,
    participant_identification, owner_id, ownership_id,
    delegating_owner_id, delegation_document_ref,
    unit_number, unit_building_name, unit_voting_weight
) VALUES (
             ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
         ) RETURNING *;
--

-- name: CheckInParticipant :exec
UPDATE gathering_participants SET
                                  check_in_time = CURRENT_TIMESTAMP,
                                  updated_at = CURRENT_TIMESTAMP
WHERE id = ? AND gathering_id = ?;
--

-- name: MarkVotingCompleted :exec
UPDATE gathering_participants SET
                                  voting_completed = TRUE,
                                  updated_at = CURRENT_TIMESTAMP
WHERE id = ?;
--

-- name: GetParticipantByUnit :one
SELECT * FROM gathering_participants
WHERE gathering_id = ? AND unit_id = ?;
--

-- name: GetQualifiedUnits :many
SELECT
    u.id,
    u.unit_number,
    u.cadastral_number,
    u.floor,
    u.entrance,
    u.area,
    u.part,
    u.unit_type,
    b.name as building_name,
    b.address as building_address
FROM units u
         JOIN buildings b ON u.building_id = b.id
WHERE b.association_id = ?
  AND (false=? OR u.unit_type in(sqlc.slice('unit_types')))
  AND (false=? OR u.floor in(sqlc.slice('unit_floors')))
  AND (false=? OR u.entrance in(sqlc.slice('unit_entrances')))
ORDER BY b.name, u.unit_number;
--

-- name: GetNonParticipatingOwners :many
SELECT DISTINCT
    o.id,
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
  AND u.id IN (
    SELECT id FROM units u2
                       JOIN buildings b2 ON u2.building_id = b2.id
    WHERE b2.association_id = ?
      AND (false=? OR u2.unit_type in(sqlc.slice('unit_types')))
      AND (false=? OR u2.floor in(sqlc.slice('unit_floors')))
      AND (false=? OR u2.entrance in(sqlc.slice('unit_entrances')))
)
  AND u.id NOT IN (
    SELECT unit_id FROM gathering_participants
    WHERE gathering_id = ?
)
GROUP BY o.id, o.name, o.identification_number, o.contact_email, o.contact_phone
ORDER BY o.name;
--

-- name: CreateVote :one
INSERT INTO votes (
    gathering_id, participant_id, voting_matter_id,
    voting_option_id, vote_value, vote_weight
) VALUES (
             ?, ?, ?, ?, ?, ?
         ) RETURNING *;
--

-- name: GetVotes :many
SELECT
    v.*,
    vm.title as matter_title,
    vo.option_text,
    gp.participant_name,
    gp.unit_number
FROM votes v
         JOIN voting_matters vm ON v.voting_matter_id = vm.id
         LEFT JOIN voting_options vo ON v.voting_option_id = vo.id
         JOIN gathering_participants gp ON v.participant_id = gp.id
WHERE v.gathering_id = ?
ORDER BY vm.order_index, v.submitted_at;
--

-- name: GetVotesByMatter :many
SELECT
    v.*,
    vo.option_text,
    gp.participant_name,
    gp.unit_number,
    gp.participant_type
FROM votes v
         LEFT JOIN voting_options vo ON v.voting_option_id = vo.id
         JOIN gathering_participants gp ON v.participant_id = gp.id
WHERE v.voting_matter_id = ?
ORDER BY v.submitted_at;
--

-- name: GetVoteResults :many
SELECT
    vm.id as matter_id,
    vm.title as matter_title,
    vm.required_majority_type,
    vm.required_majority_value,
    vm.is_anonymous,
    vo.id as option_id,
    vo.option_text,
    COALESCE(SUM(v.vote_weight), 0) as total_weight,
    COUNT(v.id) as vote_count
FROM voting_matters vm
         LEFT JOIN voting_options vo ON vm.id = vo.voting_matter_id
         LEFT JOIN votes v ON vo.id = v.voting_option_id
WHERE vm.gathering_id = ?
GROUP BY vm.id, vm.title, vm.required_majority_type,
         vm.required_majority_value, vm.is_anonymous, vo.id, vo.option_text
ORDER BY vm.order_index, vo.order_index;
--

-- name: GetParticipantVotes :many
SELECT
    v.*,
    vm.title as matter_title,
    vo.option_text
FROM votes v
         JOIN voting_matters vm ON v.voting_matter_id = vm.id
         LEFT JOIN voting_options vo ON v.voting_option_id = vo.id
WHERE v.participant_id = ?
ORDER BY vm.order_index;
--

-- name: CreateNotification :one
INSERT INTO gathering_notifications (
    gathering_id, owner_id, notification_type
) VALUES (
             ?, ?, ?
         ) ON CONFLICT (gathering_id, owner_id, notification_type)
DO UPDATE SET sent_at = CURRENT_TIMESTAMP
                  RETURNING *;
--

-- name: UpdateNotificationStatus :exec
UPDATE gathering_notifications SET
                                   read_at = CURRENT_TIMESTAMP,
                                   response_status = ?
WHERE gathering_id = ? AND owner_id = ? AND notification_type = ?;
--

-- name: GetNotifications :many
SELECT * FROM gathering_notifications
WHERE gathering_id = ?
ORDER BY sent_at DESC;
--

-- name: CreateAuditLog :exec
INSERT INTO vote_audit_log (
    vote_id, action, performed_by, ip_address, user_agent, details
) VALUES (
             ?, ?, ?, ?, ?, ?
         );
--

-- name: GetGatheringStats :one
SELECT
    g.*,
    COUNT(DISTINCT gp.id) as participant_count,
    COUNT(DISTINCT CASE WHEN gp.voting_completed THEN gp.id END) as completed_count,
    COALESCE(SUM(gp.unit_voting_weight), 0) as total_participating_weight
FROM gatherings g
         LEFT JOIN gathering_participants gp ON g.id = gp.gathering_id
WHERE g.id = ? AND g.association_id = ?
GROUP BY g.id;