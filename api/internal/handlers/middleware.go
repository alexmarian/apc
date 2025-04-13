package handlers

import (
	"github.com/alexmarian/apc/api/internal/auth"
	"net/http"
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
		next.ServeHTTP(w, AddUserIdToContext(AddAssotiationIdsToContext(r, associations), userLogin))
	}
}
