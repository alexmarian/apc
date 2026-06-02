package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/alexmarian/apc/api/internal/database"
	"github.com/alexmarian/apc/api/internal/handlers"
	"github.com/alexmarian/apc/api/internal/handlers/gathering/domain"
	"github.com/alexmarian/apc/api/internal/logging"
	"go.uber.org/zap"
)

const defaultInvitationTTL = 365 * 24 * time.Hour

type InvitationHandler struct {
	cfg *handlers.ApiConfig
}

func NewInvitationHandler(cfg *handlers.ApiConfig) *InvitationHandler {
	return &InvitationHandler{cfg: cfg}
}

type invitationResponse struct {
	ID          int64     `json:"id"`
	GatheringID int64     `json:"gathering_id"`
	OwnerID     int64     `json:"owner_id"`
	ExpiresAt   time.Time `json:"expires_at"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type createInvitationResponse struct {
	invitationResponse
	Token string `json:"token"`
}

func invitationStatus(inv database.MemberInvitation) string {
	if inv.RevokedAt != nil {
		return "revoked"
	}
	if inv.ExpiresAt.Before(time.Now()) {
		return "expired"
	}
	return "active"
}

func toInvitationResponse(inv database.MemberInvitation) invitationResponse {
	return invitationResponse{
		ID:          inv.ID,
		GatheringID: inv.GatheringID,
		OwnerID:     inv.OwnerID,
		ExpiresAt:   inv.ExpiresAt,
		Status:      invitationStatus(inv),
		CreatedAt:   inv.CreatedAt,
	}
}

// HandleCreateInvitation generates an opaque token for a member to access a gathering.
// The plaintext token is returned once and never stored.
func (h *InvitationHandler) HandleCreateInvitation() func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationID, _ := strconv.Atoi(req.PathValue(handlers.AssociationIdPathValue))
		gatheringID, _ := strconv.Atoi(req.PathValue(domain.GatheringIDPathValue))

		var body struct {
			OwnerID   int64      `json:"owner_id"`
			ExpiresAt *time.Time `json:"expires_at,omitempty"`
		}
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil || body.OwnerID == 0 {
			handlers.RespondWithError(rw, http.StatusBadRequest, "owner_id is required")
			return
		}

		// Verify gathering belongs to association
		gathering, err := h.cfg.Db.GetGathering(req.Context(), database.GetGatheringParams{
			ID:            int64(gatheringID),
			AssociationID: int64(associationID),
		})
		if err != nil {
			handlers.RespondWithError(rw, http.StatusNotFound, "gathering not found")
			return
		}
		_ = gathering

		// Check for existing non-revoked invitation (active or expired)
		existing, err := h.cfg.Db.GetActiveMemberInvitationByOwnerAndGathering(req.Context(),
			database.GetActiveMemberInvitationByOwnerAndGatheringParams{
				GatheringID: int64(gatheringID),
				OwnerID:     body.OwnerID,
			})
		if err == nil && existing.ID != 0 {
			handlers.RespondWithError(rw, http.StatusConflict,
				"an invitation already exists for this owner; revoke it before creating a new one")
			return
		}

		expiresAt := time.Now().Add(defaultInvitationTTL)
		if body.ExpiresAt != nil {
			expiresAt = *body.ExpiresAt
		}

		// Generate 32 random bytes → base64url token
		raw := make([]byte, 32)
		if _, err := rand.Read(raw); err != nil {
			logging.Logger.Log(zap.ErrorLevel, "failed to generate token", zap.Error(err))
			handlers.RespondWithError(rw, http.StatusInternalServerError, "failed to generate token")
			return
		}
		token := base64.RawURLEncoding.EncodeToString(raw)
		hash := sha256.Sum256([]byte(token))
		tokenHash := hex.EncodeToString(hash[:])

		inv, err := h.cfg.Db.CreateMemberInvitation(req.Context(), database.CreateMemberInvitationParams{
			GatheringID: int64(gatheringID),
			OwnerID:     body.OwnerID,
			TokenHash:   tokenHash,
			ExpiresAt:   expiresAt,
		})
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "failed to create invitation", zap.Error(err))
			handlers.RespondWithError(rw, http.StatusInternalServerError, "failed to create invitation")
			return
		}

		handlers.RespondWithJSON(rw, http.StatusCreated, createInvitationResponse{
			invitationResponse: toInvitationResponse(inv),
			Token:              token,
		})
	}
}

// HandleListInvitations returns all invitations for a gathering (never returns token plaintext).
func (h *InvitationHandler) HandleListInvitations() func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationID, _ := strconv.Atoi(req.PathValue(handlers.AssociationIdPathValue))
		gatheringID, _ := strconv.Atoi(req.PathValue(domain.GatheringIDPathValue))

		_, err := h.cfg.Db.GetGathering(req.Context(), database.GetGatheringParams{
			ID:            int64(gatheringID),
			AssociationID: int64(associationID),
		})
		if err != nil {
			handlers.RespondWithError(rw, http.StatusNotFound, "gathering not found")
			return
		}

		invitations, err := h.cfg.Db.ListMemberInvitationsByGathering(req.Context(), int64(gatheringID))
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "failed to list invitations", zap.Error(err))
			handlers.RespondWithError(rw, http.StatusInternalServerError, "failed to list invitations")
			return
		}

		response := make([]invitationResponse, len(invitations))
		for i, inv := range invitations {
			response[i] = toInvitationResponse(inv)
		}
		handlers.RespondWithJSON(rw, http.StatusOK, response)
	}
}

// HandleRevokeInvitation immediately invalidates a member invitation.
func (h *InvitationHandler) HandleRevokeInvitation() func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationID, _ := strconv.Atoi(req.PathValue(handlers.AssociationIdPathValue))
		gatheringID, _ := strconv.Atoi(req.PathValue(domain.GatheringIDPathValue))
		invitationID, _ := strconv.Atoi(req.PathValue(domain.InvitationIDPathValue))

		_, err := h.cfg.Db.GetGathering(req.Context(), database.GetGatheringParams{
			ID:            int64(gatheringID),
			AssociationID: int64(associationID),
		})
		if err != nil {
			handlers.RespondWithError(rw, http.StatusNotFound, "gathering not found")
			return
		}

		if err := h.cfg.Db.RevokeMemberInvitation(req.Context(), int64(invitationID)); err != nil {
			logging.Logger.Log(zap.WarnLevel, "failed to revoke invitation", zap.Error(err),
				zap.Int("invitation_id", invitationID))
			if err == sql.ErrNoRows {
				handlers.RespondWithError(rw, http.StatusNotFound, "invitation not found or already revoked")
				return
			}
			handlers.RespondWithError(rw, http.StatusInternalServerError, "failed to revoke invitation")
			return
		}

		handlers.RespondWithJSON(rw, http.StatusOK, map[string]string{"status": "revoked"})
	}
}
