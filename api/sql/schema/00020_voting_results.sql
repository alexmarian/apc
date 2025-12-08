-- +goose Up
-- +goose StatementBegin
SELECT 'Creating voting_results cache table';

-- Table to store computed voting results (generated when gathering closes)
CREATE TABLE voting_results (
    id INTEGER PRIMARY KEY,
    gathering_id INTEGER NOT NULL REFERENCES gatherings(id) ON DELETE CASCADE,

    -- Computed results (JSON)
    results_data TEXT NOT NULL, -- Complete VoteResults struct as JSON

    -- Metadata
    voting_mode TEXT NOT NULL,
    gathering_type TEXT NOT NULL,
    total_possible_votes_weight NUMERIC NOT NULL,
    total_possible_votes_count INTEGER NOT NULL,
    quorum_threshold_percentage NUMERIC NOT NULL,
    quorum_met BOOLEAN NOT NULL,

    -- Timestamps
    computed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Ensure one result per gathering
    UNIQUE(gathering_id)
);

CREATE INDEX idx_voting_results_gathering ON voting_results(gathering_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'Dropping voting_results table';

DROP TABLE IF EXISTS voting_results;

-- +goose StatementEnd
