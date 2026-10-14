package handlers

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/alexmarian/apc/api/internal/database"
	"github.com/alexmarian/apc/api/internal/logging"
	"go.uber.org/zap"
)

const MemberTokenPathValue = "memberToken"

type memberContextKey string

const memberInvitationKey memberContextKey = "memberInvitation"

// MiddlewareMemberToken validates the opaque token from the URL path and injects
// the resolved MemberInvitation into the request context.
func (cfg *ApiConfig) MiddlewareMemberToken(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.PathValue(MemberTokenPathValue)
		if token == "" {
			RespondWithError(w, http.StatusUnauthorized, "missing token")
			return
		}

		hash := sha256.Sum256([]byte(token))
		tokenHash := hex.EncodeToString(hash[:])

		inv, err := cfg.Db.GetMemberInvitationByTokenHash(r.Context(), tokenHash)
		if err != nil {
			if err == sql.ErrNoRows {
				RespondWithError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}
			logging.Logger.Log(zap.WarnLevel, "token lookup failed", zap.Error(err))
			RespondWithError(w, http.StatusInternalServerError, "token validation failed")
			return
		}

		ctx := context.WithValue(r.Context(), memberInvitationKey, inv)
		handler(w, r.WithContext(ctx))
	}
}

func memberInvitationFromContext(ctx context.Context) (database.MemberInvitation, bool) {
	inv, ok := ctx.Value(memberInvitationKey).(database.MemberInvitation)
	return inv, ok
}

// MemberInvitationFromContext is the exported version for use by sub-packages.
func MemberInvitationFromContext(ctx context.Context) (database.MemberInvitation, bool) {
	return memberInvitationFromContext(ctx)
}

// MemberGatheringResponse is the shape returned to member clients.
type MemberGatheringResponse struct {
	Gathering memberGatheringInfo `json:"gathering"`
	Owner     memberOwnerInfo     `json:"owner"`
	Units     []memberUnitInfo    `json:"units"`
	Matters   []memberMatterInfo  `json:"matters"`
	Ballot    *memberBallotInfo   `json:"ballot"`
	Results   json.RawMessage     `json:"results"`
}

type memberMatterInfo struct {
	ID            int64           `json:"id"`
	Title         string          `json:"title"`
	TitleRu       string          `json:"title_ru"`
	Description   string          `json:"description"`
	DescriptionRu string          `json:"description_ru"`
	MatterType    string          `json:"matter_type"`
	OrderIndex    int64           `json:"order_index"`
	VotingConfig  json.RawMessage `json:"voting_config"`
	IsInformative bool            `json:"is_informative"`
}

type memberGatheringInfo struct {
	ID                      int64     `json:"id"`
	Title                   string    `json:"title"`
	Description             string    `json:"description"`
	GatheringDate           time.Time `json:"gathering_date"`
	GatheringType           string    `json:"gathering_type"`
	Status                  string    `json:"status"`
	VotingMode              string    `json:"voting_mode"`
	QualifiedUnitsCount     int64     `json:"qualified_units_count"`
	QualifiedUnitsTotalPart float64   `json:"qualified_units_total_part"`
}

type memberOwnerInfo struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	Identification string `json:"identification"`
}

type memberUnitInfo struct {
	UnitID       int64   `json:"unit_id"`
	UnitNumber   string  `json:"unit_number"`
	UnitType     string  `json:"unit_type"`
	BuildingName string  `json:"building_name"`
	Area         float64 `json:"area"`
	VotingWeight float64 `json:"voting_weight"`
	IsAvailable  bool    `json:"is_available"`
}

type memberBallotInfo struct {
	BallotID      int64           `json:"ballot_id"`
	BallotHash    string          `json:"ballot_hash"`
	SubmittedAt   *time.Time      `json:"submitted_at"`
	BallotContent json.RawMessage `json:"ballot_content"`
	IsValid       bool            `json:"is_valid"`
}

// HandleGetMemberContext is the single token-scoped read endpoint for the member app.
// Route: GET /v1/api/member/gatherings/{memberToken}
func HandleGetMemberContext(cfg *ApiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		inv, ok := memberInvitationFromContext(r.Context())
		if !ok {
			RespondWithError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		gathering, err := cfg.Db.GetGatheringByID(r.Context(), inv.GatheringID)
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "failed to get gathering", zap.Error(err))
			RespondWithError(w, http.StatusInternalServerError, "failed to load gathering")
			return
		}
		response, err := MemberContext(r.Context(), cfg, gathering, inv.OwnerID)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "failed to load gathering")
			return
		}
		RespondWithJSON(w, http.StatusOK, response)
	}
}

// MemberContext is what an owner sees of a gathering — from a member link or
// their own account: their units in it (and which can still vote), the
// agenda, their ballot if they voted, and the results once tallied.
func MemberContext(ctx context.Context, cfg *ApiConfig, gathering database.Gathering, ownerID int64) (MemberGatheringResponse, error) {
	owner, err := cfg.Db.GetOwnerById(ctx, ownerID)
	if err != nil {
		logging.Logger.Log(zap.WarnLevel, "failed to get owner", zap.Error(err))
		return MemberGatheringResponse{}, err
	}

	eligibleRows, err := cfg.Db.GetEligibleVotersWithUnits(ctx, database.GetEligibleVotersWithUnitsParams{
		GatheringID:   gathering.ID,
		AssociationID: gathering.AssociationID,
	})
	if err != nil {
		logging.Logger.Log(zap.WarnLevel, "failed to get eligible units", zap.Error(err))
		return MemberGatheringResponse{}, err
	}

	units := make([]memberUnitInfo, 0)
	for _, row := range eligibleRows {
		if row.OwnerID == ownerID {
			units = append(units, memberUnitInfo{
				UnitID:       row.UnitID,
				UnitNumber:   row.UnitNumber,
				UnitType:     row.UnitType,
				BuildingName: row.BuildingName,
				Area:         row.Area,
				VotingWeight: row.VotingWeight,
				IsAvailable:  row.IsAvailable == 1,
			})
		}
	}

	var ballotInfo *memberBallotInfo
	participant, err := cfg.Db.GetGatheringParticipantByOwner(ctx, database.GetGatheringParticipantByOwnerParams{
		GatheringID: gathering.ID,
		OwnerID:     sql.NullInt64{Int64: ownerID, Valid: true},
	})
	if err == nil {
		ballot, err := cfg.Db.GetBallotByParticipant(ctx, database.GetBallotByParticipantParams{
			GatheringID:   gathering.ID,
			ParticipantID: participant.ID,
		})
		if err == nil {
			var submittedAt *time.Time
			if ballot.SubmittedAt.Valid {
				submittedAt = &ballot.SubmittedAt.Time
			}
			ballotInfo = &memberBallotInfo{
				BallotID:      ballot.ID,
				BallotHash:    ballot.BallotHash,
				SubmittedAt:   submittedAt,
				BallotContent: json.RawMessage(ballot.BallotContent),
				IsValid:       ballot.IsValid.Bool,
			}
		}
	}

	dbMatters, err := cfg.Db.GetVotingMatters(ctx, gathering.ID)
	if err != nil {
		logging.Logger.Log(zap.WarnLevel, "failed to get voting matters", zap.Error(err))
		// Non-fatal — return empty matters rather than failing the whole request
		dbMatters = nil
	}
	matterInfos := make([]memberMatterInfo, 0, len(dbMatters))
	for _, m := range dbMatters {
		matterInfos = append(matterInfos, memberMatterInfo{
			ID:            m.ID,
			Title:         m.Title,
			TitleRu:       m.TitleRu,
			Description:   m.Description.String,
			DescriptionRu: m.DescriptionRu.String,
			MatterType:    m.MatterType,
			OrderIndex:    m.OrderIndex,
			VotingConfig:  json.RawMessage(m.VotingConfig),
			IsInformative: m.IsInformative != 0,
		})
	}

	var resultsRaw json.RawMessage
	if gathering.Status == "tallied" {
		result, err := cfg.Db.GetVotingResults(ctx, gathering.ID)
		if err == nil {
			resultsRaw = json.RawMessage(result.ResultsData)
		}
	}

	return MemberGatheringResponse{
		Gathering: memberGatheringInfo{
			ID:                      gathering.ID,
			Title:                   gathering.Title,
			Description:             gathering.Description,
			GatheringDate:           gathering.GatheringDate,
			GatheringType:           gathering.GatheringType,
			Status:                  gathering.Status,
			VotingMode:              gathering.VotingMode,
			QualifiedUnitsCount:     gathering.QualifiedUnitsCount.Int64,
			QualifiedUnitsTotalPart: gathering.QualifiedUnitsTotalPart.Float64,
		},
		Owner: memberOwnerInfo{
			ID:             owner.ID,
			Name:           owner.Name,
			Identification: owner.IdentificationNumber,
		},
		Units:   units,
		Matters: matterInfos,
		Ballot:  ballotInfo,
		Results: resultsRaw,
	}, nil
}
