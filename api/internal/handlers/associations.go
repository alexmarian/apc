package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type Association struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Address       string    `json:"address"`
	Administrator string    `json:"administrator"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func HandleGetUserAssociations(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		userAssociationsIds := GetAssotiationIdsToContext(req)
		associationsFromList, err := cfg.Db.GetAssociationsFromList(req.Context(), userAssociationsIds)
		if err != nil {
			var errors = fmt.Sprintf("Error getting associations: %s", err)
			log.Printf(errors)
			RespondWithError(rw, http.StatusInternalServerError, errors)
			return
		}
		associations := make([]Association, len(associationsFromList))
		for i, association := range associationsFromList {
			associations[i] = Association{
				ID:            association.ID,
				Name:          association.Name,
				Address:       association.Address,
				Administrator: association.Administrator,
				CreatedAt:     association.CreatedAt.Time,
				UpdatedAt:     association.UpdatedAt.Time,
			}
		}
		type response struct {
			Associations []Association `json:"associations"`
		}
		responseData := response{
			Associations: associations,
		}
		RespondWithJSON(rw, http.StatusCreated, responseData)
	}
}
