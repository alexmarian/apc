package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// unitsFixture seeds two associations, each with one building and unit:
// association A (building bA, unit uA, owners alice + bob, alice the voting
// owner) and association B (building bB, unit uB, owner zed).
type unitsFixture struct {
	cfg                  *ApiConfig
	assocA, assocB       int64
	buildingA, buildingB int64
	unitA, unitB         int64
	alice, bob, zed      int64
	aliceShip            int64
}

func newUnitsFixture(t *testing.T) unitsFixture {
	t.Helper()
	cfg := newAssociationsTestConfig(t)
	f := unitsFixture{cfg: cfg}
	exec := func(query string, args ...any) int64 {
		t.Helper()
		res, err := cfg.DB.Exec(query, args...)
		if err != nil {
			t.Fatal(err)
		}
		id, _ := res.LastInsertId()
		return id
	}
	f.assocA = exec(`INSERT INTO associations (name, address, administrator) VALUES ('A', 'a', '')`)
	f.assocB = exec(`INSERT INTO associations (name, address, administrator) VALUES ('B', 'b', '')`)
	f.buildingA = exec(`INSERT INTO buildings (name, address, cadastral_number, total_area, association_id) VALUES ('Bloc A', 'Str. A', 'CB-A', 1000, ?)`, f.assocA)
	f.buildingB = exec(`INSERT INTO buildings (name, address, cadastral_number, total_area, association_id) VALUES ('Bloc B', 'Str. B', 'CB-B', 900, ?)`, f.assocB)
	f.unitA = exec(`INSERT INTO units (cadastral_number, building_id, unit_number, address, entrance, area, part, unit_type, floor, room_count) VALUES ('CU-A1', ?, '1', 'Str. A 1', 1, 50, 0.05, 'apartment', 1, 2)`, f.buildingA)
	f.unitB = exec(`INSERT INTO units (cadastral_number, building_id, unit_number, address, entrance, area, part, unit_type, floor, room_count) VALUES ('CU-B1', ?, '1', 'Str. B 1', 1, 60, 0.06, 'apartment', 2, 3)`, f.buildingB)
	f.alice = exec(`INSERT INTO owners (name, normalized_name, association_id) VALUES ('Alice', 'alice', ?)`, f.assocA)
	f.bob = exec(`INSERT INTO owners (name, normalized_name, association_id) VALUES ('Bob', 'bob', ?)`, f.assocA)
	f.zed = exec(`INSERT INTO owners (name, normalized_name, association_id) VALUES ('Zed', 'zed', ?)`, f.assocB)
	f.aliceShip = exec(`INSERT INTO ownerships (unit_id, owner_id, association_id, start_date, is_active, is_voting, registration_document, registration_date) VALUES (?, ?, ?, '2020-01-01 00:00:00', true, true, 'DOC-1', '2020-01-01 00:00:00')`, f.unitA, f.alice, f.assocA)
	exec(`INSERT INTO ownerships (unit_id, owner_id, association_id, start_date, is_active, registration_document, registration_date) VALUES (?, ?, ?, '2021-01-01 00:00:00', true, 'DOC-2', '2021-01-01 00:00:00')`, f.unitA, f.bob, f.assocA)
	return f
}

// call runs handler with the given path values (association/building/unit/
// ownership ids, 0 = unset).
func (f unitsFixture) call(handler http.HandlerFunc, method string, ids [4]int64, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/", strings.NewReader(body))
	for i, name := range []string{AssociationIdPathValue, BuildingIdPathValue, UnitIdPathValue, OwnershipIdPathValue} {
		if ids[i] != 0 {
			req.SetPathValue(name, fmt.Sprint(ids[i]))
		}
	}
	rec := httptest.NewRecorder()
	handler(rec, req)
	return rec
}

func (f unitsFixture) unitIds() [4]int64 { return [4]int64{f.assocA, f.buildingA, f.unitA, 0} }

// crossIds pairs association A with association B's building and unit.
func (f unitsFixture) crossIds() [4]int64 { return [4]int64{f.assocA, f.buildingB, f.unitB, 0} }

func TestUnits_ScopedToAssociation(t *testing.T) {
	f := newUnitsFixture(t)

	rec := f.call(HandleGetBuildingUnits(f.cfg), http.MethodGet, [4]int64{f.assocA, f.buildingA}, "")
	var units []Unit
	_ = json.Unmarshal(rec.Body.Bytes(), &units)
	if rec.Code != http.StatusOK || len(units) != 1 || units[0].OwnerNames == "" {
		t.Fatalf("own units: %d %s", rec.Code, rec.Body.String())
	}
	if rec := f.call(HandleGetBuildingUnit(f.cfg), http.MethodGet, f.unitIds(), ""); rec.Code != http.StatusOK {
		t.Fatalf("own unit: %d", rec.Code)
	}

	// Association A's id with B's building/unit: nothing leaks, nothing changes.
	for name, h := range map[string]http.HandlerFunc{
		"list":       HandleGetBuildingUnits(f.cfg),
		"get":        HandleGetBuildingUnit(f.cfg),
		"report":     HandleGetUnitReport(f.cfg),
		"ownerships": HandleGetBuildingUnitOwnerships(f.cfg),
		"owners":     HandleGetBuildingUnitOwner(f.cfg),
	} {
		if rec := f.call(h, http.MethodGet, f.crossIds(), ""); rec.Code != http.StatusNotFound {
			t.Errorf("%s across associations: %d, want 404", name, rec.Code)
		}
	}
	if rec := f.call(HandleUpdateBuildingUnit(f.cfg), http.MethodPut, f.crossIds(), `{"unit_number":"hacked"}`); rec.Code != http.StatusNotFound {
		t.Fatalf("update across associations: %d", rec.Code)
	}
	unitB, _ := f.cfg.Db.GetBuildingUnit(t.Context(), struct{ ID, BuildingID int64 }{f.unitB, f.buildingB})
	if unitB.UnitNumber != "1" {
		t.Fatal("foreign unit was modified")
	}
}

func TestUpdateUnit(t *testing.T) {
	f := newUnitsFixture(t)
	update := func(body string) *httptest.ResponseRecorder {
		return f.call(HandleUpdateBuildingUnit(f.cfg), http.MethodPut, f.unitIds(), body)
	}

	for body, want := range map[string]string{
		`{"unit_number":"  "}`: "Unit number is required",
		`{"entrance":-3}`:      "Entrance must be at least -2",
		`{"room_count":-1}`:    "Room count must be at least 0",
		`{"area":10}`:          "Area, share and cadastral number can only be changed by an application admin",
		`{"unit_type":""}`:     "Unit type is required",
	} {
		if rec := update(body); rec.Code != http.StatusBadRequest || errorMsg(rec) != want {
			t.Errorf("%s: %d %q, want 400 %q", body, rec.Code, errorMsg(rec), want)
		}
	}

	rec := update(`{"unit_number":" 1A ","entrance":-2,"room_count":4,"unit_type":"office"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("update: %d %s", rec.Code, rec.Body.String())
	}
	var got Unit
	_ = json.Unmarshal(rec.Body.Bytes(), &got)
	if got.UnitNumber != "1A" || got.Entrance != -2 || got.RoomCount != 4 || got.UnitType != "office" || got.Floor != 1 || got.Address != "Str. A 1" {
		t.Fatalf("after update: %+v", got)
	}
}

func TestUnitReport(t *testing.T) {
	f := newUnitsFixture(t)
	// A past ownership of the same owner must not duplicate alice.
	if _, err := f.cfg.DB.Exec(`INSERT INTO ownerships (unit_id, owner_id, association_id, start_date, end_date, is_active, registration_document, registration_date) VALUES (?, ?, ?, '2010-01-01', '2019-12-31', false, 'OLD', '2010-01-01')`, f.unitA, f.alice, f.assocA); err != nil {
		t.Fatal(err)
	}
	rec := f.call(HandleGetUnitReport(f.cfg), http.MethodGet, f.unitIds(), "")
	var report struct {
		Building struct{ Name string } `json:"building_details"`
		Owners   []struct {
			Name     string
			IsActive bool `json:"is_active"`
		} `json:"current_owners"`
		History []struct {
			OwnerName string `json:"owner_name"`
			IsVoting  bool   `json:"is_voting"`
			IsActive  bool   `json:"is_active"`
		} `json:"ownership_history"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &report)
	if rec.Code != http.StatusOK || report.Building.Name != "Bloc A" || len(report.Owners) != 2 || len(report.History) != 3 {
		t.Fatalf("report: %d %s", rec.Code, rec.Body.String())
	}
	voting := 0
	for _, h := range report.History {
		if h.IsVoting {
			voting++
			if h.OwnerName != "Alice" {
				t.Fatalf("voting owner = %s", h.OwnerName)
			}
		}
	}
	if voting != 1 {
		t.Fatalf("voting ownerships in history = %d", voting)
	}
}

func TestCreateOwnership(t *testing.T) {
	f := newUnitsFixture(t)
	create := func(body string) *httptest.ResponseRecorder {
		return f.call(HandleCreateUnitOwnership(f.cfg), http.MethodPost, f.unitIds(), body)
	}
	body := func(owner int64, extra string) string {
		return fmt.Sprintf(`{"owner_id":%d,"start_date":"2024-03-01T00:00:00Z","registration_document":"DOC-9","registration_date":"2024-03-02T00:00:00Z"%s}`, owner, extra)
	}

	if rec := create(body(f.zed, "")); rec.Code != http.StatusBadRequest {
		t.Fatalf("owner from another association: %d", rec.Code)
	}
	if rec := create(fmt.Sprintf(`{"owner_id":%d,"start_date":"2024-03-01T00:00:00Z","registration_date":"2024-03-02T00:00:00Z"}`, f.alice)); rec.Code != http.StatusBadRequest {
		t.Fatalf("missing document: %d", rec.Code)
	}
	if rec := create(body(f.alice, `,"end_date":"2023-01-01T00:00:00Z"`)); rec.Code != http.StatusBadRequest {
		t.Fatalf("end before start: %d", rec.Code)
	}
	if rec := create(body(f.alice, "")); rec.Code != http.StatusConflict {
		t.Fatalf("second active ownership for the same owner: %d", rec.Code)
	}

	// A new owner, exclusively: everyone else's ownership ends, the new one votes.
	carol := f.call(HandleCreateOwner(f.cfg), http.MethodPost, [4]int64{f.assocA}, `{"name":"Carol"}`)
	var owner Owner
	_ = json.Unmarshal(carol.Body.Bytes(), &owner)
	rec := create(body(owner.ID, `,"is_exclusive":true`))
	if rec.Code != http.StatusCreated {
		t.Fatalf("exclusive create: %d %s", rec.Code, rec.Body.String())
	}
	var created Ownership
	_ = json.Unmarshal(rec.Body.Bytes(), &created)
	if created.OwnerName != "Carol" || !created.IsVoting || !created.IsActive || created.StartDate.Year() != 2024 {
		t.Fatalf("created: %+v", created)
	}
	ships, _ := f.cfg.Db.GetActiveUnitOwnerships(t.Context(), struct{ UnitID, AssociationID int64 }{f.unitA, f.assocA})
	if len(ships) != 1 || ships[0].OwnerID != owner.ID {
		t.Fatalf("active ownerships after exclusive create: %+v", ships)
	}
}

func TestSetVotingOwner(t *testing.T) {
	f := newUnitsFixture(t)
	setVoting := func(ownershipId int64) *httptest.ResponseRecorder {
		ids := f.unitIds()
		ids[3] = ownershipId
		return f.call(HandleUnitVotingOwnership(f.cfg), http.MethodPost, ids, "")
	}
	votingOwner := func() string {
		ships, _ := f.cfg.Db.GetUnitOwnerships(t.Context(), struct{ UnitID, AssociationID int64 }{f.unitA, f.assocA})
		for _, s := range ships {
			if s.IsVoting {
				return s.OwnerName
			}
		}
		return ""
	}

	// An unknown ownership leaves the current voting owner in place.
	if rec := setVoting(999); rec.Code != http.StatusNotFound || votingOwner() != "Alice" {
		t.Fatalf("unknown ownership: %d, voting owner %q", rec.Code, votingOwner())
	}
	ships, _ := f.cfg.Db.GetActiveUnitOwnerships(t.Context(), struct{ UnitID, AssociationID int64 }{f.unitA, f.assocA})
	var bobShip int64
	for _, s := range ships {
		if s.OwnerName == "Bob" {
			bobShip = s.ID
		}
	}
	if rec := setVoting(bobShip); rec.Code != http.StatusOK || votingOwner() != "Bob" {
		t.Fatalf("switch voting owner: %d, voting owner %q", rec.Code, votingOwner())
	}

	// Ended ownerships can't vote.
	disable := f.call(HandleDisableOwnership(f.cfg), http.MethodPost, [4]int64{f.assocA, 0, 0, f.aliceShip}, `{}`)
	if disable.Code != http.StatusOK {
		t.Fatalf("disable: %d", disable.Code)
	}
	if rec := setVoting(f.aliceShip); rec.Code != http.StatusBadRequest || votingOwner() != "Bob" {
		t.Fatalf("inactive ownership: %d, voting owner %q", rec.Code, votingOwner())
	}
}
