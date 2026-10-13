package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/alexmarian/apc/api/internal/database"
	"github.com/alexmarian/apc/api/internal/logging"
	"go.uber.org/zap"
)

const AssociationIdPathValue = "associationId"

type Association struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
	// Administrator is the legacy free-text column, now kept in sync with
	// the comma-joined logins of Administrators for older clients.
	Administrator  string    `json:"administrator"`
	Administrators []UserRef `json:"administrators"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func toAssociation(a database.Association, administrators []UserRef) Association {
	if administrators == nil {
		administrators = []UserRef{}
	}
	return Association{
		ID:             a.ID,
		Name:           a.Name,
		Address:        a.Address,
		Administrator:  a.Administrator,
		Administrators: administrators,
		CreatedAt:      a.CreatedAt.Time,
		UpdatedAt:      a.UpdatedAt.Time,
	}
}

// associationAdministrators groups every association's linked users by
// association id.
func associationAdministrators(ctx context.Context, q *database.Queries) (map[int64][]UserRef, error) {
	rows, err := q.ListAssociationAdministrators(ctx)
	if err != nil {
		return nil, err
	}
	byAssociation := make(map[int64][]UserRef)
	for _, r := range rows {
		byAssociation[r.AssociationID] = append(byAssociation[r.AssociationID], UserRef{ID: r.ID, Login: r.Login})
	}
	return byAssociation, nil
}

// HandleGetUserAssociations lists every association for application admins,
// and only the caller's linked associations for everyone else.
func HandleGetUserAssociations(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		var all []database.Association
		var err error
		if claims := GetClaimsFromContext(req); claims != nil && claims.IsAdmin {
			all, err = cfg.Db.ListAssociations(req.Context())
		} else {
			var ids []int64
			ids, err = cfg.Db.GetUserAssociationsByLogin(req.Context(), GetUserIdFromContext(req))
			if err == nil && len(ids) > 0 {
				all, err = cfg.Db.GetAssociationsFromList(req.Context(), ids)
			}
		}
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error listing associations", zap.String("error", err.Error()))
			RespondWithError(rw, http.StatusInternalServerError, fmt.Sprintf("Error getting associations: %s", err))
			return
		}
		administrators, err := associationAdministrators(req.Context(), cfg.Db)
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error listing association administrators", zap.String("error", err.Error()))
			RespondWithError(rw, http.StatusInternalServerError, "Error getting associations")
			return
		}
		result := make([]Association, len(all))
		for i, a := range all {
			result[i] = toAssociation(a, administrators[a.ID])
		}
		RespondWithJSON(rw, http.StatusOK, result)
	}
}

func HandleGetUserAssociation(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		association, err := cfg.Db.GetAssociations(req.Context(), int64(associationId))
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting association")
			RespondWithError(rw, http.StatusInternalServerError, fmt.Sprintf("Error getting association: %s", err))
			return
		}
		administrators, err := associationAdministrators(req.Context(), cfg.Db)
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error getting association")
			return
		}
		result := toAssociation(association, administrators[association.ID])
		RespondWithJSON(rw, http.StatusOK, &result)
	}
}

// associationFieldMaxLength caps each free-text association field.
const associationFieldMaxLength = 255

// associationInput is a validated create/update body.
type associationInput struct {
	Name           string
	Address        string
	Administrators []UserRef
}

// legacyAdministrator is what goes into the old free-text column.
func (in associationInput) legacyAdministrator() string {
	logins := make([]string, len(in.Administrators))
	for i, u := range in.Administrators {
		logins[i] = u.Login
	}
	return strings.Join(logins, ", ")
}

// parseAssociationRequest decodes and validates a create/update body:
// name and address trimmed, required, at most associationFieldMaxLength
// characters; administratorIds de-duplicated, possibly empty, each an
// existing, enabled user that isn't an application admin. On failure it has already
// written the error response and returns ok=false.
func parseAssociationRequest(cfg *ApiConfig, rw http.ResponseWriter, req *http.Request) (in associationInput, ok bool) {
	var request struct {
		Name             string  `json:"name"`
		Address          string  `json:"address"`
		AdministratorIds []int64 `json:"administratorIds"`
	}
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
		return in, false
	}

	in.Name = strings.TrimSpace(request.Name)
	in.Address = strings.TrimSpace(request.Address)
	for _, field := range []struct{ label, value string }{
		{"Name", in.Name},
		{"Address", in.Address},
	} {
		if field.value == "" {
			RespondWithError(rw, http.StatusBadRequest, fmt.Sprintf("%s is required", field.label))
			return in, false
		}
		if utf8.RuneCountInString(field.value) > associationFieldMaxLength {
			RespondWithError(rw, http.StatusBadRequest,
				fmt.Sprintf("%s must be at most %d characters", field.label, associationFieldMaxLength))
			return in, false
		}
	}

	ids := slices.Clone(request.AdministratorIds)
	slices.Sort(ids)
	ids = slices.Compact(ids)
	if len(ids) == 0 {
		return in, true
	}
	users, err := cfg.Db.GetUsersByIds(req.Context(), ids)
	if err != nil {
		logging.Logger.Log(zap.WarnLevel, "Error loading administrators", zap.String("error", err.Error()))
		RespondWithError(rw, http.StatusInternalServerError, "Failed to load administrators")
		return in, false
	}
	if len(users) != len(ids) {
		RespondWithError(rw, http.StatusBadRequest, "Unknown or invalid administrator")
		return in, false
	}
	for _, u := range users {
		if u.IsAdmin || u.DisabledAt.Valid {
			RespondWithError(rw, http.StatusBadRequest, "Unknown or invalid administrator")
			return in, false
		}
		in.Administrators = append(in.Administrators, UserRef{ID: u.ID, Login: u.Login})
	}
	slices.SortFunc(in.Administrators, func(a, b UserRef) int { return strings.Compare(a.Login, b.Login) })
	return in, true
}

// saveAssociation runs write (the row insert/update) and then replaces the
// association's administrator links, all in one transaction.
func saveAssociation(
	ctx context.Context,
	cfg *ApiConfig,
	in associationInput,
	write func(q *database.Queries) (database.Association, error),
) (database.Association, error) {
	tx, err := cfg.DB.BeginTx(ctx, nil)
	if err != nil {
		return database.Association{}, err
	}
	defer tx.Rollback()
	q := cfg.Db.WithTx(tx)

	association, err := write(q)
	if err != nil {
		return database.Association{}, err
	}
	if err := q.DeleteAssociationUsers(ctx, association.ID); err != nil {
		return database.Association{}, err
	}
	for _, u := range in.Administrators {
		if err := q.AddAssociationUser(ctx, database.AddAssociationUserParams{UserID: u.ID, AssociationID: association.ID}); err != nil {
			return database.Association{}, err
		}
	}
	return association, tx.Commit()
}

// HandleCreateAssociation creates an association and links its
// administrators. Admin-only (see the route in main.go); responds 201.
func HandleCreateAssociation(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		in, ok := parseAssociationRequest(cfg, rw, req)
		if !ok {
			return
		}
		association, err := saveAssociation(req.Context(), cfg, in, func(q *database.Queries) (database.Association, error) {
			return q.CreateAssociation(req.Context(), database.CreateAssociationParams{
				Name:          in.Name,
				Address:       in.Address,
				Administrator: in.legacyAdministrator(),
			})
		})
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error creating association", zap.String("error", err.Error()))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to create association")
			return
		}
		result := toAssociation(association, in.Administrators)
		RespondWithJSON(rw, http.StatusCreated, &result)
	}
}

// HandleUpdateAssociation updates an association's details and replaces its
// administrators. Admin-only (see the route in main.go); responds 200, or
// 404 when the association doesn't exist.
func HandleUpdateAssociation(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, err := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		if err != nil || associationId <= 0 {
			RespondWithError(rw, http.StatusBadRequest, "invalid association id")
			return
		}
		in, ok := parseAssociationRequest(cfg, rw, req)
		if !ok {
			return
		}
		association, err := saveAssociation(req.Context(), cfg, in, func(q *database.Queries) (database.Association, error) {
			return q.UpdateAssociation(req.Context(), database.UpdateAssociationParams{
				Name:          in.Name,
				Address:       in.Address,
				Administrator: in.legacyAdministrator(),
				ID:            int64(associationId),
			})
		})
		if errors.Is(err, sql.ErrNoRows) {
			RespondWithError(rw, http.StatusNotFound, "Association not found")
			return
		}
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error updating association", zap.String("error", err.Error()))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to update association")
			return
		}
		result := toAssociation(association, in.Administrators)
		RespondWithJSON(rw, http.StatusOK, &result)
	}
}
