package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func (f unitsFixture) ownerCall(handler http.HandlerFunc, method string, ownerId int64, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/?units=true&co_owners=true", strings.NewReader(body))
	req.SetPathValue(AssociationIdPathValue, fmt.Sprint(f.assocA))
	if ownerId != 0 {
		req.SetPathValue(OwnerIdPathValue, fmt.Sprint(ownerId))
	}
	rec := httptest.NewRecorder()
	handler(rec, req)
	return rec
}

func TestUpdateOwner(t *testing.T) {
	f := newUnitsFixture(t)
	update := func(ownerId int64, body string) *httptest.ResponseRecorder {
		return f.ownerCall(HandleUpdateAssociationOwner(f.cfg), http.MethodPut, ownerId, body)
	}

	if rec := update(f.zed, `{"name":"Zed"}`); rec.Code != http.StatusNotFound {
		t.Fatalf("owner of another association: %d", rec.Code)
	}
	if rec := update(f.alice, `{"name":"  "}`); rec.Code != http.StatusBadRequest || errorMsg(rec) != "Owner name is required" {
		t.Fatalf("empty name: %d %s", rec.Code, rec.Body.String())
	}
	if rec := update(f.alice, `{"name":"Alice","contact_email":"not-an-email"}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("bad email: %d", rec.Code)
	}

	rec := update(f.alice, `{"name":" Alice Pop ","identification_number":"123","contact_phone":"0711","contact_email":"alice@x.ro"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("update: %d %s", rec.Code, rec.Body.String())
	}
	// A full replace: an empty value clears the field.
	rec = update(f.alice, `{"name":"Alice Pop","identification_number":"123"}`)
	var got Owner
	_ = json.Unmarshal(rec.Body.Bytes(), &got)
	stored, _ := f.cfg.Db.GetOwnerById(t.Context(), f.alice)
	if got.Name != "Alice Pop" || stored.NormalizedName != "alice pop" || stored.ContactPhone != "" || stored.ContactEmail != "" || stored.IdentificationNumber != "123" {
		t.Fatalf("after update: response %+v, stored %+v", got, stored)
	}

	if rec := f.ownerCall(HandleGetAssociationOwner(f.cfg), http.MethodGet, f.zed, ""); rec.Code != http.StatusNotFound {
		t.Fatalf("get owner of another association: %d", rec.Code)
	}
}

func TestOwnerReport(t *testing.T) {
	f := newUnitsFixture(t)
	rec := f.ownerCall(HandleGetOwnerReport(f.cfg), http.MethodGet, 0, "")
	var report []struct {
		Owner      Owner `json:"owner"`
		Statistics struct {
			TotalUnits int     `json:"total_units"`
			TotalPart  float64 `json:"total_condo_part"`
		} `json:"statistics"`
		CoOwners []struct {
			Name           string   `json:"name"`
			SharedUnitNums []string `json:"shared_unit_nums"`
		} `json:"co_owners"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &report)
	// Alice and Bob co-own unit A1; zed (association B) is not listed.
	if rec.Code != http.StatusOK || len(report) != 2 || report[0].Owner.Name != "Alice" || report[1].Owner.Name != "Bob" {
		t.Fatalf("report: %d %s", rec.Code, rec.Body.String())
	}
	if report[0].Statistics.TotalUnits != 1 || len(report[0].CoOwners) != 1 || report[0].CoOwners[0].Name != "Bob" {
		t.Fatalf("alice: %+v", report[0])
	}
}
