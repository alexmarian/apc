package handlers

import (
	"github.com/alexmarian/apc/api/internal/auth"
	"net/http"
	"strconv"
)

func (cfg *ApiConfig) MiddlewareAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			RespondWithError(w, http.StatusUnauthorized, err.Error())
			return
		}
		userLogin, associations, err := auth.ValidateJWT(token, cfg.Secret)
		if err != nil {
			RespondWithError(w, http.StatusUnauthorized, err.Error())
			return
		}
		reqWithAssoction := AddAssotiationIdsToContext(r, associations)
		reqWithUserId := AddUserIdToContext(reqWithAssoction, userLogin)
		next.ServeHTTP(w, reqWithUserId)
	}
}

func (cfg *ApiConfig) MiddlewareAssociationResource(next http.HandlerFunc) http.HandlerFunc {
	return cfg.MiddlewareAuth(func(w http.ResponseWriter, r *http.Request) {
		associationId, _ := strconv.Atoi(r.PathValue(AssociationIdPathValue))
		userAssociationsIds := GetAssotiationIdsToContext(r)
		found := false
		for _, id := range userAssociationsIds {
			if id == int64(associationId) {
				found = true
				break
			}
		}
		if !found {
			RespondWithError(w, http.StatusForbidden, "You don't have access to this association")
			return
		}
		next.ServeHTTP(w, r)
	})
}
func (cfg *ApiConfig) MiddlewareAdminOnly(next http.HandlerFunc) http.HandlerFunc {
	return cfg.MiddlewareAuth(func(w http.ResponseWriter, r *http.Request) {
		userLogin := GetUserIdFromContext(r)

		user, err := cfg.Db.GetUserByLogin(r.Context(), userLogin)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Failed to verify user privileges")
			return
		}

		if !user.IsAdmin {
			RespondWithError(w, http.StatusForbidden, "Admin privileges required")
			return
		}

		next.ServeHTTP(w, r)
	})
}
