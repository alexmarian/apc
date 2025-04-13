package auth

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pquerna/otp"
	"golang.org/x/crypto/bcrypt"
	"image/png"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pquerna/otp/totp"
)

type ApcClaims struct {
	jwt.RegisteredClaims
	Ass string `json:"ass"`
}

func HashPassword(password string) (string, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(passwordHash), nil
}
func CheckPasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return err
	}
	return nil
}
func MakeJWT(userLogin, tokenSecret string, expiresIn time.Duration, associations []int64) (string, error) {
	associations_str := make([]string, len(associations))
	for i, association := range associations {
		associations_str[i] = strconv.Itoa(int(association))
	}
	claims := ApcClaims{
		jwt.RegisteredClaims{

			Issuer:    "APC",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
			Subject:   userLogin,
		}, strings.Join(associations_str, ","),
	}
	jwt := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedString, err := jwt.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}
	return signedString, nil
}

func ValidateJWT(tokenString, tokenSecret string) (string, []int64, error) {
	claims := &ApcClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return "", nil, err
	}
	if token.Valid != true {
		return "", nil, errors.New("Token is invalid")
	}
	userLogin, err := token.Claims.GetSubject()
	if err != nil {
		return "", nil, err
	}
	asssociationsStrs := strings.Split(claims.Ass, ",")
	associations := make([]int64, len(asssociationsStrs))
	for i, association := range asssociationsStrs {
		associations[i], _ = strconv.ParseInt(association, 10, 64)
	}
	return userLogin, associations, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authorizationHeaderParts, s, err := getAuthorizationHeadersParts(headers)
	if err != nil {
		return s, err
	}
	if authorizationHeaderParts[0] != "Bearer" {
		return "", errors.New("Authorization header is invalid")
	}
	return authorizationHeaderParts[1], nil
}

func getAuthorizationHeadersParts(headers http.Header) ([]string, string, error) {
	authorizationHeader := headers.Get("Authorization")
	if authorizationHeader == "" {
		return nil, "", errors.New("Authorization header is missing")
	}
	authorizationHeaderParts := strings.Split(authorizationHeader, " ")
	if len(authorizationHeaderParts) != 2 {
		return nil, "", errors.New("Authorization header is invalid")
	}
	return authorizationHeaderParts, "", nil
}

func GetApiKey(headers http.Header) (string, error) {
	authorizationHeaderParts, s, err := getAuthorizationHeadersParts(headers)
	if err != nil {
		return s, err
	}
	if authorizationHeaderParts[0] != "ApiKey" {
		return "", errors.New("Authorization header is invalid")
	}
	return authorizationHeaderParts[1], nil
}

func MakeRefreshToken() (string, error) {
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
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

	image, err := key.Image(300, 300)
	if err != nil {
		return "", "", err
	}
	var buf bytes.Buffer
	err = png.Encode(&buf, image)
	if err != nil {
		return "", "", err
	}
	return key.Secret(), base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func VerifyTOTPCode(secret, code string) (bool, error) {
	success, err := totp.ValidateCustom(
		code, secret, time.Now().UTC(), totp.ValidateOpts{
			Period:    30,
			Digits:    otp.DigitsSix,
			Algorithm: otp.AlgorithmSHA1,
		},
	)
	return success, err
}
