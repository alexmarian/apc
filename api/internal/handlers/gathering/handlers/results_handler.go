package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/alexmarian/apc/api/internal/database"
	"github.com/alexmarian/apc/api/internal/handlers"
	"github.com/alexmarian/apc/api/internal/handlers/gathering/domain"
	"github.com/alexmarian/apc/api/internal/handlers/gathering/services"
	"github.com/alexmarian/apc/api/internal/logging"
	"go.uber.org/zap"
)

// ResultsHandler handles results and statistics operations
type ResultsHandler struct {
	cfg                  *handlers.ApiConfig
	quorumService        *services.QuorumService
	tallyService         *services.TallyService
	votingResultsService *services.VotingResultsService
}

// NewResultsHandler creates a new ResultsHandler
func NewResultsHandler(cfg *handlers.ApiConfig) *ResultsHandler {
	quorumService := services.NewQuorumService(cfg.Db)
	tallyService := services.NewTallyService(cfg.Db)

	return &ResultsHandler{
		cfg:                  cfg,
		quorumService:        quorumService,
		tallyService:         tallyService,
		votingResultsService: services.NewVotingResultsService(cfg.Db, quorumService, tallyService),
	}
}

// HandleGetVoteResults returns the voting results for a gathering
func (h *ResultsHandler) HandleGetVoteResults() func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationID, _ := strconv.Atoi(req.PathValue(handlers.AssociationIdPathValue))
		gatheringID, _ := strconv.Atoi(req.PathValue(domain.GatheringIDPathValue))

		// Use VotingResultsService to get cached or fresh results
		results, err := h.votingResultsService.GetCachedResults(req.Context(), int64(gatheringID), int64(associationID))
		if err != nil {
			logging.Logger.Log(zap.ErrorLevel, "Error getting voting results",
				zap.Int("gathering_id", gatheringID),
				zap.Error(err))
			handlers.RespondWithError(rw, http.StatusInternalServerError, "Failed to get voting results")
			return
		}

		handlers.RespondWithJSON(rw, http.StatusOK, results)
	}
}

// HandleGetGatheringStats returns statistics for a gathering
func (h *ResultsHandler) HandleGetGatheringStats() func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationID, _ := strconv.Atoi(req.PathValue(handlers.AssociationIdPathValue))
		gatheringID, _ := strconv.Atoi(req.PathValue(domain.GatheringIDPathValue))

		// Get gathering details
		gathering, err := h.cfg.Db.GetGatheringStats(req.Context(), database.GetGatheringStatsParams{
			ID:            int64(gatheringID),
			AssociationID: int64(associationID),
		})
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting gathering", zap.Error(err))
			handlers.RespondWithError(rw, http.StatusInternalServerError, "Failed to get gathering stats")
			return
		}

		// Get participant count
		participantCount, err := h.cfg.Db.GetGatheringParticipantCount(req.Context(), int64(gatheringID))
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting participant count", zap.Error(err))
		}

		// Get ballot count
		ballotCount, err := h.cfg.Db.GetGatheringBallotCount(req.Context(), int64(gatheringID))
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting ballot count", zap.Error(err))
		}

		// Calculate voted weight by getting all participants who voted
		participants, _ := h.cfg.Db.GetGatheringParticipants(req.Context(), int64(gatheringID))
		ballots, _ := h.cfg.Db.GetBallotsForGathering(req.Context(), int64(gatheringID))

		// Create a map of participant IDs who voted
		votedParticipants := make(map[int64]bool)
		for _, ballot := range ballots {
			if ballot.IsValid.Bool {
				votedParticipants[ballot.ParticipantID] = true
			}
		}

		// Calculate total voted weight
		totalVotedWeight := 0.0
		totalVotedArea := 0.0
		for _, participant := range participants {
			if votedParticipants[participant.ID] {
				totalVotedWeight += participant.UnitsPart
				totalVotedArea += participant.UnitsArea
			}
		}

		type GatheringStats struct {
			domain.Gathering
			ParticipantCount        int     `json:"participant_count"`
			VotedCount              int     `json:"voted_count"`
			TotalVotedWeight        float64 `json:"total_voted_weight"`
			TotalVotedArea          float64 `json:"total_voted_area"`
			ParticipationPercentage float64 `json:"participation_percentage"`
			VotingPercentage        float64 `json:"voting_percentage"`
		}

		response := GatheringStats{
			Gathering:        domain.DBGatheringToResponse(gathering),
			ParticipantCount: int(participantCount),
			VotedCount:       int(ballotCount),
			TotalVotedWeight: totalVotedWeight,
			TotalVotedArea:   totalVotedArea,
		}

		// Calculate percentages
		if gathering.QualifiedUnitsCount.Int64 > 0 {
			response.ParticipationPercentage = float64(participantCount) / float64(gathering.QualifiedUnitsCount.Int64) * 100
		}
		if participantCount > 0 {
			response.VotingPercentage = float64(ballotCount) / float64(participantCount) * 100
		}

		handlers.RespondWithJSON(rw, http.StatusOK, response)
	}
}

// HandleGetEligibleVoters returns eligible voters for a gathering
func (h *ResultsHandler) HandleGetEligibleVoters() func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationID, _ := strconv.Atoi(req.PathValue(handlers.AssociationIdPathValue))
		gatheringID, _ := strconv.Atoi(req.PathValue(domain.GatheringIDPathValue))

		rows, err := h.cfg.Db.GetEligibleVotersWithUnits(req.Context(), database.GetEligibleVotersWithUnitsParams{
			GatheringID:   int64(gatheringID),
			AssociationID: int64(associationID),
		})

		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting eligible voters", zap.Error(err))
			handlers.RespondWithError(rw, http.StatusInternalServerError, "Failed to get eligible voters")
			return
		}

		type VoterUnit struct {
			ID              int64   `json:"id"`
			UnitNumber      string  `json:"unit_number"`
			CadastralNumber string  `json:"cadastral_number"`
			Floor           int64   `json:"floor"`
			Entrance        int64   `json:"entrance"`
			Area            float64 `json:"area"`
			VotingWeight    float64 `json:"voting_weight"`
			UnitType        string  `json:"unit_type"`
			BuildingName    string  `json:"building_name"`
			BuildingAddress string  `json:"building_address"`
			IsAvailable     bool    `json:"is_available"`
		}

		type EligibleVoter struct {
			Owner                database.Owner `json:"owner"`
			Units                []VoterUnit    `json:"units"`
			TotalAvailableWeight float64        `json:"total_available_weight"`
			TotalAvailableArea   float64        `json:"total_available_area"`
			TotalWeight          float64        `json:"total_weight"`
			TotalArea            float64        `json:"total_area"`
			HasAvailableUnits    bool           `json:"has_available_units"`
			AvailableUnitsCount  int            `json:"available_units_count"`
		}

		votersMap := make(map[int64]*EligibleVoter)

		for _, row := range rows {
			voter, exists := votersMap[row.OwnerID]
			if !exists {
				voter = &EligibleVoter{
					Owner: database.Owner{
						ID:                   row.OwnerID,
						Name:                 row.OwnerName,
						IdentificationNumber: row.OwnerIdentification,
						ContactEmail:         row.OwnerContactEmail,
						ContactPhone:         row.OwnerContactPhone,
					},
					Units: make([]VoterUnit, 0),
				}
				votersMap[row.OwnerID] = voter
			}

			isAvailable := row.IsAvailable == 1
			unit := VoterUnit{
				ID:              row.UnitID,
				UnitNumber:      row.UnitNumber,
				CadastralNumber: row.CadastralNumber,
				Floor:           row.Floor,
				Entrance:        row.Entrance,
				Area:            row.Area,
				VotingWeight:    row.VotingWeight,
				UnitType:        row.UnitType,
				BuildingName:    row.BuildingName,
				BuildingAddress: row.BuildingAddress,
				IsAvailable:     isAvailable,
			}

			voter.Units = append(voter.Units, unit)
			voter.TotalWeight += row.VotingWeight
			voter.TotalArea += row.Area

			if isAvailable {
				voter.TotalAvailableWeight += row.VotingWeight
				voter.TotalAvailableArea += row.Area
				voter.HasAvailableUnits = true
				voter.AvailableUnitsCount++
			}
		}

		response := make([]EligibleVoter, 0, len(votersMap))
		for _, voter := range votersMap {
			response = append(response, *voter)
		}

		handlers.RespondWithJSON(rw, http.StatusOK, response)
	}
}

// HandleGetQualifiedUnits returns qualified units for a gathering
func (h *ResultsHandler) HandleGetQualifiedUnits() func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationID, _ := strconv.Atoi(req.PathValue(handlers.AssociationIdPathValue))
		gatheringID, _ := strconv.Atoi(req.PathValue(domain.GatheringIDPathValue))

		// Get gathering to check qualification rules
		gathering, err := h.cfg.Db.GetGathering(req.Context(), database.GetGatheringParams{
			ID:            int64(gatheringID),
			AssociationID: int64(associationID),
		})
		if err != nil {
			handlers.RespondWithError(rw, http.StatusNotFound, "Gathering not found")
			return
		}

		// Get qualified units with owner information
		units, err := h.cfg.Db.GetActiveOwnerUnitsForGathering(req.Context(), domain.QualificationOf(gathering).OwnerUnitsParams(int64(associationID)))

		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting qualified units", zap.Error(err))
			handlers.RespondWithError(rw, http.StatusInternalServerError, "Failed to get qualified units")
			return
		}

		type QualifiedUnit struct {
			ID              int64   `json:"id"`
			UnitNumber      string  `json:"unit_number"`
			CadastralNumber string  `json:"cadastral_number"`
			Floor           int     `json:"floor"`
			Entrance        int     `json:"entrance"`
			Area            float64 `json:"area"`
			Part            float64 `json:"part"`
			UnitType        string  `json:"unit_type"`
			BuildingName    string  `json:"building_name"`
			BuildingAddress string  `json:"building_address"`
			IsParticipating bool    `json:"is_participating"`
			OwnerID         int64   `json:"owner_id"`
			OwnerName       string  `json:"owner_name"`
		}

		// Get current participants to mark which units are already participating
		participants, _ := h.cfg.Db.GetGatheringParticipants(req.Context(), int64(gatheringID))
		participatingUnits := make(map[int64]bool)
		for _, p := range participants {
			partUnits := make([]int64, 0)
			json.Unmarshal([]byte(p.UnitsInfo), &partUnits)
			for _, unitID := range partUnits {
				participatingUnits[unitID] = true
			}
		}

		response := make([]QualifiedUnit, len(units))
		for i, u := range units {
			response[i] = QualifiedUnit{
				ID:              u.UnitID,
				UnitNumber:      u.UnitNumber,
				CadastralNumber: u.CadastralNumber,
				Floor:           int(u.Floor),
				Entrance:        int(u.Entrance),
				Area:            u.Area,
				Part:            u.VotingWeight,
				UnitType:        u.UnitType,
				BuildingName:    u.BuildingName,
				BuildingAddress: u.BuildingAddress,
				IsParticipating: participatingUnits[u.UnitID],
				OwnerID:         u.OwnerID,
				OwnerName:       u.OwnerName,
			}
		}

		handlers.RespondWithJSON(rw, http.StatusOK, response)
	}
}

// HandleGetNonParticipatingOwners returns owners who haven't participated yet
func (h *ResultsHandler) HandleGetNonParticipatingOwners() func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationID, _ := strconv.Atoi(req.PathValue(handlers.AssociationIdPathValue))
		gatheringID, _ := strconv.Atoi(req.PathValue(domain.GatheringIDPathValue))

		// Get gathering qualification rules
		gathering, err := h.cfg.Db.GetGathering(req.Context(), database.GetGatheringParams{
			ID:            int64(gatheringID),
			AssociationID: int64(associationID),
		})
		if err != nil {
			handlers.RespondWithError(rw, http.StatusNotFound, "Gathering not found")
			return
		}

		owners, err := h.cfg.Db.GetNonParticipatingOwners(req.Context(), domain.QualificationOf(gathering).NonParticipatingParams(int64(associationID), int64(gatheringID)))

		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting non-participating owners", zap.Error(err))
			handlers.RespondWithError(rw, http.StatusInternalServerError, "Failed to get non-participating owners")
			return
		}

		type NonParticipatingOwner struct {
			ID                   int64  `json:"id"`
			Name                 string `json:"name"`
			IdentificationNumber string `json:"identification_number"`
			ContactEmail         string `json:"contact_email"`
			ContactPhone         string `json:"contact_phone"`
			UnitsCount           int    `json:"units_count"`
		}

		response := make([]NonParticipatingOwner, len(owners))
		for i, o := range owners {
			response[i] = NonParticipatingOwner{
				ID:                   o.ID,
				Name:                 o.Name,
				IdentificationNumber: o.IdentificationNumber,
				ContactEmail:         o.ContactEmail,
				ContactPhone:         o.ContactPhone,
				UnitsCount:           int(o.UnitsCount),
			}
		}

		handlers.RespondWithJSON(rw, http.StatusOK, response)
	}
}
