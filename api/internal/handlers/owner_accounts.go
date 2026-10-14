package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/alexmarian/apc/api/internal/auth"
	"github.com/alexmarian/apc/api/internal/database"
	"github.com/alexmarian/apc/api/internal/logging"
	"go.uber.org/zap"
)

// Account types (users.account_type).
const (
	AccountTypeStaff = "staff"
	AccountTypeOwner = "owner"
)

// ActivationTokenPathValue is the public activation routes' path value.
const ActivationTokenPathValue = "activationToken"

// defaultActivationTTL: how long an activation link stays valid.
const defaultActivationTTL = 30 * 24 * time.Hour

// newActivationToken returns a random URL-safe token and its SHA-256 hash
// (only the hash is stored; the token is shown once).
func newActivationToken() (string, string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	return token, hashActivationToken(token), nil
}

func hashActivationToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func activationUsable(a database.OwnerActivation, now time.Time) bool {
	return !a.UsedAt.Valid && !a.RevokedAt.Valid && a.ExpiresAt.After(now)
}

// --- staff: an owner's app account ------------------------------------------

type ownerAccountResponse struct {
	OwnerID         int64      `json:"owner_id"`
	Status          string     `json:"status"` // none | invited | active | disabled
	Login           string     `json:"login,omitempty"`
	InviteExpiresAt *time.Time `json:"invite_expires_at,omitempty"`
	HasPhoto        bool       `json:"has_photo"`
}

type createdActivationResponse struct {
	OwnerID   int64     `json:"owner_id"`
	OwnerName string    `json:"owner_name"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ownerAccounts returns every owner's account status in the association.
func ownerAccounts(cfg *ApiConfig, req *http.Request, associationId int64) (map[int64]ownerAccountResponse, error) {
	ctx := req.Context()
	accounts, err := cfg.Db.ListAssociationOwnerAccounts(ctx, associationId)
	if err != nil {
		return nil, err
	}
	pending, err := cfg.Db.ListPendingOwnerActivations(ctx, associationId)
	if err != nil {
		return nil, err
	}
	out := map[int64]ownerAccountResponse{}
	now := time.Now()
	for _, a := range pending {
		if !activationUsable(a, now) {
			continue
		}
		expires := a.ExpiresAt
		out[a.OwnerID] = ownerAccountResponse{OwnerID: a.OwnerID, Status: "invited", InviteExpiresAt: &expires}
	}
	for _, a := range accounts {
		status := "active"
		if a.DisabledAt.Valid {
			status = "disabled"
		}
		out[a.OwnerID] = ownerAccountResponse{OwnerID: a.OwnerID, Status: status, Login: a.Login, HasPhoto: a.PhotoUpdatedAt.Valid}
	}
	return out, nil
}

// HandleListOwnerAccounts: GET /associations/{aid}/owners/accounts — the
// owners that have an account or a pending link (others have none).
func HandleListOwnerAccounts(cfg *ApiConfig) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := pathInt64(req, AssociationIdPathValue)
		accounts, err := ownerAccounts(cfg, req, associationId)
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error listing accounts")
			return
		}
		list := make([]ownerAccountResponse, 0, len(accounts))
		for _, a := range accounts {
			list = append(list, a)
		}
		RespondWithJSON(rw, http.StatusOK, list)
	}
}

// HandleGetOwnerAccount: GET /associations/{aid}/owners/{ownerId}/account.
func HandleGetOwnerAccount(cfg *ApiConfig) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		owner, ok := associationOwner(cfg, rw, req)
		if !ok {
			return
		}
		accounts, err := ownerAccounts(cfg, req, owner.AssociationID)
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error getting account")
			return
		}
		account, found := accounts[owner.ID]
		if !found {
			account = ownerAccountResponse{OwnerID: owner.ID, Status: "none"}
		}
		RespondWithJSON(rw, http.StatusOK, account)
	}
}

// createActivation makes a link for owner unless they have an account or a
// usable link already (errAccountExists / errActivationPending).
var (
	errAccountExists     = errors.New("the owner already has an app account")
	errActivationPending = errors.New("the owner already has an activation link; revoke it to create a new one")
)

func createActivation(cfg *ApiConfig, req *http.Request, owner database.Owner, expiresAt time.Time) (createdActivationResponse, error) {
	ctx := req.Context()
	if _, err := cfg.Db.GetOwnerAccount(ctx, owner.ID); err == nil {
		return createdActivationResponse{}, errAccountExists
	} else if !errors.Is(err, sql.ErrNoRows) {
		return createdActivationResponse{}, err
	}
	pending, err := cfg.Db.ListPendingOwnerActivations(ctx, owner.AssociationID)
	if err != nil {
		return createdActivationResponse{}, err
	}
	for _, a := range pending {
		if a.OwnerID == owner.ID && activationUsable(a, time.Now()) {
			return createdActivationResponse{}, errActivationPending
		}
	}
	token, hash, err := newActivationToken()
	if err != nil {
		return createdActivationResponse{}, err
	}
	created, err := cfg.Db.CreateOwnerActivation(ctx, database.CreateOwnerActivationParams{
		OwnerID:   owner.ID,
		TokenHash: hash,
		ExpiresAt: expiresAt,
		CreatedBy: GetUserIdFromContext(req),
	})
	if err != nil {
		return createdActivationResponse{}, err
	}
	return createdActivationResponse{OwnerID: owner.ID, OwnerName: owner.Name, Token: token, ExpiresAt: created.ExpiresAt}, nil
}

// activationExpiry reads an optional {expires_at} body (default 30 days;
// must be in the future).
func activationExpiry(rw http.ResponseWriter, req *http.Request) (time.Time, bool) {
	var body struct {
		ExpiresAt *time.Time `json:"expires_at"`
	}
	if req.Body != nil && req.ContentLength != 0 {
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return time.Time{}, false
		}
	}
	expires := time.Now().Add(defaultActivationTTL)
	if body.ExpiresAt != nil {
		if !body.ExpiresAt.After(time.Now()) {
			RespondWithError(rw, http.StatusBadRequest, "The expiry must be in the future")
			return time.Time{}, false
		}
		expires = *body.ExpiresAt
	}
	return expires, true
}

// HandleCreateOwnerActivation: POST /associations/{aid}/owners/{ownerId}/activation
// — a one-time link the owner opens to create their account (token shown once).
func HandleCreateOwnerActivation(cfg *ApiConfig) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		owner, ok := associationOwner(cfg, rw, req)
		if !ok {
			return
		}
		expires, ok := activationExpiry(rw, req)
		if !ok {
			return
		}
		created, err := createActivation(cfg, req, owner, expires)
		switch {
		case errors.Is(err, errAccountExists), errors.Is(err, errActivationPending):
			RespondWithError(rw, http.StatusConflict, err.Error())
		case err != nil:
			logging.Logger.Log(zap.WarnLevel, "Error creating activation", zap.Error(err))
			RespondWithError(rw, http.StatusInternalServerError, "Error creating activation link")
		default:
			RespondWithJSON(rw, http.StatusCreated, created)
		}
	}
}

// HandleBulkOwnerActivations: POST /associations/{aid}/owners/activations —
// links for every owner without an account or a pending link.
func HandleBulkOwnerActivations(cfg *ApiConfig) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := pathInt64(req, AssociationIdPathValue)
		expires, ok := activationExpiry(rw, req)
		if !ok {
			return
		}
		owners, err := cfg.Db.GetAssociationOwners(req.Context(), associationId)
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error listing owners")
			return
		}
		created := []createdActivationResponse{}
		skipped := 0
		for _, o := range owners {
			c, err := createActivation(cfg, req, o, expires)
			if errors.Is(err, errAccountExists) || errors.Is(err, errActivationPending) {
				skipped++
				continue
			}
			if err != nil {
				logging.Logger.Log(zap.WarnLevel, "Error creating activation", zap.Error(err))
				RespondWithError(rw, http.StatusInternalServerError, "Error creating activation links")
				return
			}
			created = append(created, c)
		}
		RespondWithJSON(rw, http.StatusCreated, map[string]any{"created": created, "skipped": skipped})
	}
}

// HandleRevokeOwnerActivation: DELETE /associations/{aid}/owners/{ownerId}/activation.
func HandleRevokeOwnerActivation(cfg *ApiConfig) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		owner, ok := associationOwner(cfg, rw, req)
		if !ok {
			return
		}
		n, err := cfg.Db.RevokeOwnerActivations(req.Context(), owner.ID)
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error revoking activation link")
			return
		}
		if n == 0 {
			RespondWithError(rw, http.StatusNotFound, "No activation link to revoke")
			return
		}
		RespondWithJSON(rw, http.StatusOK, map[string]string{"status": "revoked"})
	}
}

// HandleUnlinkOwnerAccount: DELETE /associations/{aid}/owners/{ownerId}/account
// — the account no longer represents this owner; an owner account left with
// no owners is disabled (it would only show an empty portal).
func HandleUnlinkOwnerAccount(cfg *ApiConfig) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		owner, ok := associationOwner(cfg, rw, req)
		if !ok {
			return
		}
		ctx := req.Context()
		user, err := cfg.Db.GetOwnerAccount(ctx, owner.ID)
		if errors.Is(err, sql.ErrNoRows) {
			RespondWithError(rw, http.StatusNotFound, "The owner has no app account")
			return
		}
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error unlinking account")
			return
		}
		tx, err := cfg.DB.BeginTx(ctx, nil)
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error unlinking account")
			return
		}
		defer tx.Rollback()
		q := cfg.Db.WithTx(tx)
		if _, err := q.UnlinkOwner(ctx, owner.ID); err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error unlinking account")
			return
		}
		left, err := q.CountUserOwners(ctx, user.ID)
		if err == nil && left == 0 && user.AccountType == AccountTypeOwner {
			err = q.SetUserDisabled(ctx, database.SetUserDisabledParams{
				DisabledAt: sql.NullTime{Time: time.Now(), Valid: true},
				DisabledBy: sql.NullString{String: GetUserIdFromContext(req), Valid: true},
				ID:         user.ID,
			})
		}
		if err != nil || tx.Commit() != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error unlinking account")
			return
		}
		RespondWithJSON(rw, http.StatusOK, map[string]string{"status": "unlinked"})
	}
}

// --- public: activation -------------------------------------------------------

// usableActivation looks a token up (404 when unknown, used, revoked or expired).
func usableActivation(cfg *ApiConfig, rw http.ResponseWriter, req *http.Request, q *database.Queries, token string) (database.OwnerActivation, database.Owner, bool) {
	a, err := q.GetOwnerActivationByHash(req.Context(), hashActivationToken(token))
	if err != nil || !activationUsable(a, time.Now()) {
		RespondWithError(rw, http.StatusNotFound, "This activation link is invalid or has expired")
		return a, database.Owner{}, false
	}
	owner, err := q.GetOwnerById(req.Context(), a.OwnerID)
	if err != nil {
		RespondWithError(rw, http.StatusNotFound, "This activation link is invalid or has expired")
		return a, database.Owner{}, false
	}
	return a, owner, true
}

// HandleGetActivation: GET /activate/{token} — who the link is for, to greet
// them before they choose a login.
func HandleGetActivation(cfg *ApiConfig) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		a, owner, ok := usableActivation(cfg, rw, req, cfg.Db, req.PathValue(ActivationTokenPathValue))
		if !ok {
			return
		}
		association, err := cfg.Db.GetAssociations(req.Context(), owner.AssociationID)
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error reading activation link")
			return
		}
		RespondWithJSON(rw, http.StatusOK, map[string]any{
			"owner_name":       owner.Name,
			"association_name": association.Name,
			"expires_at":       a.ExpiresAt,
		})
	}
}

// HandleActivate: POST /activate {token, login, password} — creates an owner
// account linked to the link's owner, uses the link, and signs in.
func HandleActivate(cfg *ApiConfig) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		var body struct {
			Token    string `json:"token"`
			Login    string `json:"login"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}
		body.Login = strings.TrimSpace(body.Login)
		if msg := validateCredentials(body.Login, body.Password); msg != "" {
			RespondWithError(rw, http.StatusBadRequest, msg)
			return
		}
		hash, err := auth.HashPassword(body.Password)
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error creating account")
			return
		}
		ctx := req.Context()
		tx, err := cfg.DB.BeginTx(ctx, nil)
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error creating account")
			return
		}
		defer tx.Rollback()
		q := cfg.Db.WithTx(tx)
		a, owner, ok := usableActivation(cfg, rw, req, q, body.Token)
		if !ok {
			return
		}
		if _, err := q.GetOwnerAccount(ctx, owner.ID); err == nil {
			RespondWithError(rw, http.StatusConflict, errAccountExists.Error())
			return
		}
		if _, err := q.GetUserByLogin(ctx, body.Login); err == nil {
			RespondWithError(rw, http.StatusConflict, "Login already taken")
			return
		} else if !errors.Is(err, sql.ErrNoRows) {
			RespondWithError(rw, http.StatusInternalServerError, "Error creating account")
			return
		}
		user, err := q.CreateOwnerUser(ctx, database.CreateOwnerUserParams{Login: body.Login, PasswordHash: hash})
		if err == nil {
			err = q.LinkUserOwner(ctx, database.LinkUserOwnerParams{UserID: user.ID, OwnerID: owner.ID})
		}
		if err == nil {
			var used int64
			used, err = q.UseOwnerActivation(ctx, a.ID)
			if err == nil && used == 0 {
				err = sql.ErrNoRows
			}
		}
		if err != nil || tx.Commit() != nil {
			logging.Logger.Log(zap.WarnLevel, "Error activating owner account", zap.Error(err))
			RespondWithError(rw, http.StatusInternalServerError, "Error creating account")
			return
		}
		session, err := issueSession(ctx, cfg, user, time.Hour)
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error creating session")
			return
		}
		RespondWithJSON(rw, http.StatusCreated, session)
	}
}

// --- the signed-in user -------------------------------------------------------

func currentUser(cfg *ApiConfig, rw http.ResponseWriter, req *http.Request) (database.User, bool) {
	user, err := cfg.Db.GetUserByLogin(req.Context(), GetUserIdFromContext(req))
	if err != nil || user.DisabledAt.Valid {
		RespondWithError(rw, http.StatusUnauthorized, "Unauthorized")
		return user, false
	}
	return user, true
}

type meOwner struct {
	ID                   int64  `json:"id"`
	Name                 string `json:"name"`
	IdentificationNumber string `json:"identification_number"`
	ContactPhone         string `json:"contact_phone"`
	ContactEmail         string `json:"contact_email"`
	AssociationID        int64  `json:"association_id"`
	AssociationName      string `json:"association_name"`
}

// HandleGetMe: GET /me — who is signed in and what they can do (the app
// builds its navigation from this).
func HandleGetMe(cfg *ApiConfig) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		user, ok := currentUser(cfg, rw, req)
		if !ok {
			return
		}
		ctx := req.Context()
		rows, err := cfg.Db.ListUserOwners(ctx, user.ID)
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error reading profile")
			return
		}
		owners := make([]meOwner, len(rows))
		for i, r := range rows {
			owners[i] = meOwner{r.ID, r.Name, r.IdentificationNumber, r.ContactPhone, r.ContactEmail, r.AssociationID, r.AssociationName}
		}
		links, _ := cfg.Db.CountUserAssociationLinks(ctx, user.ID)
		RespondWithJSON(rw, http.StatusOK, map[string]any{
			"login":                user.Login,
			"account_type":         user.AccountType,
			"is_admin":             user.IsAdmin,
			"manages_associations": user.IsAdmin || links > 0,
			"owners":               owners,
			"photo_url":            PhotoURL(cfg, user),
			"totp_enabled":         user.TotpConfirmedAt.Valid,
		})
	}
}

// HandleUpdateMyProfile: PUT /me/profile {contact_phone, contact_email} —
// the owner's contact details, on every owner record they hold (staff see
// them there). Name and ID number stay with the association's admins.
func HandleUpdateMyProfile(cfg *ApiConfig) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		user, ok := currentUser(cfg, rw, req)
		if !ok {
			return
		}
		var body struct {
			ContactPhone string `json:"contact_phone"`
			ContactEmail string `json:"contact_email"`
		}
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}
		body.ContactPhone = strings.TrimSpace(body.ContactPhone)
		body.ContactEmail = strings.TrimSpace(body.ContactEmail)
		if utf8.RuneCountInString(body.ContactPhone) > 50 || utf8.RuneCountInString(body.ContactEmail) > 255 {
			RespondWithError(rw, http.StatusBadRequest, "Contact details are too long")
			return
		}
		if body.ContactEmail != "" {
			if _, err := mail.ParseAddress(body.ContactEmail); err != nil {
				RespondWithError(rw, http.StatusBadRequest, "Invalid email address")
				return
			}
		}
		ctx := req.Context()
		owners, err := cfg.Db.ListUserOwners(ctx, user.ID)
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error saving profile")
			return
		}
		if len(owners) == 0 {
			RespondWithError(rw, http.StatusConflict, "This account doesn't represent an owner")
			return
		}
		for _, o := range owners {
			if err := cfg.Db.UpdateOwnerContact(ctx, database.UpdateOwnerContactParams{ContactPhone: body.ContactPhone, ContactEmail: body.ContactEmail, ID: o.ID}); err != nil {
				RespondWithError(rw, http.StatusInternalServerError, "Error saving profile")
				return
			}
		}
		HandleGetMe(cfg)(rw, req)
	}
}

// HandleChangePassword: PUT /me/password {current_password, new_password}.
func HandleChangePassword(cfg *ApiConfig) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		user, ok := currentUser(cfg, rw, req)
		if !ok {
			return
		}
		var body struct {
			Current string `json:"current_password"`
			New     string `json:"new_password"`
		}
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}
		if auth.CheckPasswordHash(body.Current, user.PasswordHash) != nil {
			RespondWithError(rw, http.StatusForbidden, "The current password is wrong")
			return
		}
		if msg := validatePassword(body.New); msg != "" {
			RespondWithError(rw, http.StatusBadRequest, msg)
			return
		}
		hash, err := auth.HashPassword(body.New)
		if err == nil {
			err = cfg.Db.SetUserPassword(req.Context(), database.SetUserPasswordParams{PasswordHash: hash, ID: user.ID})
		}
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error changing password")
			return
		}
		RespondWithJSON(rw, http.StatusOK, map[string]string{"status": "changed"})
	}
}

// ownerAccount answers 409 for staff accounts (their TOTP is mandatory).
func ownerAccount(cfg *ApiConfig, rw http.ResponseWriter, req *http.Request) (database.User, bool) {
	user, ok := currentUser(cfg, rw, req)
	if ok && user.AccountType != AccountTypeOwner {
		RespondWithError(rw, http.StatusConflict, "Staff accounts always use an authenticator code")
		return user, false
	}
	return user, ok
}

// HandleStartMyTotp: POST /me/totp — a new authenticator secret to scan (not
// required at sign-in until confirmed).
func HandleStartMyTotp(cfg *ApiConfig) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		user, ok := ownerAccount(cfg, rw, req)
		if !ok {
			return
		}
		enrollment, err := auth.GenerateTOTPEnrollment(user.Login)
		if err == nil {
			err = cfg.Db.SetUserTotpSecret(req.Context(), database.SetUserTotpSecretParams{ToptSecret: enrollment.Secret, ID: user.ID})
		}
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error starting authenticator setup")
			return
		}
		RespondWithJSON(rw, http.StatusOK, toEnrollmentResponse(enrollment))
	}
}

// HandleConfirmMyTotp: POST /me/totp/confirm {totp} — from now on sign-in
// needs the code.
func HandleConfirmMyTotp(cfg *ApiConfig) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		user, ok := ownerAccount(cfg, rw, req)
		if !ok {
			return
		}
		var body struct {
			TOTP string `json:"totp"`
		}
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}
		if user.ToptSecret == "" {
			RespondWithError(rw, http.StatusConflict, "Start the authenticator setup first")
			return
		}
		if ok, err := auth.VerifyTOTPCode(user.ToptSecret, body.TOTP); err != nil || !ok {
			RespondWithError(rw, http.StatusBadRequest, "The code is not correct")
			return
		}
		if err := cfg.Db.ConfirmUserTotp(req.Context(), user.ID); err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error enabling the authenticator")
			return
		}
		RespondWithJSON(rw, http.StatusOK, map[string]string{"status": "enabled"})
	}
}

// HandleDisableMyTotp: DELETE /me/totp {password} — back to password-only.
func HandleDisableMyTotp(cfg *ApiConfig) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		user, ok := ownerAccount(cfg, rw, req)
		if !ok {
			return
		}
		var body struct {
			Password string `json:"password"`
		}
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}
		if auth.CheckPasswordHash(body.Password, user.PasswordHash) != nil {
			RespondWithError(rw, http.StatusForbidden, "The password is wrong")
			return
		}
		if err := cfg.Db.SetUserTotpSecret(req.Context(), database.SetUserTotpSecretParams{ToptSecret: "", ID: user.ID}); err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error disabling the authenticator")
			return
		}
		RespondWithJSON(rw, http.StatusOK, map[string]string{"status": "disabled"})
	}
}

// HandleLinkMyOwner: POST /me/owners {token} — an activation link from
// another association, added to the signed-in account.
func HandleLinkMyOwner(cfg *ApiConfig) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		user, ok := currentUser(cfg, rw, req)
		if !ok {
			return
		}
		var body struct {
			Token string `json:"token"`
		}
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}
		ctx := req.Context()
		tx, err := cfg.DB.BeginTx(ctx, nil)
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error linking owner")
			return
		}
		defer tx.Rollback()
		q := cfg.Db.WithTx(tx)
		a, owner, ok := usableActivation(cfg, rw, req, q, strings.TrimSpace(body.Token))
		if !ok {
			return
		}
		if _, err := q.GetOwnerAccount(ctx, owner.ID); err == nil {
			RespondWithError(rw, http.StatusConflict, errAccountExists.Error())
			return
		}
		err = q.LinkUserOwner(ctx, database.LinkUserOwnerParams{UserID: user.ID, OwnerID: owner.ID})
		if err == nil {
			_, err = q.UseOwnerActivation(ctx, a.ID)
		}
		if err != nil || tx.Commit() != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error linking owner")
			return
		}
		HandleGetMe(cfg)(rw, req)
	}
}

// userIDParam parses a positive int64 (shared by the photo routes).
func userIDParam(s string) (int64, bool) {
	id, err := strconv.ParseInt(s, 10, 64)
	return id, err == nil && id > 0
}

// --- the owner's properties ---------------------------------------------------

type myUnit struct {
	OwnershipID     int64     `json:"ownership_id"`
	IsVoting        bool      `json:"is_voting"`
	Since           time.Time `json:"since"`
	UnitID          int64     `json:"unit_id"`
	UnitNumber      string    `json:"unit_number"`
	UnitType        string    `json:"unit_type"`
	Floor           int64     `json:"floor"`
	Entrance        int64     `json:"entrance"`
	Area            float64   `json:"area"`
	Part            float64   `json:"part"`
	Address         string    `json:"address"`
	BuildingID      int64     `json:"building_id"`
	BuildingName    string    `json:"building_name"`
	BuildingAddress string    `json:"building_address"`
	CoOwners        []string  `json:"co_owners"`
}

// HandleGetMyProperties: GET /me/properties — per owner record (association),
// the units currently owned, with co-owners.
func HandleGetMyProperties(cfg *ApiConfig) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		user, ok := currentUser(cfg, rw, req)
		if !ok {
			return
		}
		ctx := req.Context()
		owners, err := cfg.Db.ListUserOwners(ctx, user.ID)
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error reading properties")
			return
		}
		type group struct {
			OwnerID         int64    `json:"owner_id"`
			AssociationID   int64    `json:"association_id"`
			AssociationName string   `json:"association_name"`
			Units           []myUnit `json:"units"`
		}
		groups := make([]group, 0, len(owners))
		for _, o := range owners {
			rows, err := cfg.Db.ListOwnerActiveUnits(ctx, o.ID)
			if err != nil {
				RespondWithError(rw, http.StatusInternalServerError, "Error reading properties")
				return
			}
			units := make([]myUnit, len(rows))
			ids := make([]int64, len(rows))
			for i, r := range rows {
				units[i] = myUnit{r.OwnershipID, r.IsVoting, r.StartDate.Time, r.UnitID, r.UnitNumber, r.UnitType, r.Floor, r.Entrance,
					r.Area, r.Part, r.Address, r.BuildingID, r.BuildingName, r.BuildingAddress, []string{}}
				ids[i] = r.UnitID
			}
			if len(ids) > 0 {
				co, err := cfg.Db.ListUnitsCoOwners(ctx, database.ListUnitsCoOwnersParams{OwnerID: o.ID, UnitIds: ids})
				if err != nil {
					RespondWithError(rw, http.StatusInternalServerError, "Error reading properties")
					return
				}
				for _, c := range co {
					for i := range units {
						if units[i].UnitID == c.UnitID {
							units[i].CoOwners = append(units[i].CoOwners, c.Name)
						}
					}
				}
			}
			groups = append(groups, group{o.ID, o.AssociationID, o.AssociationName, units})
		}
		RespondWithJSON(rw, http.StatusOK, groups)
	}
}

// MyOwners returns the signed-in user's owner records (for the voting routes).
func MyOwners(cfg *ApiConfig, req *http.Request) ([]database.ListUserOwnersRow, bool) {
	user, err := cfg.Db.GetUserByLogin(req.Context(), GetUserIdFromContext(req))
	if err != nil || user.DisabledAt.Valid {
		return nil, false
	}
	owners, err := cfg.Db.ListUserOwners(req.Context(), user.ID)
	return owners, err == nil
}
