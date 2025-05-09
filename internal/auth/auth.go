package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func CheckPasswordHash(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	ts := time.Now().UTC()
	start := jwt.NewNumericDate(ts)
	end := jwt.NewNumericDate(ts.Add(expiresIn))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{Issuer: "chirpy", IssuedAt: start, ExpiresAt: end, Subject: userID.String()})
	return token.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (interface{}, error) { return []byte(tokenSecret), nil },
	)
	if err != nil {
		return uuid.UUID{}, err
	}
	rawId, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.UUID{}, err
	}
	uid, err := uuid.Parse(rawId)
	if err != nil {
		return uuid.UUID{}, err
	}
	return uid, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	raw := headers.Get("Authorization")
	if len(raw) == 0 {
		return "", errors.New("No token")
	}
	token, found := strings.CutPrefix(raw, "Bearer ")
	if !found {
		return "", errors.New("No token")
	}
	return token, nil
}

func MakeRefreshToken() (string, error) {
	randBytes := make([]byte, 32)
	rand.Read(randBytes)
	out := hex.EncodeToString(randBytes)
	return out, nil
}

func GetAPIKey(headers http.Header) (string, error) {
	raw := headers.Get("Authorization")
	if len(raw) == 0 {
		return "", errors.New("No token")
	}
	token, found := strings.CutPrefix(raw, "ApiKey ")
	if !found {
		return "", errors.New("No token")
	}
	return token, nil
}
