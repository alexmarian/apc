package handlers

import (
	"encoding/json"
	"math"
	"net/http"
	"testing"
	"time"
)

type distributionResponse struct {
	DistributionMethod string  `json:"distribution_method"`
	TotalUnits         int     `json:"total_units"`
	TotalArea          float64 `json:"total_area"`
	TotalExpenses      float64 `json:"total_expenses"`
	CategoryTotals     map[string]struct {
		Amount float64 `json:"amount"`
		ID     int64   `json:"id"`
		Name   string  `json:"name"`
	} `json:"category_totals"`
	UnitDistributions []struct {
		ID            int64              `json:"id"`
		UnitNumber    string             `json:"unit_number"`
		ExpensesShare map[string]float64 `json:"expenses_share"`
		TotalShare    float64            `json:"total_share"`
	} `json:"unit_distributions"`
}

func TestExpenseDistribution(t *testing.T) {
	b := newBookkeeping(t)
	exec := func(q string, args ...any) int64 {
		res, err := b.cfg.DB.Exec(q, args...)
		if err != nil {
			t.Fatal(err)
		}
		id, _ := res.LastInsertId()
		return id
	}
	// More units in building A, inserted out of order: 10 (area 100), 2 (area 50); unit 1 has 50.
	exec(`INSERT INTO units (cadastral_number, building_id, unit_number, address, entrance, area, part, unit_type, floor, room_count) VALUES ('CU-A10', ?, '10', 'x', 1, 100, 0.1, 'apartment', 3, 2)`, b.buildingA)
	exec(`INSERT INTO units (cadastral_number, building_id, unit_number, address, entrance, area, part, unit_type, floor, room_count) VALUES ('CU-A2', ?, '2', 'x', 1, 50, 0.05, 'commercial', 1, 1)`, b.buildingA)
	// A second "janitor_salary", in another family: its own column.
	twin := exec(`INSERT INTO categories (type, family, name, association_id) VALUES ('Budgeted', 'cleaning', 'janitor_salary', ?)`, b.assocA)
	water := exec(`INSERT INTO categories (type, family, name, association_id) VALUES ('Utilities', 'water', 'cold_water', ?)`, b.assocA)
	// Stored as the app stores them (a time.Time, not a bare string).
	expense := func(category int64, amount float64, day int) {
		date := time.Date(2026, time.January, day, 0, 0, 0, 0, time.UTC)
		exec(`INSERT INTO expenses (amount, description, destination, date, month, year, category_id, account_id) VALUES (?, 'x', 'y', ?, 1, 2026, ?, ?)`, amount, date, category, b.account)
	}
	expense(b.category, 100, 1)
	expense(twin, 40, 31)
	expense(water, 200, 15)
	expense(water, 999, 32) // February 1st: outside the range

	get := func(query string) (int, distributionResponse) {
		rec := b.req(HandleExpenseDistributionReport(b.cfg), http.MethodGet, "?"+query, "", nil)
		var got distributionResponse
		_ = json.Unmarshal(rec.Body.Bytes(), &got)
		return rec.Code, got
	}

	code, got := get("start_date=2026-01-01&end_date=2026-01-31")
	if code != http.StatusOK || got.DistributionMethod != "area" || got.TotalUnits != 3 || got.TotalArea != 200 || got.TotalExpenses != 340 {
		t.Fatalf("area: %d %+v", code, got)
	}
	if len(got.CategoryTotals) != 3 {
		t.Fatalf("category keys: %+v", got.CategoryTotals)
	}
	first, second := got.CategoryTotals["Budgeted/personnel/janitor_salary"], got.CategoryTotals["Budgeted/cleaning/janitor_salary"]
	if first.Amount != 100 || second.Amount != 40 || first.Name != "janitor_salary" || got.CategoryTotals["cold_water"].Amount != 200 {
		t.Fatalf("category totals: %+v", got.CategoryTotals)
	}
	var order []string
	for _, u := range got.UnitDistributions {
		order = append(order, u.UnitNumber)
	}
	if len(order) != 3 || order[0] != "1" || order[1] != "2" || order[2] != "10" {
		t.Fatalf("unit order: %v", order)
	}
	unit10 := got.UnitDistributions[2]
	if math.Abs(unit10.TotalShare-170) > 1e-9 || math.Abs(unit10.ExpensesShare["cold_water"]-100) > 1e-9 {
		t.Fatalf("unit 10 share: %+v", unit10)
	}

	// Per unit: equal shares; total_area is still the units' area.
	_, got = get("start_date=2026-01-01&end_date=2026-01-31&distribution_method=count&unit_type=apartment")
	if got.TotalUnits != 2 || got.TotalArea != 150 || math.Abs(got.UnitDistributions[0].TotalShare-170) > 1e-9 {
		t.Fatalf("count: %+v", got)
	}

	// A unique name keeps its plain key (the old app's columns).
	_, got = get("start_date=2026-01-01&end_date=2026-01-31&category_type=Utilities")
	if _, ok := got.CategoryTotals["cold_water"]; !ok || len(got.CategoryTotals) != 1 {
		t.Fatalf("filtered: %+v", got.CategoryTotals)
	}

	// Nothing in range: the same shape, empty.
	rec := b.req(HandleExpenseDistributionReport(b.cfg), http.MethodGet, "?start_date=2025-01-01&end_date=2025-01-31&distribution_method=equal", "", nil)
	var empty map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &empty)
	if empty["distribution_method"] != "equal" || empty["start_date"] == nil || len(empty["unit_distributions"].([]any)) != 0 {
		t.Fatalf("empty: %s", rec.Body.String())
	}

	if code, _ := get("distribution_method=volume"); code != http.StatusBadRequest {
		t.Fatalf("bad method: %d", code)
	}
}
