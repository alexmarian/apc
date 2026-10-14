package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alexmarian/apc/api/internal/auth"
	"github.com/alexmarian/apc/api/internal/database"
	"github.com/pquerna/otp/totp"
)

const testPassword = "correct-horse"

// seedUser adds a user with testPassword and a real TOTP secret (returned).
// confirmed=false leaves them pending onboarding.
func seedUser(t *testing.T, cfg *ApiConfig, login string, isAdmin, confirmed bool) (database.User, string) {
	t.Helper()
	hash, err := auth.HashPassword(testPassword)
	if err != nil {
		t.Fatal(err)
	}
	enrollment, err := auth.GenerateTOTPEnrollment(login)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := cfg.DB.Exec(
		`INSERT INTO users (login, password_hash, topt_secret, is_admin, totp_confirmed_at) VALUES (?, ?, ?, ?, CASE WHEN ? THEN datetime('now') END)`,
		login, hash, enrollment.Secret, isAdmin, confirmed,
	); err != nil {
		t.Fatal(err)
	}
	user, err := cfg.Db.GetUserByLogin(t.Context(), login)
	if err != nil {
		t.Fatal(err)
	}
	return user, enrollment.Secret
}

func currentCode(t *testing.T, secret string) string {
	t.Helper()
	code, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	return code
}

func post(handler http.HandlerFunc, body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	handler(rec, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body)))
	return rec
}

// asAdmin calls an admin endpoint as root (seeded as app admin by
// newAssociationsTestConfig), with optional {userId}.
func asAdmin(t *testing.T, cfg *ApiConfig, handler http.HandlerFunc, method, userId, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, "/", strings.NewReader(body))
	if userId != "" {
		req.SetPathValue(UserIdPathValue, userId)
	}
	return authed(t, cfg, handler, "root", true, req)
}

func createInvitation(t *testing.T, cfg *ApiConfig, body string) RegistrationToken {
	t.Helper()
	rec := asAdmin(t, cfg, HandleCreateRegistrationToken(cfg), http.MethodPost, "", body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create token: %d %s", rec.Code, rec.Body.String())
	}
	var token RegistrationToken
	_ = json.Unmarshal(rec.Body.Bytes(), &token)
	return token
}

func TestRegistrationAndOnboarding(t *testing.T) {
	cfg := newAssociationsTestConfig(t)
	association := decodeAssociation(t, createAssociation(t, cfg, `{"name":"One","address":"a"}`))
	invite := createInvitation(t, cfg, fmt.Sprintf(`{"expiration_hours":24,"description":"for carla","association_ids":[%d]}`, association.ID))
	if len(invite.Associations) != 1 || invite.Associations[0].Name != "One" {
		t.Fatalf("invite associations = %+v", invite.Associations)
	}

	// The code can be checked first, without using it.
	check := func(token string) int {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.SetPathValue(InvitationTokenPathValue, token)
		rec := httptest.NewRecorder()
		HandleGetInvitation(cfg)(rec, req)
		return rec.Code
	}
	if code := check(invite.Token); code != http.StatusOK {
		t.Fatalf("check a valid invitation: %d", code)
	}
	if code := check("nope"); code != http.StatusNotFound {
		t.Fatalf("check an unknown invitation: %d", code)
	}

	// Register: pending user, linked to the invitation's associations.
	rec := post(HandleCreateUserWithToken(cfg), fmt.Sprintf(`{"login":"carla","password":%q,"token":%q}`, testPassword, invite.Token))
	if rec.Code != http.StatusCreated {
		t.Fatalf("register: %d %s", rec.Code, rec.Body.String())
	}
	var registered struct {
		Enrollment struct{ QRCode, Secret, OtpauthURL string } `json:"enrollment"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &registered)
	if registered.Enrollment.Secret == "" || registered.Enrollment.QRCode == "" || !strings.HasPrefix(registered.Enrollment.OtpauthURL, "otpauth://") {
		t.Fatalf("missing enrollment: %s", rec.Body.String())
	}
	if ids, _ := cfg.Db.GetUserAssociationsByLogin(t.Context(), "carla"); len(ids) != 1 || ids[0] != association.ID {
		t.Fatalf("carla's associations = %v", ids)
	}

	// The token is spent.
	if code := check(invite.Token); code != http.StatusNotFound {
		t.Fatalf("check a used invitation: %d", code)
	}
	if rec := post(HandleCreateUserWithToken(cfg), fmt.Sprintf(`{"login":"carla2","password":%q,"token":%q}`, testPassword, invite.Token)); rec.Code != http.StatusUnauthorized {
		t.Fatalf("reused token: %d", rec.Code)
	}

	// Pending: login says onboarding is required (only with the right password).
	login := func(code string) *httptest.ResponseRecorder {
		return post(HandleLogin(cfg), fmt.Sprintf(`{"login":"carla","password":%q,"totp":%q}`, testPassword, code))
	}
	if rec := login("000000"); rec.Code != http.StatusForbidden || errorMsg(rec) != OnboardingRequiredMessage {
		t.Fatalf("pending login: %d %s", rec.Code, rec.Body.String())
	}
	if rec := post(HandleLogin(cfg), `{"login":"carla","password":"wrong-password","totp":"000000"}`); rec.Code != http.StatusUnauthorized {
		t.Fatalf("wrong password: %d", rec.Code)
	}

	// Resuming issues a fresh secret; the first one stops working.
	rec = post(HandleStartOnboarding(cfg), fmt.Sprintf(`{"login":"carla","password":%q}`, testPassword))
	if rec.Code != http.StatusOK {
		t.Fatalf("start onboarding: %d %s", rec.Code, rec.Body.String())
	}
	var resumed struct {
		Enrollment struct{ Secret string } `json:"enrollment"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resumed)
	if resumed.Enrollment.Secret == "" || resumed.Enrollment.Secret == registered.Enrollment.Secret {
		t.Fatal("resume must issue a new secret")
	}

	confirm := func(code string) *httptest.ResponseRecorder {
		return post(HandleConfirmOnboarding(cfg), fmt.Sprintf(`{"login":"carla","password":%q,"code":%q}`, testPassword, code))
	}
	if rec := confirm(currentCode(t, registered.Enrollment.Secret)); rec.Code != http.StatusBadRequest {
		t.Fatalf("old secret's code: %d", rec.Code)
	}
	rec = confirm(currentCode(t, resumed.Enrollment.Secret))
	if rec.Code != http.StatusOK {
		t.Fatalf("confirm: %d %s", rec.Code, rec.Body.String())
	}
	var session sessionResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &session)
	if session.Token == "" || session.RefreshToken == "" {
		t.Fatalf("confirm must sign in: %s", rec.Body.String())
	}

	// Active now: normal login works; onboarding is closed.
	if rec := login(currentCode(t, resumed.Enrollment.Secret)); rec.Code != http.StatusOK {
		t.Fatalf("login after onboarding: %d %s", rec.Code, rec.Body.String())
	}
	if rec := confirm(currentCode(t, resumed.Enrollment.Secret)); rec.Code != http.StatusConflict {
		t.Fatalf("confirm when active: %d", rec.Code)
	}
}

func TestRegistration_Validation(t *testing.T) {
	cfg := newAssociationsTestConfig(t)
	register := func(login, password string) *httptest.ResponseRecorder {
		invite := createInvitation(t, cfg, `{"expiration_hours":1,"description":"x"}`)
		return post(HandleCreateUserWithToken(cfg), fmt.Sprintf(`{"login":%q,"password":%q,"token":%q}`, login, password, invite.Token))
	}
	if rec := register("root", testPassword); rec.Code != http.StatusConflict {
		t.Fatalf("taken login: %d", rec.Code)
	}
	if rec := register("a b", testPassword); rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid login: %d", rec.Code)
	}
	if rec := register("newuser", "short"); rec.Code != http.StatusBadRequest {
		t.Fatalf("short password: %d", rec.Code)
	}
	if rec := post(HandleCreateUserWithToken(cfg), fmt.Sprintf(`{"login":"newuser","password":%q,"token":"nope"}`, testPassword)); rec.Code != http.StatusUnauthorized {
		t.Fatalf("unknown token: %d", rec.Code)
	}
	// A rejected registration must not burn the token.
	invite := createInvitation(t, cfg, `{"expiration_hours":1,"description":"x"}`)
	post(HandleCreateUserWithToken(cfg), fmt.Sprintf(`{"login":"root","password":%q,"token":%q}`, testPassword, invite.Token))
	if _, err := cfg.Db.GetValidRegistrationToken(t.Context(), invite.Token); err != nil {
		t.Fatalf("token used up by a failed registration: %v", err)
	}
	if rec := asAdmin(t, cfg, HandleCreateRegistrationToken(cfg), http.MethodPost, "", `{"expiration_hours":1,"description":"x","is_admin":true,"association_ids":[1]}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("admin invitation with associations: %d", rec.Code)
	}
}

func TestDisableUser(t *testing.T) {
	cfg := newAssociationsTestConfig(t)
	dana, secret := seedUser(t, cfg, "dana", false, true)
	session, err := issueSession(t.Context(), cfg, dana, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	id := fmt.Sprint(dana.ID)

	rec := asAdmin(t, cfg, HandleSetUserStatus(cfg), http.MethodPut, id, `{"disabled":true}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("disable: %d %s", rec.Code, rec.Body.String())
	}
	var got AdminUser
	_ = json.Unmarshal(rec.Body.Bytes(), &got)
	if got.Status != UserStatusDisabled || got.DisabledBy != "root" {
		t.Fatalf("after disable: %+v", got)
	}

	if rec := post(HandleLogin(cfg), fmt.Sprintf(`{"login":"dana","password":%q,"totp":%q}`, testPassword, currentCode(t, secret))); rec.Code != http.StatusUnauthorized {
		t.Fatalf("disabled login: %d", rec.Code)
	}
	refresh := httptest.NewRequest(http.MethodPost, "/", nil)
	refresh.Header.Set("Authorization", "Bearer "+session.RefreshToken)
	refreshRec := httptest.NewRecorder()
	HandleRefresh(cfg)(refreshRec, refresh)
	if refreshRec.Code != http.StatusUnauthorized {
		t.Fatalf("disabled refresh: %d", refreshRec.Code)
	}
	if rec := authed(t, cfg, HandleGetUserAssociations(cfg), "dana", false, httptest.NewRequest(http.MethodGet, "/", nil)); rec.Code != http.StatusUnauthorized {
		t.Fatalf("disabled user's access token: %d", rec.Code)
	}

	if rec := asAdmin(t, cfg, HandleSetUserStatus(cfg), http.MethodPut, id, `{"disabled":false}`); rec.Code != http.StatusOK {
		t.Fatalf("enable: %d", rec.Code)
	}
	if rec := post(HandleLogin(cfg), fmt.Sprintf(`{"login":"dana","password":%q,"totp":%q}`, testPassword, currentCode(t, secret))); rec.Code != http.StatusOK {
		t.Fatalf("re-enabled login: %d", rec.Code)
	}
}

func TestAdminGuardsAndRoles(t *testing.T) {
	cfg := newAssociationsTestConfig(t)
	created := decodeAssociation(t, createAssociation(t, cfg, `{"name":"One","address":"a","administratorIds":[2]}`))

	// Not on yourself (root is user 1).
	if rec := asAdmin(t, cfg, HandleSetUserStatus(cfg), http.MethodPut, "1", `{"disabled":true}`); rec.Code != http.StatusConflict {
		t.Fatalf("disable self: %d", rec.Code)
	}
	if rec := asAdmin(t, cfg, HandleSetUserRole(cfg), http.MethodPut, "1", `{"isAdmin":false}`); rec.Code != http.StatusConflict {
		t.Fatalf("demote self: %d", rec.Code)
	}
	if rec := asAdmin(t, cfg, HandleSetUserRole(cfg), http.MethodPut, "99", `{"isAdmin":true}`); rec.Code != http.StatusNotFound {
		t.Fatalf("unknown user: %d", rec.Code)
	}

	// Promoting ana (2) drops her association links.
	rec := asAdmin(t, cfg, HandleSetUserRole(cfg), http.MethodPut, "2", `{"isAdmin":true}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("promote: %d %s", rec.Code, rec.Body.String())
	}
	var ana AdminUser
	_ = json.Unmarshal(rec.Body.Bytes(), &ana)
	if !ana.IsAdmin || len(ana.Associations) != 0 {
		t.Fatalf("promoted ana: %+v", ana)
	}
	// ...and the middleware sees the new role immediately, whatever her token says.
	if rec := authed(t, cfg, cfg.MiddlewareAdminOnly(HandleListUsers(cfg)), "ana", false, httptest.NewRequest(http.MethodGet, "/", nil)); rec.Code != http.StatusOK {
		t.Fatalf("promoted user with an old non-admin token: %d", rec.Code)
	}

	// Associations for an app admin are refused; for bob (3) they replace his links.
	if rec := asAdmin(t, cfg, HandleSetUserAssociations(cfg), http.MethodPut, "2", fmt.Sprintf(`{"associationIds":[%d]}`, created.ID)); rec.Code != http.StatusBadRequest {
		t.Fatalf("associations for an admin: %d", rec.Code)
	}
	if rec := asAdmin(t, cfg, HandleSetUserAssociations(cfg), http.MethodPut, "3", `{"associationIds":[999]}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("unknown association: %d", rec.Code)
	}
	rec = asAdmin(t, cfg, HandleSetUserAssociations(cfg), http.MethodPut, "3", fmt.Sprintf(`{"associationIds":[%d,%d]}`, created.ID, created.ID))
	if rec.Code != http.StatusOK {
		t.Fatalf("set associations: %d %s", rec.Code, rec.Body.String())
	}
	var bob AdminUser
	_ = json.Unmarshal(rec.Body.Bytes(), &bob)
	if len(bob.Associations) != 1 || bob.Associations[0].Name != "One" {
		t.Fatalf("bob's associations: %+v", bob.Associations)
	}

	// The list reports role, status and associations.
	rec = asAdmin(t, cfg, HandleListUsers(cfg), http.MethodGet, "", "")
	var users []AdminUser
	_ = json.Unmarshal(rec.Body.Bytes(), &users)
	if len(users) != 3 || users[0].Login != "ana" || users[1].Login != "bob" || users[1].Status != UserStatusPending {
		t.Fatalf("users: %+v", users)
	}
}

func TestPasswordReset(t *testing.T) {
	cfg := newAssociationsTestConfig(t)
	boss, _ := seedUser(t, cfg, "boss", true, true)

	resetToken := func(login string) string {
		rec := asAdmin(t, cfg, HandleRequestPasswordReset(cfg), http.MethodPost, "", fmt.Sprintf(`{"login":%q}`, login))
		if rec.Code != http.StatusOK {
			t.Fatalf("request reset: %d %s", rec.Code, rec.Body.String())
		}
		var body struct{ Token string }
		_ = json.Unmarshal(rec.Body.Bytes(), &body)
		return body.Token
	}
	if rec := asAdmin(t, cfg, HandleRequestPasswordReset(cfg), http.MethodPost, "", `{"login":"nobody"}`); rec.Code != http.StatusNotFound {
		t.Fatalf("unknown login: %d", rec.Code)
	}

	// Password only: still an admin, still active.
	rec := post(HandleResetPassword(cfg), fmt.Sprintf(`{"token":%q,"new_password":"brand-new-pass"}`, resetToken("boss")))
	if rec.Code != http.StatusOK {
		t.Fatalf("reset: %d %s", rec.Code, rec.Body.String())
	}
	after, _ := cfg.Db.GetUserByID(t.Context(), boss.ID)
	if !after.IsAdmin || userStatus(after) != UserStatusActive || auth.CheckPasswordHash("brand-new-pass", after.PasswordHash) != nil {
		t.Fatalf("after password reset: admin=%v status=%s", after.IsAdmin, userStatus(after))
	}

	// With TOTP: a new secret, and pending until confirmed.
	rec = post(HandleResetPassword(cfg), fmt.Sprintf(`{"token":%q,"new_password":"another-pass","reset_totp_secret":true}`, resetToken("boss")))
	var body struct {
		OnboardingRequired bool `json:"onboardingRequired"`
		Enrollment         struct{ Secret string }
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	after, _ = cfg.Db.GetUserByID(t.Context(), boss.ID)
	if !body.OnboardingRequired || body.Enrollment.Secret == "" || after.ToptSecret != body.Enrollment.Secret || userStatus(after) != UserStatusPending || !after.IsAdmin {
		t.Fatalf("after TOTP reset: %s, status=%s", rec.Body.String(), userStatus(after))
	}

	// Tokens are single-use, and short passwords are refused.
	token := resetToken("boss")
	if rec := post(HandleResetPassword(cfg), fmt.Sprintf(`{"token":%q,"new_password":"short"}`, token)); rec.Code != http.StatusBadRequest {
		t.Fatalf("short password: %d", rec.Code)
	}
	post(HandleResetPassword(cfg), fmt.Sprintf(`{"token":%q,"new_password":"long-enough"}`, token))
	if rec := post(HandleResetPassword(cfg), fmt.Sprintf(`{"token":%q,"new_password":"long-enough"}`, token)); rec.Code != http.StatusBadRequest {
		t.Fatalf("reused reset token: %d", rec.Code)
	}
}
