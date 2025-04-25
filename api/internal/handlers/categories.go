package handlers

import (
	"database/sql"
	"encoding/json"
	"github.com/alexmarian/apc/api/internal/database"
	"log"
	"net/http"
	"strconv"
	"time"
)

const CategoryIdPathValue = "categoryId"

type Category struct {
	ID            int64     `json:"id"`
	Type          string    `json:"type"`
	Family        string    `json:"family"`
	Name          string    `json:"name"`
	IsDeleted     bool      `json:"is_deleted"`
	AssociationID int64     `json:"association_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func HandleGetCategory(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {

		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		categoryId, _ := strconv.Atoi(req.PathValue(CategoryIdPathValue))

		dbCategory, err := cfg.Db.GetAssociationCategory(req.Context(), database.GetAssociationCategoryParams{
			ID:            int64(categoryId),
			AssociationID: int64(associationId),
		},
		)

		if err != nil {
			log.Printf("Error creating category: %s", err)
			RespondWithError(rw, http.StatusInternalServerError, "Failed to create category")
			return
		}

		category := Category{
			ID:            dbCategory.ID,
			Type:          dbCategory.Type,
			Family:        dbCategory.Family,
			Name:          dbCategory.Name,
			IsDeleted:     dbCategory.IsDeleted,
			AssociationID: dbCategory.AssociationID,
			CreatedAt:     dbCategory.CreatedAt.Time,
			UpdatedAt:     dbCategory.UpdatedAt.Time,
		}

		RespondWithJSON(rw, http.StatusOK, category)
	}
}

func HandleGetActiveCategories(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))

		dbCategories, err := cfg.Db.GetActiveCategories(req.Context(), int64(associationId))

		if err != nil {
			log.Printf("Error creating category: %s", err)
			RespondWithError(rw, http.StatusInternalServerError, "Failed to create category")
			return
		}
		categories := make([]Category, len(dbCategories))

		for i, category := range dbCategories {
			categories[i] = Category{
				ID:            category.ID,
				Type:          category.Type,
				Family:        category.Family,
				Name:          category.Name,
				IsDeleted:     category.IsDeleted,
				AssociationID: category.AssociationID,
				CreatedAt:     category.CreatedAt.Time,
				UpdatedAt:     category.UpdatedAt.Time,
			}
		}

		RespondWithJSON(rw, http.StatusOK, categories)
	}
}

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
		RespondWithJSON(rw, http.StatusCreated, Category{
			ID:            newCategory.ID,
			Type:          newCategory.Type,
			Family:        newCategory.Family,
			Name:          newCategory.Name,
			AssociationID: newCategory.AssociationID,
			CreatedAt:     newCategory.CreatedAt.Time,
			UpdatedAt:     newCategory.UpdatedAt.Time,
		})
	}
}

// Deactivate a category (soft delete)
func HandleDeactivateCategory(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		categoryId, _ := strconv.Atoi(req.PathValue(CategoryIdPathValue))

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
