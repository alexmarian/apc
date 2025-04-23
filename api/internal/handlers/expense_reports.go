package handlers

import (
	"fmt"
	"github.com/alexmarian/apc/api/internal/database"
	"log"
	"net/http"
	"strconv"
	"time"
)

type ExpenseSummary struct {
	TotalAmount           float64            `json:"total_amount"`
	TypeSummary           map[string]float64 `json:"type_summary"`
	CategorySummary       map[string]float64 `json:"category_summary"`
	CategoryFamilySummary map[string]float64 `json:"category_family_summary"`
	AccountSummary        map[string]float64 `json:"account_summary"`
	MonthSummary          map[string]float64 `json:"month_summary"`
}

func HandleGetExpenseReport(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))

		// Parse query parameters for time range
		startDate, endDate, err := GetRequestDateRange(req, &rw)
		if err != nil {
			RespondWithError(rw, http.StatusBadRequest, err.Error())
			return
		}

		// Get expenses for specified date range
		expenses, err := cfg.Db.GetExpensesByDateRange(req.Context(), database.GetExpensesByDateRangeParams{
			AssociationID: int64(associationId),
			Date:          startDate,
			Date_2:        endDate,
		})

		if err != nil {
			log.Printf("Error getting expenses: %s", err)
			RespondWithError(rw, http.StatusInternalServerError, "Failed to retrieve expenses")
			return
		}

		type ExpenseReport struct {
			StartDate time.Time      `json:"start_date"`
			EndDate   time.Time      `json:"end_date"`
			Summary   ExpenseSummary `json:"summary"`
			Expenses  []ExpenseItem  `json:"expenses"`
		}

		// Create the report
		report := ExpenseReport{
			StartDate: startDate,
			EndDate:   endDate,
			Summary: ExpenseSummary{
				TotalAmount:           0,
				TypeSummary:           make(map[string]float64),
				CategorySummary:       make(map[string]float64),
				CategoryFamilySummary: make(map[string]float64),
				AccountSummary:        make(map[string]float64),
				MonthSummary:          make(map[string]float64),
			},
			Expenses: []ExpenseItem{},
		}

		// Process each expense
		for _, exp := range expenses {
			// Add to total
			report.Summary.TotalAmount += exp.Amount
			//Add to type summary
			typeKey := exp.CategoryType
			report.Summary.TypeSummary[typeKey] += exp.Amount

			// Add to category summary
			categoryKey := exp.CategoryName
			report.Summary.CategorySummary[categoryKey] += exp.Amount

			// Add to family summary
			familyKey := exp.CategoryFamily
			report.Summary.CategoryFamilySummary[familyKey] += exp.Amount

			// Add to account summary
			accountKey := exp.AccountNumber
			report.Summary.AccountSummary[accountKey] += exp.Amount

			// Add to month summary
			monthKey := fmt.Sprintf("%d-%02d", exp.Year, exp.Month)
			report.Summary.MonthSummary[monthKey] += exp.Amount

			// Add expense to list
			expenseItem := ExpenseItem{
				ID:             exp.ID,
				Amount:         exp.Amount,
				Description:    exp.Description,
				Destination:    exp.Destination,
				Date:           exp.Date,
				CategoryID:     exp.CategoryID,
				CategoryType:   exp.CategoryType,
				CategoryFamily: exp.CategoryFamily,
				CategoryName:   exp.CategoryName,
				AccountID:      exp.AccountID,
				AccountNumber:  exp.AccountNumber,
				AccountName:    exp.AccountName,
			}

			report.Expenses = append(report.Expenses, expenseItem)
		}

		RespondWithJSON(rw, http.StatusOK, report)
	}
}
