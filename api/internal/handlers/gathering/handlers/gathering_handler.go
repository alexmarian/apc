package handlers

import (
	"database/sql"
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

// GatheringHandler handles gathering CRUD operations
type GatheringHandler struct {
	cfg             *handlers.ApiConfig
	statsService    *services.StatsService
	unitSlotService *services.UnitSlotService
	quorumService   *services.QuorumService
	tallyService    *services.TallyService
}

// NewGatheringHandler creates a new GatheringHandler
func NewGatheringHandler(cfg *handlers.ApiConfig) *GatheringHandler {
	return &GatheringHandler{
		cfg:             cfg,
		statsService:    services.NewStatsService(cfg.Db),
		unitSlotService: services.NewUnitSlotService(cfg.Db),
		quorumService:   services.NewQuorumService(cfg.Db),
		tallyService:    services.NewTallyService(cfg.Db),
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

// HandleCreateGathering creates a new gathering
func (h *GatheringHandler) HandleCreateGathering() func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationID, _ := strconv.Atoi(req.PathValue(handlers.AssociationIdPathValue))

		var createReq domain.CreateGatheringRequest
		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&createReq); err != nil {
			handlers.RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}

		// Validate required fields
		if createReq.Title == "" || createReq.GatheringType == "" {
			handlers.RespondWithError(rw, http.StatusBadRequest, "Title and gathering type are required")
			return
		}

		if createReq.Location == "" {
			handlers.RespondWithError(rw, http.StatusBadRequest, "Location is required")
			return
		}

		// Convert arrays to JSON strings for storage
		unitTypesJSON, _ := json.Marshal(createReq.QualificationUnitTypes)
		floorsJSON, _ := json.Marshal(createReq.QualificationFloors)
		entrancesJSON, _ := json.Marshal(createReq.QualificationEntrances)

		gathering, err := h.cfg.Db.CreateGathering(req.Context(), database.CreateGatheringParams{
			AssociationID:           int64(associationID),
			Title:                   createReq.Title,
			Description:             createReq.Description,
			Intent:                  createReq.Intent,
			Location:                createReq.Location,
			GatheringDate:           createReq.GatheringDate,
			GatheringType:           createReq.GatheringType,
			Status:                  "draft",
			QualificationUnitTypes:  sql.NullString{String: string(unitTypesJSON), Valid: len(unitTypesJSON) > 2},
			QualificationFloors:     sql.NullString{String: string(floorsJSON), Valid: len(floorsJSON) > 2},
			QualificationEntrances:  sql.NullString{String: string(entrancesJSON), Valid: len(entrancesJSON) > 2},
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

		err = h.unitSlotService.SyncUnitsSlots(req.Context(), int64(associationID), gathering.ID, createReq.QualificationUnitTypes, createReq.QualificationFloors, createReq.QualificationEntrances)
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

		// If closing the gathering, calculate final results
		if statusReq.Status == "closed" {
			go h.statsService.CalculateFinalResults(int64(gatheringID), int64(associationID), h.tallyService)
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
