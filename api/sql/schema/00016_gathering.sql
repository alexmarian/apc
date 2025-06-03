-- +goose Up
-- +goose StatementBegin
SELECT 'Creating gatherings table';

-- Main gathering/voting session table
CREATE TABLE gatherings
(
    id                             INTEGER PRIMARY KEY,
    association_id                 INTEGER   NOT NULL REFERENCES associations (id),
    title                          TEXT      NOT NULL,
    description                    TEXT      NOT NULL,
    intent                         TEXT      NOT NULL,
    gathering_date                 TIMESTAMP NOT NULL,
    gathering_type                 TEXT      NOT NULL CHECK (gathering_type IN ('initial', 'repeated')),
    status                         TEXT      NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'active', 'closed', 'tallied')),

    -- Qualification criteria
    qualification_unit_types       TEXT, -- JSON array: ["apartment", "commercial"]
    qualification_floors           TEXT, -- JSON array: [1, 2, 3]
    qualification_entrances        TEXT, -- JSON array: [1, 2]
    qualification_custom_rule      TEXT, -- For complex rules

    -- Calculated fields (cached for performance)
    qualified_units_count          INTEGER            DEFAULT 0,
    qualified_units_total_part     NUMERIC            DEFAULT 0,
    qualified_units_total_area     NUMERIC            DEFAULT 0,
    participating_units_count      INTEGER            DEFAULT 0,
    participating_units_total_part NUMERIC            DEFAULT 0,
    participating_units_total_area NUMERIC            DEFAULT 0,

    created_at                     TIMESTAMP          DEFAULT CURRENT_TIMESTAMP,
    updated_at                     TIMESTAMP          DEFAULT CURRENT_TIMESTAMP
);

-- Voting matters/questions table
CREATE TABLE voting_matters
(
    id            INTEGER PRIMARY KEY,
    gathering_id  INTEGER NOT NULL REFERENCES gatherings (id) ON DELETE CASCADE,
    order_index   INTEGER NOT NULL,
    title         TEXT    NOT NULL,
    description   TEXT,
    matter_type   TEXT    NOT NULL CHECK (matter_type IN ('budget', 'election', 'policy', 'poll', 'extraordinary')),

    -- Voting configuration (JSON)
    voting_config TEXT    NOT NULL, -- JSON with: type, options[], required_majority, quorum, allow_abstention, is_anonymous

    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE (gathering_id, order_index)
);
CREATE TABLE gathering_participants
(
    id                         INTEGER PRIMARY KEY,
    gathering_id               INTEGER NOT NULL REFERENCES gatherings (id) ON DELETE CASCADE,
    -- Participant info
    participant_type           TEXT    NOT NULL CHECK (participant_type IN ('owner', 'delegate')),
    participant_name           TEXT    NOT NULL,
    participant_identification TEXT,

    -- For owner participants
    owner_id                   INTEGER REFERENCES owners (id),

    -- For delegate participants
    delegating_owner_id        INTEGER REFERENCES owners (id),
    delegation_document_ref    TEXT,

    -- JSON Array of unit IDs for multi-unit participants
    units_info                  TEXT    NOT NULL,
    units_area                  NUMERIC NOT NULL, -- Total area of all units
    units_part                  NUMERIC NOT NULL, -- Total voting weight of all units

    -- Participation tracking
    check_in_time              TIMESTAMP,

    created_at                 TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at                 TIMESTAMP DEFAULT CURRENT_TIMESTAMP

);
-- Units slots used till voting state/ deleted and archived in participant on closing the gathering
CREATE TABLE unit_slots
(
    id            INTEGER PRIMARY KEY,
    gathering_id  INTEGER NOT NULL REFERENCES gatherings (id) ON DELETE CASCADE,
    unit_id       INTEGER NOT NULL REFERENCES units (id),
    participant_id INTEGER NULL REFERENCES gathering_participants (id),

    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE (gathering_id, unit_id, participant_id),
    CONSTRAINT unique_unit_per_gathering UNIQUE (gathering_id, unit_id)
);


-- Voting ballots table - stores complete ballot as JSON
CREATE TABLE voting_ballots
(
    id                    INTEGER PRIMARY KEY,
    gathering_id          INTEGER NOT NULL REFERENCES gatherings (id) ON DELETE CASCADE,
    participant_id        INTEGER NOT NULL REFERENCES gathering_participants (id) ON DELETE CASCADE,

    -- The complete ballot content
    ballot_content        TEXT    NOT NULL, -- JSON with all votes: {matter_id: {option_id, vote_value}, ...}
    ballot_hash           TEXT    NOT NULL, -- SHA256 hash of ballot_content

    -- Submission metadata
    submitted_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    submitted_ip          TEXT,
    submitted_user_agent  TEXT,

    -- Digital signature (future)
    signature             TEXT,
    signature_timestamp   TIMESTAMP,
    signature_certificate TEXT,

    -- Status
    is_valid              BOOLEAN   DEFAULT TRUE,
    invalidated_at        TIMESTAMP,
    invalidation_reason   TEXT,

    -- Ensure one ballot per participant
    CONSTRAINT unique_ballot_per_participant UNIQUE (gathering_id, participant_id)
);
-- Materialized view for vote counting (refreshed after each ballot submission)
CREATE TABLE vote_tallies
(
    id               INTEGER PRIMARY KEY,
    gathering_id     INTEGER NOT NULL REFERENCES gatherings (id) ON DELETE CASCADE,
    voting_matter_id INTEGER NOT NULL REFERENCES voting_matters (id) ON DELETE CASCADE,

    -- Tally data (JSON)
    tally_data       TEXT    NOT NULL, -- JSON with vote counts, weights, and details per option

    last_updated     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE (gathering_id, voting_matter_id)
);

-- Audit log for all voting actions
CREATE TABLE voting_audit_log
(
    id           INTEGER PRIMARY KEY,
    gathering_id INTEGER NOT NULL REFERENCES gatherings (id),
    entity_type  TEXT    NOT NULL CHECK (entity_type IN ('gathering', 'matter', 'participant', 'ballot')),
    entity_id    INTEGER NOT NULL,
    action       TEXT    NOT NULL,
    performed_by TEXT,
    performed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ip_address   TEXT,
    details      TEXT -- JSON with action-specific details
);

-- Notification tracking
CREATE TABLE voting_notifications
(
    id                INTEGER PRIMARY KEY,
    gathering_id      INTEGER NOT NULL REFERENCES gatherings (id) ON DELETE CASCADE,
    owner_id          INTEGER NOT NULL REFERENCES owners (id),
    notification_type TEXT    NOT NULL CHECK (notification_type IN ('invitation', 'reminder', 'results')),
    sent_at           TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    sent_via          TEXT, -- 'email', 'sms', etc.
    read_at           TIMESTAMP,

    UNIQUE (gathering_id, owner_id, notification_type)
);

-- Adding active owner units view

-- View showing all active ownership relationships (SQLite compatible)
CREATE VIEW active_owner_units AS
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
WHERE own.is_active = TRUE;

-- Create indexes for performance
CREATE INDEX idx_gatherings_association_status ON gatherings (association_id, status);
CREATE INDEX idx_voting_matters_gathering ON voting_matters (gathering_id);
CREATE INDEX idx_participants_gathering ON gathering_participants (gathering_id);
CREATE INDEX idx_ballots_gathering ON voting_ballots (gathering_id);
CREATE INDEX idx_ballots_participant ON voting_ballots (participant_id);
CREATE INDEX idx_vote_tallies_gathering ON vote_tallies (gathering_id);
CREATE INDEX idx_audit_log_gathering ON voting_audit_log (gathering_id, performed_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'Dropping simplified ballot-based voting system tables';

DROP VIEW IF EXISTS active_owner_units;
DROP TABLE IF EXISTS voting_notifications;
DROP TABLE IF EXISTS voting_audit_log;
DROP TABLE IF EXISTS vote_tallies;
DROP TABLE IF EXISTS voting_ballots;
DROP TABLE IF EXISTS gathering_participants;
DROP TABLE IF EXISTS voting_matters;
DROP TABLE IF EXISTS gatherings;

-- +goose StatementEnd