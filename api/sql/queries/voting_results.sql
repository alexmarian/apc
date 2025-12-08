-- name: GetVotingResults :one
SELECT * FROM voting_results
WHERE gathering_id = ?;

-- name: CreateVotingResults :one
INSERT INTO voting_results (
    gathering_id,
    results_data,
    voting_mode,
    gathering_type,
    total_possible_votes_weight,
    total_possible_votes_count,
    quorum_threshold_percentage,
    quorum_met
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?) RETURNING *;

-- name: UpdateVotingResults :one
UPDATE voting_results
SET results_data = ?,
    voting_mode = ?,
    gathering_type = ?,
    total_possible_votes_weight = ?,
    total_possible_votes_count = ?,
    quorum_threshold_percentage = ?,
    quorum_met = ?,
    computed_at = CURRENT_TIMESTAMP
WHERE gathering_id = ? RETURNING *;

-- name: DeleteVotingResults :exec
DELETE FROM voting_results
WHERE gathering_id = ?;
