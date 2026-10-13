package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/alexmarian/apc/api/internal/database"
	"github.com/alexmarian/apc/api/internal/logging"
	"go.uber.org/zap"
)

const UserIdPathValue = "userId"

// User statuses, derived from the users row (never stored).
const (
	UserStatusActive   = "active"
	UserStatusPending  = "pending"
	UserStatusDisabled = "disabled"
)

func userStatus(u database.User) string {
	switch {
	case u.DisabledAt.Valid:
		return UserStatusDisabled
	case !u.TotpConfirmedAt.Valid:
		return UserStatusPending
	default:
		return UserStatusActive
	}
}

// AssociationRef is an association's id and name, as shown next to users
// and invitations.
type AssociationRef struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// AdminUser is a user as the admin Users module sees it.
type AdminUser struct {
	ID           int64            `json:"id"`
	Login        string           `json:"login"`
	IsAdmin      bool             `json:"isAdmin"`
	Status       string           `json:"status"`
	CreatedAt    time.Time        `json:"createdAt"`
	DisabledBy   string           `json:"disabledBy,omitempty"`
	Associations []AssociationRef `json:"associations"`
}

func loadAdminUsers(req *http.Request, q *database.Queries) ([]AdminUser, error) {
	users, err := q.ListUsers(req.Context())
	if err != nil {
		return nil, err
	}
	links, err := q.ListUserAssociationLinks(req.Context())
	if err != nil {
		return nil, err
	}
	byUser := make(map[int64][]AssociationRef)
	for _, l := range links {
		byUser[l.UserID] = append(byUser[l.UserID], AssociationRef{ID: l.ID, Name: l.Name})
	}
	result := make([]AdminUser, len(users))
	for i, u := range users {
		associations := byUser[u.ID]
		if associations == nil {
			associations = []AssociationRef{}
		}
		result[i] = AdminUser{
			ID:           u.ID,
			Login:        u.Login,
			IsAdmin:      u.IsAdmin,
			Status:       userStatus(u),
			CreatedAt:    u.CreatedAt,
			DisabledBy:   u.DisabledBy.String,
			Associations: associations,
		}
	}
	return result, nil
}

// respondWithAdminUser answers with the current state of one user.
func respondWithAdminUser(cfg *ApiConfig, rw http.ResponseWriter, req *http.Request, id int64) {
	users, err := loadAdminUsers(req, cfg.Db)
	if err == nil {
		for _, u := range users {
			if u.ID == id {
				RespondWithJSON(rw, http.StatusOK, u)
				return
			}
		}
	}
	RespondWithError(rw, http.StatusInternalServerError, "Error loading user")
}

// HandleListUsers (GET /admin/users) lists every user with role, status and
// associations. Admin-only.
func HandleListUsers(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		users, err := loadAdminUsers(req, cfg.Db)
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error listing users", zap.String("error", err.Error()))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to list users")
			return
		}
		RespondWithJSON(rw, http.StatusOK, users)
	}
}

// targetUser resolves the {userId} path value. On failure it has already
// written the response.
func targetUser(cfg *ApiConfig, rw http.ResponseWriter, req *http.Request) (database.User, bool) {
	id, err := strconv.ParseInt(req.PathValue(UserIdPathValue), 10, 64)
	if err != nil || id <= 0 {
		RespondWithError(rw, http.StatusBadRequest, "invalid user id")
		return database.User{}, false
	}
	user, err := cfg.Db.GetUserByID(req.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		RespondWithError(rw, http.StatusNotFound, "User not found")
		return database.User{}, false
	}
	if err != nil {
		RespondWithError(rw, http.StatusInternalServerError, "Error loading user")
		return database.User{}, false
	}
	return user, true
}

// guardAdminChange refuses changes an admin may not make to user: anything
// on their own account, and anything that would leave no active app admin
// (removing is true when the change takes user out of the active admins).
func guardAdminChange(cfg *ApiConfig, rw http.ResponseWriter, req *http.Request, user database.User, removing bool) bool {
	if user.Login == GetUserIdFromContext(req) {
		RespondWithError(rw, http.StatusConflict, "You can't change your own role or status")
		return false
	}
	if removing && user.IsAdmin && !user.DisabledAt.Valid {
		count, err := cfg.Db.CountActiveAdmins(req.Context())
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error checking admins")
			return false
		}
		if count <= 1 {
			RespondWithError(rw, http.StatusConflict, "At least one active application admin is required")
			return false
		}
	}
	return true
}

// HandleSetUserRole (PUT /admin/users/{userId}/role, {isAdmin}) promotes or
// demotes an application admin. Promoting removes the user's association
// links: admins reach every association anyway, and an admin may not be an
// association administrator.
func HandleSetUserRole(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		var request struct {
			IsAdmin *bool `json:"isAdmin"`
		}
		if err := json.NewDecoder(req.Body).Decode(&request); err != nil || request.IsAdmin == nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}
		user, ok := targetUser(cfg, rw, req)
		if !ok || !guardAdminChange(cfg, rw, req, user, !*request.IsAdmin) {
			return
		}

		ctx := req.Context()
		tx, err := cfg.DB.BeginTx(ctx, nil)
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error updating role")
			return
		}
		defer tx.Rollback()
		q := cfg.Db.WithTx(tx)
		err = q.SetUserAdmin(ctx, database.SetUserAdminParams{IsAdmin: *request.IsAdmin, ID: user.ID})
		if err == nil && *request.IsAdmin {
			err = q.DeleteUserAssociations(ctx, user.ID)
		}
		if err == nil {
			err = tx.Commit()
		}
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error updating role", zap.String("error", err.Error()))
			RespondWithError(rw, http.StatusInternalServerError, "Error updating role")
			return
		}
		respondWithAdminUser(cfg, rw, req, user.ID)
	}
}

// HandleSetUserStatus (PUT /admin/users/{userId}/status, {disabled})
// disables or re-enables a user. Disabling also revokes their refresh
// tokens; their access tokens stop working on the next request
// (MiddlewareAuth checks the DB).
func HandleSetUserStatus(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		var request struct {
			Disabled *bool `json:"disabled"`
		}
		if err := json.NewDecoder(req.Body).Decode(&request); err != nil || request.Disabled == nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}
		user, ok := targetUser(cfg, rw, req)
		if !ok || !guardAdminChange(cfg, rw, req, user, *request.Disabled) {
			return
		}

		ctx := req.Context()
		params := database.SetUserDisabledParams{ID: user.ID}
		if *request.Disabled {
			params.DisabledAt = sql.NullTime{Time: time.Now().UTC(), Valid: true}
			params.DisabledBy = sql.NullString{String: GetUserIdFromContext(req), Valid: true}
		}
		err := cfg.Db.SetUserDisabled(ctx, params)
		if err == nil && *request.Disabled {
			err = cfg.Db.RevokeUserRefreshTokens(ctx, user.Login)
		}
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error updating user status", zap.String("error", err.Error()))
			RespondWithError(rw, http.StatusInternalServerError, "Error updating status")
			return
		}
		respondWithAdminUser(cfg, rw, req, user.ID)
	}
}

// HandleSetUserAssociations (PUT /admin/users/{userId}/associations,
// {associationIds}) replaces the associations a user administers — the
// same links the association form edits. Not for app admins.
func HandleSetUserAssociations(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		var request struct {
			AssociationIds []int64 `json:"associationIds"`
		}
		if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}
		user, ok := targetUser(cfg, rw, req)
		if !ok {
			return
		}
		if user.IsAdmin {
			RespondWithError(rw, http.StatusBadRequest, "Application admins already have access to every association")
			return
		}
		ids, ok := validAssociationIds(cfg, rw, req, request.AssociationIds)
		if !ok {
			return
		}

		ctx := req.Context()
		tx, err := cfg.DB.BeginTx(ctx, nil)
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error updating associations")
			return
		}
		defer tx.Rollback()
		q := cfg.Db.WithTx(tx)
		err = q.DeleteUserAssociations(ctx, user.ID)
		for _, id := range ids {
			if err != nil {
				break
			}
			err = q.AddAssociationUser(ctx, database.AddAssociationUserParams{UserID: user.ID, AssociationID: id})
		}
		if err == nil {
			err = tx.Commit()
		}
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error updating user associations", zap.String("error", err.Error()))
			RespondWithError(rw, http.StatusInternalServerError, "Error updating associations")
			return
		}
		respondWithAdminUser(cfg, rw, req, user.ID)
	}
}

// validAssociationIds de-duplicates ids and checks every one exists. On
// failure it has already written the response.
func validAssociationIds(cfg *ApiConfig, rw http.ResponseWriter, req *http.Request, ids []int64) ([]int64, bool) {
	ids = slices.Clone(ids)
	slices.Sort(ids)
	ids = slices.Compact(ids)
	if len(ids) == 0 {
		return ids, true
	}
	found, err := cfg.Db.GetAssociationsFromList(req.Context(), ids)
	if err != nil {
		RespondWithError(rw, http.StatusInternalServerError, "Error loading associations")
		return nil, false
	}
	if len(found) != len(ids) {
		RespondWithError(rw, http.StatusBadRequest, "Unknown association")
		return nil, false
	}
	return ids, true
}
