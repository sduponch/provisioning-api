package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"

	b64 "encoding/base64"
	"errors"
	"strings"
	"time"
)

func (app *App) Auth(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	login, err := app.parseLogin(r)
	if err != nil {
		app.render.JSON(w, http.StatusBadRequest, map[string]string{"message": "malformed login"})
		return
	}
	password, err := app.parsePassword(r)
	if err != nil {
		app.render.JSON(w, http.StatusBadRequest, map[string]string{"message": "malformed password"})
		return
	}
	token, err := app.Login(login, password)
	if err != nil {
		app.render.JSON(w, http.StatusNotAcceptable, map[string]string{"message": "invalid credentials"})
		return
	}
	app.render.JSON(w, http.StatusOK, map[string]string{"bearer_token": b64.StdEncoding.EncodeToString(token), "expire_at": app.tokens[string(token)].expire.UTC().Format(time.RFC3339)})
}

func (app *App) Me(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	user, err := app.parseBearer(r)
	if err != nil {
		app.render.JSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}
	app.render.JSON(w, http.StatusOK, map[string]string{"login": user.login})
}

func (app *App) parseBearer(r *http.Request) (User, error) {
	bearerToken := r.Header.Get("Authorization")
	if bearerToken == "" {
		return User{}, errors.New("bearer token is missing")
	}
	reqToken := strings.Split(bearerToken, " ")[1]
	token, err := b64.StdEncoding.DecodeString(reqToken)
	if err != nil || len(token) != 256 {
		return User{}, errors.New("malformed bearer token")
	}
	user, ok := app.tokens[string(token)]
	if !ok {
		return User{}, errors.New("invalid bearer token")
	}
	return user, nil
}

func (app *App) parseLogin(r *http.Request) (string, error) {
	login := r.URL.Query().Get("login")
	return login, nil
}

func (app *App) parsePassword(r *http.Request) (string, error) {
	password := r.URL.Query().Get("password")
	return password, nil
}
