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
	ID             int64              `json:"id"`
	Type           string             `json:"type"`
	Family         string             `json:"family"`
	Name           string             `json:"name"`
	IsDeleted      bool               `json:"is_deleted"`
	AssociationID  int64              `json:"association_id"`
	OriginalLabels map[string]string  `json:"original_labels,omitempty"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
}

func parseOriginalLabels(ns sql.NullString) map[string]string {
	if !ns.Valid || ns.String == "" {
		return nil
	}
	var m map[string]string
	if err := json.Unmarshal([]byte(ns.String), &m); err != nil {
		return nil
	}
	return m
}

func serializeOriginalLabels(m map[string]string) sql.NullString {
	if len(m) == 0 {
		return sql.NullString{}
	}
	b, err := json.Marshal(m)
	if err != nil {
		return sql.NullString{}
	}
	return sql.NullString{String: string(b), Valid: true}
}

func dbCategoryToResponse(c database.Category) Category {
	return Category{
		ID:             c.ID,
		Type:           c.Type,
		Family:         c.Family,
		Name:           c.Name,
		IsDeleted:      c.IsDeleted,
		AssociationID:  c.AssociationID,
		OriginalLabels: parseOriginalLabels(c.OriginalLabels),
		CreatedAt:      c.CreatedAt.Time,
		UpdatedAt:      c.UpdatedAt.Time,
	}
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

		RespondWithJSON(rw, http.StatusOK, dbCategoryToResponse(dbCategory))
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
			categories[i] = dbCategoryToResponse(category)
		}

		RespondWithJSON(rw, http.StatusOK, categories)
	}
}

// Add a category
func HandleCreateCategory(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))

		var category struct {
			Type           string            `json:"type"`
			Family         string            `json:"family"`
			Name           string            `json:"name"`
			OriginalLabels map[string]string `json:"original_labels"`
		}

		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&category); err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}

		if category.Type == "" || category.Family == "" || category.Name == "" {
			RespondWithError(rw, http.StatusBadRequest, "Type, family, and name are required")
			return
		}

		newCategory, err := cfg.Db.CreateCategory(req.Context(), database.CreateCategoryParams{
			Type:           category.Type,
			Family:         category.Family,
			Name:           category.Name,
			AssociationID:  int64(associationId),
			OriginalLabels: serializeOriginalLabels(category.OriginalLabels),
		})

		if err != nil {
			log.Printf("Error creating category: %s", err)
			RespondWithError(rw, http.StatusInternalServerError, "Failed to create category")
			return
		}

		RespondWithJSON(rw, http.StatusCreated, dbCategoryToResponse(newCategory))
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

// Get all categories (with optional inactive filter)
func HandleGetAllCategories(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		includeInactive := req.URL.Query().Get("include_inactive") == "true"

		dbCategories, err := cfg.Db.GetAllCategories(req.Context(), database.GetAllCategoriesParams{
			AssociationID: int64(associationId),
			Column2:       includeInactive,
		})

		if err != nil {
			log.Printf("Error getting categories: %s", err)
			RespondWithError(rw, http.StatusInternalServerError, "Failed to get categories")
			return
		}

		categories := make([]Category, len(dbCategories))
		for i, category := range dbCategories {
			categories[i] = dbCategoryToResponse(category)
		}

		RespondWithJSON(rw, http.StatusOK, categories)
	}
}

// Update a category
func HandleUpdateCategory(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		categoryId, _ := strconv.Atoi(req.PathValue(CategoryIdPathValue))

		var category struct {
			Type           string            `json:"type"`
			Family         string            `json:"family"`
			Name           string            `json:"name"`
			OriginalLabels map[string]string `json:"original_labels"`
		}

		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&category); err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}

		if category.Type == "" || category.Family == "" || category.Name == "" {
			RespondWithError(rw, http.StatusBadRequest, "Type, family, and name are required")
			return
		}

		count, err := cfg.Db.CheckCategoryUniqueness(req.Context(), database.CheckCategoryUniquenessParams{
			AssociationID: int64(associationId),
			Type:          category.Type,
			Family:        category.Family,
			Name:          category.Name,
			Column5:       int64(categoryId),
			ID:            int64(categoryId),
		})
		if err != nil {
			log.Printf("Error checking category uniqueness: %s", err)
			RespondWithError(rw, http.StatusInternalServerError, "Failed to validate category")
			return
		}
		if count > 0 {
			RespondWithError(rw, http.StatusConflict, "A category with these values already exists")
			return
		}

		updatedCategory, err := cfg.Db.UpdateCategory(req.Context(), database.UpdateCategoryParams{
			Type:           category.Type,
			Family:         category.Family,
			Name:           category.Name,
			OriginalLabels: serializeOriginalLabels(category.OriginalLabels),
			ID:             int64(categoryId),
			AssociationID:  int64(associationId),
		})

		if err != nil {
			log.Printf("Error updating category: %s", err)
			RespondWithError(rw, http.StatusInternalServerError, "Failed to update category")
			return
		}

		RespondWithJSON(rw, http.StatusOK, dbCategoryToResponse(updatedCategory))
	}
}

// Reactivate a category
func HandleReactivateCategory(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
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

		// Check for uniqueness before reactivating
		count, err := cfg.Db.CheckCategoryUniqueness(req.Context(), database.CheckCategoryUniquenessParams{
			AssociationID: int64(associationId),
			Type:          category.Type,
			Family:        category.Family,
			Name:          category.Name,
			Column5:       0,
			ID:            0,
		})
		if err != nil {
			log.Printf("Error checking category uniqueness: %s", err)
			RespondWithError(rw, http.StatusInternalServerError, "Failed to validate category")
			return
		}
		if count > 0 {
			RespondWithError(rw, http.StatusConflict, "Cannot reactivate: a category with these values already exists")
			return
		}

		// Reactivate category
		err = cfg.Db.ReactivateCategory(req.Context(), database.ReactivateCategoryParams{
			ID:            int64(categoryId),
			AssociationID: int64(associationId),
		})

		if err != nil {
			log.Printf("Error reactivating category: %s", err)
			RespondWithError(rw, http.StatusInternalServerError, "Failed to reactivate category")
			return
		}

		RespondWithJSON(rw, http.StatusOK, map[string]string{
			"message": "Category has been reactivated successfully",
		})
	}
}

// Get category usage statistics
func HandleGetCategoryUsage(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
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

		// Get usage count
		usageCount, err := cfg.Db.GetCategoryUsageCount(req.Context(), int64(categoryId))
		if err != nil {
			log.Printf("Error getting category usage count: %s", err)
			RespondWithError(rw, http.StatusInternalServerError, "Failed to get usage statistics")
			return
		}

		// Get usage details
		usageDetails, err := cfg.Db.GetCategoryUsageDetails(req.Context(), int64(categoryId))
		if err != nil {
			log.Printf("Error getting category usage details: %s", err)
			RespondWithError(rw, http.StatusInternalServerError, "Failed to get usage details")
			return
		}

		type ExpenseDetail struct {
			ID          int64     `json:"id"`
			Description string    `json:"description"`
			Amount      float64   `json:"amount"`
			Date        time.Time `json:"date"`
			CreatedAt   time.Time `json:"created_at"`
		}

		expenses := make([]ExpenseDetail, len(usageDetails))
		for i, expense := range usageDetails {
			createdAt := time.Time{}
			if expense.CreatedAt.Valid {
				createdAt = expense.CreatedAt.Time
			}
			expenses[i] = ExpenseDetail{
				ID:          expense.ID,
				Description: expense.Description,
				Amount:      expense.Amount,
				Date:        expense.Date,
				CreatedAt:   createdAt,
			}
		}

		response := struct {
			CategoryID     int64           `json:"category_id"`
			UsageCount     int64           `json:"usage_count"`
			RecentExpenses []ExpenseDetail `json:"recent_expenses"`
		}{
			CategoryID:     int64(categoryId),
			UsageCount:     usageCount,
			RecentExpenses: expenses,
		}

		RespondWithJSON(rw, http.StatusOK, response)
	}
}

// Bulk deactivate categories
func HandleBulkDeactivateCategories(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))

		// Parse request
		var requestBody struct {
			IDs []int64 `json:"ids"`
		}

		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&requestBody); err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}

		if len(requestBody.IDs) == 0 {
			RespondWithError(rw, http.StatusBadRequest, "No category IDs provided")
			return
		}

		// Bulk deactivate
		err := cfg.Db.BulkDeactivateCategories(req.Context(), database.BulkDeactivateCategoriesParams{
			CategoryIds:   requestBody.IDs,
			AssociationID: int64(associationId),
		})

		if err != nil {
			log.Printf("Error bulk deactivating categories: %s", err)
			RespondWithError(rw, http.StatusInternalServerError, "Failed to deactivate categories")
			return
		}

		RespondWithJSON(rw, http.StatusOK, map[string]interface{}{
			"message": "Categories deactivated successfully",
			"count":   len(requestBody.IDs),
		})
	}
}

// Bulk reactivate categories
func HandleBulkReactivateCategories(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))

		// Parse request
		var requestBody struct {
			IDs []int64 `json:"ids"`
		}

		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&requestBody); err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}

		if len(requestBody.IDs) == 0 {
			RespondWithError(rw, http.StatusBadRequest, "No category IDs provided")
			return
		}

		// Bulk reactivate
		err := cfg.Db.BulkReactivateCategories(req.Context(), database.BulkReactivateCategoriesParams{
			CategoryIds:   requestBody.IDs,
			AssociationID: int64(associationId),
		})

		if err != nil {
			log.Printf("Error bulk reactivating categories: %s", err)
			RespondWithError(rw, http.StatusInternalServerError, "Failed to reactivate categories")
			return
		}

		RespondWithJSON(rw, http.StatusOK, map[string]interface{}{
			"message": "Categories reactivated successfully",
			"count":   len(requestBody.IDs),
		})
	}
}
