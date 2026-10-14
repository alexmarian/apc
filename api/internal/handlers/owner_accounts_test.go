package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alexmarian/apc/api/internal/auth"
	"github.com/pquerna/otp/totp"
)

// ownerAccounts is the units fixture (association A: Alice and Bob; B: Zed;
// staff users root/ana/bob) served through the real routes and middleware.
type ownerAccountsFixture struct {
	unitsFixture
	t   *testing.T
	mux *http.ServeMux
}

func newOwnerAccountsFixture(t *testing.T) ownerAccountsFixture {
	f := ownerAccountsFixture{unitsFixture: newUnitsFixture(t), t: t, mux: http.NewServeMux()}
	f.cfg.PhotoDir = t.TempDir()
	cfg := f.cfg
	m := f.mux
	m.HandleFunc("POST /login", HandleLogin(cfg))
	m.HandleFunc("GET /activate/{activationToken}", HandleGetActivation(cfg))
	m.HandleFunc("POST /activate", HandleActivate(cfg))
	m.HandleFunc("GET /me", cfg.MiddlewareAuth(HandleGetMe(cfg)))
	m.HandleFunc("GET /me/properties", cfg.MiddlewareAuth(HandleGetMyProperties(cfg)))
	m.HandleFunc("PUT /me/profile", cfg.MiddlewareAuth(HandleUpdateMyProfile(cfg)))
	m.HandleFunc("PUT /me/password", cfg.MiddlewareAuth(HandleChangePassword(cfg)))
	m.HandleFunc("POST /me/totp", cfg.MiddlewareAuth(HandleStartMyTotp(cfg)))
	m.HandleFunc("POST /me/totp/confirm", cfg.MiddlewareAuth(HandleConfirmMyTotp(cfg)))
	m.HandleFunc("DELETE /me/totp", cfg.MiddlewareAuth(HandleDisableMyTotp(cfg)))
	m.HandleFunc("POST /me/owners", cfg.MiddlewareAuth(HandleLinkMyOwner(cfg)))
	m.HandleFunc("PUT /me/photo", cfg.MiddlewareAuth(HandleUploadMyPhoto(cfg)))
	m.HandleFunc("DELETE /me/photo", cfg.MiddlewareAuth(HandleDeleteMyPhoto(cfg)))
	m.HandleFunc("GET /v1/api/photos/{photoUserId}", HandleGetPhoto(cfg))
	base := "/associations/{associationId}/owners"
	m.HandleFunc("GET "+base+"/accounts", cfg.MiddlewareAssociationResource(HandleListOwnerAccounts(cfg)))
	m.HandleFunc("POST "+base+"/activations", cfg.MiddlewareAssociationResource(HandleBulkOwnerActivations(cfg)))
	m.HandleFunc("GET "+base+"/{ownerId}", cfg.MiddlewareAssociationResource(HandleGetAssociationOwner(cfg)))
	m.HandleFunc("GET "+base+"/{ownerId}/account", cfg.MiddlewareAssociationResource(HandleGetOwnerAccount(cfg)))
	m.HandleFunc("DELETE "+base+"/{ownerId}/account", cfg.MiddlewareAssociationResource(HandleUnlinkOwnerAccount(cfg)))
	m.HandleFunc("POST "+base+"/{ownerId}/activation", cfg.MiddlewareAssociationResource(HandleCreateOwnerActivation(cfg)))
	m.HandleFunc("DELETE "+base+"/{ownerId}/activation", cfg.MiddlewareAssociationResource(HandleRevokeOwnerActivation(cfg)))
	m.HandleFunc("PUT /admin/users/{userId}/role", cfg.MiddlewareAdminOnly(HandleSetUserRole(cfg)))
	return f
}

// as calls the API as login (admin when isAdmin); body may be JSON or nil.
func (f ownerAccountsFixture) as(login string, isAdmin bool, method, path string, body any) (int, map[string]any) {
	f.t.Helper()
	var reader *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, reader)
	if login != "" {
		token, _, _ := auth.MakeJWT(login, f.cfg.Secret, time.Hour, isAdmin)
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	f.mux.ServeHTTP(rec, req)
	var out map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	return rec.Code, out
}

func (f ownerAccountsFixture) root(method, path string, body any) (int, map[string]any) {
	return f.as("root", true, method, path, body)
}

func (f ownerAccountsFixture) owner(id int64, suffix string) string {
	return fmt.Sprintf("/associations/%d/owners/%d%s", f.assocA, id, suffix)
}

func (f ownerAccountsFixture) activate(ownerID int64, login string) string {
	f.t.Helper()
	code, link := f.root("POST", f.owner(ownerID, "/activation"), nil)
	if code != http.StatusCreated {
		f.t.Fatalf("create activation: %d %v", code, link)
	}
	token := link["token"].(string)
	if login != "" {
		if code, out := f.as("", false, "POST", "/activate", map[string]string{"token": token, "login": login, "password": "owner-pass-1"}); code != http.StatusCreated {
			f.t.Fatalf("activate: %d %v", code, out)
		}
	}
	return token
}

func TestOwnerActivationAndLogin(t *testing.T) {
	f := newOwnerAccountsFixture(t)

	code, link := f.root("POST", f.owner(f.alice, "/activation"), nil)
	if code != http.StatusCreated || link["token"] == "" || link["owner_name"] != "Alice" {
		t.Fatalf("create: %d %v", code, link)
	}
	token := link["token"].(string)
	if code, _ := f.root("POST", f.owner(f.alice, "/activation"), nil); code != http.StatusConflict {
		t.Fatalf("second pending link: %d", code)
	}
	if code, status := f.root("GET", f.owner(f.alice, "/account"), nil); code != 200 || status["status"] != "invited" {
		t.Fatalf("status invited: %d %v", code, status)
	}

	// The owner opens the link.
	if code, info := f.as("", false, "GET", "/activate/"+token, nil); code != 200 || info["owner_name"] != "Alice" || info["association_name"] != "A" {
		t.Fatalf("activation info: %d %v", code, info)
	}
	for name, tc := range map[string]struct {
		login, password string
		code            int
	}{
		"weak password": {"alice.pop", "short", 400},
		"taken login":   {"root", "owner-pass-1", 409},
	} {
		if code, out := f.as("", false, "POST", "/activate", map[string]string{"token": token, "login": tc.login, "password": tc.password}); code != tc.code {
			t.Errorf("%s: %d %v", name, code, out)
		}
	}
	code, session := f.as("", false, "POST", "/activate", map[string]string{"token": token, "login": "alice.pop", "password": "owner-pass-1"})
	if code != http.StatusCreated || session["token"] == "" {
		t.Fatalf("activate: %d %v", code, session)
	}
	if code, _ := f.as("", false, "POST", "/activate", map[string]string{"token": token, "login": "alice2", "password": "owner-pass-1"}); code != 404 {
		t.Fatalf("link reused: %d", code)
	}
	if code, status := f.root("GET", f.owner(f.alice, "/account"), nil); code != 200 || status["status"] != "active" || status["login"] != "alice.pop" {
		t.Fatalf("status active: %d %v", code, status)
	}
	if code, _ := f.root("POST", f.owner(f.alice, "/activation"), nil); code != http.StatusConflict {
		t.Fatalf("link for an owner with an account: %d", code)
	}

	// Password-only sign-in for the owner account.
	if code, out := f.as("", false, "POST", "/login", map[string]string{"login": "alice.pop", "password": "owner-pass-1"}); code != 200 || out["token"] == "" {
		t.Fatalf("owner login: %d %v", code, out)
	}

	// Their own profile; no way into association routes or staff roles.
	code, me := f.as("alice.pop", false, "GET", "/me", nil)
	owners, _ := me["owners"].([]any)
	if code != 200 || me["account_type"] != "owner" || me["manages_associations"] != false || len(owners) != 1 {
		t.Fatalf("me: %d %v", code, me)
	}
	if code, _ := f.as("alice.pop", false, "GET", f.owner(f.alice, ""), nil); code != http.StatusForbidden {
		t.Fatalf("owner reached an association route: %d", code)
	}
	var userID int64
	f.cfg.DB.QueryRow(`SELECT id FROM users WHERE login = 'alice.pop'`).Scan(&userID)
	if code, _ := f.root("PUT", fmt.Sprintf("/admin/users/%d/role", userID), map[string]bool{"isAdmin": true}); code != http.StatusConflict {
		t.Fatalf("owner made admin: %d", code)
	}
}

func TestOwnerProfileTotpAndPassword(t *testing.T) {
	f := newOwnerAccountsFixture(t)
	f.activate(f.alice, "alice.pop")
	me := func(method, path string, body any) (int, map[string]any) {
		return f.as("alice.pop", false, method, path, body)
	}
	login := func(password, code string) int {
		c, _ := f.as("", false, "POST", "/login", map[string]string{"login": "alice.pop", "password": password, "totp": code})
		return c
	}

	if code, _ := me("PUT", "/me/profile", map[string]string{"contact_phone": "0700", "contact_email": "not-an-email"}); code != 400 {
		t.Fatalf("bad email: %d", code)
	}
	if code, out := me("PUT", "/me/profile", map[string]string{"contact_phone": " 0700 111 ", "contact_email": "alice@x.ro"}); code != 200 || out["owners"].([]any)[0].(map[string]any)["contact_phone"] != "0700 111" {
		t.Fatalf("profile: %d %v", code, out)
	}
	var phone string
	f.cfg.DB.QueryRow(`SELECT contact_phone FROM owners WHERE id = ?`, f.alice).Scan(&phone)
	if phone != "0700 111" {
		t.Fatalf("owner record not updated: %q", phone)
	}

	// Optional TOTP: once confirmed, sign-in needs the code; disabling needs the password.
	code, enrollment := me("POST", "/me/totp", nil)
	secret, _ := enrollment["secret"].(string)
	if code != 200 || secret == "" || login("owner-pass-1", "") != 200 {
		t.Fatalf("start totp (not yet required): %d", code)
	}
	if code, _ := me("POST", "/me/totp/confirm", map[string]string{"totp": "000000"}); code != 400 {
		t.Fatalf("wrong code accepted: %d", code)
	}
	now, _ := totp.GenerateCode(secret, time.Now())
	if code, _ := me("POST", "/me/totp/confirm", map[string]string{"totp": now}); code != 200 {
		t.Fatalf("confirm: %d", code)
	}
	if login("owner-pass-1", "") != 401 || login("owner-pass-1", now) != 200 {
		t.Fatal("TOTP not enforced after enabling")
	}
	if code, _ := me("DELETE", "/me/totp", map[string]string{"password": "wrong"}); code != 403 {
		t.Fatalf("disable with a wrong password: %d", code)
	}
	if code, _ := me("DELETE", "/me/totp", map[string]string{"password": "owner-pass-1"}); code != 200 || login("owner-pass-1", "") != 200 {
		t.Fatalf("disable: %d", code)
	}
	if code, _ := f.as("ana", false, "POST", "/me/totp", nil); code != http.StatusConflict {
		t.Fatalf("staff turned TOTP optional: %d", code)
	}

	// Password change.
	if code, _ := me("PUT", "/me/password", map[string]string{"current_password": "nope", "new_password": "another-pass-2"}); code != 403 {
		t.Fatalf("wrong current password: %d", code)
	}
	if code, _ := me("PUT", "/me/password", map[string]string{"current_password": "owner-pass-1", "new_password": "another-pass-2"}); code != 200 || login("another-pass-2", "") != 200 {
		t.Fatalf("change password: %d", code)
	}
}

func TestOwnerActivationsBulkLinkRevokeUnlink(t *testing.T) {
	f := newOwnerAccountsFixture(t)
	f.activate(f.alice, "alice.pop")

	code, bulk := f.root("POST", fmt.Sprintf("/associations/%d/owners/activations", f.assocA), nil)
	created, _ := bulk["created"].([]any)
	if code != http.StatusCreated || len(created) != 1 || bulk["skipped"] != float64(1) {
		t.Fatalf("bulk: %d %v", code, bulk)
	}
	if code, _ := f.root("DELETE", f.owner(f.bob, "/activation"), nil); code != 200 {
		t.Fatalf("revoke: %d", code)
	}
	if code, _ := f.root("DELETE", f.owner(f.bob, "/activation"), nil); code != 404 {
		t.Fatalf("revoke twice: %d", code)
	}
	if code, _ := f.root("GET", fmt.Sprintf("/associations/%d/owners/accounts", f.assocA), nil); code != 200 {
		t.Fatalf("list: %d", code)
	}

	// A second association's owner record on the same account.
	code, zedLink := f.root("POST", fmt.Sprintf("/associations/%d/owners/%d/activation", f.assocB, f.zed), nil)
	if code != http.StatusCreated {
		t.Fatalf("zed link: %d", code)
	}
	code, me := f.as("alice.pop", false, "POST", "/me/owners", map[string]string{"token": zedLink["token"].(string)})
	if code != 200 || len(me["owners"].([]any)) != 2 {
		t.Fatalf("link second owner: %d %v", code, me)
	}

	// Unlinking the last owner disables the owner account.
	f.root("DELETE", f.owner(f.alice, "/account"), nil)
	if code, _ := f.root("DELETE", fmt.Sprintf("/associations/%d/owners/%d/account", f.assocB, f.zed), nil); code != 200 {
		t.Fatalf("unlink: %d", code)
	}
	if code, _ := f.as("", false, "POST", "/login", map[string]string{"login": "alice.pop", "password": "owner-pass-1"}); code != 401 {
		t.Fatalf("account without owners can still sign in: %d", code)
	}
}

func TestOwnerPhoto(t *testing.T) {
	f := newOwnerAccountsFixture(t)
	f.activate(f.alice, "alice.pop")

	img := image.NewRGBA(image.Rect(0, 0, 600, 400))
	for x := 0; x < 600; x++ {
		for y := 0; y < 400; y++ {
			img.Set(x, y, color.RGBA{uint8(x), uint8(y), 120, 255})
		}
	}
	var pngData bytes.Buffer
	png.Encode(&pngData, img)
	upload := func(data []byte) (int, map[string]any) {
		var body bytes.Buffer
		w := multipart.NewWriter(&body)
		part, _ := w.CreateFormFile("photo", "me.png")
		part.Write(data)
		w.Close()
		req := httptest.NewRequest("PUT", "/me/photo", &body)
		req.Header.Set("Content-Type", w.FormDataContentType())
		token, _, _ := auth.MakeJWT("alice.pop", f.cfg.Secret, time.Hour, false)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		f.mux.ServeHTTP(rec, req)
		var out map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &out)
		return rec.Code, out
	}

	if code, _ := upload([]byte("not an image")); code != 400 {
		t.Fatalf("not an image: %d", code)
	}
	code, me := upload(pngData.Bytes())
	url, _ := me["photo_url"].(string)
	if code != 200 || !strings.HasPrefix(url, "/v1/api/photos/") {
		t.Fatalf("upload: %d %v", code, me)
	}
	rec := httptest.NewRecorder()
	f.mux.ServeHTTP(rec, httptest.NewRequest("GET", url, nil))
	decoded, err := jpeg.Decode(rec.Body)
	if rec.Code != 200 || err != nil || decoded.Bounds().Dx() != photoSize || decoded.Bounds().Dy() != photoSize {
		t.Fatalf("serve: %d %v", rec.Code, err)
	}
	for name, bad := range map[string]string{
		"tampered": strings.Replace(url, "sig=", "sig=0", 1),
		"expired":  strings.Split(url, "&exp=")[0] + "&exp=1&sig=" + photoSignature(f.cfg, 0, 1),
	} {
		rec := httptest.NewRecorder()
		f.mux.ServeHTTP(rec, httptest.NewRequest("GET", bad, nil))
		if rec.Code != http.StatusForbidden {
			t.Errorf("%s: %d", name, rec.Code)
		}
	}
	// Staff see it next to the owner; the account being disabled hides it.
	if _, owner := f.root("GET", f.owner(f.alice, ""), nil); owner["photo_url"] == nil || owner["photo_url"] == "" {
		t.Fatalf("owner detail without photo: %v", owner)
	}
	if _, owner := f.root("GET", f.owner(f.bob, ""), nil); owner["photo_url"] != nil {
		t.Fatalf("bob has alice's photo: %v", owner)
	}
	f.cfg.DB.Exec(`UPDATE users SET disabled_at = CURRENT_TIMESTAMP WHERE login = 'alice.pop'`)
	if _, owner := f.root("GET", f.owner(f.alice, ""), nil); owner["photo_url"] != nil {
		t.Fatalf("disabled account's photo shown: %v", owner)
	}
	f.cfg.DB.Exec(`UPDATE users SET disabled_at = NULL WHERE login = 'alice.pop'`)

	if code, me := f.as("alice.pop", false, "DELETE", "/me/photo", nil); code != 200 || me["photo_url"] != "" {
		t.Fatalf("delete: %d %v", code, me)
	}
	if _, owner := f.root("GET", f.owner(f.alice, ""), nil); owner["photo_url"] != nil {
		t.Fatalf("deleted photo still shown: %v", owner)
	}
}

type myProperties []struct {
	OwnerID         int64  `json:"owner_id"`
	AssociationName string `json:"association_name"`
	Units           []struct {
		UnitID       int64    `json:"unit_id"`
		BuildingName string   `json:"building_name"`
		IsVoting     bool     `json:"is_voting"`
		Area         float64  `json:"area"`
		CoOwners     []string `json:"co_owners"`
	} `json:"units"`
}

func TestMyProperties(t *testing.T) {
	f := newOwnerAccountsFixture(t)
	f.activate(f.alice, "alice-owner")
	get := func(login string) myProperties {
		t.Helper()
		req := httptest.NewRequest("GET", "/me/properties", nil)
		token, _, _ := auth.MakeJWT(login, f.cfg.Secret, time.Hour, false)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		f.mux.ServeHTTP(rec, req)
		var out myProperties
		if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil || rec.Code != 200 {
			t.Fatalf("properties: %d %s", rec.Code, rec.Body.String())
		}
		return out
	}
	props := get("alice-owner")
	if len(props) != 1 || props[0].OwnerID != f.alice || props[0].AssociationName != "A" || len(props[0].Units) != 1 {
		t.Fatalf("alice: %+v", props)
	}
	u := props[0].Units[0]
	if u.UnitID != f.unitA || u.BuildingName != "Bloc A" || !u.IsVoting || u.Area != 50 || len(u.CoOwners) != 1 || u.CoOwners[0] != "Bob" {
		t.Fatalf("alice's unit: %+v", u)
	}
	// A staff account without owner records has nothing.
	if props := get("ana"); len(props) != 0 {
		t.Fatalf("staff: %+v", props)
	}
}
