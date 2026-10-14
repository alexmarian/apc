package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"sort"
	"strconv"

	"github.com/alexmarian/apc/api/internal/database"
	"github.com/alexmarian/apc/api/internal/handlers"
	"github.com/alexmarian/apc/api/internal/handlers/gathering/domain"
)

// A signed-in owner's gatherings: the same view and voting as a member link,
// for every association where the account holds an owner record.

type myGathering struct {
	handlers.MemberGatheringResponse
	AssociationID   int64  `json:"association_id"`
	AssociationName string `json:"association_name"`
	Location        string `json:"location"`
}

// myGatheringOwner finds the gathering in the path and the user's owner
// record in its association (404 when the gathering isn't theirs).
func myGatheringOwner(cfg *handlers.ApiConfig, rw http.ResponseWriter, req *http.Request) (database.Gathering, database.ListUserOwnersRow, bool) {
	owners, ok := handlers.MyOwners(cfg, req)
	if !ok {
		handlers.RespondWithError(rw, http.StatusUnauthorized, "Unauthorized")
		return database.Gathering{}, database.ListUserOwnersRow{}, false
	}
	id, err := strconv.ParseInt(req.PathValue(domain.GatheringIDPathValue), 10, 64)
	if err != nil {
		handlers.RespondWithError(rw, http.StatusBadRequest, "Invalid gathering id")
		return database.Gathering{}, database.ListUserOwnersRow{}, false
	}
	gathering, err := cfg.Db.GetGatheringByID(req.Context(), id)
	if err == nil && gathering.Status != "draft" {
		for _, o := range owners {
			if o.AssociationID == gathering.AssociationID {
				return gathering, o, true
			}
		}
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		handlers.RespondWithError(rw, http.StatusInternalServerError, "Error reading gathering")
		return database.Gathering{}, database.ListUserOwnersRow{}, false
	}
	handlers.RespondWithError(rw, http.StatusNotFound, "Gathering not found")
	return database.Gathering{}, database.ListUserOwnersRow{}, false
}

// HandleMyGatherings: GET /me/gatherings — published, active, closed and
// tallied gatherings where one of the user's owner records has qualifying
// units, newest first.
func HandleMyGatherings(cfg *handlers.ApiConfig) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		owners, ok := handlers.MyOwners(cfg, req)
		if !ok {
			handlers.RespondWithError(rw, http.StatusUnauthorized, "Unauthorized")
			return
		}
		list := []myGathering{}
		for _, o := range owners {
			gatherings, err := cfg.Db.GetGatherings(req.Context(), o.AssociationID)
			if err != nil {
				handlers.RespondWithError(rw, http.StatusInternalServerError, "Error reading gatherings")
				return
			}
			for _, g := range gatherings {
				if g.Status == "draft" {
					continue
				}
				view, err := handlers.MemberContext(req.Context(), cfg, g, o.ID)
				if err != nil {
					handlers.RespondWithError(rw, http.StatusInternalServerError, "Error reading gatherings")
					return
				}
				if len(view.Units) == 0 {
					continue // the owner's units don't qualify
				}
				list = append(list, myGathering{view, o.AssociationID, o.AssociationName, g.Location})
			}
		}
		sort.SliceStable(list, func(i, j int) bool { return list[i].Gathering.GatheringDate.After(list[j].Gathering.GatheringDate) })
		handlers.RespondWithJSON(rw, http.StatusOK, list)
	}
}

// HandleMyGathering: GET /me/gatherings/{gatheringId}.
func HandleMyGathering(cfg *handlers.ApiConfig) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		gathering, owner, ok := myGatheringOwner(cfg, rw, req)
		if !ok {
			return
		}
		view, err := handlers.MemberContext(req.Context(), cfg, gathering, owner.ID)
		if err != nil {
			handlers.RespondWithError(rw, http.StatusInternalServerError, "Error reading gathering")
			return
		}
		handlers.RespondWithJSON(rw, http.StatusOK, myGathering{view, owner.AssociationID, owner.AssociationName, gathering.Location})
	}
}

// HandleMyBallot: POST /me/gatherings/{gatheringId}/ballot — the owner votes
// with all their available units (see submitOwnerBallot).
func HandleMyBallot(cfg *handlers.ApiConfig) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		gathering, owner, ok := myGatheringOwner(cfg, rw, req)
		if !ok {
			return
		}
		respondOwnerBallot(rw, req, cfg, gathering, owner.ID, "user:"+handlers.GetUserIdFromContext(req), "owner_account")
	}
}
