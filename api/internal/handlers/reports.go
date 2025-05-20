package handlers

import (
	"github.com/alexmarian/apc/api/internal/database"
	"github.com/alexmarian/apc/api/internal/logging"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

func HandleExpenseDistributionReport(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))

		// Parse query parameters for time range
		startDate, endDate, err := GetRequestDateRange(req, &rw)
		if err != nil {
			RespondWithError(rw, http.StatusBadRequest, err.Error())
			return
		}

		// Parse unit type filter (optional)
		unitType := req.URL.Query().Get("unit_type")

		// Parse category filters (optional)
		categoryType := req.URL.Query().Get("category_type")
		categoryFamily := req.URL.Query().Get("category_family")
		categoryIdStr := req.URL.Query().Get("category_id")

		var categoryId int64 = 0
		if categoryIdStr != "" {
			categoryId, err = strconv.ParseInt(categoryIdStr, 10, 64)
			if err != nil {
				RespondWithError(rw, http.StatusBadRequest, "Invalid category_id parameter")
				return
			}

			// Verify category exists and belongs to this association
			if categoryId > 0 {
				category, err := cfg.Db.GetCategory(req.Context(), categoryId)
				if err != nil {
					logging.Logger.Log(zap.WarnLevel, "Error getting category", zap.String("error", err.Error()))
					RespondWithError(rw, http.StatusNotFound, "Category not found")
					return
				}

				if category.AssociationID != int64(associationId) {
					RespondWithError(rw, http.StatusForbidden, "Category does not belong to this association")
					return
				}
			}
		}

		// Parse distribution method (area, count, equal)
		distributionMethod := req.URL.Query().Get("distribution_method")
		if distributionMethod == "" {
			distributionMethod = "area" // Default to area-based distribution
		}

		if distributionMethod != "area" && distributionMethod != "count" && distributionMethod != "equal" {
			RespondWithError(rw, http.StatusBadRequest, "Invalid distribution_method. Must be 'area', 'count', or 'equal'")
			return
		}

		// Step 1: Get all expenses for the specified date range
		expenses, err := cfg.Db.GetExpensesByDateRangeWithFilters(req.Context(), database.GetExpensesByDateRangeWithFiltersParams{
			AssociationID: int64(associationId),
			Date:          startDate,
			Date_2:        endDate,
			CategoryID:    categoryId,     // Will be 0 if not provided
			Column4:       categoryId,     // Same value, needed twice for SQL
			Column6:       categoryType,   // Will be empty if not provided
			Type:          categoryType,   // Same value
			Column8:       categoryFamily, // Will be empty if not provided
			Family:        categoryFamily, // Same value
		})
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting expenses", zap.String("error", err.Error()))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to retrieve expenses")
			return
		}

		// Filter expenses by category criteria if provided
		var filteredExpenses []database.GetExpensesByDateRangeWithFiltersRow
		for _, expense := range expenses {
			// Apply category filters
			if categoryId > 0 && expense.CategoryID != categoryId {
				continue
			}

			if categoryType != "" && expense.CategoryType != categoryType {
				continue
			}

			if categoryFamily != "" && expense.CategoryFamily != categoryFamily {
				continue
			}

			filteredExpenses = append(filteredExpenses, expense)
		}

		// If we filtered out all expenses, return early
		if len(filteredExpenses) == 0 {
			RespondWithError(rw, http.StatusNotFound, "No expenses found matching the criteria")
			return
		}

		// Step 2: Get all buildings in the association
		buildings, err := cfg.Db.GetAssociationBuildings(req.Context(), int64(associationId))
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting buildings", zap.String("error", err.Error()))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to retrieve buildings")
			return
		}

		// Step 3: Get all units for each building
		type UnitInfo struct {
			ID              int64   `json:"id"`
			BuildingID      int64   `json:"building_id"`
			UnitNumber      string  `json:"unit_number"`
			BuildingName    string  `json:"building_name"`
			BuildingAddress string  `json:"building_address"`
			UnitType        string  `json:"unit_type"`
			Area            float64 `json:"area"`
			Part            float64 `json:"part"`
		}

		units := []UnitInfo{}

		for _, building := range buildings {
			buildingUnits, err := cfg.Db.GetBuildingUnits(req.Context(), building.ID)
			if err != nil {
				logging.Logger.Log(zap.WarnLevel, "Error getting units for building", zap.Int64("building_id", building.ID), zap.String("error", err.Error()))
				continue
			}

			for _, unit := range buildingUnits {
				// Filter by unit type if specified
				if unitType != "" && unit.UnitType != unitType {
					continue
				}

				units = append(units, UnitInfo{
					ID:              unit.ID,
					BuildingID:      unit.BuildingID,
					UnitNumber:      unit.UnitNumber,
					BuildingName:    building.Name,
					BuildingAddress: building.Address,
					UnitType:        unit.UnitType,
					Area:            unit.Area,
					Part:            unit.Part,
				})
			}
		}

		if len(units) == 0 {
			RespondWithError(rw, http.StatusNotFound, "No units found matching the criteria")
			return
		}
		type ExpenseShare struct {
			ExpenseID      int64     `json:"expense_id"`
			Description    string    `json:"description"`
			Date           time.Time `json:"date"`
			CategoryName   string    `json:"category_name"`
			CategoryType   string    `json:"category_type"`
			CategoryFamily string    `json:"category_family"`
			CategoryID     int64     `json:"category_id"`
			TotalAmount    float64   `json:"total_amount"`
			UnitShare      float64   `json:"unit_share"`
		}
		// Step 4: Calculate distribution factors
		type UnitDistribution struct {
			UnitInfo
			DistributionFactor float64                `json:"distribution_factor"`
			ExpensesShare      map[string]float64     `json:"expenses_share"` // By category
			TotalShare         float64                `json:"total_share"`
			DetailedExpenses   map[int64]ExpenseShare `json:"detailed_expenses,omitempty"`
		}

		unitDistributions := make(map[int64]*UnitDistribution)

		// Calculate total for distribution factors based on the method
		var totalFactor float64 = 0

		switch distributionMethod {
		case "area":
			for _, unit := range units {
				totalFactor += unit.Area
			}
		case "count":
			totalFactor = float64(len(units))
		case "equal":
			totalFactor = float64(len(units)) // Same as count for equal distribution
		}

		// Calculate distribution factor for each unit
		for _, unit := range units {
			var factor float64

			switch distributionMethod {
			case "area":
				factor = unit.Area / totalFactor
			case "count", "equal":
				factor = 1.0 / totalFactor
			}

			unitDistributions[unit.ID] = &UnitDistribution{
				UnitInfo:           unit,
				DistributionFactor: factor,
				ExpensesShare:      make(map[string]float64),
				DetailedExpenses:   make(map[int64]ExpenseShare),
				TotalShare:         0,
			}
		}

		// Step 5: Distribute expenses
		type CategoryTotal struct {
			Total          float64 `json:"total"`
			Count          int     `json:"count"`
			CategoryID     int64   `json:"category_id"`
			CategoryType   string  `json:"category_type"`
			CategoryFamily string  `json:"category_family"`
		}

		categoryTotals := make(map[string]*CategoryTotal)

		for _, expense := range filteredExpenses {
			categoryKey := expense.CategoryName

			// Initialize category in the map if it doesn't exist
			if _, exists := categoryTotals[categoryKey]; !exists {
				categoryTotals[categoryKey] = &CategoryTotal{
					Total:          0,
					Count:          0,
					CategoryID:     expense.CategoryID,
					CategoryType:   expense.CategoryType,
					CategoryFamily: expense.CategoryFamily,
				}
			}

			// Update category totals
			categoryTotals[categoryKey].Total += expense.Amount
			categoryTotals[categoryKey].Count++

			// Distribute this expense to each unit
			for _, unitDist := range unitDistributions {
				unitShare := expense.Amount * unitDist.DistributionFactor

				// Add to category total for this unit
				unitDist.ExpensesShare[categoryKey] += unitShare

				// Add to unit's total share
				unitDist.TotalShare += unitShare

				// Add detailed expense record
				unitDist.DetailedExpenses[expense.ID] = ExpenseShare{
					ExpenseID:      expense.ID,
					Description:    expense.Description,
					Date:           expense.Date,
					CategoryName:   expense.CategoryName,
					CategoryType:   expense.CategoryType,
					CategoryFamily: expense.CategoryFamily,
					CategoryID:     expense.CategoryID,
					TotalAmount:    expense.Amount,
					UnitShare:      unitShare,
				}
			}
		}
		type CategoryDetail struct {
			Amount float64 `json:"amount"`
			ID     int64   `json:"id"`
			Type   string  `json:"type"`
			Family string  `json:"family"`
		}
		// Step 6: Prepare response
		type ExpenseDistributionResponse struct {
			StartDate          time.Time                 `json:"start_date"`
			EndDate            time.Time                 `json:"end_date"`
			UnitType           string                    `json:"unit_type,omitempty"`
			CategoryType       string                    `json:"category_type,omitempty"`
			CategoryFamily     string                    `json:"category_family,omitempty"`
			CategoryID         int64                     `json:"category_id,omitempty"`
			DistributionMethod string                    `json:"distribution_method"`
			TotalUnits         int                       `json:"total_units"`
			TotalArea          float64                   `json:"total_area"`
			CategoryTotals     map[string]CategoryDetail `json:"category_totals"`
			TotalExpenses      float64                   `json:"total_expenses"`
			UnitDistributions  []*UnitDistribution       `json:"unit_distributions"`
		}

		response := ExpenseDistributionResponse{
			StartDate:          startDate,
			EndDate:            endDate,
			UnitType:           unitType,
			CategoryType:       categoryType,
			CategoryFamily:     categoryFamily,
			CategoryID:         categoryId,
			DistributionMethod: distributionMethod,
			TotalUnits:         len(units),
			TotalArea:          totalFactor,
			CategoryTotals:     make(map[string]CategoryDetail),
			TotalExpenses:      0,
			UnitDistributions:  []*UnitDistribution{},
		}

		// Calculate total expenses by category
		for category, totals := range categoryTotals {
			response.CategoryTotals[category] = CategoryDetail{
				Amount: totals.Total,
				ID:     totals.CategoryID,
				Type:   totals.CategoryType,
				Family: totals.CategoryFamily,
			}
			response.TotalExpenses += totals.Total
		}

		// Convert unit distributions map to slice
		for _, unitDist := range unitDistributions {
			response.UnitDistributions = append(response.UnitDistributions, unitDist)
		}

		// Check if we should include detailed expenses
		includeDetails := req.URL.Query().Get("include_details") == "true"
		if !includeDetails {
			// Remove detailed expenses to reduce response size
			for _, unitDist := range response.UnitDistributions {
				unitDist.DetailedExpenses = nil
			}
		}

		RespondWithJSON(rw, http.StatusOK, response)
	}
}
