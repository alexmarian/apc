-- +goose Up
-- +goose StatementBegin
SELECT 'Creating gatherings table';

-- Main gathering/voting session table
CREATE TABLE gatherings (
                            id INTEGER PRIMARY KEY,
                            association_id INTEGER NOT NULL REFERENCES associations(id),
                            title TEXT NOT NULL,
                            description TEXT NOT NULL,
                            intent TEXT NOT NULL,
                            gathering_date TIMESTAMP NOT NULL,
                            gathering_type TEXT NOT NULL CHECK (gathering_type IN ('initial', 'repeated')),
                            status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'active', 'closed', 'tallied')),

    -- Qualification criteria
                            qualification_unit_types TEXT, -- JSON array: ["apartment", "commercial"]
                            qualification_floors TEXT, -- JSON array: [1, 2, 3]
                            qualification_entrances TEXT, -- JSON array: [1, 2]
                            qualification_custom_rule TEXT, -- For complex rules

    -- Calculated fields (cached for performance)
                            qualified_units_count INTEGER DEFAULT 0,
                            qualified_units_total_part NUMERIC DEFAULT 0,
                            participating_units_count INTEGER DEFAULT 0,
                            participating_units_total_part NUMERIC DEFAULT 0,

                            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Voting matters/questions table
CREATE TABLE voting_matters (
                                id INTEGER PRIMARY KEY,
                                gathering_id INTEGER NOT NULL REFERENCES gatherings(id) ON DELETE CASCADE,
                                order_index INTEGER NOT NULL,
                                title TEXT NOT NULL,
                                description TEXT,
                                matter_type TEXT NOT NULL CHECK (matter_type IN ('budget', 'election', 'policy', 'poll', 'extraordinary')),
                                voting_type TEXT NOT NULL CHECK (voting_type IN ('yes_no', 'multiple_choice', 'ranking')),

    -- Voting requirements
                                required_majority_type TEXT NOT NULL CHECK (required_majority_type IN ('simple', 'supermajority', 'three_quarters', 'unanimous', 'custom')),
                                required_majority_value NUMERIC NOT NULL CHECK (required_majority_value > 0 AND required_majority_value <= 1),
                                required_quorum NUMERIC DEFAULT 0.5 CHECK (required_quorum >= 0 AND required_quorum <= 1),

    -- Privacy settings
                                is_anonymous BOOLEAN DEFAULT FALSE,
                                show_results_during_voting BOOLEAN DEFAULT FALSE,
                                allow_abstention BOOLEAN DEFAULT TRUE,

    -- Status tracking
                                is_locked BOOLEAN DEFAULT FALSE,
                                locked_at TIMESTAMP,

                                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Options for each voting matter
CREATE TABLE voting_options (
                                id INTEGER PRIMARY KEY,
                                voting_matter_id INTEGER NOT NULL REFERENCES voting_matters(id) ON DELETE CASCADE,
                                option_text TEXT NOT NULL,
                                order_index INTEGER NOT NULL,
                                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

                                UNIQUE(voting_matter_id, order_index)
);

-- Participants in the gathering
CREATE TABLE gathering_participants (
                                        id INTEGER PRIMARY KEY,
                                        gathering_id INTEGER NOT NULL REFERENCES gatherings(id) ON DELETE CASCADE,
                                        unit_id INTEGER NOT NULL REFERENCES units(id),

    -- Participant info
                                        participant_type TEXT NOT NULL CHECK (participant_type IN ('owner', 'delegate')),
                                        participant_name TEXT NOT NULL,
                                        participant_identification TEXT,

    -- For owner participants
                                        owner_id INTEGER REFERENCES owners(id),
                                        ownership_id INTEGER REFERENCES ownerships(id),

    -- For delegate participants
                                        delegating_owner_id INTEGER REFERENCES owners(id),
                                        delegation_document_ref TEXT,

    -- Unit details (cached for history)
                                        unit_number TEXT NOT NULL,
                                        unit_building_name TEXT NOT NULL,
                                        unit_voting_weight NUMERIC NOT NULL,

    -- Participation tracking
                                        check_in_time TIMESTAMP,
                                        voting_completed BOOLEAN DEFAULT FALSE,

                                        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Ensure one participant per unit per gathering
                                        CONSTRAINT unique_unit_per_gathering UNIQUE (gathering_id, unit_id)
);
-- NEW: Voting ballots table - one per participant
CREATE TABLE voting_ballots (
                                id INTEGER PRIMARY KEY,
                                gathering_id INTEGER NOT NULL REFERENCES gatherings(id) ON DELETE CASCADE,
                                participant_id INTEGER NOT NULL REFERENCES gathering_participants(id) ON DELETE CASCADE,

    -- Ballot metadata
                                ballot_hash TEXT, -- SHA256 hash of the ballot content for integrity
                                submitted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                submitted_ip TEXT,
                                submitted_user_agent TEXT,

    -- Future: Digital signature fields
                                signature TEXT, -- Digital signature of the ballot
                                signature_timestamp TIMESTAMP,
                                signature_certificate TEXT, -- Public key or certificate reference

    -- Ensure one ballot per participant per gathering
                                CONSTRAINT unique_ballot_per_participant UNIQUE (gathering_id, participant_id)
);

-- Individual votes within a ballot
CREATE TABLE ballot_votes (
                              id INTEGER PRIMARY KEY,
                              ballot_id INTEGER NOT NULL REFERENCES voting_ballots(id) ON DELETE CASCADE,
                              voting_matter_id INTEGER NOT NULL REFERENCES voting_matters(id) ON DELETE CASCADE,
                              voting_option_id INTEGER REFERENCES voting_options(id),

    -- Vote details
                              vote_value TEXT, -- For abstention or write-in
                              vote_weight NUMERIC NOT NULL, -- The unit's voting weight at time of voting

    -- Ensure one vote per matter per ballot
                              CONSTRAINT unique_vote_per_matter_ballot UNIQUE (ballot_id, voting_matter_id)
);

-- Ballot audit trail
CREATE TABLE ballot_audit_log (
                                  id INTEGER PRIMARY KEY,
                                  ballot_id INTEGER NOT NULL REFERENCES voting_ballots(id),
                                  action TEXT NOT NULL CHECK (action IN ('created', 'submitted', 'verified', 'invalidated')),
                                  performed_by TEXT,
                                  performed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                  ip_address TEXT,
                                  details TEXT -- JSON details
);

-- Notification tracking
CREATE TABLE gathering_notifications (
                                         id INTEGER PRIMARY KEY,
                                         gathering_id INTEGER NOT NULL REFERENCES gatherings(id) ON DELETE CASCADE,
                                         owner_id INTEGER NOT NULL REFERENCES owners(id),
                                         notification_type TEXT NOT NULL CHECK (notification_type IN ('invitation', 'reminder', 'deadline', 'results')),
                                         sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                         read_at TIMESTAMP,
                                         response_status TEXT CHECK (response_status IN ('will_attend', 'delegated', 'declined')),

                                         UNIQUE(gathering_id, owner_id, notification_type)
);

-- Create a view for easy vote counting
CREATE VIEW vote_results AS
SELECT
    vm.gathering_id,
    vm.id as voting_matter_id,
    vm.title as matter_title,
    vm.required_majority_value,
    vo.id as option_id,
    vo.option_text,
    COUNT(bv.id) as vote_count,
    COALESCE(SUM(bv.vote_weight), 0) as total_weight
FROM voting_matters vm
         LEFT JOIN voting_options vo ON vm.id = vo.voting_matter_id
         LEFT JOIN ballot_votes bv ON vo.id = bv.voting_option_id
         LEFT JOIN voting_ballots b ON bv.ballot_id = b.id
GROUP BY vm.gathering_id, vm.id, vm.title, vm.required_majority_value, vo.id, vo.option_text;

-- Create indexes for performance
CREATE INDEX idx_gatherings_association ON gatherings(association_id);
CREATE INDEX idx_gatherings_status ON gatherings(status);
CREATE INDEX idx_voting_matters_gathering ON voting_matters(gathering_id);
CREATE INDEX idx_participants_gathering ON gathering_participants(gathering_id);
CREATE INDEX idx_participants_unit ON gathering_participants(unit_id);
CREATE INDEX idx_ballots_gathering ON voting_ballots(gathering_id);
CREATE INDEX idx_ballots_participant ON voting_ballots(participant_id);
CREATE INDEX idx_ballot_votes_ballot ON ballot_votes(ballot_id);
CREATE INDEX idx_ballot_votes_matter ON ballot_votes(voting_matter_id);
CREATE INDEX idx_notifications_gathering_owner ON gathering_notifications(gathering_id, owner_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'Dropping voting system tables';

DROP VIEW IF EXISTS vote_results;
DROP TABLE IF EXISTS ballot_audit_log;
DROP TABLE IF EXISTS ballot_votes;
DROP TABLE IF EXISTS voting_ballots;
DROP TABLE IF EXISTS gathering_notifications;
DROP TABLE IF EXISTS gathering_participants;
DROP TABLE IF EXISTS voting_options;
DROP TABLE IF EXISTS voting_matters;
DROP TABLE IF EXISTS gatherings;

-- +goose StatementEnd