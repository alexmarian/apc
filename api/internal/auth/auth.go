package auth

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"image/png"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

type ApcClaims struct {
	jwt.RegisteredClaims
	IsAdmin bool `json:"admin"`
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// MakeJWT returns the signed token string and the jti (for revocation).
func MakeJWT(userLogin, tokenSecret string, expiresIn time.Duration, isAdmin bool) (string, string, error) {
	jti := uuid.NewString()
	claims := ApcClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "APC",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
			Subject:   userLogin,
			ID:        jti,
		},
		IsAdmin: isAdmin,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", "", err
	}
	return signed, jti, nil
}

func ValidateJWT(tokenString, tokenSecret string) (*ApcClaims, error) {
	claims := &ApcClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("token is invalid")
	}
	return claims, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	parts, err := authHeaderParts(headers)
	if err != nil {
		return "", err
	}
	if parts[0] != "Bearer" {
		return "", errors.New("authorization header is invalid")
	}
	return parts[1], nil
}

func GetApiKey(headers http.Header) (string, error) {
	parts, err := authHeaderParts(headers)
	if err != nil {
		return "", err
	}
	if parts[0] != "ApiKey" {
		return "", errors.New("authorization header is invalid")
	}
	return parts[1], nil
}

func MakeRefreshToken() (string, error) {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}
	return hex.EncodeToString(token), nil
}

func GenerateQRCode(username string) (string, string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "APC",
		AccountName: username,
		Period:      30,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		return "", "", err
	}
	img, err := key.Image(300, 300)
	if err != nil {
		return "", "", err
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", "", err
	}
	return key.Secret(), base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func VerifyTOTPCode(secret, code string) (bool, error) {
	return totp.ValidateCustom(code, secret, time.Now().UTC(), totp.ValidateOpts{
		Period:    30,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
}

func authHeaderParts(headers http.Header) ([]string, error) {
	h := headers.Get("Authorization")
	if h == "" {
		return nil, errors.New("authorization header is missing")
	}
	parts := strings.Split(h, " ")
	if len(parts) != 2 {
		return nil, errors.New("authorization header is invalid")
	}
	return parts, nil
}
