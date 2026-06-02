package handlers

import (
	"net/http"
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
		r = AddUserIdToContext(r, userLogin)
		r = AddClaimsToContext(r, claims)
		next.ServeHTTP(w, r)
	}
}

// MiddlewareAssociationResource validates the association ID path value.
// With single-association deployments every authenticated user has access;
// the check is kept so route definitions stay unchanged.
func (cfg *ApiConfig) MiddlewareAssociationResource(next http.HandlerFunc) http.HandlerFunc {
	return cfg.MiddlewareAuth(func(w http.ResponseWriter, r *http.Request) {
		associationId, err := strconv.Atoi(r.PathValue(AssociationIdPathValue))
		if err != nil || associationId <= 0 {
			RespondWithError(w, http.StatusBadRequest, "invalid association id")
			return
		}
		next.ServeHTTP(w, r)
	})
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
