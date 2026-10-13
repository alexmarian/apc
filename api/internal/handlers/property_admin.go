package handlers

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"

	"github.com/alexmarian/apc/api/internal/database"
	"github.com/alexmarian/apc/api/internal/logging"
	"go.uber.org/zap"
)

// App-admin-only management of buildings and units — the structural data
// that used to be loaded with api/sql/samples/load.sh. Routes in main.go use
// MiddlewareAdminOnly; the in-association scoping comes from scope.go.

func toBuilding(b database.Building) Building {
	return Building{
		ID:              b.ID,
		Name:            b.Name,
		Address:         b.Address,
		CadastralNumber: b.CadastralNumber,
		TotalArea:       b.TotalArea,
		CreatedAt:       b.CreatedAt.Time,
		UpdatedAt:       b.UpdatedAt.Time,
	}
}

type buildingInput struct {
	Name            string  `json:"name"`
	Address         string  `json:"address"`
	CadastralNumber string  `json:"cadastral_number"`
	TotalArea       float64 `json:"total_area"`
}

func parseBuildingInput(rw http.ResponseWriter, req *http.Request) (buildingInput, bool) {
	var in buildingInput
	if err := json.NewDecoder(req.Body).Decode(&in); err != nil {
		RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
		return in, false
	}
	in.Name, in.Address, in.CadastralNumber = strings.TrimSpace(in.Name), strings.TrimSpace(in.Address), strings.TrimSpace(in.CadastralNumber)
	switch {
	case in.Name == "":
		RespondWithError(rw, http.StatusBadRequest, "Building name is required")
		return in, false
	case len([]rune(in.Name)) > 255 || len([]rune(in.Address)) > 255 || len([]rune(in.CadastralNumber)) > 255:
		RespondWithError(rw, http.StatusBadRequest, "Text fields must be at most 255 characters")
		return in, false
	case in.TotalArea < 0:
		RespondWithError(rw, http.StatusBadRequest, "Total area can't be negative")
		return in, false
	}
	return in, true
}

func HandleCreateBuilding(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, ok := pathInt64(req, AssociationIdPathValue)
		if !ok {
			RespondWithError(rw, http.StatusBadRequest, "Invalid id")
			return
		}
		in, ok := parseBuildingInput(rw, req)
		if !ok {
			return
		}
		building, err := cfg.Db.CreateBuilding(req.Context(), database.CreateBuildingParams{
			Name: in.Name, Address: in.Address, CadastralNumber: in.CadastralNumber, TotalArea: in.TotalArea, AssociationID: associationId,
		})
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error creating building", zap.String("error", err.Error()))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to create building")
			return
		}
		RespondWithJSON(rw, http.StatusCreated, toBuilding(building))
	}
}

func HandleUpdateBuilding(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		building, ok := associationBuilding(cfg, rw, req)
		if !ok {
			return
		}
		in, ok := parseBuildingInput(rw, req)
		if !ok {
			return
		}
		if err := cfg.Db.UpdateBuilding(req.Context(), database.UpdateBuildingParams{
			Name: in.Name, Address: in.Address, CadastralNumber: in.CadastralNumber, TotalArea: in.TotalArea,
			ID: building.ID, AssociationID: building.AssociationID,
		}); err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Failed to update building")
			return
		}
		building.Name, building.Address, building.CadastralNumber, building.TotalArea = in.Name, in.Address, in.CadastralNumber, in.TotalArea
		result := toBuilding(building)
		result.ShareTotal, _ = cfg.Db.GetBuildingShareTotal(req.Context(), building.ID)
		RespondWithJSON(rw, http.StatusOK, result)
	}
}

// HandleDeleteBuilding deletes an empty building; 409 while it has units.
func HandleDeleteBuilding(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		building, ok := associationBuilding(cfg, rw, req)
		if !ok {
			return
		}
		count, err := cfg.Db.CountBuildingUnits(req.Context(), building.ID)
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Failed to delete building")
			return
		}
		if count > 0 {
			RespondWithError(rw, http.StatusConflict, "The building still has units; delete them first")
			return
		}
		if err := cfg.Db.DeleteBuilding(req.Context(), database.DeleteBuildingParams{ID: building.ID, AssociationID: building.AssociationID}); err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Failed to delete building")
			return
		}
		rw.WriteHeader(http.StatusNoContent)
	}
}

// unitTypes are the types new units may have (older data may hold others).
var unitTypes = []string{"apartment", "commercial", "office", "parking", "storage"}

// unitTypeAliases accepts the Romanian names import files tend to use.
var unitTypeAliases = map[string]string{
	"apartament": "apartment", "comercial": "commercial", "spatiu comercial": "commercial",
	"birou": "office", "oficiu": "office", "parcare": "parking", "boxa": "storage", "debara": "storage",
}

func normalizeUnitType(t string) string {
	t = strings.ToLower(strings.TrimSpace(t))
	if alias, ok := unitTypeAliases[t]; ok {
		return alias
	}
	return t
}

// unitFields are a new unit's data (create and import). Part is a fraction.
type unitFields struct {
	CadastralNumber string  `json:"cadastral_number"`
	UnitNumber      string  `json:"unit_number"`
	Address         string  `json:"address"`
	UnitType        string  `json:"unit_type"`
	Entrance        int64   `json:"entrance"`
	Floor           int64   `json:"floor"`
	RoomCount       int64   `json:"room_count"`
	Area            float64 `json:"area"`
	Part            float64 `json:"part"`
}

// fieldError is a validation failure tied to one field.
type fieldError struct{ Field, Code, Message string }

// normalize trims and canonicalizes, then validates; nil when valid.
func (u *unitFields) normalize(building database.Building) *fieldError {
	u.CadastralNumber, u.UnitNumber, u.Address = strings.TrimSpace(u.CadastralNumber), strings.TrimSpace(u.UnitNumber), strings.TrimSpace(u.Address)
	u.UnitType = normalizeUnitType(u.UnitType)
	if u.Address == "" {
		u.Address = strings.TrimSpace(building.Address + ", " + u.UnitNumber)
		u.Address = strings.TrimPrefix(u.Address, ", ")
	}
	switch {
	case u.CadastralNumber == "":
		return &fieldError{"cadastral_number", "required", "Cadastral number is required"}
	case u.UnitNumber == "":
		return &fieldError{"unit_number", "required", "Unit number is required"}
	case len([]rune(u.CadastralNumber)) > 255 || len([]rune(u.UnitNumber)) > 255 || len([]rune(u.Address)) > 255:
		return &fieldError{"", "too_long", "Text fields must be at most 255 characters"}
	case !slices.Contains(unitTypes, u.UnitType):
		return &fieldError{"unit_type", "invalid_type", "Unit type must be apartment, commercial, office, parking or storage"}
	case u.Entrance < unitEntranceMin:
		return &fieldError{"entrance", "out_of_range", "Entrance must be at least -2"}
	case u.RoomCount < -1:
		return &fieldError{"rooms", "out_of_range", "Room count must be at least 0"}
	case u.Area <= 0:
		return &fieldError{"area", "out_of_range", "Area must be greater than 0"}
	case u.Part <= 0 || u.Part > 1:
		return &fieldError{"share_percent", "out_of_range", "Share must be greater than 0 and at most 100%"}
	}
	return nil
}

// HandleCreateUnit adds one unit to a building. 409 when the cadastral
// number is already used (it's unique across the database).
func HandleCreateUnit(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		building, ok := associationBuilding(cfg, rw, req)
		if !ok {
			return
		}
		in := unitFields{Entrance: 1, RoomCount: -1}
		if err := json.NewDecoder(req.Body).Decode(&in); err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}
		if fe := in.normalize(building); fe != nil {
			RespondWithError(rw, http.StatusBadRequest, fe.Message)
			return
		}
		unit, err := cfg.Db.CreateUnit(req.Context(), database.CreateUnitParams{
			CadastralNumber: in.CadastralNumber, BuildingID: building.ID, UnitNumber: in.UnitNumber, Address: in.Address,
			Entrance: in.Entrance, Area: in.Area, Part: in.Part, UnitType: in.UnitType, Floor: in.Floor, RoomCount: in.RoomCount,
		})
		if isUniqueViolation(err) {
			RespondWithError(rw, http.StatusConflict, "Cadastral number already used by another unit")
			return
		}
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error creating unit", zap.String("error", err.Error()))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to create unit")
			return
		}
		RespondWithJSON(rw, http.StatusCreated, toUnit(unit))
	}
}

// unitInUse explains why a unit can't be deleted, or "" when it can.
// Ownership history and gathering participation are records worth keeping.
func unitInUse(cfg *ApiConfig, req *http.Request, unitID int64) (string, error) {
	ctx := req.Context()
	if n, err := cfg.Db.CountUnitOwnerships(ctx, unitID); err != nil || n > 0 {
		return "The unit has ownership records; it can't be deleted", err
	}
	if n, err := cfg.Db.CountUnitSlots(ctx, unitID); err != nil || n > 0 {
		return "The unit is part of a gathering; it can't be deleted", err
	}
	infos, err := cfg.Db.ListParticipantUnitsInfo(ctx)
	if err != nil {
		return "", err
	}
	for _, info := range infos {
		var ids []int64
		if json.Unmarshal([]byte(info), &ids) == nil && slices.Contains(ids, unitID) {
			return "The unit is part of a gathering; it can't be deleted", nil
		}
	}
	return "", nil
}

// HandleDeleteUnit deletes a unit that was never owned nor used in a
// gathering (typically a mistaken entry); 409 with the reason otherwise.
func HandleDeleteUnit(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		building, unit, ok := associationUnit(cfg, rw, req)
		if !ok {
			return
		}
		reason, err := unitInUse(cfg, req, unit.ID)
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Failed to delete unit")
			return
		}
		if reason != "" {
			RespondWithError(rw, http.StatusConflict, reason)
			return
		}
		if err := cfg.Db.DeleteUnit(req.Context(), database.DeleteUnitParams{ID: unit.ID, BuildingID: building.ID}); err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Failed to delete unit")
			return
		}
		rw.WriteHeader(http.StatusNoContent)
	}
}
