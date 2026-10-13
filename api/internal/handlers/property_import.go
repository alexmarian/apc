package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/alexmarian/apc/api/internal/database"
	"github.com/alexmarian/apc/api/internal/logging"
	"go.uber.org/zap"
)

// Bulk import of a building's units (and optionally their owners), replacing
// the manual api/sql/samples/load.sh. The client parses the CSV and sends
// normalized rows; this re-validates everything. Each call runs inside a
// transaction: a dry run (preview) rolls it back, a commit keeps it only
// when there were no errors — all or nothing. See
// ui-admin/implementation/18-feature-property-import.md.

const importMaxRows = 2000

// shareTolerance is how far a building's share total may drift from 100%
// before the import warns.
const shareTolerance = 0.0001

// ImportRow is one CSV line, normalized by the client. Pointers are
// optional/“empty cell” values.
type ImportRow struct {
	Line            int      `json:"line"`
	CadastralNumber string   `json:"cadastral_number"`
	UnitNumber      string   `json:"unit_number"`
	UnitType        string   `json:"unit_type"`
	Floor           *int64   `json:"floor"`
	Entrance        *int64   `json:"entrance"`
	Area            *float64 `json:"area"`
	SharePercent    *float64 `json:"share_percent"`
	Rooms           *int64   `json:"rooms"`
	Address         string   `json:"address"`

	OwnerName            string `json:"owner_name"`
	OwnerIDNumber        string `json:"owner_id_number"`
	OwnerPhone           string `json:"owner_phone"`
	OwnerEmail           string `json:"owner_email"`
	OwnershipStart       string `json:"ownership_start"`
	RegistrationDocument string `json:"registration_document"`
	RegistrationDate     string `json:"registration_date"`
}

type ImportIssue struct {
	Line    int    `json:"line"`
	Field   string `json:"field,omitempty"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ImportSummary struct {
	Rows              int     `json:"rows"`
	UnitsCreated      int     `json:"units_created"`
	UnitsUpdated      int     `json:"units_updated"`
	OwnersCreated     int     `json:"owners_created"`
	OwnersMatched     int     `json:"owners_matched"`
	OwnershipsCreated int     `json:"ownerships_created"`
	OwnershipsSkipped int     `json:"ownerships_skipped"`
	ShareTotalBefore  float64 `json:"share_total_before"`
	ShareTotalAfter   float64 `json:"share_total_after"`
}

type ImportReport struct {
	OK       bool          `json:"ok"`
	DryRun   bool          `json:"dry_run"`
	Errors   []ImportIssue `json:"errors"`
	Warnings []ImportIssue `json:"warnings"`
	Summary  ImportSummary `json:"summary"`
}

// parseImportDate accepts YYYY-MM-DD and DD.MM.YYYY; empty → fallback.
func parseImportDate(s string, fallback time.Time) (time.Time, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return fallback, true
	}
	for _, layout := range []string{"2006-01-02", "02.01.2006", "2.1.2006"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// importer carries one import's state inside its transaction.
type importer struct {
	ctx      context.Context
	q        *database.Queries
	building database.Building
	report   *ImportReport
	// New owners created by this import, so later rows reuse them.
	newOwnersByID   map[string]int64
	newOwnersByName map[string]int64
}

func (im *importer) fail(line int, field, code, message string) {
	im.report.Errors = append(im.report.Errors, ImportIssue{Line: line, Field: field, Code: code, Message: message})
}

func (im *importer) warn(line int, code, message string) {
	im.report.Warnings = append(im.report.Warnings, ImportIssue{Line: line, Code: code, Message: message})
}

// unitFieldsOf builds a row's unit, reporting missing required numbers.
func (im *importer) unitFieldsOf(r ImportRow) (unitFields, bool) {
	u := unitFields{
		CadastralNumber: r.CadastralNumber, UnitNumber: r.UnitNumber, Address: r.Address, UnitType: r.UnitType,
		Entrance: 1, RoomCount: -1,
	}
	for _, req := range []struct {
		field string
		ok    bool
	}{{"floor", r.Floor != nil}, {"area", r.Area != nil}, {"share_percent", r.SharePercent != nil}} {
		if !req.ok {
			im.fail(r.Line, req.field, "required", fmt.Sprintf("%s is required", req.field))
			return u, false
		}
	}
	u.Floor, u.Area, u.Part = *r.Floor, *r.Area, *r.SharePercent/100
	if r.Entrance != nil {
		u.Entrance = *r.Entrance
	}
	if r.Rooms != nil {
		if *r.Rooms < 0 {
			im.fail(r.Line, "rooms", "out_of_range", "Room count must be at least 0")
			return u, false
		}
		u.RoomCount = *r.Rooms
	}
	if fe := u.normalize(im.building); fe != nil {
		im.fail(r.Line, fe.Field, fe.Code, fe.Message)
		return u, false
	}
	return u, true
}

// importUnit creates or updates one unit; returns its id and whether it
// already existed.
func (im *importer) importUnit(line int, u unitFields) (int64, bool, bool) {
	existing, err := im.q.GetUnitByCadastralNumber(im.ctx, u.CadastralNumber)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		created, err := im.q.CreateUnit(im.ctx, database.CreateUnitParams{
			CadastralNumber: u.CadastralNumber, BuildingID: im.building.ID, UnitNumber: u.UnitNumber, Address: u.Address,
			Entrance: u.Entrance, Area: u.Area, Part: u.Part, UnitType: u.UnitType, Floor: u.Floor, RoomCount: u.RoomCount,
		})
		if err != nil {
			im.fail(line, "", "database", "Could not create the unit: "+err.Error())
			return 0, false, false
		}
		im.report.Summary.UnitsCreated++
		return created.ID, false, true
	case err != nil:
		im.fail(line, "", "database", err.Error())
		return 0, false, false
	case existing.BuildingID != im.building.ID:
		im.fail(line, "cadastral_number", "cadastral_elsewhere", "This cadastral number belongs to a unit of another building")
		return 0, false, false
	}
	if err := im.q.UpdateUnitAll(im.ctx, database.UpdateUnitAllParams{
		CadastralNumber: u.CadastralNumber, UnitNumber: u.UnitNumber, Address: u.Address, Entrance: u.Entrance,
		Area: u.Area, Part: u.Part, UnitType: u.UnitType, Floor: u.Floor, RoomCount: u.RoomCount,
		ID: existing.ID, BuildingID: im.building.ID,
	}); err != nil {
		im.fail(line, "", "database", "Could not update the unit: "+err.Error())
		return 0, false, false
	}
	im.report.Summary.UnitsUpdated++
	return existing.ID, true, true
}

// ownerFor matches (by ID number, then by name — with a warning) or
// creates the row's owner.
func (im *importer) ownerFor(r ImportRow) (int64, bool) {
	name := strings.TrimSpace(r.OwnerName)
	idNumber := strings.TrimSpace(r.OwnerIDNumber)
	email := strings.TrimSpace(r.OwnerEmail)
	if email != "" && !emailPattern.MatchString(email) {
		im.fail(r.Line, "owner_email", "invalid_email", "Invalid email address")
		return 0, false
	}
	normalized := strings.ToLower(name)
	if idNumber != "" {
		if id, ok := im.newOwnersByID[idNumber]; ok {
			return id, true
		}
		found, err := im.q.GetAssociationOwnersByIdentification(im.ctx, database.GetAssociationOwnersByIdentificationParams{
			AssociationID: im.building.AssociationID, IdentificationNumber: idNumber,
		})
		if err != nil {
			im.fail(r.Line, "", "database", err.Error())
			return 0, false
		}
		if len(found) > 0 {
			im.report.Summary.OwnersMatched++
			return found[0].ID, true
		}
	} else {
		if id, ok := im.newOwnersByName[normalized]; ok {
			return id, true
		}
		found, err := im.q.GetAssociationOwnersByNormalizedName(im.ctx, database.GetAssociationOwnersByNormalizedNameParams{
			AssociationID: im.building.AssociationID, NormalizedName: normalized,
		})
		if err != nil {
			im.fail(r.Line, "", "database", err.Error())
			return 0, false
		}
		if len(found) > 0 {
			im.report.Summary.OwnersMatched++
			im.warn(r.Line, "owner_matched_by_name", fmt.Sprintf("Owner %q matched an existing owner by name only", name))
			return found[0].ID, true
		}
	}
	created, err := im.q.CreateOwner(im.ctx, database.CreateOwnerParams{
		Name: name, NormalizedName: normalized, IdentificationNumber: idNumber,
		ContactPhone: strings.TrimSpace(r.OwnerPhone), ContactEmail: email, AssociationID: im.building.AssociationID,
	})
	if err != nil {
		im.fail(r.Line, "", "database", "Could not create the owner: "+err.Error())
		return 0, false
	}
	im.report.Summary.OwnersCreated++
	if idNumber != "" {
		im.newOwnersByID[idNumber] = created.ID
	} else {
		im.newOwnersByName[normalized] = created.ID
	}
	return created.ID, true
}

// importOwnerships records the unit's owners from its rows. The first
// listed owner votes, unless the unit already has a voting owner. Existing
// active ownerships are left alone (re-imports don't duplicate them).
func (im *importer) importOwnerships(unitID int64, rows []ImportRow, today time.Time) {
	active, err := im.q.GetActiveUnitOwnerships(im.ctx, database.GetActiveUnitOwnershipsParams{UnitID: unitID, AssociationID: im.building.AssociationID})
	if err != nil {
		im.fail(rows[0].Line, "", "database", err.Error())
		return
	}
	hasVoting := false
	activeOwners := map[int64]bool{}
	for _, o := range active {
		activeOwners[o.OwnerID] = true
		hasVoting = hasVoting || o.IsVoting
	}
	seen := map[int64]int{}
	for _, r := range rows {
		if strings.TrimSpace(r.OwnerName) == "" {
			if strings.TrimSpace(r.OwnerIDNumber+r.OwnerPhone+r.OwnerEmail+r.RegistrationDocument) != "" {
				im.fail(r.Line, "owner_name", "required", "Owner name is required when owner details are given")
			}
			continue
		}
		doc := strings.TrimSpace(r.RegistrationDocument)
		if doc == "" {
			im.fail(r.Line, "registration_document", "required", "Registration document is required for an owner")
			continue
		}
		start, ok := parseImportDate(r.OwnershipStart, today)
		if !ok {
			im.fail(r.Line, "ownership_start", "invalid_date", "Dates must be YYYY-MM-DD or DD.MM.YYYY")
			continue
		}
		registered, ok := parseImportDate(r.RegistrationDate, start)
		if !ok {
			im.fail(r.Line, "registration_date", "invalid_date", "Dates must be YYYY-MM-DD or DD.MM.YYYY")
			continue
		}
		ownerID, ok := im.ownerFor(r)
		if !ok {
			continue
		}
		if line, dup := seen[ownerID]; dup {
			im.fail(r.Line, "owner_name", "duplicate_owner", fmt.Sprintf("The same owner is already listed for this unit on line %d", line))
			continue
		}
		seen[ownerID] = r.Line
		if activeOwners[ownerID] {
			im.report.Summary.OwnershipsSkipped++
			continue
		}
		voting := !hasVoting
		if _, err := im.q.CreateOwnership(im.ctx, database.CreateOwnershipParams{
			UnitID: unitID, OwnerID: ownerID, AssociationID: im.building.AssociationID,
			StartDate: sql.NullTime{Time: start, Valid: true}, IsActive: true, IsVoting: voting,
			RegistrationDocument: doc, RegistrationDate: registered,
		}); err != nil {
			im.fail(r.Line, "", "database", "Could not record the ownership: "+err.Error())
			continue
		}
		hasVoting = hasVoting || voting
		activeOwners[ownerID] = true
		im.report.Summary.OwnershipsCreated++
	}
}

// HandleImportUnits (POST .../buildings/{bid}/units/import?dry_run=true|false)
// validates and applies an import. Always 200 with an ImportReport; `ok`
// false means errors and nothing was written.
func HandleImportUnits(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		building, ok := associationBuilding(cfg, rw, req)
		if !ok {
			return
		}
		dryRun := req.URL.Query().Get("dry_run") != "false"
		var body struct {
			Rows []ImportRow `json:"rows"`
		}
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}
		if len(body.Rows) == 0 {
			RespondWithError(rw, http.StatusBadRequest, "The file has no rows")
			return
		}
		if len(body.Rows) > importMaxRows {
			RespondWithError(rw, http.StatusBadRequest, fmt.Sprintf("At most %d rows per import", importMaxRows))
			return
		}

		ctx := req.Context()
		tx, err := cfg.DB.BeginTx(ctx, nil)
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Import failed")
			return
		}
		defer tx.Rollback()

		report := &ImportReport{DryRun: dryRun, Errors: []ImportIssue{}, Warnings: []ImportIssue{}}
		report.Summary.Rows = len(body.Rows)
		im := &importer{
			ctx: ctx, q: cfg.Db.WithTx(tx), building: building, report: report,
			newOwnersByID: map[string]int64{}, newOwnersByName: map[string]int64{},
		}
		report.Summary.ShareTotalBefore, _ = im.q.GetBuildingShareTotal(ctx, building.ID)

		// Group rows by cadastral number, in file order: repeated rows are
		// the unit's co-owners and must agree on the unit's details.
		type group struct {
			unit unitFields
			rows []ImportRow
		}
		var order []string
		groups := map[string]*group{}
		for _, r := range body.Rows {
			u, ok := im.unitFieldsOf(r)
			if !ok {
				continue
			}
			if g, seen := groups[u.CadastralNumber]; seen {
				if g.unit != u {
					im.fail(r.Line, "cadastral_number", "unit_mismatch",
						fmt.Sprintf("Unit details differ from line %d for the same cadastral number", g.rows[0].Line))
					continue
				}
				g.rows = append(g.rows, r)
				continue
			}
			groups[u.CadastralNumber] = &group{unit: u, rows: []ImportRow{r}}
			order = append(order, u.CadastralNumber)
		}

		today := time.Now().UTC().Truncate(24 * time.Hour)
		unitLines := map[string]int{}
		for _, cadastral := range order {
			g := groups[cadastral]
			unitID, _, ok := im.importUnit(g.rows[0].Line, g.unit)
			if !ok {
				continue
			}
			if line, dup := unitLines[g.unit.UnitNumber]; dup {
				im.warn(g.rows[0].Line, "duplicate_unit_number", fmt.Sprintf("Unit number %q is also used on line %d", g.unit.UnitNumber, line))
			}
			unitLines[g.unit.UnitNumber] = g.rows[0].Line
			im.importOwnerships(unitID, g.rows, today)
		}

		report.Summary.ShareTotalAfter, _ = im.q.GetBuildingShareTotal(ctx, building.ID)
		if math.Abs(report.Summary.ShareTotalAfter-1) > shareTolerance {
			im.warn(0, "share_total", fmt.Sprintf("After the import the building's shares total %.4f%%, not 100%%", report.Summary.ShareTotalAfter*100))
		}

		report.OK = len(report.Errors) == 0
		if report.OK && !dryRun {
			if err := tx.Commit(); err != nil {
				logging.Logger.Log(zap.WarnLevel, "Error committing import", zap.String("error", err.Error()))
				RespondWithError(rw, http.StatusInternalServerError, "Import failed")
				return
			}
		}
		RespondWithJSON(rw, http.StatusOK, report)
	}
}
