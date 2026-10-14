package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alexmarian/apc/api/internal/auth"
	"github.com/alexmarian/apc/api/internal/handlers/gathering/domain"
)

// An owner account votes for its owner record exactly like a member link:
// same units, same slot rules, and one ballot per unit across both channels.
func TestMyGatheringsAndBallot(t *testing.T) {
	f := newFlow(t)
	account := func(login string, owners ...int64) {
		res, err := f.cfg.DB.Exec(`INSERT INTO users (login, password_hash, topt_secret, account_type) VALUES (?, 'x', '', 'owner')`, login)
		if err != nil {
			t.Fatal(err)
		}
		id, _ := res.LastInsertId()
		for _, o := range owners {
			if _, err := f.cfg.DB.Exec(`INSERT INTO user_owners (user_id, owner_id) VALUES (?, ?)`, id, o); err != nil {
				t.Fatal(err)
			}
		}
	}
	account("alice", f.alice)
	account("carol", f.carol)
	account("nobody")
	f.mux.HandleFunc("GET /v1/api/me/gatherings", f.cfg.MiddlewareAuth(HandleMyGatherings(f.cfg)))
	f.mux.HandleFunc("GET /v1/api/me/gatherings/{gatheringId}", f.cfg.MiddlewareAuth(HandleMyGathering(f.cfg)))
	f.mux.HandleFunc("POST /v1/api/me/gatherings/{gatheringId}/ballot", f.cfg.MiddlewareAuth(HandleMyBallot(f.cfg)))
	f.mux.HandleFunc("POST /v1/api/member/gatherings/{memberToken}/ballot", f.cfg.MiddlewareMemberToken(NewMemberBallotHandler(f.cfg).HandleSubmitMemberBallot()))
	as := func(login, method, path, body string) (int, string) {
		t.Helper()
		token, _, _ := auth.MakeJWT(login, f.cfg.Secret, time.Hour, false)
		req := httptest.NewRequest(method, "/v1/api"+path, strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		f.mux.ServeHTTP(rec, req)
		return rec.Code, rec.Body.String()
	}

	var g domain.Gathering
	f.do("POST", fmt.Sprintf("/associations/%d/gatherings", f.assocA), `{"title":"Annual","description":"d","intent":"d","location":"Hall","gathering_date":"2026-11-01T18:00:00Z","gathering_type":"initial","voting_mode":"by_weight","qualification_unit_types":[]}`, &g)
	var budget domain.VotingMatter
	f.do("POST", f.g(g.ID, "/matters"), `{"order_index":1,"title":"Budget","matter_type":"budget","voting_config":{"type":"yes_no","required_majority":"simple","allow_abstention":true}}`, &budget)

	// Drafts stay hidden.
	if code, body := as("alice", "GET", "/me/gatherings", ""); code != 200 || strings.TrimSpace(body) != "[]" {
		t.Fatalf("draft listed: %d %s", code, body)
	}
	if code, _ := as("alice", "GET", fmt.Sprintf("/me/gatherings/%d", g.ID), ""); code != 404 {
		t.Fatalf("draft readable: %d", code)
	}

	for _, s := range []string{"published", "active"} {
		f.do("PUT", f.g(g.ID, "/status"), `{"status":"`+s+`"}`, nil)
	}
	code, body := as("alice", "GET", "/me/gatherings", "")
	if code != 200 || !strings.Contains(body, `"title":"Annual"`) || !strings.Contains(body, `"association_name":"A"`) || strings.Count(body, `"is_available":true`) != 2 {
		t.Fatalf("alice's gatherings: %d %s", code, body)
	}
	if _, body := as("carol", "GET", "/me/gatherings", ""); strings.Contains(body, "Annual") {
		t.Fatalf("carol sees A's gathering: %s", body)
	}
	if code, _ := as("carol", "GET", fmt.Sprintf("/me/gatherings/%d", g.ID), ""); code != 404 {
		t.Fatalf("carol reads A's gathering: %d", code)
	}
	if code, _ := as("carol", "POST", fmt.Sprintf("/me/gatherings/%d/ballot", g.ID), "{}"); code != 404 {
		t.Fatalf("carol votes in A: %d", code)
	}
	if code, body := as("nobody", "GET", "/me/gatherings", ""); code != 200 || strings.TrimSpace(body) != "[]" {
		t.Fatalf("account without owners: %d %s", code, body)
	}

	content := fmt.Sprintf(`{"ballot_content":{"%d":{"matter_id":%d,"values":["yes"]}}}`, budget.ID, budget.ID)
	if code, body := as("alice", "POST", fmt.Sprintf("/me/gatherings/%d/ballot", g.ID), content); code != http.StatusCreated || !strings.Contains(body, "ballot_hash") {
		t.Fatalf("alice votes: %d %s", code, body)
	}
	var performedBy, details string
	f.cfg.DB.QueryRow(`SELECT performed_by, details FROM voting_audit_log WHERE entity_type = 'ballot'`).Scan(&performedBy, &details)
	if performedBy != "user:alice" || !strings.Contains(details, `"source":"owner_account"`) {
		t.Fatalf("audit: %q %q", performedBy, details)
	}
	if code, body := as("alice", "GET", fmt.Sprintf("/me/gatherings/%d", g.ID), ""); code != 200 || strings.Contains(body, `"is_available":true`) || !strings.Contains(body, `"ballot_hash"`) {
		t.Fatalf("after voting: %d %s", code, body)
	}
	if code, _ := as("alice", "POST", fmt.Sprintf("/me/gatherings/%d/ballot", g.ID), content); code != http.StatusConflict {
		t.Fatalf("second ballot: %d", code)
	}

	// The member link of the same owner finds nothing left to vote.
	var inv map[string]any
	f.do("POST", f.g(g.ID, "/invitations"), fmt.Sprintf(`{"owner_id":%d}`, f.alice), &inv)
	if code, _ := as("", "POST", fmt.Sprintf("/member/gatherings/%v/ballot", inv["token"]), content); code != http.StatusConflict {
		t.Fatalf("link after account vote: %d", code)
	}
}
