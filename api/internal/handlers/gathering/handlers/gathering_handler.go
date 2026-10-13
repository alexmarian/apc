package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/alexmarian/apc/api/internal/database"
	"github.com/alexmarian/apc/api/internal/handlers"
	"github.com/alexmarian/apc/api/internal/handlers/gathering/domain"
	"github.com/alexmarian/apc/api/internal/handlers/gathering/services"
	"github.com/alexmarian/apc/api/internal/logging"
	"go.uber.org/zap"
)

// GatheringHandler handles gathering CRUD operations
type GatheringHandler struct {
	cfg                  *handlers.ApiConfig
	statsService         *services.StatsService
	unitSlotService      *services.UnitSlotService
	quorumService        *services.QuorumService
	tallyService         *services.TallyService
	votingResultsService *services.VotingResultsService
}

// NewGatheringHandler creates a new GatheringHandler
func NewGatheringHandler(cfg *handlers.ApiConfig) *GatheringHandler {
	quorumService := services.NewQuorumService(cfg.Db)
	tallyService := services.NewTallyService(cfg.Db)

	return &GatheringHandler{
		cfg:                  cfg,
		statsService:         services.NewStatsService(cfg.Db),
		unitSlotService:      services.NewUnitSlotService(cfg.Db),
		quorumService:        quorumService,
		tallyService:         tallyService,
		votingResultsService: services.NewVotingResultsService(cfg.Db, quorumService, tallyService),
	}
}

// HandleGetGatherings returns all gatherings for an association
func (h *GatheringHandler) HandleGetGatherings() func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationID, _ := strconv.Atoi(req.PathValue(handlers.AssociationIdPathValue))

		gatherings, err := h.cfg.Db.GetGatherings(req.Context(), int64(associationID))
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting gatherings", zap.Error(err))
			handlers.RespondWithError(rw, http.StatusInternalServerError, "Failed to get gatherings")
			return
		}

		response := make([]domain.Gathering, len(gatherings))
		for i, g := range gatherings {
			response[i] = domain.DBGatheringToResponse(g)
		}

		handlers.RespondWithJSON(rw, http.StatusOK, response)
	}
}

// HandleGetGathering returns a single gathering by ID
func (h *GatheringHandler) HandleGetGathering() func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationID, _ := strconv.Atoi(req.PathValue(handlers.AssociationIdPathValue))
		gatheringID, _ := strconv.Atoi(req.PathValue(domain.GatheringIDPathValue))

		gathering, err := h.cfg.Db.GetGathering(req.Context(), database.GetGatheringParams{
			ID:            int64(gatheringID),
			AssociationID: int64(associationID),
		})
		if err != nil {
			if err == sql.ErrNoRows {
				handlers.RespondWithError(rw, http.StatusNotFound, "Gathering not found")
			} else {
				logging.Logger.Log(zap.WarnLevel, "Error getting gathering", zap.Error(err))
				handlers.RespondWithError(rw, http.StatusInternalServerError, "Failed to get gathering")
			}
			return
		}

		handlers.RespondWithJSON(rw, http.StatusOK, domain.DBGatheringToResponse(gathering))
	}
}

// parseGatheringRequest reads and validates a create/update body (title,
// type and location required; voting mode defaults to by_weight). On
// failure it has answered 400.
func parseGatheringRequest(rw http.ResponseWriter, req *http.Request) (domain.CreateGatheringRequest, bool) {
	var r domain.CreateGatheringRequest
	if err := json.NewDecoder(req.Body).Decode(&r); err != nil {
		handlers.RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
		return r, false
	}
	r.Title = strings.TrimSpace(r.Title)
	r.Location = strings.TrimSpace(r.Location)
	if r.Title == "" || r.GatheringType == "" {
		handlers.RespondWithError(rw, http.StatusBadRequest, "Title and gathering type are required")
		return r, false
	}
	if r.GatheringType != "initial" && r.GatheringType != "repeated" && r.GatheringType != "remote" {
		handlers.RespondWithError(rw, http.StatusBadRequest, "gathering_type must be 'initial', 'repeated' or 'remote'")
		return r, false
	}
	if r.Location == "" {
		handlers.RespondWithError(rw, http.StatusBadRequest, "Location is required")
		return r, false
	}
	if r.VotingMode == "" {
		r.VotingMode = "by_weight" // Default for backward compatibility
	}
	if r.VotingMode != "by_weight" && r.VotingMode != "by_unit" {
		handlers.RespondWithError(rw, http.StatusBadRequest, "voting_mode must be 'by_weight' or 'by_unit'")
		return r, false
	}
	return r, true
}

// checkBuildings de-duplicates the request's buildings and makes sure each
// is one of the association's (400 otherwise).
func (h *GatheringHandler) checkBuildings(rw http.ResponseWriter, req *http.Request, associationID int64, r *domain.CreateGatheringRequest) bool {
	if len(r.QualificationBuildings) == 0 {
		return true
	}
	buildings, err := h.cfg.Db.GetAssociationBuildings(req.Context(), associationID)
	if err != nil {
		handlers.RespondWithError(rw, http.StatusInternalServerError, "Failed to check buildings")
		return false
	}
	own := map[int64]bool{}
	for _, b := range buildings {
		own[b.ID] = true
	}
	seen := map[int64]bool{}
	ids := make([]int64, 0, len(r.QualificationBuildings))
	for _, id := range r.QualificationBuildings {
		if !own[id] {
			handlers.RespondWithError(rw, http.StatusBadRequest, "Unknown building")
			return false
		}
		if !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	r.QualificationBuildings = ids
	return true
}

// qualificationJSON stores a qualification list (NULL when empty: no restriction).
func qualificationJSON[T any](values []T) sql.NullString {
	b, _ := json.Marshal(values)
	return sql.NullString{String: string(b), Valid: len(values) > 0}
}

// HandleCreateGathering creates a new gathering
func (h *GatheringHandler) HandleCreateGathering() func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationID, _ := strconv.Atoi(req.PathValue(handlers.AssociationIdPathValue))

		createReq, ok := parseGatheringRequest(rw, req)
		if !ok || !h.checkBuildings(rw, req, int64(associationID), &createReq) {
			return
		}
		votingMode := createReq.VotingMode

		gathering, err := h.cfg.Db.CreateGathering(req.Context(), database.CreateGatheringParams{
			AssociationID:           int64(associationID),
			Title:                   createReq.Title,
			Description:             createReq.Description,
			Intent:                  createReq.Intent,
			Location:                createReq.Location,
			GatheringDate:           createReq.GatheringDate,
			GatheringType:           createReq.GatheringType,
			VotingMode:              votingMode,
			Status:                  "draft",
			QualificationUnitTypes:  qualificationJSON(createReq.QualificationUnitTypes),
			QualificationFloors:     qualificationJSON(createReq.QualificationFloors),
			QualificationEntrances:  qualificationJSON(createReq.QualificationEntrances),
			QualificationBuildings:  qualificationJSON(createReq.QualificationBuildings),
			QualificationCustomRule: sql.NullString{String: createReq.QualificationCustomRule, Valid: createReq.QualificationCustomRule != ""},
		})

		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error creating gathering", zap.Error(err))
			handlers.RespondWithError(rw, http.StatusInternalServerError, "Failed to create gathering")
			return
		}

		// Calculate qualified units
		qualifiedCount, qualifiedPart, qualifiedArea := h.statsService.UpdateGatheringStats(gathering.ID, int64(associationID))

		// Log audit
		h.cfg.Db.CreateAuditLog(req.Context(), database.CreateAuditLogParams{
			GatheringID: gathering.ID,
			EntityType:  "gathering",
			EntityID:    gathering.ID,
			Action:      "created",
			PerformedBy: sql.NullString{String: req.Context().Value("userID").(string), Valid: true},
			IpAddress:   sql.NullString{String: req.RemoteAddr, Valid: true},
			Details:     sql.NullString{String: "{}", Valid: true},
		})

		err = h.unitSlotService.SyncUnitsSlots(req.Context(), int64(associationID), gathering.ID, createReq.Qualification())
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error syncing units slots", zap.Error(err))
			handlers.RespondWithError(rw, http.StatusInternalServerError, "Failed to sync units slots")
			return
		}

		response := domain.DBGatheringToResponse(gathering)
		response.QualifiedUnitsCount = qualifiedCount
		response.QualifiedUnitsTotalPart = qualifiedPart
		response.QualifiedUnitsTotalArea = qualifiedArea
		handlers.RespondWithJSON(rw, http.StatusCreated, response)
	}
}

// HandleUpdateGatheringStatus updates the status of a gathering
func (h *GatheringHandler) HandleUpdateGatheringStatus() func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationID, _ := strconv.Atoi(req.PathValue(handlers.AssociationIdPathValue))
		gatheringID, _ := strconv.Atoi(req.PathValue(domain.GatheringIDPathValue))

		var statusReq struct {
			Status string `json:"status"`
		}
		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&statusReq); err != nil {
			handlers.RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}

		// Validate status transition
		validStatuses := map[string]bool{
			"draft": true, "published": true, "active": true, "closed": true, "tallied": true,
		}
		if !validStatuses[statusReq.Status] {
			handlers.RespondWithError(rw, http.StatusBadRequest, "Invalid status")
			return
		}

		previous, err := h.cfg.Db.GetGathering(req.Context(), database.GetGatheringParams{
			ID:            int64(gatheringID),
			AssociationID: int64(associationID),
		})
		if err != nil {
			handlers.RespondWithError(rw, http.StatusNotFound, "Gathering not found")
			return
		}

		gathering, err := h.cfg.Db.UpdateGatheringStatus(req.Context(), database.UpdateGatheringStatusParams{
			Status:        statusReq.Status,
			ID:            int64(gatheringID),
			AssociationID: int64(associationID),
		})

		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error updating gathering status", zap.Error(err))
			handlers.RespondWithError(rw, http.StatusInternalServerError, "Failed to update gathering status")
			return
		}

		// If closing the gathering, calculate and cache final results. In the
		// background, so not with the request's context (cancelled once we answer).
		if statusReq.Status == "closed" {
			go func() {
				_, err := h.votingResultsService.ComputeAndStoreResults(context.Background(), int64(gatheringID), int64(associationID))
				if err != nil {
					logging.Logger.Log(zap.ErrorLevel, "Error computing final results",
						zap.Int64("gathering_id", int64(gatheringID)),
						zap.Error(err))
				} else {
					logging.Logger.Log(zap.InfoLevel, "Successfully computed and cached voting results",
						zap.Int64("gathering_id", int64(gatheringID)))
				}
			}()
		}

		// Leaving "closed" (e.g. reopening): the cached final results are stale.
		// (This compared the updated row's status before, so it never ran.)
		if previous.Status == "closed" && statusReq.Status != "closed" {
			if err := h.votingResultsService.InvalidateResults(req.Context(), int64(gatheringID)); err != nil {
				logging.Logger.Log(zap.WarnLevel, "Error invalidating cached results",
					zap.Int64("gathering_id", int64(gatheringID)),
					zap.Error(err))
			}
		}

		handlers.RespondWithJSON(rw, http.StatusOK, domain.DBGatheringToResponse(gathering))
	}
}

// ValidateGatheringStateWithFetch fetches a gathering and validates its state
func (h *GatheringHandler) ValidateGatheringStateWithFetch(req *http.Request, gatheringID int, associationID int, targetStatus string) bool {
	gathering, err := h.cfg.Db.GetGathering(req.Context(), database.GetGatheringParams{
		ID:            int64(gatheringID),
		AssociationID: int64(associationID),
	})
	if err != nil {
		logging.Logger.Log(zap.WarnLevel, "Error getting gathering", zap.Error(err))
		return false
	}
	return h.quorumService.ValidateGatheringState(gathering, targetStatus)
}

// HandleUpdateGathering edits a gathering (same body and rules as create)
// while it is draft or published. Qualification changes re-derive the
// qualified units and their voting slots, which is only safe before anyone
// takes part: with participants it answers 409.
func (h *GatheringHandler) HandleUpdateGathering() func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationID, _ := strconv.Atoi(req.PathValue(handlers.AssociationIdPathValue))
		gatheringID, _ := strconv.Atoi(req.PathValue(domain.GatheringIDPathValue))

		current, err := h.cfg.Db.GetGathering(req.Context(), database.GetGatheringParams{
			ID:            int64(gatheringID),
			AssociationID: int64(associationID),
		})
		if err != nil {
			handlers.RespondWithError(rw, http.StatusNotFound, "Gathering not found")
			return
		}
		if current.Status != "draft" && current.Status != "published" {
			handlers.RespondWithError(rw, http.StatusConflict, "Only draft or published gatherings can be edited")
			return
		}
		updateReq, ok := parseGatheringRequest(rw, req)
		if !ok || !h.checkBuildings(rw, req, int64(associationID), &updateReq) {
			return
		}
		participants, err := h.cfg.Db.GetGatheringParticipantCount(req.Context(), int64(gatheringID))
		if err != nil {
			handlers.RespondWithError(rw, http.StatusInternalServerError, "Failed to update gathering")
			return
		}
		if participants > 0 {
			handlers.RespondWithError(rw, http.StatusConflict, "The gathering already has participants")
			return
		}

		tx, err := h.cfg.DB.BeginTx(req.Context(), nil)
		if err != nil {
			handlers.RespondWithError(rw, http.StatusInternalServerError, "Failed to update gathering")
			return
		}
		defer tx.Rollback()
		qtx := h.cfg.Db.WithTx(tx)

		_, err = qtx.UpdateGathering(req.Context(), database.UpdateGatheringParams{
			Title:                       updateReq.Title,
			Description:                 updateReq.Description,
			Intent:                      updateReq.Intent,
			Location:                    updateReq.Location,
			GatheringDate:               updateReq.GatheringDate,
			GatheringType:               updateReq.GatheringType,
			VotingMode:                  updateReq.VotingMode,
			QualificationUnitTypes:      qualificationJSON(updateReq.QualificationUnitTypes),
			QualificationFloors:         qualificationJSON(updateReq.QualificationFloors),
			QualificationEntrances:      qualificationJSON(updateReq.QualificationEntrances),
			QualificationBuildings:      qualificationJSON(updateReq.QualificationBuildings),
			QualificationCustomRule:     sql.NullString{String: updateReq.QualificationCustomRule, Valid: updateReq.QualificationCustomRule != ""},
			QualifiedUnitsCount:         current.QualifiedUnitsCount,
			QualifiedUnitsTotalPart:     current.QualifiedUnitsTotalPart,
			QualifiedUnitsTotalArea:     current.QualifiedUnitsTotalArea,
			ParticipatingUnitsCount:     current.ParticipatingUnitsCount,
			ParticipatingUnitsTotalPart: current.ParticipatingUnitsTotalPart,
			ParticipatingUnitsTotalArea: current.ParticipatingUnitsTotalArea,
			ID:                          int64(gatheringID),
			AssociationID:               int64(associationID),
		})
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error updating gathering", zap.Error(err))
			handlers.RespondWithError(rw, http.StatusInternalServerError, "Failed to update gathering")
			return
		}
		// The voting slots follow the (possibly changed) qualification.
		if err := qtx.RemoveUnitSlot(req.Context(), int64(gatheringID)); err != nil {
			handlers.RespondWithError(rw, http.StatusInternalServerError, "Failed to update gathering")
			return
		}
		if err := services.NewUnitSlotService(qtx).SyncUnitsSlots(req.Context(), int64(associationID), int64(gatheringID), updateReq.Qualification()); err != nil {
			handlers.RespondWithError(rw, http.StatusInternalServerError, "Failed to sync units slots")
			return
		}
		if err := tx.Commit(); err != nil {
			handlers.RespondWithError(rw, http.StatusInternalServerError, "Failed to update gathering")
			return
		}

		h.statsService.UpdateGatheringStats(int64(gatheringID), int64(associationID))
		updated, err := h.cfg.Db.GetGathering(req.Context(), database.GetGatheringParams{
			ID:            int64(gatheringID),
			AssociationID: int64(associationID),
		})
		if err != nil {
			handlers.RespondWithError(rw, http.StatusInternalServerError, "Failed to get gathering")
			return
		}

		h.cfg.Db.CreateAuditLog(req.Context(), database.CreateAuditLogParams{
			GatheringID: int64(gatheringID),
			EntityType:  "gathering",
			EntityID:    int64(gatheringID),
			Action:      "updated",
			PerformedBy: sql.NullString{String: handlers.GetUserIdFromContext(req), Valid: true},
			IpAddress:   sql.NullString{String: req.RemoteAddr, Valid: true},
			Details:     sql.NullString{String: "{}", Valid: true},
		})

		handlers.RespondWithJSON(rw, http.StatusOK, domain.DBGatheringToResponse(updated))
	}
}
