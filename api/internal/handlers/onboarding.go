package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"time"
	"unicode/utf8"

	"github.com/alexmarian/apc/api/internal/auth"
	"github.com/alexmarian/apc/api/internal/database"
	"github.com/alexmarian/apc/api/internal/logging"
	"go.uber.org/zap"
)

// A user goes through onboarding twice at most: after registering with an
// invitation token, and after a password reset that also reset their TOTP.
// Until they confirm a code from their authenticator (totp_confirmed_at),
// login answers 403 OnboardingRequiredMessage and they finish here.

var loginPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{4,64}$`)

const (
	passwordMinLength = 8
	// bcrypt ignores (and the Go implementation rejects) anything longer.
	passwordMaxBytes = 72
)

// validateCredentials checks a new login/password pair, returning a message
// for a 400, or "".
func validateCredentials(login, password string) string {
	if !loginPattern.MatchString(login) {
		return "Login must be 4-64 characters: letters, digits, '.', '_' or '-'"
	}
	return validatePassword(password)
}

func validatePassword(password string) string {
	if utf8.RuneCountInString(password) < passwordMinLength {
		return "Password must be at least 8 characters"
	}
	if len(password) > passwordMaxBytes {
		return "Password is too long"
	}
	return ""
}

// totpEnrollmentResponse is what a client shows to set up an authenticator.
type totpEnrollmentResponse struct {
	QRCode     string `json:"qrCode"`
	Secret     string `json:"secret"`
	OtpauthURL string `json:"otpauthUrl"`
}

func toEnrollmentResponse(e auth.TOTPEnrollment) totpEnrollmentResponse {
	return totpEnrollmentResponse{QRCode: e.QRCode, Secret: e.Secret, OtpauthURL: e.OtpauthURL}
}

var errLoginTaken = errors.New("login taken")

// HandleCreateUserWithToken registers a user with an invitation token
// (POST /users). In one transaction it claims the token (so it can't be
// used twice), creates the user as pending, and links them to the token's
// associations. Responds 201 with the TOTP enrollment to confirm next.
func HandleCreateUserWithToken(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		var request struct {
			Login    string `json:"login"`
			Password string `json:"password"`
			Token    string `json:"token"`
		}
		if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}
		if msg := validateCredentials(request.Login, request.Password); msg != "" {
			RespondWithError(rw, http.StatusBadRequest, msg)
			return
		}
		// Hash and generate outside the transaction — bcrypt is slow.
		passwordHash, err := auth.HashPassword(request.Password)
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error processing registration")
			return
		}
		enrollment, err := auth.GenerateTOTPEnrollment(request.Login)
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error processing registration")
			return
		}

		ctx := req.Context()
		tx, err := cfg.DB.BeginTx(ctx, nil)
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error processing registration")
			return
		}
		defer tx.Rollback()
		q := cfg.Db.WithTx(tx)

		user, err := func() (database.User, error) {
			token, err := q.GetValidRegistrationToken(ctx, request.Token)
			if err != nil {
				return database.User{}, err
			}
			if _, err := q.GetUserByLogin(ctx, request.Login); err == nil {
				return database.User{}, errLoginTaken
			} else if !errors.Is(err, sql.ErrNoRows) {
				return database.User{}, err
			}
			claimed, err := q.UseRegistrationToken(ctx, database.UseRegistrationTokenParams{
				UsedBy: sql.NullString{String: request.Login, Valid: true},
				Token:  request.Token,
			})
			if err != nil {
				return database.User{}, err
			}
			if claimed == 0 {
				return database.User{}, sql.ErrNoRows
			}
			user, err := q.CreateUser(ctx, database.CreateUserParams{
				Login:        request.Login,
				PasswordHash: passwordHash,
				ToptSecret:   enrollment.Secret,
				IsAdmin:      token.IsAdmin,
			})
			if err != nil {
				return database.User{}, err
			}
			if !token.IsAdmin {
				ids, err := q.GetRegistrationTokenAssociationIds(ctx, request.Token)
				if err != nil {
					return database.User{}, err
				}
				for _, id := range ids {
					if err := q.AddAssociationUser(ctx, database.AddAssociationUserParams{UserID: user.ID, AssociationID: id}); err != nil {
						return database.User{}, err
					}
				}
			}
			return user, tx.Commit()
		}()
		switch {
		case errors.Is(err, sql.ErrNoRows):
			RespondWithError(rw, http.StatusUnauthorized, "Invalid or expired registration token")
			return
		case errors.Is(err, errLoginTaken):
			RespondWithError(rw, http.StatusConflict, "Login already taken")
			return
		case err != nil:
			logging.Logger.Log(zap.WarnLevel, "Error registering user", zap.String("error", err.Error()))
			RespondWithError(rw, http.StatusInternalServerError, "Error processing registration")
			return
		}

		RespondWithJSON(rw, http.StatusCreated, struct {
			Login      string                 `json:"login"`
			IsAdmin    bool                   `json:"isAdmin"`
			CreatedAt  time.Time              `json:"createdAt"`
			Enrollment totpEnrollmentResponse `json:"enrollment"`
			// QrCode duplicates Enrollment.QRCode for the legacy Vue register page.
			QrCode string `json:"qrCode"`
		}{user.Login, user.IsAdmin, user.CreatedAt, toEnrollmentResponse(enrollment), enrollment.QRCode})
	}
}

// pendingUser authenticates login/password for an onboarding call and
// requires the user to still be pending. On failure it has already written
// the response: the generic 401 for unknown/wrong/disabled (as login does),
// 409 for an already-onboarded user.
func pendingUser(cfg *ApiConfig, rw http.ResponseWriter, req *http.Request, login, password string) (database.User, bool) {
	user, err := cfg.Db.GetUserByLogin(req.Context(), login)
	if err != nil || auth.CheckPasswordHash(password, user.PasswordHash) != nil || user.DisabledAt.Valid {
		RespondWithError(rw, http.StatusUnauthorized, "Login failure")
		return user, false
	}
	if user.TotpConfirmedAt.Valid {
		RespondWithError(rw, http.StatusConflict, "Two-factor authentication is already set up")
		return user, false
	}
	return user, true
}

// HandleStartOnboarding (POST /onboarding/totp) issues a fresh TOTP secret
// to a pending user — for resuming setup after the first QR was lost. The
// previous secret stops working.
func HandleStartOnboarding(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		var request struct {
			Login    string `json:"login"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}
		user, ok := pendingUser(cfg, rw, req, request.Login, request.Password)
		if !ok {
			return
		}
		enrollment, err := auth.GenerateTOTPEnrollment(user.Login)
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error generating two-factor secret")
			return
		}
		if err := cfg.Db.SetUserTotpSecret(req.Context(), database.SetUserTotpSecretParams{ToptSecret: enrollment.Secret, ID: user.ID}); err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error generating two-factor secret")
			return
		}
		RespondWithJSON(rw, http.StatusOK, struct {
			Login      string                 `json:"login"`
			Enrollment totpEnrollmentResponse `json:"enrollment"`
		}{user.Login, toEnrollmentResponse(enrollment)})
	}
}

// HandleConfirmOnboarding (POST /onboarding/totp/confirm) checks a code from
// the user's authenticator against their pending secret, marks TOTP as set
// up, and signs them in (same response as login).
func HandleConfirmOnboarding(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		var request struct {
			Login    string `json:"login"`
			Password string `json:"password"`
			Code     string `json:"code"`
		}
		if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}
		user, ok := pendingUser(cfg, rw, req, request.Login, request.Password)
		if !ok {
			return
		}
		if valid, err := auth.VerifyTOTPCode(user.ToptSecret, request.Code); err != nil || !valid {
			RespondWithError(rw, http.StatusBadRequest, "Invalid code")
			return
		}
		if err := cfg.Db.ConfirmUserTotp(req.Context(), user.ID); err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error completing setup")
			return
		}
		session, err := issueSession(req.Context(), cfg, user, time.Hour)
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error creating session")
			return
		}
		RespondWithJSON(rw, http.StatusOK, session)
	}
}
