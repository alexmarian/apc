package domain

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/alexmarian/apc/api/internal/database"
)

// Path value constants
const (
	GatheringIDPathValue    = "gatheringId"
	VotingMatterIDPathValue = "matterId"
	ParticipantIDPathValue  = "participantId"
)

// Gathering represents a gathering event
type Gathering struct {
	ID                          int64     `json:"id"`
	AssociationID               int64     `json:"association_id"`
	Title                       string    `json:"title"`
	Description                 string    `json:"description"`
	Intent                      string    `json:"intent"`
	Location                    string    `json:"location"`
	GatheringDate               time.Time `json:"scheduled_date"` // Frontend expects scheduled_date
	GatheringType               string    `json:"type"`           // Frontend expects type
	VotingMode                  string    `json:"voting_mode"`    // by_weight or by_unit
	Status                      string    `json:"status"`
	QualificationUnitTypes      []string  `json:"qualification_unit_types"`
	QualificationFloors         []int64   `json:"qualification_floors"`
	QualificationEntrances      []int64   `json:"qualification_entrances"`
	QualificationCustomRule     string    `json:"qualification_custom_rule"`
	QualifiedUnitsCount         int       `json:"qualified_units"`
	QualifiedUnitsTotalPart     float64   `json:"qualified_weight"`
	QualifiedUnitsTotalArea     float64   `json:"qualified_area"`
	ParticipatingUnitsCount     int       `json:"participating_units"`
	ParticipatingUnitsTotalPart float64   `json:"participating_weight"`
	ParticipatingUnitsTotalArea float64   `json:"participating_area"`
	CreatedAt                   time.Time `json:"created_at"`
	UpdatedAt                   time.Time `json:"updated_at"`
}

// CreateGatheringRequest represents the request to create a gathering
type CreateGatheringRequest struct {
	Title                   string    `json:"title"`
	Description             string    `json:"description"`
	Intent                  string    `json:"intent"`
	Location                string    `json:"location"`
	GatheringDate           time.Time `json:"gathering_date"`
	GatheringType           string    `json:"gathering_type"`
	VotingMode              string    `json:"voting_mode"` // by_weight or by_unit
	QualificationUnitTypes  []string  `json:"qualification_unit_types"`
	QualificationFloors     []int64   `json:"qualification_floors"`
	QualificationEntrances  []int64   `json:"qualification_entrances"`
	QualificationCustomRule string    `json:"qualification_custom_rule"`
}

// QuorumInfo contains detailed information about quorum calculation
type QuorumInfo struct {
	Required           float64 `json:"required"`            // Required threshold (weight or count)
	Achieved           float64 `json:"achieved"`            // Achieved participation (weight or count)
	RequiredPercentage float64 `json:"required_percentage"` // Threshold percentage (25, 50, 100)
	AchievedPercentage float64 `json:"achieved_percentage"` // Actual participation percentage
	Met                bool    `json:"met"`                 // Whether quorum is met
	VotingMode         string  `json:"voting_mode"`         // by_weight or by_unit
	GatheringType      string  `json:"gathering_type"`      // initial, repeated, remote
}

// VotingResultsCached represents cached voting results in the database
type VotingResultsCached struct {
	ID                        int64       `json:"id"`
	GatheringID               int64       `json:"gathering_id"`
	ResultsData               VoteResults `json:"results_data"`
	VotingMode                string      `json:"voting_mode"`
	GatheringType             string      `json:"gathering_type"`
	TotalPossibleVotesWeight  float64     `json:"total_possible_votes_weight"`
	TotalPossibleVotesCount   int         `json:"total_possible_votes_count"`
	QuorumThresholdPercentage float64     `json:"quorum_threshold_percentage"`
	QuorumMet                 bool        `json:"quorum_met"`
	ComputedAt                time.Time   `json:"computed_at"`
}

// VotingMatter represents a matter to be voted on
type VotingMatter struct {
	ID           int64        `json:"id"`
	GatheringID  int64        `json:"gathering_id"`
	OrderIndex   int          `json:"order_index"`
	Title        string       `json:"title"`
	Description  string       `json:"description"`
	MatterType   string       `json:"matter_type"`
	VotingConfig VotingConfig `json:"voting_config"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

// VotingConfig contains the configuration for a voting matter
type VotingConfig struct {
	Type                    string         `json:"type"` // yes_no, multiple_choice, ranking
	Options                 []VotingOption `json:"options,omitempty"`
	RequiredMajority        string         `json:"required_majority"` // simple, supermajority, custom
	RequiredMajorityValue   float64        `json:"required_majority_value,omitempty"`
	Quorum                  float64        `json:"quorum"`
	AllowAbstention         bool           `json:"allow_abstention"`
	IsAnonymous             bool           `json:"is_anonymous"`
	ShowResultsDuringVoting bool           `json:"show_results_during_voting"`
}

// VotingOption represents an option in a multiple choice vote
type VotingOption struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// GatheringParticipant represents a participant in a gathering
type GatheringParticipant struct {
	ID                        int64      `json:"id"`
	GatheringID               int64      `json:"gathering_id"`
	ParticipantType           string     `json:"participant_type"`
	ParticipantName           string     `json:"participant_name"`
	ParticipantIdentification string     `json:"participant_identification"`
	OwnerID                   *int64     `json:"owner_id"`
	DelegatingOwnerID         *int64     `json:"delegating_owner_id"`
	DelegationDocumentRef     string     `json:"delegation_document_ref"`
	UnitsInfo                 []int64    `json:"units_info"`
	UnitsPart                 float64    `json:"units_part"`
	UnitsArea                 float64    `json:"units_area"`
	CheckInTime               *time.Time `json:"check_in_time"`
	HasVoted                  bool       `json:"has_voted"`
	CreatedAt                 time.Time  `json:"created_at"`
	UpdatedAt                 time.Time  `json:"updated_at"`
}

// Ballot represents a submitted ballot
type Ballot struct {
	ID                 int64                 `json:"id"`
	GatheringID        int64                 `json:"gathering_id"`
	ParticipantID      int64                 `json:"participant_id"`
	BallotContent      map[string]BallotVote `json:"ballot_content"`
	BallotHash         string                `json:"ballot_hash"`
	SubmittedAt        time.Time             `json:"submitted_at"`
	SubmittedIP        string                `json:"submitted_ip"`
	SubmittedUserAgent string                `json:"submitted_user_agent"`
	IsValid            bool                  `json:"is_valid"`
}

// BallotVote represents a vote on a single matter
type BallotVote struct {
	MatterID  int64  `json:"matter_id"`
	OptionID  string `json:"option_id,omitempty"`
	VoteValue string `json:"vote_value"` // yes, no, abstain, or option ID
}

// VoteResults represents the results of all voting matters
type VoteResults struct {
	GatheringID int64              `json:"gathering_id"`
	Results     []VoteMatterResult `json:"results"`
	Summary     GatheringSummary   `json:"statistics"` // JSON tag 'statistics' for frontend compatibility
	GeneratedAt string             `json:"generated_at"`
}

// VoteMatterResult represents the result for a single voting matter
type VoteMatterResult struct {
	MatterID     int64            `json:"matter_id"`
	MatterTitle  string           `json:"matter_title"`
	MatterType   string           `json:"matter_type"`
	VotingConfig VotingConfig     `json:"voting_config"`
	Votes        []VoteResult     `json:"votes"`
	Statistics   MatterStatistics `json:"statistics"`
	QuorumInfo   *QuorumInfo      `json:"quorum_info,omitempty"` // Detailed quorum information
	Result       string           `json:"result"`
	IsPassed     bool             `json:"is_passed"`
	// Keep internal fields for calculations
	Tally          map[string]TallyResult `json:"-"`
	TotalVoted     float64                `json:"-"`
	TotalAbstained float64                `json:"-"`
}

// VoteResult represents voting results for a specific choice
type VoteResult struct {
	Choice           string  `json:"choice"`
	VoteCount        int     `json:"vote_count"`
	WeightSum        float64 `json:"weight_sum"`
	Percentage       float64 `json:"percentage"`
	WeightPercentage float64 `json:"weight_percentage"`
}

// TallyResult holds the tallied votes for an option
type TallyResult struct {
	Count      int     `json:"count"`
	Weight     float64 `json:"weight"`
	Area       float64 `json:"area"`
	Percentage float64 `json:"percentage"`
}

// MatterStatistics holds statistics for a voting matter
type MatterStatistics struct {
	TotalParticipants int     `json:"total_participants"`
	TotalVotes        int     `json:"total_votes"`
	TotalWeight       float64 `json:"total_weight"`
	Abstentions       int     `json:"abstentions"`
	ParticipationRate float64 `json:"participation_rate"`
}

// GatheringSummary holds summary statistics for a gathering
type GatheringSummary struct {
	QualifiedUnits       int     `json:"qualified_units"`
	QualifiedWeight      float64 `json:"qualified_weight"`
	QualifiedArea        float64 `json:"qualified_area"`
	ParticipatingUnits   int     `json:"participating_units"`
	ParticipatingWeight  float64 `json:"participating_weight"`
	ParticipatingArea    float64 `json:"participating_area"`
	VotedUnits           int     `json:"voted_units"`
	VotedWeight          float64 `json:"voted_weight"`
	VotedArea            float64 `json:"voted_area"`
	ParticipationRate    float64 `json:"participation_rate"`
	VotingCompletionRate float64 `json:"voting_completion_rate"`
	VotingMode           string  `json:"voting_mode,omitempty"` // by_weight or by_unit
}

// Mapper functions from database models to domain models

// DBGatheringToResponse converts a database Gathering to a response Gathering
func DBGatheringToResponse(g database.Gathering) Gathering {
	var unitTypes []string
	var floors []int64
	var entrances []int64

	if g.QualificationUnitTypes.Valid {
		json.Unmarshal([]byte(g.QualificationUnitTypes.String), &unitTypes)
	}
	if g.QualificationFloors.Valid {
		json.Unmarshal([]byte(g.QualificationFloors.String), &floors)
	}
	if g.QualificationEntrances.Valid {
		json.Unmarshal([]byte(g.QualificationEntrances.String), &entrances)
	}

	return Gathering{
		ID:                          g.ID,
		AssociationID:               g.AssociationID,
		Title:                       g.Title,
		Description:                 g.Description,
		Intent:                      g.Intent,
		Location:                    g.Location,
		GatheringDate:               g.GatheringDate,
		GatheringType:               g.GatheringType,
		VotingMode:                  g.VotingMode,
		Status:                      g.Status,
		QualificationUnitTypes:      unitTypes,
		QualificationFloors:         floors,
		QualificationEntrances:      entrances,
		QualificationCustomRule:     g.QualificationCustomRule.String,
		QualifiedUnitsCount:         int(g.QualifiedUnitsCount.Int64),
		QualifiedUnitsTotalPart:     g.QualifiedUnitsTotalPart.Float64,
		QualifiedUnitsTotalArea:     g.QualifiedUnitsTotalArea.Float64,
		ParticipatingUnitsCount:     int(g.ParticipatingUnitsCount.Int64),
		ParticipatingUnitsTotalPart: g.ParticipatingUnitsTotalPart.Float64,
		ParticipatingUnitsTotalArea: g.ParticipatingUnitsTotalArea.Float64,
		CreatedAt:                   g.CreatedAt.Time,
		UpdatedAt:                   g.UpdatedAt.Time,
	}
}

// DBVotingMatterToResponse converts a database VotingMatter to a response VotingMatter
func DBVotingMatterToResponse(m database.VotingMatter) VotingMatter {
	var config VotingConfig
	json.Unmarshal([]byte(m.VotingConfig), &config)

	return VotingMatter{
		ID:           m.ID,
		GatheringID:  m.GatheringID,
		OrderIndex:   int(m.OrderIndex),
		Title:        m.Title,
		Description:  m.Description.String,
		MatterType:   m.MatterType,
		VotingConfig: config,
		CreatedAt:    m.CreatedAt.Time,
		UpdatedAt:    m.UpdatedAt.Time,
	}
}

// DBParticipantToResponse converts a database GatheringParticipant to a response GatheringParticipant
func DBParticipantToResponse(p database.GatheringParticipant) GatheringParticipant {
	var unitsInfo []int64
	json.Unmarshal([]byte(p.UnitsInfo), &unitsInfo)

	return GatheringParticipant{
		ID:                        p.ID,
		GatheringID:               p.GatheringID,
		ParticipantType:           p.ParticipantType,
		ParticipantName:           p.ParticipantName,
		ParticipantIdentification: p.ParticipantIdentification.String,
		OwnerID:                   NullInt64ToPtr(p.OwnerID),
		DelegatingOwnerID:         NullInt64ToPtr(p.DelegatingOwnerID),
		DelegationDocumentRef:     p.DelegationDocumentRef.String,
		UnitsInfo:                 unitsInfo,
		UnitsPart:                 p.UnitsPart,
		UnitsArea:                 p.UnitsArea,
		CheckInTime:               NullTimeToPtr(p.CheckInTime),
		CreatedAt:                 p.CreatedAt.Time,
		UpdatedAt:                 p.UpdatedAt.Time,
	}
}

// DBParticipantRowToResponse converts a database GetGatheringParticipantsRow to a response GatheringParticipant
func DBParticipantRowToResponse(p database.GetGatheringParticipantsRow) GatheringParticipant {
	var unitsInfo []int64
	json.Unmarshal([]byte(p.UnitsInfo), &unitsInfo)

	return GatheringParticipant{
		ID:                        p.ID,
		GatheringID:               p.GatheringID,
		ParticipantType:           p.ParticipantType,
		ParticipantName:           p.ParticipantName,
		ParticipantIdentification: p.ParticipantIdentification.String,
		OwnerID:                   NullInt64ToPtr(p.OwnerID),
		DelegatingOwnerID:         NullInt64ToPtr(p.DelegatingOwnerID),
		DelegationDocumentRef:     p.DelegationDocumentRef.String,
		UnitsInfo:                 unitsInfo,
		UnitsPart:                 p.UnitsPart,
		UnitsArea:                 p.UnitsArea,
		CheckInTime:               NullTimeToPtr(p.CheckInTime),
		CreatedAt:                 p.CreatedAt.Time,
		UpdatedAt:                 p.UpdatedAt.Time,
	}
}

// Helper functions

// NullInt64ToPtr converts sql.NullInt64 to *int64
func NullInt64ToPtr(n sql.NullInt64) *int64 {
	if n.Valid {
		return &n.Int64
	}
	return nil
}

// NullTimeToPtr converts sql.NullTime to *time.Time
func NullTimeToPtr(n sql.NullTime) *time.Time {
	if n.Valid {
		return &n.Time
	}
	return nil
}
