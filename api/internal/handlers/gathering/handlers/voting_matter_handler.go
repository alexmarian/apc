package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/alexmarian/apc/api/internal/database"
	"github.com/alexmarian/apc/api/internal/handlers"
	"github.com/alexmarian/apc/api/internal/handlers/gathering/domain"
	"github.com/alexmarian/apc/api/internal/logging"
	"go.uber.org/zap"
)

// VotingMatterHandler handles voting matter operations
type VotingMatterHandler struct {
	cfg              *handlers.ApiConfig
	gatheringHandler *GatheringHandler
}

// NewVotingMatterHandler creates a new VotingMatterHandler
func NewVotingMatterHandler(cfg *handlers.ApiConfig, gatheringHandler *GatheringHandler) *VotingMatterHandler {
	return &VotingMatterHandler{
		cfg:              cfg,
		gatheringHandler: gatheringHandler,
	}
}

// HandleGetVotingMatters returns all voting matters for a gathering
func (h *VotingMatterHandler) HandleGetVotingMatters() func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		gatheringID, _ := strconv.Atoi(req.PathValue(domain.GatheringIDPathValue))

		matters, err := h.cfg.Db.GetVotingMatters(req.Context(), int64(gatheringID))
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting voting matters", zap.Error(err))
			handlers.RespondWithError(rw, http.StatusInternalServerError, "Failed to get voting matters")
			return
		}

		response := make([]domain.VotingMatter, len(matters))
		for i, m := range matters {
			response[i] = domain.DBVotingMatterToResponse(m)
		}

		handlers.RespondWithJSON(rw, http.StatusOK, response)
	}
}

// HandleCreateVotingMatter creates a new voting matter
func (h *VotingMatterHandler) HandleCreateVotingMatter() func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		gatheringID, _ := strconv.Atoi(req.PathValue(domain.GatheringIDPathValue))
		associationID, _ := strconv.Atoi(req.PathValue(handlers.AssociationIdPathValue))

		draft := h.gatheringHandler.ValidateGatheringStateWithFetch(req, gatheringID, associationID, "draft")
		if !draft {
			handlers.RespondWithError(rw, http.StatusBadRequest, "Cannot create voting matter in non-draft gathering")
			return
		}

		var createReq domain.VotingMatter
		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&createReq); err != nil {
			handlers.RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}

		// Convert voting config to JSON
		configJSON, err := json.Marshal(createReq.VotingConfig)
		if err != nil {
			handlers.RespondWithError(rw, http.StatusBadRequest, "Invalid voting configuration")
			return
		}

		matter, err := h.cfg.Db.CreateVotingMatter(req.Context(), database.CreateVotingMatterParams{
			GatheringID:  int64(gatheringID),
			OrderIndex:   int64(createReq.OrderIndex),
			Title:        createReq.Title,
			Description:  sql.NullString{String: createReq.Description, Valid: createReq.Description != ""},
			MatterType:   createReq.MatterType,
			VotingConfig: string(configJSON),
		})

		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error creating voting matter", zap.Error(err))
			handlers.RespondWithError(rw, http.StatusInternalServerError, "Failed to create voting matter")
			return
		}

		handlers.RespondWithJSON(rw, http.StatusCreated, domain.DBVotingMatterToResponse(matter))
	}
}

// HandleUpdateVotingMatter updates an existing voting matter
func (h *VotingMatterHandler) HandleUpdateVotingMatter() func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		gatheringID, _ := strconv.Atoi(req.PathValue(domain.GatheringIDPathValue))
		matterID, _ := strconv.Atoi(req.PathValue(domain.VotingMatterIDPathValue))
		associationID, _ := strconv.Atoi(req.PathValue(handlers.AssociationIdPathValue))

		draft := h.gatheringHandler.ValidateGatheringStateWithFetch(req, gatheringID, associationID, "draft")
		if !draft {
			handlers.RespondWithError(rw, http.StatusBadRequest, "Cannot update voting matter in non-draft gathering")
			return
		}

		var createReq domain.VotingMatter
		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&createReq); err != nil {
			handlers.RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}

		// Convert voting config to JSON
		configJSON, err := json.Marshal(createReq.VotingConfig)
		if err != nil {
			handlers.RespondWithError(rw, http.StatusBadRequest, "Invalid voting configuration")
			return
		}

		matter, err := h.cfg.Db.UpdateVotingMatter(req.Context(), database.UpdateVotingMatterParams{
			GatheringID:  int64(gatheringID),
			ID:           int64(matterID),
			Title:        createReq.Title,
			Description:  sql.NullString{String: createReq.Description, Valid: createReq.Description != ""},
			MatterType:   createReq.MatterType,
			OrderIndex:   int64(createReq.OrderIndex),
			VotingConfig: string(configJSON),
		})

		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error updating voting matter", zap.Error(err))
			handlers.RespondWithError(rw, http.StatusInternalServerError, "Failed to update voting matter")
			return
		}

		handlers.RespondWithJSON(rw, http.StatusCreated, domain.DBVotingMatterToResponse(matter))
	}
}

// HandleDeleteVotingMatter deletes a voting matter
func (h *VotingMatterHandler) HandleDeleteVotingMatter() func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		gatheringID, _ := strconv.Atoi(req.PathValue(domain.GatheringIDPathValue))
		matterID, _ := strconv.Atoi(req.PathValue(domain.VotingMatterIDPathValue))
		associationID, _ := strconv.Atoi(req.PathValue(handlers.AssociationIdPathValue))

		draft := h.gatheringHandler.ValidateGatheringStateWithFetch(req, gatheringID, associationID, "draft")
		if !draft {
			handlers.RespondWithError(rw, http.StatusBadRequest, "Cannot delete voting matter in non-draft gathering")
			return
		}

		err := h.cfg.Db.DeleteVotingMatter(req.Context(), database.DeleteVotingMatterParams{
			GatheringID: int64(gatheringID),
			ID:          int64(matterID),
		})

		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error deleting voting matter", zap.Error(err))
			handlers.RespondWithError(rw, http.StatusInternalServerError, "Failed to delete voting matter")
			return
		}

		handlers.RespondWithJSON(rw, http.StatusOK, map[string]string{"result": "success"})
	}
}
