package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// asRole calls handler through the auth middleware as root (app admin) or
// ana (association admin), with path ids as in unitsFixture.call.
func (f unitsFixture) asRole(t *testing.T, admin bool, handler http.HandlerFunc, method string, ids [4]int64, query, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, "/"+query, strings.NewReader(body))
	for i, name := range []string{AssociationIdPathValue, BuildingIdPathValue, UnitIdPathValue, OwnershipIdPathValue} {
		if ids[i] != 0 {
			req.SetPathValue(name, fmt.Sprint(ids[i]))
		}
	}
	login := "ana"
	if admin {
		login = "root"
	}
	return authed(t, f.cfg, handler, login, admin, req)
}

func TestBuildingCRUD(t *testing.T) {
	f := newUnitsFixture(t)
	rec := f.call(HandleCreateBuilding(f.cfg), http.MethodPost, [4]int64{f.assocA}, `{"name":"  ","address":"x"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("nameless building: %d", rec.Code)
	}
	rec = f.call(HandleCreateBuilding(f.cfg), http.MethodPost, [4]int64{f.assocA}, `{"name":"Bloc C","address":"Str. C","cadastral_number":"CB-C","total_area":500}`)
	var b Building
	_ = json.Unmarshal(rec.Body.Bytes(), &b)
	if rec.Code != http.StatusCreated || b.ID == 0 || b.Name != "Bloc C" {
		t.Fatalf("create: %d %s", rec.Code, rec.Body.String())
	}
	ids := [4]int64{f.assocA, b.ID}
	if rec := f.call(HandleUpdateBuilding(f.cfg), http.MethodPut, ids, `{"name":"Bloc C2","address":"Str. C","total_area":510}`); rec.Code != http.StatusOK {
		t.Fatalf("update: %d", rec.Code)
	}
	if rec := f.call(HandleUpdateBuilding(f.cfg), http.MethodPut, [4]int64{f.assocA, f.buildingB}, `{"name":"x"}`); rec.Code != http.StatusNotFound {
		t.Fatalf("update another association's building: %d", rec.Code)
	}
	if rec := f.call(HandleDeleteBuilding(f.cfg), http.MethodDelete, [4]int64{f.assocA, f.buildingA}, ""); rec.Code != http.StatusConflict {
		t.Fatalf("delete building with units: %d", rec.Code)
	}
	if rec := f.call(HandleDeleteBuilding(f.cfg), http.MethodDelete, ids, ""); rec.Code != http.StatusNoContent {
		t.Fatalf("delete empty building: %d", rec.Code)
	}

	// The list reports each building's share total.
	rec = f.call(HandleGetAssociationBuildings(f.cfg), http.MethodGet, [4]int64{f.assocA}, "")
	var list []Building
	_ = json.Unmarshal(rec.Body.Bytes(), &list)
	if len(list) != 1 || list[0].ShareTotal < 0.0499 || list[0].ShareTotal > 0.0501 {
		t.Fatalf("buildings: %s", rec.Body.String())
	}
}

func TestCreateAndDeleteUnit(t *testing.T) {
	f := newUnitsFixture(t)
	create := func(body string) *httptest.ResponseRecorder {
		return f.call(HandleCreateUnit(f.cfg), http.MethodPost, [4]int64{f.assocA, f.buildingA}, body)
	}
	for body, want := range map[string]string{
		`{"cadastral_number":"CU-A2","unit_number":"2","unit_type":"villa","floor":1,"area":40,"part":0.04}`:    "Unit type must be apartment, commercial, office, parking or storage",
		`{"cadastral_number":"CU-A2","unit_number":"2","unit_type":"apartment","floor":1,"area":0,"part":0.04}`: "Area must be greater than 0",
		`{"cadastral_number":"CU-A2","unit_number":"2","unit_type":"apartment","floor":1,"area":40,"part":1.5}`: "Share must be greater than 0 and at most 100%",
		`{"cadastral_number":"","unit_number":"2","unit_type":"apartment","floor":1,"area":40,"part":0.04}`:     "Cadastral number is required",
	} {
		if rec := create(body); rec.Code != http.StatusBadRequest || errorMsg(rec) != want {
			t.Errorf("%s: %d %q", body, rec.Code, errorMsg(rec))
		}
	}
	if rec := create(`{"cadastral_number":"CU-B1","unit_number":"2","unit_type":"apartment","floor":1,"area":40,"part":0.04}`); rec.Code != http.StatusConflict {
		t.Fatalf("cadastral number of another association's unit: %d", rec.Code)
	}
	rec := create(`{"cadastral_number":"CU-A2","unit_number":"2","unit_type":"Parcare","floor":-1,"entrance":-1,"area":12.5,"part":0.01}`)
	var u Unit
	_ = json.Unmarshal(rec.Body.Bytes(), &u)
	if rec.Code != http.StatusCreated || u.UnitType != "parking" || u.Address != "Str. A, 2" || u.RoomCount != -1 {
		t.Fatalf("create: %d %s", rec.Code, rec.Body.String())
	}

	del := func(unitId int64) *httptest.ResponseRecorder {
		return f.call(HandleDeleteUnit(f.cfg), http.MethodDelete, [4]int64{f.assocA, f.buildingA, unitId}, "")
	}
	if rec := del(f.unitA); rec.Code != http.StatusConflict || !strings.Contains(errorMsg(rec), "ownership") {
		t.Fatalf("delete owned unit: %d %s", rec.Code, rec.Body.String())
	}
	// A unit listed by a gathering participant is kept too.
	if _, err := f.cfg.DB.Exec(`INSERT INTO gatherings (association_id, title, description, intent, gathering_date, gathering_type) VALUES (?, 'G', 'd', 'i', '2026-01-01', 'initial')`, f.assocA); err != nil {
		t.Fatal(err)
	}
	if _, err := f.cfg.DB.Exec(`INSERT INTO gathering_participants (gathering_id, participant_type, participant_name, units_info, units_area, units_part) VALUES (1, 'owner', 'X', ?, 1, 1)`, fmt.Sprintf("[%d]", u.ID)); err != nil {
		t.Fatal(err)
	}
	if rec := del(u.ID); rec.Code != http.StatusConflict || !strings.Contains(errorMsg(rec), "gathering") {
		t.Fatalf("delete unit used in a gathering: %d %s", rec.Code, rec.Body.String())
	}
	if _, err := f.cfg.DB.Exec(`DELETE FROM gathering_participants`); err != nil {
		t.Fatal(err)
	}
	if rec := del(u.ID); rec.Code != http.StatusNoContent {
		t.Fatalf("delete free unit: %d", rec.Code)
	}
}

func TestUpdateUnitStructuralFieldsAdminOnly(t *testing.T) {
	f := newUnitsFixture(t)
	body := `{"area":55.5,"part":0.06,"cadastral_number":"CU-A1-new"}`
	if rec := f.asRole(t, false, HandleUpdateBuildingUnit(f.cfg), http.MethodPut, f.unitIds(), "", body); rec.Code != http.StatusBadRequest {
		t.Fatalf("association admin: %d", rec.Code)
	}
	rec := f.asRole(t, true, HandleUpdateBuildingUnit(f.cfg), http.MethodPut, f.unitIds(), "", body)
	var u Unit
	_ = json.Unmarshal(rec.Body.Bytes(), &u)
	if rec.Code != http.StatusOK || u.Area != 55.5 || u.Part != 0.06 || u.CadastralNumber != "CU-A1-new" || u.UnitNumber != "1" {
		t.Fatalf("app admin: %d %s", rec.Code, rec.Body.String())
	}
	if rec := f.asRole(t, true, HandleUpdateBuildingUnit(f.cfg), http.MethodPut, f.unitIds(), "", `{"cadastral_number":"CU-B1"}`); rec.Code != http.StatusConflict {
		t.Fatalf("taken cadastral number: %d", rec.Code)
	}
}

type importReportT struct {
	OK       bool          `json:"ok"`
	Errors   []ImportIssue `json:"errors"`
	Warnings []ImportIssue `json:"warnings"`
	Summary  ImportSummary `json:"summary"`
}

func (f unitsFixture) runImport(t *testing.T, dryRun bool, rows string) importReportT {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/?dry_run=%v", dryRun), strings.NewReader(`{"rows":[`+rows+`]}`))
	req.SetPathValue(AssociationIdPathValue, fmt.Sprint(f.assocA))
	req.SetPathValue(BuildingIdPathValue, fmt.Sprint(f.buildingA))
	rec := httptest.NewRecorder()
	HandleImportUnits(f.cfg)(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("import: %d %s", rec.Code, rec.Body.String())
	}
	var report importReportT
	_ = json.Unmarshal(rec.Body.Bytes(), &report)
	return report
}

func (f unitsFixture) unitCount(t *testing.T) int64 {
	n, _ := f.cfg.Db.CountBuildingUnits(t.Context(), f.buildingA)
	return n
}

const importRows = `
{"line":2,"cadastral_number":"CU-A2","unit_number":"2","unit_type":"apartament","floor":1,"area":60,"share_percent":45,"owner_name":"Dan Pop","owner_id_number":"1900","registration_document":"D-2","ownership_start":"01.02.2024"},
{"line":3,"cadastral_number":"CU-A2","unit_number":"2","unit_type":"apartament","floor":1,"area":60,"share_percent":45,"owner_name":"Eva Pop","registration_document":"D-2b"},
{"line":4,"cadastral_number":"CU-A3","unit_number":"3","unit_type":"parking","floor":-1,"entrance":-1,"area":12,"share_percent":50,"owner_name":"alice","registration_document":"D-3","registration_date":"2024-02-03"}`

func TestImportUnits(t *testing.T) {
	f := newUnitsFixture(t)

	// Dry run: a full report, nothing written.
	report := f.runImport(t, true, importRows)
	if !report.OK || report.Summary.UnitsCreated != 2 || report.Summary.OwnersCreated != 2 || report.Summary.OwnersMatched != 1 ||
		report.Summary.OwnershipsCreated != 3 || f.unitCount(t) != 1 {
		t.Fatalf("dry run: %+v, units in db %d", report, f.unitCount(t))
	}
	if report.Summary.ShareTotalBefore < 0.0499 || report.Summary.ShareTotalAfter < 0.9999 || report.Summary.ShareTotalAfter > 1.0001 {
		t.Fatalf("share totals: %+v", report.Summary)
	}
	if len(report.Warnings) != 1 || report.Warnings[0].Code != "owner_matched_by_name" || report.Warnings[0].Line != 4 {
		t.Fatalf("warnings: %+v", report.Warnings)
	}

	// Commit.
	if report := f.runImport(t, false, importRows); !report.OK || f.unitCount(t) != 3 {
		t.Fatalf("commit: %+v", report)
	}
	unit, _ := f.cfg.Db.GetUnitByCadastralNumber(t.Context(), "CU-A2")
	ships, _ := f.cfg.Db.GetActiveUnitOwnerships(t.Context(), struct{ UnitID, AssociationID int64 }{unit.ID, f.assocA})
	if len(ships) != 2 {
		t.Fatalf("co-owners: %+v", ships)
	}
	for _, s := range ships {
		if s.IsVoting != (s.OwnerName == "Dan Pop") || !s.StartDate.Valid {
			t.Fatalf("first listed owner votes, start date stored: %+v", s)
		}
	}

	// Re-import with a corrected area: units updated in place, ownerships not duplicated.
	corrected := strings.Replace(importRows, `"area":12,`, `"area":13,`, 1)
	report = f.runImport(t, false, corrected)
	if !report.OK || report.Summary.UnitsUpdated != 2 || report.Summary.UnitsCreated != 0 ||
		report.Summary.OwnershipsCreated != 0 || report.Summary.OwnershipsSkipped != 3 || report.Summary.OwnersMatched != 3 {
		t.Fatalf("re-import: %+v", report.Summary)
	}
}

func TestImportUnits_ErrorsWriteNothing(t *testing.T) {
	f := newUnitsFixture(t)
	report := f.runImport(t, false, `
{"line":2,"cadastral_number":"CU-A2","unit_number":"2","unit_type":"apartment","floor":1,"area":60,"share_percent":45},
{"line":3,"cadastral_number":"CU-A2","unit_number":"2","unit_type":"apartment","floor":2,"area":60,"share_percent":45},
{"line":4,"cadastral_number":"CU-B1","unit_number":"9","unit_type":"apartment","floor":1,"area":60,"share_percent":5},
{"line":5,"cadastral_number":"CU-A5","unit_number":"5","unit_type":"apartment","area":60,"share_percent":5},
{"line":6,"cadastral_number":"CU-A6","unit_number":"6","unit_type":"apartment","floor":1,"area":60,"share_percent":5,"owner_name":"Zoe","registration_document":"D","ownership_start":"31/12/2024"},
{"line":7,"cadastral_number":"CU-A7","unit_number":"7","unit_type":"apartment","floor":1,"area":60,"share_percent":5,"owner_name":"Zoe"}`)
	byLine := map[int]string{}
	for _, e := range report.Errors {
		byLine[e.Line] = e.Code
	}
	want := map[int]string{3: "unit_mismatch", 4: "cadastral_elsewhere", 5: "required", 6: "invalid_date", 7: "required"}
	if report.OK || fmt.Sprint(byLine) != fmt.Sprint(want) {
		t.Fatalf("errors: %+v", report.Errors)
	}
	if f.unitCount(t) != 1 {
		t.Fatalf("a failed import wrote %d units", f.unitCount(t)-1)
	}
}
