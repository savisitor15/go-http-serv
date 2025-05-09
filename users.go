package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/savisitor15/go-http-serv/internal/auth"
	"github.com/savisitor15/go-http-serv/internal/database"
)

type UserJSONAll struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	ChirpyRed    bool      `json:"is_chirpy_red"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

type UserJSON struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	ChirpyRed bool      `json:"is_chirpy_red"`
}

func (u UserJSONAll) GetWithoutTokens() UserJSON {
	return UserJSON{
		ID:        u.ID,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		Email:     u.Email,
		ChirpyRed: u.ChirpyRed,
	}
}

func convertDbUserToJSON(u database.User, token string, refresh_token string) UserJSONAll {
	return UserJSONAll{
		ID:           u.ID,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
		Email:        u.Email,
		ChirpyRed:    u.IsChirpyRed,
		Token:        token,
		RefreshToken: refresh_token,
	}
}

func (cfg *apiConfig) handleUserCreation(w http.ResponseWriter, r *http.Request) {
	type reqIn struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	reqin := reqIn{}
	err := decodeRequestBody(r, &reqin)
	if err != nil {
		log.Println("unable to parse user post body", err)
		errorJSONBody(w, 500, err)
		return
	}
	if len(reqin.Email) == 0 || len(reqin.Password) == 0 {
		log.Println("invalid user request!")
		errorJSONBody(w, 400, errors.New("Missing email or password"))
		return
	}
	ts := time.Now()
	hashPass, err := auth.HashPassword(reqin.Password)
	if err != nil {
		log.Println("CreateUser password hashing failed", err)
		errorJSONBody(w, 500, err)
		return
	}
	params := database.CreateUserParams{CreatedAt: ts, UpdatedAt: ts, Email: reqin.Email, HashedPassword: hashPass}
	dbuser, err := cfg.dbConnection.CreateUser(r.Context(), params)
	if err != nil {
		log.Println("user creation failed!", err)
		errorJSONBody(w, 500, err)
		return
	}
	respondJSONBody(w, 201, convertDbUserToJSON(dbuser, "", ""))
}

func ClampExpiryTime(in float64) float64 {
	def := time.Hour.Seconds()
	if in > def || in <= 0.0 {
		return def
	}
	return in
}

func (cfg *apiConfig) RevokeRefreshToken(r *http.Request, in database.RefreshToken) error {
	return cfg.dbConnection.ExpireToken(r.Context(), database.ExpireTokenParams{Token: in.Token, RevokedAt: sql.NullTime{Valid: true, Time: time.Now().UTC()}})
}

func (cfg *apiConfig) FindRefreshToken(r *http.Request, in string) (database.RefreshToken, error) {
	// get the current record in database for the token
	token, err := cfg.dbConnection.GetToken(r.Context(), in)
	if err != nil {
		return database.RefreshToken{}, err
	}
	// Sanity check
	if token.ExpiresAt.Before(time.Now()) {
		// expired make sure it's flagged
		if token.RevokedAt.Valid == false {
			// this should not happen
			cfg.RevokeRefreshToken(r, token)
		}
	}
	return token, nil
}

func (cfg *apiConfig) NewRefreshToken(uid uuid.UUID, r *http.Request) (string, error) {
	token, _ := auth.MakeRefreshToken()
	ts := time.Now()
	_, err := cfg.FindRefreshToken(r, token)
	if err == nil {
		// token EXISTS
		log.Println("Token conflict!", token)
		return "", errors.New("Server side error occured")
	}
	params := database.CreateRefreshTokenParams{
		Token:     token,
		CreatedAt: ts,
		UserID:    uid,
		ExpiresAt: time.Now().AddDate(0, 0, 60).UTC(),
	}
	ref_token, err := cfg.dbConnection.CreateRefreshToken(r.Context(), params)
	if err != nil {
		log.Println("NewRefreshToken() failed to create the token in DB")
		return "", err
	}
	return ref_token.Token, nil
}

func (cfg *apiConfig) handlerCheckLogin(w http.ResponseWriter, r *http.Request) {
	type reqIn struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	// Out is a UserJSON
	reqin := reqIn{}
	err := decodeRequestBody(r, &reqin)
	if err != nil {
		log.Println("unable to parse user post body", err)
		errorJSONBody(w, 500, err)
		return
	}
	dbuser, err := cfg.dbConnection.GetUserByEmail(r.Context(), reqin.Email)
	if err != nil {
		errorJSONBody(w, 404, errors.New("User not found"))
		return
	}
	// check password
	err = auth.CheckPasswordHash(dbuser.HashedPassword, reqin.Password)
	if err != nil {
		errorJSONBody(w, 401, errors.New("Unauthorized access"))
		return
	}
	token, err := auth.MakeJWT(dbuser.ID, cfg.supserSecret, time.Hour)
	if err != nil {
		log.Println("handlerCheckLogin() failed to generate token", err)
		errorJSONBody(w, 500, errors.New("Failed to generate token"))
		return
	}
	ref_token, err := cfg.NewRefreshToken(dbuser.ID, r)
	respondJSONBody(w, 200, convertDbUserToJSON(dbuser, token, ref_token))
}

func (cfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	type responseOut struct {
		Token string `json:"token"`
	}
	inToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		errorJSONBody(w, 403, err)
		return
	}
	token, err := cfg.FindRefreshToken(r, inToken)
	if err != nil {
		log.Println("handlerRefreshToken() error finding token in database", err)
		errorJSONBody(w, 500, errors.New("failed to read token"))
		return
	}
	ts := time.Now().UTC()
	if token.RevokedAt.Valid && token.RevokedAt.Time.Before(ts) {
		log.Println("handlerRefreshToken() token expired")
		errorJSONBody(w, 401, errors.New("expired"))
		return
	}
	newAccess, err := auth.MakeJWT(token.UserID, cfg.supserSecret, time.Hour)
	if err != nil {
		log.Println("handlerRefreshToken() jwt token failed", err)
		errorJSONBody(w, 500, errors.New("server side error"))
		return
	}
	out := responseOut{Token: newAccess}
	respondJSONBody(w, 200, out)
}

func (cfg *apiConfig) handlerRevokeToken(w http.ResponseWriter, r *http.Request) {
	inToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		errorJSONBody(w, 403, err)
		return
	}
	token, err := cfg.FindRefreshToken(r, inToken)
	if err != nil {
		log.Println("handlerRevokeToken() error finding token in database", err)
		errorJSONBody(w, 500, errors.New("failed to read token"))
		return
	}
	err = cfg.RevokeRefreshToken(r, token)
	if err != nil {
		log.Println("handlerRevokeToken() error updating refresh token", err)
		errorJSONBody(w, 500, errors.New("error updating token"))
		return
	}
	respondJSONBody(w, 204, nil)
}

func (cfg *apiConfig) GetUserAuthority(r *http.Request) (database.User, error) {
	inToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Println("handlerUpdatePassword() error getting the token", err)
		return database.User{}, errors.New("error unauthenticated")
	}
	uid, err := auth.ValidateJWT(inToken, cfg.supserSecret)
	if err != nil {
		log.Println("handlerUpdatePassword() error getting uuid from token", err)
		return database.User{}, errors.New("error unauthenticated")
	}
	user, err := cfg.dbConnection.GetUserByID(r.Context(), uid)
	if err != nil {
		log.Println("handlerUpdatePassword() error getting the user from the token uuid", err)
		return database.User{}, errors.New("error unauthenticated")
	}
	return user, nil
}

func (cfg *apiConfig) handlerUpdatePassword(w http.ResponseWriter, r *http.Request) {
	type reqIn struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	user, err := cfg.GetUserAuthority(r)
	if err != nil {
		errorJSONBody(w, 401, err)
		return
	}
	reqin := reqIn{}
	err = decodeRequestBody(r, &reqin)
	if err != nil {
		log.Println("handlerUpdatePassword() error decoding request body", err)
		errorJSONBody(w, 500, err)
		return
	}
	// Sanity check
	user_sec, err := cfg.dbConnection.GetUserByEmail(r.Context(), reqin.Email)
	if err == nil {
		// user is already in the database, same user?
		if user_sec.ID != user.ID {
			// Conflict!
			log.Println("handlerUpdatePassword() user tried to update email to another user's email Offender:", user.ID, "Victim:", user_sec.ID, "Email:", reqin.Email)
			errorJSONBody(w, 403, nil)
			return
		}
	}
	// We should have everything to prepare to update the user
	hashed, err := auth.HashPassword(reqin.Password)
	if err != nil {
		log.Println("handlerUpdatePassword() hashing password error", err)
		errorJSONBody(w, 500, errors.New("server side error"))
		return
	}
	params := database.UpdateUserByIDParams{ID: user.ID, HashedPassword: hashed, Email: reqin.Email}
	user_new, err := cfg.dbConnection.UpdateUserByID(r.Context(), params)
	if err != nil {
		log.Println("handlerUpdatePassword() error update user record", err)
		errorJSONBody(w, 500, errors.New("error updating user record"))
		return
	}
	respondJSONBody(w, 200, convertDbUserToJSON(user_new, "", "").GetWithoutTokens())
}
