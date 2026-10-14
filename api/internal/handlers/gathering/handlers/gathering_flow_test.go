package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/alexmarian/apc/api/internal/auth"
	"github.com/alexmarian/apc/api/internal/database"
	"github.com/alexmarian/apc/api/internal/handlers"
	"github.com/alexmarian/apc/api/internal/handlers/gathering/domain"
	"github.com/alexmarian/apc/api/internal/logging"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	logging.Logger = zap.NewNop()
	os.Exit(m.Run())
}

// flow is association A (Alice owns apartments 1 and 2, Bob commercial 3)
// and association B (one gathering), served through the real middleware.
type flow struct {
	t                   *testing.T
	cfg                 *handlers.ApiConfig
	mux                 *http.ServeMux
	token               string
	assocA, assocB      int64
	alice, bob, carol   int64
	unit1, unit2, unit3 int64
	foreignGathering    int64
}

func newFlow(t *testing.T) *flow {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })
	if err := database.RunMigrations(db, os.DirFS("../../../../sql/schema")); err != nil {
		t.Fatal(err)
	}
	f := &flow{t: t, cfg: &handlers.ApiConfig{Db: database.New(db), DB: db, Secret: "test-secret"}}
	exec := func(q string, args ...any) int64 {
		t.Helper()
		res, err := db.Exec(q, args...)
		if err != nil {
			t.Fatal(err)
		}
		id, _ := res.LastInsertId()
		return id
	}
	exec(`INSERT INTO users (login, password_hash, topt_secret, is_admin) VALUES ('root', 'x', 'x', true)`)
	f.assocA = exec(`INSERT INTO associations (name, address, administrator) VALUES ('A', 'a', '')`)
	f.assocB = exec(`INSERT INTO associations (name, address, administrator) VALUES ('B', 'b', '')`)
	building := exec(`INSERT INTO buildings (name, address, cadastral_number, total_area, association_id) VALUES ('Bloc A', 'x', 'CB', 1000, ?)`, f.assocA)
	unit := func(n, typ string, area, part float64) int64 {
		return exec(`INSERT INTO units (cadastral_number, building_id, unit_number, address, entrance, area, part, unit_type, floor) VALUES (?, ?, ?, 'x', 1, ?, ?, ?, 1)`, "CU-"+n, building, n, area, part, typ)
	}
	f.unit1, f.unit2, f.unit3 = unit("1", "apartment", 60, 0.3), unit("2", "apartment", 40, 0.2), unit("3", "commercial", 100, 0.5)
	owner := func(name string, assoc int64) int64 {
		return exec(`INSERT INTO owners (name, normalized_name, identification_number, association_id) VALUES (?, ?, ?, ?)`, name, strings.ToLower(name), "ID-"+name, assoc)
	}
	f.alice, f.bob, f.carol = owner("Alice", f.assocA), owner("Bob", f.assocA), owner("Carol", f.assocB)
	own := func(unit, owner int64) {
		exec(`INSERT INTO ownerships (unit_id, owner_id, association_id, is_active, is_voting, registration_document, registration_date) VALUES (?, ?, ?, true, true, 'D', '2020-01-01 00:00:00')`, unit, owner, f.assocA)
	}
	own(f.unit1, f.alice)
	own(f.unit2, f.alice)
	own(f.unit3, f.bob)
	f.foreignGathering = exec(`INSERT INTO gatherings (association_id, title, description, intent, location, gathering_date, gathering_type, voting_mode, status) VALUES (?, 'B meeting', '', '', 'x', '2026-01-01 00:00:00', 'initial', 'by_weight', 'draft')`, f.assocB)

	router := NewGatheringHandler(f.cfg)
	matters := NewVotingMatterHandler(f.cfg, router)
	participants := NewParticipantHandler(f.cfg, router)
	ballots := NewBallotHandler(f.cfg, router)
	results := NewResultsHandler(f.cfg)
	invitations := NewInvitationHandler(f.cfg)
	f.mux = http.NewServeMux()
	base := "/v1/api/associations/{associationId}/gatherings"
	g := base + "/{gatheringId}"
	for pattern, h := range map[string]http.HandlerFunc{
		"POST " + base:               router.HandleCreateGathering(),
		"PUT " + g:                   router.HandleUpdateGathering(),
		"PUT " + g + "/status":       router.HandleUpdateGatheringStatus(),
		"GET " + g + "/matters":      matters.HandleGetVotingMatters(),
		"POST " + g + "/matters":     matters.HandleCreateVotingMatter(),
		"GET " + g + "/participants": participants.HandleGetParticipants(),
		"POST " + g + "/participants/{participantId}/checkin": participants.HandleCheckInParticipant(),
		"POST " + g + "/ballot":                               ballots.HandleSubmitBallot(),
		"GET " + g + "/ballots":                               ballots.HandleGetBallots(),
		"GET " + g + "/results":                               results.HandleGetVoteResults(),
		"GET " + g + "/eligible-voters":                       results.HandleGetEligibleVoters(),
		"POST " + g + "/invitations":                          invitations.HandleCreateInvitation(),
		"DELETE " + g + "/invitations/{invitationId}":         invitations.HandleRevokeInvitation(),
	} {
		f.mux.HandleFunc(pattern, f.cfg.MiddlewareAssociationResource(h))
	}
	f.mux.HandleFunc("POST /v1/api/ballot/verify", ballots.HandleVerifyBallot())
	f.token, _, _ = auth.MakeJWT("root", f.cfg.Secret, time.Hour, true)
	return f
}

func (f *flow) do(method, path, body string, out any) int {
	f.t.Helper()
	req := httptest.NewRequest(method, "/v1/api"+path, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+f.token)
	rec := httptest.NewRecorder()
	f.mux.ServeHTTP(rec, req)
	if out != nil {
		if err := json.Unmarshal(rec.Body.Bytes(), out); err != nil {
			f.t.Fatalf("%s %s: %d %s", method, path, rec.Code, rec.Body.String())
		}
	}
	return rec.Code
}

func (f *flow) g(id int64, rest string) string {
	return fmt.Sprintf("/associations/%d/gatherings/%d%s", f.assocA, id, rest)
}

func TestGatheringFlow(t *testing.T) {
	f := newFlow(t)

	// Create: apartments only, then edit to every unit type.
	var created domain.Gathering
	body := `{"title":" Annual ","description":"d","intent":"d","location":"Hall","gathering_date":"2026-11-01T18:00:00Z","gathering_type":"initial","voting_mode":"by_weight","qualification_unit_types":["apartment"]}`
	if code := f.do("POST", fmt.Sprintf("/associations/%d/gatherings", f.assocA), body, &created); code != 201 || created.Title != "Annual" || created.QualifiedUnitsCount != 2 {
		t.Fatalf("create: %d %+v", code, created)
	}
	if code := f.do("POST", fmt.Sprintf("/associations/%d/gatherings", f.assocA), strings.Replace(body, "initial", "secret", 1), nil); code != 400 {
		t.Fatalf("bad type: %d", code)
	}
	var edited domain.Gathering
	if code := f.do("PUT", f.g(created.ID, ""), strings.Replace(strings.Replace(body, `"qualification_unit_types":["apartment"]`, `"qualification_unit_types":[]`, 1), "Hall", "Lobby", 1), &edited); code != 200 || edited.Location != "Lobby" || edited.QualifiedUnitsCount != 3 || edited.QualificationUnitTypes != nil {
		t.Fatalf("edit: %d %+v", code, edited)
	}
	var voters []struct {
		Owner struct{ ID int64 } `json:"owner"`
	}
	if f.do("GET", f.g(created.ID, "/eligible-voters"), "", &voters); len(voters) != 2 {
		t.Fatalf("slots not re-synced: %+v", voters)
	}

	// Another association's gathering is out of reach through A.
	if code := f.do("GET", f.g(f.foreignGathering, "/matters"), "", nil); code != 404 {
		t.Fatalf("foreign gathering: %d", code)
	}

	// Matters: yes/no with abstention, a single choice, an informative item.
	var yesNo, choice domain.VotingMatter
	f.do("POST", f.g(created.ID, "/matters"), `{"order_index":1,"title":"Budget","matter_type":"budget","voting_config":{"type":"yes_no","required_majority":"simple","allow_abstention":true}}`, &yesNo)
	f.do("POST", f.g(created.ID, "/matters"), `{"order_index":2,"title":"Color","matter_type":"poll","voting_config":{"type":"single_choice","required_majority":"simple","options":[{"id":"1","text":"Red"},{"id":"2","text":"Blue"}]}}`, &choice)
	f.do("POST", f.g(created.ID, "/matters"), `{"order_index":3,"title":"Report","matter_type":"policy","is_informative":true,"voting_config":{"type":"yes_no","required_majority":"simple"}}`, nil)

	for _, s := range []string{"published", "active"} {
		if code := f.do("PUT", f.g(created.ID, "/status"), `{"status":"`+s+`"}`, nil); code != 200 {
			t.Fatalf("status %s: %d", s, code)
		}
	}
	if code := f.do("PUT", f.g(created.ID, ""), body, nil); code != 409 {
		t.Fatalf("edit while active: %d", code)
	}

	ballot := func(voter string, units []int64, yn, color string) (int, map[string]any) {
		content := fmt.Sprintf(`{"%d":{"matter_id":%d,"values":["%s"]},"%d":{"matter_id":%d,"values":["%s"]}}`, yesNo.ID, yesNo.ID, yn, choice.ID, choice.ID, color)
		unitsJSON, _ := json.Marshal(units)
		var out map[string]any
		code := f.do("POST", f.g(created.ID, "/ballot"), fmt.Sprintf(`{%s,"unit_ids":%s,"ballot_content":%s}`, voter, unitsJSON, content), &out)
		return code, out
	}
	owner := func(id int64) string { return fmt.Sprintf(`"voter_type":"owner","owner_id":%d`, id) }

	for name, tc := range map[string]struct {
		yn, color, msg string
	}{
		"missing matter": {"yes", "", "Missing vote for \"Color\""},
		"unknown option": {"yes", "9", "Invalid vote for \"Color\""},
		"no abstention":  {"yes", "abstain", "Abstention is not allowed for \"Color\""},
		"bad yes/no":     {"maybe", "1", "Invalid vote for \"Budget\""},
	} {
		content := map[string]domain.BallotVote{fmt.Sprint(yesNo.ID): {MatterID: yesNo.ID, Values: []string{tc.yn}}}
		if tc.color != "" {
			content[fmt.Sprint(choice.ID)] = domain.BallotVote{MatterID: choice.ID, Values: []string{tc.color}}
		}
		c, _ := json.Marshal(content)
		var out map[string]any
		if code := f.do("POST", f.g(created.ID, "/ballot"), fmt.Sprintf(`{%s,"unit_ids":[%d],"ballot_content":%s}`, owner(f.alice), f.unit1, c), &out); code != 400 || out["msg"] != tc.msg {
			t.Errorf("%s: %d %v", name, code, out)
		}
	}

	code, aliceBallot := ballot(owner(f.alice), []int64{f.unit1}, "yes", "1")
	if code != 201 {
		t.Fatalf("alice: %d %v", code, aliceBallot)
	}
	if code, _ := ballot(owner(f.alice), []int64{f.unit1}, "no", "2"); code != 400 {
		t.Fatalf("unit voted twice: %d", code)
	}
	delegate := fmt.Sprintf(`"voter_type":"delegate","delegating_owner_id":%d,"delegate_name":"Dan Delegate","delegate_identification":"ID-DAN","delegation_document_ref":"PROXY-1"`, f.alice)
	if code, out := ballot(delegate, []int64{f.unit2}, "abstain", "2"); code != 201 {
		t.Fatalf("delegate: %d %v", code, out)
	}
	if code, _ := ballot(fmt.Sprintf(`"voter_type":"owner","owner_id":%d`, f.carol), []int64{f.unit3}, "yes", "1"); code != 400 {
		t.Fatalf("foreign owner: %d", code)
	}

	var participants []domain.GatheringParticipant
	f.do("GET", f.g(created.ID, "/participants"), "", &participants)
	if len(participants) != 2 || participants[1].ParticipantName != "Dan Delegate" || participants[1].DelegatingOwnerName != "Alice" || participants[1].ParticipantIdentification != "ID-DAN" || !participants[1].HasVoted {
		t.Fatalf("participants: %+v", participants)
	}
	if code := f.do("POST", f.g(created.ID, fmt.Sprintf("/participants/%d/checkin", participants[0].ID)), "", nil); code != 200 {
		t.Fatalf("check in: %d", code)
	}
	if code := f.do("POST", f.g(created.ID, "/participants/999/checkin"), "", nil); code != 404 {
		t.Fatalf("check in unknown: %d", code)
	}

	var ballotsList []map[string]any
	f.do("GET", f.g(created.ID, "/ballots"), "", &ballotsList)
	if len(ballotsList) != 2 || ballotsList[1]["participant_type"] != "delegate" || ballotsList[1]["delegating_owner_name"] != "Alice" {
		t.Fatalf("ballots: %v", ballotsList)
	}

	// Results reflect both ballots right away (tallies before the cache drop).
	var results domain.VoteResults
	f.do("GET", f.g(created.ID, "/results"), "", &results)
	budget := results.Results[0]
	if len(results.Results) != 3 || budget.Statistics.Abstentions != 1 || budget.Votes[0].Choice != "yes" || budget.Votes[len(budget.Votes)-1].Choice != "abstain" {
		t.Fatalf("results: %+v", results)
	}

	// Verify by ballot id + hash.
	var verified map[string]any
	f.do("POST", "/ballot/verify", fmt.Sprintf(`{"ballot_id":%v,"ballot_hash":"%v"}`, aliceBallot["ballot_id"], aliceBallot["ballot_hash"]), &verified)
	if verified["valid"] != true {
		t.Fatalf("verify: %v", verified)
	}
	f.do("POST", "/ballot/verify", fmt.Sprintf(`{"ballot_id":%v,"ballot_hash":"nope"}`, aliceBallot["ballot_id"]), &verified)
	if verified["valid"] != false {
		t.Fatalf("verify wrong hash: %v", verified)
	}

	// Closing caches the final results; reopening drops them.
	f.do("PUT", f.g(created.ID, "/status"), `{"status":"closed"}`, nil)
	cached := func() int {
		var n int
		f.cfg.DB.QueryRow(`SELECT COUNT(*) FROM voting_results WHERE gathering_id = ?`, created.ID).Scan(&n)
		return n
	}
	if cached() != 1 {
		t.Fatal("results not cached on close")
	}
	f.do("PUT", f.g(created.ID, "/status"), `{"status":"active"}`, nil)
	if cached() != 0 {
		t.Fatal("results still cached after reopening")
	}

	// Invitations: only eligible owners; revoke only this gathering's.
	if code := f.do("POST", f.g(created.ID, "/invitations"), fmt.Sprintf(`{"owner_id":%d}`, f.carol), nil); code != 400 {
		t.Fatalf("invite foreign owner: %d", code)
	}
	var inv map[string]any
	if code := f.do("POST", f.g(created.ID, "/invitations"), fmt.Sprintf(`{"owner_id":%d}`, f.bob), &inv); code != 201 {
		t.Fatalf("invite: %d %v", code, inv)
	}
	other := f.do("POST", fmt.Sprintf("/associations/%d/gatherings", f.assocA), body, &created)
	_ = other
	if code := f.do("DELETE", f.g(created.ID, fmt.Sprintf("/invitations/%v", inv["id"])), "", nil); code != 404 {
		t.Fatalf("revoke through another gathering: %d", code)
	}
}

// A unit whose slot is taken meanwhile: nothing of the attempt is kept.
func TestRecordBallot_AllOrNothing(t *testing.T) {
	f := newFlow(t)
	var g domain.Gathering
	f.do("POST", fmt.Sprintf("/associations/%d/gatherings", f.assocA), `{"title":"T","location":"x","gathering_type":"initial","gathering_date":"2026-11-01T18:00:00Z"}`, &g)
	if _, err := f.cfg.DB.Exec(`UPDATE gatherings SET status = 'active' WHERE id = ?`, g.ID); err != nil {
		t.Fatal(err)
	}
	rows, err := f.cfg.Db.GetEligibleVotersWithUnits(context.Background(), database.GetEligibleVotersWithUnitsParams{GatheringID: g.ID, AssociationID: f.assocA})
	if err != nil {
		t.Fatal(err)
	}
	var aliceUnits []database.GetEligibleVotersWithUnitsRow
	for _, r := range rows {
		if r.OwnerID == f.alice {
			aliceUnits = append(aliceUnits, r)
		}
	}
	// Unit 2 is claimed by someone else between the check and the insert.
	f.cfg.DB.Exec(`INSERT INTO gathering_participants (gathering_id, participant_type, participant_name, units_info, units_area, units_part) VALUES (?, 'owner', 'Other', '[]', 0, 0)`, g.ID)
	f.cfg.DB.Exec(`UPDATE unit_slots SET participant_id = (SELECT MAX(id) FROM gathering_participants) WHERE gathering_id = ? AND unit_id = ?`, g.ID, f.unit2)

	_, _, err = recordBallot(context.Background(), f.cfg, ballotRecord{
		GatheringID: g.ID, ParticipantType: "owner", ParticipantName: "Alice",
		OwnerID: sql.NullInt64{Int64: f.alice, Valid: true}, Units: aliceUnits, Content: map[string]domain.BallotVote{},
	})
	if !errors.Is(err, errUnitTaken) {
		t.Fatalf("err = %v", err)
	}
	var participants, claimed int
	f.cfg.DB.QueryRow(`SELECT COUNT(*) FROM gathering_participants WHERE participant_name = 'Alice'`).Scan(&participants)
	f.cfg.DB.QueryRow(`SELECT COUNT(*) FROM unit_slots WHERE unit_id = ? AND participant_id IS NOT NULL`, f.unit1).Scan(&claimed)
	if participants != 0 || claimed != 0 {
		t.Fatalf("partial ballot kept: participants=%d unit1 claimed=%d", participants, claimed)
	}
}

// Buildings: a gathering applies to the selected ones (all when none); with
// several it is one common decision over all their qualifying units.
func TestGatheringBuildings(t *testing.T) {
	f := newFlow(t)
	exec := func(q string, args ...any) int64 {
		res, err := f.cfg.DB.Exec(q, args...)
		if err != nil {
			t.Fatal(err)
		}
		id, _ := res.LastInsertId()
		return id
	}
	var blocA int64
	f.cfg.DB.QueryRow(`SELECT building_id FROM units WHERE id = ?`, f.unit1).Scan(&blocA)
	blocB := exec(`INSERT INTO buildings (name, address, cadastral_number, total_area, association_id) VALUES ('Bloc B', 'y', 'CB-2', 500, ?)`, f.assocA)
	unitB := exec(`INSERT INTO units (cadastral_number, building_id, unit_number, address, entrance, area, part, unit_type, floor) VALUES ('CU-B1', ?, '1', 'y', 1, 70, 0.4, 'apartment', 2)`, blocB)
	dana := exec(`INSERT INTO owners (name, normalized_name, association_id) VALUES ('Dana', 'dana', ?)`, f.assocA)
	exec(`INSERT INTO ownerships (unit_id, owner_id, association_id, is_active, is_voting, registration_document, registration_date) VALUES (?, ?, ?, true, true, 'D', '2020-01-01 00:00:00')`, unitB, dana, f.assocA)
	foreign := exec(`INSERT INTO buildings (name, address, cadastral_number, total_area, association_id) VALUES ('Other', 'z', 'CB-3', 1, ?)`, f.assocB)

	create := func(buildings string, extra string) (int, domain.Gathering) {
		var g domain.Gathering
		code := f.do("POST", fmt.Sprintf("/associations/%d/gatherings", f.assocA),
			`{"title":"T","location":"x","gathering_type":"initial","gathering_date":"2026-11-01T18:00:00Z","qualification_buildings":`+buildings+extra+`}`, &g)
		return code, g
	}
	voters := func(g domain.Gathering) map[int64]bool {
		var list []struct {
			Owner struct{ ID int64 } `json:"owner"`
		}
		f.do("GET", f.g(g.ID, "/eligible-voters"), "", &list)
		out := map[int64]bool{}
		for _, v := range list {
			out[v.Owner.ID] = true
		}
		return out
	}

	if code, g := create(fmt.Sprintf("[%d]", blocB), ""); code != 201 || g.QualifiedUnitsCount != 1 || len(g.QualificationBuildings) != 1 || !voters(g)[dana] || voters(g)[f.alice] {
		t.Fatalf("Bloc B only: %d %+v %v", code, g, voters(g))
	}
	// Both buildings: one gathering, every unit of both, one pool of voters.
	if code, g := create(fmt.Sprintf("[%d,%d,%d]", blocA, blocB, blocA), ""); code != 201 || g.QualifiedUnitsCount != 4 || len(g.QualificationBuildings) != 2 || len(voters(g)) != 3 {
		t.Fatalf("both: %d %+v", code, g)
	}
	// Combined with the other rules (AND): floor 2 of both buildings.
	if _, g := create(fmt.Sprintf("[%d,%d]", blocA, blocB), `,"qualification_floors":[2]`); g.QualifiedUnitsCount != 1 {
		t.Fatalf("floor 2: %+v", g)
	}
	// None: every building.
	if _, g := create("[]", ""); g.QualifiedUnitsCount != 4 || g.QualificationBuildings != nil {
		t.Fatalf("all: %+v", g)
	}
	if code, _ := create(fmt.Sprintf("[%d]", foreign), ""); code != 400 {
		t.Fatalf("foreign building: %d", code)
	}
}
