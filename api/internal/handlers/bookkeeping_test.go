package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// bookkeeping seeds association A (from unitsFixture) with an account and
// a category of its own, a disabled account and an inactive category, and
// association B with one category.
type bookkeeping struct {
	unitsFixture
	account, disabledAccount, category, inactiveCategory, foreignCategory int64
}

func newBookkeeping(t *testing.T) bookkeeping {
	t.Helper()
	b := bookkeeping{unitsFixture: newUnitsFixture(t)}
	exec := func(q string, args ...any) int64 {
		res, err := b.cfg.DB.Exec(q, args...)
		if err != nil {
			t.Fatal(err)
		}
		id, _ := res.LastInsertId()
		return id
	}
	b.account = exec(`INSERT INTO accounts (number, destination, description, association_id) VALUES ('RO01', 'Bank', 'Main', ?)`, b.assocA)
	b.disabledAccount = exec(`INSERT INTO accounts (number, destination, description, association_id, is_active) VALUES ('RO02', 'Bank', 'Old', ?, false)`, b.assocA)
	b.category = exec(`INSERT INTO categories (type, family, name, association_id) VALUES ('Budgeted', 'personnel', 'janitor_salary', ?)`, b.assocA)
	b.inactiveCategory = exec(`INSERT INTO categories (type, family, name, association_id, is_deleted) VALUES ('Budgeted', 'personnel', 'old_salary', ?, true)`, b.assocA)
	b.foreignCategory = exec(`INSERT INTO categories (type, family, name, association_id) VALUES ('Budgeted', 'personnel', 'janitor_salary', ?)`, b.assocB)
	return b
}

// req calls handler with association A and the given extra path values.
func (b bookkeeping) req(handler http.HandlerFunc, method, query, body string, pathValues map[string]int64) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, "/"+query, strings.NewReader(body))
	r.SetPathValue(AssociationIdPathValue, fmt.Sprint(b.assocA))
	for k, v := range pathValues {
		r.SetPathValue(k, fmt.Sprint(v))
	}
	rec := httptest.NewRecorder()
	handler(rec, r)
	return rec
}

func (b bookkeeping) expenseBody(category, account int64, extra string) string {
	return fmt.Sprintf(`{"amount":120.5,"description":"  Salary October  ","destination":"Ion","date":"2026-10-01T00:00:00Z","category_id":%d,"account_id":%d%s}`, category, account, extra)
}

func TestExpenses(t *testing.T) {
	b := newBookkeeping(t)
	create := func(body string) *httptest.ResponseRecorder {
		return b.req(HandleCreateExpense(b.cfg), http.MethodPost, "", body, nil)
	}

	for name, tc := range map[string]struct{ body, msg string }{
		"inactive category": {b.expenseBody(b.inactiveCategory, b.account, ""), "The category is inactive"},
		"disabled account":  {b.expenseBody(b.category, b.disabledAccount, ""), "The account is disabled"},
		"foreign category":  {b.expenseBody(b.foreignCategory, b.account, ""), "Invalid category"},
		"long doc ref":      {b.expenseBody(b.category, b.account, `,"document_ref":"`+strings.Repeat("x", 51)+`"`), "Document reference must be at most 50 characters"},
		"zero amount":       {strings.Replace(b.expenseBody(b.category, b.account, ""), "120.5", "0", 1), "Amount must be positive"},
	} {
		if rec := create(tc.body); rec.Code != http.StatusBadRequest || errorMsg(rec) != tc.msg {
			t.Errorf("%s: %d %q", name, rec.Code, errorMsg(rec))
		}
	}

	rec := create(b.expenseBody(b.category, b.account, `,"document_ref":"F-1"`))
	var created ExpenseItem
	_ = json.Unmarshal(rec.Body.Bytes(), &created)
	if rec.Code != http.StatusCreated || created.Description != "Salary October" || created.CategoryName != "janitor_salary" || created.AccountNumber != "RO01" {
		t.Fatalf("create: %d %s", rec.Code, rec.Body.String())
	}

	// The list includes an expense dated on the first day of the range.
	rec = b.req(HandleGetExpenses(b.cfg), http.MethodGet, "?start_date=2026-10-01&end_date=2026-10-31", "", nil)
	var list []ExpenseItem
	_ = json.Unmarshal(rec.Body.Bytes(), &list)
	if rec.Code != http.StatusOK || len(list) != 1 {
		t.Fatalf("list from the expense's own date: %d %s", rec.Code, rec.Body.String())
	}

	// Update answers with names too; switching to a disabled account is refused,
	// keeping the current (since deactivated) category is fine.
	if _, err := b.cfg.DB.Exec(`UPDATE categories SET is_deleted = true WHERE id = ?`, b.category); err != nil {
		t.Fatal(err)
	}
	update := func(body string) *httptest.ResponseRecorder {
		return b.req(HandleUpdateExpense(b.cfg), http.MethodPut, "", body, map[string]int64{ExpensesIdPathValue: created.ID})
	}
	if rec := update(fmt.Sprintf(`{"account_id":%d}`, b.disabledAccount)); rec.Code != http.StatusBadRequest {
		t.Fatalf("switch to disabled account: %d", rec.Code)
	}
	rec = update(fmt.Sprintf(`{"amount":99,"category_id":%d}`, b.category))
	var updated ExpenseItem
	_ = json.Unmarshal(rec.Body.Bytes(), &updated)
	if rec.Code != http.StatusOK || updated.Amount != 99 || updated.CategoryType != "Budgeted" || updated.AccountNumber != "RO01" {
		t.Fatalf("update: %d %s", rec.Code, rec.Body.String())
	}
}

func TestCategories(t *testing.T) {
	b := newBookkeeping(t)
	create := func(body string) *httptest.ResponseRecorder {
		return b.req(HandleCreateCategory(b.cfg), http.MethodPost, "", body, nil)
	}
	if rec := create(`{"type":"Budgeted","family":"personnel","name":" janitor_salary "}`); rec.Code != http.StatusConflict {
		t.Fatalf("duplicate create: %d", rec.Code)
	}
	// The inactive duplicate doesn't block creating an active one…
	rec := create(`{"type":"Budgeted","family":"personnel","name":"old_salary","original_labels":{"name":"Old salary"}}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", rec.Code, rec.Body.String())
	}
	// …but then the inactive one can't come back, alone or in bulk.
	if rec := b.req(HandleReactivateCategory(b.cfg), http.MethodPut, "", "", map[string]int64{CategoryIdPathValue: b.inactiveCategory}); rec.Code != http.StatusConflict {
		t.Fatalf("reactivate duplicate: %d", rec.Code)
	}
	if rec := b.req(HandleBulkReactivateCategories(b.cfg), http.MethodPost, "", fmt.Sprintf(`{"ids":[%d]}`, b.inactiveCategory), nil); rec.Code != http.StatusConflict {
		t.Fatalf("bulk reactivate duplicate: %d %s", rec.Code, rec.Body.String())
	}

	if rec := b.req(HandleGetCategory(b.cfg), http.MethodGet, "", "", map[string]int64{CategoryIdPathValue: b.foreignCategory}); rec.Code != http.StatusNotFound {
		t.Fatalf("get another association's category: %d", rec.Code)
	}
	if rec := b.req(HandleUpdateCategory(b.cfg), http.MethodPut, "", `{"type":"a","family":"b","name":"c"}`, map[string]int64{CategoryIdPathValue: b.foreignCategory}); rec.Code != http.StatusNotFound {
		t.Fatalf("update another association's category: %d", rec.Code)
	}
}

func TestAccounts(t *testing.T) {
	b := newBookkeeping(t)
	if rec := b.req(HandleCreateAccount(b.cfg), http.MethodPost, "", `{"number":"  ","description":"x"}`, nil); rec.Code != http.StatusBadRequest {
		t.Fatalf("blank number: %d", rec.Code)
	}
	if rec := b.req(HandleUpdateAccount(b.cfg), http.MethodPut, "", `{"number":"RO02","description":"Renamed"}`, map[string]int64{AccountIdPathValue: b.disabledAccount}); rec.Code != http.StatusConflict {
		t.Fatalf("edit a disabled account: %d", rec.Code)
	}
	rec := b.req(HandleUpdateAccount(b.cfg), http.MethodPut, "", `{"number":" RO01 ","description":"Main account","destination":""}`, map[string]int64{AccountIdPathValue: b.account})
	var a Account
	_ = json.Unmarshal(rec.Body.Bytes(), &a)
	if rec.Code != http.StatusOK || a.Number != "RO01" || a.Description != "Main account" {
		t.Fatalf("update: %d %s", rec.Code, rec.Body.String())
	}
}
