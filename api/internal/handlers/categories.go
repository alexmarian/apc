package handlers

import (
	"database/sql"
	"encoding/json"
	"github.com/alexmarian/apc/api/internal/database"
	"log"
	"net/http"
	"strconv"
)

// Add a category
func HandleCreateCategory(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))

		// Parse request
		var category struct {
			Type   string `json:"type"`
			Family string `json:"family"`
			Name   string `json:"name"`
		}

		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&category); err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}

		// Validate required fields
		if category.Type == "" || category.Family == "" || category.Name == "" {
			RespondWithError(rw, http.StatusBadRequest, "Type, family, and name are required")
			return
		}

		// Create category
		newCategory, err := cfg.Db.CreateCategory(req.Context(), database.CreateCategoryParams{
			Type:          category.Type,
			Family:        category.Family,
			Name:          category.Name,
			AssociationID: int64(associationId),
		})

		if err != nil {
			log.Printf("Error creating category: %s", err)
			RespondWithError(rw, http.StatusInternalServerError, "Failed to create category")
			return
		}

		// Return created category
		RespondWithJSON(rw, http.StatusCreated, map[string]interface{}{
			"id":             newCategory.ID,
			"type":           newCategory.Type,
			"family":         newCategory.Family,
			"name":           newCategory.Name,
			"association_id": newCategory.AssociationID,
			"created_at":     newCategory.CreatedAt,
			"updated_at":     newCategory.UpdatedAt,
		})
	}
}

// Deactivate a category (soft delete)
func HandleDeactivateCategory(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		categoryId, _ := strconv.Atoi(req.PathValue("categoryId"))

		// Check if category exists and belongs to association
		category, err := cfg.Db.GetCategory(req.Context(), int64(categoryId))
		if err != nil {
			if err == sql.ErrNoRows {
				RespondWithError(rw, http.StatusNotFound, "Category not found")
			} else {
				log.Printf("Error retrieving category: %s", err)
				RespondWithError(rw, http.StatusInternalServerError, "Failed to retrieve category")
			}
			return
		}

		if category.AssociationID != int64(associationId) {
			RespondWithError(rw, http.StatusForbidden, "Category does not belong to this association")
			return
		}

		// Update category name to mark as inactive
		err = cfg.Db.DeactivateCategory(req.Context(), int64(categoryId))

		if err != nil {
			log.Printf("Error deactivating category: %s", err)
			RespondWithError(rw, http.StatusInternalServerError, "Failed to deactivate category")
			return
		}

		RespondWithJSON(rw, http.StatusOK, map[string]string{
			"message": "Category has been deactivated successfully",
		})
	}
}
