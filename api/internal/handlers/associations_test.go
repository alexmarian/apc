package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/alexmarian/apc/api/internal/auth"
	"github.com/alexmarian/apc/api/internal/database"
	_ "github.com/mattn/go-sqlite3"
)

// newAssociationsTestConfig opens an in-memory database migrated with the
// real schema and seeds users: root (app admin, id 1), ana (2), bob (3).
func newAssociationsTestConfig(t *testing.T) *ApiConfig {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	// Every pooled connection to ":memory:" would be a separate, empty database.
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })
	if err := database.RunMigrations(db, os.DirFS("../../sql/schema")); err != nil {
		t.Fatal(err)
	}
	for _, u := range []struct {
		login   string
		isAdmin bool
	}{{"root", true}, {"ana", false}, {"bob", false}} {
		if _, err := db.Exec(`INSERT INTO users (login, password_hash, topt_secret, is_admin) VALUES (?, 'x', 'x', ?)`, u.login, u.isAdmin); err != nil {
			t.Fatal(err)
		}
	}
	return &ApiConfig{Db: database.New(db), DB: db, Secret: "test-secret"}
}

// authed wraps handler in the auth middleware and calls it as login.
func authed(t *testing.T, cfg *ApiConfig, handler http.HandlerFunc, login string, isAdmin bool, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	token, _, err := auth.MakeJWT(login, cfg.Secret, time.Hour, isAdmin)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	cfg.MiddlewareAuth(handler)(rec, req)
	return rec
}

func createAssociation(t *testing.T, cfg *ApiConfig, body string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	HandleCreateAssociation(cfg)(rec, httptest.NewRequest(http.MethodPost, "/v1/api/associations", strings.NewReader(body)))
	return rec
}

func updateAssociation(cfg *ApiConfig, id string, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPut, "/v1/api/associations/"+id, strings.NewReader(body))
	req.SetPathValue(AssociationIdPathValue, id)
	rec := httptest.NewRecorder()
	HandleUpdateAssociation(cfg)(rec, req)
	return rec
}

func decodeAssociation(t *testing.T, rec *httptest.ResponseRecorder) Association {
	t.Helper()
	var got Association
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode %q: %v", rec.Body.String(), err)
	}
	return got
}

func errorMsg(rec *httptest.ResponseRecorder) string {
	var got ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &got)
	return got.Msg
}

func TestHandleCreateAssociation_Created(t *testing.T) {
	cfg := newAssociationsTestConfig(t)
	rec := createAssociation(t, cfg, `{"name":"  Blocul Verde ","address":"Str. Florilor 1","administratorIds":[3,2,3]}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	got := decodeAssociation(t, rec)
	if got.ID == 0 || got.Name != "Blocul Verde" || got.Address != "Str. Florilor 1" {
		t.Fatalf("unexpected association: %+v", got)
	}
	want := []UserRef{{2, "ana"}, {3, "bob"}}
	if fmt.Sprint(got.Administrators) != fmt.Sprint(want) {
		t.Fatalf("administrators = %v, want %v", got.Administrators, want)
	}
	if got.Administrator != "ana, bob" {
		t.Fatalf("legacy administrator = %q", got.Administrator)
	}
	ids, _ := cfg.Db.GetUserAssociationsByLogin(httptest.NewRequest(http.MethodGet, "/", nil).Context(), "ana")
	if len(ids) != 1 || ids[0] != got.ID {
		t.Fatalf("ana's associations = %v", ids)
	}
}

func TestHandleCreateAssociation_NoAdministrators(t *testing.T) {
	cfg := newAssociationsTestConfig(t)
	rec := createAssociation(t, cfg, `{"name":"n","address":"a"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if got := decodeAssociation(t, rec); len(got.Administrators) != 0 || got.Administrator != "" {
		t.Fatalf("unexpected administrators: %+v", got)
	}
}

func TestHandleCreateAssociation_Validation(t *testing.T) {
	cfg := newAssociationsTestConfig(t)
	tests := []struct {
		name    string
		body    string
		wantMsg string
	}{
		{"malformed", `{`, "Invalid request format"},
		{"missing name", `{"name":"  ","address":"a"}`, "Name is required"},
		{"missing address", `{"name":"n"}`, "Address is required"},
		{"too long", `{"name":"` + strings.Repeat("x", 256) + `","address":"a"}`, "Name must be at most 255 characters"},
		{"unknown administrator", `{"name":"n","address":"a","administratorIds":[2,99]}`, "Unknown or invalid administrator"},
		{"app admin as administrator", `{"name":"n","address":"a","administratorIds":[1]}`, "Unknown or invalid administrator"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := createAssociation(t, cfg, tt.body)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", rec.Code)
			}
			if got := errorMsg(rec); got != tt.wantMsg {
				t.Fatalf("msg = %q, want %q", got, tt.wantMsg)
			}
		})
	}
}

func TestHandleUpdateAssociation(t *testing.T) {
	cfg := newAssociationsTestConfig(t)
	created := decodeAssociation(t, createAssociation(t, cfg, `{"name":"Old","address":"A","administratorIds":[2]}`))
	id := fmt.Sprint(created.ID)

	rec := updateAssociation(cfg, id, `{"name":"New","address":"B","administratorIds":[3]}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	got := decodeAssociation(t, rec)
	if got.Name != "New" || got.Address != "B" || got.Administrator != "bob" || fmt.Sprint(got.Administrators) != fmt.Sprint([]UserRef{{3, "bob"}}) {
		t.Fatalf("unexpected association: %+v", got)
	}
	ctx := httptest.NewRequest(http.MethodGet, "/", nil).Context()
	if ids, _ := cfg.Db.GetUserAssociationsByLogin(ctx, "ana"); len(ids) != 0 {
		t.Fatalf("ana should have lost access, has %v", ids)
	}
	if ids, _ := cfg.Db.GetUserAssociationsByLogin(ctx, "bob"); len(ids) != 1 {
		t.Fatalf("bob should have access, has %v", ids)
	}

	if rec := updateAssociation(cfg, "999", `{"name":"n","address":"a"}`); rec.Code != http.StatusNotFound {
		t.Fatalf("missing association status = %d, want 404", rec.Code)
	}
	if rec := updateAssociation(cfg, "abc", `{"name":"n","address":"a"}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("bad id status = %d, want 400", rec.Code)
	}
	// A failed validation must not touch the existing links.
	if rec := updateAssociation(cfg, id, `{"name":"n","address":"a","administratorIds":[1]}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("admin administrator status = %d, want 400", rec.Code)
	}
	if ids, _ := cfg.Db.GetUserAssociationsByLogin(ctx, "bob"); len(ids) != 1 {
		t.Fatalf("bob's access changed by a rejected update: %v", ids)
	}
}

func TestHandleGetUserAssociations_Scoping(t *testing.T) {
	cfg := newAssociationsTestConfig(t)
	createAssociation(t, cfg, `{"name":"One","address":"a","administratorIds":[2]}`)
	createAssociation(t, cfg, `{"name":"Two","address":"a","administratorIds":[3]}`)

	list := func(login string, isAdmin bool) []Association {
		rec := authed(t, cfg, HandleGetUserAssociations(cfg), login, isAdmin, httptest.NewRequest(http.MethodGet, "/v1/api/associations", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
		}
		var got []Association
		_ = json.Unmarshal(rec.Body.Bytes(), &got)
		return got
	}

	if got := list("root", true); len(got) != 2 {
		t.Fatalf("app admin sees %d associations, want 2", len(got))
	}
	if got := list("ana", false); len(got) != 1 || got[0].Name != "One" || got[0].Administrators[0].Login != "ana" {
		t.Fatalf("ana sees %+v, want only One", got)
	}
	if _, err := cfg.DB.Exec(`INSERT INTO users (login, password_hash, topt_secret) VALUES ('nobody', 'x', 'x')`); err != nil {
		t.Fatal(err)
	}
	if got := list("nobody", false); len(got) != 0 {
		t.Fatalf("unlinked user sees %+v, want none", got)
	}
}

func TestMiddlewareAssociationResource_Membership(t *testing.T) {
	cfg := newAssociationsTestConfig(t)
	created := decodeAssociation(t, createAssociation(t, cfg, `{"name":"One","address":"a","administratorIds":[2]}`))
	id := fmt.Sprint(created.ID)
	ok := func(rw http.ResponseWriter, _ *http.Request) { rw.WriteHeader(http.StatusNoContent) }

	call := func(login string, isAdmin bool, associationId string) int {
		token, _, _ := auth.MakeJWT(login, cfg.Secret, time.Hour, isAdmin)
		req := httptest.NewRequest(http.MethodGet, "/v1/api/associations/"+associationId+"/buildings", nil)
		req.SetPathValue(AssociationIdPathValue, associationId)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		cfg.MiddlewareAssociationResource(ok)(rec, req)
		return rec.Code
	}

	if code := call("root", true, id); code != http.StatusNoContent {
		t.Fatalf("app admin: %d", code)
	}
	if code := call("ana", false, id); code != http.StatusNoContent {
		t.Fatalf("linked user: %d", code)
	}
	if code := call("bob", false, id); code != http.StatusForbidden {
		t.Fatalf("unlinked user: %d, want 403", code)
	}
	if code := call("ana", false, "0"); code != http.StatusBadRequest {
		t.Fatalf("bad id: %d, want 400", code)
	}
}

func TestHandleCreateAssociation_AdminOnly(t *testing.T) {
	// Db is nil so the auth middleware skips the revoked-token lookup.
	cfg := &ApiConfig{Secret: "test-secret"}
	handler := cfg.MiddlewareAdminOnly(HandleCreateAssociation(cfg))

	token, _, err := auth.MakeJWT("member", cfg.Secret, time.Hour, false)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/v1/api/associations", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("non-admin status = %d, want 403", rec.Code)
	}

	rec = httptest.NewRecorder()
	handler(rec, httptest.NewRequest(http.MethodPost, "/v1/api/associations", strings.NewReader(`{}`)))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous status = %d, want 401", rec.Code)
	}
}

func TestHandleUpdateUser_CannotSelfPromote(t *testing.T) {
	cfg := newAssociationsTestConfig(t)
	body := `{"login":"ana","password":"new-password-123","isAdmin":true}`
	rec := authed(t, cfg, HandleUpdateUser(cfg), "ana", false, httptest.NewRequest(http.MethodPut, "/v1/api/users", strings.NewReader(body)))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	user, err := cfg.Db.GetUserByLogin(httptest.NewRequest(http.MethodGet, "/", nil).Context(), "ana")
	if err != nil {
		t.Fatal(err)
	}
	if user.IsAdmin {
		t.Fatal("a non-admin promoted themselves to app admin")
	}
}
