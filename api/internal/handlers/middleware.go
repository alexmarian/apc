package handlers

import (
	"net/http"
	"slices"
	"strconv"

	"github.com/alexmarian/apc/api/internal/auth"
)

func (cfg *ApiConfig) MiddlewareAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			RespondWithError(w, http.StatusUnauthorized, err.Error())
			return
		}
		claims, err := auth.ValidateJWT(token, cfg.Secret)
		if err != nil {
			RespondWithError(w, http.StatusUnauthorized, err.Error())
			return
		}
		if cfg.Db != nil {
			revoked, err := cfg.Db.IsTokenRevoked(r.Context(), claims.ID)
			if err != nil || revoked != 0 {
				RespondWithError(w, http.StatusUnauthorized, "token has been revoked")
				return
			}
		}
		userLogin, _ := claims.GetSubject()
		if cfg.Db != nil {
			// The DB, not the token, is the source of truth for the account:
			// disabling or a role change applies on the next request instead
			// of when the token expires.
			user, err := cfg.Db.GetUserByLogin(r.Context(), userLogin)
			if err != nil || user.DisabledAt.Valid {
				RespondWithError(w, http.StatusUnauthorized, "user is not active")
				return
			}
			claims.IsAdmin = user.IsAdmin
		}
		r = AddUserIdToContext(r, userLogin)
		r = AddClaimsToContext(r, claims)
		next.ServeHTTP(w, r)
	}
}

// MiddlewareAssociationResource guards every association-scoped route:
// application admins may access any association; everyone else only the
// associations they're linked to in users_associations (looked up per
// request, so assignment changes apply without a new login).
func (cfg *ApiConfig) MiddlewareAssociationResource(next http.HandlerFunc) http.HandlerFunc {
	return cfg.MiddlewareAuth(func(w http.ResponseWriter, r *http.Request) {
		associationId, err := strconv.Atoi(r.PathValue(AssociationIdPathValue))
		if err != nil || associationId <= 0 {
			RespondWithError(w, http.StatusBadRequest, "invalid association id")
			return
		}
		if claims := GetClaimsFromContext(r); claims == nil || !claims.IsAdmin {
			if !cfg.userHasAssociation(r, int64(associationId)) {
				RespondWithError(w, http.StatusForbidden, "You don't have access to this association")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (cfg *ApiConfig) userHasAssociation(r *http.Request, associationId int64) bool {
	if cfg.Db == nil {
		return false
	}
	ids, err := cfg.Db.GetUserAssociationsByLogin(r.Context(), GetUserIdFromContext(r))
	if err != nil {
		return false
	}
	return slices.Contains(ids, associationId)
}

func (cfg *ApiConfig) MiddlewareAdminOnly(next http.HandlerFunc) http.HandlerFunc {
	return cfg.MiddlewareAuth(func(w http.ResponseWriter, r *http.Request) {
		claims := GetClaimsFromContext(r)
		if claims == nil || !claims.IsAdmin {
			RespondWithError(w, http.StatusForbidden, "admin privileges required")
			return
		}
		next.ServeHTTP(w, r)
	})
}
